package models

import "time"

type SalesHistory struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	BotID          uint      `json:"bot_id"`
	SellerID       uint      `json:"seller_id"`
	BuyerID        uint      `json:"buyer_id"`
	Amount         float64   `json:"amount"`
	TransactionRef string    `json:"transaction_ref"`
	SoldAt         time.Time `json:"sold_at"`
}
