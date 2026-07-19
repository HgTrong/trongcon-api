package saved_workout

import (
	"errors"
	"net/http"
	"strconv"

	"trongcon-api/api/swagger"
	savedv1 "trongcon-api/api/saved_workout/v1"
	"trongcon-api/internal/http/middleware"
	"trongcon-api/internal/service"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	svc service.SavedWorkoutService
}

func NewController(svc service.SavedWorkoutService) *Controller {
	return &Controller{svc: svc}
}

func (c *Controller) Save(ctx *gin.Context) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "unauthorized"})
		return
	}
	var req savedv1.SaveReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.Save(ctx.Request.Context(), userID, req.WorkoutID)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) Unsave(ctx *gin.Context) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "unauthorized"})
		return
	}
	workoutID, err := parseUintParam(ctx, "workout_id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid workout_id"})
		return
	}
	if err := c.svc.Unsave(ctx.Request.Context(), userID, workoutID); err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, savedv1.DeleteRes{Status: "ok"})
}

func (c *Controller) ListIDs(ctx *gin.Context) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "unauthorized"})
		return
	}
	res, err := c.svc.ListIDs(ctx.Request.Context(), userID)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) List(ctx *gin.Context) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "unauthorized"})
		return
	}
	var req savedv1.ListReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.List(ctx.Request.Context(), userID, &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func parseUintParam(ctx *gin.Context, name string) (uint, error) {
	s := ctx.Param(name)
	if s == "" {
		return 0, strconv.ErrRange
	}
	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}

func writeErr(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrWorkoutNotFound):
		ctx.JSON(http.StatusNotFound, swagger.ErrBody{Error: err.Error()})
	case errors.Is(err, service.ErrSavedWorkoutNotFound):
		ctx.JSON(http.StatusNotFound, swagger.ErrBody{Error: err.Error()})
	default:
		ctx.JSON(http.StatusInternalServerError, swagger.ErrBody{Error: err.Error()})
	}
}
