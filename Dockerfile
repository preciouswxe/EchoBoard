FROM golang:alpine AS builder

# 为我们的镜像设置必要的环境变量
ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# 移动到工作目录：/build
WORKDIR /build

# 将代码复制到容器中
COPY . .

# 下载依赖信息
RUN go mod download

# 将我们的代码编译成二进制可执行文件 bubble
RUN go build -o echoboard_app .

###################
# 接下来创建一个小镜像
###################
FROM debian:bullseye-slim

# 把脚本复制到根目录
COPY ./wait-for.sh /

# 从 builder 镜像中把静态文件拷贝到当前目录
COPY ./templates /templates
COPY ./static /static

# 从builder镜像中把配置文件拷贝到当前目录
COPY ./conf /conf

# 从builder镜像中把/dist/app 拷贝到当前目录
COPY --from=builder /build/echoboard_app /

# 更新依赖并安装 给予脚本运行权限
RUN set -eux; \
    apt-get update; \
    apt-get install -y \
        --no-install-recommends \
        netcat; \
        chmod 755 wait-for.sh

# 声明服务端口 （不是执行只是说明）
EXPOSE 8081

# 需要运行的命令（使用 docker-compose 时不需要）
#ENTRYPOINT ["/echoboard_app", "conf/config.yaml"]
