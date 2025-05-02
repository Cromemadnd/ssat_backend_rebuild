package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	UUID           uuid.UUID `json:"uuid" gorm:"primaryKey;type:char(36)"`
	Username       string    `json:"name" gorm:"type:varchar(32);not null"`
	HashedPassword string    `json:"-" gorm:"type:char(32);not null"`
	Devices        []Device  `json:"devices" gorm:"foreignKey:OwnerID"`
	Permissions    uint8     `json:"permissions" gorm:"type:tinyint(1) unsigned;default:0"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.UUID == uuid.Nil {
		u.UUID = uuid.New()
	}
	return nil
}
