package main

type DBConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	Charset  string
}

var dbConfig = DBConfig{
	Host:     "localhost",
	Port:     "3306",
	Username: "root",         // 替换为实际用户名
	Password: "123456",       // 替换为实际密码
	DBName:   "AeroSentinel", // 替换为实际数据库名称
	Charset:  "utf8mb4",
}
