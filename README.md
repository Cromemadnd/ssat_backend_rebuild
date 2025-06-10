# AeroSentinel Backend

[![Go Version](https://img.shields.io/badge/Go-1.24.2-blue.svg)](https://golang.org/)
[![Gin Framework](https://img.shields.io/badge/Gin-v1.10.0-green.svg)](https://gin-gonic.com/)
[![MySQL](https://img.shields.io/badge/MySQL-8.0+-orange.svg)](https://www.mysql.com/)
[![MongoDB](https://img.shields.io/badge/MongoDB-6.0+-green.svg)](https://www.mongodb.com/)


## 🏗️ 技术架构

- **后端框架**: Gin (Go)
- **数据库**: MySQL 8.0+ (主要数据) + MongoDB (时序数据)
- **认证**: JWT Token
- **缓存**: Go-Cache (内存缓存)
- **文档**: RESTful API

## 📋 系统要求

### 运行环境
- Go 1.24.2+
- MySQL 8.0+
- MongoDB 6.0+
- Linux/Windows/macOS

## 🚀 快速部署

### 方式一：从源码编译部署

#### 1. 准备环境

```bash
# 安装Go (如果未安装)
# 下载地址: https://golang.org/dl/

# 验证Go安装
go version

# 安装MySQL
# Ubuntu/Debian:
sudo apt update
sudo apt install mysql-server

# CentOS/RHEL:
sudo yum install mysql-server

# 安装MongoDB
# Ubuntu/Debian:
sudo apt install mongodb

# CentOS/RHEL:
sudo yum install mongodb-server
```

#### 2. 获取源码

```bash
# 克隆项目 (如果是从Git仓库)
git clone <your-repository-url>
cd ssat_backend_rebuild

# 或者直接使用现有代码目录
cd /path/to/ssat_backend_rebuild
```

#### 3. 安装依赖

```bash
# 下载Go模块依赖
go mod download

# 验证依赖
go mod verify
```

#### 4. 配置数据库

**MySQL配置:**
```sql
-- 连接MySQL
mysql -u root -p

-- 创建数据库
CREATE DATABASE AeroSentinel CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 创建用户 (可选，建议生产环境使用)
CREATE USER 'ssat_user'@'localhost' IDENTIFIED BY 'your_password';
GRANT ALL PRIVILEGES ON AeroSentinel.* TO 'ssat_user'@'localhost';
FLUSH PRIVILEGES;
```

**MongoDB配置:**
```bash
# 启动MongoDB服务
sudo systemctl start mongod
sudo systemctl enable mongod

# 验证MongoDB运行
mongo --eval "db.runCommand('ping')"
```

#### 5. 配置应用

复制并编辑配置文件：
```bash
# 配置文件已存在，直接编辑
cp config.json config.json.backup  # 备份原配置
```

编辑 `config.json`：
```json
{
  "mysql": {
    "host": "localhost",
    "port": 3306,
    "username": "root",
    "password": "你的MySQL密码",
    "db_name": "AeroSentinel",
    "charset": "utf8mb4"
  },
  "mongodb": {
    "host": "localhost",
    "port": 27017,
    "db_name": "AeroSentinel",
    "collection": "EnvData"
  },
  "jwt": {
    "secret": "your-jwt-secret-key-change-in-production",
    "expires": 3600,
    "refresh": 7200
  },
  "wechat": {
    "app_id": "your-wechat-app-id",
    "secret": "your-wechat-secret"
  },
  "admins": [
    {
      "username": "admin",
      "password": "admin123"
    }
  ],
  "mongo_to_sql_threshold": 1000,
  "ai_api_url": "your-ai-api-url",
  "ai_api_key": "your-ai-api-key",
  "server_addr": ":8080"
}
```

#### 6. 编译应用

```bash
# 编译应用
go build -o ssat_backend_rebuild

# 或者直接运行
go run main.go
```

#### 7. 启动服务

```bash
# 方式1: 直接运行编译好的二进制文件
./ssat_backend_rebuild

# 方式2: 使用go run
go run main.go

# 方式3: 后台运行
nohup ./ssat_backend_rebuild > app.log 2>&1 &
```

### 方式二：Docker部署 (推荐生产环境)

#### 1. 创建Dockerfile

```dockerfile
# 创建 Dockerfile
FROM golang:1.24.2-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ssat_backend_rebuild

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/ssat_backend_rebuild .
COPY --from=builder /app/config.json .

EXPOSE 8080
CMD ["./ssat_backend_rebuild"]
```

#### 2. 创建docker-compose.yml

```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - mysql
      - mongodb
    environment:
      - GIN_MODE=release
    volumes:
      - ./config.json:/root/config.json

  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: 123456
      MYSQL_DATABASE: AeroSentinel
      MYSQL_CHARACTER_SET_SERVER: utf8mb4
      MYSQL_COLLATION_SERVER: utf8mb4_unicode_ci
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql

  mongodb:
    image: mongo:6.0
    ports:
      - "27017:27017"
    volumes:
      - mongodb_data:/data/db

volumes:
  mysql_data:
  mongodb_data:
```

#### 3. 部署命令

```bash
# 构建并启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

## 🔧 配置说明

### 环境变量

可以使用环境变量覆盖配置文件中的设置：

```bash
export MYSQL_HOST=localhost
export MYSQL_PORT=3306
export MYSQL_USER=root
export MYSQL_PASSWORD=your_password
export MONGODB_HOST=localhost
export MONGODB_PORT=27017
export JWT_SECRET=your-jwt-secret
export SERVER_PORT=8080
```

### 生产环境配置建议

1. **安全配置**:
   - 修改默认的JWT密钥
   - 使用强密码
   - 启用HTTPS
   - 配置防火墙

2. **性能优化**:
   - 调整数据库连接池大小
   - 配置缓存策略
   - 设置合适的MongoDB阈值

3. **监控配置**:
   - 配置日志级别
   - 设置监控告警
   - 定期备份数据库

## 📡 API接口

服务启动后，API接口将在配置的端口上可用（默认8080）。

### 主要接口模块

- `POST /auth/login` - 管理员登录
- `POST /auth/wechat_login` - 微信登录
- `GET /devices/` - 设备列表 (管理员)
- `GET /devices/my_devices` - 我的设备 (用户)
- `POST /data/upload` - 数据上传
- `GET /data/my_data` - 我的数据 (用户)
- `GET /tickets/my_tickets` - 我的工单 (用户)
- `GET /announcements/` - 公告列表

### API认证

大部分接口需要JWT认证，在请求头中添加：
```
Authorization: Bearer <your-jwt-token>
```

## 🔍 验证部署

### 1. 健康检查

```bash
# 检查服务状态
curl http://localhost:8080/announcements/

# 检查数据库连接
# 查看应用日志确认数据库连接成功
```

### 2. 测试登录

```bash
# 管理员登录测试
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

### 3. 检查日志

```bash
# 查看应用日志
tail -f app.log

# 检查系统资源
top
df -h
```

## 🛠️ 常见问题

### 1. 数据库连接失败

**问题**: `connection refused` 或 `access denied`

**解决方案**:
```bash
# 检查MySQL服务状态
sudo systemctl status mysql

# 检查MongoDB服务状态
sudo systemctl status mongod

# 重启服务
sudo systemctl restart mysql
sudo systemctl restart mongod

# 检查配置文件中的数据库连接信息
```

### 2. 端口已被占用

**问题**: `bind: address already in use`

**解决方案**:
```bash
# 查找占用端口的进程
lsof -i :8080
netstat -tlnp | grep :8080

# 杀死进程或修改配置文件中的端口
```

### 3. 权限问题

**问题**: `permission denied`

**解决方案**:
```bash
# 给二进制文件执行权限
chmod +x ssat_backend_rebuild

# 检查文件所有权
ls -la ssat_backend_rebuild
```

### 4. 依赖模块问题

**问题**: `go module` 相关错误

**解决方案**:
```bash
# 清理模块缓存
go clean -modcache

# 重新下载依赖
go mod download

# 更新依赖
go mod tidy
```

## 📊 系统监控

### 1. 日志监控

```bash
# 实时查看日志
tail -f app.log

# 搜索错误日志
grep "ERROR" app.log

# 查看访问统计
grep "GET\|POST\|PUT\|DELETE" app.log | wc -l
```

### 2. 性能监控

```bash
# 查看进程资源使用
ps aux | grep ssat_backend_rebuild

# 查看系统负载
htop

# 查看数据库状态
mysql -u root -p -e "SHOW PROCESSLIST;"
```

## 🔄 更新升级

### 1. 更新应用

```bash
# 停止服务
pkill ssat_backend_rebuild

# 备份当前版本
cp ssat_backend_rebuild ssat_backend_rebuild.backup

# 编译新版本
go build -o ssat_backend_rebuild

# 启动新版本
./ssat_backend_rebuild
```

### 2. 数据库迁移

```bash
# 备份数据库
mysqldump -u root -p AeroSentinel > backup.sql
mongodump --db AeroSentinel --out mongodb_backup/

# 执行升级后再验证数据完整性
```

## 📞 技术支持

- **项目地址**: [项目仓库地址]
- **文档**: [在线文档地址]
- **问题反馈**: [Issues地址]

## 📄 许可证

本项目采用 [MIT License](LICENSE) 许可证。

---

**⚠️ 重要提醒**:
- 生产环境部署前请务必修改默认密码和JWT密钥
- 定期备份数据库数据
- 监控系统资源使用情况
- 及时更新安全补丁

如有部署问题，请查看常见问题部分或联系技术支持。
