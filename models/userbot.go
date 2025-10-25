package models

import "time"

type UserBot struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	UserID        uint       `json:"user_id"`
	BotID         uint       `json:"bot_id"`
	AccessType    string     `json:"access_type"` // "purchase" or "rent" (replaces Type)
	IsActive      bool       `json:"is_active"`   // replaces Active
	TransactionID *uint      `json:"transaction_id,omitempty"`
	Price         float64    `json:"price,omitempty"`
	ExpiryDate    *time.Time `json:"expiry_date,omitempty"` // for rentals
	PurchaseDate  time.Time  `json:"purchase_date"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	Type          string     `json:"type"`
	ResaleAllowed bool
}
