package models

import (
	"ssat_backend_rebuild/utils"

	"github.com/google/uuid"
)

type Log struct {
	LogType uint8            `json:"log_type" gorm:"type:tinyint(1)"` // 0: 设备日志，1: 用户日志 2: 管理员日志
	Subject *uuid.UUID       `json:"subject" gorm:"type:char(36)"`    // 操作主体的UUID
	Path    string           `json:"path" gorm:"type:varchar(128)"`   // 请求路径
	Method  string           `json:"method" gorm:"type:varchar(8)"`   // 请求方法
	IP      string           `json:"ip" gorm:"type:varchar(16)"`      // 请求者IP
	Status  *utils.ErrorCode `json:"status" gorm:"embedded"`          // 请求状态
	BaseModel
}
