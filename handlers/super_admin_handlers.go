package handlers

import (
	"Api/database"
	"Api/models"
	"Api/utils"
	"crypto/subtle"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Retrieve super admin secret from environment variable

// POST /super/register
func SuperAdminRegisterHandler(ctx *gin.Context) {
	var superAdminSecret = os.Getenv("SUPER_ADMIN_SECRET")
	var payload struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Secret   string `json:"secret"`
	}

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Check secret using constant-time comparison
	if subtle.ConstantTimeCompare([]byte(payload.Secret), []byte(superAdminSecret)) == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Sanitize inputs
	payload.Name = strings.TrimSpace(payload.Name)
	payload.Email = strings.ToLower(strings.TrimSpace(payload.Email))
	payload.Password = strings.TrimSpace(payload.Password)

	// Validate inputs
	if payload.Name == "" || payload.Email == "" || payload.Password == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "all fields required"})
		return
	}
	if !govalidator.IsEmail(payload.Email) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid email format"})
		return
	}
	if len(payload.Password) < 8 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 8 characters"})
		return
	}

	// Check for existing email
	var existing models.Person
	if err := database.DB.Where("email = ?", payload.Email).First(&existing).Error; err == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "email already exists"})
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(payload.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// Create super admin
	superAdmin := models.Person{
		Name:      payload.Name,
		Email:     payload.Email,
		Password:  hashedPassword,
		Role:      "superadmin",
		CreatedAt: utils.FormattedTime(time.Now()),
		UpdatedAt: utils.FormattedTime(time.Now()),
	}

	if err := database.DB.Create(&superAdmin).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not create superadmin"})
		return
	}

	// Generate token
	token, err := utils.GenerateToken(superAdmin.ID, superAdmin.Email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "superadmin registered", "token": token})
}

// POST /super/login
func SuperAdminLoginHandler(ctx *gin.Context) {
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Sanitize inputs
	payload.Email = strings.ToLower(strings.TrimSpace(payload.Email))
	payload.Password = strings.TrimSpace(payload.Password)

	// Validate inputs
	if payload.Email == "" || payload.Password == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "email and password required"})
		return
	}
	if !govalidator.IsEmail(payload.Email) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid email format"})
		return
	}

	// Find super admin
	var superAdmin models.Person
	if err := database.DB.Where("email = ? AND role = ?", payload.Email, "superadmin").First(&superAdmin).Error; err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Check password
	if !utils.CheckPasswordHash(payload.Password, superAdmin.Password) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Generate token
	token, err := utils.GenerateToken(superAdmin.ID, superAdmin.Email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "login successful", "token": token})
}

func SuperAdminProfileHandler(ctx *gin.Context) {
	userIDInterface, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	var user models.Person
	if err := database.DB.First(&user, userID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if user.Role != "superadmin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
		"role":  user.Role,
	})
}

func SuperAdminDashboardHandler(ctx *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	userIDInterface, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	// Fetch superadmin from database
	var user models.Person
	if err := database.DB.First(&user, userID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Verify role
	if user.Role != "superadmin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Return dashboard data (expand as needed)
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Welcome to the SuperAdmin Dashboard",
		"user": gin.H{
			"id":        user.ID,
			"name":      user.Name,
			"email":     user.Email,
			"role":      user.Role,
			"joined":    user.CreatedAt,
			"updatedat": user.UpdatedAt,
		},
		// You can add additional stats, bot counts, etc. here
	})
}

// superadmin creates user, must provide password
func CreateUser(ctx *gin.Context) {
	var payload struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	payload.Email = strings.ToLower(strings.TrimSpace(payload.Email))
	payload.Name = strings.TrimSpace(payload.Name)
	payload.Password = strings.TrimSpace(payload.Password)

	if payload.Email == "" || payload.Name == "" || payload.Password == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "all fields required"})
		return
	}

	var existing models.Person
	if err := database.DB.Where("email = ?", payload.Email).First(&existing).Error; err == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user already exists"})
		return
	}

	hashedPassword, err := utils.HashPassword(payload.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	user := models.Person{
		Name:      payload.Name,
		Email:     payload.Email,
		Password:  hashedPassword,
		Role:      "User",
		CreatedAt: utils.FormattedTime(time.Now()),
		UpdatedAt: utils.FormattedTime(time.Now()),
	}

	if err := database.DB.Create(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user", "details": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "user created successfully",
		"user_id": user.ID,
		"email":   user.Email,
	})
}

// ✏️ Update user
func UpdateUser(ctx *gin.Context) {
	id := ctx.Param("id")
	var user models.Person

	if err := database.DB.First(&user, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var updateData models.Person
	if err := ctx.ShouldBindJSON(&updateData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if err := database.DB.Model(&user).Updates(updateData).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "User updated successfully", "user": user})
}

// ❌ Delete user
func DeleteUser(ctx *gin.Context) {
	id := ctx.Param("id")
	var user models.Person

	if err := database.DB.First(&user, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := database.DB.Delete(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// 📋 Get all users
func GetAllUsers(ctx *gin.Context) {
	var users []models.Person

	if err := database.DB.Find(&users).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"users": users})
}

// 🔍 Get single user by ID
func GetUserByID(ctx *gin.Context) {
	id := ctx.Param("id")
	var user models.Person

	if err := database.DB.First(&user, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"user": user})
}

//Admins

// func RequestAdminUpgrade(ctx *gin.Context) {
//     userID := ctx.GetUint("user_id") // matches your ProfileHandler
//     var user models.Person
//     if err := database.DB.First(&user, userID).Error; err != nil {
//         ctx.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
//         return
//     }

//     if user.Role != "USER" {
//         ctx.JSON(http.StatusBadRequest, gin.H{"message": "Only normal users can request upgrade"})
//         return
//     }

//     user.UpgradeRequestStatus = "pending"
//     user.UpdatedAt = time.Now()
//     database.DB.Save(&user)

//     ctx.JSON(http.StatusOK, gin.H{"message": "Upgrade request submitted"})
// }

func GetPendingRequests(ctx *gin.Context) {
	var users []models.Person
	if err := database.DB.Where("upgrade_request_status = ?", "pending").Find(&users).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch pending requests"})
		return
	}

	var result []gin.H
	for _, u := range users {
		result = append(result, gin.H{
			"id":    u.ID,
			"name":  u.Name,
			"email": u.Email,
			"role":  u.Role,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{"pending_requests": result})
}

// func ApproveUpgrade(ctx *gin.Context) {
// 	// Super Admin only
// 	id := ctx.Param("id")
// 	var user models.Person
// 	if err := database.DB.First(&user, id).Error; err != nil {
// 		ctx.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
// 		return
// 	}

// 	if user.UpgradeRequestStatus != "pending" {
// 		ctx.JSON(http.StatusBadRequest, gin.H{"message": "No pending upgrade request for this user"})
// 		return
// 	}

// 	user.Role = "ADMIN"
// 	user.UpgradeRequestStatus = "approved"
// 	user.UpdatedAt = time.Now()
// 	database.DB.Save(&user)

// 	ctx.JSON(http.StatusOK, gin.H{"message": "User promoted to admin"})
// }

func RejectUpgrade(ctx *gin.Context) {
	id := ctx.Param("id")
	var user models.Person
	if err := database.DB.First(&user, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	if user.UpgradeRequestStatus != "pending" {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "No pending upgrade request for this user"})
		return
	}

	user.UpgradeRequestStatus = "rejected"
	user.UpdatedAt = utils.FormattedTime(time.Now())

	database.DB.Save(&user)

	ctx.JSON(http.StatusOK, gin.H{"message": "User upgrade request rejected"})
}

func RequestAdminUpgrade(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var user models.Person
	if err := database.DB.First(&user, userID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	if user.Role != "USER" {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Only normal users can request upgrade"})
		return
	}

	user.UpgradeRequestStatus = "pending"
	user.UpdatedAt = utils.FormattedTime(time.Now())

	database.DB.Save(&user)

	// Notify all superadmins
	Hub.BroadcastToSuperAdmins(fmt.Sprintf("📩 New upgrade request from %s", user.Email))

	ctx.JSON(http.StatusOK, gin.H{"message": "Upgrade request submitted"})
}

func ApproveUpgrade(ctx *gin.Context) {
	id := ctx.Param("id")

	var user models.Person
	if err := database.DB.First(&user, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	if user.UpgradeRequestStatus != "pending" {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "No pending upgrade request for this user"})
		return
	}

	user.Role = "ADMIN"
	user.UpgradeRequestStatus = "approved"
	user.UpdatedAt = utils.FormattedTime(time.Now())

	database.DB.Save(&user)

	// 🔔 Send WebSocket notifications
	messageToUser := fmt.Sprintf("🎉 Congratulations %s! Your admin upgrade request was approved.", user.Name)
	messageToSuperAdmins := fmt.Sprintf("✅ Upgrade approved for user %s (%s)", user.Name, user.Email)

	Hub.SendToUser(user.ID, messageToUser)
	Hub.BroadcastToSuperAdmins(messageToSuperAdmins)

	ctx.JSON(http.StatusOK, gin.H{"message": "User promoted to admin"})
}

// func RequestAdminUpgrade(ctx *gin.Context) {
// 	userID := ctx.GetUint("user_id")

// 	var user models.Person
// 	if err := database.DB.First(&user, userID).Error; err != nil {
// 		ctx.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
// 		return
// 	}

// 	if user.Role != "USER" {
// 		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Only normal users can request upgrade"})
// 		return
// 	}

// 	user.UpgradeRequestStatus = "pending"
// 	user.UpdatedAt = time.Now()
// 	database.DB.Save(&user)

// 	// 🔔 Notify all connected superadmins immediately
// 	Hub.BroadcastToSuperAdmins(fmt.Sprintf("📩 New upgrade request from %s", user.Email))

// 	ctx.JSON(http.StatusOK, gin.H{"message": "Upgrade request submitted"})
// }

// func ApproveUpgrade(ctx *gin.Context) {
// 	id := ctx.Param("id")

// 	var user models.Person
// 	if err := database.DB.First(&user, id).Error; err != nil {
// 		ctx.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
// 		return
// 	}

// 	if user.UpgradeRequestStatus != "pending" {
// 		ctx.JSON(http.StatusBadRequest, gin.H{"message": "No pending upgrade request for this user"})
// 		return
// 	}

// 	user.Role = "ADMIN"
// 	user.UpgradeRequestStatus = "approved"
// 	user.UpdatedAt = time.Now()
// 	database.DB.Save(&user)

// 	// 🔔 Send WebSocket notifications
// 	messageToUser := fmt.Sprintf("🎉 Congratulations %s! Your admin upgrade request was approved.", user.Name)
// 	messageToSuperAdmins := fmt.Sprintf("✅ Upgrade approved for user %s (%s)", user.Name, user.Email)

// 	Hub.SendToUser(user.ID, messageToUser)
// 	Hub.BroadcastToSuperAdmins(messageToSuperAdmins)

// 	ctx.JSON(http.StatusOK, gin.H{"message": "User promoted to admin"})
// }

func GetAllAdmins(ctx *gin.Context) {
	var admins []models.Person

	if err := database.DB.Where("role IN ?", []string{"Admin", "Senior Admin", "Super Admin", "ADMIN"}).Find(&admins).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch admins"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"admins": admins})
}

func ToggleAdminStatus(ctx *gin.Context) {
	id := ctx.Param("id")
	var admin models.Person

	if err := database.DB.First(&admin, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Admin not found"})
		return
	}

	if admin.UpgradeRequestStatus == "Active" || admin.UpgradeRequestStatus == "" {
		admin.UpgradeRequestStatus = "Suspended"
	} else {
		admin.UpgradeRequestStatus = "Active"
	}

	if err := database.DB.Save(&admin).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"admin": admin})
}

func DeleteAdmin(ctx *gin.Context) {
	id := ctx.Param("id")
	var admin models.Person

	if err := database.DB.First(&admin, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Admin not found"})
		return
	}

	if err := database.DB.Delete(&admin).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete admin"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Admin deleted successfully"})
}

func CreateAdmin(ctx *gin.Context) {
	var payload struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	payload.Email = strings.ToLower(strings.TrimSpace(payload.Email))
	payload.Name = strings.TrimSpace(payload.Name)
	payload.Password = strings.TrimSpace(payload.Password)

	if payload.Email == "" || payload.Name == "" || payload.Password == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "all fields required"})
		return
	}

	// 🔍 Check if email already exists
	var existing models.Person
	if err := database.DB.Where("email = ?", payload.Email).First(&existing).Error; err == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user already exists"})
		return
	}

	// 🔐 Hash password
	hashedPassword, err := utils.HashPassword(payload.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// 👤 Create Person
	person := models.Person{
		Name:      payload.Name,
		Email:     payload.Email,
		Password:  hashedPassword,
		Role:      "Admin",
		CreatedAt: utils.FormattedTime(time.Now()),
		UpdatedAt: utils.FormattedTime(time.Now()),
	}

	if err := database.DB.Create(&person).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create person", "details": err.Error()})
		return
	}

	// 🧩 Create Admin linked to Person
	admin := models.Admin{
		PersonID: person.ID,
	}
	if err := database.DB.Create(&admin).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create admin record", "details": err.Error()})
		return
	}

	// ✅ Response
	ctx.JSON(http.StatusCreated, gin.H{
		"message":    "admin created successfully",
		"person_id":  person.ID,
		"admin_id":   admin.ID,
		"email":      person.Email,
		"role":       person.Role,
		"created_at": person.CreatedAt,
		"updated_at": person.UpdatedAt,
	})
}

func UpdateAdmin(ctx *gin.Context) {
	id := ctx.Param("id")
	var admin models.Person

	if err := database.DB.First(&admin, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Admin not found"})
		return
	}

	var input struct {
		Name    string `json:"name"`
		Email   string `json:"email" binding:"email"`
		Role    string `json:"role"`
		Phone   string `json:"phone"`
		Country string `json:"country"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Name != "" {
		admin.Name = input.Name
	}
	if input.Email != "" {
		admin.Email = input.Email
	}
	if input.Role != "" {
		admin.Role = input.Role
	}
	if input.Phone != "" {
		admin.Phone = input.Phone
	}
	if input.Country != "" {
		admin.Country = input.Country
	}

	if err := database.DB.Save(&admin).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update admin"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"admin": admin})
}

func GetBotsHandler(ctx *gin.Context) {
	var bots []models.Bot

	// Get all bots
	if err := database.DB.Find(&bots).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bots"})
		return
	}

	var response []gin.H

	for _, bot := range bots {
		// 🧩 Try to get the owner (admin)
		var owner models.Person
		if err := database.DB.Select("id", "name", "email").First(&owner, bot.OwnerID).Error; err != nil {
			log.Printf("⚠️ Owner not found for bot %d (owner_id=%d): %v", bot.ID, bot.OwnerID, err)
			owner = models.Person{
				Name:  "Unknown",
				Email: "N/A",
			}
		}

		// 🧩 Count subscribers safely
		var subscriberCount int64
		if err := database.DB.Table("bot_users").Where("bot_id = ?", bot.ID).Count(&subscriberCount).Error; err != nil {
			log.Printf("⚠️ Failed to count subscribers for bot %d: %v", bot.ID, err)
			subscriberCount = 0
		}

		// ✅ Build response entry
		response = append(response, gin.H{
			"id":                 bot.ID,
			"name":               bot.Name,
			"subscriptionType":   bot.SubscriptionType,
			"price":              bot.Price,
			"subscriberCount":    subscriberCount,
			"subscriptionExpiry": bot.SubscriptionExpiry,
			"status":             bot.Status,
			"owner": gin.H{
				"id":    owner.ID,
				"name":  owner.Name,
				"email": owner.Email,
			},
		})
	}

	ctx.JSON(http.StatusOK, gin.H{"bots": response})
}

// ScanAllBotsHandler scans all bot files to ensure they contain only the allowed app ID
func ScanAllBotsHandler(c *gin.Context) {
	rootDir := "./uploads"
	var invalidBots []map[string]interface{}

	// Regex to match app_id or appId assignments
	re := regexp.MustCompile(`(?i)(app[_]?id)\s*[:=]\s*['"]?(\d+)['"]?`)

	// Walk through all files in uploads
	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Only scan .html or .js files
		if !d.IsDir() && (filepath.Ext(path) == ".html" || filepath.Ext(path) == ".js") {
			bytes, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			matches := re.FindAllStringSubmatch(string(bytes), -1)
			for _, m := range matches {
				if len(m) > 2 && m[2] != "1089" {
					// Try to find bot record
					var bot models.Bot
					err := database.DB.Where("html_file = ?", path).First(&bot).Error
					if errors.Is(err, gorm.ErrRecordNotFound) {
						bot = models.Bot{} // fallback
					} else if err != nil {
						return err
					}

					// Try to find owner
					var owner models.Person
					// ownerName := "Unknown Owner"
					if bot.OwnerID != 0 {
						err := database.DB.Where("id = ?", bot.OwnerID).First(&owner).Error
						if err == nil {
							// ownerName = owner.Name
						}
					}

					invalidBots = append(invalidBots, map[string]interface{}{
						"bot_name": bot.Name,
						"owner":    bot.OwnerID,
						"file":     path,
						"app_id":   m[2],
						"name":     bot.Name,
						"filename": bot.HTMLFile,
					})
				}
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to scan bots",
			"details": err.Error(),
		})
		return
	}

	if len(invalidBots) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"message":      fmt.Sprintf("Scan completed. Found %d bots with invalid App IDs.", len(invalidBots)),
			"invalid_bots": invalidBots,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All bots are valid"})
}

// GetAllTransactions retrieves all transactions for superadmin
func GetAllTransactions(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	// Verify superadmin role
	var user models.Person
	if err := database.DB.First(&user, userIDUint).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if user.Role != "superadmin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var transactions []models.Transaction
	if err := database.DB.Order("created_at DESC").Find(&transactions).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error fetching all transactions",
			"details": err.Error(),
		})
		return
	}

	totalSales := 0.0
	totalCompanyShare := 0.0
	totalAdminShare := 0.0
	byAdmin := make(map[uint][]models.Transaction)
	for _, tx := range transactions {
		if tx.PaymentType == "purchase" || tx.PaymentType == "rent" {
			totalSales += tx.Amount
			totalCompanyShare += tx.CompanyShare
			totalAdminShare += tx.AdminShare
		}
		byAdmin[tx.AdminID] = append(byAdmin[tx.AdminID], tx)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"transactions":        transactions,
		"total_sales":         totalSales,
		"total_company_share": totalCompanyShare,
		"total_admin_share":   totalAdminShare,
		"total_transactions":  len(transactions),
		"by_admin":            byAdmin,
	})
}

// RecordTransaction creates a new transaction
func RecordTransaction(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	// Verify admin role
	var user models.Person
	if err := database.DB.First(&user, userIDUint).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if user.Role != "ADMIN" && user.Role != "superadmin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var input struct {
		UserID         uint    `json:"user_id" binding:"required"`
		BotID          uint    `json:"bot_id" binding:"required"`
		Amount         float64 `json:"amount" binding:"required,gt=0"`
		CompanyShare   float64 `json:"company_share" binding:"gte=0"`
		AdminShare     float64 `json:"admin_share" binding:"gte=0"`
		Reference      string  `json:"reference" binding:"required"`
		Status         string  `json:"status" binding:"required,oneof=pending success failed"`
		PaymentChannel string  `json:"payment_channel" binding:"required"`
		PaymentType    string  `json:"payment_type" binding:"required,oneof=purchase rent"`
		Description    string  `json:"description"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid input",
			"details": err.Error(),
		})
		return
	}

	// Additional validation using govalidator
	if !govalidator.IsIn(input.Status, "pending", "success", "failed") ||
		!govalidator.IsIn(input.PaymentType, "purchase", "rent") {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid status or payment type",
		})
		return
	}

	transaction := models.Transaction{
		UserID:         input.UserID,
		AdminID:        userIDUint,
		BotID:          input.BotID,
		Amount:         input.Amount,
		CompanyShare:   input.CompanyShare,
		AdminShare:     input.AdminShare,
		Reference:      input.Reference,
		Status:         input.Status,
		PaymentChannel: input.PaymentChannel,
		PaymentType:    input.PaymentType,
		Description:    input.Description,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := database.DB.Create(&transaction).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error recording transaction",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, transaction)
}
