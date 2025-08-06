package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	SupabaseKey string
	SupabaseURL string
	SupabaseAdminKey string
}

func LoadConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not found, reading config from environment")
	}

	return Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		DBSSLMode:  os.Getenv("DB_SSLMODE"),
		SupabaseKey: os.Getenv("SUPABASE_KEY"),
		SupabaseURL: os.Getenv("SUPABASE_URL"),
		SupabaseAdminKey: os.Getenv("SUPABASE_ADMIN_KEY"),
	}
}
