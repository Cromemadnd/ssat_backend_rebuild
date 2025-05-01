package models

import "github.com/google/uuid"

type User struct {
	UUID    uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name    string    `json:"name" gorm:"type:varchar(100);not null"`
	Devices []Device  `json:"devices" gorm:"foreignKey:OwnerID"`
}
