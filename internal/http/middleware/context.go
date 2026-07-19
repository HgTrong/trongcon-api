package middleware

import (
	"github.com/gin-gonic/gin"
)

const (
	ContextUserID = "userID"
	ContextRoles  = "roles"
)

func SetAuthContext(c *gin.Context, userID uint, roles []string) {
	c.Set(ContextUserID, userID)
	c.Set(ContextRoles, roles)
}

func GetUserID(c *gin.Context) (uint, bool) {
	v, ok := c.Get(ContextUserID)
	if !ok {
		return 0, false
	}
	id, ok := v.(uint)
	return id, ok
}

func GetRoles(c *gin.Context) []string {
	v, ok := c.Get(ContextRoles)
	if !ok {
		return nil
	}
	roles, ok := v.([]string)
	if !ok {
		return nil
	}
	return roles
}

func HasRole(c *gin.Context, role string) bool {
	for _, r := range GetRoles(c) {
		if r == role {
			return true
		}
	}
	return false
}
