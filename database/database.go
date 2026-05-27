package database

import (
	"log"

	"github.com/C9b3rD3vi1/forge/models"
	"gorm.io/driver/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DB is the database connection
var DB *gorm.DB

// DBConnection is the function to connect to the database

func InitDB() (*gorm.DB, error) {
	// Connect to the SQLite database
	db, err := gorm.Open(sqlite.Open("server.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
		//return nil, err
	}
	// Set the global DB variable
	DB = db
	// Log the database connection
	log.Println("Connected to the database")
{}

   // Drop table
   //db.Migrator().DropTable(&models.User{},&models.ContactMessage{},)

	
	
	// Migrate the schema
	if err := db.AutoMigrate(
		&models.Post{}, 
		&models.User{}, 
		&models.Comment{}, 
		&models.Projects{}, 
		&models.Services{}, 
		&models.Tag{},
	 	&models.ContactMessage{},
		&models.TechStack{}); err != nil {
		log.Fatal("Failed to migrate the database schema:", err)
		return nil, err
	}
	// Log the migration
	log.Println("Database schema migrated successfully")

	return db, nil
}

// CreateAdminUser creates the initial admin user
func CreateAdminUser(db *gorm.DB) error {
    // Check if admin user already exists
    var count int64
    db.Model(&models.User{}).Where("username = ?", "admin").Count(&count)
    
    if count > 0 {
        return nil // Admin already exists
    }

    admin := models.User{
    	ID: uuid.New(),
    	Username: "admin",
        Email:    "admin@example.com",
        Password: "admin123", // Change this in production!
        IsAdmin:  true,
        IsActive: true,
    }

    if err := admin.HashPassword(); err != nil {
        return err
    }

    return db.Create(&admin).Error
}