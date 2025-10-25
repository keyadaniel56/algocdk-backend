package tasks

import (
	"Api/database"
	"Api/models"
	"log"
	"time"
)

// DeactivateExpiredBots automatically disables expired rental bots.

func DeactivateExpiredBots() {
	log.Println("[Scheduler] Checking for expired rented bots...")

	var expiredBots []models.UserBot
	// OLD QUERY (caused error):
	// database.DB.Where("type = ? AND expiry < ? AND active = ?", "rent", time.Now(), true).Find(&expiredBots)

	// ✅ NEW QUERY (matches your current model)
	database.DB.Where("access_type = ? AND expiry_date < ? AND is_active = ?", "rent", time.Now(), true).Find(&expiredBots)

	if len(expiredBots) == 0 {
		log.Println("[Scheduler] No expired bots found.")
		return
	}

	for _, bot := range expiredBots {
		bot.IsActive = false
		database.DB.Save(&bot)
		log.Printf("[Scheduler] Deactivated bot ID %d (UserID: %d)\n", bot.BotID, bot.UserID)
	}
}
