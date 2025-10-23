package models

import "time"

type Transaction struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	UserID         uint      `json:"user_id"`
	AdminID        uint      `json:"admin_id"`
	BotID          uint      `json:"bot_id"`
	Amount         float64   `json:"amount"`
	CompanyShare   float64   `json:"company_share"`
	AdminShare     float64   `json:"admin_share"`
	Reference      string    `json:"reference"`
	Status         string    `json:"status"`          // "pending", "success", "failed"
	PaymentChannel string    `json:"payment_channel"` // e.g. "Paystack"
	PaymentType    string    `json:"payment_type"`    // "purchase" or "rent"
	Description    string    `json:"description"`
	CreatedAt      time.Time `json:"created_at"`
}
