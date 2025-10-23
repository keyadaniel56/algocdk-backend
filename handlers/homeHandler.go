package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HomeController(ctx *gin.Context) {
	ctx.JSON(http.StatusAccepted, gin.H{"message": "welcome home"})
}
