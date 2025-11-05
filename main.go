package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"Api/database"
	"Api/middleware"
	"Api/routes"
	"Api/tasks"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env locally
	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			log.Println("⚠️ No .env found, using system env vars")
		}
	}

	fmt.Println("SUPER_ADMIN_SECRET:", os.Getenv("SUPER_ADMIN_SECRET"))

	// Connect to DB + run expired bot task
	database.InitDB()
	tasks.DeactivateExpiredBots()

	// Gin config
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.Use(middleware.CORSMiddleware())

	// Set up API routes
	routes.SetUpRouter(r)

	// Frontend path
	frontendPath := "./Frontend" // Use relative to main.go or adjust as needed

	// Serve assets (CSS, JS, images)
	r.Static("/assets", frontendPath)

	// Serve all HTML files dynamically
	htmlFiles, err := filepath.Glob(filepath.Join(frontendPath, "*.html"))
	if err != nil {
		log.Fatal("Failed to scan frontend HTML files:", err)
	}
	for _, file := range htmlFiles {
		file := file // capture for closure
		_, filename := filepath.Split(file)
		route := "/" + filename[:len(filename)-len(".html")] // /auth for auth.html
		if route == "/index" {
			route = "/" // index.html should be root
		}
		r.GET(route, func(c *gin.Context) {
			c.File(file)
		})
	}

	// SPA fallback: serve index.html for all unmatched routes
	r.NoRoute(func(c *gin.Context) {
		c.File(filepath.Join(frontendPath, "index.html"))
	})

	// Port config
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 Server running on http://localhost:%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}
