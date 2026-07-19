package workout_session

import (
	"errors"
	"net/http"
	"strconv"

	"trongcon-api/api/swagger"
	sessionv1 "trongcon-api/api/workout_session/v1"
	"trongcon-api/internal/http/middleware"
	"trongcon-api/internal/service"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	svc service.WorkoutSessionService
}

func NewController(svc service.WorkoutSessionService) *Controller {
	return &Controller{svc: svc}
}

func writeErr(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrSessionNotFound),
		errors.Is(err, service.ErrSetNotFound),
		errors.Is(err, service.ErrWorkoutNotFound),
		errors.Is(err, service.ErrExerciseNotFound):
		ctx.JSON(http.StatusNotFound, swagger.ErrBody{Error: err.Error()})
	case errors.Is(err, service.ErrForbiddenWorkout):
		ctx.JSON(http.StatusForbidden, swagger.ErrBody{Error: err.Error()})
	default:
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
	}
}

func (c *Controller) Create(ctx *gin.Context) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "unauthorized"})
		return
	}
	var req sessionv1.CreateSessionReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.Create(ctx.Request.Context(), userID, &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, res)
}

func (c *Controller) List(ctx *gin.Context) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "unauthorized"})
		return
	}
	var req sessionv1.ListSessionsReq
	_ = ctx.ShouldBindQuery(&req)
	res, err := c.svc.List(ctx.Request.Context(), userID, &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) Get(ctx *gin.Context) {
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
	res, err := c.svc.Get(ctx.Request.Context(), userID, uint(id))
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) UpdateSet(ctx *gin.Context) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "unauthorized"})
		return
	}
	setID, err := strconv.ParseUint(ctx.Param("set_id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid set_id"})
		return
	}
	var req sessionv1.UpdateSetReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.UpdateSet(ctx.Request.Context(), userID, uint(setID), &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) AddItem(ctx *gin.Context) {
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
	var req sessionv1.AddSessionItemReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.AddItem(ctx.Request.Context(), userID, uint(id), &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, res)
}

func (c *Controller) Complete(ctx *gin.Context) {
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
	res, err := c.svc.Complete(ctx.Request.Context(), userID, uint(id))
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) Abandon(ctx *gin.Context) {
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
	res, err := c.svc.Abandon(ctx.Request.Context(), userID, uint(id))
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) ExercisePerformance(ctx *gin.Context) {
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
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "12"))
	res, err := c.svc.ExercisePerformance(ctx.Request.Context(), userID, uint(id), limit)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}
