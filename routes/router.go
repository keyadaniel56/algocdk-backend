// package routes

// import (
// 	"Api/handlers"
// 	"Api/middleware"
// 	"Api/paystack"

// 	"github.com/gin-gonic/gin"
// )

// func SetUpRouter(router *gin.Engine) {
// 	// Root route
// 	router.GET("/marketplace", handlers.MarketplaceHandler)

// 	// Serve static files for bots
// 	router.Static("/uploads", "./uploads")

// 	// -----------------------------
// 	// 🌐 Main API group
// 	// -----------------------------
// 	api := router.Group("/api")
// 	{
// 		// -----------------------------
// 		// 🔓 AUTH ROUTES
// 		// -----------------------------
// 		auth := api.Group("/auth")
// 		{
// 			auth.POST("/login", handlers.LoginHandler)
// 			auth.POST("/register", handlers.SignupHandler)
// 		}

// 		// -----------------------------
// 		// 🔐 USER ROUTES
// 		// -----------------------------
// 		user := api.Group("/user")
// 		user.Use(middleware.AuthMiddleware()) // user must be logged in
// 		{
// 			user.GET("/me", handlers.ProfileHandler)
// 			user.GET("/me/favorites", handlers.GetUserFavorites)
// 			user.POST("/favorites/:bot_id", handlers.ToggleFavorite)
// 			user.POST("/request-upgrade", handlers.RequestAdminUpgrade)
// 			user.GET("/ws", handlers.WebSocketHandler)

// 			// Paystack routes for logged-in users
// 			user.POST("/paystack/initialize", paystack.InitializePayment) // Initialize payment
// 			user.GET("/paystack/verify", paystack.VerifyPayment)    // Verify payment callback
// 			// Paystack webhook / callback route (no auth, because Paystack calls this)
// 			api.POST("/paystack/callback", paystack.PaystackCallback)

// 		}

// 		// -----------------------------
// 		// 🛠 ADMIN ROUTES
// 		// -----------------------------
// 		admin := api.Group("/admin")
// 		admin.Use(middleware.AuthMiddleware())
// 		{
// 			admin.GET("/dashboard", handlers.AdminDashboardHandler)
// 			admin.POST("/create-bot", handlers.CreateBotHandler)
// 			admin.PUT("/update-bot/:id", handlers.UpdateBotHandler)
// 			admin.DELETE("/delete-bot/:id", handlers.DeleteBotHandler)
// 			admin.GET("/bots", handlers.ListAdminBotsHandler)
// 			admin.GET("/profile", handlers.AdminProfileHandler)

// 		}

// 		// -----------------------------
// 		// 👑 SUPER ADMIN ROUTES
// 		// -----------------------------
// 		superAdmin := api.Group("/superadmin")
// 		{
// 			// Public routes
// 			superAdmin.POST("/register", handlers.SuperAdminRegisterHandler)
// 			superAdmin.POST("/login", handlers.SuperAdminLoginHandler)

// 			// Protected routes
// 			superAdmin.Use(middleware.AuthMiddleware())
// 			{
// 				superAdmin.GET("/profile", handlers.SuperAdminProfileHandler)
// 				superAdmin.GET("/dashboard", handlers.SuperAdminDashboardHandler)
// 				superAdmin.GET("/users", handlers.GetAllUsers)
// 				superAdmin.POST("/create-user", handlers.CreateUser)
// 				superAdmin.POST("/update-user", handlers.UpdateUser)
// 				superAdmin.DELETE("/delete-user", handlers.DeleteUser)
// 				superAdmin.GET("/pending-requests", handlers.GetPendingRequests)
// 				superAdmin.POST("/promote/:id", handlers.ApproveUpgrade)
// 				superAdmin.POST("/reject/:id", handlers.RejectUpgrade)
// 				superAdmin.GET("/admins", handlers.GetAllAdmins)
// 				superAdmin.POST("/create-admin", handlers.CreateAdmin)
// 				superAdmin.PUT("/update-admin/:id", handlers.UpdateAdmin)
// 				superAdmin.PATCH("/toggle-admin/:id", handlers.ToggleAdminStatus)
// 				superAdmin.DELETE("/delete-admin/:id", handlers.DeleteAdmin)
// 				superAdmin.GET("/bots", handlers.GetBotsHandler)
// 				// routes/routes.go
// 				superAdmin.Handle("GET", "/scan-bots", handlers.ScanAllBotsHandler)
// 				superAdmin.Handle("POST", "/scan-bots", handlers.ScanAllBotsHandler)

// 				superAdmin.GET("/ws", handlers.WebSocketHandler)
// 			}
// 		}
// 	}
// }

package routes

import (
	"Api/handlers"
	"Api/middleware"
	"Api/paystack"

	"github.com/gin-gonic/gin"
)

func SetUpRouter(router *gin.Engine) {
	// Root route
	router.GET("/marketplace", handlers.MarketplaceHandler)
	router.GET("/api/paystack/callback", paystack.HandleCallbackRedirect)
	// Serve static files for bots
	router.Static("/uploads", "./Uploads")

	// -----------------------------
	// 🌐 Main API group
	// -----------------------------
	api := router.Group("/api")
	{
		// -----------------------------
		// 🔓 AUTH ROUTES
		// -----------------------------
		auth := api.Group("/auth")
		{
			auth.POST("/login", handlers.LoginHandler)
			auth.POST("/register", handlers.SignupHandler)
		}
		api.GET("/bots/:id", handlers.GetBotDetails)
		// -----------------------------
		// 🔐 USER ROUTES
		// -----------------------------
		user := api.Group("/user")
		user.Use(middleware.AuthMiddleware()) // user must be logged in
		{
			user.GET("/me", handlers.ProfileHandler)
			user.GET("/me/favorites", handlers.GetUserFavorites)
			user.POST("/favorites/:bot_id", handlers.ToggleFavorite)
			user.POST("/request-upgrade", handlers.RequestAdminUpgrade)
			user.GET("/ws", handlers.WebSocketHandler)
		}

		// -----------------------------
		// 💳 PAYSTACK ROUTES
		// -----------------------------
		paystackGroup := api.Group("/paystack")
		{
			// Authenticated routes (require logged-in user)
			paystackGroup.Use(middleware.AuthMiddleware())
			{
				paystackGroup.POST("/initialize", paystack.InitializePayment) // Initialize payment
				paystackGroup.GET("/verify", paystack.VerifyPayment)          // Verify payment via reference
				paystackGroup.POST("/callback", paystack.FrontendCallback)    // Handle frontend callback
				paystackGroup.POST("update-transaction", paystack.UpdateTransaction)

			}

			// Unauthenticated route for Paystack webhook
			paystackGroup.POST("/webhook", paystack.PaystackCallback) // Paystack webhook for charge.success
		}

		// -----------------------------
		// 🛠 ADMIN ROUTES
		// -----------------------------
		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware())
		{
			admin.GET("/dashboard", handlers.AdminDashboardHandler)
			admin.POST("/create-bot", handlers.CreateBotHandler)
			admin.PUT("/update-bot/:id", handlers.UpdateBotHandler)
			admin.DELETE("/delete-bot/:id", handlers.DeleteBotHandler)
			admin.GET("/bots", handlers.ListAdminBotsHandler)
			admin.GET("/profile", handlers.AdminProfileHandler)
			admin.PUT("/bank-details", handlers.UpdateAdminBankDetails)
			admin.GET("/transactions", handlers.GetAdminTransactions)
			admin.POST("/transactions", handlers.RecordTransaction)
		}

		// -----------------------------
		// 👑 SUPER ADMIN ROUTES
		// -----------------------------
		superAdmin := api.Group("/superadmin")
		{
			// Public routes
			superAdmin.POST("/register", handlers.SuperAdminRegisterHandler)
			superAdmin.POST("/login", handlers.SuperAdminLoginHandler)

			// Protected routes
			superAdmin.Use(middleware.AuthMiddleware())
			{
				superAdmin.GET("/profile", handlers.SuperAdminProfileHandler)
				superAdmin.GET("/dashboard", handlers.SuperAdminDashboardHandler)
				superAdmin.GET("/users", handlers.GetAllUsers)
				superAdmin.POST("/create-user", handlers.CreateUser)
				superAdmin.POST("/update-user", handlers.UpdateUser)
				superAdmin.DELETE("/delete-user", handlers.DeleteUser)
				superAdmin.GET("/pending-requests", handlers.GetPendingRequests)
				superAdmin.POST("/promote/:id", handlers.ApproveUpgrade)
				superAdmin.POST("/reject/:id", handlers.RejectUpgrade)
				superAdmin.GET("/admins", handlers.GetAllAdmins)
				superAdmin.POST("/create-admin", handlers.CreateAdmin)
				superAdmin.PUT("/update-admin/:id", handlers.UpdateAdmin)
				superAdmin.PATCH("/toggle-admin/:id", handlers.ToggleAdminStatus)
				superAdmin.DELETE("/delete-admin/:id", handlers.DeleteAdmin)
				superAdmin.GET("/bots", handlers.GetBotsHandler)
				superAdmin.Handle("GET", "/scan-bots", handlers.ScanAllBotsHandler)
				superAdmin.Handle("POST", "/scan-bots", handlers.ScanAllBotsHandler)
				superAdmin.GET("/ws", handlers.WebSocketHandler)
				superAdmin.GET("/transactions", handlers.GetAllTransactions)
			}
		}
	}
}
