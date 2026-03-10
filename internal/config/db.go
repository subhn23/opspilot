package config

import (
	"fmt"
	"log"
	"opspilot/internal/models"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB initializes the connection to the database
func InitDB(dialector gorm.Dialector) *gorm.DB {
	if dialector == nil {
		dialector = GetDialector()
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	DB = db
	log.Println("Database connection established")

	if err := AutoMigrate(DB); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	return DB
}

// GetDialector returns a GORM dialector for PostgreSQL based on environment variables
func GetDialector() gorm.Dialector {
	host := getEnv("DB_HOST", "localhost")
	user := getEnv("DB_USER", "postgres")
	pass := getEnv("DB_PASS", "postgres")
	name := getEnv("DB_NAME", "opspilot")
	port := getEnv("DB_PORT", "5432")

	// HA Connection string (SSL Mode disabled for local dev, enable for prod)
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		host, user, pass, name, port)
	return postgres.Open(dsn)
}

// AutoMigrate runs GORM's AutoMigrate for all system models
func AutoMigrate(db *gorm.DB) error {
	log.Println("Running database migrations...")
	return db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.Certificate{},
		&models.ProxyRoute{},
		&models.Environment{},
		&models.Deployment{},
		&models.AuditLog{},
		&models.CertTestOverride{},
	)
}
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
