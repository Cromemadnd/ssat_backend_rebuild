package setup

import (
	"encoding/json"
	"os"
)

type SQLConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	DBName   string `json:"db_name"`
	Charset  string `json:"charset"`
}

type MongoConfig struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	DBName     string `json:"db_name"`
	Collection string `json:"collection"`
}

type JWTConfig struct {
	Secret  string `json:"secret"`
	Expires int    `json:"expires"`
	Refresh int    `json:"refresh"`
}

type WechatConfig struct {
	AppID  string `json:"app_id"`
	Secret string `json:"secret"`
}

type AdminEntry struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Config struct {
	SQLConfig    SQLConfig    `json:"mysql"`
	MongoConfig  MongoConfig  `json:"mongodb"`
	JWTConfig    JWTConfig    `json:"jwt"`
	WechatConfig WechatConfig `json:"wechat"`
	AdminsConfig []AdminEntry `json:"admins"`
	ServerAddr   string       `json:"server_addr"`
}

func LoadConfig() Config {
	config := Config{}
	file, err := os.Open("config.json")
	if err != nil {
		panic("Failed to open config.json: " + err.Error())
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		panic("Failed to decode config.json: " + err.Error())
	}
	return config
}
