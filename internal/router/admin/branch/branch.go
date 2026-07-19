package branch

import (
	gymctl "trongcon-api/internal/controller/gym"

	"github.com/gin-gonic/gin"
)

func Register(g *gin.RouterGroup, c *gymctl.Controller) {
	bg := g.Group("/branches")
	{
		bg.POST("", c.CreateBranch)
		bg.GET("", c.ListBranches)
		bg.GET("/:id", c.GetBranch)
		bg.PUT("/:id", c.UpdateBranch)
		bg.DELETE("/:id", c.DeleteBranch)
	}
}
