package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseModelInterface interface {
	BeforeCreate(tx *gorm.DB) error
}

type BaseModel struct {
	UUID      uuid.UUID `json:"uuid" gorm:"primaryKey;type:char(36);not null;<-:create"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime;<-:create"`
}

func (m *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if m.UUID == uuid.Nil { // 检查 UUID 是否为零值
		m.UUID = uuid.New() // 生成新的 UUID
	}
	return nil
}
