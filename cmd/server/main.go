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

	// ✅ Load env only locally
	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			log.Println("⚠️ No .env found, using system env vars")
		}
	}

	fmt.Println("SUPER_ADMIN_SECRET:", os.Getenv("SUPER_ADMIN_SECRET"))

	// ✅ Connect DB + run expired bot task
	database.InitDB()
	tasks.DeactivateExpiredBots()

	// ✅ Gin config
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.Use(middleware.CORSMiddleware())

	// ✅ API routes — PASS ENGINE, NOT GROUP
	routes.SetUpRouter(r)

	// ✅ Frontend path (for Tailwind static HTML project)
	frontendPath := "../../Frontend"

	// Serve assets folder
	r.Static("/assets", frontendPath)

	// Serve index.html at root
	r.GET("/", func(c *gin.Context) {
		c.File(frontendPath + "/index.html")
	})

	// Fallback — browser routing SPA
	r.NoRoute(func(c *gin.Context) {
		c.File(frontendPath + "/index.html")
	})

	// ✅ Port config
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 Server running http://localhost:%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}
