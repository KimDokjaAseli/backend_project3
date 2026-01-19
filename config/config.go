package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerHost     string
	ServerPort     string
	ServerAddress  string
	GinMode        string
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	JWTSecret      string
	JWTExpiryHours int
	AllowedOrigins string
	MaxUploadSize  int64
	UploadPath     string
}

func LoadConfig() *Config {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Parse JWT expiry hours
	jwtExpiry, err := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	if err != nil {
		jwtExpiry = 24
	}

	// Parse max upload size
	maxUploadSize, err := strconv.ParseInt(getEnv("MAX_UPLOAD_SIZE", "10485760"), 10, 64)
	if err != nil {
		maxUploadSize = 10485760 // 10MB default
	}

	serverHost := getEnv("SERVER_HOST", "localhost")
	serverPort := getEnv("SERVER_PORT", "8102")

	return &Config{
		ServerHost:     serverHost,
		ServerPort:     serverPort,
		ServerAddress:  serverHost + ":" + serverPort,
		GinMode:        getEnv("GIN_MODE", "debug"),
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "3306"),
		DBUser:         getEnv("DB_USER", "root"),
		DBPassword:     getEnv("DB_PASSWORD", ""),
		DBName:         getEnv("DB_NAME", "wallet_point"),
		JWTSecret:      getEnv("JWT_SECRET", "change-this-secret-key-in-production"),
		JWTExpiryHours: jwtExpiry,
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "https://walletpoint.xeroon.my.id"),
		MaxUploadSize:  maxUploadSize,
		UploadPath:     getEnv("UPLOAD_PATH", "./uploads"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
