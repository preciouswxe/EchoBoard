# EchoBoard

## 技术栈

Go 1.24 + Gin + SQLX + MySQL + Redis + JWT + Zap + Viper + Snowflake + validator + swagger 

## 功能亮点
- 用户注册 / 登录与基于 JWT 的 Token 管理
- 帖子投票系统，支持 Redis 缓存
- 配置热加载和日志管理
- 基于雪花算法分布式 ID 生成 (Snowflake)
- 支持 Redis pipeline 批量读写提高性能
- 支持优雅关停，捕获系统信号并在超时内完成请求，确保服务平滑下线

## 项目启动命令


```bash
go run main.go
```

## 已配置热加载

```bash
air
```

## 访问接口文档

```bash
swag init
```
访问：
http://127.0.0.1:8081/swagger/index.html

