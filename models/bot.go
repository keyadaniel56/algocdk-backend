package models

import "time"

type Bot struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	HTMLFile  string    `json:"html_file"`
	Image     string    `json:"image"`
	Price     float64   `json:"price"`      // 💰 Main purchase price
	RentPrice float64   `json:"rent_price"` // 💰 Rental price per period
	Strategy  string    `json:"strategy"`
	OwnerID   uint      `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Status    string    `json:"status" gorm:"default:'inactive'"`

	// 🕒 Subscription / Rent Info
	SubscriptionType   string `json:"subscription_type"`   // e.g. "monthly", "weekly", "lifetime"
	SubscriptionExpiry string `json:"subscription_expiry"` // optional: template expiry or plan info

	// 💵 Optional metadata for display
	Description string `json:"description"` // for showing in UI
	Category    string `json:"category"`    // e.g., "digit", "rise/fall", etc.
	Version     string `json:"version"`     // if admin uploads new bot version
}
