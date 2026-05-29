package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"github.com/google/uuid"
)

// User struct represents a user entity with personal and contact details.
type User struct {
	ID              uuid.UUID `gorm:"primaryKey"`
	FullName        string `gorm:"unique;not null"`
	Username        string `gorm:"unique;not null"`
	Email           string `gorm:"unique;not null"`
	Password        string `gorm:"required"`
	PasswordConfirm string
	Address         string
	Phone           string
	TwoFASecret     string // save TOTP secret here
	IsActive        bool

	// admin
	IsAdmin bool `gorm:"default:false"`

	gorm.Model
}

// Comment struct represents a comment entity with user and post details.
type Comment struct {
	ID      uuid.UUID `gorm:"primaryKey"`
	Content string    `gorm:"type:text;not null"`
	UserID  uuid.UUID `gorm:"not null"`
	User    User      `gorm:"foreignKey:UserID"`
	PostID  uuid.UUID `gorm:"not null"`
	Post    Post      `gorm:"foreignKey:PostID"`
	ParentID *uuid.UUID `gorm:""`

	gorm.Model
}


// ContactMessage struct represents a contact message entity with user and contact details.
type ContactMessage struct {
	ID       uuid.UUID `gorm:"primaryKey"`
	Name     string
	Email    string
	Subject  string
	Message  string
	Services string `gorm:"type:text"`

	IsRead bool `gorm:"default:false"`

	gorm.Model
}


func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
    if u.ID == uuid.Nil {
        u.ID = uuid.New()
    }
    return
}

func (u *ContactMessage) BeforeCreate(tx *gorm.DB) (err error) {
    if u.ID == uuid.Nil {
        u.ID = uuid.New()
    }
    return
}


// HashPassword hashes the user's password
func (u *User) HashPassword() error {
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    u.Password = string(hashedPassword)
    return nil
}

// CheckPassword compares plain password with hashed password
func (u *User) CheckPassword(password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
    return err == nil
}
