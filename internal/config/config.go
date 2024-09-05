package config

import "os"

const (
	ENV_LOCAL = "local"
)

func New() *Config {
	return &Config{
		DB_URL: os.Getenv("DB_URL"),
		PORT:   os.Getenv("PORT"),
		ENV:    os.Getenv("ENV"),
	}
}

type Config struct {
	DB_URL string `json:"db_url"`
	PORT   string `json:"port"`
	ENV    string `json:"env"`
}
