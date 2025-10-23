package models

type BotUser struct {
	ID     uint `gorm:"primaryKey"`
	BotID  uint
	UserID uint
}
