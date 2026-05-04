package dashboard

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"affiliatetrack/backend/internal/auth"
	"affiliatetrack/backend/internal/config"
	"affiliatetrack/backend/internal/users"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	cfg config.Config
	db  *sql.DB
}

type affiliateProduct struct {
	ID                int64   `json:"id"`
	Name              string  `json:"name"`
	PriceCents        int64   `json:"price_cents"`
	CommissionPercent float64 `json:"commission_percent"`
	SellerEmail       string  `json:"seller_email"`
	ReferralURL       string  `json:"referral_url"`
}

type sellerProductMetric struct {
	ID                  int64   `json:"id"`
	Name                string  `json:"name"`
	Description         string  `json:"description"`
	PriceCents          int64   `json:"price_cents"`
	CommissionPercent   float64 `json:"commission_percent"`
	Sales               int64   `json:"sales"`
	GrossRevenueCents   int64   `json:"gross_revenue_cents"`
	CommissionPaidCents int64   `json:"commission_paid_cents"`
	NetRevenueCents     int64   `json:"net_revenue_cents"`
}

type sellerAffiliateMetric struct {
	AffiliateID     int64  `json:"affiliate_id"`
	AffiliateEmail  string `json:"affiliate_email"`
	Conversions     int64  `json:"conversions"`
	RevenueCents    int64  `json:"revenue_cents"`
	CommissionCents int64  `json:"commission_cents"`
}

type transaction struct {
	ID                    int64  `json:"id"`
	AffiliateID           *int64 `json:"affiliate_id"`
	AffiliateEmail        string `json:"affiliate_email"`
	SellerID              int64  `json:"seller_id"`
	SellerEmail           string `json:"seller_email"`
	ProductID             int64  `json:"product_id"`
	ProductName           string `json:"product_name"`
	AmountCents           int64  `json:"amount_cents"`
	CommissionAmountCents int64  `json:"commission_amount_cents"`
	PaymentReference      string `json:"payment_reference"`
	CreatedAt             string `json:"created_at"`
}

func RegisterRoutes(group *gin.RouterGroup, cfg config.Config, database *sql.DB) {
	handler := Handler{cfg: cfg, db: database}

	group.GET("/affiliate", auth.RequireAuth(cfg.JWTSecret), auth.RequireRole(auth.RoleAffiliate), handler.affiliate)
	group.GET("/seller", auth.RequireAuth(cfg.JWTSecret), auth.RequireRole(auth.RoleSeller), handler.seller)
	group.GET("/admin", auth.RequireAuth(cfg.JWTSecret), auth.RequireRole(auth.RoleAdmin), handler.admin)
}

func (h Handler) affiliate(c *gin.Context) {
	user, _ := auth.CurrentUser(c)

	var clicks int64
	var conversions int64
	var earningsCents int64

	if err := h.db.QueryRow(`
		SELECT COUNT(*)
		FROM clicks
		WHERE referral_code = $1
	`, user.ReferralCode).Scan(&clicks); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load clicks"})
		return
	}

	if err := h.db.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(commission_amount_cents), 0)::BIGINT
		FROM conversions
		WHERE affiliate_id = $1
	`, user.UserID).Scan(&conversions, &earningsCents); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load conversions"})
		return
	}

	products, err := h.affiliateProducts(user.ReferralCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load referral links"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"referral_code":  user.ReferralCode,
		"clicks":         clicks,
		"conversions":    conversions,
		"earnings_cents": earningsCents,
		"products":       products,
	})
}

func (h Handler) seller(c *gin.Context) {
	user, _ := auth.CurrentUser(c)

	var productsCount int64
	var sales int64
	var grossRevenueCents int64
	var commissionPaidCents int64

	if err := h.db.QueryRow(`
		SELECT COUNT(*)
		FROM products
		WHERE seller_id = $1
	`, user.UserID).Scan(&productsCount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load product count"})
		return
	}

	if err := h.db.QueryRow(`
		SELECT COUNT(*),
			COALESCE(SUM(amount_cents), 0)::BIGINT,
			COALESCE(SUM(commission_amount_cents), 0)::BIGINT
		FROM conversions
		WHERE seller_id = $1
	`, user.UserID).Scan(&sales, &grossRevenueCents, &commissionPaidCents); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load seller totals"})
		return
	}

	productMetrics, err := h.sellerProducts(user.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load product metrics"})
		return
	}

	affiliateMetrics, err := h.sellerAffiliates(user.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load affiliate metrics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"products_count":        productsCount,
		"sales":                 sales,
		"gross_revenue_cents":   grossRevenueCents,
		"commission_paid_cents": commissionPaidCents,
		"net_revenue_cents":     grossRevenueCents - commissionPaidCents,
		"products":              productMetrics,
		"affiliates":            affiliateMetrics,
	})
}

func (h Handler) admin(c *gin.Context) {
	var usersCount int64
	var productsCount int64
	var conversionsCount int64
	var grossRevenueCents int64
	var commissionCents int64

	if err := h.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&usersCount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load users count"})
		return
	}
	if err := h.db.QueryRow(`SELECT COUNT(*) FROM products`).Scan(&productsCount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load products count"})
		return
	}
	if err := h.db.QueryRow(`
		SELECT COUNT(*),
			COALESCE(SUM(amount_cents), 0)::BIGINT,
			COALESCE(SUM(commission_amount_cents), 0)::BIGINT
		FROM conversions
	`).Scan(&conversionsCount, &grossRevenueCents, &commissionCents); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load platform totals"})
		return
	}

	userList, err := h.users()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load users"})
		return
	}

	transactions, err := h.transactions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users_count":          usersCount,
		"products_count":       productsCount,
		"conversions_count":    conversionsCount,
		"gross_revenue_cents":  grossRevenueCents,
		"commission_cents":     commissionCents,
		"seller_revenue_cents": grossRevenueCents - commissionCents,
		"users":                userList,
		"transactions":         transactions,
	})
}

func (h Handler) affiliateProducts(referralCode string) ([]affiliateProduct, error) {
	rows, err := h.db.Query(`
		SELECT p.id, p.name, p.price_cents, CAST(p.commission_percent AS DOUBLE PRECISION), u.email
		FROM products p
		JOIN users u ON u.id = p.seller_id
		ORDER BY p.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]affiliateProduct, 0)
	for rows.Next() {
		var product affiliateProduct
		if err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.PriceCents,
			&product.CommissionPercent,
			&product.SellerEmail,
		); err != nil {
			return nil, err
		}
		product.ReferralURL = strings.TrimRight(h.cfg.FrontendURL, "/") + "/product/" + strconv.FormatInt(product.ID, 10) + "?ref=" + referralCode
		products = append(products, product)
	}

	return products, rows.Err()
}

func (h Handler) sellerProducts(sellerID int64) ([]sellerProductMetric, error) {
	rows, err := h.db.Query(`
		SELECT p.id, p.name, p.description, p.price_cents, CAST(p.commission_percent AS DOUBLE PRECISION),
			COUNT(c.id),
			COALESCE(SUM(c.amount_cents), 0)::BIGINT,
			COALESCE(SUM(c.commission_amount_cents), 0)::BIGINT
		FROM products p
		LEFT JOIN conversions c ON c.product_id = p.id
		WHERE p.seller_id = $1
		GROUP BY p.id, p.name, p.description, p.price_cents, p.commission_percent
		ORDER BY p.created_at DESC
	`, sellerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]sellerProductMetric, 0)
	for rows.Next() {
		var product sellerProductMetric
		if err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.PriceCents,
			&product.CommissionPercent,
			&product.Sales,
			&product.GrossRevenueCents,
			&product.CommissionPaidCents,
		); err != nil {
			return nil, err
		}
		product.NetRevenueCents = product.GrossRevenueCents - product.CommissionPaidCents
		products = append(products, product)
	}

	return products, rows.Err()
}

func (h Handler) sellerAffiliates(sellerID int64) ([]sellerAffiliateMetric, error) {
	rows, err := h.db.Query(`
		SELECT u.id, u.email, COUNT(c.id),
			COALESCE(SUM(c.amount_cents), 0)::BIGINT,
			COALESCE(SUM(c.commission_amount_cents), 0)::BIGINT
		FROM conversions c
		JOIN users u ON u.id = c.affiliate_id
		WHERE c.seller_id = $1
		GROUP BY u.id, u.email
		ORDER BY COUNT(c.id) DESC, u.email ASC
	`, sellerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	affiliates := make([]sellerAffiliateMetric, 0)
	for rows.Next() {
		var affiliate sellerAffiliateMetric
		if err := rows.Scan(
			&affiliate.AffiliateID,
			&affiliate.AffiliateEmail,
			&affiliate.Conversions,
			&affiliate.RevenueCents,
			&affiliate.CommissionCents,
		); err != nil {
			return nil, err
		}
		affiliates = append(affiliates, affiliate)
	}

	return affiliates, rows.Err()
}

func (h Handler) users() ([]users.User, error) {
	rows, err := h.db.Query(`
		SELECT id, email, role, referral_code, created_at
		FROM users
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	userList := make([]users.User, 0)
	for rows.Next() {
		var user users.User
		if err := rows.Scan(&user.ID, &user.Email, &user.Role, &user.ReferralCode, &user.CreatedAt); err != nil {
			return nil, err
		}
		userList = append(userList, user)
	}

	return userList, rows.Err()
}

func (h Handler) transactions() ([]transaction, error) {
	paymentReferenceColumn, err := h.paymentReferenceColumn()
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		SELECT c.id, c.affiliate_id, COALESCE(a.email, ''), s.id, s.email,
			p.id, p.name, c.amount_cents, c.commission_amount_cents,
			c.%s, c.created_at::TEXT
		FROM conversions c
		LEFT JOIN users a ON a.id = c.affiliate_id
		JOIN users s ON s.id = c.seller_id
		JOIN products p ON p.id = c.product_id
		ORDER BY c.created_at DESC
	`, paymentReferenceColumn)

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := make([]transaction, 0)
	for rows.Next() {
		var item transaction
		var affiliateID sql.NullInt64
		if err := rows.Scan(
			&item.ID,
			&affiliateID,
			&item.AffiliateEmail,
			&item.SellerID,
			&item.SellerEmail,
			&item.ProductID,
			&item.ProductName,
			&item.AmountCents,
			&item.CommissionAmountCents,
			&item.PaymentReference,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		if affiliateID.Valid {
			id := affiliateID.Int64
			item.AffiliateID = &id
		}
		transactions = append(transactions, item)
	}

	return transactions, rows.Err()
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
