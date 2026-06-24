package tools

import (
	"errors"
	"net/http"

	"trongcon-api/api/swagger"
	toolsv1 "trongcon-api/api/tools/v1"
	"trongcon-api/internal/service"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	tdeeSvc    service.TDEEService
	macroSvc   service.MacroService
	oneRepMaxSvc service.OneRepMaxService
}

func NewController(tdeeSvc service.TDEEService, macroSvc service.MacroService, oneRepMaxSvc service.OneRepMaxService) *Controller {
	return &Controller{tdeeSvc: tdeeSvc, macroSvc: macroSvc, oneRepMaxSvc: oneRepMaxSvc}
}

func (c *Controller) CalculateTDEE(ctx *gin.Context) {
	var req toolsv1.CalculateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.tdeeSvc.Calculate(&req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) CalculateOneRepMax(ctx *gin.Context) {
	var req toolsv1.OneRepMaxCalculateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.oneRepMaxSvc.Calculate(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) CalculateMacro(ctx *gin.Context) {
	var req toolsv1.MacroCalculateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.macroSvc.Calculate(&req)
	if err != nil {
		writeMacroErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func writeErr(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidActivityLevel),
		errors.Is(err, service.ErrInvalidGoal):
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
	default:
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
	}
}

func writeMacroErr(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidMacroPreset):
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
	default:
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
	}
}
