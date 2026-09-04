// Package main khởi chạy API TrongCon.
// @title TrongCon API
// @version 1.0
// @description REST API: auth, admin CRUD, user signup/login. Swagger: /swagger/index.html
// @termsOfService http://swagger.io/terms/

// @contact.name TrongCon
// @contact.url https://github.com
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
// @schemes http
package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"trongcon-api/internal/config"
	aictl "trongcon-api/internal/controller/ai"
	articlectl "trongcon-api/internal/controller/article"
	authctl "trongcon-api/internal/controller/auth"
	categoryctl "trongcon-api/internal/controller/category"
	contentsharectl "trongcon-api/internal/controller/content_share"
	emailtemplatectl "trongcon-api/internal/controller/email_template"
	equipmentctl "trongcon-api/internal/controller/equipment"
	exercisectl "trongcon-api/internal/controller/exercise"
	faqctl "trongcon-api/internal/controller/faq"
	foodctl "trongcon-api/internal/controller/food"
	foodlogctl "trongcon-api/internal/controller/food_log"
	gymctl "trongcon-api/internal/controller/gym"
	gymcommercectl "trongcon-api/internal/controller/gym_commerce"
	mealplanctl "trongcon-api/internal/controller/meal_plan"
	musclectl "trongcon-api/internal/controller/muscle"
	mytrainctl "trongcon-api/internal/controller/my_train"
	revenuectl "trongcon-api/internal/controller/revenue"
	routinectl "trongcon-api/internal/controller/routine"
	savedworkoutctl "trongcon-api/internal/controller/saved_workout"
	statsctl "trongcon-api/internal/controller/stats"
	planctl "trongcon-api/internal/controller/subscription_plan"
	toolsctl "trongcon-api/internal/controller/tools"
	enrollctl "trongcon-api/internal/controller/training_enrollment"
	uploadctl "trongcon-api/internal/controller/upload"
	userctl "trongcon-api/internal/controller/user"
	subctl "trongcon-api/internal/controller/user_subscription"
	workoutctl "trongcon-api/internal/controller/workout"
	sessionctl "trongcon-api/internal/controller/workout_session"
	httpserver "trongcon-api/internal/http"
	"trongcon-api/internal/mail"
	oaiclient "trongcon-api/internal/openai"
	"trongcon-api/internal/repository"
	adminrouter "trongcon-api/internal/router/admin"
	publicrouter "trongcon-api/internal/router/public"
	"trongcon-api/internal/service"
	"trongcon-api/internal/storage/postgres"
)

func main() {
	loadDotEnv()

	cfg := config.Load()
	db := postgres.GetDatabaseConnection()
	emailTplRepo := repository.NewEmailTemplateRepository(db.Connection)
	mailSender := mail.NewSender(mail.SMTPConfig{
		Enabled:  cfg.SMTP.Enabled,
		Name:     cfg.SMTP.Name,
		Host:     cfg.SMTP.Host,
		Port:     cfg.SMTP.Port,
		Username: cfg.SMTP.Username,
		Password: cfg.SMTP.Password,
		From:     cfg.SMTP.From,
	})
	if mailSender.Enabled() {
		log.Printf("📧 Mail: AWS SES SMTP enabled (from=%s host=%s)", cfg.SMTP.From, cfg.SMTP.Host)
	} else {
		log.Printf("📧 Mail: SMTP disabled or incomplete — outbound email will fail until SMTP_* is set")
	}
	emailTplSvc := service.NewEmailTemplateService(emailTplRepo, mailSender)
	mailerSvc := service.NewMailerService(emailTplRepo, mailSender)
	userRepo := repository.NewUserRepository(db.Connection)
	roleRepo := repository.NewRoleRepository(db.Connection)
	categoryRepo := repository.NewCategoryRepository(db.Connection)
	articleRepo := repository.NewArticleRepository(db.Connection)
	equipmentRepo := repository.NewEquipmentRepository(db.Connection)
	exerciseRepo := repository.NewExerciseRepository(db.Connection)
	foodRepo := repository.NewFoodRepository(db.Connection)
	mealPlanRepo := repository.NewMealPlanRepository(db.Connection)
	routineRepo := repository.NewRoutineRepository(db.Connection)
	workoutRepo := repository.NewWorkoutRepository(db.Connection)
	savedWorkoutRepo := repository.NewSavedWorkoutRepository(db.Connection)
	foodLogRepo := repository.NewFoodLogRepository(db.Connection)
	sessionRepo := repository.NewWorkoutSessionRepository(db.Connection)
	enrollmentRepo := repository.NewTrainingEnrollmentRepository(db.Connection)
	muscleRepo := repository.NewMuscleRepository(db.Connection)
	aiChatRepo := repository.NewAiChatRepository(db.Connection)
	branchRepo := repository.NewGymBranchRepository(db.Connection)
	trainerRepo := repository.NewTrainerProfileRepository(db.Connection)
	ptStatRepo := repository.NewPTContentStatRepository(db.Connection)
	ptAttrRepo := repository.NewPTAttributionRepository(db.Connection)
	ptReviewRepo := repository.NewPTReviewRepository(db.Connection)
	ptGrowth := service.NewPTGrowthTracker(trainerRepo, ptStatRepo, ptAttrRepo)

	uploadSvc := service.NewUploadService(cfg.S3)
	userSvc := service.NewUserService(userRepo, roleRepo, uploadSvc)
	categorySvc := service.NewCategoryService(categoryRepo)
	articleSvc := service.NewArticleService(articleRepo, categoryRepo, userRepo, trainerRepo, ptGrowth)
	equipmentSvc := service.NewEquipmentService(equipmentRepo)
	exerciseSvc := service.NewExerciseService(exerciseRepo, equipmentRepo, muscleRepo)
	foodSvc := service.NewFoodService(foodRepo)
	contentShareRepo := repository.NewContentShareRepository(db.Connection)
	mealPlanSvc := service.NewMealPlanService(mealPlanRepo, foodRepo, userRepo, trainerRepo, contentShareRepo, ptGrowth)
	routineSvc := service.NewRoutineService(routineRepo, workoutRepo, userRepo, trainerRepo, contentShareRepo, ptGrowth)
	workoutSvc := service.NewWorkoutService(workoutRepo, exerciseRepo, userRepo, trainerRepo, contentShareRepo, ptGrowth)
	savedWorkoutSvc := service.NewSavedWorkoutService(savedWorkoutRepo, workoutRepo, ptGrowth)
	macroSvc := service.NewMacroService()
	foodLogSvc := service.NewFoodLogService(foodLogRepo, foodRepo, macroSvc)
	myTrainSvc := service.NewMyTrainService(workoutRepo, exerciseRepo, routineRepo)
	sessionSvc := service.NewWorkoutSessionService(sessionRepo, workoutRepo, exerciseRepo)
	enrollmentSvc := service.NewTrainingEnrollmentService(enrollmentRepo, routineRepo, workoutRepo, sessionRepo, sessionSvc)
	muscleSvc := service.NewMuscleService(muscleRepo)
	tdeeSvc := service.NewTDEEService()
	oneRepMaxSvc := service.NewOneRepMaxService()
	statsSvc := service.NewStatsService(db.Connection)
	openaiClient := oaiclient.NewClient(cfg.OpenAI)
	aiCoachSvc := service.NewAICoachService(openaiClient, foodRepo, exerciseRepo, muscleRepo, mealPlanSvc, myTrainSvc, aiChatRepo)
	gymSvc := service.NewGymService(branchRepo, trainerRepo, userRepo, roleRepo, workoutRepo, routineRepo, mealPlanRepo, ptGrowth)

	planRepo := repository.NewSubscriptionPlanRepository(db.Connection)
	userSubRepo := repository.NewUserSubscriptionRepository(db.Connection)
	paymentHistoryRepo := repository.NewPaymentHistoryRepository(db.Connection)
	vnpaySvc := service.NewVNPayService(cfg.VNPay)
	if !vnpaySvc.Enabled() {
		log.Printf("VNPay: TMN_CODE/SECRET_KEY missing — VNPay checkout disabled")
	} else {
		log.Printf("VNPay: enabled (return %s)", cfg.VNPay.ReturnURL)
	}
	planSvc := service.NewSubscriptionPlanService(planRepo)
	stripeSvc := service.NewStripeService(cfg.Stripe)
	userSubSvc := service.NewUserSubscriptionService(userSubRepo, planRepo, paymentHistoryRepo, userRepo, vnpaySvc, stripeSvc)

	gymMembershipPlanRepo := repository.NewGymMembershipPlanRepository(db.Connection)
	userGymMembershipRepo := repository.NewUserGymMembershipRepository(db.Connection)
	groupClassRepo := repository.NewGroupClassRepository(db.Connection)
	classSessionRepo := repository.NewClassSessionRepository(db.Connection)
	classBookingRepo := repository.NewClassBookingRepository(db.Connection)
	ptPackageRepo := repository.NewPTPackageRepository(db.Connection)
	userPTPackageRepo := repository.NewUserPTPackageRepository(db.Connection)
	ptSessionLogRepo := repository.NewPTSessionLogRepository(db.Connection)
	ptChatRepo := repository.NewPTPackageChatRepository(db.Connection)
	revenueShareRepo := repository.NewRevenueShareSettingRepository(db.Connection)
	ptEarningRepo := repository.NewPTEarningRepository(db.Connection)
	ptHoursRepo := repository.NewPTWorkingHoursRepository(db.Connection)
	ptBlockedRepo := repository.NewPTBlockedSlotRepository(db.Connection)
	ptRecurringRepo := repository.NewPTRecurringBookingRepository(db.Connection)
	gymCommerceSvc := service.NewGymCommerceService(
		gymMembershipPlanRepo, userGymMembershipRepo, groupClassRepo, classSessionRepo, classBookingRepo,
		ptPackageRepo, userPTPackageRepo, ptSessionLogRepo,
		repository.NewPTSessionOfferRepository(db.Connection),
		ptChatRepo, ptHoursRepo, ptBlockedRepo, ptRecurringRepo,
		ptStatRepo, ptAttrRepo, ptReviewRepo, ptGrowth,
		revenueShareRepo, ptEarningRepo, trainerRepo, userRepo,
		vnpaySvc, stripeSvc,
		cfg.VNPay.MembershipReturnURL, cfg.VNPay.PackageReturnURL,
		cfg.Stripe.MembershipSuccessURL, cfg.Stripe.MembershipCancelURL,
		cfg.Stripe.PackageSuccessURL, cfg.Stripe.PackageCancelURL,
		db.Connection,
	)
	gymCommerceSvc.ConfigureOps(mailerSvc, userSubSvc, cfg.JWTSecret, repository.NewGymCheckInRepository(db.Connection))

	contentShareSvc := service.NewContentShareService(contentShareRepo, workoutRepo, routineRepo, mealPlanRepo, trainerRepo, userPTPackageRepo, userRepo, ptChatRepo)
	contentShareController := contentsharectl.NewController(contentShareSvc)

	// Auto-confirm PT session proofs after 1 day if the student forgets.
	go func() {
		ticker := time.NewTicker(15 * time.Minute)
		defer ticker.Stop()
		for {
			n, err := gymCommerceSvc.AutoConfirmExpiredSessionProofs(context.Background(), 24*time.Hour, 100)
			if err != nil {
				log.Printf("pt auto-confirm: %v", err)
			} else if n > 0 {
				log.Printf("pt auto-confirm: confirmed %d session(s)", n)
			}
			<-ticker.C
		}
	}()

	// Cancel session offers left "pending" for 2+ days — otherwise a forgotten
	// proposal keeps blocking the trainer's slot / the package's session credit.
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			n, err := gymCommerceSvc.ExpireStalePendingOffers(context.Background(), 48*time.Hour)
			if err != nil {
				log.Printf("pt stale-offer cleanup: %v", err)
			} else if n > 0 {
				log.Printf("pt stale-offer cleanup: cancelled %d offer(s)", n)
			}
			<-ticker.C
		}
	}()

	// Actively flip expired gym memberships / PT packages to "expired" instead of
	// relying only on lazy expiry the next time someone reads the list.
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for {
			if err := gymCommerceSvc.RunExpiryHousekeeping(context.Background()); err != nil {
				log.Printf("membership/pt-package expiry housekeeping: %v", err)
			}
			<-ticker.C
		}
	}()

	// Cancel gym-membership / PT-package / Premium subscription orders left "pending"
	// (checkout opened but never paid) for 6+ hours — an abandoned Stripe/VNPay
	// checkout would otherwise stay "pending" forever.
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for {
			n, err := gymCommerceSvc.CancelStalePendingOrders(context.Background(), 6*time.Hour)
			if err != nil {
				log.Printf("stale pending order cleanup: %v", err)
			} else if n > 0 {
				log.Printf("stale pending order cleanup: cancelled %d order(s)", n)
			}
			if userSubSvc != nil {
				if n, err := userSubSvc.CancelStalePending(context.Background(), 6*time.Hour); err != nil {
					log.Printf("stale pending subscription cleanup: %v", err)
				} else if n > 0 {
					log.Printf("stale pending subscription cleanup: cancelled %d subscription(s)", n)
				}
			}
			<-ticker.C
		}
	}()

	// Keep every active standing (recurring) PT booking materialized ~3 weeks
	// ahead into real session offers, and auto-pause any that ran out of
	// session credits or whose package expired.
	go func() {
		ticker := time.NewTicker(6 * time.Hour)
		defer ticker.Stop()
		for {
			n, err := gymCommerceSvc.MaterializeRecurringBookings(context.Background(), 0)
			if err != nil {
				log.Printf("recurring booking materialization: %v", err)
			} else if n > 0 {
				log.Printf("recurring booking materialization: created %d occurrence(s)", n)
			}
			<-ticker.C
		}
	}()

	authSvc := service.NewAuthService(
		userRepo,
		roleRepo,
		repository.NewEmailOTPRepository(db.Connection),
		mailerSvc,
		cfg.JWTSecret,
		cfg.JWTExpiration,
	)

	userController := userctl.NewController(userSvc)
	authController := authctl.NewController(authSvc)
	categoryController := categoryctl.NewController(categorySvc)
	articleController := articlectl.NewController(articleSvc)
	equipmentController := equipmentctl.NewController(equipmentSvc)
	exerciseController := exercisectl.NewController(exerciseSvc)
	foodController := foodctl.NewController(foodSvc)
	mealPlanController := mealplanctl.NewController(mealPlanSvc)
	routineController := routinectl.NewController(routineSvc)
	workoutController := workoutctl.NewController(workoutSvc)
	savedWorkoutController := savedworkoutctl.NewController(savedWorkoutSvc)
	foodLogController := foodlogctl.NewController(foodLogSvc)
	myTrainController := mytrainctl.NewController(myTrainSvc)
	sessionController := sessionctl.NewController(sessionSvc)
	enrollmentController := enrollctl.NewController(enrollmentSvc)
	aiController := aictl.NewController(aiCoachSvc, mealPlanSvc)
	gymController := gymctl.NewController(gymSvc)
	muscleController := musclectl.NewController(muscleSvc)
	toolsController := toolsctl.NewController(tdeeSvc, macroSvc, oneRepMaxSvc)
	uploadController := uploadctl.NewController(uploadSvc)
	statsController := statsctl.NewController(statsSvc)
	revenueSvc := service.NewRevenueService(db.Connection)
	revenueController := revenuectl.NewController(revenueSvc)
	planController := planctl.NewController(planSvc)
	userSubController := subctl.NewController(userSubSvc, cfg.Stripe)
	userSubController.SetGymCommerce(gymCommerceSvc)
	gymCommerceController := gymcommercectl.NewController(gymCommerceSvc)
	emailTemplateController := emailtemplatectl.NewController(emailTplSvc)

	faqRepo := repository.NewFAQRepository(db.Connection)
	faqSvc := service.NewFAQService(faqRepo)
	faqController := faqctl.NewController(faqSvc)

	router := httpserver.NewRouter(cfg, httpserver.Deps{
		Auth:         authController,
		Saved:        savedWorkoutController,
		FoodLog:      foodLogController,
		MyTrain:      myTrainController,
		Sessions:     sessionController,
		Enrollment:   enrollmentController,
		AI:           aiController,
		Subscription: userSubController,
		ContentShare: contentShareController,
		Premium:      userSubSvc,
		Admin: adminrouter.Controllers{
			User:             userController,
			Category:         categoryController,
			Article:          articleController,
			Equipment:        equipmentController,
			Exercise:         exerciseController,
			Food:             foodController,
			MealPlan:         mealPlanController,
			Routine:          routineController,
			Workout:          workoutController,
			Muscle:           muscleController,
			Tools:            toolsController,
			Upload:           uploadController,
			Stats:            statsController,
			Revenue:          revenueController,
			Gym:              gymController,
			SubscriptionPlan: planController,
			UserSubscription: userSubController,
			GymCommerce:      gymCommerceController,
			EmailTemplate:    emailTemplateController,
			FAQ:              faqController,
		},
		Public: publicrouter.Controllers{
			Exercise:         exerciseController,
			Muscle:           muscleController,
			Equipment:        equipmentController,
			Workout:          workoutController,
			Routine:          routineController,
			Food:             foodController,
			MealPlan:         mealPlanController,
			Article:          articleController,
			Category:         categoryController,
			Tools:            toolsController,
			Gym:              gymController,
			SubscriptionPlan: planController,
			GymCommerce:      gymCommerceController,
			FAQ:              faqController,
		},
	})

	log.Printf("API listening on :%s — Swagger: http://localhost:%s/swagger/index.html", cfg.Port, cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}

// loadDotEnv loads .env from cwd or parent dirs (so `go run` from cmd/api still works).
func loadDotEnv() {
	candidates := []string{".env"}
	if wd, err := os.Getwd(); err == nil {
		dir := wd
		for i := 0; i < 4; i++ {
			candidates = append(candidates, filepath.Join(dir, ".env"))
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	for _, p := range candidates {
		if err := godotenv.Load(p); err == nil {
			return
		}
	}
}
