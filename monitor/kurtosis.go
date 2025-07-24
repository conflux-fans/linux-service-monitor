package monitor

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/conflux-fans/service-monitor/config"
	"github.com/conflux-fans/service-monitor/logger"
	"github.com/nft-rainbow/rainbow-goutils/utils/alertutils"
)

// KurtosisMonitor Kurtosis 监控器
type KurtosisMonitor struct {
	config        *config.KurtosisConfig
	logger        *logger.Logger
	restartCounts map[string]int
}

// ServiceInfo 服务信息结构
type ServiceInfo struct {
	UUID   string
	Name   string
	Status string
}

// NewKurtosisMonitor 创建新的 Kurtosis 监控器
func NewKurtosisMonitor(cfg *config.KurtosisConfig, log *logger.Logger) *KurtosisMonitor {
	return &KurtosisMonitor{
		config:        cfg,
		logger:        log,
		restartCounts: make(map[string]int),
	}
}

// Start 启动 Kurtosis 监控
func (km *KurtosisMonitor) Start() {
	km.logger.Info("启动 Kurtosis 监控服务，监控间隔: %d 秒", km.config.Interval)

	ticker := time.NewTicker(time.Duration(km.config.Interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			km.checkEnclaves()
		}
	}
}

// checkEnclaves 检查所有 enclave 状态
func (km *KurtosisMonitor) checkEnclaves() {
	for _, enclaveName := range km.config.Enclaves {
		km.logger.Info("检查 enclave '%s' 的服务状态", enclaveName)
		services, err := km.getEnclaveServices(enclaveName)
		if err != nil {
			km.logger.Error("获取 enclave '%s' 服务状态失败: %v", enclaveName, err)
			continue
		}

		if len(services) == 0 {
			km.logger.Warn("enclave '%s' 中没有找到服务", enclaveName)
			continue
		}

		// 检查每个服务状态
		stoppedServices := []ServiceInfo{}
		runningCount := 0

		for _, service := range services {
			if service.Status == "STOPPED" {
				stoppedServices = append(stoppedServices, service)
				km.logger.Warn("检测到服务 '%s' (UUID: %s) 已停止", service.Name, service.UUID)

				// 发送钉钉报警
				alertutils.DingWarnf("enclave %s 服务 %s 已停止", enclaveName, service.Name)

			} else if service.Status == "RUNNING" {
				runningCount++
			}
		}

		km.logger.Info("enclave '%s': 总服务数 %d, 运行中 %d, 停止 %d",
			enclaveName, len(services), runningCount, len(stoppedServices))

		// 处理停止的服务
		for _, service := range stoppedServices {
			km.handleStoppedService(enclaveName, service)
		}

		// 如果所有服务都在运行，重置重启计数
		if len(stoppedServices) == 0 && runningCount > 0 {
			for serviceName := range km.restartCounts {
				if strings.Contains(serviceName, enclaveName) {
					km.logger.Info("enclave '%s' 所有服务正常，重置重启计数", enclaveName)
					delete(km.restartCounts, serviceName)
				}
			}
		}
	}
}

// getEnclaveServices 获取 enclave 中的服务列表
func (km *KurtosisMonitor) getEnclaveServices(enclaveName string) ([]ServiceInfo, error) {
	cmd := exec.Command("kurtosis", "enclave", "inspect", enclaveName)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("执行 kurtosis enclave inspect 命令失败: %v", err)
	}

	return km.parseServiceInfo(string(output))
}

// parseServiceInfo 解析服务信息
func (km *KurtosisMonitor) parseServiceInfo(output string) ([]ServiceInfo, error) {
	var services []ServiceInfo
	lines := strings.Split(output, "\n")

	// 查找 "User Services" 部分
	inUserServices := false
	serviceHeaderFound := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 检查是否到达 User Services 部分
		if strings.Contains(line, "User Services") {
			inUserServices = true
			continue
		}

		// 如果在 User Services 部分，查找表头
		if inUserServices && !serviceHeaderFound {
			if strings.Contains(line, "UUID") && strings.Contains(line, "Name") && strings.Contains(line, "Status") {
				serviceHeaderFound = true
				continue
			}
		}

		// 如果找到了表头，开始解析服务信息
		if inUserServices && serviceHeaderFound {
			// 跳过空行
			if line == "" {
				continue
			}

			// 如果遇到新的分割线，说明 User Services 部分结束
			if strings.Contains(line, "=") {
				break
			}

			// 解析服务行
			service := km.parseServiceLine(line)
			if service.UUID != "" && service.Name != "" {
				services = append(services, service)
			}
		}
	}

	return services, nil
}

// parseServiceLine 解析单行服务信息
func (km *KurtosisMonitor) parseServiceLine(line string) ServiceInfo {
	// 使用正则表达式解析服务行
	// 格式: UUID   Name   Ports   Status
	re := regexp.MustCompile(`^([a-f0-9]+)\s+(\S+)\s+.*?\s+(RUNNING|STOPPED|STARTING|STOPPING)`)
	matches := re.FindStringSubmatch(line)

	if len(matches) >= 4 {
		return ServiceInfo{
			UUID:   matches[1],
			Name:   matches[2],
			Status: matches[3],
		}
	}

	return ServiceInfo{}
}

// handleStoppedService 处理停止的服务
func (km *KurtosisMonitor) handleStoppedService(enclaveName string, service ServiceInfo) {
	serviceKey := fmt.Sprintf("%s:%s", enclaveName, service.Name)
	currentCount := km.restartCounts[serviceKey]

	if currentCount >= km.config.MaxRestartAttempts {
		km.logger.Error("服务 '%s' 已连续重启失败 %d 次，停止重启尝试", service.Name, currentCount)

		// 发送钉钉报警
		alertutils.DingWarnf("enclave %s 服务 %s 已连续重启失败 %d 次，停止重启尝试", enclaveName, service.Name, currentCount)
		return
	}

	km.restartCounts[serviceKey]++
	km.logger.Info("尝试重启服务 '%s' (第 %d/%d 次)", service.Name, km.restartCounts[serviceKey], km.config.MaxRestartAttempts)

	// 发送钉钉报警
	alertutils.DingWarnf("enclave %s 服务 %s 第 %d/%d 次重启尝试", enclaveName, service.Name, km.restartCounts[serviceKey], km.config.MaxRestartAttempts)

	if km.restartService(service) {
		km.logger.Info("服务 '%s' 重启成功", service.Name)

		// 发送钉钉报警
		alertutils.DingWarnf("enclave %s 服务 %s 重启成功", enclaveName, service.Name)

		// 重启成功后等待一段时间再检查
		time.Sleep(5 * time.Second)
	} else {
		km.logger.Error("服务 '%s' 重启失败 (第 %d/%d 次)", service.Name, km.restartCounts[serviceKey], km.config.MaxRestartAttempts)

		// 发送钉钉报警
		alertutils.DingWarnf("enclave %s 服务 %s 第 %d/%d 次重启失败", enclaveName, service.Name, km.restartCounts[serviceKey], km.config.MaxRestartAttempts)
	}
}

// restartService 重启服务
func (km *KurtosisMonitor) restartService(service ServiceInfo) bool {
	// 1. 查找对应的 Docker 容器
	containerID, err := km.findDockerContainer(service.Name)
	if err != nil {
		km.logger.Error("查找服务 '%s' 对应的 Docker 容器失败: %v", service.Name, err)
		return false
	}

	if containerID == "" {
		km.logger.Error("未找到服务 '%s' 对应的 Docker 容器", service.Name)
		return false
	}

	km.logger.Info("找到服务 '%s' 对应的容器 ID: %s", service.Name, containerID)

	// 2. 启动容器
	cmd := exec.Command("docker", "start", containerID)
	err = cmd.Run()
	if err != nil {
		km.logger.Error("启动容器 '%s' 失败: %v", containerID, err)
		return false
	}

	km.logger.Info("容器 '%s' 启动命令执行成功", containerID)
	return true
}

// findDockerContainer 查找服务对应的 Docker 容器
func (km *KurtosisMonitor) findDockerContainer(serviceName string) (string, error) {
	// 使用 docker ps -a 查找容器
	cmd := exec.Command("docker", "ps", "-a")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("执行 docker ps -a 命令失败: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		// 查找包含服务名称的行
		if strings.Contains(line, serviceName) {
			// 检查是否是已退出的容器
			if strings.Contains(line, "Exited") || strings.Contains(line, "Created") {
				// 提取容器 ID (第一列)
				fields := strings.Fields(line)
				if len(fields) > 0 {
					containerID := fields[0]
					km.logger.Info("找到已停止的容器: %s (服务: %s)", containerID, serviceName)
					return containerID, nil
				}
			}
		}
	}

	return "", nil
}
