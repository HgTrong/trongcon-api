package gym

import (
	"errors"
	"net/http"
	"strconv"

	gymv1 "trongcon-api/api/gym/v1"
	"trongcon-api/api/swagger"
	"trongcon-api/internal/service"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	svc service.GymService
}

func NewController(svc service.GymService) *Controller {
	return &Controller{svc: svc}
}

func writeErr(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrBranchNotFound),
		errors.Is(err, service.ErrTrainerNotFound),
		errors.Is(err, service.ErrUserNotFound):
		ctx.JSON(http.StatusNotFound, swagger.ErrBody{Error: err.Error()})
	case errors.Is(err, service.ErrTrainerExists):
		ctx.JSON(http.StatusConflict, swagger.ErrBody{Error: err.Error()})
	default:
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
	}
}

func parseID(ctx *gin.Context) (uint, error) {
	v, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	return uint(v), err
}

func (c *Controller) CreateBranch(ctx *gin.Context) {
	var req gymv1.CreateBranchReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.CreateBranch(ctx.Request.Context(), &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, res)
}

func (c *Controller) ListBranches(ctx *gin.Context) {
	var req gymv1.ListBranchReq
	_ = ctx.ShouldBindQuery(&req)
	res, err := c.svc.ListBranches(ctx.Request.Context(), &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) ListBranchesPublic(ctx *gin.Context) {
	var req gymv1.ListBranchReq
	_ = ctx.ShouldBindQuery(&req)
	res, err := c.svc.ListBranchesPublic(ctx.Request.Context(), &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) GetBranch(ctx *gin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	res, err := c.svc.GetBranch(ctx.Request.Context(), id)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) GetBranchPublic(ctx *gin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	res, err := c.svc.GetBranchPublic(ctx.Request.Context(), id)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) UpdateBranch(ctx *gin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	var req gymv1.UpdateBranchReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.UpdateBranch(ctx.Request.Context(), id, &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) DeleteBranch(ctx *gin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	if err := c.svc.DeleteBranch(ctx.Request.Context(), id); err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gymv1.DeleteRes{Status: "ok"})
}

func (c *Controller) CreateTrainer(ctx *gin.Context) {
	var req gymv1.CreateTrainerReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.CreateTrainer(ctx.Request.Context(), &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, res)
}

func (c *Controller) ListTrainers(ctx *gin.Context) {
	var req gymv1.ListTrainerReq
	_ = ctx.ShouldBindQuery(&req)
	res, err := c.svc.ListTrainers(ctx.Request.Context(), &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) ListTrainersPublic(ctx *gin.Context) {
	var req gymv1.ListTrainerReq
	_ = ctx.ShouldBindQuery(&req)
	res, err := c.svc.ListTrainersPublic(ctx.Request.Context(), &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) GetTrainer(ctx *gin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	res, err := c.svc.GetTrainer(ctx.Request.Context(), id)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) GetTrainerPublic(ctx *gin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	res, err := c.svc.GetTrainerPublic(ctx.Request.Context(), id)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) UpdateTrainer(ctx *gin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	var req gymv1.UpdateTrainerReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.UpdateTrainer(ctx.Request.Context(), id, &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) DeleteTrainer(ctx *gin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	if err := c.svc.DeleteTrainer(ctx.Request.Context(), id); err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gymv1.DeleteRes{Status: "ok"})
}

func (c *Controller) ListTrainerWorkouts(ctx *gin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	res, err := c.svc.ListTrainerWorkouts(ctx.Request.Context(), id)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) ListTrainerRoutines(ctx *gin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	res, err := c.svc.ListTrainerRoutines(ctx.Request.Context(), id)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) ListTrainerMealPlans(ctx *gin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	res, err := c.svc.ListTrainerMealPlans(ctx.Request.Context(), id)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}
