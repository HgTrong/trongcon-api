package exercise

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	exercisev1 "trongcon-api/api/exercise/v1"
	"trongcon-api/api/swagger"
	"trongcon-api/internal/service"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	svc service.ExerciseService
}

func NewController(svc service.ExerciseService) *Controller {
	return &Controller{svc: svc}
}

func (c *Controller) Create(ctx *gin.Context) {
	var req exercisev1.CreateReq
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
	var req exercisev1.ListReq
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
	var req exercisev1.ListReq
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

func (c *Controller) GetBySlug(ctx *gin.Context) {
	slug := strings.TrimSpace(ctx.Param("slug"))
	if slug == "" {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid slug"})
		return
	}
	res, err := c.svc.GetBySlug(ctx.Request.Context(), slug)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) IncrementViews(ctx *gin.Context) {
	slug := strings.TrimSpace(ctx.Param("slug"))
	if slug == "" {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid slug"})
		return
	}
	res, err := c.svc.IncrementViews(ctx.Request.Context(), slug)
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

func (c *Controller) GetByIDPublic(ctx *gin.Context) {
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	res, err := c.svc.GetByIDPublic(ctx.Request.Context(), id)
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
	var req exercisev1.UpdateReq
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
	ctx.JSON(http.StatusOK, exercisev1.DeleteRes{Status: "ok"})
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
	case errors.Is(err, service.ErrExerciseNotFound):
		ctx.JSON(http.StatusNotFound, swagger.ErrBody{Error: err.Error()})
	case errors.Is(err, service.ErrEquipmentNotFound),
		errors.Is(err, service.ErrMuscleNotFound):
		ctx.JSON(http.StatusNotFound, swagger.ErrBody{Error: err.Error()})
	default:
		ctx.JSON(http.StatusInternalServerError, swagger.ErrBody{Error: err.Error()})
	}
}
