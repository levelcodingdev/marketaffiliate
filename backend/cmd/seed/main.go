package main

import (
	"database/sql"
	"fmt"
	"log"

	"affiliatetrack/backend/internal/config"
	"affiliatetrack/backend/internal/db"

	"golang.org/x/crypto/bcrypt"
)

const password = "password"

type seedUser struct {
	Email        string
	Role         string
	ReferralCode string
}

type seedProduct struct {
	SellerEmail       string
	Name              string
	Description       string
	PriceCents        int64
	CommissionPercent float64
}

var users = []seedUser{
	{Email: "admin@affiliatetrack.local", Role: "admin", ReferralCode: "ADMIN"},
	{Email: "seller1@affiliatetrack.local", Role: "seller", ReferralCode: "SELLER-1"},
	{Email: "seller2@affiliatetrack.local", Role: "seller", ReferralCode: "SELLER-2"},
	{Email: "affiliate@affiliatetrack.local", Role: "affiliate", ReferralCode: "AFF-1"},
}

var products = []seedProduct{
	{
		SellerEmail:       "seller1@affiliatetrack.local",
		Name:              "Creator Launch Playbook",
		Description:       "A tactical guide for planning, pricing, and launching a digital product in two weeks.",
		PriceCents:        7900,
		CommissionPercent: 30,
	},
	{
		SellerEmail:       "seller1@affiliatetrack.local",
		Name:              "Content Calendar Kit",
		Description:       "Editable templates for social posts, newsletters, launch emails, and weekly planning.",
		PriceCents:        3900,
		CommissionPercent: 25,
	},
	{
		SellerEmail:       "seller1@affiliatetrack.local",
		Name:              "Premium Notion CRM",
		Description:       "A Notion workspace for tracking leads, deals, follow-ups, and customer history.",
		PriceCents:        4900,
		CommissionPercent: 20,
	},
	{
		SellerEmail:       "seller2@affiliatetrack.local",
		Name:              "SaaS Metrics Dashboard",
		Description:       "Spreadsheet dashboard for MRR, churn, CAC, LTV, cohorts, and growth reporting.",
		PriceCents:        9900,
		CommissionPercent: 35,
	},
	{
		SellerEmail:       "seller2@affiliatetrack.local",
		Name:              "Landing Page Audit",
		Description:       "A practical checklist and teardown template for improving conversion rates.",
		PriceCents:        5900,
		CommissionPercent: 30,
	},
	{
		SellerEmail:       "seller2@affiliatetrack.local",
		Name:              "Email Funnel Blueprint",
		Description:       "A complete welcome, nurture, and sales email sequence blueprint for small teams.",
		PriceCents:        6900,
		CommissionPercent: 28,
	},
}

func main() {
	cfg := config.Load()

	database, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	tx, err := database.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	userIDs := make(map[string]int64, len(users))
	for _, user := range users {
		id, err := upsertUser(tx, user)
		if err != nil {
			log.Fatalf("seed user %s: %v", user.Email, err)
		}
		userIDs[user.Email] = id
	}

	for _, product := range products {
		sellerID, ok := userIDs[product.SellerEmail]
		if !ok {
			log.Fatalf("seller %s was not seeded", product.SellerEmail)
		}
		if err := upsertProduct(tx, sellerID, product); err != nil {
			log.Fatalf("seed product %s: %v", product.Name, err)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Seed complete.")
}

func upsertUser(tx *sql.Tx, user seedUser) (int64, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	var id int64
	err = tx.QueryRow(`
		INSERT INTO users (email, password_hash, role, referral_code)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (email) DO UPDATE
		SET password_hash = EXCLUDED.password_hash,
			role = EXCLUDED.role,
			referral_code = EXCLUDED.referral_code
		RETURNING id
	`, user.Email, string(passwordHash), user.Role, user.ReferralCode).Scan(&id)

	return id, err
}

func upsertProduct(tx *sql.Tx, sellerID int64, product seedProduct) error {
	var id int64
	err := tx.QueryRow(`
		UPDATE products
		SET description = $1,
			price_cents = $2,
			commission_percent = $3
		WHERE seller_id = $4 AND name = $5
		RETURNING id
	`, product.Description, product.PriceCents, product.CommissionPercent, sellerID, product.Name).Scan(&id)
	if err == nil {
		return nil
	}
	if err != sql.ErrNoRows {
		return err
	}

	return tx.QueryRow(`
		INSERT INTO products (seller_id, name, description, price_cents, commission_percent)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, sellerID, product.Name, product.Description, product.PriceCents, product.CommissionPercent).Scan(&id)
}
