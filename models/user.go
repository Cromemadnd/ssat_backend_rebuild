package models

type User struct {
	Username       string   `json:"name" gorm:"type:varchar(32);not null"`
	HashedPassword string   `json:"-" gorm:"type:char(32);not null"`
	Devices        []Device `json:"devices" gorm:"foreignKey:OwnerID"`
	Permissions    uint8    `json:"permissions" gorm:"type:tinyint(1) unsigned;default:0"`
	BaseModel
}

func (u User) HasPerm(permission uint8) bool {
	return u.Permissions&permission != 0
}
