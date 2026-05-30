package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PageView struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	Path      string    `gorm:"index;not null"`
	Entity    string    `gorm:"size:50;index"` // "project", "service", "post", "page"
	EntityID  string    `gorm:"index"`          // UUID of the entity
	IP        string    `gorm:"size:45"`
	UserAgent string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"index"`
}

type DailyStat struct {
	ID          uuid.UUID `gorm:"primaryKey"`
	Date        time.Time `gorm:"index"`
	Entity      string    `gorm:"size:50;index"`
	EntityID    string    `gorm:"index"`
	ViewCount   int       `gorm:"default:0"`
	UniqueViews int       `gorm:"default:0"`
}

func (p *PageView) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return
}

func (d *DailyStat) BeforeCreate(tx *gorm.DB) (err error) {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return
}
