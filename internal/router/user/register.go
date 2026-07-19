package user

import (
	authctl "trongcon-api/internal/controller/auth"
	aictl "trongcon-api/internal/controller/ai"
	enrollctl "trongcon-api/internal/controller/training_enrollment"
	foodlogctl "trongcon-api/internal/controller/food_log"
	mytrainctl "trongcon-api/internal/controller/my_train"
	savedctl "trongcon-api/internal/controller/saved_workout"
	sessionctl "trongcon-api/internal/controller/workout_session"
	subctl "trongcon-api/internal/controller/user_subscription"
	userctl "trongcon-api/internal/controller/user"
	"trongcon-api/internal/http/middleware"
	"trongcon-api/internal/service"

	"github.com/gin-gonic/gin"
)

type Controllers struct {
	Auth         *authctl.Controller
	User         *userctl.Controller
	Saved        *savedctl.Controller
	FoodLog      *foodlogctl.Controller
	MyTrain      *mytrainctl.Controller
	Sessions     *sessionctl.Controller
	Enrollment   *enrollctl.Controller
	AI           *aictl.Controller
	Subscription *subctl.Controller
}

func Register(g *gin.RouterGroup, c Controllers, jwtSecret string, premium service.UserSubscriptionService) {
	g.POST("/signup", c.Auth.Signup)
	g.POST("/login", c.Auth.UserLogin)

	authed := g.Group("")
	authed.Use(middleware.RequireAuth(jwtSecret))
	{
		authed.GET("/me", c.User.GetMe)
		authed.PATCH("/me", c.User.UpdateMe)
		authed.PUT("/avatar", c.User.UpdateAvatar)
		authed.POST("/change-password", c.User.ChangePassword)

		if c.Subscription != nil {
			authed.POST("/subscriptions/checkout", c.Subscription.Checkout)
			authed.POST("/subscriptions/checkout/paypal", c.Subscription.CheckoutPayPal)
			authed.POST("/subscriptions/checkout/stripe", c.Subscription.CheckoutStripe)
			authed.POST("/subscriptions/capture", c.Subscription.Capture)
			authed.POST("/subscriptions/stripe/confirm", c.Subscription.ConfirmStripe)
			authed.GET("/subscriptions/me", c.Subscription.Me)
		}

		authed.GET("/saved-workouts", c.Saved.List)
		authed.GET("/saved-workouts/ids", c.Saved.ListIDs)
		authed.POST("/saved-workouts", c.Saved.Save)
		authed.DELETE("/saved-workouts/:workout_id", c.Saved.Unsave)

		premiumGroup := authed.Group("")
		if premium != nil {
			premiumGroup.Use(middleware.RequirePremium(premium))
		}

		premiumGroup.GET("/nutrition-goals", c.FoodLog.GetGoals)
		premiumGroup.PUT("/nutrition-goals", c.FoodLog.UpdateGoals)
		premiumGroup.POST("/nutrition-goals/from-calories", c.FoodLog.SaveFromCalories)
		premiumGroup.GET("/member-stats", c.FoodLog.GetMemberStats)
		premiumGroup.GET("/food-log", c.FoodLog.GetDay)
		premiumGroup.POST("/food-log/meals", c.FoodLog.CreateMeal)
		premiumGroup.PATCH("/food-log/meals/:id", c.FoodLog.UpdateMeal)
		premiumGroup.DELETE("/food-log/meals/:id", c.FoodLog.DeleteMeal)
		premiumGroup.GET("/food-log/recent", c.FoodLog.ListRecent)
		premiumGroup.POST("/food-log/entries", c.FoodLog.CreateEntry)
		premiumGroup.PATCH("/food-log/entries/:id", c.FoodLog.UpdateEntry)
		premiumGroup.DELETE("/food-log/entries/:id", c.FoodLog.DeleteEntry)

		premiumGroup.GET("/my-workouts", c.MyTrain.ListWorkouts)
		premiumGroup.POST("/my-workouts", c.MyTrain.CreateWorkout)
		premiumGroup.POST("/my-workouts/from-catalog", c.MyTrain.CloneWorkout)
		premiumGroup.GET("/my-workouts/:id", c.MyTrain.GetWorkout)
		premiumGroup.PUT("/my-workouts/:id", c.MyTrain.UpdateWorkout)
		premiumGroup.DELETE("/my-workouts/:id", c.MyTrain.DeleteWorkout)

		premiumGroup.GET("/my-routines", c.MyTrain.ListRoutines)
		premiumGroup.POST("/my-routines", c.MyTrain.CreateRoutine)
		premiumGroup.GET("/my-routines/:id", c.MyTrain.GetRoutine)
		premiumGroup.PUT("/my-routines/:id", c.MyTrain.UpdateRoutine)
		premiumGroup.DELETE("/my-routines/:id", c.MyTrain.DeleteRoutine)

		premiumGroup.GET("/workout-sessions", c.Sessions.List)
		premiumGroup.POST("/workout-sessions", c.Sessions.Create)
		premiumGroup.GET("/workout-sessions/:id", c.Sessions.Get)
		premiumGroup.POST("/workout-sessions/:id/complete", c.Sessions.Complete)
		premiumGroup.POST("/workout-sessions/:id/abandon", c.Sessions.Abandon)
		premiumGroup.POST("/workout-sessions/:id/items", c.Sessions.AddItem)
		premiumGroup.PATCH("/workout-sessions/sets/:set_id", c.Sessions.UpdateSet)
		premiumGroup.GET("/exercises/:id/performance", c.Sessions.ExercisePerformance)

		premiumGroup.GET("/training-enrollments", c.Enrollment.List)
		premiumGroup.POST("/training-enrollments", c.Enrollment.Create)
		premiumGroup.GET("/training-enrollments/this-week", c.Enrollment.ThisWeek)
		premiumGroup.GET("/training-enrollments/:id", c.Enrollment.Get)
		premiumGroup.POST("/training-enrollments/:id/cancel", c.Enrollment.Cancel)
		premiumGroup.GET("/training-enrollments/:id/compare", c.Enrollment.Compare)
		premiumGroup.POST("/training-enrollments/slots/:slot_id/start", c.Enrollment.StartSlot)

		if c.AI != nil {
			premiumGroup.POST("/ai/meal-plan/generate", c.AI.GenerateMealPlan)
			premiumGroup.POST("/ai/routine/generate", c.AI.GenerateRoutine)
			premiumGroup.POST("/ai/chat", c.AI.Chat)
			premiumGroup.GET("/ai/chat/threads", c.AI.ListThreads)
			premiumGroup.GET("/ai/chat/threads/:thread_id/messages", c.AI.ListMessages)

			premiumGroup.GET("/my-meal-plans", c.AI.ListMyMealPlans)
			premiumGroup.POST("/my-meal-plans", c.AI.CreateMyMealPlan)
			premiumGroup.GET("/my-meal-plans/:id", c.AI.GetMyMealPlan)
			premiumGroup.PUT("/my-meal-plans/:id", c.AI.UpdateMyMealPlan)
			premiumGroup.DELETE("/my-meal-plans/:id", c.AI.DeleteMyMealPlan)
		}
	}
}
