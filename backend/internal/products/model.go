package products

import "time"

type Product struct {
	ID                int64     `json:"id"`
	SellerID          int64     `json:"seller_id"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	PriceCents        int64     `json:"price_cents"`
	CommissionPercent float64   `json:"commission_percent"`
	CreatedAt         time.Time `json:"created_at"`
}
