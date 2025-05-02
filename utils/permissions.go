package utils

const (
	// 读权限(低4位)
	ReadData    uint8 = 1 << 0
	ReadLogs    uint8 = 1 << 1
	ReadDevices uint8 = 1 << 2
	ReadUsers   uint8 = 1 << 3

	// 写权限(高4位)
	WriteData    uint8 = 1 << 4 // 0001 0000
	WriteLogs    uint8 = 1 << 5 // 0010 0000
	WriteDevices uint8 = 1 << 6 // 0100 0000
	WriteUsers   uint8 = 1 << 7 // 1000 0000
)
