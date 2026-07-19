package stats

import (
	"net/http"

	"trongcon-api/internal/service"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	svc service.StatsService
}

func NewController(svc service.StatsService) *Controller {
	return &Controller{svc: svc}
}

func (c *Controller) Overview(ctx *gin.Context) {
	res, err := c.svc.Overview(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, res)
}
