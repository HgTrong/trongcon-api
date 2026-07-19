package training_enrollment

import (
	"errors"
	"net/http"
	"strconv"

	"trongcon-api/api/swagger"
	enrollv1 "trongcon-api/api/training_enrollment/v1"
	"trongcon-api/internal/http/middleware"
	"trongcon-api/internal/service"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	svc service.TrainingEnrollmentService
}

func NewController(svc service.TrainingEnrollmentService) *Controller {
	return &Controller{svc: svc}
}

func writeErr(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrEnrollmentNotFound),
		errors.Is(err, service.ErrSlotNotFound),
		errors.Is(err, service.ErrRoutineNotFound),
		errors.Is(err, service.ErrWorkoutNotFound):
		ctx.JSON(http.StatusNotFound, swagger.ErrBody{Error: err.Error()})
	case errors.Is(err, service.ErrForbiddenRoutine),
		errors.Is(err, service.ErrForbiddenWorkout):
		ctx.JSON(http.StatusForbidden, swagger.ErrBody{Error: err.Error()})
	case errors.Is(err, service.ErrSlotAlreadyStarted):
		ctx.JSON(http.StatusConflict, swagger.ErrBody{Error: err.Error()})
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
	var req enrollv1.CreateEnrollmentReq
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
	var req enrollv1.ListEnrollmentsReq
	_ = ctx.ShouldBindQuery(&req)
	res, err := c.svc.List(ctx.Request.Context(), userID, &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) ThisWeek(ctx *gin.Context) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "unauthorized"})
		return
	}
	res, err := c.svc.ThisWeek(ctx.Request.Context(), userID)
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

func (c *Controller) Cancel(ctx *gin.Context) {
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
	res, err := c.svc.Cancel(ctx.Request.Context(), userID, uint(id))
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) Compare(ctx *gin.Context) {
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
	res, err := c.svc.Compare(ctx.Request.Context(), userID, uint(id))
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) StartSlot(ctx *gin.Context) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "unauthorized"})
		return
	}
	slotID, err := strconv.ParseUint(ctx.Param("slot_id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid slot_id"})
		return
	}
	res, err := c.svc.StartSlot(ctx.Request.Context(), userID, uint(slotID))
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, res)
}
