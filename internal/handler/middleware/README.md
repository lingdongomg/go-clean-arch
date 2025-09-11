# 错误处理中间件

这个包提供了一个通用的错误处理中间件，用于统一处理 Gin 框架中的错误响应。

## 功能特性

- 🎯 **统一错误响应格式**：所有错误都返回一致的 JSON 格式
- 🏷️ **自定义错误类型**：支持应用级别的自定义错误
- 📝 **结构化日志**：自动记录错误日志，包含请求上下文信息
- 🔍 **错误分类**：区分客户端错误（4xx）和服务器错误（5xx）
- 🌐 **国际化支持**：提供中文错误消息
- 🛡️ **安全性**：避免敏感信息泄露
- 🔄 **双重保护**：支持 panic 恢复和手动错误处理

## 使用方法

### 1. 注册中间件

```go
package main

import (
    "os"
    "github.com/gin-gonic/gin"
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
    "your-project/internal/handler/middleware"
)

func main() {
    r := gin.New()
    
    // 配置 zerolog
    zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
    log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
    
    // 注册错误处理中间件
    r.Use(middleware.ErrorHandler())      // panic 恢复
    r.Use(middleware.ErrorMiddleware())   // 手动错误处理
    
    // 你的路由...
    r.Run(":8080")
}
```

### 2. 在业务代码中使用

```go
func GetArticle(c *gin.Context) {
    id := c.Param("id")
    if id == "" {
        // 使用预定义错误
        middleware.HandleError(c, middleware.ErrBadRequest)
        return
    }
    
    article, err := articleService.GetByID(id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            // 创建自定义错误
            middleware.HandleError(c, middleware.NewAppError(404, "文章不存在", "article not found"))
            return
        }
        // 包装原始错误
        middleware.HandleError(c, middleware.NewAppErrorWithErr(500, "获取文章失败", err))
        return
    }
    
    c.JSON(200, article)
}
```

### 3. Panic 处理

```go
func DangerousHandler(c *gin.Context) {
    // 如果这里发生 panic，ErrorHandler 会自动捕获并处理
    panic("something went wrong")
}
```

## 错误类型

### 预定义错误

- `ErrBadRequest` (400): 请求参数错误
- `ErrUnauthorized` (401): 未授权访问
- `ErrForbidden` (403): 禁止访问
- `ErrNotFound` (404): 资源不存在
- `ErrConflict` (409): 资源冲突
- `ErrInternalServerError` (500): 服务器内部错误

### 自定义错误

```go
// 创建简单的自定义错误
err := middleware.NewAppError(400, "用户名不能为空", "username is required")

// 创建包含原始错误的自定义错误
err := middleware.NewAppErrorWithErr(500, "数据库操作失败", originalErr)
```

## 响应格式

所有错误都会返回统一的 JSON 格式：

```json
{
    "code": 404,
    "message": "资源不存在",
    "details": "article with id 123 not found"
}
```

- `code`: HTTP 状态码
- `message`: 用户友好的错误消息
- `details`: 可选的详细错误信息（仅在开发环境或需要时提供）

## 日志记录

中间件会自动记录错误日志，包含以下信息：

- HTTP 方法和 URI
- 用户代理和客户端 IP
- 错误详情
- 根据错误类型选择日志级别（客户端错误用 WARN，服务器错误用 ERROR）

## 中间件对比

### ErrorHandler vs ErrorMiddleware

| 特性 | ErrorHandler | ErrorMiddleware |
|------|-------------|-----------------|
| 用途 | Panic 恢复 | 手动错误处理 |
| 触发方式 | 自动（panic时） | 手动（调用HandleError） |
| 使用场景 | 意外错误 | 业务逻辑错误 |
| 推荐 | 必须使用 | 推荐使用 |

## 最佳实践

1. **使用预定义错误**：对于常见的 HTTP 错误，优先使用预定义的错误类型
2. **提供有意义的错误消息**：错误消息应该对用户友好，避免技术细节
3. **保护敏感信息**：不要在错误响应中暴露敏感的系统信息
4. **适当使用 details 字段**：仅在开发环境或确实需要时提供详细错误信息
5. **包装原始错误**：使用 `NewAppErrorWithErr` 来保留原始错误信息用于日志记录
6. **双重保护**：同时使用 ErrorHandler 和 ErrorMiddleware 确保完整的错误处理覆盖

## 与 Echo 框架的区别

- **中间件注册**：使用 `r.Use()` 而不是 `e.Use()`
- **上下文对象**：使用 `*gin.Context` 而不是 `echo.Context`
- **错误处理**：使用 `HandleError()` 函数而不是直接返回错误
- **路由参数**：使用 `c.Param()` 和 `c.Query()` 获取参数
- **响应方法**：使用 `c.JSON()` 而不是 `c.JSON()`（参数顺序相同）