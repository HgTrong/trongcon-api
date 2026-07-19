package trainer

import (
	gymctl "trongcon-api/internal/controller/gym"

	"github.com/gin-gonic/gin"
)

func Register(g *gin.RouterGroup, c *gymctl.Controller) {
	tg := g.Group("/trainers")
	{
		tg.POST("", c.CreateTrainer)
		tg.GET("", c.ListTrainers)
		tg.GET("/:id", c.GetTrainer)
		tg.PUT("/:id", c.UpdateTrainer)
		tg.DELETE("/:id", c.DeleteTrainer)
	}
}
