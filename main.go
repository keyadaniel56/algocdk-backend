package main

import (
	"fmt"
	"log"
	"os"

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
	frontendPath := "./Frontend"

	// Serve assets
	r.Static("/assets", frontendPath)

	// Serve HTML files manually
	r.GET("/", func(c *gin.Context) {
		c.File(frontendPath + "/index.html")
	})
	r.GET("/auth", func(c *gin.Context) {
		c.File(frontendPath + "/auth.html")
	})
	r.GET("/app", func(c *gin.Context) {
		c.File(frontendPath + "/app.html")
	})
	r.GET("/mybots", func(c *gin.Context) {
		c.File(frontendPath + "/mybots.html")
	})
	r.GET("/botstore", func(c *gin.Context) {
		c.File(frontendPath + "/botstore.html")
	})
	r.GET("/admin_dashboard", func(c *gin.Context) {
		c.File(frontendPath + "/admin_dashboard.html")
	})
	r.GET("/superadmin_dashboard", func(c *gin.Context) {
		c.File(frontendPath + "/superadmin_dashboard.html")
	})
	r.GET("/support", func(c *gin.Context) {
		c.File(frontendPath + "/support.html")
	})
	r.GET("/privacy", func(c *gin.Context) {
		c.File(frontendPath + "/privacy.html")
	})
	r.GET("/terms", func(c *gin.Context) {
		c.File(frontendPath + "/terms.html")
	})
	r.GET("/legal", func(c *gin.Context) {
		c.File(frontendPath + "/legal.html")
	})
	r.GET("/marketchart", func(c *gin.Context) {
		c.File(frontendPath + "/marketchart.html")
	})
	r.GET("/test_upgrade", func(c *gin.Context) {
		c.File(frontendPath + "/test_upgrade.html")
	})
	r.GET("/video", func(c *gin.Context) {
		c.File(frontendPath + "/video.html")
	})

	// SPA fallback
	r.NoRoute(func(c *gin.Context) {
		c.File(frontendPath + "/index.html")
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
