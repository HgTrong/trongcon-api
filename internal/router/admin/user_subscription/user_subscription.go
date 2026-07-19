package user_subscription

import (
	subctl "trongcon-api/internal/controller/user_subscription"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.RouterGroup, c *subctl.Controller) {
	r.GET("/user-subscriptions", c.ListAdmin)
}
