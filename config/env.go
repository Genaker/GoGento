package config

import (
	"github.com/joho/godotenv"
	"log"
)

func LoadEnv() {
	_ = godotenv.Load()
	// If .env is missing, ignore error (env vars can be set by other means)
	log.Println("Environment variables loaded (if .env present)")
}
