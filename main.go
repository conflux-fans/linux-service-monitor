package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"service_monitor/config"
	"service_monitor/logger"
	"service_monitor/monitor"
	"sync"
	"syscall"

	"github.com/nft-rainbow/rainbow-goutils/utils/alertutils"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置
	config.Init(*configPath)
	cfg := config.Get()

	// 初始化日志记录器
	log, err := logger.NewLogger(cfg.Log.File)
	if err != nil {
		fmt.Printf("初始化日志记录器失败: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	log.Info("=== 服务监控系统启动 ===")
	log.Info("配置文件: %s", *configPath)
	log.Info("日志文件: %s", cfg.Log.File)

	alertutils.MustInitFromViper()
	alertutils.DingInfof("服务监控系统启动")

	// 创建监控器
	kurtosisMonitor := monitor.NewKurtosisMonitor(&cfg.Kurtosis, log)
	processMonitor := monitor.NewProcessMonitor(&cfg.Process, log)

	// 使用 WaitGroup 来管理 goroutine
	var wg sync.WaitGroup

	// 启动 Kurtosis 监控
	if len(cfg.Kurtosis.Enclaves) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Info("启动 Kurtosis 监控，监控的 enclaves: %v", cfg.Kurtosis.Enclaves)
			kurtosisMonitor.Start()
		}()
	} else {
		log.Info("未配置 Kurtosis enclaves，跳过 Kurtosis 监控")
	}

	// 启动进程监控
	if len(cfg.Process.Processes) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Info("启动进程监控，监控的进程: %v", cfg.Process.Processes)
			processMonitor.Start()
		}()
	} else {
		log.Info("未配置监控进程，跳过进程监控")
	}

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 等待退出信号
	log.Info("监控服务已启动，按 Ctrl+C 退出")
	<-sigChan

	log.Info("接收到退出信号，正在关闭监控服务...")
	log.Info("=== 服务监控系统关闭 ===")
}
