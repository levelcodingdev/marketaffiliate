package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"affiliatetrack/backend/internal/config"
	"affiliatetrack/backend/internal/users"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	cfg config.Config
	db  *sql.DB
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string     `json:"token"`
	User  users.User `json:"user"`
}

func RegisterRoutes(group *gin.RouterGroup, cfg config.Config, database *sql.DB) {
	handler := Handler{cfg: cfg, db: database}

	group.POST("/register", handler.register)
	group.POST("/login", handler.login)
}

func (h Handler) register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}

	email := normalizeEmail(req.Email)
	role, err := h.signupRole(req.Role)
	if email == "" || len(req.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and password with at least 8 characters are required"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not validate signup role"})
		return
	}
	if role == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role must be affiliate or seller"})
		return
	}

	exists, err := h.emailExists(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not check email"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not secure password"})
		return
	}

	user, err := h.createUser(email, string(hash), role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create user"})
		return
	}

	token, err := tokenForUser(h.cfg.JWTSecret, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create session"})
		return
	}

	c.JSON(http.StatusCreated, authResponse{Token: token, User: user})
}

func (h Handler) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}

	user, err := h.findUserByEmail(normalizeEmail(req.Email))
	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load user"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	token, err := tokenForUser(h.cfg.JWTSecret, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create session"})
		return
	}

	c.JSON(http.StatusOK, authResponse{Token: token, User: user})
}

func (h Handler) emailExists(email string) (bool, error) {
	var exists bool
	err := h.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email).Scan(&exists)
	return exists, err
}

func (h Handler) createUser(email string, passwordHash string, role string) (users.User, error) {
	var user users.User

	for attempt := 0; attempt < 5; attempt++ {
		referralCode, err := newReferralCode()
		if err != nil {
			return user, err
		}

		err = h.db.QueryRow(`
			INSERT INTO users (email, password_hash, role, referral_code)
			VALUES ($1, $2, $3, $4)
			RETURNING id, email, password_hash, role, referral_code, created_at
		`, email, passwordHash, role, referralCode).Scan(
			&user.ID,
			&user.Email,
			&user.PasswordHash,
			&user.Role,
			&user.ReferralCode,
			&user.CreatedAt,
		)
		if err == nil {
			return user, nil
		}
	}

	return user, sql.ErrNoRows
}

func (h Handler) findUserByEmail(email string) (users.User, error) {
	var user users.User
	err := h.db.QueryRow(`
		SELECT id, email, password_hash, role, referral_code, created_at
		FROM users
		WHERE email = $1
	`, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.ReferralCode,
		&user.CreatedAt,
	)

	return user, err
}

func tokenForUser(secret string, user users.User) (string, error) {
	return GenerateToken(secret, Claims{
		UserID:       user.ID,
		Email:        user.Email,
		Role:         user.Role,
		ReferralCode: user.ReferralCode,
	}, 7*24*time.Hour)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (h Handler) signupRole(role string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "", RoleAffiliate:
		return RoleAffiliate, nil
	case RoleSeller:
		return RoleSeller, nil
	default:
		return "", nil
	}
}

func newReferralCode() (string, error) {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "AFF-" + strings.ToUpper(hex.EncodeToString(bytes)), nil
}
