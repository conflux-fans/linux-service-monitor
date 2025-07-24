package monitor

import (
	"os/exec"
	"strings"
	"time"

	"github.com/conflux-fans/service-monitor/config"
	"github.com/conflux-fans/service-monitor/logger"
	"github.com/nft-rainbow/rainbow-goutils/utils/alertutils"
)

// ProcessMonitor 进程监控器
type ProcessMonitor struct {
	config        *config.ProcessConfig
	logger        *logger.Logger
	processStates map[string]bool // 记录进程上次的状态
}

// NewProcessMonitor 创建新的进程监控器
func NewProcessMonitor(cfg *config.ProcessConfig, log *logger.Logger) *ProcessMonitor {
	return &ProcessMonitor{
		config:        cfg,
		logger:        log,
		processStates: make(map[string]bool),
	}
}

// Start 启动进程监控
func (pm *ProcessMonitor) Start() {
	pm.logger.Info("启动进程监控服务，监控间隔: %d 秒", pm.config.Interval)

	ticker := time.NewTicker(time.Duration(pm.config.Interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pm.checkProcesses()
		}
	}
}

// checkProcesses 检查所有进程状态
func (pm *ProcessMonitor) checkProcesses() {
	for _, processName := range pm.config.Processes {
		isRunning := pm.isProcessRunning(processName)
		lastState, existed := pm.processStates[processName]

		// 更新进程状态
		pm.processStates[processName] = isRunning

		if !isRunning {
			pm.logger.Error("⚠️ 进程监控告警: 进程 '%s' 已停止运行！", processName)

			// 如果进程从运行状态变为停止状态，或者是第一次检查就发现停止，发送钉钉报警
			if !existed || (existed && lastState) {
				alertutils.DingWarnf("进程 %s 已停止运行", processName)
			}
		} else {
			pm.logger.Info("进程 '%s' 正常运行", processName)

			// 如果进程从停止状态变为运行状态，发送恢复通知
			if existed && !lastState {
				pm.logger.Info("进程 '%s' 状态已恢复", processName)
				alertutils.DingWarnf("进程 %s 状态已恢复", processName)
			}
		}
	}
}

// isProcessRunning 检查进程是否在运行
func (pm *ProcessMonitor) isProcessRunning(processName string) bool {
	// 使用 pgrep 命令查找进程
	cmd := exec.Command("pgrep", "-f", processName)
	output, err := cmd.Output()

	if err != nil {
		// pgrep 返回非零退出码表示没有找到进程
		return false
	}

	// 如果有输出，说明找到了进程
	pids := strings.TrimSpace(string(output))
	if pids != "" {
		pm.logger.Info("进程 '%s' 运行中，PID: %s", processName, strings.ReplaceAll(pids, "\n", ", "))
		return true
	}

	return false
}

// GetProcessInfo 获取进程详细信息
func (pm *ProcessMonitor) GetProcessInfo(processName string) []ProcessInfo {
	var processes []ProcessInfo

	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		pm.logger.Error("执行 ps 命令失败: %v", err)
		return processes
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, processName) && !strings.Contains(line, "ps aux") {
			fields := strings.Fields(line)
			if len(fields) >= 11 {
				processes = append(processes, ProcessInfo{
					User:    fields[0],
					PID:     fields[1],
					CPU:     fields[2],
					Memory:  fields[3],
					Command: strings.Join(fields[10:], " "),
				})
			}
		}
	}

	return processes
}

// ProcessInfo 进程信息结构
type ProcessInfo struct {
	User    string
	PID     string
	CPU     string
	Memory  string
	Command string
}
