package subscription_plan

import (
	planctl "trongcon-api/internal/controller/subscription_plan"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.RouterGroup, c *planctl.Controller) {
	r.GET("/subscription-plans", c.ListPublic)
	r.GET("/subscription-plans/:id", c.GetByID)
}
