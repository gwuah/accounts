package database

import (
	"github.com/gwuah/accounts/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func RunMigrations(db *gorm.DB, models ...interface{}) error {
	return db.AutoMigrate(models...)
}

func NewPGConnection(config *config.Config) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(config.DB_URL), &gorm.Config{})
}
