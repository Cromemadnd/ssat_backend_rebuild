package models

type DataEntry struct {
	Temperature float32 `json:"temperature" binding:"required"   gorm:"type:float;default:0"`
	Humidity    float32 `json:"humidity"    binding:"required"   gorm:"type:float;default:0"`
	FreshAir    float32 `json:"fresh_air"   binding:"required"   gorm:"type:float;default:0"`
	Ozone       float32 `json:"ozone"       binding:"required"   gorm:"type:float;default:0"`
	NitroDio    float32 `json:"nitro_dio"   binding:"required"   gorm:"type:float;default:0"`
	Methanal    float32 `json:"methanal"    binding:"required"   gorm:"type:float;default:0"`
	Pm25        float32 `json:"pm2_5"       binding:"required"   gorm:"type:float;default:0"`
	CarbMomo    float32 `json:"carb_momo"   binding:"required"   gorm:"type:float;default:0"`
	Bacteria    float32 `json:"bacteria"    binding:"required"   gorm:"type:float;default:0"`
	Radon       float32 `json:"radon"       binding:"required"   gorm:"type:float;default:0"`
}

type Data struct {
	MyDeviceID string    `json:"device_id" gorm:"type:char(16)"`
	MyDevice   *Device   `json:"my_device" gorm:"foreignKey:MyDeviceID"`
	Avg        DataEntry `json:"avg" gorm:"embedded;embeddedPrefix:avg_"`
	Var        DataEntry `json:"var" gorm:"embedded;embeddedPrefix:var_"`
	Min        DataEntry `json:"min" gorm:"embedded;embeddedPrefix:min_"`
	Max        DataEntry `json:"max" gorm:"embedded;embeddedPrefix:max_"`
	BaseModel
}
