package handlers

import (
	"Api/database"
	"Api/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetBotDetails(ctx *gin.Context) {
	botID := ctx.Param("id")
	var bot models.Bot
	if err := database.DB.First(&bot, botID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "Bot not found"})
		return
	}

	var admin models.Admin
	if err := database.DB.Where("person_id = ?", bot.OwnerID).First(&admin).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "Admin not found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Bot details retrieved",
		"data": map[string]interface{}{
			"id":           bot.ID,
			"admin_id":     admin.ID,
			"price":        bot.Price,
			"rent_price":   bot.RentPrice,
			"payment_type": bot.SubscriptionType,
			"name":         bot.Name,
			"description":  bot.Description,
		},
	})
}
