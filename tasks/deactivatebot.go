package tasks

import (
	"Api/database"
	"Api/models"
	"log"
	"time"
)

// DeactivateExpiredBots automatically disables expired rental bots.
// It should be run periodically (e.g. every hour or once a day via goroutine or cron).
func DeactivateExpiredBots() {
	log.Println("[Scheduler] Checking for expired rented bots...")

	var expiredBots []models.UserBot

	// Find all active rental bots that have expired
	if err := database.DB.
		Where("type = ? AND expiry < ? AND active = ?", "rent", time.Now(), true).
		Find(&expiredBots).Error; err != nil {
		log.Printf("[Scheduler] Error fetching expired bots: %v\n", err)
		return
	}

	if len(expiredBots) == 0 {
		log.Println("[Scheduler] No expired bots found.")
		return
	}

	// Deactivate all expired bots
	result := database.DB.Model(&models.UserBot{}).
		Where("type = ? AND expiry < ? AND active = ?", "rent", time.Now(), true).
		Update("active", false)

	if result.Error != nil {
		log.Printf("[Scheduler] Error deactivating expired bots: %v\n", result.Error)
		return
	}

	log.Printf("[Scheduler] Successfully deactivated %d expired rented bots.\n", result.RowsAffected)
}
