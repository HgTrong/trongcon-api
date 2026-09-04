package revenue

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"trongcon-api/internal/service"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	svc service.RevenueService
}

func NewController(svc service.RevenueService) *Controller {
	return &Controller{svc: svc}
}

// parseDateRange reads ?from=YYYY-MM-DD&to=YYYY-MM-DD (both required, both UTC
// calendar days, `to` inclusive) — missing/invalid input means "no range"
// (nil, nil), which callers treat as their own default (usually "today").
func parseDateRange(ctx *gin.Context) (*time.Time, *time.Time) {
	fromStr := strings.TrimSpace(ctx.Query("from"))
	toStr := strings.TrimSpace(ctx.Query("to"))
	if fromStr == "" || toStr == "" {
		return nil, nil
	}
	from, err1 := time.Parse("2006-01-02", fromStr)
	to, err2 := time.Parse("2006-01-02", toStr)
	if err1 != nil || err2 != nil {
		return nil, nil
	}
	toExclusive := to.AddDate(0, 0, 1)
	return &from, &toExclusive
}

func (c *Controller) Overview(ctx *gin.Context) {
	limit := 10
	if v := ctx.Query("leaderboard_limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	from, to := parseDateRange(ctx)
	res, err := c.svc.Overview(ctx.Request.Context(), limit, from, to)
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
	from, to := parseDateRange(ctx)
	res, err := c.svc.ListPayments(ctx.Request.Context(), page, limit, source, from, to)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) PTLeaderboard(ctx *gin.Context) {
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	sortBy := ctx.DefaultQuery("sort_by", "pt_share")
	from, to := parseDateRange(ctx)
	res, err := c.svc.PTLeaderboard(ctx.Request.Context(), limit, sortBy, from, to)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) TodayActivity(ctx *gin.Context) {
	res, err := c.svc.TodayActivity(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, res)
}
