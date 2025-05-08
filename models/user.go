package models

type User struct {
	Username       string   `json:"username" gorm:"type:varchar(32);not null"`
	HashedPassword string   `json:"-" gorm:"type:char(32);not null"`
	IsAdmin        bool     `json:"is_admin" gorm:"default:false"`
	Devices        []Device `json:"devices" gorm:"foreignKey:OwnerID"`
	BaseModel
}
