package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Health() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "UP")
	}
}
