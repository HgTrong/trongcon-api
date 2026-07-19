package ai

import (
	"errors"
	"net/http"
	"strconv"

	aiv1 "trongcon-api/api/ai/v1"
	"trongcon-api/api/swagger"
	"trongcon-api/internal/http/middleware"
	"trongcon-api/internal/service"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	ai       service.AICoachService
	mealPlan service.MealPlanService
}

func NewController(ai service.AICoachService, mealPlan service.MealPlanService) *Controller {
	return &Controller{ai: ai, mealPlan: mealPlan}
}

func writeErr(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrAIUnavailable):
		ctx.JSON(http.StatusServiceUnavailable, swagger.ErrBody{Error: err.Error()})
	case errors.Is(err, service.ErrMealPlanNotFound),
		errors.Is(err, service.ErrAIChatThreadNotFound),
		errors.Is(err, service.ErrFoodNotFound),
		errors.Is(err, service.ErrExerciseNotFound):
		ctx.JSON(http.StatusNotFound, swagger.ErrBody{Error: err.Error()})
	case errors.Is(err, service.ErrForbiddenMealPlan):
		ctx.JSON(http.StatusForbidden, swagger.ErrBody{Error: err.Error()})
	default:
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
	}
}

func (c *Controller) GenerateMealPlan(ctx *gin.Context) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "unauthorized"})
		return
	}
	var req aiv1.GenerateMealPlanReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	if req.Diet == "" {
		req.Diet = "none"
	}
	res, err := c.ai.GenerateMealPlan(ctx.Request.Context(), userID, &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) GenerateRoutine(ctx *gin.Context) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "unauthorized"})
		return
	}
	var req aiv1.GenerateRoutineReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.ai.GenerateRoutine(ctx.Request.Context(), userID, &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) Chat(ctx *gin.Context) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "unauthorized"})
		return
	}
	var req aiv1.ChatReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.ai.Chat(ctx.Request.Context(), userID, &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) ListThreads(ctx *gin.Context) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "unauthorized"})
		return
	}
	res, err := c.ai.ListThreads(ctx.Request.Context(), userID)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) ListMessages(ctx *gin.Context) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "unauthorized"})
		return
	}
	id, err := strconv.ParseUint(ctx.Param("thread_id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid thread_id"})
		return
	}
	res, err := c.ai.ListMessages(ctx.Request.Context(), userID, uint(id))
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) ListMyMealPlans(ctx *gin.Context) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "unauthorized"})
		return
	}
	var req aiv1.ListMyMealPlansReq
	_ = ctx.ShouldBindQuery(&req)
	res, err := c.mealPlan.ListForUser(ctx.Request.Context(), userID, req.Page, req.Limit, req.Q)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) GetMyMealPlan(ctx *gin.Context) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "unauthorized"})
		return
	}
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	res, err := c.mealPlan.GetForUser(ctx.Request.Context(), userID, uint(id))
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) CreateMyMealPlan(ctx *gin.Context) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "unauthorized"})
		return
	}
	var req aiv1.CreateMyMealPlanReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.mealPlan.CreateForUser(ctx.Request.Context(), userID, req.Title, req.Description, req.IsPublic, req.Meals)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, res)
}

func (c *Controller) UpdateMyMealPlan(ctx *gin.Context) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "unauthorized"})
		return
	}
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	var req aiv1.UpdateMyMealPlanReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.mealPlan.UpdateForUser(ctx.Request.Context(), userID, uint(id), req.Title, req.Description, req.IsPublic, req.Meals)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) DeleteMyMealPlan(ctx *gin.Context) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "unauthorized"})
		return
	}
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	if err := c.mealPlan.DeleteForUser(ctx.Request.Context(), userID, uint(id)); err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}
