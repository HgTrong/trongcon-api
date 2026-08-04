package revenue

import (
	"net/http"
	"strconv"

	"trongcon-api/internal/service"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	svc service.RevenueService
}

func NewController(svc service.RevenueService) *Controller {
	return &Controller{svc: svc}
}

func (c *Controller) Overview(ctx *gin.Context) {
	limit := 10
	if v := ctx.Query("leaderboard_limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	res, err := c.svc.Overview(ctx.Request.Context(), limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) Payments(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	source := ctx.Query("source")
	res, err := c.svc.ListPayments(ctx.Request.Context(), page, limit, source)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) PTLeaderboard(ctx *gin.Context) {
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	sortBy := ctx.DefaultQuery("sort_by", "pt_share")
	res, err := c.svc.PTLeaderboard(ctx.Request.Context(), limit, sortBy)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, res)
}
