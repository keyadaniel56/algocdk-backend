package database

import (
	"Api/models"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dsn := os.Getenv("DATABASE_URL") // Render environment variable

	if dsn == "" {
		log.Fatal("❌ DATABASE_URL is not set — did you add it on Render?")
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to connect to PostgreSQL: ", err)
	}

	log.Println("✅ PostgreSQL connected successfully")

	// Auto migrate models
	err = DB.AutoMigrate(
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

	if err != nil {
		log.Fatal("❌ Auto migration failed: ", err)
	}

	log.Println("✅ Tables migrated successfully")

	// Ensure uploads folder exists (still valid)
	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		os.Mkdir("uploads", os.ModePerm)
	}
}
