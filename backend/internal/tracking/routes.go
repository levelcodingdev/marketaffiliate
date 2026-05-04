package tracking

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"affiliatetrack/backend/internal/config"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	db *sql.DB
}

func RegisterRoutes(group *gin.RouterGroup, cfg config.Config, database *sql.DB) {
	handler := Handler{db: database}
	group.GET("", handler.track)
}

func (h Handler) track(c *gin.Context) {
	referralCode := strings.TrimSpace(c.Query("ref"))
	productID, err := strconv.ParseInt(c.Query("product_id"), 10, 64)
	if referralCode == "" || err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ref and product_id are required"})
		return
	}

	var clickID int64
	err = h.db.QueryRow(`
		INSERT INTO clicks (referral_code, product_id)
		VALUES ($1, $2)
		RETURNING id
	`, referralCode, productID).Scan(&clickID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not track click"})
		return
	}

	maxAge := 30 * 24 * 60 * 60
	c.SetCookie("referral_code", referralCode, maxAge, "/", "", false, true)
	c.SetCookie("referral_product_id", strconv.FormatInt(productID, 10), maxAge, "/", "", false, true)

	c.JSON(http.StatusCreated, gin.H{
		"tracked":    true,
		"click_id":   clickID,
		"product_id": productID,
		"ref":        referralCode,
	})
}
