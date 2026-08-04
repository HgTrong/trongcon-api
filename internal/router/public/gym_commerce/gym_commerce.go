package gym_commerce

import (
	gcctl "trongcon-api/internal/controller/gym_commerce"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.RouterGroup, c *gcctl.Controller) {
	r.GET("/gym-membership-plans", c.ListPlansPublic)
	r.GET("/gym-membership-plans/highlighted", c.ListHighlightedPlansPublic)
	r.GET("/class-sessions/upcoming", c.ListUpcomingClassSessionsPublic)
	r.GET("/trainers/:id/pt-packages", c.ListPublicPTPackagesByTrainer)
	r.GET("/trainers/:id/available-slots", c.ListAvailableSlots)
	r.GET("/trainers/:id/reviews", c.ListPTReviews)
}
