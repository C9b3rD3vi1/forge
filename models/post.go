package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Post represents a blog post in the database
// Post represents a blog post in the database
type Post struct {
    ID        uuid.UUID `gorm:"primaryKey"`
    Title     string    `gorm:"not null"`
    Slug      string    `gorm:"not null;uniqueIndex"`
    ImageURL  string    `gorm:"not null"`
    Content   string    `gorm:"not null"`
    Author    string    `gorm:"not null"`
    Tags      []Tag     `gorm:"many2many:post_tags;"`
    CreatedAt time.Time `gorm:"autoCreateTime"`
    UpdatedAt time.Time `gorm:"autoUpdateTime"`
    Published bool      `gorm:"not null;default:true"`
}

// Tag represents a tag in the database
type Tag struct {
    ID        uuid.UUID `gorm:"primaryKey"`
    Name      string    `gorm:"not null;uniqueIndex"`
    Posts     []Post    `gorm:"many2many:post_tags;"`
    CreatedAt time.Time `gorm:"autoCreateTime"`
    UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (t *Tag) BeforeCreate(tx *gorm.DB) (err error) {
    if t.ID == uuid.Nil {
        t.ID = uuid.New()
    }
    return
}

func (p *Post) BeforeCreate(tx *gorm.DB) (err error) {
    if p.ID == uuid.Nil {
        p.ID = uuid.New()
    }
    return
}

func (t *Tag) BeforeUpdate(tx *gorm.DB) (err error) {
    if t.ID == uuid.Nil {
        t.ID = uuid.New()
    }
    return
}
