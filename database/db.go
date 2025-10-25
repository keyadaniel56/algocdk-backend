package database

import (
	"Api/models"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	// Persistent path for SQLite (works locally and on Koyeb)
	dbPath := "app.db"
	if _, exists := os.LookupEnv("KOYEB_DEPLOYMENT_ID"); exists {
		// Use persistent volume path when running on Koyeb
		dbPath = "/app/data/app.db"
	}

	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	// Auto migrate tables
	DB.AutoMigrate(
		&models.Person{},
		&models.Bot{},
		&models.Favorite{},
		&models.BotUser{},
		&models.Admin{},
		&models.Transaction{},
		&models.SalesHistory{},
		&models.UserBot{},
		&models.Sale{},
	)

	// Ensure uploads directory exists
	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		os.Mkdir("uploads", os.ModePerm)
	}
}
