package main

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
	Host   string `json:"host"`
	Port   int    `json:"port"`
	DBName string `json:"db_name"`
}

type JWTConfig struct {
	Secret  string `json:"secret"`
	Expires int    `json:"expires"`
	Refresh int    `json:"refresh"`
}

type Config struct {
	SQLConfig   SQLConfig   `json:"mysql"`
	MongoConfig MongoConfig `json:"mongodb"`
	JWTConfig   JWTConfig   `json:"jwt"`
}

var MyConfig Config

func LoadConfig() {
	file, err := os.Open("config.json")
	if err != nil {
		panic("Failed to open config.json: " + err.Error())
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&MyConfig)
	if err != nil {
		panic("Failed to decode config.json: " + err.Error())
	}
}
