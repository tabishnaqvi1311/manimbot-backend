package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        string         `gorm:"primaryKey" json:"id"`
	ClerkID   string         `gorm:"uniqueIndex;not null" json:"clerk_id"`
	Email     string         `gorm:"uniqueIndex;not null" json:"email"`
	FullName  string         `json:"full_name"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Chats     []Chat         `gorm:"foreignKey:UserID" json:"chats,omitempty"`
}

type Chat struct {
	ID        string         `gorm:"primaryKey" json:"id"`
	UserID    string         `gorm:"not null;index" json:"user_id"`
	Title     string         `json:"title"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Messages  []Message      `gorm:"foreignKey:ChatID" json:"messages,omitempty"`
}

type Message struct {
	ID          string         `gorm:"primaryKey" json:"id"`
	ChatID      string         `gorm:"not null;index" json:"chat_id"`
	Role        string         `gorm:"not null" json:"role"`
	Content     string         `gorm:"type:text" json:"content"`
	VideoURL    string         `json:"video_url,omitempty"`
	Explanation string         `gorm:"type:text" json:"explanation,omitempty"`
	Duration    int            `json:"duration,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
