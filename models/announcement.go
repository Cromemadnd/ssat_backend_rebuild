package models

import "time"

type Announcement struct {
	Title      string     `json:"title" gorm:"type:varchar(128);not null"`
	Content    string     `json:"content" gorm:"type:text;not null"`
	Publisher  string     `json:"publisher" gorm:"type:varchar(32);null"`
	Type       uint8      `json:"type" gorm:"type:tinyint(1);default:0"` // 0: 普通公告，1: 紧急公告
	ModifiedAt *time.Time `json:"modified_at" gorm:"autoUpdateTime"`
	BaseModel
}
