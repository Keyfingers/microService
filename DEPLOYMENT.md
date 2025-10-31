# 部署指南

本文档说明如何将微服务系统部署到生产环境。

## 部署架构

```
┌─────────────┐
│   用户请求   │
└──────┬──────┘
       │
       v
┌─────────────────┐
│   负载均衡器     │  (Nginx/ALB)
└────────┬────────┘
         │
    ┌────┴────┬────────────┐
    │         │            │
    v         v            v
┌────────┐ ┌────────┐  ┌────────┐
│ Gateway│ │ Gateway│  │ Gateway│  (多实例)
│  :8080 │ │  :8080 │  │  :8080 │
└───┬────┘ └───┬────┘  └───┬────┘
    │          │            │
    └──────────┴────────────┘
               │
    ┌──────────┼──────────┬──────────┐
    │          │          │          │
    v          v          v          v
┌────────┐ ┌─────────┐ ┌─────┐  ┌──────┐
│ gRPC   │ │RabbitMQ │ │ S3  │  │Cron  │
│:50051  │ │ :5672   │ │     │  │Server│
└───┬────┘ └────┬────┘ └─────┘  └──────┘
    │           │
    v           v
┌──────────────────────┐
│   PostgreSQL + Redis  │
└──────────────────────┘
```

## 部署选项

### 选项 1: Docker 部署（推荐）

#### 1.1 准备 Dockerfile

创建 `Dockerfile`:

```dockerfile
# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 安装依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制代码
COPY . .

# 编译服务
RUN go build -o bin/gateway cmd/gateway/main.go
RUN go build -o bin/grpc-server cmd/grpc-server/main.go
RUN go build -o bin/cron-server cmd/cron-server/main.go

# 运行阶段
FROM alpine:latest

WORKDIR /app

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 复制编译好的二进制文件
COPY --from=builder /app/bin/ ./bin/
COPY --from=builder /app/config/ ./config/

# 暴露端口
EXPOSE 8080 50051

# 默认运行网关服务
CMD ["./bin/gateway"]
```

#### 1.2 构建镜像

```bash
# 构建网关镜像
docker build -t microservice-gateway:latest \
  --target gateway .

# 构建 gRPC 镜像
docker build -t microservice-grpc:latest \
  --target grpc-server .

# 构建 Cron 镜像
docker build -t microservice-cron:latest \
  --target cron-server .
```

#### 1.3 运行容器

```bash
# 启动所有服务
docker-compose -f docker-compose.prod.yml up -d
```

创建 `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  gateway:
    image: microservice-gateway:latest
    ports:
      - "8080:8080"
    environment:
      - DATABASE_HOST=postgres
      - REDIS_HOST=redis
      - RABBITMQ_HOST=rabbitmq
    depends_on:
      - postgres
      - redis
      - rabbitmq
    restart: unless-stopped

  grpc-server:
    image: microservice-grpc:latest
    ports:
      - "50051:50051"
    environment:
      - DATABASE_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis
    restart: unless-stopped

  cron-server:
    image: microservice-cron:latest
    environment:
      - DATABASE_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis
    restart: unless-stopped

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: microservice
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data
    restart: unless-stopped

  rabbitmq:
    image: rabbitmq:3.12-management-alpine
    environment:
      RABBITMQ_DEFAULT_USER: ${MQ_USER}
      RABBITMQ_DEFAULT_PASS: ${MQ_PASSWORD}
    volumes:
      - rabbitmq_data:/var/lib/rabbitmq
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
  rabbitmq_data:
```

---

### 选项 2: Kubernetes 部署

#### 2.1 创建配置文件

`k8s/deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gateway
spec:
  replicas: 3
  selector:
    matchLabels:
      app: gateway
  template:
    metadata:
      labels:
        app: gateway
    spec:
      containers:
      - name: gateway
        image: microservice-gateway:latest
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_HOST
          valueFrom:
            configMapKeyRef:
              name: app-config
              key: database_host
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
---
apiVersion: v1
kind: Service
metadata:
  name: gateway
spec:
  selector:
    app: gateway
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
```

#### 2.2 部署到 Kubernetes

```bash
# 创建命名空间
kubectl create namespace microservice

# 应用配置
kubectl apply -f k8s/ -n microservice

# 查看状态
kubectl get pods -n microservice
```

---

### 选项 3: 传统服务器部署

#### 3.1 编译二进制文件

```bash
# 在开发机器上编译
GOOS=linux GOARCH=amd64 go build -o bin/gateway cmd/gateway/main.go
GOOS=linux GOARCH=amd64 go build -o bin/grpc-server cmd/grpc-server/main.go
GOOS=linux GOARCH=amd64 go build -o bin/cron-server cmd/cron-server/main.go
```

#### 3.2 上传到服务器

```bash
# 上传文件
scp -r bin config user@server:/opt/microservice/
```

#### 3.3 创建 Systemd 服务

创建 `/etc/systemd/system/gateway.service`:

```ini
[Unit]
Description=Microservice Gateway
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=microservice
WorkingDirectory=/opt/microservice
ExecStart=/opt/microservice/bin/gateway
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

启动服务：

```bash
# 重载配置
sudo systemctl daemon-reload

# 启动服务
sudo systemctl start gateway
sudo systemctl enable gateway

# 查看状态
sudo systemctl status gateway
```

---

## 环境配置

### 生产环境配置清单

**config/config.prod.yaml**:

```yaml
server:
  gateway_port: 8080
  grpc_port: 50051
  mode: release  # 生产模式
  shutdown_timeout: 30

database:
  host: ${DATABASE_HOST}
  port: 5432
  user: ${DATABASE_USER}
  password: ${DATABASE_PASSWORD}
  dbname: microservice
  max_idle_conns: 10
  max_open_conns: 100

redis:
  host: ${REDIS_HOST}
  port: 6379
  password: ${REDIS_PASSWORD}

aws:
  region: ${AWS_REGION}
  access_key: ${AWS_ACCESS_KEY}
  secret_key: ${AWS_SECRET_KEY}

logger:
  level: info  # 生产环境使用 info 级别
  format: json
```

### 环境变量

创建 `.env.prod`:

```bash
# 数据库
DATABASE_HOST=your-db-host
DATABASE_USER=microservice_user
DATABASE_PASSWORD=secure_password

# Redis
REDIS_HOST=your-redis-host
REDIS_PASSWORD=redis_password

# RabbitMQ
RABBITMQ_HOST=your-mq-host
RABBITMQ_USER=mq_user
RABBITMQ_PASSWORD=mq_password

# AWS
AWS_REGION=us-east-1
AWS_ACCESS_KEY=your_access_key
AWS_SECRET_KEY=your_secret_key
AWS_S3_BUCKET=your-bucket
```

---

## 监控和日志

### 1. 日志收集

使用 ELK Stack 或 Loki 收集日志：

```yaml
# Filebeat 配置
filebeat.inputs:
- type: log
  paths:
    - /opt/microservice/logs/*.log
  json.keys_under_root: true

output.elasticsearch:
  hosts: ["elasticsearch:9200"]
```

### 2. 性能监控

集成 Prometheus + Grafana：

```go
// 添加到代码中
import "github.com/prometheus/client_golang/prometheus/promhttp"

router.GET("/metrics", gin.WrapH(promhttp.Handler()))
```

### 3. 健康检查

配置负载均衡器的健康检查：

```
Health Check URL: /health
Interval: 10s
Timeout: 5s
Healthy Threshold: 2
Unhealthy Threshold: 3
```

---

## 安全配置

### 1. 防火墙规则

```bash
# 只开放必要的端口
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw deny 8080/tcp  # 通过负载均衡器访问
```

### 2. SSL/TLS 配置

使用 Nginx 作为反向代理：

```nginx
server {
    listen 443 ssl http2;
    server_name api.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### 3. 数据库安全

```sql
-- 创建专用数据库用户
CREATE USER microservice_user WITH PASSWORD 'secure_password';
GRANT CONNECT ON DATABASE microservice TO microservice_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO microservice_user;
```

---

## 备份策略

### 1. 数据库备份

```bash
#!/bin/bash
# backup.sh

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backup/postgres"

pg_dump -h localhost -U postgres microservice | \
  gzip > ${BACKUP_DIR}/microservice_${DATE}.sql.gz

# 保留最近 7 天的备份
find ${BACKUP_DIR} -name "*.sql.gz" -mtime +7 -delete
```

添加到 crontab:
```
0 2 * * * /opt/scripts/backup.sh
```

### 2. 配置备份

定期备份配置文件到版本控制系统。

---

## 扩展和性能优化

### 1. 水平扩展

- 网关服务：可扩展到多个实例，通过负载均衡器分发请求
- gRPC 服务：可扩展到多个实例
- Cron 服务：使用 Redis 分布式锁，确保任务不重复执行

### 2. 数据库优化

```sql
-- 添加索引
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_created_at ON users(created_at);

-- 分析查询性能
EXPLAIN ANALYZE SELECT * FROM users WHERE email = 'test@example.com';
```

### 3. 缓存策略

- 使用 Redis 缓存热点数据
- 设置合理的过期时间
- 使用缓存预热

---

## 故障排查

### 常见问题

1. **服务无法启动**
   - 检查配置文件路径
   - 检查依赖服务是否运行
   - 查看日志文件

2. **数据库连接失败**
   - 检查网络连接
   - 验证数据库凭据
   - 检查防火墙规则

3. **内存不足**
   - 调整连接池大小
   - 增加服务器内存
   - 检查是否有内存泄漏

### 监控指标

关键指标：
- CPU 使用率 < 70%
- 内存使用率 < 80%
- 磁盘使用率 < 85%
- API 响应时间 < 200ms
- 错误率 < 0.1%

---

## 回滚策略

### 快速回滚

```bash
# Docker 回滚
docker service update --rollback gateway

# Kubernetes 回滚
kubectl rollout undo deployment/gateway -n microservice

# 传统部署回滚
cp /backup/bin/gateway.backup /opt/microservice/bin/gateway
systemctl restart gateway
```

---

## 灾难恢复

### 恢复步骤

1. 恢复数据库备份
2. 恢复配置文件
3. 重新部署服务
4. 验证服务功能
5. 切换流量

### RTO 和 RPO 目标

- **RTO**（恢复时间目标）: < 1 小时
- **RPO**（恢复点目标）: < 15 分钟

---

## 检查清单

部署前检查：

- [ ] 所有配置文件已更新
- [ ] 环境变量已设置
- [ ] 数据库已初始化
- [ ] 依赖服务正常运行
- [ ] SSL 证书有效
- [ ] 监控和告警已配置
- [ ] 备份策略已实施
- [ ] 负载均衡器已配置
- [ ] 防火墙规则已设置
- [ ] 文档已更新

部署后验证：

- [ ] 健康检查通过
- [ ] API 接口可访问
- [ ] 数据库连接正常
- [ ] Redis 连接正常
- [ ] 消息队列正常
- [ ] 日志正常记录
- [ ] 监控指标正常
- [ ] 性能测试通过

