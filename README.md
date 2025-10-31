# Golang 微服务基础架构

这是一个基于 Golang 和 Gin 框架的高性能微服务基础系统，包含了企业级应用所需的所有核心组件。

## 项目简介

本项目提供了一套完整的微服务解决方案，包括：
- 🚪 **API 网关** - 基于 Gin 框架的高性能 HTTP 网关
- 📨 **消息队列** - RabbitMQ 集成，支持异步消息处理
- ⏰ **定时任务** - Cron 定时服务，支持分布式任务调度
- 🔌 **gRPC 服务** - 高性能 RPC 通信服务
- ☁️ **AWS S3** - 文件上传和存储服务
- 📝 **日志系统** - 基于 Zap 的结构化日志
- 🗄️ **PostgreSQL** - 关系型数据库连接
- 🔴 **Redis** - 缓存和分布式锁

## 项目结构

```
microService/
├── cmd/                    # 应用程序入口
│   ├── gateway/           # 网关服务
│   ├── grpc-server/       # gRPC 服务
│   └── cron-server/       # 定时任务服务
├── internal/              # 内部代码包
│   ├── config/           # 配置管理
│   ├── database/         # 数据库连接
│   ├── cache/            # Redis 缓存
│   ├── logger/           # 日志系统
│   ├── queue/            # 消息队列
│   ├── storage/          # 文件存储(S3)
│   ├── handler/          # HTTP 处理器
│   ├── service/          # 业务逻辑层
│   └── middleware/       # 中间件
├── proto/                 # gRPC proto 文件
├── config/               # 配置文件
│   └── config.yaml       # 主配置文件
├── scripts/              # 脚本文件
├── go.mod               # Go 模块依赖
└── README.md            # 项目说明文档
```

## 功能说明

### 1. API 网关服务
- **用途**: 提供 RESTful API 接口，作为所有 HTTP 请求的入口
- **端口**: 默认 8080
- **功能**: 
  - 路由管理
  - 请求日志记录
  - CORS 跨域支持
  - 请求限流
  - 统一错误处理

### 2. 消息队列
- **用途**: 异步消息处理，解耦服务间通信
- **实现**: RabbitMQ
- **功能**:
  - 消息发布/订阅
  - 消息确认机制
  - 死信队列处理
  - 自动重连机制

### 3. 定时任务服务
- **用途**: 执行周期性任务
- **实现**: robfig/cron
- **功能**:
  - Cron 表达式支持
  - 任务日志记录
  - 分布式锁防止重复执行
  - 动态添加/删除任务

### 4. gRPC 服务
- **用途**: 微服务间高性能通信
- **端口**: 默认 50051
- **功能**:
  - Protocol Buffers 序列化
  - 双向流支持
  - 服务健康检查
  - 拦截器支持

### 5. AWS S3 上传服务
- **用途**: 文件存储和管理
- **功能**:
  - 文件上传
  - 文件下载
  - 预签名 URL 生成
  - 文件删除

### 6. 日志系统
- **用途**: 统一的日志记录和追踪
- **实现**: uber-go/zap
- **功能**:
  - 结构化日志
  - 日志分级（Debug/Info/Warn/Error）
  - 日志文件轮转
  - 请求 ID 追踪

### 7. PostgreSQL 数据库
- **用途**: 持久化数据存储
- **实现**: gorm
- **功能**:
  - 连接池管理
  - 自动重连
  - 事务支持
  - SQL 日志记录

### 8. Redis 缓存
- **用途**: 高速缓存和分布式锁
- **实现**: go-redis
- **功能**:
  - 键值存储
  - 过期时间设置
  - 分布式锁
  - 发布/订阅

## 快速开始

### 环境要求
- Go 1.21+
- PostgreSQL 13+
- Redis 6+
- RabbitMQ 3.9+
- AWS 账号（用于 S3）

### 安装依赖
```bash
go mod download
```

### 配置文件
编辑 `config/config.yaml` 文件，填写您的配置信息：
- 数据库连接信息
- Redis 连接信息
- RabbitMQ 连接信息
- AWS S3 凭证

### 运行服务

#### 启动网关服务
```bash
go run cmd/gateway/main.go
```

#### 启动 gRPC 服务
```bash
go run cmd/grpc-server/main.go
```

#### 启动定时任务服务
```bash
go run cmd/cron-server/main.go
```

## API 接口文档

### 健康检查
- **URL**: `GET /health`
- **说明**: 检查服务健康状态
- **返回**: 
```json
{
  "status": "ok",
  "timestamp": "2025-10-31T10:00:00Z"
}
```

### 文件上传
- **URL**: `POST /api/v1/upload`
- **说明**: 上传文件到 S3
- **参数**: 
  - `file`: 文件内容（multipart/form-data）
- **返回**: 
```json
{
  "url": "https://s3.amazonaws.com/bucket/file.jpg",
  "key": "uploads/file.jpg"
}
```

### 发送消息
- **URL**: `POST /api/v1/message`
- **说明**: 发送消息到队列
- **参数**: 
```json
{
  "queue": "task_queue",
  "message": "your message"
}
```

## 技术栈

- **Web 框架**: Gin
- **gRPC**: google.golang.org/grpc
- **数据库**: GORM + PostgreSQL
- **缓存**: go-redis
- **消息队列**: RabbitMQ (streadway/amqp)
- **定时任务**: robfig/cron
- **日志**: uber-go/zap
- **配置**: viper
- **文件存储**: AWS SDK for Go

## 最佳实践

1. **错误处理**: 所有错误都会被记录到日志系统，并返回统一的错误格式
2. **性能监控**: 关键操作都有耗时监控和日志记录
3. **优雅关闭**: 所有服务都支持优雅关闭，确保正在处理的请求完成
4. **配置管理**: 使用环境变量覆盖配置文件，方便不同环境部署
5. **安全性**: 敏感信息不记录到日志，使用环境变量管理密钥

## 注意事项

1. 首次运行前，请确保所有外部服务（PostgreSQL、Redis、RabbitMQ）已启动
2. AWS S3 需要配置正确的访问密钥和区域
3. 生产环境请修改默认的密钥和密码
4. 建议使用 Docker Compose 统一管理所有服务

## 项目改进计划

### 短期改进（1-2周）
- [ ] 完善单元测试覆盖率（目标 >80%）
- [ ] 添加集成测试
- [ ] 补充 API 文档（Swagger/OpenAPI）
- [ ] 添加性能测试脚本
- [ ] 完善错误处理和日志记录

### 中期改进（1-2月）
- [ ] 添加服务注册与发现（Consul/Etcd）
- [ ] 添加链路追踪（Jaeger/Zipkin）
- [ ] 添加指标监控（Prometheus + Grafana）
- [ ] 实现熔断器和限流（Hystrix/Sentinel）
- [ ] 添加 Docker 和 Kubernetes 完整部署配置
- [ ] 实现配置中心（Apollo/Nacos）

### 长期改进（3-6月）
- [ ] 添加 CI/CD 流程（Jenkins/GitLab CI）
- [ ] 实现灰度发布
- [ ] 添加 API 网关高级功能（认证、鉴权、限流）
- [ ] 实现分布式事务（Saga/TCC）
- [ ] 添加全链路监控
- [ ] 实现多租户支持

## 项目反思与总结

### 优点
1. **模块化设计**：各个功能模块职责清晰，便于维护和扩展
2. **完善的注释**：所有函数都有详细的注释说明，降低学习成本
3. **统一的错误处理**：使用日志系统记录所有错误，便于排查问题
4. **配置化管理**：通过 YAML 文件集中管理配置，支持环境变量覆盖
5. **优雅关闭**：所有服务都实现了优雅关闭，避免数据丢失
6. **易于部署**：提供多种部署方式（Docker、Kubernetes、传统部署）

### 可改进之处
1. **测试覆盖**：当前缺少完整的单元测试和集成测试
2. **监控告警**：缺少完整的监控指标和告警机制
3. **服务发现**：当前是硬编码的服务地址，应该使用服务发现
4. **认证授权**：缺少完整的认证授权机制（JWT/OAuth2）
5. **限流策略**：限流中间件较为简单，应该使用 Redis 实现分布式限流
6. **文档完善**：需要补充更多的架构图和流程图

### 潜在问题
1. **分布式事务**：当前没有处理分布式事务问题
2. **数据一致性**：缓存和数据库的一致性需要进一步优化
3. **并发控制**：高并发场景下需要更完善的并发控制机制
4. **资源限制**：需要设置合理的资源限制防止资源耗尽

### 建议使用场景
- ✅ 中小型微服务项目
- ✅ 企业内部系统
- ✅ API 服务平台
- ✅ 学习微服务架构
- ⚠️ 超高并发场景（需要进一步优化）
- ⚠️ 金融级系统（需要加强安全性和事务处理）

## 相关文档

- [快速开始指南](QUICKSTART.md) - 5分钟快速启动项目
- [API 接口文档](API.md) - 详细的 API 接口说明
- [部署指南](DEPLOYMENT.md) - 生产环境部署说明

## 开发者

本项目为微服务基础架构模板，可根据实际业务需求进行扩展和定制。

**设计理念**：简单、可靠、易扩展

**技术选型原则**：
1. 优先使用成熟稳定的技术
2. 避免过度设计
3. 注重代码可读性和可维护性
4. 遵循 SOLID 原则

## 贡献指南

欢迎贡献代码！请遵循以下步骤：

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

### 代码规范

- 遵循 Go 语言官方代码规范
- 所有公开函数必须有注释
- 提交前运行 `go fmt` 和 `go vet`
- 新功能需要添加相应的测试

## 许可证

本项目采用 MIT 许可证 - 详见 LICENSE 文件

## 致谢

感谢以下开源项目：

- [Gin](https://github.com/gin-gonic/gin) - HTTP Web 框架
- [GORM](https://gorm.io/) - ORM 库
- [Zap](https://github.com/uber-go/zap) - 日志库
- [Viper](https://github.com/spf13/viper) - 配置管理
- [gRPC](https://grpc.io/) - RPC 框架

---

**最后更新**: 2025-10-31

如有问题或建议，欢迎提 Issue 或 Pull Request！

