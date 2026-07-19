package branch

import (
	gymctl "trongcon-api/internal/controller/gym"

	"github.com/gin-gonic/gin"
)

func Register(g *gin.RouterGroup, c *gymctl.Controller) {
	bg := g.Group("/branches")
	{
		bg.GET("", c.ListBranchesPublic)
		bg.GET("/:id", c.GetBranchPublic)
	}
}
