package models

import (
	"time"
)

// Sale represents a record of a bot sale or rental transaction
type Sale struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	BotID     uint      `json:"bot_id"`
	SellerID  uint      `json:"seller_id"`
	BuyerID   uint      `json:"buyer_id"`
	Amount    float64   `json:"amount"`
	SaleType  string    `json:"sale_type"` // "purchase" or "rent"
	SaleDate  time.Time `json:"sale_date"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
