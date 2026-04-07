package middleware

import (
	"net/http"
	"strings"

	"trongcon-api/internal/entity"
	"trongcon-api/internal/jwtutil"

	"github.com/gin-gonic/gin"
)

func RequireSuper(jwtSecret string) gin.HandlerFunc {
	sec := []byte(jwtSecret)
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" || !strings.HasPrefix(strings.ToLower(h), "bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization"})
			return
		}
		raw := strings.TrimSpace(h[7:])
		claims, err := jwtutil.Parse(raw, sec)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		ok := false
		for _, r := range claims.Roles {
			if r == entity.RoleSuper {
				ok = true
				break
			}
		}
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "super role required"})
			return
		}
		c.Set("userID", claims.UserID)
		c.Next()
	}
}
