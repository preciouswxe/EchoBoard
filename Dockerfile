# ---------------- 构建阶段 ----------------
FROM golang:alpine AS builder

# 设置环境变量
ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build

# 下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 拷贝代码
COPY . .

# 编译 Go API 和 Consumer
RUN go build -o echoboard_app .                           # 编译 main.go
RUN go build -o consumer_app cmd/consumer/consumer.go     # 编译 Consumer

# ---------------- 运行阶段 ----------------
FROM debian:bullseye-slim

WORKDIR /app

# 拷贝 wait-for 脚本
COPY ./wait-for.sh /

# 拷贝配置文件
COPY ./conf /conf

# 从 builder 镜像拷贝可执行文件
COPY --from=builder /build/echoboard_app .
COPY --from=builder /build/consumer_app .

# 安装依赖工具并授权 wait-for.sh
RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends netcat; \
    chmod 755 wait-for.sh

# 声明服务端口
EXPOSE 8081

# Docker Compose 会覆盖 CMD/command，这里不写 ENTRYPOINT