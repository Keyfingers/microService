# API 接口文档

本文档详细说明了微服务系统提供的所有 HTTP API 接口。

## 基础信息

- **基础 URL**: `http://localhost:8080`
- **内容类型**: `application/json`
- **字符编码**: `UTF-8`

## 通用响应格式

### 成功响应
```json
{
  "data": {},
  "message": "操作成功"
}
```

### 错误响应
```json
{
  "error": "错误描述信息"
}
```

## API 端点

### 1. 健康检查

#### 1.1 基础健康检查

**端点**: `GET /health`

**说明**: 检查服务是否正常运行

**请求示例**:
```bash
curl http://localhost:8080/health
```

**响应示例**:
```json
{
  "status": "ok",
  "timestamp": "2025-10-31T10:00:00Z"
}
```

**响应字段**:
| 字段 | 类型 | 说明 |
|------|------|------|
| status | string | 服务状态，"ok" 表示正常 |
| timestamp | string | ISO 8601 格式的时间戳 |

---

#### 1.2 详细健康检查

**端点**: `GET /health/detail`

**说明**: 检查服务及其所有依赖的健康状态

**请求示例**:
```bash
curl http://localhost:8080/health/detail
```

**响应示例**:
```json
{
  "status": "ok",
  "timestamp": "2025-10-31T10:00:00Z",
  "services": {
    "database": {
      "status": "ok"
    },
    "redis": {
      "status": "ok"
    }
  }
}
```

**响应字段**:
| 字段 | 类型 | 说明 |
|------|------|------|
| status | string | 整体状态："ok" 或 "degraded" |
| timestamp | string | ISO 8601 格式的时间戳 |
| services | object | 各个依赖服务的状态 |
| services.*.status | string | 服务状态："ok" 或 "error" |
| services.*.message | string | 错误信息（仅在出错时） |

**HTTP 状态码**:
- `200`: 所有服务正常
- `503`: 有服务异常

---

### 2. 文件上传

#### 2.1 上传文件到 S3

**端点**: `POST /api/v1/upload`

**说明**: 上传文件到 AWS S3 存储

**请求类型**: `multipart/form-data`

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | file | 是 | 要上传的文件 |

**请求示例**:
```bash
curl -X POST http://localhost:8080/api/v1/upload \
  -F "file=@/path/to/image.jpg"
```

**响应示例**:
```json
{
  "url": "https://your-bucket.s3.amazonaws.com/uploads/image_20251031100000.jpg",
  "key": "uploads/image_20251031100000.jpg"
}
```

**响应字段**:
| 字段 | 类型 | 说明 |
|------|------|------|
| url | string | 文件的完整访问 URL |
| key | string | 文件在 S3 中的 Key |

**错误码**:
- `400`: 未提供文件或文件格式错误
- `500`: 上传失败

---

#### 2.2 获取预签名 URL

**端点**: `GET /api/v1/presigned-url`

**说明**: 生成文件的临时访问 URL（有效期可配置）

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| key | string | 是 | 文件在 S3 中的 Key |

**请求示例**:
```bash
curl "http://localhost:8080/api/v1/presigned-url?key=uploads/image_20251031100000.jpg"
```

**响应示例**:
```json
{
  "url": "https://your-bucket.s3.amazonaws.com/uploads/image.jpg?X-Amz-Algorithm=..."
}
```

**响应字段**:
| 字段 | 类型 | 说明 |
|------|------|------|
| url | string | 预签名的临时访问 URL |

**错误码**:
- `400`: 未提供 key 参数
- `500`: 生成 URL 失败

---

### 3. 消息队列

#### 3.1 发送消息

**端点**: `POST /api/v1/message`

**说明**: 发送消息到消息队列

**请求类型**: `application/json`

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| queue | string | 是 | 队列名称（如 "task", "email"） |
| message | object | 是 | 消息内容（任意 JSON 对象） |

**请求示例**:
```bash
curl -X POST http://localhost:8080/api/v1/message \
  -H "Content-Type: application/json" \
  -d '{
    "queue": "task",
    "message": {
      "type": "send_email",
      "to": "user@example.com",
      "subject": "欢迎注册",
      "body": "感谢您注册我们的服务"
    }
  }'
```

**响应示例**:
```json
{
  "message": "消息发送成功"
}
```

**错误码**:
- `400`: 请求参数错误
- `500`: 消息发送失败

---

## gRPC 接口

### UserService

gRPC 服务提供用户管理功能，端口：`50051`

#### 方法列表

1. **GetUser** - 获取用户信息
2. **CreateUser** - 创建用户
3. **UpdateUser** - 更新用户
4. **DeleteUser** - 删除用户

详细的 gRPC 接口定义请查看 `proto/service.proto` 文件。

---

## 错误代码

| HTTP 状态码 | 说明 |
|------------|------|
| 200 | 请求成功 |
| 400 | 请求参数错误 |
| 401 | 未授权 |
| 403 | 禁止访问 |
| 404 | 资源不存在 |
| 429 | 请求过于频繁 |
| 500 | 服务器内部错误 |
| 503 | 服务不可用 |

---

## 请求示例（完整）

### 使用 cURL

```bash
# 健康检查
curl http://localhost:8080/health

# 上传文件
curl -X POST http://localhost:8080/api/v1/upload \
  -F "file=@./test.jpg"

# 发送消息
curl -X POST http://localhost:8080/api/v1/message \
  -H "Content-Type: application/json" \
  -d '{
    "queue": "task",
    "message": {
      "type": "test",
      "data": "hello"
    }
  }'
```

### 使用 JavaScript (Fetch)

```javascript
// 健康检查
fetch('http://localhost:8080/health')
  .then(res => res.json())
  .then(data => console.log(data));

// 上传文件
const formData = new FormData();
formData.append('file', fileInput.files[0]);

fetch('http://localhost:8080/api/v1/upload', {
  method: 'POST',
  body: formData
})
  .then(res => res.json())
  .then(data => console.log(data));

// 发送消息
fetch('http://localhost:8080/api/v1/message', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    queue: 'task',
    message: {
      type: 'test',
      data: 'hello'
    }
  })
})
  .then(res => res.json())
  .then(data => console.log(data));
```

### 使用 Python (requests)

```python
import requests

# 健康检查
response = requests.get('http://localhost:8080/health')
print(response.json())

# 上传文件
files = {'file': open('test.jpg', 'rb')}
response = requests.post('http://localhost:8080/api/v1/upload', files=files)
print(response.json())

# 发送消息
data = {
    'queue': 'task',
    'message': {
        'type': 'test',
        'data': 'hello'
    }
}
response = requests.post('http://localhost:8080/api/v1/message', json=data)
print(response.json())
```

---

## 速率限制

默认配置：
- 每秒最多 100 个请求
- 突发请求数：200

可在 `config/config.yaml` 中修改：
```yaml
middleware:
  rate_limit:
    enable: true
    requests_per_second: 100
    burst: 200
```

---

## CORS 配置

默认允许所有来源的跨域请求。生产环境建议修改配置：

```yaml
middleware:
  cors:
    enable: true
    allow_origins:
      - "https://your-domain.com"
```

---

## 日志追踪

每个请求都会生成唯一的 `request_id`，可用于日志追踪和问题排查。

查看日志文件：
- 应用日志：`logs/app.log`
- 错误日志：`logs/error.log`

