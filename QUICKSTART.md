# 快速开始指南

本指南帮助您快速启动和运行微服务项目。

## 前置要求

在开始之前，请确保您的系统已安装以下软件：

1. **Go 1.21+**
   ```bash
   go version
   ```

2. **Docker 和 Docker Compose**（可选，用于快速启动依赖服务）
   ```bash
   docker --version
   docker-compose --version
   ```

3. **Protocol Buffers 编译器**（如果需要使用 gRPC）
   ```bash
   # macOS
   brew install protobuf
   
   # Linux
   apt-get install -y protobuf-compiler
   
   # 验证安装
   protoc --version
   ```

## 快速启动步骤

### 步骤 1: 安装依赖

```bash
# 下载 Go 模块依赖
go mod download
go mod tidy
```

### 步骤 2: 启动依赖服务

**使用 Docker Compose（推荐）：**

```bash
# 启动 PostgreSQL、Redis 和 RabbitMQ
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

**手动安装（如果不使用 Docker）：**

- PostgreSQL: 安装并启动，创建数据库 `microservice`
- Redis: 安装并启动
- RabbitMQ: 安装并启动

### 步骤 3: 配置应用

编辑 `config/config.yaml` 文件，修改数据库和其他服务的连接信息：

```yaml
database:
  host: localhost
  port: 5432
  user: postgres
  password: your_password  # 修改为您的密码
  dbname: microservice

redis:
  host: localhost
  port: 6379
  password: ""

rabbitmq:
  host: localhost
  port: 5672
  user: guest
  password: guest

aws:
  region: us-east-1
  access_key: your_access_key  # 修改为您的 AWS 密钥
  secret_key: your_secret_key
  s3:
    bucket: your-bucket-name
```

### 步骤 4: 生成 gRPC 代码（可选）

如果您需要使用 gRPC 服务：

```bash
# 方式 1: 使用脚本
chmod +x scripts/generate_proto.sh
./scripts/generate_proto.sh

# 方式 2: 使用 Makefile
make proto
```

### 步骤 5: 运行服务

您可以选择运行一个或多个服务：

**运行网关服务（HTTP API）：**

```bash
# 使用 go run
go run cmd/gateway/main.go

# 或使用 Makefile
make run-gateway
```

服务启动后访问: http://localhost:8080

**运行 gRPC 服务：**

```bash
# 使用 go run
go run cmd/grpc-server/main.go

# 或使用 Makefile
make run-grpc
```

gRPC 服务监听端口: 50051

**运行定时任务服务：**

```bash
# 使用 go run
go run cmd/cron-server/main.go

# 或使用 Makefile
make run-cron
```

## 测试 API

### 1. 健康检查

```bash
curl http://localhost:8080/health
```

预期响应：
```json
{
  "status": "ok",
  "timestamp": "2025-10-31T10:00:00Z"
}
```

### 2. 详细健康检查

```bash
curl http://localhost:8080/health/detail
```

### 3. 上传文件（需要配置 AWS S3）

```bash
curl -X POST http://localhost:8080/api/v1/upload \
  -F "file=@/path/to/your/file.jpg"
```

### 4. 发送消息到队列

```bash
curl -X POST http://localhost:8080/api/v1/message \
  -H "Content-Type: application/json" \
  -d '{
    "queue": "task",
    "message": {
      "type": "email",
      "to": "user@example.com",
      "content": "Hello World"
    }
  }'
```

## 常见问题

### 1. 数据库连接失败

**错误**: `连接数据库失败: connection refused`

**解决方案**:
- 检查 PostgreSQL 是否正在运行
- 确认配置文件中的数据库连接信息正确
- 如果使用 Docker，等待容器完全启动（约30秒）

### 2. Redis 连接失败

**错误**: `Redis 连接失败: connection refused`

**解决方案**:
- 检查 Redis 是否正在运行
- 确认配置文件中的 Redis 连接信息正确

### 3. RabbitMQ 连接失败

**错误**: `连接 RabbitMQ 失败`

**解决方案**:
- 检查 RabbitMQ 是否正在运行
- 访问管理界面: http://localhost:15672 (用户名/密码: guest/guest)
- 确认配置文件中的连接信息正确

### 4. AWS S3 上传失败

**错误**: `上传文件到 S3 失败`

**解决方案**:
- 检查 AWS 访问密钥和秘密密钥是否正确
- 确认 S3 存储桶存在且有写入权限
- 检查 AWS 区域设置是否正确

### 5. Proto 文件生成失败

**错误**: `protoc: command not found`

**解决方案**:
```bash
# macOS
brew install protobuf

# Ubuntu/Debian
sudo apt-get install -y protobuf-compiler

# 安装 Go 插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## 开发建议

### 1. 使用 Makefile

查看所有可用命令：
```bash
make help
```

常用命令：
- `make build`: 编译所有服务
- `make run-gateway`: 运行网关服务
- `make run-grpc`: 运行 gRPC 服务
- `make run-cron`: 运行定时任务服务
- `make proto`: 生成 Proto 文件
- `make clean`: 清理编译文件
- `make docker-up`: 启动 Docker 服务
- `make docker-down`: 停止 Docker 服务

### 2. 查看日志

日志文件位置：
- 应用日志: `logs/app.log`
- 错误日志: `logs/error.log`
- 标准输出: 终端

### 3. 调试模式

在 `config/config.yaml` 中设置：
```yaml
server:
  mode: debug  # debug, release, test
```

### 4. 停止服务

- 终端中按 `Ctrl+C` 优雅停止服务
- Docker 服务: `docker-compose down`

## 下一步

1. 查看 [README.md](README.md) 了解项目架构和功能
2. 阅读各个模块的代码和注释
3. 根据您的业务需求扩展功能
4. 添加单元测试和集成测试
5. 配置 CI/CD 流程

## 技术支持

如果遇到问题：
1. 检查日志文件
2. 查看 Docker 容器日志: `docker-compose logs`
3. 确认所有依赖服务正常运行
4. 检查配置文件是否正确

