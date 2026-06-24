package workout

import (
	workoutctl "trongcon-api/internal/controller/workout"

	"github.com/gin-gonic/gin"
)

func Register(g *gin.RouterGroup, c *workoutctl.Controller) {
	wg := g.Group("/workouts")
	{
		wg.POST("", c.Create)
		wg.GET("", c.List)
		wg.GET("/:id", c.GetByID)
		wg.PUT("/:id", c.Update)
		wg.DELETE("/:id", c.Delete)
	}
}
