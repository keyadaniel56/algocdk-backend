package models

type Favorite struct {
	ID     uint `json:"id" gorm:"primaryKey"`
	UserID uint `json:"user_id"`
	BotID  uint `json:"bot_id"`

	User Person `json:"user" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	Bot  Bot    `json:"bot" gorm:"foreignKey:BotID;constraint:OnDelete:CASCADE;"`
}
