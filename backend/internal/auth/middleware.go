package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	RoleAdmin     = "admin"
	RoleSeller    = "seller"
	RoleAffiliate = "affiliate"
)

const contextUserKey = "auth_user"

type Claims struct {
	UserID       int64  `json:"user_id"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	ReferralCode string `json:"referral_code"`
	ExpiresAt    int64  `json:"exp"`
}

func GenerateToken(secret string, claims Claims, ttl time.Duration) (string, error) {
	claims.ExpiresAt = time.Now().Add(ttl).Unix()

	headerJSON, err := json.Marshal(map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	})
	if err != nil {
		return "", err
	}

	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	signingInput := base64.RawURLEncoding.EncodeToString(headerJSON) + "." + base64.RawURLEncoding.EncodeToString(payloadJSON)
	signature := sign(secret, signingInput)

	return signingInput + "." + signature, nil
}

func ParseToken(secret string, token string) (Claims, error) {
	var claims Claims

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return claims, errors.New("invalid token format")
	}

	signingInput := parts[0] + "." + parts[1]
	expected := sign(secret, signingInput)
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return claims, errors.New("invalid token signature")
	}

	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return claims, err
	}

	var header struct {
		Alg string `json:"alg"`
		Typ string `json:"typ"`
	}
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return claims, err
	}
	if header.Alg != "HS256" || header.Typ != "JWT" {
		return claims, errors.New("unsupported token header")
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return claims, err
	}
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return claims, err
	}
	if claims.ExpiresAt < time.Now().Unix() {
		return claims, errors.New("token expired")
	}

	return claims, nil
}

func RequireAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearerToken(c.GetHeader("Authorization"))
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			c.Abort()
			return
		}

		claims, err := ParseToken(secret, token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid bearer token"})
			c.Abort()
			return
		}

		c.Set(contextUserKey, claims)
		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := CurrentUser(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}

		for _, role := range roles {
			if claims.Role == role {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient role"})
		c.Abort()
	}
}

func CurrentUser(c *gin.Context) (Claims, bool) {
	value, ok := c.Get(contextUserKey)
	if !ok {
		return Claims{}, false
	}

	claims, ok := value.(Claims)
	return claims, ok
}

func bearerToken(header string) string {
	prefix := "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}

func sign(secret string, value string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(value))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
