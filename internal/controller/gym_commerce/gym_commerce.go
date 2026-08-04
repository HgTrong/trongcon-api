package gym_commerce

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	gcv1 "trongcon-api/api/gym_commerce/v1"
	"trongcon-api/api/swagger"
	"trongcon-api/internal/http/middleware"
	"trongcon-api/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Controller struct {
	svc service.GymCommerceService
}

func NewController(svc service.GymCommerceService) *Controller {
	return &Controller{svc: svc}
}

// —— helpers ——

func parseUintParam(ctx *gin.Context, name string) (uint, error) {
	n, err := strconv.ParseUint(ctx.Param(name), 10, 64)
	return uint(n), err
}

func queryPageLimit(ctx *gin.Context) (int, int) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	return page, limit
}

func queryUintPtr(ctx *gin.Context, name string) *uint {
	v := strings.TrimSpace(ctx.Query(name))
	if v == "" {
		return nil
	}
	n, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return nil
	}
	u := uint(n)
	return &u
}

func queryBoolPtr(ctx *gin.Context, name string) *bool {
	v := strings.TrimSpace(ctx.Query(name))
	if v == "" {
		return nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return nil
	}
	return &b
}

func writeErr(ctx *gin.Context, err error) {
	msg := err.Error()
	status := http.StatusBadRequest
	if errors.Is(err, gorm.ErrRecordNotFound) || strings.Contains(msg, "not found") {
		status = http.StatusNotFound
	}
	if strings.Contains(msg, "unauthorized") {
		status = http.StatusUnauthorized
	}
	ctx.JSON(status, swagger.ErrBody{Error: msg})
}

func requireUserID(ctx *gin.Context) (uint, bool) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok || userID == 0 {
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "unauthorized"})
		return 0, false
	}
	return userID, true
}

// ============================== Admin: membership plans ==============================

func (c *Controller) CreatePlan(ctx *gin.Context) {
	var req gcv1.MembershipPlanReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.CreatePlan(ctx.Request.Context(), &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, res)
}

func (c *Controller) ListPlans(ctx *gin.Context) {
	page, limit := queryPageLimit(ctx)
	q := ctx.Query("q")
	active := queryBoolPtr(ctx, "active")
	res, err := c.svc.ListPlansAdmin(ctx.Request.Context(), page, limit, q, active)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) UpdatePlan(ctx *gin.Context) {
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	var req gcv1.MembershipPlanReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.UpdatePlan(ctx.Request.Context(), id, &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) DeletePlan(ctx *gin.Context) {
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	if err := c.svc.DeletePlan(ctx.Request.Context(), id); err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ListPlansPublic serves the public catalog of active membership plans.
func (c *Controller) ListPlansPublic(ctx *gin.Context) {
	page, limit := queryPageLimit(ctx)
	res, err := c.svc.ListPublicPlans(ctx.Request.Context(), page, limit)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

// ListHighlightedPlansPublic returns active plans flagged for the marketing home.
func (c *Controller) ListHighlightedPlansPublic(ctx *gin.Context) {
	page, limit := queryPageLimit(ctx)
	res, err := c.svc.ListHighlightedPlans(ctx.Request.Context(), page, limit)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) SetPlanHighlight(ctx *gin.Context) {
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	var req gcv1.HighlightPlanReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.SetPlanHighlight(ctx.Request.Context(), id, req.IsHighlighted)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

// ============================== Admin: user gym memberships ==============================

func (c *Controller) ListUserGymMemberships(ctx *gin.Context) {
	page, limit := queryPageLimit(ctx)
	status := ctx.Query("status")
	var userID uint
	if v := queryUintPtr(ctx, "user_id"); v != nil {
		userID = *v
	}
	res, err := c.svc.ListUserGymMembershipsAdmin(ctx.Request.Context(), page, limit, status, userID)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) AdminActivateMembership(ctx *gin.Context) {
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	res, err := c.svc.AdminActivateMembership(ctx.Request.Context(), id)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) AdminCancelMembership(ctx *gin.Context) {
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	res, err := c.svc.AdminCancelMembership(ctx.Request.Context(), id)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

// ============================== Admin: group classes ==============================

func (c *Controller) CreateGroupClass(ctx *gin.Context) {
	var req gcv1.GroupClassReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.CreateGroupClass(ctx.Request.Context(), &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, res)
}

func (c *Controller) ListGroupClasses(ctx *gin.Context) {
	page, limit := queryPageLimit(ctx)
	q := ctx.Query("q")
	branchID := queryUintPtr(ctx, "branch_id")
	res, err := c.svc.ListGroupClasses(ctx.Request.Context(), page, limit, q, branchID)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) UpdateGroupClass(ctx *gin.Context) {
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	var req gcv1.GroupClassReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.UpdateGroupClass(ctx.Request.Context(), id, &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) DeleteGroupClass(ctx *gin.Context) {
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	if err := c.svc.DeleteGroupClass(ctx.Request.Context(), id); err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ============================== Admin: class sessions ==============================

func (c *Controller) CreateClassSession(ctx *gin.Context) {
	var req gcv1.ClassSessionReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.CreateClassSession(ctx.Request.Context(), &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, res)
}

func (c *Controller) ListClassSessions(ctx *gin.Context) {
	page, limit := queryPageLimit(ctx)
	groupClassID := queryUintPtr(ctx, "group_class_id")
	res, err := c.svc.ListClassSessionsAdmin(ctx.Request.Context(), page, limit, groupClassID)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) DeleteClassSession(ctx *gin.Context) {
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	if err := c.svc.DeleteClassSession(ctx.Request.Context(), id); err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ListUpcomingClassSessionsPublic serves upcoming sessions for public/user browsing.
func (c *Controller) ListUpcomingClassSessionsPublic(ctx *gin.Context) {
	page, limit := queryPageLimit(ctx)
	res, err := c.svc.ListPublicUpcomingSessions(ctx.Request.Context(), page, limit)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

// ============================== Admin: revenue share ==============================

func (c *Controller) GetRevenueShare(ctx *gin.Context) {
	res, err := c.svc.GetRevenueShare(ctx.Request.Context())
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) UpdateRevenueShare(ctx *gin.Context) {
	var req gcv1.RevenueShareReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.UpdateRevenueShare(ctx.Request.Context(), &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

// ============================== Admin: PT earnings / packages catalog ==============================

func (c *Controller) ListPTEarnings(ctx *gin.Context) {
	page, limit := queryPageLimit(ctx)
	var trainerID uint
	if v := queryUintPtr(ctx, "trainer_id"); v != nil {
		trainerID = *v
	}
	res, err := c.svc.ListPTEarningsAdmin(ctx.Request.Context(), page, limit, trainerID)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) SetPTEarningPaidOut(ctx *gin.Context) {
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	var req struct {
		Paid bool `json:"paid_out"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	if err := c.svc.AdminSetPTEarningPaidOut(ctx.Request.Context(), id, req.Paid); err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

func (c *Controller) ListPTPackagesAdmin(ctx *gin.Context) {
	page, limit := queryPageLimit(ctx)
	q := ctx.Query("q")
	var trainerID uint
	if v := queryUintPtr(ctx, "trainer_id"); v != nil {
		trainerID = *v
	}
	res, err := c.svc.ListPTPackagesAdmin(ctx.Request.Context(), page, limit, q, trainerID)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) ListUserPTPackagesAdmin(ctx *gin.Context) {
	page, limit := queryPageLimit(ctx)
	status := ctx.Query("status")
	var trainerID, userID uint
	if v := queryUintPtr(ctx, "trainer_id"); v != nil {
		trainerID = *v
	}
	if v := queryUintPtr(ctx, "user_id"); v != nil {
		userID = *v
	}
	res, err := c.svc.ListUserPTPackagesAdmin(ctx.Request.Context(), page, limit, status, trainerID, userID)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) ListPTSessionsAdmin(ctx *gin.Context) {
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	res, err := c.svc.ListPTSessionsAdmin(ctx.Request.Context(), id)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

// ============================== User: gym membership checkout ==============================

func (c *Controller) CheckoutMembershipVNPay(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	var req gcv1.CheckoutReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.CheckoutMembershipVNPay(ctx.Request.Context(), userID, req.PlanID, ctx.ClientIP())
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) ConfirmMembershipVNPay(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	var req gcv1.ConfirmVNPayReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.ConfirmMembershipVNPay(ctx.Request.Context(), userID, req.Params)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) CheckoutMembershipStripe(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	var req gcv1.CheckoutReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.CheckoutMembershipStripe(ctx.Request.Context(), userID, req.PlanID)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) ConfirmMembershipStripe(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	var req gcv1.ConfirmStripeReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.ConfirmMembershipStripe(ctx.Request.Context(), userID, req.SessionID)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) MyMembership(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	res, err := c.svc.MyMembership(ctx.Request.Context(), userID)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) IssueCheckInToken(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	res, err := c.svc.IssueCheckInToken(ctx.Request.Context(), userID)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) VerifyCheckIn(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	var req gcv1.VerifyCheckInReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.VerifyCheckIn(ctx.Request.Context(), userID, req.Token, req.BranchID, req.Note)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) ListRecentCheckIns(ctx *gin.Context) {
	res, err := c.svc.ListRecentCheckIns(ctx.Request.Context(), 50)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

// ============================== User: group classes ==============================

func (c *Controller) UpcomingClassSessions(ctx *gin.Context) {
	page, limit := queryPageLimit(ctx)
	res, err := c.svc.ListUpcomingClassSessions(ctx.Request.Context(), page, limit)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) BookClassSession(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	sessionID, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	res, err := c.svc.BookClassSession(ctx.Request.Context(), userID, sessionID)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, res)
}

func (c *Controller) CancelClassBooking(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	bookingID, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	if err := c.svc.CancelClassBooking(ctx.Request.Context(), userID, bookingID); err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (c *Controller) MyClassBookings(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	res, err := c.svc.MyClassBookings(ctx.Request.Context(), userID)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

// ============================== User: PT packages (owned) ==============================

func (c *Controller) CreatePTPackage(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	var req gcv1.PTPackageReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.CreatePTPackage(ctx.Request.Context(), userID, &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, res)
}

func (c *Controller) ListMyPTPackages(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	page, limit := queryPageLimit(ctx)
	res, err := c.svc.ListMyPTPackages(ctx.Request.Context(), userID, page, limit)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) UpdatePTPackage(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	var req gcv1.PTPackageReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.UpdatePTPackage(ctx.Request.Context(), userID, id, &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) DeletePTPackage(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	if err := c.svc.DeletePTPackage(ctx.Request.Context(), userID, id); err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ============================== User: PT packages (purchased) ==============================

func (c *Controller) ListPurchasedPTPackages(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	res, err := c.svc.ListPurchasedPTPackages(ctx.Request.Context(), userID)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) CheckoutPTPackageVNPay(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	var req gcv1.PackageCheckoutReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.CheckoutPTPackageVNPay(ctx.Request.Context(), userID, req.PackageID, ctx.ClientIP())
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) ConfirmPTPackageVNPay(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	var req gcv1.ConfirmVNPayReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.ConfirmPTPackageVNPay(ctx.Request.Context(), userID, req.Params)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) CheckoutPTPackageStripe(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	var req gcv1.PackageCheckoutReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.CheckoutPTPackageStripe(ctx.Request.Context(), userID, req.PackageID)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) ConfirmPTPackageStripe(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	var req gcv1.ConfirmStripeReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.ConfirmPTPackageStripe(ctx.Request.Context(), userID, req.SessionID)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) LogPTSession(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	var req gcv1.LogPTSessionReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.LogPTSession(ctx.Request.Context(), userID, id, &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) ListPTSessions(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	res, err := c.svc.ListPTSessions(ctx.Request.Context(), userID, id)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) ListChatMessages(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	var afterID uint
	if v := queryUintPtr(ctx, "after_id"); v != nil {
		afterID = *v
	}
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "100"))
	res, err := c.svc.ListChatMessages(ctx.Request.Context(), userID, id, afterID, limit)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) SendChatMessage(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	var req gcv1.SendChatMessageReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.SendChatMessage(ctx.Request.Context(), userID, id, &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) ListSessionOffers(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	res, err := c.svc.ListSessionOffers(ctx.Request.Context(), userID, id)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) CreateSessionOffer(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	var req gcv1.CreateSessionOfferReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.CreateSessionOffer(ctx.Request.Context(), userID, id, &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, res)
}

func (c *Controller) AcceptSessionOffer(ctx *gin.Context) {
	c.mutateSessionOffer(ctx, func(userID, pkgID, offerID uint) (*gcv1.SessionOfferRes, error) {
		return c.svc.AcceptSessionOffer(ctx.Request.Context(), userID, pkgID, offerID)
	})
}

func (c *Controller) DeclineSessionOffer(ctx *gin.Context) {
	c.mutateSessionOffer(ctx, func(userID, pkgID, offerID uint) (*gcv1.SessionOfferRes, error) {
		return c.svc.DeclineSessionOffer(ctx.Request.Context(), userID, pkgID, offerID)
	})
}

func (c *Controller) CancelSessionOffer(ctx *gin.Context) {
	c.mutateSessionOffer(ctx, func(userID, pkgID, offerID uint) (*gcv1.SessionOfferRes, error) {
		return c.svc.CancelSessionOffer(ctx.Request.Context(), userID, pkgID, offerID)
	})
}

func (c *Controller) RescheduleSessionOffer(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	pkgID, offerID, ok := parsePackageOfferIDs(ctx)
	if !ok {
		return
	}
	var req gcv1.RescheduleSessionOfferReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.RescheduleSessionOffer(ctx.Request.Context(), userID, pkgID, offerID, req.StartsAt)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) CompleteSessionOffer(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	pkgID, offerID, ok := parsePackageOfferIDs(ctx)
	if !ok {
		return
	}
	var req gcv1.CompleteSessionOfferReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.CompleteSessionOffer(ctx.Request.Context(), userID, pkgID, offerID, &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) ConfirmSessionOffer(ctx *gin.Context) {
	c.mutateSessionOffer(ctx, func(userID, pkgID, offerID uint) (*gcv1.SessionOfferRes, error) {
		return c.svc.ConfirmSessionOffer(ctx.Request.Context(), userID, pkgID, offerID)
	})
}

func (c *Controller) RejectSessionOfferProof(ctx *gin.Context) {
	c.mutateSessionOffer(ctx, func(userID, pkgID, offerID uint) (*gcv1.SessionOfferRes, error) {
		return c.svc.RejectSessionOfferProof(ctx.Request.Context(), userID, pkgID, offerID)
	})
}

func (c *Controller) MarkSessionNoShow(ctx *gin.Context) {
	c.mutateSessionOffer(ctx, func(userID, pkgID, offerID uint) (*gcv1.SessionOfferRes, error) {
		return c.svc.MarkSessionNoShow(ctx.Request.Context(), userID, pkgID, offerID)
	})
}

func (c *Controller) mutateSessionOffer(ctx *gin.Context, fn func(userID, pkgID, offerID uint) (*gcv1.SessionOfferRes, error)) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	pkgID, offerID, ok := parsePackageOfferIDs(ctx)
	if !ok {
		return
	}
	res, err := fn(userID, pkgID, offerID)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func parsePackageOfferIDs(ctx *gin.Context) (pkgID, offerID uint, ok bool) {
	var err error
	pkgID, err = parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return 0, 0, false
	}
	offerID, err = parseUintParam(ctx, "offerId")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid offer id"})
		return 0, 0, false
	}
	return pkgID, offerID, true
}

func (c *Controller) ListSoldPTPackages(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	page, limit := queryPageLimit(ctx)
	status := ctx.Query("status")
	res, err := c.svc.ListSoldPTPackages(ctx.Request.Context(), userID, page, limit, status)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) GetUserPTPackage(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	res, err := c.svc.GetUserPTPackage(ctx.Request.Context(), userID, id)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

// ============================== User: PT earnings ==============================

func (c *Controller) MyPTEarnings(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	page, limit := queryPageLimit(ctx)
	res, err := c.svc.MyPTEarnings(ctx.Request.Context(), userID, page, limit)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

// ============================== Public: PT packages by trainer ==============================

func (c *Controller) ListPublicPTPackagesByTrainer(ctx *gin.Context) {
	trainerID, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	page, limit := queryPageLimit(ctx)
	res, err := c.svc.ListPublicPTPackagesByTrainer(ctx.Request.Context(), trainerID, page, limit)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

// ============================== User: PT schedule / slot booking ==============================

func parseRFC3339Query(ctx *gin.Context, name string) (time.Time, error) {
	v := strings.TrimSpace(ctx.Query(name))
	if v == "" {
		return time.Time{}, fmt.Errorf("%s required", name)
	}
	return time.Parse(time.RFC3339, v)
}

func (c *Controller) GetMyBookingSettings(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	res, err := c.svc.GetMyBookingSettings(ctx.Request.Context(), userID)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) UpdateMyBookingSettings(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	var req gcv1.BookingSettingsReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.UpdateMyBookingSettings(ctx.Request.Context(), userID, &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) GetMyWorkingHours(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	res, err := c.svc.GetMyWorkingHours(ctx.Request.Context(), userID)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) SetMyWorkingHours(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	var req gcv1.SetWorkingHoursReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.SetMyWorkingHours(ctx.Request.Context(), userID, &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) ListMyBlockedSlots(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	from, err := parseRFC3339Query(ctx, "from")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	to, err := parseRFC3339Query(ctx, "to")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.ListMyBlockedSlots(ctx.Request.Context(), userID, from, to)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) BlockMySlot(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	var req gcv1.BlockSlotReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.BlockMySlot(ctx.Request.Context(), userID, &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, res)
}

func (c *Controller) UnblockMySlot(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	if err := c.svc.UnblockMySlot(ctx.Request.Context(), userID, id); err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *Controller) ListAvailableSlots(ctx *gin.Context) {
	trainerID, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	from, err := parseRFC3339Query(ctx, "from")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	to, err := parseRFC3339Query(ctx, "to")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.ListAvailableSlots(ctx.Request.Context(), trainerID, from, to)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) BookSlot(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	var req gcv1.BookSlotReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.BookSlot(ctx.Request.Context(), userID, &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, res)
}

func (c *Controller) AdminTrainerOpsOverview(ctx *gin.Context) {
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	res, err := c.svc.AdminTrainerOpsOverview(ctx.Request.Context(), id)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) AdminListTrainerClients(ctx *gin.Context) {
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	page, limit := queryPageLimit(ctx)
	status := strings.TrimSpace(ctx.Query("status"))
	res, err := c.svc.AdminListTrainerClients(ctx.Request.Context(), id, page, limit, status)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) AdminListTrainerBookings(ctx *gin.Context) {
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	from, err := parseRFC3339Query(ctx, "from")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	to, err := parseRFC3339Query(ctx, "to")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.AdminListTrainerBookings(ctx.Request.Context(), id, from, to)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) AdminContentFunnel(ctx *gin.Context) {
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	res, err := c.svc.AdminContentFunnel(ctx.Request.Context(), id)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) AdminTrainerQuality(ctx *gin.Context) {
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	res, err := c.svc.AdminTrainerQuality(ctx.Request.Context(), id)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) AdminTrainerCalendar(ctx *gin.Context) {
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	from, err := parseRFC3339Query(ctx, "from")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	to, err := parseRFC3339Query(ctx, "to")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.AdminTrainerCalendar(ctx.Request.Context(), id, from, to)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) CreatePTReview(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	var req gcv1.CreatePTReviewReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.CreatePTReview(ctx.Request.Context(), userID, &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, res)
}

func (c *Controller) ListPTReviews(ctx *gin.Context) {
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	page, limit := queryPageLimit(ctx)
	res, err := c.svc.ListPTReviews(ctx.Request.Context(), id, page, limit)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) TouchPTFunnel(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	var req gcv1.TouchPTFunnelReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	if err := c.svc.TouchPTFunnel(ctx.Request.Context(), userID, &req); err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}
