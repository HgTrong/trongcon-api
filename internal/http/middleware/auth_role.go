package middleware

import (
	"net/http"
	"strings"

	"trongcon-api/internal/jwtutil"

	"github.com/gin-gonic/gin"
)

// parseBearerClaims đọc JWT từ Authorization header (pattern tương tự strongbody-api).
func parseBearerClaims(c *gin.Context, jwtSecret string) (*jwtutil.Claims, error) {
	h := c.GetHeader("Authorization")
	if h == "" || !strings.HasPrefix(strings.ToLower(h), "bearer ") {
		return nil, jwtutil.ErrInvalidToken
	}
	raw := strings.TrimSpace(h[7:])
	claims, err := jwtutil.Parse(raw, []byte(jwtSecret))
	if err != nil {
		return nil, err
	}
	purpose := claims.Purpose
	if purpose == "" {
		purpose = jwtutil.PurposeAccess
	}
	if purpose != jwtutil.PurposeAccess {
		return nil, jwtutil.ErrInvalidToken
	}
	return claims, nil
}

func abortUnauthorized(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": msg})
}

func abortForbidden(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": msg})
}

// RequireAuth — bắt buộc JWT hợp lệ (bất kỳ role).
func RequireAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := parseBearerClaims(c, jwtSecret)
		if err != nil {
			abortUnauthorized(c, "invalid token")
			return
		}
		SetAuthContext(c, claims.UserID, claims.Roles)
		c.Next()
	}
}

// RequireAnyRole — JWT hợp lệ + ít nhất một trong các role.
func RequireAnyRole(jwtSecret string, roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(c *gin.Context) {
		claims, err := parseBearerClaims(c, jwtSecret)
		if err != nil {
			abortUnauthorized(c, "missing or invalid authorization")
			return
		}
		ok := false
		for _, r := range claims.Roles {
			if _, found := allowed[r]; found {
				ok = true
				break
			}
		}
		if !ok {
			abortForbidden(c, "insufficient role")
			return
		}
		SetAuthContext(c, claims.UserID, claims.Roles)
		c.Next()
	}
}

// OptionalAuth — nếu có JWT hợp lệ thì gắn context; không có vẫn cho qua (anonymous).
func OptionalAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := parseBearerClaims(c, jwtSecret)
		if err == nil && claims != nil {
			SetAuthContext(c, claims.UserID, claims.Roles)
		}
		c.Next()
	}
}
