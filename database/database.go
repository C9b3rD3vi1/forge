package database

import (
	"log"
	"os"
	"path/filepath"

	"github.com/C9b3rD3vi1/forge/models"
	"gorm.io/driver/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DB is the database connection
var DB *gorm.DB

// DBConnection is the function to connect to the database

func InitDB() (*gorm.DB, error) {
    // Get database path from environment variable or use default
    dbPath := os.Getenv("DB_PATH")
    if dbPath == "" {
        dbPath = "server.db"
    }
    
    // Ensure the directory exists
    dir := filepath.Dir(dbPath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        log.Fatal("Failed to create data directory:", err)
        return nil, err
    }
    
    // Connect to the SQLite database
    db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to the database:", err)
        return nil, err
    }
    
    // Set the global DB variable
    DB = db
    
    // Log the database connection
    log.Println("Connected to the database:", dbPath)
    
    // Migrate the schema
    if err := db.AutoMigrate(
        &models.Post{},
        &models.User{},
        &models.Comment{},
        &models.Projects{},
        &models.Services{},
        &models.Tag{},
        &models.ContactMessage{},
        &models.TechStack{},
        &models.Setting{}); err != nil {
        log.Fatal("Failed to migrate the database schema:", err)
        return nil, err
    }
    
    // Log the migration
    log.Println("Database schema migrated successfully")
    
    // Seed default settings if they don't exist
    SeedSettings(db)
    
    return db, nil
}


// getEnv returns the value of an environment variable or a default value
func getEnv(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}

// CreateAdminUser creates the initial admin user from environment variables
func CreateAdminUser(db *gorm.DB) error {
    username := getEnv("ADMIN_USERNAME", "admin")
    email := getEnv("ADMIN_EMAIL", "admin@example.com")
    password := getEnv("ADMIN_PASSWORD", "admin123")

    var count int64
    db.Model(&models.User{}).Where("username = ?", username).Count(&count)

    if count > 0 {
        return nil
    }

    admin := models.User{
    	ID:       uuid.New(),
    	Username: username,
        Email:    email,
        Password: password,
        IsAdmin:  true,
        IsActive: true,
    }

    if err := admin.HashPassword(); err != nil {
        return err
    }

    return db.Create(&admin).Error
}

func SeedSettings(db *gorm.DB) {
	defaults := map[string]string{
		"site_name":        "Forge.Hub",
		"site_tagline":     "Cybersecurity & DevOps Engineer",
		"site_description": "Professional portfolio and blog covering cybersecurity, DevOps, and software engineering.",
		"site_keywords":    "cybersecurity, devops, go, security auditing, cloud infrastructure",
		"github_username":  "C9b3rD3vi1",
		"twitter_url":      "",
		"linkedin_url":     "",
		"hero_title":       "Cybersecurity & DevOps Engineer",
		"hero_subtitle":    "Securing systems, automating workflows, building resilient infrastructure.",
		"contact_email":    "hello@simuxtech.com",
		"stat_faster_deploys":      "5×",
		"stat_uptime":             "99.9%",
		"stat_certifications":     "7",
		"stat_yrs_production":     "3+",
		"open_source_projects": `[{"icon":"📦","name":"forgebuild","desc":"auto-build CLI for Go projects"},{"icon":"🔐","name":"openvpn-manager","desc":"VPN user management API"},{"icon":"⚙️","name":"deploy-scripts","desc":"Infrastructure automation toolkit"},{"icon":"🧪","name":"security-audit-tools","desc":"Hardening & compliance scripts"}]`,
	}
	for key, value := range defaults {
		var count int64
		db.Model(&models.Setting{}).Where("key = ?", key).Count(&count)
		if count == 0 {
			db.Create(&models.Setting{Key: key, Value: value})
		}
	}
	log.Println("Default settings seeded")
}