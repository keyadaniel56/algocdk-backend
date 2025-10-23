package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GreetHandler — basic greeting for Rasa↔Sara connection test
func GreetHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello, I’m Sara 🤖 — your AlgocdK assistant!",
	})
}
