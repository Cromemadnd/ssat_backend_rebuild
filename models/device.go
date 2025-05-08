package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Device struct {
	DeviceID     string     `json:"device_id" gorm:"type:char(16)"`
	Secret       string     `json:"-" gorm:"type:char(16)"`
	Status       int        `json:"status" gorm:"type:int;default:0"`
	LastReceived *time.Time `json:"last_received" gorm:"null"`
	OwnerID      *uuid.UUID `json:"-" gorm:"type:char(36);null"`
	Owner        *User      `json:"owner" gorm:"foreignKey:OwnerID"`
	Data         *[]Data    `json:"data" gorm:"foreignKey:MyDeviceID"`
	BaseModel
}

func (d *Device) BeforeCreate(tx *gorm.DB) (err error) {
	if err = d.BaseModel.BeforeCreate(tx); err != nil {
		return err
	}
	return nil
}
