package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	apierrors "github.com/hhftechnology/middleware-manager/api/errors"
)

func respondError(c *gin.Context, status int, message string, err error) {
	apierrors.HandleAPIError(c, status, message, err)
}

func respondJSON(c *gin.Context, payload any) {
	c.JSON(http.StatusOK, payload)
}
