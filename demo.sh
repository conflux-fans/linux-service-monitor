#!/bin/bash

# 服务监控系统演示脚本

echo "=== 服务监控系统演示 ==="
echo

# 检查依赖
echo "1. 检查依赖环境..."
echo "检查 Go 环境:"
if command -v go &> /dev/null; then
    echo "✓ Go 已安装: $(go version)"
else
    echo "✗ Go 未安装，请先安装 Go 1.21+"
    exit 1
fi

echo "检查 Docker 环境:"
if command -v docker &> /dev/null; then
    echo "✓ Docker 已安装: $(docker --version)"
else
    echo "✗ Docker 未安装，重启功能将不可用"
fi

echo "检查 Kurtosis 环境:"
if command -v kurtosis &> /dev/null; then
    echo "✓ Kurtosis 已安装: $(kurtosis version)"
else
    echo "✗ Kurtosis 未安装，Kurtosis 监控功能将不可用"
fi

echo

# 构建程序
echo "2. 构建监控程序..."
if go build -o bin/service_monitor main.go; then
    echo "✓ 构建成功"
else
    echo "✗ 构建失败"
    exit 1
fi

echo

# 展示配置文件
echo "3. 配置文件说明:"
echo "主配置文件: config.yaml"
cat config.yaml
echo
echo "测试配置文件: test_config.yaml"
cat test_config.yaml
echo

# 运行演示
echo "4. 运行监控程序演示 (使用测试配置)..."
echo "程序将监控以下进程: systemd, bash"
echo "按 Ctrl+C 可以停止程序"
echo

read -p "按 Enter 键开始演示..." -r

echo "启动监控程序..."
./bin/service_monitor -config=test_config.yaml &
MONITOR_PID=$!

echo "监控程序已启动 (PID: $MONITOR_PID)"
echo "等待 30 秒查看监控日志..."

sleep 30

echo
echo "停止监控程序..."
kill $MONITOR_PID 2>/dev/null || true
wait $MONITOR_PID 2>/dev/null || true

echo
echo "5. 查看生成的日志文件:"
if [ -f "test_monitor.log" ]; then
    echo "=== 日志内容 ==="
    tail -20 test_monitor.log
else
    echo "未找到日志文件"
fi

echo
echo "=== 演示完成 ==="
echo
echo "接下来你可以:"
echo "1. 修改 config.yaml 配置你要监控的 enclave 和进程"
echo "2. 运行: ./bin/service_monitor -config=config.yaml"
echo "3. 或者安装为系统服务: sudo ./install.sh"
echo
echo "更多信息请查看 README_USAGE.md" 