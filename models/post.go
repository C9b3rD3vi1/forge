package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Post struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	Title     string    `gorm:"not null"`
	Slug      string    `gorm:"not null;uniqueIndex"`
	Excerpt   string    `gorm:"type:text"`
	Content   string    `gorm:"type:text;not null"`
	ImageURL  string    `gorm:"type:text"`
	Category  string    `gorm:"size:100;default:'general'"`
	Author    string    `gorm:"size:100;not null;default:'Nickson Wekongo'"`
	ReadingTime int     `gorm:"default:0"`

	ViewCount   int       `gorm:"default:0"`
	ContentText string    `gorm:"type:text"`

	Featured    bool      `gorm:"default:false"`
	Published   bool      `gorm:"default:true"`
	PublishedAt time.Time `gorm:"autoCreateTime"`
	CanonicalURL string   `gorm:"type:text"`

	Tags     []Tag     `gorm:"many2many:post_tags;"`
	Comments []Comment `gorm:"foreignKey:PostID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

type Tag struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	Name      string    `gorm:"not null;uniqueIndex"`
	Posts     []Post    `gorm:"many2many:post_tags;"`
	CreatedAt time.Time
	UpdatedAt time.Time
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
	if p.ContentText == "" && p.Content != "" {
		p.ContentText = StripMarkdown(p.Content)
	}
	return
}

func (p *Post) BeforeUpdate(tx *gorm.DB) (err error) {
	if p.Content != "" {
		p.ContentText = StripMarkdown(p.Content)
	}
	return
}

func (t *Tag) BeforeUpdate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return
}
