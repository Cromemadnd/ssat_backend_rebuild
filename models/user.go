package models

type User struct {
	WechatID string   `json:"wechat_id" gorm:"type:varchar(32);not null"`
	Devices  []Device `json:"devices" gorm:"foreignKey:OwnerID"`
	BaseModel
}

type Admin struct {
	Username       string `json:"username" gorm:"type:varchar(32);not null"`
	HashedPassword string `json:"-" gorm:"type:char(32);not null"`
	BaseModel
}
