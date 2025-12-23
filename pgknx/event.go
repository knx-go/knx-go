package pgknx

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Event struct {
	ID          uuid.UUID `gorm:"primaryKey"`
	Timestamp   time.Time `gorm:"index"`
	Command     string
	Source      string
	Destination string
	Data        []byte
	Group       string `gorm:"index"`
	DPT         string
	Decoded     string
}

func (e *Event) BeforeCreate(tx *gorm.DB) (err error) {
	e.ID = uuid.New()
	return nil
}
