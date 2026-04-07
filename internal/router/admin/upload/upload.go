package upload

import (
	uploadctl "trongcon-api/internal/controller/upload"

	"github.com/gin-gonic/gin"
)

func Register(g *gin.RouterGroup, c *uploadctl.Controller) {
	g.POST("/upload", c.Upload)
}
