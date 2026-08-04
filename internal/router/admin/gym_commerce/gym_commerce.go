package gym_commerce

import (
	gcctl "trongcon-api/internal/controller/gym_commerce"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.RouterGroup, c *gcctl.Controller) {
	plans := r.Group("/gym-membership-plans")
	{
		plans.GET("", c.ListPlans)
		plans.POST("", c.CreatePlan)
		plans.PUT("/:id", c.UpdatePlan)
		plans.PATCH("/:id/highlight", c.SetPlanHighlight)
		plans.DELETE("/:id", c.DeletePlan)
	}

	r.GET("/user-gym-memberships", c.ListUserGymMemberships)
	r.POST("/user-gym-memberships/:id/activate", c.AdminActivateMembership)
	r.POST("/user-gym-memberships/:id/cancel", c.AdminCancelMembership)
	r.POST("/gym-check-ins/verify", c.VerifyCheckIn)
	r.GET("/gym-check-ins", c.ListRecentCheckIns)

	classes := r.Group("/group-classes")
	{
		classes.GET("", c.ListGroupClasses)
		classes.POST("", c.CreateGroupClass)
		classes.PUT("/:id", c.UpdateGroupClass)
		classes.DELETE("/:id", c.DeleteGroupClass)
	}

	sessions := r.Group("/class-sessions")
	{
		sessions.GET("", c.ListClassSessions)
		sessions.POST("", c.CreateClassSession)
		sessions.DELETE("/:id", c.DeleteClassSession)
	}

	revShare := r.Group("/revenue-share")
	{
		revShare.GET("", c.GetRevenueShare)
		revShare.PUT("", c.UpdateRevenueShare)
	}

	r.GET("/pt-earnings", c.ListPTEarnings)
	r.PATCH("/pt-earnings/:id/paid-out", c.SetPTEarningPaidOut)
	r.GET("/pt-packages", c.ListPTPackagesAdmin)
	r.GET("/user-pt-packages", c.ListUserPTPackagesAdmin)
	r.GET("/user-pt-packages/:id/sessions", c.ListPTSessionsAdmin)
	r.GET("/trainers/:id/ops-overview", c.AdminTrainerOpsOverview)
	r.GET("/trainers/:id/clients", c.AdminListTrainerClients)
	r.GET("/trainers/:id/bookings", c.AdminListTrainerBookings)
	r.GET("/trainers/:id/content-funnel", c.AdminContentFunnel)
	r.GET("/trainers/:id/quality", c.AdminTrainerQuality)
	r.GET("/trainers/:id/calendar", c.AdminTrainerCalendar)
	r.GET("/trainers/:id/reviews", c.ListPTReviews)
}
