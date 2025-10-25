package handlers

import (
	"Api/database"
	"Api/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

// UpdateAdminBankDetails allows admins to update their bank details
func UpdateAdminBankDetails(ctx *gin.Context) {
	var input struct {
		BankCode      string `json:"bank_code"`
		AccountNumber string `json:"account_number"`
		AccountName   string `json:"account_name"`
	}
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input"})
		return
	}

	adminID := ctx.GetUint("admin_id") // Assuming middleware provides admin_id
	var admin models.Admin
	if err := database.DB.First(&admin, adminID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "Admin not found"})
		return
	}

	admin.BankCode = input.BankCode
	admin.AccountNumber = input.AccountNumber
	admin.AccountName = input.AccountName
	if err := database.DB.Save(&admin).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update bank details"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Bank details updated successfully"})
}
