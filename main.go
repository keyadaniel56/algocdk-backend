package main

import (
	"Api/database"
	"Api/middleware"
	"Api/routes"
	"Api/tasks"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	// Load .env locally only
	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			log.Println("⚠️ No .env found, using system env vars")
		}
	}

	fmt.Println("SUPER_ADMIN_SECRET:", os.Getenv("SUPER_ADMIN_SECRET"))

	// Initialize DB + deactivate expired bots
	database.InitDB()
	tasks.DeactivateExpiredBots()

	// Gin config
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.Use(middleware.CORSMiddleware())

	// API routes
	routes.SetUpRouter(r)

	// Frontend folder relative to main.go (repo root)
	frontendPath := "Frontend"

	// Serve static assets (CSS, JS, images)
	r.Static("/assets", frontendPath)

	// Serve index.html at root
	r.GET("/", func(c *gin.Context) {
		c.File(frontendPath + "/index.html")
	})

	// Fallback for SPA routes
	r.NoRoute(func(c *gin.Context) {
		c.File(frontendPath + "/index.html")
	})

	// Port configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 Server running on port %s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}
