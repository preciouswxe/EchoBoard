#!/bin/bash

# Windows Git Bash 下 Go 绝对路径
GO_EXE="/mnt/d/Golang/bin/go.exe"

# 如果 GO_EXE 不存在,尝试自动检测
if [ ! -f "$GO_EXE" ]; then
    echo "警告: $GO_EXE 不存在，尝试自动检测 Go..."
    GO_EXE=$(which go 2>/dev/null || echo "")
    if [ -z "$GO_EXE" ]; then
        echo "错误: 找不到 Go，请设置正确的 GO_EXE 路径"
        exit 1
    fi
    echo "使用: $GO_EXE"
fi

trap "echo 'Stopping...'; kill 0" EXIT

echo "=============================="
echo "Starting EchoBoard Dev Environment"
echo "=============================="
echo "Go 路径: $GO_EXE"
echo ""

# ================= 启动 Docker 容器 =================
echo "[1/4] 启动 Docker 容器..."
containers=("zoo1" "kafka1" "redisdb" "elasticsearch" "kibana" "kafka-ui")
for c in "${containers[@]}"; do
    status=$(docker inspect -f '{{.State.Status}}' $c 2>/dev/null || echo "notfound")
    if [ "$status" = "running" ]; then
        echo "✓ 容器 $c 已运行"
    elif [ "$status" = "exited" ]; then
        echo "→ 启动容器 $c ..."
        docker start $c
    else
        echo "✗ 容器 $c 不存在，请先创建"
    fi
done

echo ""
echo "等待服务就绪..."
sleep 5

# ================= 启动 Go API =================
echo "[2/4] 启动 Go API..."
"$GO_EXE" run ./main.go > tmp/api.log 2>&1 &
API_PID=$!
echo "  PID: $API_PID"

# ================= 启动 Kafka Consumer =================
echo "[3/4] 启动 Kafka Consumer..."
"$GO_EXE" run ./cmd/consumer/*.go > tmp/consumer.log 2>&1 &
CONSUMER_PID=$!
echo "  PID: $CONSUMER_PID"

# ================= 启动 Vue 前端 =================
echo "[4/4] 启动 Vue 前端..."
(cd ../EchoBoard_frontend && npm run dev > ../EchoBoard_backend/tmp/frontend.log 2>&1) &
FRONTEND_PID=$!
echo "  PID: $FRONTEND_PID"

echo ""
echo "=============================="
echo "所有服务已启动！"
echo "=============================="
echo "日志文件:"
echo "  - API:      api.log"
echo "  - Consumer: consumer.log"
echo "  - Frontend: frontend.log"
echo ""
echo "实时查看日志:"
echo "  tail -f api.log"
echo "  tail -f consumer.log"
echo "  tail -f frontend.log"
echo ""
echo "按 Ctrl+C 停止所有服务"
echo "=============================="

wait