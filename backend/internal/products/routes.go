package products

import (
	"database/sql"
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"

	"affiliatetrack/backend/internal/auth"
	"affiliatetrack/backend/internal/config"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	db *sql.DB
}

type productRequest struct {
	Name              string  `json:"name"`
	Description       string  `json:"description"`
	PriceCents        int64   `json:"price_cents"`
	Price             float64 `json:"price"`
	CommissionPercent float64 `json:"commission_percent"`
}

type productResponse struct {
	Product
	SellerEmail string `json:"seller_email"`
}

func RegisterRoutes(group *gin.RouterGroup, cfg config.Config, database *sql.DB) {
	handler := Handler{db: database}

	group.GET("", handler.list)
	group.GET("/:id", handler.detail)
	group.POST("", auth.RequireAuth(cfg.JWTSecret), auth.RequireRole(auth.RoleSeller), handler.create)
	group.PUT("/:id", auth.RequireAuth(cfg.JWTSecret), auth.RequireRole(auth.RoleSeller, auth.RoleAdmin), handler.update)
	group.DELETE("/:id", auth.RequireAuth(cfg.JWTSecret), auth.RequireRole(auth.RoleSeller, auth.RoleAdmin), handler.delete)
}

func (h Handler) list(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT p.id, p.seller_id, p.name, p.description, p.price_cents,
			CAST(p.commission_percent AS DOUBLE PRECISION), p.created_at, u.email
		FROM products p
		JOIN users u ON u.id = p.seller_id
		ORDER BY p.created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list products"})
		return
	}
	defer rows.Close()

	products := make([]productResponse, 0)
	for rows.Next() {
		var product productResponse
		if err := rows.Scan(
			&product.ID,
			&product.SellerID,
			&product.Name,
			&product.Description,
			&product.PriceCents,
			&product.CommissionPercent,
			&product.CreatedAt,
			&product.SellerEmail,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not read products"})
			return
		}
		products = append(products, product)
	}

	c.JSON(http.StatusOK, gin.H{"products": products})
}

func (h Handler) detail(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}

	product, err := h.find(id)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"product": product})
}

func (h Handler) create(c *gin.Context) {
	user, _ := auth.CurrentUser(c)

	req, ok := bindProductRequest(c)
	if !ok {
		return
	}

	var product Product
	err := h.db.QueryRow(`
		INSERT INTO products (seller_id, name, description, price_cents, commission_percent)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, seller_id, name, description, price_cents,
			CAST(commission_percent AS DOUBLE PRECISION), created_at
	`, user.UserID, req.Name, req.Description, req.PriceCents, req.CommissionPercent).Scan(
		&product.ID,
		&product.SellerID,
		&product.Name,
		&product.Description,
		&product.PriceCents,
		&product.CommissionPercent,
		&product.CreatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create product"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"product": product})
}

func (h Handler) update(c *gin.Context) {
	user, _ := auth.CurrentUser(c)
	id, ok := paramID(c, "id")
	if !ok {
		return
	}

	current, err := h.find(id)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load product"})
		return
	}
	if user.Role != auth.RoleAdmin && current.SellerID != user.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only update your own products"})
		return
	}

	req, ok := bindProductRequest(c)
	if !ok {
		return
	}

	var product Product
	err = h.db.QueryRow(`
		UPDATE products
		SET name = $1, description = $2, price_cents = $3, commission_percent = $4
		WHERE id = $5
		RETURNING id, seller_id, name, description, price_cents,
			CAST(commission_percent AS DOUBLE PRECISION), created_at
	`, req.Name, req.Description, req.PriceCents, req.CommissionPercent, id).Scan(
		&product.ID,
		&product.SellerID,
		&product.Name,
		&product.Description,
		&product.PriceCents,
		&product.CommissionPercent,
		&product.CreatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"product": product})
}

func (h Handler) delete(c *gin.Context) {
	user, _ := auth.CurrentUser(c)
	id, ok := paramID(c, "id")
	if !ok {
		return
	}

	current, err := h.find(id)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load product"})
		return
	}
	if user.Role != auth.RoleAdmin && current.SellerID != user.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only delete your own products"})
		return
	}

	var conversions int64
	if err := h.db.QueryRow(`SELECT COUNT(*) FROM conversions WHERE product_id = $1`, id).Scan(&conversions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not check product conversions"})
		return
	}
	if conversions > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "products with conversions cannot be deleted"})
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not start product delete"})
		return
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM clicks WHERE product_id = $1`, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete product clicks"})
		return
	}
	result, err := tx.Exec(`DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete product"})
		return
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not confirm product delete"})
		return
	}
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not finish product delete"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (h Handler) find(id int64) (productResponse, error) {
	var product productResponse
	err := h.db.QueryRow(`
		SELECT p.id, p.seller_id, p.name, p.description, p.price_cents,
			CAST(p.commission_percent AS DOUBLE PRECISION), p.created_at, u.email
		FROM products p
		JOIN users u ON u.id = p.seller_id
		WHERE p.id = $1
	`, id).Scan(
		&product.ID,
		&product.SellerID,
		&product.Name,
		&product.Description,
		&product.PriceCents,
		&product.CommissionPercent,
		&product.CreatedAt,
		&product.SellerEmail,
	)
	return product, err
}

func bindProductRequest(c *gin.Context) (productRequest, bool) {
	var req productRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return req, false
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	if req.PriceCents == 0 && req.Price > 0 {
		req.PriceCents = int64(math.Round(req.Price * 100))
	}

	if err := validateProduct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return req, false
	}

	return req, true
}

func validateProduct(req productRequest) error {
	if req.Name == "" {
		return errors.New("name is required")
	}
	if req.PriceCents <= 0 {
		return errors.New("price_cents must be greater than zero")
	}
	if req.CommissionPercent < 0 || req.CommissionPercent > 100 {
		return errors.New("commission_percent must be between 0 and 100")
	}
	return nil
}

func paramID(c *gin.Context, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + name})
		return 0, false
	}
	return id, true
}
