package models

import (
	"time"

	"github.com/google/uuid"
)

type Device struct {
	DeviceID     string     `json:"device_id" gorm:"type:char(16);uniqueIndex;not null"`
	Nickname     string     `json:"nickname" gorm:"type:varchar(64);not null"`
	Secret       string     `json:"-" gorm:"type:char(255)"`
	Status       int        `json:"status" gorm:"type:int;default:0"`
	LastReceived *time.Time `json:"last_received" gorm:"null"`
	OwnerID      *uuid.UUID `json:"-" gorm:"type:char(36);null"`
	Owner        *User      `json:"owner" gorm:"foreignKey:OwnerID"`
	Data         *[]Data    `json:"data" gorm:"foreignKey:MyDeviceID"`
	BaseModel
}
