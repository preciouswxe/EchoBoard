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
- 支持令牌桶限流中间件

## 项目启动命令


```bash
go run main.go
```

## 已配置热加载启动

```bash
air
```

## 访问接口文档

```bash
swag init
```
访问：
http://127.0.0.1:8081/swagger/index.html

## 压测结果

使用外部压测工具 `hey`（ https://github.com/rakyll/hey ）

示例:（不启动限流中间件）
```
hey -n 10000 -c 100 -cpus 8 "http://127.0.0.1:8081/api/v1/posts" 
```
测试结果：
```
Summary:
  Total:        0.7265 secs
  Slowest:      0.0316 secs
  Fastest:      0.0001 secs
  Average:      0.0072 secs
  Requests/sec: 13764.9356
 
Status code distribution:
  [200] 10000 responses
```

测试结果2（启动令牌桶限流）：

```
Summary:
  Total:        0.8275 secs
  Slowest:      0.0593 secs
  Fastest:      0.0001 secs
  Average:      0.0082 secs
  Requests/sec: 12084.5162

Status code distribution:
  [200] 10 responses
  [429] 9990 responses
```

## pprof 性能分析

示例（内存篇）:

```
go tool pprof -inuse_space http://127.0.0.1:8081/debug/pprof/heap
```

![](photo/pprof_svgtojpg.jpg)

注：可视化工具 [graphviz](https://graphviz.gitlab.io/)
