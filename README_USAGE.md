# 服务监控系统使用说明

## 概述

这是一个用 Go 语言开发的服务监控系统，主要功能包括：

1. **Kurtosis Enclave 监控**: 监控 Kurtosis enclave 的运行状态，自动重启停止的 enclave
2. **进程监控**: 监控指定的系统进程，检测进程停止并报警

## 功能特性

### Kurtosis 监控
- 监控配置的 enclave 列表
- 检测到停止的 enclave 时自动尝试重启
- 支持配置最大重启尝试次数（默认3次）
- 连续重启失败后停止尝试并报警

### 进程监控
- 监控配置的进程列表
- 检测进程停止时立即报警
- 支持进程名称模糊匹配

### 通用特性
- 可配置的监控间隔时间
- 详细的日志记录（文件+控制台）
- 支持 systemd 服务管理
- 优雅的信号处理和关闭

## 安装和部署

### 方法一：只安装 binary

```
go install github.com/conflux-fans/linux-service-monitor
```

### 方法二：使用脚本自动化服务（推荐）

```bash
# 给安装脚本执行权限
chmod +x install.sh

# 运行安装脚本（需要 root 权限）
sudo ./install.sh
```

### 方法二：手动安装

```bash
# 1. 构建程序
make build

# 2. 复制配置文件
sudo cp config.yaml /etc/service_monitor.yaml

# 3. 复制二进制文件
sudo cp bin/service_monitor /usr/local/bin/

# 4. 创建 systemd 服务
make systemd
sudo cp service_monitor.service /etc/systemd/system/
sudo systemctl daemon-reload
```

## 配置说明

示例文件为 config.sample.yaml。 配置文件使用 YAML 格式，主要包含以下部分：
- Kurtosis 配置
- 进程监控配置
- 日志配置
- 报警（Alert）配置

## 使用方法

### 开发和测试

```bash
# 开发模式运行
make dev

# 使用自定义配置文件
go run main.go -config=test_config.yaml

# 构建程序
make build

# 运行构建的程序
make run
```

### 生产环境

```bash
# 启动服务
sudo systemctl start service_monitor

# 停止服务
sudo systemctl stop service_monitor

# 启用开机自启
sudo systemctl enable service_monitor

# 查看服务状态
sudo systemctl status service_monitor

# 查看实时日志
sudo journalctl -u service_monitor -f

# 查看历史日志
sudo journalctl -u service_monitor --since "1 hour ago"
```

## 日志说明

系统会生成详细的日志，包括：

- **INFO**: 正常运行信息，如服务启动、检查结果等
- **WARN**: 警告信息，如检测到服务停止
- **ERROR**: 错误信息，如重启失败、命令执行错误等

日志同时输出到：
1. 配置的日志文件（默认 `monitor.log`）
2. 控制台（开发模式）
3. systemd journal（服务模式）

## 故障排除

### 常见问题

1. **Kurtosis 命令不可用**
   - 确保已安装 Kurtosis CLI
   - 检查 PATH 环境变量

2. **权限问题**
   - 确保程序有足够权限执行系统命令
   - 建议以 root 用户运行

3. **配置文件错误**
   - 检查 YAML 语法是否正确
   - 验证文件路径是否存在

### 调试方法

```bash
# 检查配置文件语法
go run main.go -config=config.yaml

# 查看详细日志
tail -f monitor.log

# 检查系统服务状态
sudo systemctl status service_monitor

# 手动测试 Kurtosis 命令
kurtosis enclave ls

# 手动测试进程检查
pgrep -f nginx
```

## 扩展开发

### 添加新的监控类型

1. 在 `monitor/` 目录下创建新的监控器
2. 实现 `Start()` 方法
3. 在 `main.go` 中集成新监控器
4. 在配置文件中添加相应配置

### 自定义报警方式

当前系统仅支持日志报警，可以扩展支持：
- 邮件通知
- 钉钉/企业微信机器人
- Slack 通知
- HTTP Webhook

## 依赖要求

- Go 1.21+
- Linux 系统
- Kurtosis CLI（如果使用 Kurtosis 监控）
- systemd（如果使用服务模式）

## 许可证

本项目采用 MIT 许可证。 