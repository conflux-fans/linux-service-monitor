# Makefile for Service Monitor

.PHONY: build run clean test install deps

# 默认目标
all: build

# 构建项目
build:
	@echo "构建服务监控程序..."
	go build -o bin/service_monitor main.go

# 运行项目
run: build
	@echo "启动服务监控..."
	./bin/service_monitor -config=config.yaml

# 安装依赖
deps:
	@echo "安装依赖..."
	go mod tidy
	go mod download

# 清理构建文件
clean:
	@echo "清理构建文件..."
	rm -rf bin/
	rm -f *.log

# 测试
test:
	@echo "运行测试..."
	go test ./...

# 安装到系统
install: build
	@echo "安装到系统..."
	sudo cp bin/service_monitor /usr/local/bin/
	sudo cp config.yaml /etc/service_monitor.yaml

# 创建systemd服务文件
systemd:
	@echo "创建systemd服务文件..."
	@cat > service_monitor.service << EOF
[Unit]
Description=Service Monitor
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/service_monitor
ExecStart=/usr/local/bin/service_monitor -config=/etc/service_monitor.yaml
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF
	@echo "服务文件已创建: service_monitor.service"
	@echo "使用以下命令安装服务:"
	@echo "  sudo cp service_monitor.service /etc/systemd/system/"
	@echo "  sudo systemctl daemon-reload"
	@echo "  sudo systemctl enable service_monitor"
	@echo "  sudo systemctl start service_monitor"

# 开发模式运行
dev:
	@echo "开发模式运行..."
	go run main.go -config=config.yaml

# 显示帮助
help:
	@echo "可用的命令:"
	@echo "  build    - 构建项目"
	@echo "  run      - 构建并运行项目"
	@echo "  dev      - 开发模式运行"
	@echo "  deps     - 安装依赖"
	@echo "  clean    - 清理构建文件"
	@echo "  test     - 运行测试"
	@echo "  install  - 安装到系统"
	@echo "  systemd  - 创建systemd服务文件"
	@echo "  help     - 显示此帮助信息" 