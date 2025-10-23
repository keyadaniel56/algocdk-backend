package handlers

import (
	"Api/database"
	"Api/models"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"

	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AdminDashboardHandler shows all bots created by the logged-in admin
func AdminDashboardHandler(ctx *gin.Context) {
	// Get user_id from context
	userIDInterface, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "invalid user_id"})
		return
	}

	// Fetch admin from DB
	var admin models.Person
	if err := database.DB.First(&admin, userID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	roleLower := strings.ToLower(admin.Role)
	if roleLower != "admin" && roleLower != "superadmin" {
		ctx.JSON(http.StatusForbidden, gin.H{"message": "Access denied"})
		return
	}

	// Fetch all bots created by this admin
	var bots []models.Bot
	if err := database.DB.Where("owner_id = ?", admin.ID).Find(&bots).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "failed to fetch bots"})
		return
	}

	// Prepare bot data
	botData := []gin.H{}
	for _, bot := range bots {
		// You can also fetch users of this bot here if you have a bot_users table
		botData = append(botData, gin.H{
			"id":         bot.ID,
			"name":       bot.Name,
			"strategy":   bot.Strategy,
			"price":      bot.Price,
			"link":       "https://yourfrontend.com/bots/" + strconv.FormatUint(uint64(bot.ID), 10),
			"image":      bot.Image,
			"html_file":  bot.HTMLFile,
			"created_at": bot.CreatedAt.Format(time.RFC3339),
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"admin_id":   admin.ID,
		"admin_name": admin.Name,
		"role":       admin.Role,
		"bots":       botData,
	})
}

type CreateBotPayload struct {
	Name     string  `json:"name"`
	HTMLFile string  `json:"html_file"`
	Image    string  `json:"image"`
	Price    float64 `json:"price"`
	Strategy string  `json:"strategy"`
}

// CreateBotHandler allows an admin to create a bot

// GET /api/admin/bots
func GetAdminBotsHandler(ctx *gin.Context) {
	userIDInterface, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	roleInterface, _ := ctx.Get("role")
	userID, ok := userIDInterface.(uint)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Invalid user_id"})
		return
	}

	role, ok := roleInterface.(string)
	if !ok || (strings.ToLower(role) != "admin" && strings.ToLower(role) != "superadmin") {
		ctx.JSON(http.StatusForbidden, gin.H{"message": "Access denied"})
		return
	}

	// Fetch bots created by this admin
	var bots []models.Bot
	if err := database.DB.Where("owner_id = ?", userID).Find(&bots).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch bots", "error": err.Error()})
		return
	}

	// Prepare response
	var botData []gin.H
	for _, bot := range bots {
		// Fetch users of this bot
		var botUsers []struct {
			ID    uint
			Name  string
			Email string
		}
		database.DB.Table("people").
			Select("people.id, people.name, people.email").
			Joins("JOIN bot_users ON bot_users.user_id = people.id").
			Where("bot_users.bot_id = ?", bot.ID).
			Scan(&botUsers)

		userList := []gin.H{}
		for _, u := range botUsers {
			userList = append(userList, gin.H{
				"id":    u.ID,
				"name":  u.Name,
				"email": u.Email,
			})
		}

		// Bot link (frontend route)
		botLink := "/bots/" + string(bot.ID)

		botData = append(botData, gin.H{
			"id":     bot.ID,
			"name":   bot.Name,
			"status": bot.Status, // ensure Status exists in your model
			"link":   botLink,
			"users":  userList,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"admin_id": userID,
		"role":     role,
		"bots":     botData,
	})
}

// func CreateBotHandler(c *gin.Context) {
// 	// Get user_id from context
// 	userIDInterface, exists := c.Get("user_id")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
// 		return
// 	}
// 	userID, ok := userIDInterface.(uint)
// 	if !ok {
// 		c.JSON(http.StatusInternalServerError, gin.H{"message": "invalid user_id"})
// 		return
// 	}

// 	// Parse form values
// 	name := c.PostForm("name")
// 	priceStr := c.PostForm("price")
// 	strategy := c.PostForm("strategy")

// 	if name == "" || priceStr == "" || strategy == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required fields (name, price, strategy)"})
// 		return
// 	}

// 	price, err := strconv.ParseFloat(priceStr, 64)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid price"})
// 		return
// 	}

// 	now := time.Now()
// 	baseFolder := fmt.Sprintf("uploads/user_%d/%d/%02d/%02d", userID, now.Year(), now.Month(), now.Day())

// 	// Helper function to save uploaded file
// 	saveFile := func(fileHeader *multipart.FileHeader, folder string) (string, error) {
// 		if err := os.MkdirAll(folder, os.ModePerm); err != nil {
// 			return "", err
// 		}
// 		newName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(fileHeader.Filename))
// 		path := filepath.Join(folder, newName)
// 		if err := c.SaveUploadedFile(fileHeader, path); err != nil {
// 			return "", err
// 		}
// 		return path, nil
// 	}

// 	// Save HTML file
// 	htmlFile, err := c.FormFile("html_file")
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "html_file required"})
// 		return
// 	}
// 	htmlPath, err := saveFile(htmlFile, filepath.Join(baseFolder, "html"))
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save html file"})
// 		return
// 	}

// 	// Save image file
// 	imageFile, err := c.FormFile("image")
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "image required"})
// 		return
// 	}
// 	imagePath, err := saveFile(imageFile, filepath.Join(baseFolder, "images"))
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save image file"})
// 		return
// 	}

// 	// Create bot record
// 	bot := models.Bot{
// 		Name:      name,
// 		HTMLFile:  htmlPath,
// 		Image:     imagePath,
// 		Price:     price,
// 		Strategy:  strategy,
// 		OwnerID:   userID,
// 		CreatedAt: now,
// 		UpdatedAt: now,
// 	}

// 	if err := database.DB.Create(&bot).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save bot"})
// 		return
// 	}

// 	// Generate bot link for frontend
// 	botLink := fmt.Sprintf("https://yourfrontend.com/bots/%d", bot.ID)

// 	c.JSON(http.StatusCreated, gin.H{
// 		"message":  "Bot created successfully",
// 		"bot_id":   bot.ID,
// 		"bot_link": botLink,
// 	})
// }

// -------------------------
// Create Bot (already working)
// -------------------------
// func CreateBotHandler(c *gin.Context) { ... }

// -------------------------
// Update Bot
// -------------------------
func UpdateBotHandler(c *gin.Context) {
	userID := c.GetUint("user_id")

	// Fetch bot
	var bot models.Bot
	if err := database.DB.First(&bot, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}

	// Check ownership
	if bot.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "not your bot"})
		return
	}

	// Optional fields
	name := c.PostForm("name")
	priceStr := c.PostForm("price")
	strategy := c.PostForm("strategy")

	if name != "" {
		bot.Name = name
	}
	if strategy != "" {
		bot.Strategy = strategy
	}
	if priceStr != "" {
		if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
			bot.Price = price
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid price"})
			return
		}
	}

	now := time.Now()
	baseFolder := fmt.Sprintf("uploads/user_%d/%d/%02d/%02d", userID, now.Year(), now.Month(), now.Day())

	// Helper to save file
	saveFile := func(file *multipart.FileHeader, folder string) (string, error) {
		os.MkdirAll(folder, os.ModePerm)
		newName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename))
		path := filepath.Join(folder, newName)
		if err := c.SaveUploadedFile(file, path); err != nil {
			return "", err
		}
		return path, nil
	}

	// Update HTML file if provided
	if file, err := c.FormFile("html_file"); err == nil {
		if htmlPath, err := saveFile(file, filepath.Join(baseFolder, "html")); err == nil {
			bot.HTMLFile = htmlPath
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save html file"})
			return
		}
	}

	// Update image if provided
	if file, err := c.FormFile("image"); err == nil {
		if imagePath, err := saveFile(file, filepath.Join(baseFolder, "images")); err == nil {
			bot.Image = imagePath
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save image file"})
			return
		}
	}

	// Update timestamp
	bot.UpdatedAt = time.Now()

	// Save changes
	if err := database.DB.Save(&bot).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update bot"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "bot updated", "bot": bot})
}

// -------------------------
// Delete Bot
// -------------------------
func DeleteBotHandler(c *gin.Context) {
	userID := c.GetUint("user_id")
	botIDStr := c.Param("id")
	botID, err := strconv.ParseUint(botIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bot ID"})
		return
	}

	var bot models.Bot
	if err := database.DB.First(&bot, botID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}

	if bot.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "not allowed to delete this bot"})
		return
	}

	if err := database.DB.Delete(&bot).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete bot"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "bot deleted successfully"})
}

// -------------------------
// List Users using a Bot
// -------------------------
func BotUsersHandler(c *gin.Context) {
	userID := c.GetUint("user_id")
	botIDStr := c.Param("id")
	botID, err := strconv.ParseUint(botIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bot ID"})
		return
	}

	var bot models.Bot
	if err := database.DB.First(&bot, botID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}

	if bot.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "not allowed to view this bot's users"})
		return
	}

	// Get users from bot_users table (assuming many-to-many relationship)
	var users []models.Person
	database.DB.Joins("JOIN bot_users ON bot_users.user_id = people.id").
		Where("bot_users.bot_id = ?", bot.ID).
		Find(&users)

	userList := []gin.H{}
	for _, u := range users {
		userList = append(userList, gin.H{"id": u.ID, "name": u.Name, "email": u.Email})
	}

	c.JSON(http.StatusOK, gin.H{"bot_id": bot.ID, "bot_name": bot.Name, "users": userList})
}

// -------------------------
// Remove user from Bot
// -------------------------
func RemoveUserFromBotHandler(c *gin.Context) {
	userID := c.GetUint("user_id")
	botIDStr := c.Param("bot_id")
	botID, err := strconv.ParseUint(botIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bot ID"})
		return
	}

	var bot models.Bot
	if err := database.DB.First(&bot, botID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}

	if bot.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "not allowed to modify this bot"})
		return
	}

	// user_id to remove
	removeUserIDStr := c.PostForm("user_id")
	removeUserID, err := strconv.ParseUint(removeUserIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	// Delete from bot_users table
	if err := database.DB.Exec("DELETE FROM bot_users WHERE bot_id = ? AND user_id = ?", bot.ID, removeUserID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user removed from bot"})
}

func ListAdminBotsHandler(c *gin.Context) {
	// Get user_id from context
	userID := c.GetUint("user_id")

	// Fetch all bots created by this admin
	var bots []models.Bot
	if err := database.DB.Where("owner_id = ?", userID).Find(&bots).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch bots"})
		return
	}

	var botList []gin.H
	for _, bot := range bots {
		// Fetch users associated with this bot
		var users []models.Person
		database.DB.Joins("JOIN bot_users ON bot_users.user_id = people.id").
			Where("bot_users.bot_id = ?", bot.ID).
			Find(&users)

		userList := []gin.H{}
		for _, u := range users {
			userList = append(userList, gin.H{
				"id":    u.ID,
				"name":  u.Name,
				"email": u.Email,
			})
		}

		botList = append(botList, gin.H{
			"id":       bot.ID,
			"name":     bot.Name,
			"price":    bot.Price,
			"strategy": bot.Strategy,
			"html":     bot.HTMLFile,
			"image":    bot.Image,
			"users":    userList,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"admin_id": userID,
		"bots":     botList,
	})
}

func CreateBotHandler(c *gin.Context) {
	// Get user_id from context
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}
	userID, ok := userIDInterface.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "invalid user_id"})
		return
	}

	// Parse form values
	name := c.PostForm("name")
	priceStr := c.PostForm("price")
	rentPriceStr := c.PostForm("rent_price")
	strategy := c.PostForm("strategy")
	subscriptionType := c.PostForm("subscription_type")
	description := c.PostForm("description")
	category := c.PostForm("category")
	version := c.PostForm("version")

	if name == "" || priceStr == "" || strategy == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required fields (name, price, strategy)"})
		return
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid price"})
		return
	}

	var rentPrice float64
	if rentPriceStr != "" {
		rentPrice, err = strconv.ParseFloat(rentPriceStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rent_price"})
			return
		}
	}

	now := time.Now()
	baseFolder := fmt.Sprintf("uploads/user_%d/%d/%02d/%02d", userID, now.Year(), now.Month(), now.Day())

	// Helper function to save uploaded file
	saveFile := func(fileHeader *multipart.FileHeader, folder string) (string, error) {
		if err := os.MkdirAll(folder, os.ModePerm); err != nil {
			return "", err
		}
		newName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(fileHeader.Filename))
		path := filepath.Join(folder, newName)
		if err := c.SaveUploadedFile(fileHeader, path); err != nil {
			return "", err
		}
		return path, nil
	}

	// Save HTML file
	htmlFile, err := c.FormFile("html_file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "html_file required"})
		return
	}
	htmlPath, err := saveFile(htmlFile, filepath.Join(baseFolder, "html"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save html file"})
		return
	}

	// Save image file
	imageFile, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image required"})
		return
	}
	imagePath, err := saveFile(imageFile, filepath.Join(baseFolder, "images"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save image file"})
		return
	}

	// Create bot record
	bot := models.Bot{
		Name:             name,
		HTMLFile:         htmlPath,
		Image:            imagePath,
		Price:            price,
		RentPrice:        rentPrice,
		Strategy:         strategy,
		OwnerID:          userID,
		CreatedAt:        now,
		UpdatedAt:        now,
		Status:           "inactive", // Admin can activate later
		SubscriptionType: subscriptionType,
		Description:      description,
		Category:         category,
		Version:          version,
	}

	if err := database.DB.Create(&bot).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save bot"})
		return
	}

	// Generate bot link for frontend
	botLink := fmt.Sprintf("https://yourfrontend.com/bots/%d", bot.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Bot created successfully",
		"bot_id":   bot.ID,
		"bot_link": botLink,
	})
}



// GET /api/admin/profile
func AdminProfileHandler(ctx *gin.Context) {
	emailValue, exists := ctx.Get("email") // ✅ fixed key
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized, missing token or claims"})
		return
	}

	email, ok := emailValue.(string)
	if !ok || email == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email in token"})
		return
	}

	var person models.Person
	if err := database.DB.Where("email = ?", email).First(&person).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "admin not found"})
		return
	}

	var admin models.Admin
	if err := database.DB.Where("person_id = ?", person.ID).First(&admin).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "admin record not found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Admin profile loaded successfully",
		"admin": gin.H{
			"id":                     admin.ID,
			"person_id":              admin.PersonID,
			"name":                   person.Name,
			"email":                  person.Email,
			"role":                   person.Role,
			"country":                person.Country,
			"phone":                  person.Phone,
			"membership":             person.Membership,
			"total_profits":          person.TotalProfits,
			"active_bots":            person.ActiveBots,
			"total_trades":           person.TotalTrades,
			"bank_code":              admin.BankCode,
			"account_number":         admin.AccountNumber,
			"account_name":           admin.AccountName,
			"paystack_subaccount":    admin.PaystackSubaccountCode,
			"kyc_status":             admin.KYCStatus,
			"verified_at":            admin.VerifiedAt,
			"created_at":             person.CreatedAt,
			"updated_at":             person.UpdatedAt,
			"subscription_expiry":    person.SubscriptionExpiry,
			"upgrade_request_status": person.UpgradeRequestStatus,
		},
	})
}
