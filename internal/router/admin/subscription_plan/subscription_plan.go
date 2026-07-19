package subscription_plan

import (
	planctl "trongcon-api/internal/controller/subscription_plan"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.RouterGroup, c *planctl.Controller) {
	g := r.Group("/subscription-plans")
	{
		g.POST("", c.Create)
		g.GET("", c.List)
		g.GET("/:id", c.GetByID)
		g.PUT("/:id", c.Update)
		g.DELETE("/:id", c.Delete)
	}
}
