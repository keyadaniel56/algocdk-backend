package handlers

import (
	"Api/database"

	"Api/models"
	"Api/utils"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func LoginHandler(ctx *gin.Context) {
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Bind JSON
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Sanitize input
	payload.Email = strings.ToLower(strings.TrimSpace(payload.Email))
	payload.Password = strings.TrimSpace(payload.Password)

	if payload.Email == "" || payload.Password == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "email and password are required"})
		return
	}

	// Look up user
	var user models.Person
	if err := database.DB.Where("email = ?", payload.Email).First(&user).Error; err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	// Verify password
	if !utils.CheckPasswordHash(payload.Password, user.Password) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid password"})
		return
	}

	// Generate token
	token, _ := utils.GenerateToken(user.ID, user.Email)

	// Respond with user info, role, and membership
	ctx.JSON(http.StatusOK, gin.H{
		"message":    "login successful",
		"token":      token,
		"role":       user.Role,       // "User" or "Admin"
		"membership": user.Membership, // free, silver, gold, etc.
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
		"last_login": time.Now(),
	})

	fmt.Println("Login attempt:")
fmt.Println("Email:", payload.Email)
fmt.Println("Entered password:", payload.Password)
fmt.Println("Stored hash:", user.Password)
fmt.Println("Check result:", utils.CheckPasswordHash(payload.Password, user.Password))
fmt.Println("Role:", user.Role)

}

// func SignupHandler(ctx *gin.Context) {
// 	var payload struct {
// 		Name     string `json:"name"`
// 		Email    string `json:"email"`
// 		Password string `json:"password"`
// 	}
// 	if err := ctx.ShouldBindJSON(&payload); err != nil {
//     ctx.JSON(http.StatusBadRequest, gin.H{
//         "error": "invalid request body",
//         "details": err.Error(),
//     })
//     return
// }

// 	payload.Email = strings.ToLower(strings.TrimSpace(payload.Email))
// 	payload.Name = strings.TrimSpace(payload.Name)
// 	payload.Password = strings.TrimSpace(payload.Password)

// 	if payload.Email == "" || payload.Name == "" || payload.Password == "" {
// 		ctx.JSON(http.StatusBadRequest, gin.H{"error": "all fields required"})
// 		return
// 	}

// 	hashedPassword, _ := utils.HashPassword(payload.Password)
// 	user := models.Person{
// 		Name:      payload.Name,
// 		Email:     payload.Email,
// 		Password:  hashedPassword,
// 		CreatedAt: time.Now(),
// 		UpdatedAt: time.Now(),
// 	}

// 	if err := database.DB.Create(&user).Error; err != nil {
// 		// return DB error text for debugging (remove details in production)
// 		ctx.JSON(http.StatusBadRequest, gin.H{"error": "could not create user", "details": err.Error()})
// 		return
// 	}

// 	token, _ := utils.GenerateToken(user.ID, user.Email)
// 	ctx.JSON(http.StatusOK, gin.H{"message": "signup successful", "token": token})
// }

// func ProfileHandler(ctx *gin.Context) {
// 	userID := ctx.GetUint("user_id")
// 	var user models.Person
// 	if err := database.DB.First(&user, userID).Error; err != nil {
// 		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, gin.H{
// 		"id":     user.ID,
// 		"name":   user.Name,
// 		"email":  user.Email,
// 		"joined": user.CreatedAt.Format(time.RFC3339),
// 		"membership": user.Membership,
// 		"role": user.Role,
// 	})
// }

func ProfileHandler(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	var user models.Person
	if err := database.DB.First(&user, userID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Determine upgrade status message
	upgradeMessage := ""
	switch user.UpgradeRequestStatus {
	case "pending":
		upgradeMessage = "Your request to become admin is pending."
	case "approved":
		upgradeMessage = "Your request has been approved. You are now an admin."
	case "rejected":
		upgradeMessage = "Your request to become admin has been rejected."
	default: // user never submitted a request
		upgradeMessage = "You have not requested to become an admin."
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":             user.ID,
		"name":           user.Name,
		"email":          user.Email,
		"joined": time.Time(user.CreatedAt).Format(time.RFC3339),


		"membership":     user.Membership,
		"role":           user.Role,
		"upgrade_status": upgradeMessage,
	})
}

func UpdateProfileHandler(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	var payload struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	var user models.Person
	if err := database.DB.First(&user, userID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if payload.Name != "" {
		user.Name = strings.TrimSpace(payload.Name)
	}
	if payload.Password != "" {
		hashedPassword, _ := utils.HashPassword(payload.Password)
		user.Password = hashedPassword
	}
	user.UpdatedAt = utils.FormattedTime(time.Now())

	if err := database.DB.Save(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile", "details": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "profile updated"})
}

func DeleteAccountHandler(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	if err := database.DB.Delete(&models.Person{}, userID).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete account", "details": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "account deleted"})
}

func ForgotPasswordHandler(ctx *gin.Context) {
	var payload struct {
		Email string `json:"email"`
	}
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	var user models.Person
	if err := database.DB.Where("email = ?", payload.Email).First(&user).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	resetToken := utils.GenerateResetToken()
	user.ResetToken = resetToken
	user.ResetExpiry = time.Now().Add(15 * time.Minute)
	if err := database.DB.Save(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save reset token", "details": err.Error()})
		return
	}
	// Normally send via email; returning for demo
	ctx.JSON(http.StatusOK, gin.H{"message": "reset token generated", "reset_token": resetToken})
}

func ResetPasswordHandler(ctx *gin.Context) {
	var payload struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	var user models.Person
	if err := database.DB.Where("reset_token = ?", payload.Token).First(&user).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid reset token"})
		return
	}
	if time.Now().After(user.ResetExpiry) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "reset token expired"})
		return
	}
	hashedPassword, _ := utils.HashPassword(payload.NewPassword)
	user.Password = hashedPassword
	user.ResetToken = ""
	if err := database.DB.Save(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reset password", "details": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "password reset successful"})
}

// MarketplaceHandler returns all bots available in the marketplace
// func MarketplaceHandler(c *gin.Context) {
// 	var bots []models.Bot
// 	if err := database.DB.Find(&bots).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to fetch bots"})
// 		return
// 	}

// 	var botList []gin.H
// 	for _, b := range bots {
// 		// Trim the "uploads/" prefix because Gin serves it at /bots
// 		botPath := strings.TrimPrefix(b.HTMLFile, "uploads/")

// 		botList = append(botList, gin.H{
// 			"id":        b.ID,
// 			"name":      b.Name,
// 			"image":     b.Image,
// 			"price":     b.Price,
// 			"strategy":  b.Strategy,
// 			"bot_link":  fmt.Sprintf("http://localhost:8080/bots/%s", botPath),
// 		})
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "Marketplace bots fetched successfully",
// 		"bots":    botList,
// 	})
// }

// ServeBotHandler serves the bot HTML page by bot ID
func ServeBotHandler(c *gin.Context) {
	botID := c.Param("id")

	var bot models.Bot
	if err := database.DB.First(&bot, botID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "bot not found"})
		return
	}

	c.File(bot.HTMLFile) // serves the HTML file directly
}

// func MarketplaceHandler(c *gin.Context) {
// 	var bots []models.Bot
// 	if err := database.DB.Find(&bots).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to fetch bots"})
// 		return
// 	}

// 	var botList []gin.H
// 	for _, b := range bots {
// 		// Trim "uploads/" because Gin serves /uploads as static
// 		botPath := strings.TrimPrefix(b.HTMLFile, "uploads/")
// 		botList = append(botList, gin.H{
// 			"id":        b.ID,
// 			"name":      b.Name,
// 			"image":     b.Image,
// 			"price":     b.Price,
// 			"strategy":  b.Strategy,
// 			"bot_link":  fmt.Sprintf("http://localhost:8080/uploads/%s", botPath),
// 		})
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "Marketplace bots fetched successfully",
// 		"bots":    botList,
// 	})
// }

func MarketplaceHandler(c *gin.Context) {
	// Get logged-in user ID
	userIDVal, exists := c.Get("user_id")
	var userID uint
	if exists {
		userID = userIDVal.(uint)
	}

	// -----------------------------
	// 📄 Pagination params
	// -----------------------------
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	// -----------------------------
	// ⚙️ Query bots with pagination
	// -----------------------------
	var bots []models.Bot
	var total int64

	database.DB.Model(&models.Bot{}).Count(&total)

	if err := database.DB.
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&bots).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch bots"})
		return
	}

	// -----------------------------
	// 💖 Fetch user's favorites
	// -----------------------------
	favoriteMap := make(map[uint]bool)

	if userID != 0 {
		var favorites []models.Favorite
		if err := database.DB.
			Where("user_id = ?", userID).
			Find(&favorites).Error; err == nil {

			for _, f := range favorites {
				favoriteMap[f.BotID] = true
			}
		}
	}

	// -----------------------------
	// 🧱 Build custom bot list
	// -----------------------------
	var botList []gin.H
	for _, b := range bots {
		botPath := strings.TrimPrefix(b.HTMLFile, "uploads/")
		botList = append(botList, gin.H{
			"id":          b.ID,
			"name":        b.Name,
			"image":       b.Image,
			"price":       b.Price,
			"strategy":    b.Strategy,
			"status":      b.Status,
			"bot_link":    fmt.Sprintf("http://localhost:8080/uploads/%s", botPath),
			"is_favorite": favoriteMap[b.ID],
		})
	}

	// -----------------------------
	// ⭐ Sort: Favorites first
	// -----------------------------
	sort.SliceStable(botList, func(i, j int) bool {
		return botList[i]["is_favorite"].(bool) && !botList[j]["is_favorite"].(bool)
	})

	// -----------------------------
	// 📦 JSON response
	// -----------------------------
	c.JSON(http.StatusOK, gin.H{
		"message":     "Marketplace bots fetched successfully",
		"page":        page,
		"limit":       limit,
		"total_bots":  total,
		"total_pages": (total + int64(limit) - 1) / int64(limit),
		"bots":        botList,
	})
}

// // GET /api/user/me/favorites
// func GetUserFavorites(c *gin.Context) {
// 	db := c.MustGet("db").(*gorm.DB)
// 	userID := c.GetUint("user_id")

// 	var favorites []models.Favorite

// 	if err := db.Preload("Bot").Where("user_id = ?", userID).Find(&favorites).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch favorites"})
// 		return
// 	}

// 	var bots []models.Bot
// 	for _, fav := range favorites {
// 		bots = append(bots, fav.Bot)
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "Favorite bots retrieved successfully",
// 		"data":    bots,
// 	})
// }

// // POST /api/user/favorites/:bot_id
// func ToggleFavorite(c *gin.Context) {
// 	db := c.MustGet("db").(*gorm.DB)
// 	userID := c.GetUint("user_id")

// 	botID, err := strconv.Atoi(c.Param("bot_id"))
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bot ID"})
// 		return
// 	}

// 	var favorite models.Favorite
// 	err = db.Where("user_id = ? AND bot_id = ?", userID, botID).First(&favorite).Error

// 	if err == nil {
// 		db.Delete(&favorite)
// 		c.JSON(http.StatusOK, gin.H{
// 			"message": "Bot removed from favorites",
// 			"status":  "unfavorited",
// 		})
// 		return
// 	}

// 	if err != gorm.ErrRecordNotFound {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
// 		return
// 	}

// 	newFavorite := models.Favorite{UserID: userID, BotID: uint(botID)}
// 	if err := db.Create(&newFavorite).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add favorite"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "Bot added to favorites",
// 		"status":  "favorited",
// 	})
// }

// -----------------------------
// 🔹 GET /api/user/me/favorites
// -----------------------------
func GetUserFavorites(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	var favorites []models.Favorite

	// Load all favorites with their bots
	if err := database.DB.Preload("Bot").Where("user_id = ?", userID).Find(&favorites).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch favorites"})
		return
	}

	// Prepare bot list for response
	var bots []gin.H
	for _, fav := range favorites {
		b := fav.Bot
		bots = append(bots, gin.H{
			"id":          b.ID,
			"name":        b.Name,
			"image":       b.Image,
			"price":       b.Price,
			"strategy":    b.Strategy,
			"status":      b.Status,
			"bot_link":    b.HTMLFile,
			"is_favorite": true,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Favorite bots retrieved successfully",
		"data":    bots,
	})
}

// -----------------------------
// 🔹 POST /api/user/favorites/:bot_id
// -----------------------------
func ToggleFavorite(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	// Parse bot_id
	botID, err := strconv.Atoi(c.Param("bot_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid bot ID"})
		return
	}

	// Ensure bot exists
	var bot models.Bot
	if err := database.DB.First(&bot, botID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Bot not found"})
		return
	}

	var favorite models.Favorite
	err = database.DB.Where("user_id = ? AND bot_id = ?", userID, botID).First(&favorite).Error

	if err == nil {
		// Exists → unfavorite
		if err := database.DB.Delete(&favorite).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to unfavorite bot"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":     "Bot removed from favorites",
			"is_favorite": false,
		})
		return
	}

	if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error"})
		return
	}

	// Otherwise add to favorites
	newFavorite := models.Favorite{UserID: userID.(uint), BotID: uint(botID)}
	if err := database.DB.Create(&newFavorite).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to add favorite"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Bot added to favorites",
		"is_favorite": true,
	})
}

func DetectCountry(ip string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://ipapi.co/%s/country_name/", ip))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	country := strings.TrimSpace(string(body))
	if country == "" {
		country = "Unknown"
	}
	return country, nil
}

func SignupHandler(ctx *gin.Context) {
	var payload struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	payload.Email = strings.ToLower(strings.TrimSpace(payload.Email))
	payload.Name = strings.TrimSpace(payload.Name)
	payload.Password = strings.TrimSpace(payload.Password)

	if payload.Email == "" || payload.Name == "" || payload.Password == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "all fields required"})
		return
	}

	// 1️⃣ Detect country automatically from IP
	clientIP := ctx.ClientIP()
	country, err := DetectCountry(clientIP)
	if err != nil || country == "" {
		country = "Unknown"
	}

	// 2️⃣ Hash password
	hashedPassword, _ := utils.HashPassword(payload.Password)

	// 3️⃣ Create user with detected country
	user := models.Person{
		Name:      payload.Name,
		Email:     payload.Email,
		Password:  hashedPassword,
		Country:   country,
		CreatedAt: utils.FormattedTime(time.Now()),
UpdatedAt: utils.FormattedTime(time.Now()),

	}

	if err := database.DB.Create(&user).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "could not create user", "details": err.Error()})
		return
	}

	// 4️⃣ Generate JWT token
	token, _ := utils.GenerateToken(user.ID, user.Email)

	ctx.JSON(http.StatusOK, gin.H{
		"message": "signup successful",
		"token":   token,
		"country": country,
	})
}
