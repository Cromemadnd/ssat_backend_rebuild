package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Device struct {
	UUID         uuid.UUID `json:"uuid" gorm:"primaryKey;type:char(36)"`
	Status       int       `json:"status" gorm:"type:int"`
	LastReceived time.Time `json:"last_received"`
	OwnerID      uuid.UUID `json:"-" gorm:"type:char(36)"`
	Owner        User      `json:"owner" gorm:"foreignKey:OwnerID"`
}

func (d *Device) BeforeCreate(tx *gorm.DB) error {
	if d.UUID == uuid.Nil {
		d.UUID = uuid.New()
	}
	return nil
}
