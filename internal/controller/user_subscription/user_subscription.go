package user_subscription

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"trongcon-api/api/swagger"
	subv1 "trongcon-api/api/user_subscription/v1"
	"trongcon-api/internal/config"
	"trongcon-api/internal/http/middleware"
	"trongcon-api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
)

type Controller struct {
	svc       service.UserSubscriptionService
	stripeCfg config.StripeConfig
}

func NewController(svc service.UserSubscriptionService, stripeCfg config.StripeConfig) *Controller {
	return &Controller{svc: svc, stripeCfg: stripeCfg}
}

func (c *Controller) Checkout(ctx *gin.Context) {
	// Back-compat: PayPal
	c.CheckoutPayPal(ctx)
}

func (c *Controller) CheckoutPayPal(ctx *gin.Context) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "unauthorized"})
		return
	}
	var req subv1.CheckoutReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.CheckoutPayPal(ctx.Request.Context(), userID, req.PlanID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) CheckoutStripe(ctx *gin.Context) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "unauthorized"})
		return
	}
	var req subv1.CheckoutReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.CheckoutStripe(ctx.Request.Context(), userID, req.PlanID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) Capture(ctx *gin.Context) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "unauthorized"})
		return
	}
	var req subv1.CaptureReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.CapturePayPal(ctx.Request.Context(), userID, req.Token, req.OrderID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) ConfirmStripe(ctx *gin.Context) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "unauthorized"})
		return
	}
	var req subv1.ConfirmStripeReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.ConfirmStripe(ctx.Request.Context(), userID, req.SessionID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) Me(ctx *gin.Context) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "unauthorized"})
		return
	}
	res, err := c.svc.Me(ctx.Request.Context(), userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) ListAdmin(ctx *gin.Context) {
	var req subv1.ListAdminReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.ListAdmin(ctx.Request.Context(), &req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, res)
}

// StripeWebhook handles checkout.session.completed (raw body + signature).
func (c *Controller) StripeWebhook(ctx *gin.Context) {
	const maxBody = int64(65536)
	payload, err := io.ReadAll(io.LimitReader(ctx.Request.Body, maxBody))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "read body failed"})
		return
	}
	sig := ctx.GetHeader("Stripe-Signature")
	var event stripe.Event
	if strings.TrimSpace(c.stripeCfg.WebhookSecret) != "" {
		event, err = webhook.ConstructEvent(payload, sig, c.stripeCfg.WebhookSecret)
		if err != nil {
			log.Printf("[stripe webhook] signature: %v", err)
			ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid signature"})
			return
		}
	} else {
		if err := json.Unmarshal(payload, &event); err != nil {
			ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid json"})
			return
		}
		log.Printf("[stripe webhook] WARNING: STRIPE_WEBHOOK_SECRET empty — skipping verify")
	}

	switch event.Type {
	case "checkout.session.completed":
		var sess stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &sess); err != nil {
			ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "unmarshal session failed"})
			return
		}
		if sess.Metadata["type"] != "" && sess.Metadata["type"] != "user_subscription" {
			ctx.JSON(http.StatusOK, gin.H{"ok": true, "skipped": true})
			return
		}
		if _, err := c.svc.ActivateFromStripeSession(ctx.Request.Context(), &sess); err != nil {
			log.Printf("[stripe webhook] activate: %v", err)
			ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
			return
		}
	default:
		// ignore other events for now
	}
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

func EnsurePremiumMessage(err error) string {
	if err == nil {
		return ""
	}
	if strings.Contains(err.Error(), "premium") {
		return "premium_required"
	}
	return err.Error()
}
