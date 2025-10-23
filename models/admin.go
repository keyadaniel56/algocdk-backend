package models

import (
	"gorm.io/gorm"
	"time"
)

type Admin struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	PersonID uint   `gorm:"uniqueIndex" json:"person_id"` // links to Person table
	Person   Person `gorm:"foreignKey:PersonID" json:"person"`

	// Payment / KYC / Paystack
	BankCode               string     `json:"bank_code"`
	AccountNumber          string     `json:"account_number"`
	AccountName            string     `json:"account_name"`
	PaystackSubaccountCode string     `json:"paystack_subaccount_code"`
	KYCStatus              string     `gorm:"default:unverified" json:"kyc_status"`
	VerifiedAt             *time.Time `json:"verified_at"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
