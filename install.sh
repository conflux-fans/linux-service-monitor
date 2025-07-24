#!/bin/bash

# 服务监控系统安装脚本

set -e

echo "=== 服务监控系统安装脚本 ==="

# 检查是否为 root 用户
if [[ $EUID -ne 0 ]]; then
   echo "此脚本需要 root 权限运行"
   echo "请使用: sudo $0"
   exit 1
fi

# 检查 Go 是否已安装
if ! command -v go &> /dev/null; then
    echo "错误: 未找到 Go 环境，请先安装 Go"
    exit 1
fi

echo "检测到 Go 版本: $(go version)"

# 创建安装目录
INSTALL_DIR="/opt/service_monitor"
CONFIG_FILE="/etc/service_monitor.yaml"
BINARY_FILE="/usr/local/bin/service_monitor"
SERVICE_FILE="/etc/systemd/system/service_monitor.service"

echo "创建安装目录: $INSTALL_DIR"
mkdir -p $INSTALL_DIR

# 复制文件
echo "安装依赖..."
go mod tidy

echo "构建程序..."
go build -o service_monitor main.go

echo "安装二进制文件到 $BINARY_FILE"
cp service_monitor $BINARY_FILE
chmod +x $BINARY_FILE

echo "安装配置文件到 $CONFIG_FILE"
cp config.yaml $CONFIG_FILE

# 创建 systemd 服务文件
echo "创建 systemd 服务文件..."
cat > $SERVICE_FILE << EOF
[Unit]
Description=Service Monitor - Kurtosis and Process Monitor
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
ExecStart=$BINARY_FILE -config=$CONFIG_FILE
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# 重新加载 systemd
echo "重新加载 systemd..."
systemctl daemon-reload

echo ""
echo "=== 安装完成 ==="
echo ""
echo "配置文件位置: $CONFIG_FILE"
echo "二进制文件位置: $BINARY_FILE"
echo "服务文件位置: $SERVICE_FILE"
echo ""
echo "使用以下命令管理服务:"
echo "  启动服务: sudo systemctl start service_monitor"
echo "  停止服务: sudo systemctl stop service_monitor"
echo "  启用开机自启: sudo systemctl enable service_monitor"
echo "  查看服务状态: sudo systemctl status service_monitor"
echo "  查看日志: sudo journalctl -u service_monitor -f"
echo ""
echo "请根据需要编辑配置文件: $CONFIG_FILE" 