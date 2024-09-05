package models

import "time"

type Model struct {
	ID        int        `gorm:"primary_key" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `sql:"index" json:"deleted_at,omitempty"`
}

type Account struct {
	Model
	UserID    int  `json:"user_id"`
	IsPrimary bool `json:"is_primary,omitempty"`
}
