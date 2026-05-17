package server

import (
	"database/sql"
	"net/http"

	"affiliatetrack/backend/internal/auth"
	"affiliatetrack/backend/internal/config"
	"affiliatetrack/backend/internal/dashboard"
	"affiliatetrack/backend/internal/payments"
	"affiliatetrack/backend/internal/products"
	"affiliatetrack/backend/internal/tracking"

	"github.com/gin-gonic/gin"
)

func NewRouter(cfg config.Config, database *sql.DB) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.HandleMethodNotAllowed = true
	router.Use(cors(cfg))
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
	})
	router.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	auth.RegisterRoutes(router.Group("/auth"), cfg, database)
	products.RegisterRoutes(router.Group("/products"), cfg, database)
	tracking.RegisterRoutes(router.Group("/track"), cfg, database)
	payments.RegisterRoutes(router.Group("/checkout-session"), router.Group("/webhook"), cfg, database)
	dashboard.RegisterRoutes(router.Group("/dashboard"), cfg, database)

	return router
}

func cors(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == cfg.FrontendURL || cfg.AppEnv != "production" {
			if origin == "" {
				origin = cfg.FrontendURL
			}
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}

		c.Writer.Header().Set("Vary", "Origin")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Stripe-Signature")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
