package gorm

import "time"

const (
	// Database definition version
	Version = 0

	// Table names
	TblConfigurations = "configurations"
	TblAPIKeys        = "apiKeys"

	// DB Table Columns
	ColID         = "ID"
	ColCreateDate = "createDate"
	ColUpdateDate = "updateDate"
	ColUserID     = "userID"
	ColKey        = "key"
	ColValue      = "value"
)

type ApiKey struct {
	ID        uint      `json:"id"gorm:"primary_key"`
	UserId    int       `json:"user_id"gorm:"not null"`
	Key       string    `json:"key"gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Configuration struct {
	Key       string    `json:"key"gorm:"primary_key,not null"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
