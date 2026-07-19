package subscription_plan

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"trongcon-api/api/swagger"
	planv1 "trongcon-api/api/subscription_plan/v1"
	"trongcon-api/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Controller struct {
	svc service.SubscriptionPlanService
}

func NewController(svc service.SubscriptionPlanService) *Controller {
	return &Controller{svc: svc}
}

func (c *Controller) Create(ctx *gin.Context) {
	var req planv1.CreateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.Create(ctx.Request.Context(), &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, res)
}

func (c *Controller) List(ctx *gin.Context) {
	var req planv1.ListReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.List(ctx.Request.Context(), &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) ListPublic(ctx *gin.Context) {
	var req planv1.ListReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.ListPublic(ctx.Request.Context(), &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) GetByID(ctx *gin.Context) {
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	res, err := c.svc.GetByID(ctx.Request.Context(), id)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) Update(ctx *gin.Context) {
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	var req planv1.UpdateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.Update(ctx.Request.Context(), id, &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) Delete(ctx *gin.Context) {
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	if err := c.svc.Delete(ctx.Request.Context(), id); err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, planv1.DeleteRes{Status: "ok"})
}

func parseUintParam(ctx *gin.Context, name string) (uint, error) {
	s := ctx.Param(name)
	n, err := strconv.ParseUint(s, 10, 64)
	return uint(n), err
}

func writeErr(ctx *gin.Context, err error) {
	msg := err.Error()
	status := http.StatusBadRequest
	if errors.Is(err, gorm.ErrRecordNotFound) || strings.Contains(msg, "not found") {
		status = http.StatusNotFound
	}
	ctx.JSON(status, swagger.ErrBody{Error: msg})
}
