package models

import (
	"Api/utils"
	"time"
)

type Person struct {
	ID                   uint      `json:"id" gorm:"primaryKey"`
	Name                 string    `json:"name"`
	Email                string    `json:"email" gorm:"uniqueIndex"`
	Password             string    `json:"-"` // hide in responses
	Role                 string    `json:"role" gorm:"default:USER"`
	Phone                string    `json:"phone"`
	Country              string    `json:"country"`
	RefreshToken         string    `json:"refresh_token"`
	Token                string    `json:"token,omitempty"`
	ResetToken           string    `json:"-"` // for password reset
	ResetExpiry          time.Time `json:"-"`
    CreatedAt utils.FormattedTime `json:"created_at"`
	UpdatedAt utils.FormattedTime `json:"updated_at"`
	TotalProfits         uint      `json:"total_profits"`
	ActiveBots           uint      `json:"active_bots"`
	TotalTrades          uint      `json:"total_trades"`
	Membership           string    `json:"member_ship_type"`
	SubscriptionExpiry   time.Time `json:"subscription_expiry"`
	UpgradeRequestStatus string    `json:"upgrade_request_status" gorm:"type:varchar(20);default:null"`
}
