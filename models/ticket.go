package models

import "github.com/google/uuid"

type Ticket struct {
	UserUUID    *uuid.UUID   `json:"user_uuid" gorm:"type:char(36);not null"`
	User        *User        `json:"-" gorm:"foreignKey:UserUUID"`
	Title       string       `json:"title" gorm:"type:varchar(128);not null"`
	Content     string       `json:"content" gorm:"type:text;not null"`
	Type        uint8        `json:"type" gorm:"type:tinyint(1);default:0"`
	DeviceUUID  *uuid.UUID   `json:"device_uuid" gorm:"type:char(36);not null"`
	Device      *Device      `json:"-" gorm:"foreignKey:DeviceUUID"`
	Status      uint8        `json:"status" gorm:"type:tinyint(1);default:0"` // 0: 未处理，1: 处理中，2: 已解决
	ChatHistory []TicketChat `json:"chat_history" gorm:"foreignKey:TicketID"`
	BaseModel
}

type TicketChat struct {
	TicketID *uuid.UUID `json:"-" gorm:"type:char(36);not null"`       // 关联的工单ID
	Ticket   *Ticket    `json:"-" gorm:"foreignKey:TicketID"`          // 关联的工单
	Type     uint8      `json:"type" gorm:"type:tinyint(1);default:0"` // 0: 用户消息，1: 管理员消息
	Subject  string     `json:"subject" gorm:"type:char(36);not null"` // 回复的管理员用户名
	Content  string     `json:"content" gorm:"type:text;not null"`
	BaseModel
}
