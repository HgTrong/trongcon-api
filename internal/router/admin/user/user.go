package user

import (
	userctl "trongcon-api/internal/controller/user"

	"github.com/gin-gonic/gin"
)

func Register(g *gin.RouterGroup, c *userctl.Controller) {
	users := g.Group("/users")
	{
		users.POST("", c.Create)
		users.GET("", c.List)
		users.GET("/:id", c.GetByID)
		users.PUT("/:id", c.Update)
		users.DELETE("/:id", c.Delete)
	}
}
