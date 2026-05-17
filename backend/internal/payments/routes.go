package payments

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"

	"affiliatetrack/backend/internal/config"
	"affiliatetrack/backend/internal/products"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v85"
	checkoutsession "github.com/stripe/stripe-go/v85/checkout/session"
	"github.com/stripe/stripe-go/v85/webhook"
)

type Handler struct {
	cfg config.Config
	db  *sql.DB
}

type checkoutRequest struct {
	ProductID    int64  `json:"product_id"`
	ReferralCode string `json:"referral_code"`
	SuccessURL   string `json:"success_url"`
	CancelURL    string `json:"cancel_url"`
}

func RegisterRoutes(checkout *gin.RouterGroup, webhookGroup *gin.RouterGroup, cfg config.Config, database *sql.DB) {
	handler := Handler{cfg: cfg, db: database}

	checkout.POST("", handler.createCheckoutSession)
	webhookGroup.POST("/stripe", handler.stripeWebhook)
}

func (h Handler) createCheckoutSession(c *gin.Context) {
	if h.cfg.StripeSecretKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "STRIPE_SECRET_KEY is not configured"})
		return
	}

	var req checkoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	if req.ProductID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product_id is required"})
		return
	}
	if strings.TrimSpace(req.ReferralCode) == "" {
		if cookie, err := c.Cookie("referral_code"); err == nil {
			req.ReferralCode = cookie
		}
	}

	product, err := h.findProduct(req.ProductID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load product"})
		return
	}

	session, err := h.createStripeSession(product, req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "stripe checkout failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"checkout_session_id": session.ID,
		"checkout_url":        session.URL,
	})
}

func (h Handler) stripeWebhook(c *gin.Context) {
	if h.cfg.StripeWebhookSecret == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "STRIPE_WEBHOOK_SECRET is not configured"})
		return
	}

	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not read webhook payload"})
		return
	}

	event, err := webhook.ConstructEvent(payload, c.GetHeader("Stripe-Signature"), h.cfg.StripeWebhookSecret)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid stripe signature"})
		return
	}

	if event.Type == "checkout.session.completed" {
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid checkout session"})
			return
		}

		if err := h.recordConversion(&session); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not record conversion"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}

func (h Handler) createStripeSession(product products.Product, req checkoutRequest) (*stripe.CheckoutSession, error) {
	successURL := strings.TrimSpace(req.SuccessURL)
	cancelURL := strings.TrimSpace(req.CancelURL)
	productPath := "/product/" + strconv.FormatInt(product.ID, 10)
	if successURL == "" {
		successURL = strings.TrimRight(h.cfg.FrontendURL, "/") + productPath + "?checkout=success"
	}
	if cancelURL == "" {
		cancelURL = strings.TrimRight(h.cfg.FrontendURL, "/") + productPath + "?checkout=cancelled"
	}

	stripe.Key = h.cfg.StripeSecretKey

	metadata := map[string]string{
		"product_id": strconv.FormatInt(product.ID, 10),
	}
	if strings.TrimSpace(req.ReferralCode) != "" {
		metadata["referral_code"] = strings.TrimSpace(req.ReferralCode)
	}

	productData := &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
		Name: stripe.String(product.Name),
	}
	if strings.TrimSpace(product.Description) != "" {
		productData.Description = stripe.String(product.Description)
	}

	return checkoutsession.New(&stripe.CheckoutSessionParams{
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		Metadata:   metadata,
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Quantity: stripe.Int64(1),
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency:    stripe.String(strings.ToLower(h.cfg.StripeCurrency)),
					UnitAmount:  stripe.Int64(product.PriceCents),
					ProductData: productData,
				},
			},
		},
	})
}

func (h Handler) recordConversion(session *stripe.CheckoutSession) error {
	if session == nil || session.ID == "" {
		return errors.New("missing stripe checkout session id")
	}

	productID, err := strconv.ParseInt(session.Metadata["product_id"], 10, 64)
	if err != nil || productID <= 0 {
		return errors.New("missing product metadata")
	}

	product, err := h.findProduct(productID)
	if err != nil {
		return err
	}

	paymentReferenceColumn, err := h.paymentReferenceColumn()
	if err != nil {
		return err
	}

	amountCents := session.AmountTotal
	if amountCents <= 0 {
		amountCents = product.PriceCents
	}

	var affiliateID sql.NullInt64
	referralCode := strings.TrimSpace(session.Metadata["referral_code"])
	if referralCode != "" {
		err := h.db.QueryRow(`
			SELECT id
			FROM users
			WHERE referral_code = $1 AND role = 'affiliate'
		`, referralCode).Scan(&affiliateID)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
	}

	commissionAmount := int64(0)
	var affiliateValue any
	if affiliateID.Valid {
		affiliateValue = affiliateID.Int64
		commissionAmount = int64(math.Round(float64(amountCents) * (product.CommissionPercent / 100)))
	}

	query := fmt.Sprintf(`
		INSERT INTO conversions (
			affiliate_id,
			seller_id,
			product_id,
			amount_cents,
			commission_amount_cents,
			%s
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (%s) DO NOTHING
	`, paymentReferenceColumn, paymentReferenceColumn)

	_, err = h.db.Exec(query, affiliateValue, product.SellerID, product.ID, amountCents, commissionAmount, session.ID)
	return err
}

func (h Handler) paymentReferenceColumn() (string, error) {
	var column string
	err := h.db.QueryRow(`
		SELECT column_name
		FROM information_schema.columns
		WHERE table_name = 'conversions'
			AND column_name IN ('payment_reference', 'stripe_session_id')
		ORDER BY CASE WHEN column_name = 'payment_reference' THEN 0 ELSE 1 END
		LIMIT 1
	`).Scan(&column)
	if err == sql.ErrNoRows {
		return "", errors.New("conversions payment reference column is missing")
	}
	return column, err
}

func (h Handler) findProduct(id int64) (products.Product, error) {
	var product products.Product
	err := h.db.QueryRow(`
		SELECT id, seller_id, name, description, price_cents,
			CAST(commission_percent AS DOUBLE PRECISION), created_at
		FROM products
		WHERE id = $1
	`, id).Scan(
		&product.ID,
		&product.SellerID,
		&product.Name,
		&product.Description,
		&product.PriceCents,
		&product.CommissionPercent,
		&product.CreatedAt,
	)
	return product, err
}
