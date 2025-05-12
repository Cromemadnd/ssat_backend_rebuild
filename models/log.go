package models

import "github.com/google/uuid"

type Log struct {
	IsUserLog bool       `json:"is_user_log" gorm:"default:false"`
	Subject   *uuid.UUID `json:"subject" gorm:"type:char(36)"`
	Content   string     `json:"content" gorm:"type:varchar(256)"`
	BaseModel
}
