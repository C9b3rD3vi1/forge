package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Setting struct {
	ID    uuid.UUID `gorm:"primaryKey"`
	Key   string    `gorm:"uniqueIndex;size:100;not null"`
	Value string    `gorm:"type:text"`
}

func (s *Setting) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}
