package controllers

import (
	"Api/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// POST /users/me/favorites/:bot_id
func ToggleFavorite(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.GetUint("user_id")

	botID, err := strconv.Atoi(c.Param("bot_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bot ID"})
		return
	}

	var favorite models.Favorite

	// Check if this favorite already exists
	err = db.Where("user_id = ? AND bot_id = ?", userID, botID).First(&favorite).Error

	if err == nil {
		// Exists → unfavorite (delete)
		db.Delete(&favorite)
		c.JSON(http.StatusOK, gin.H{
			"message": "Bot removed from favorites",
			"status":  "unfavorited",
		})
		return
	}

	if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Not found → favorite it
	newFavorite := models.Favorite{
		UserID: userID,
		BotID:  uint(botID),
	}

	if err := db.Create(&newFavorite).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add favorite"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Bot added to favorites",
		"status":  "favorited",
	})
}
