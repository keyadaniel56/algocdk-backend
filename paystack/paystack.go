package paystack

import (
	"Api/database"
	"Api/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ✅ Paystack Verification Response Struct
type PaystackVerifyResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Status     string `json:"status"`
		Amount     int    `json:"amount"`
		Reference  string `json:"reference"`
		Subaccount string `json:"subaccount"`
		Fees       int    `json:"fees"`
		Customer   struct {
			Email string `json:"email"`
		} `json:"customer"`
	} `json:"data"`
}

// ✅ PAYMENT INITIALIZATION
func InitializePayment(ctx *gin.Context) {
	var input struct {
		Amount      float64 `json:"amount"`
		AdminID     uint    `json:"admin_id"`
		BotID       uint    `json:"bot_id"`
		PaymentType string  `json:"payment_type"` // "purchase" or "rent"
		Description string  `json:"description"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input"})
		return
	}

	userID := ctx.GetUint("user_id")

	// ✅ Fetch user and admin info
	var user models.Person
	if err := database.DB.First(&user, userID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	var admin models.Admin
	if err := database.DB.First(&admin, input.AdminID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "Admin not found"})
		return
	}

	// ✅ Prevent double purchase
	if input.PaymentType == "purchase" {
		var existing models.Transaction
		if err := database.DB.
			Where("user_id = ? AND bot_id = ? AND payment_type = ? AND status = ?", userID, input.BotID, "purchase", "success").
			First(&existing).Error; err == nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "You already purchased this bot"})
			return
		}
	}

	// ✅ Determine company share %
	var companyPercent float64
	switch input.PaymentType {
	case "purchase":
		companyPercent = 0.30 // 30% for purchases
	case "rent":
		companyPercent = 0.20 // 20% for rentals
	default:
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Invalid payment type"})
		return
	}

	companyShare := input.Amount * companyPercent
	adminShare := input.Amount - companyShare
	reference := fmt.Sprintf("ALG_%d_%d", userID, time.Now().Unix())

	// ✅ Build Paystack payload
	payload := map[string]interface{}{
		"email":              user.Email,
		"amount":             int(input.Amount * 100),
		"reference":          reference,
		"callback_url":       os.Getenv("PAYSTACK_CALLBACK_URL"),
		"subaccount":         admin.PaystackSubaccountCode,
		"bearer":             "subaccount",
		"transaction_charge": int(companyShare * 100),
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://api.paystack.co/transaction/initialize", bytes.NewReader(body))
	req.Header.Add("Authorization", "Bearer "+os.Getenv("PAYSTACK_SECRET_KEY"))
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Paystack request failed"})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	if result["status"] != true {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Failed to initialize payment", "error": result["message"]})
		return
	}

	// ✅ Save pending transaction
	transaction := models.Transaction{
		UserID:         userID,
		AdminID:        input.AdminID,
		BotID:          input.BotID,
		Amount:         input.Amount,
		CompanyShare:   companyShare,
		AdminShare:     adminShare,
		Status:         "pending",
		Reference:      reference,
		PaymentChannel: "Paystack",
		PaymentType:    input.PaymentType,
		Description:    input.Description,
		CreatedAt:      time.Now(),
	}

	database.DB.Create(&transaction)

	data := result["data"].(map[string]interface{})
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Payment initialized",
		"data":    data,
	})
}

// ✅ VERIFY PAYMENT CALLBACK
func VerifyPayment(ctx *gin.Context) {
	reference := ctx.Query("reference")
	if reference == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Missing reference"})
		return
	}

	url := fmt.Sprintf("https://api.paystack.co/transaction/verify/%s", reference)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Bearer "+os.Getenv("PAYSTACK_SECRET_KEY"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Verification failed"})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result PaystackVerifyResponse
	json.Unmarshal(body, &result)

	if !result.Status || result.Data.Status != "success" {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Payment verification failed"})
		return
	}

	var tx models.Transaction
	if err := database.DB.Where("reference = ?", reference).First(&tx).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "Transaction not found"})
		return
	}

	// Update transaction status
	tx.Status = "success"
	database.DB.Save(&tx)

	// Grant bot access based on payment type
	switch tx.PaymentType {
	case "purchase":
		// Permanent ownership
		var existing models.UserBot
		err := database.DB.Where("user_id = ? AND bot_id = ?", tx.UserID, tx.BotID).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			database.DB.Create(&models.UserBot{
				UserID:       tx.UserID,
				BotID:        tx.BotID,
				AccessType:   "purchase",
				IsActive:     true,
				PurchaseDate: time.Now(),
			})
		}
	case "rent":
		// Limited-time access
		expiry := time.Now().Add(30 * 24 * time.Hour) // 30 days rental
		database.DB.Create(&models.UserBot{
			UserID:       tx.UserID,
			BotID:        tx.BotID,
			AccessType:   "rent",
			IsActive:     true,
			ExpiryDate:   &expiry,
			PurchaseDate: time.Now(),
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Payment verified successfully",
		"data":    tx,
	})
}

// ✅ Create Paystack Subaccount for Admin
func CreatePaystackSubaccount(admin *models.Admin) error {
	payload := map[string]interface{}{
		"business_name":     admin.Person.Name,
		"settlement_bank":   admin.BankCode,
		"account_number":    admin.AccountNumber,
		"percentage_charge": 10, // Default company cut (10%) — can be dynamic later
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://api.paystack.co/subaccount", bytes.NewReader(body))
	req.Header.Add("Authorization", "Bearer "+os.Getenv("PAYSTACK_SECRET_KEY"))
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Paystack error: %s", string(respBody))
	}

	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	if data, ok := result["data"].(map[string]interface{}); ok {
		admin.PaystackSubaccountCode = data["subaccount_code"].(string)
		return database.DB.Save(&admin).Error
	}

	return fmt.Errorf("failed to parse Paystack response")
}

// PaystackCallback handles webhook calls from Paystack
func PaystackCallback(ctx *gin.Context) {
	// Read body
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
		return
	}

	var event map[string]interface{}
	if err := json.Unmarshal(body, &event); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Invalid JSON"})
		return
	}

	// Check event type
	if event["event"] != "charge.success" {
		ctx.JSON(http.StatusOK, gin.H{"message": "Event ignored"})
		return
	}

	// Get payment reference
	data := event["data"].(map[string]interface{})
	ref := data["reference"].(string)

	var tx models.Transaction
	if err := database.DB.Where("reference = ?", ref).First(&tx).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "Transaction not found"})
		return
	}

	// Update transaction status
	tx.Status = "success"
	database.DB.Save(&tx)

	// Grant bot access based on payment type
	switch tx.PaymentType {
	case "purchase":
		var existing models.UserBot
		if err := database.DB.Where("user_id = ? AND bot_id = ?", tx.UserID, tx.BotID).First(&existing).Error; err == gorm.ErrRecordNotFound {
			database.DB.Create(&models.UserBot{
				UserID:       tx.UserID,
				BotID:        tx.BotID,
				AccessType:   "purchase",
				IsActive:     true,
				PurchaseDate: time.Now(),
			})
		}
	case "rent":
		expiry := time.Now().Add(30 * 24 * time.Hour) // 30 days rental
		database.DB.Create(&models.UserBot{
			UserID:       tx.UserID,
			BotID:        tx.BotID,
			AccessType:   "rent",
			IsActive:     true,
			ExpiryDate:   &expiry,
			PurchaseDate: time.Now(),
		})
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Payment processed"})
}
