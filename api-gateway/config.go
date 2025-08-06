package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	JwtSecret string
}

func LoadConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not found, reading config from environment")
	}

	return Config{
		JwtSecret: os.Getenv("JWT_SECRET"),
	}
}
