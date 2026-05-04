package payments

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"affiliatetrack/backend/internal/config"
	"affiliatetrack/backend/internal/products"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	db *sql.DB
}

type purchaseRequest struct {
	ProductID    int64  `json:"product_id"`
	ReferralCode string `json:"referral_code"`
}

type conversionResponse struct {
	ID                    int64     `json:"id"`
	AffiliateID           *int64    `json:"affiliate_id"`
	SellerID              int64     `json:"seller_id"`
	ProductID             int64     `json:"product_id"`
	AmountCents           int64     `json:"amount_cents"`
	CommissionAmountCents int64     `json:"commission_amount_cents"`
	PaymentReference      string    `json:"payment_reference"`
	CreatedAt             time.Time `json:"created_at"`
}

func RegisterRoutes(purchase *gin.RouterGroup, _ config.Config, database *sql.DB) {
	handler := Handler{db: database}

	purchase.POST("", handler.completeTestPurchase)
}

func (h Handler) completeTestPurchase(c *gin.Context) {
	var req purchaseRequest
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

	conversion, err := h.recordTestConversion(product, req.ReferralCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not record test purchase"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "test purchase recorded",
		"conversion": conversion,
	})
}

func (h Handler) recordTestConversion(product products.Product, referralCode string) (conversionResponse, error) {
	var conversion conversionResponse

	paymentReferenceColumn, err := h.paymentReferenceColumn()
	if err != nil {
		return conversion, err
	}

	var affiliateID sql.NullInt64
	referralCode = strings.TrimSpace(referralCode)
	if referralCode != "" {
		err := h.db.QueryRow(`
			SELECT id
			FROM users
			WHERE referral_code = $1 AND role = 'affiliate'
		`, referralCode).Scan(&affiliateID)
		if err != nil && err != sql.ErrNoRows {
			return conversion, err
		}
	}

	commissionAmount := int64(0)
	var affiliateValue any
	if affiliateID.Valid {
		affiliateValue = affiliateID.Int64
		commissionAmount = int64(math.Round(float64(product.PriceCents) * (product.CommissionPercent / 100)))
	}

	paymentReference, err := newPaymentReference()
	if err != nil {
		return conversion, err
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
		RETURNING id, affiliate_id, seller_id, product_id, amount_cents,
			commission_amount_cents, %s, created_at
	`, paymentReferenceColumn, paymentReferenceColumn)

	var returnedAffiliateID sql.NullInt64
	err = h.db.QueryRow(
		query,
		affiliateValue,
		product.SellerID,
		product.ID,
		product.PriceCents,
		commissionAmount,
		paymentReference,
	).Scan(
		&conversion.ID,
		&returnedAffiliateID,
		&conversion.SellerID,
		&conversion.ProductID,
		&conversion.AmountCents,
		&conversion.CommissionAmountCents,
		&conversion.PaymentReference,
		&conversion.CreatedAt,
	)
	if err != nil {
		return conversion, err
	}

	if returnedAffiliateID.Valid {
		id := returnedAffiliateID.Int64
		conversion.AffiliateID = &id
	}

	return conversion, nil
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

func newPaymentReference() (string, error) {
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "test_" + hex.EncodeToString(bytes), nil
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
