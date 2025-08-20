# EchoBoard

## 技术栈

Go 1.24 + Gin + SQLX + MySQL + Redis + JWT + Zap + Viper + Snowflake + validator 


## 项目启动命令


```bash
go run main.go
```

## 已配置热加载

```bash
air
```

## 功能亮点
- 用户注册 / 登录 token 管理
- 帖子投票系统，支持 Redis 缓存
- 配置热加载和日志管理
- 支持分布式 ID 生成 (Snowflake)