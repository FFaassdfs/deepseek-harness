package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	defaultPort    = 3080
	defaultCommand = "dsh web"
)

// DesktopConfig 是桌面壳的运行配置，位于 %APPDATA%\dsh-desktop\config.json。
// 缺失或字段为空时回落到默认值，保证桌面壳无需配置即可运行。
type DesktopConfig struct {
	// Port 是 harness Web UI 的监听端口，默认 3080。
	Port int `json:"port"`
	// Command 是拉起 harness Web UI 的命令，默认 "dsh web"；支持带空格的 shell 命令（如 "pnpm dsh web"）。
	Command string `json:"command"`
	// AutoUpdateHarness 是否在拉起新实例前自动把全局 dsh 更新到最新版，默认 true。
	AutoUpdateHarness bool `json:"autoUpdateHarness"`
	// WorkDir 是拉起 harness 进程时的工作目录（用于定位 .env / cordis 配置），默认空=继承桌面壳自身目录。
	WorkDir string `json:"workdir"`
}

func configPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = os.TempDir()
	}
	return filepath.Join(dir, "dsh-desktop", "config.json")
}

func defaultConfig() *DesktopConfig {
	return &DesktopConfig{Port: defaultPort, Command: defaultCommand, AutoUpdateHarness: true}
}

// loadConfigFrom 从指定路径加载配置；文件缺失时返回默认值。
// 用指针区分「字段缺失」（回落默认）与「显式设值」（含显式 false）。
func loadConfigFrom(path string) (*DesktopConfig, error) {
	cfg := defaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}
	var raw struct {
		Port              *int    `json:"port"`
		Command           *string `json:"command"`
		AutoUpdateHarness *bool   `json:"autoUpdateHarness"`
		WorkDir           *string `json:"workdir"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	if raw.Port != nil && *raw.Port > 0 {
		cfg.Port = *raw.Port
	}
	if raw.Command != nil && *raw.Command != "" {
		cfg.Command = *raw.Command
	}
	if raw.AutoUpdateHarness != nil {
		cfg.AutoUpdateHarness = *raw.AutoUpdateHarness
	}
	if raw.WorkDir != nil {
		cfg.WorkDir = *raw.WorkDir
	}
	return cfg, nil
}

// loadConfig 加载用户配置，任何错误都回落到默认值（永不返回 nil）。
func loadConfig() *DesktopConfig {
	cfg, err := loadConfigFrom(configPath())
	if err != nil {
		return defaultConfig()
	}
	return cfg
}

// addr 返回 TCP 监听地址。
func (c *DesktopConfig) addr() string {
	return fmt.Sprintf("127.0.0.1:%d", c.Port)
}

// URL 返回 Web UI 的访问地址。
func (c *DesktopConfig) URL() string {
	return fmt.Sprintf("http://%s", c.addr())
}
