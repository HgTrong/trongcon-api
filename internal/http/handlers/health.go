package handlers

import (
	"net/http"

	"trongcon-api/api/swagger"

	"github.com/gin-gonic/gin"
)

// HealthCheck godoc
// @Summary Health check
// @Description Check whether the API is up.
// @Tags health
// @Produce json
// @Success 200 {object} swagger.HealthOK
// @Router /health [get]
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, swagger.HealthOK{Status: "ok"})
}
