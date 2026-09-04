package user

import (
	authctl "trongcon-api/internal/controller/auth"
	aictl "trongcon-api/internal/controller/ai"
	contentsharectl "trongcon-api/internal/controller/content_share"
	enrollctl "trongcon-api/internal/controller/training_enrollment"
	foodlogctl "trongcon-api/internal/controller/food_log"
	gymcommercectl "trongcon-api/internal/controller/gym_commerce"
	mytrainctl "trongcon-api/internal/controller/my_train"
	savedctl "trongcon-api/internal/controller/saved_workout"
	sessionctl "trongcon-api/internal/controller/workout_session"
	subctl "trongcon-api/internal/controller/user_subscription"
	uploadctl "trongcon-api/internal/controller/upload"
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
	Upload       *uploadctl.Controller
	GymCommerce  *gymcommercectl.Controller
	ContentShare *contentsharectl.Controller
}

func Register(g *gin.RouterGroup, c Controllers, jwtSecret string, premium service.UserSubscriptionService) {
	g.POST("/signup/request-otp", c.Auth.RequestSignupOTP)
	g.POST("/signup", c.Auth.Signup)
	g.POST("/login", c.Auth.UserLogin)
	g.POST("/forgot-password", c.Auth.ForgotPassword)
	g.POST("/forgot-password/verify", c.Auth.VerifyForgotOTP)
	g.PATCH("/forgot-password/reset", c.Auth.ResetPassword)

	authed := g.Group("")
	authed.Use(middleware.RequireAuth(jwtSecret))
	{
		authed.GET("/me", c.User.GetMe)
		authed.PATCH("/me", c.User.UpdateMe)
		authed.PUT("/avatar", c.User.UpdateAvatar)
		authed.POST("/change-password", c.User.ChangePassword)

		if c.Subscription != nil {
			authed.POST("/subscriptions/checkout", c.Subscription.Checkout)
			authed.POST("/subscriptions/checkout/vnpay", c.Subscription.CheckoutVNPay)
			authed.POST("/subscriptions/checkout/paypal", c.Subscription.CheckoutPayPal) // legacy → VNPay
			authed.POST("/subscriptions/checkout/stripe", c.Subscription.CheckoutStripe)
			authed.POST("/subscriptions/capture", c.Subscription.Capture)
			authed.POST("/subscriptions/vnpay/confirm", c.Subscription.ConfirmVNPay)
			authed.POST("/subscriptions/stripe/confirm", c.Subscription.ConfirmStripe)
			authed.GET("/subscriptions/me", c.Subscription.Me)
			authed.POST("/subscriptions/trial", c.Subscription.StartTrial)
		}

		authed.GET("/saved-workouts", c.Saved.List)
		authed.GET("/saved-workouts/ids", c.Saved.ListIDs)
		authed.POST("/saved-workouts", c.Saved.Save)
		authed.DELETE("/saved-workouts/:workout_id", c.Saved.Unsave)

		if c.Upload != nil {
			authed.POST("/upload", c.Upload.Upload)
		}

		if c.GymCommerce != nil {
			gc := c.GymCommerce
			authed.POST("/gym-memberships/checkout/vnpay", gc.CheckoutMembershipVNPay)
			authed.POST("/gym-memberships/vnpay/confirm", gc.ConfirmMembershipVNPay)
			authed.POST("/gym-memberships/checkout/stripe", gc.CheckoutMembershipStripe)
			authed.POST("/gym-memberships/stripe/confirm", gc.ConfirmMembershipStripe)
			authed.GET("/gym-memberships/me", gc.MyMembership)
			authed.GET("/gym-memberships/check-in-token", gc.IssueCheckInToken)

			authed.GET("/class-sessions/upcoming", gc.UpcomingClassSessions)
			authed.POST("/class-sessions/:id/book", gc.BookClassSession)
			authed.DELETE("/class-bookings/:id", gc.CancelClassBooking)
			authed.GET("/class-bookings/me", gc.MyClassBookings)

			authed.GET("/pt-packages", gc.ListMyPTPackages)
			authed.POST("/pt-packages", gc.CreatePTPackage)
			authed.GET("/pt-packages/mine", gc.ListPurchasedPTPackages)
			authed.GET("/pt-packages/sold", gc.ListSoldPTPackages)
			authed.POST("/pt-packages/checkout/vnpay", gc.CheckoutPTPackageVNPay)
			authed.POST("/pt-packages/vnpay/confirm", gc.ConfirmPTPackageVNPay)
			authed.POST("/pt-packages/checkout/stripe", gc.CheckoutPTPackageStripe)
			authed.POST("/pt-packages/stripe/confirm", gc.ConfirmPTPackageStripe)
			authed.PUT("/pt-packages/:id", gc.UpdatePTPackage)
			authed.DELETE("/pt-packages/:id", gc.DeletePTPackage)
			authed.GET("/user-pt-packages/:id", gc.GetUserPTPackage)
			authed.GET("/user-pt-packages/:id/sessions", gc.ListPTSessions)
			authed.POST("/user-pt-packages/:id/sessions", gc.LogPTSession)
			authed.GET("/user-pt-packages/:id/messages", gc.ListChatMessages)
			authed.POST("/user-pt-packages/:id/messages", gc.SendChatMessage)
			authed.GET("/user-pt-packages/:id/session-offers", gc.ListSessionOffers)
			authed.POST("/user-pt-packages/:id/session-offers", gc.CreateSessionOffer)
			authed.POST("/user-pt-packages/:id/session-offers/log-direct", gc.LogSessionDirect)
			authed.POST("/user-pt-packages/:id/session-offers/:offerId/accept", gc.AcceptSessionOffer)
			authed.POST("/user-pt-packages/:id/session-offers/:offerId/decline", gc.DeclineSessionOffer)
			authed.POST("/user-pt-packages/:id/session-offers/:offerId/cancel", gc.CancelSessionOffer)
			authed.POST("/user-pt-packages/:id/session-offers/:offerId/reschedule", gc.RescheduleSessionOffer)
			authed.POST("/user-pt-packages/:id/session-offers/:offerId/complete", gc.CompleteSessionOffer)
			authed.POST("/user-pt-packages/:id/session-offers/:offerId/confirm", gc.ConfirmSessionOffer)
			authed.POST("/user-pt-packages/:id/session-offers/:offerId/reject-proof", gc.RejectSessionOfferProof)
			authed.POST("/user-pt-packages/:id/session-offers/:offerId/no-show", gc.MarkSessionNoShow)

			authed.GET("/pt-booking/settings", gc.GetMyBookingSettings)
			authed.PUT("/pt-booking/settings", gc.UpdateMyBookingSettings)
			authed.GET("/pt-booking/working-hours", gc.GetMyWorkingHours)
			authed.PUT("/pt-booking/working-hours", gc.SetMyWorkingHours)
			authed.GET("/pt-booking/blocked-slots", gc.ListMyBlockedSlots)
			authed.POST("/pt-booking/blocked-slots", gc.BlockMySlot)
			authed.DELETE("/pt-booking/blocked-slots/:id", gc.UnblockMySlot)
			authed.POST("/pt-booking/book-slot", gc.BookSlot)
			authed.POST("/pt-booking/recurring", gc.CreateRecurringBooking)
			authed.GET("/pt-booking/recurring", gc.ListMyRecurringBookings)
			authed.POST("/pt-booking/recurring/:id/cancel", gc.CancelRecurringBooking)
			authed.POST("/pt-funnel/touch", gc.TouchPTFunnel)
			authed.POST("/pt-reviews", gc.CreatePTReview)

			authed.GET("/pt-earnings/me", gc.MyPTEarnings)
			authed.GET("/pt-earnings/today", gc.MyTodayActivity)
			authed.POST("/pt-earnings/mark-seen", gc.MarkStudentsSeen)
		}

		if c.ContentShare != nil {
			cs := c.ContentShare
			authed.POST("/content-shares", cs.Share)
			authed.GET("/content-shares", cs.ListRecipients)
			authed.GET("/content-shares/students", cs.ListStudents)
			authed.GET("/content-shares/mine", cs.ListMine)
			authed.DELETE("/content-shares/:content_type/:content_id/:recipient_user_id", cs.Unshare)
		}

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
