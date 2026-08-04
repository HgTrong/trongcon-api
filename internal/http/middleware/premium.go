package middleware

import (
	"net/http"

	"trongcon-api/internal/entity"
	"trongcon-api/internal/service"

	"github.com/gin-gonic/gin"
)

// RequirePremium blocks non-premium authenticated users.
// PT / super bypass — trainers manage catalog content in PT Studio without a paid sub.
func RequirePremium(checker service.UserSubscriptionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := GetUserID(c)
		if !ok || userID == 0 {
			abortUnauthorized(c, "unauthorized")
			return
		}
		if HasRole(c, entity.RoleSuper) || HasRole(c, entity.RolePT) {
			c.Next()
			return
		}
		okPrem, err := checker.IsPremium(c.Request.Context(), userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "premium check failed"})
			return
		}
		if !okPrem {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "premium_required"})
			return
		}
		c.Next()
	}
}

// RequirePremiumOrAuthDetail: for public detail routes with OptionalAuth already applied.
// Anonymous or free users get 403 premium_required. PT / super bypass.
func RequirePremiumDetail(checker service.UserSubscriptionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := GetUserID(c)
		if !ok || userID == 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "premium_required"})
			return
		}
		if HasRole(c, entity.RoleSuper) || HasRole(c, entity.RolePT) {
			c.Next()
			return
		}
		okPrem, err := checker.IsPremium(c.Request.Context(), userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "premium check failed"})
			return
		}
		if !okPrem {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "premium_required"})
			return
		}
		c.Next()
	}
}
