package trainer

import (
	gymctl "trongcon-api/internal/controller/gym"

	"github.com/gin-gonic/gin"
)

func Register(g *gin.RouterGroup, c *gymctl.Controller) {
	tg := g.Group("/trainers")
	{
		tg.GET("", c.ListTrainersPublic)
		tg.GET("/:id", c.GetTrainerPublic)
		tg.GET("/:id/workouts", c.ListTrainerWorkouts)
		tg.GET("/:id/routines", c.ListTrainerRoutines)
		tg.GET("/:id/meal-plans", c.ListTrainerMealPlans)
	}
}
