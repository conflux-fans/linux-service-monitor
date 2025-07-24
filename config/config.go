package config

import (
	"github.com/nft-rainbow/rainbow-goutils/utils/configutils"
	"github.com/sirupsen/logrus"
)

// Config 总配置结构
type Config struct {
	Kurtosis KurtosisConfig `yaml:"kurtosis"`
	Process  ProcessConfig  `yaml:"process"`
	Log      LogConfig      `yaml:"log"`
	Alert    Alert          `yaml:"alert"`
}

// KurtosisConfig Kurtosis 监控配置
type KurtosisConfig struct {
	Enclaves           []string `yaml:"enclaves"`
	Interval           int      `yaml:"interval"`
	MaxRestartAttempts int      `yaml:"max_restart_attempts"`
}

// ProcessConfig 进程监控配置
type ProcessConfig struct {
	Processes []string `yaml:"processes"`
	Interval  int      `yaml:"interval"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level string `yaml:"level"`
	File  string `yaml:"file"`
}

type Alert struct {
	CustomTags []string `yaml:"customTags"`
	Channels   map[string]struct {
		Platform  string   `yaml:"platform"`
		Webhook   string   `yaml:"webhook"`
		Secret    string   `yaml:"secret"`
		AtMobiles []string `yaml:"atMobiles"`
		IsAtAll   bool     `yaml:"isAtAll"`
	} `yaml:"channels"`
}

var (
	configVal *Config
)

func Init(configPath string) {
	if configPath == "" {
		configPath = "config.yaml"
	}
	configVal = configutils.MustLoadByFile[Config](configPath)
	logrus.Info("load config done, start to decrypt")
}

func Get() *Config {
	return configVal
}
