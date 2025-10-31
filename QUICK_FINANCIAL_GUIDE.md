# 金融级系统快速入门指南

Spark大人，这份指南帮助您在**15分钟内**快速了解和使用金融级功能。

---

## 🎯 3 分钟速览

### 核心升级点
1. **数据加密** - AES-256-GCM 加密敏感数据 ✅
2. **JWT 认证** - 保护 API 接口，防止未授权访问 ✅
3. **审计日志** - 记录所有操作，满足合规要求 ⚠️
4. **幂等性** - 防止重复扣款等问题 ⚠️
5. **分布式事务** - 保证数据一致性 ⚠️
6. **限流熔断** - 防止系统过载 ⚠️

### 当前状态
- ✅ **已完成**: 数据加密、JWT 认证（可立即使用）
- ⚠️ **设计完成**: 其他功能（需要实施）
- 📚 **文档完备**: 60+页详细方案

---

## 🚀 5 分钟快速使用

### 1. 使用数据加密

```go
package main

import (
    "fmt"
    "github.com/zhang/microservice/internal/security"
)

func main() {
    // 创建加密器（密钥必须是32字节）
    encryptor, err := security.NewEncryptor("12345678901234567890123456789012")
    if err != nil {
        panic(err)
    }

    // 加密用户手机号
    phone := "13800138000"
    encrypted, _ := encryptor.Encrypt(phone)
    fmt.Println("加密后:", encrypted)
    // 输出: YxRtZm... (Base64密文)

    // 解密
    decrypted, _ := encryptor.Decrypt(encrypted)
    fmt.Println("解密后:", decrypted)
    // 输出: 13800138000

    // 日志中脱敏显示
    masked := security.MaskSensitiveData(phone, "phone")
    fmt.Println("脱敏显示:", masked)
    // 输出: 138****8000
}
```

### 2. 使用 JWT 认证

#### 2.1 配置 JWT
```go
// cmd/gateway/main.go

import "github.com/zhang/microservice/internal/middleware"

func main() {
    // 设置 JWT 配置
    middleware.SetJWTConfig(&middleware.JWTConfig{
        Secret:     []byte(os.Getenv("JWT_SECRET")),
        ExpireTime: 2 * time.Hour, // 金融系统建议2小时
    })

    router := gin.New()
    // ... 其他配置
}
```

#### 2.2 保护路由
```go
// 公开路由（无需认证）
router.POST("/api/v1/login", handler.Login())

// 需要认证的路由
authorized := router.Group("/api/v1")
authorized.Use(middleware.JWTAuth())
{
    // 所有用户都可以访问
    authorized.GET("/profile", handler.GetProfile())
    
    // 只有管理员可以访问
    admin := authorized.Group("/admin")
    admin.Use(middleware.RequireRole("admin"))
    {
        admin.GET("/users", handler.ListUsers())
        admin.DELETE("/users/:id", handler.DeleteUser())
    }
    
    // 只有财务人员可以访问
    finance := authorized.Group("/finance")
    finance.Use(middleware.RequireRole("admin", "finance"))
    {
        finance.POST("/transfer", handler.Transfer())
    }
}
```

#### 2.3 登录处理
```go
// internal/handler/auth.go

func Login() gin.HandlerFunc {
    return func(c *gin.Context) {
        var req LoginRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(400, gin.H{"error": "参数错误"})
            return
        }

        // 验证用户名密码（从数据库查询）
        user, err := userService.Login(req.Username, req.Password)
        if err != nil {
            c.JSON(401, gin.H{"error": "用户名或密码错误"})
            return
        }

        // 生成 JWT token
        token, err := middleware.GenerateToken(
            user.ID,
            user.Username,
            user.Role,
        )
        if err != nil {
            c.JSON(500, gin.H{"error": "生成令牌失败"})
            return
        }

        c.JSON(200, gin.H{
            "token": token,
            "user": gin.H{
                "id":       user.ID,
                "username": user.Username,
                "role":     user.Role,
            },
        })
    }
}
```

#### 2.4 在业务代码中获取用户信息
```go
func GetProfile() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 获取当前用户 ID
        userID, _ := middleware.GetUserID(c)
        
        // 查询用户信息
        user, err := userService.GetUser(c.Request.Context(), userID)
        if err != nil {
            c.JSON(500, gin.H{"error": "查询失败"})
            return
        }

        c.JSON(200, user)
    }
}
```

---

## 📱 客户端使用示例

### 1. 登录获取 token
```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "password123"
  }'

# 响应
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "admin",
    "role": "admin"
  }
}
```

### 2. 使用 token 访问保护接口
```bash
curl -X GET http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### 3. JavaScript 示例
```javascript
// 登录
async function login(username, password) {
    const response = await fetch('http://localhost:8080/api/v1/login', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ username, password })
    });
    
    const data = await response.json();
    // 保存 token
    localStorage.setItem('token', data.token);
    return data;
}

// 调用保护接口
async function getProfile() {
    const token = localStorage.getItem('token');
    const response = await fetch('http://localhost:8080/api/v1/profile', {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    });
    
    return await response.json();
}
```

---

## 💡 实战场景

### 场景 1: 用户注册（加密敏感信息）

```go
func Register(c *gin.Context) {
    var req RegisterRequest
    c.ShouldBindJSON(&req)

    // 创建加密器
    encryptor, _ := security.NewEncryptor(config.EncryptionKey)

    // 加密敏感信息
    encryptedPhone, _ := encryptor.Encrypt(req.Phone)
    encryptedIDCard, _ := encryptor.Encrypt(req.IDCard)

    // 创建用户
    user := &User{
        Username: req.Username,
        Phone:    encryptedPhone,   // 存储加密后的数据
        IDCard:   encryptedIDCard,
    }

    // 保存到数据库
    db.Create(user)

    c.JSON(200, gin.H{"message": "注册成功"})
}
```

### 场景 2: 查询用户（解密显示）

```go
func GetUser(c *gin.Context) {
    userID, _ := middleware.GetUserID(c)

    // 从数据库查询
    var user User
    db.First(&user, userID)

    // 解密敏感信息
    encryptor, _ := security.NewEncryptor(config.EncryptionKey)
    decryptedPhone, _ := encryptor.Decrypt(user.Phone)

    // 返回（脱敏显示）
    c.JSON(200, gin.H{
        "id":       user.ID,
        "username": user.Username,
        "phone":    security.MaskSensitiveData(decryptedPhone, "phone"),
    })
}
```

### 场景 3: 转账（需要高级权限）

```go
func Transfer(c *gin.Context) {
    var req TransferRequest
    c.ShouldBindJSON(&req)

    // 获取当前用户
    userID, _ := middleware.GetUserID(c)
    userRole, _ := middleware.GetUserRole(c)

    // 检查权限
    if userRole != "admin" && userRole != "finance" {
        c.JSON(403, gin.H{"error": "权限不足"})
        return
    }

    // 执行转账
    err := transferService.Transfer(
        c.Request.Context(),
        req.FromAccount,
        req.ToAccount,
        req.Amount,
    )

    if err != nil {
        c.JSON(500, gin.H{"error": "转账失败"})
        return
    }

    c.JSON(200, gin.H{"message": "转账成功"})
}
```

---

## 🔧 配置说明

### 环境变量配置

```bash
# .env
JWT_SECRET=your-very-long-secret-key-here-change-in-production
ENCRYPTION_KEY=12345678901234567890123456789012
```

### 配置文件

```yaml
# config/config.yaml
security:
  jwt:
    secret_key: ${JWT_SECRET}
    expire_hours: 2
  encryption:
    key: ${ENCRYPTION_KEY}
    algorithm: AES-256-GCM
```

---

## 📋 检查清单

### 上线前必做
- [ ] 更换 JWT 密钥为随机生成的强密钥
- [ ] 更换加密密钥为32字节随机密钥
- [ ] 配置 HTTPS/TLS 证书
- [ ] 设置合理的 token 过期时间（建议2小时）
- [ ] 启用审计日志
- [ ] 进行安全测试

### 最佳实践
- [ ] 密钥使用环境变量，不写入代码
- [ ] 生产环境使用密钥管理系统（AWS KMS等）
- [ ] 定期轮换密钥
- [ ] 监控异常登录
- [ ] 实施IP白名单（如果需要）

---

## ⚠️ 常见问题

### Q1: token 过期怎么办？
A: 实现 token 刷新机制，或要求用户重新登录。

```go
// 刷新 token
newToken, _ := middleware.RefreshToken(oldToken)
```

### Q2: 密钥泄露怎么办？
A: 立即更换密钥，强制所有用户重新登录。

### Q3: 如何实现单点登录？
A: 使用 Redis 存储 token，实现 token 黑名单机制。

### Q4: 如何实现多设备互踢？
A: 在 Redis 中维护用户活跃 token，新登录时使旧 token 失效。

---

## 📚 深入学习

### 完整文档
1. **FINANCIAL_GRADE_UPGRADE.md** - 60+页技术方案
2. **IMPLEMENTATION_PRIORITY.md** - 实施优先级
3. **FINANCIAL_SUMMARY.md** - 项目总结

### 代码示例
1. **internal/security/encryption.go** - 加密实现
2. **internal/security/encryption_test.go** - 测试示例
3. **internal/middleware/auth.go** - 认证实现

---

## 🎯 下一步

### 今天可以做
1. ✅ 运行加密测试: `go test -v internal/security/...`
2. ✅ 在代码中集成 JWT 认证
3. ✅ 测试登录和权限控制

### 本周计划
1. ⚠️ 实施审计日志系统
2. ⚠️ 实施幂等性保证
3. ⚠️ 配置 HTTPS
4. ⚠️ 编写集成测试

### 技术支持
- 查看代码注释
- 运行测试用例
- 阅读详细文档

---

**创建时间**: 2025-10-31
**适用人群**: 开发人员、架构师
**预计阅读**: 15分钟
**上手时间**: 30分钟

祝您使用愉快！如有问题，请查看详细文档。

