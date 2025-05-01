package models

import (
	"time"

	"github.com/google/uuid"
)

type Device struct {
	UUID       uuid.UUID `json:"id" gorm:"primaryKey"`
	Name       string    `json:"name" gorm:"type:varchar(100);not null"`
	Status     int       `json:"status" gorm:"type:int;not null"`
	LastOnline time.Time `json:"last_online"`
	OwnerID    uuid.UUID `json:"owner_id"`
	Owner      User      `json:"owner" gorm:"foreignKey:OwnerID"`
}
