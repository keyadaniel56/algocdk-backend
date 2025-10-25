package main

import (
	"Api/database"
	"Api/middleware"
	"Api/routes"
	"Api/tasks"

	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 1️⃣ Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
	fmt.Println("SUPER_ADMIN_SECRET:", os.Getenv("SUPER_ADMIN_SECRET"))

	// 2️⃣ Initialize database
	database.InitDB()
	tasks.DeactivateExpiredBots()
	// 3️⃣ Create Gin router
	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.Use(middleware.CORSMiddleware())
	// 4️⃣ Setup routes
	routes.SetUpRouter(r)

	// 5️⃣ Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Static("/Frontend", "../../Frontend")
	r.StaticFile("/output.css", "../../Frontend/output.css")

	// Serve static files (HTML, CSS, JS)
	r.GET("/", func(c *gin.Context) {
		c.File("../../Frontend/index.html")
	})

	r.GET("/admin", func(ctx *gin.Context) {
		ctx.File("../../Frontend/admin_dashboard.html")
	})
	// main.go

	r.GET("/auth", func(ctx *gin.Context) {
		ctx.File("../../Frontend/auth.html")
	})

	r.GET("/superadmin", func(ctx *gin.Context) {
		ctx.File("../../Frontend/superadmin_dashboard.html")
	})

	r.GET("/terms", func(ctx *gin.Context) {
		ctx.File("../../Frontend/terms.html")
	})

	r.GET("/privacy", func(ctx *gin.Context) {
		ctx.File("../../Frontend/privacy.html")
	})

	r.GET("/support", func(ctx *gin.Context) {
		ctx.File("../../Frontend/support.html")
	})

	r.GET("/app", func(ctx *gin.Context) {
		ctx.File("../../Frontend/app.html")
	})

	r.GET("/botstore", func(ctx *gin.Context) {
		ctx.File("../../Frontend/botstore.html")
	})

	r.GET("/mybots", func(ctx *gin.Context) {
		ctx.File("../../Frontend/mybots.html")
	})

	r.GET("/chart", func(ctx *gin.Context) {
		ctx.File("../../Frontend/marketchart.html")
	})

	r.GET("/greet", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hey, I’m Sara from AlgoCDK! How’s trading going?",
		})
	})

	// Example: market signal endpoint
	r.GET("/market_signal", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"symbol": "Volatility 75 Index",
			"signal": "Buy",
		})
	})
	fmt.Printf("🚀 Server running on http://localhost:%s\n", port)
	r.Run(":" + port)

}
