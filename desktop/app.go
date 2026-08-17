package main

import (
	"context"
	"net"
	"net/http"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	waitTimeout         = 30 * time.Second
	pollInterval        = 300 * time.Millisecond
	healthCheckInterval = 5 * time.Second
)

type App struct {
	ctx          context.Context
	winCtx       context.Context
	cfg          *DesktopConfig
	cmd          *exec.Cmd
	owns         bool
	mu           sync.Mutex
	booting      bool
	booted       bool
	recovering   bool
	shuttingDown bool
	healthStop   chan struct{}
	bridgePort   int
}

func NewApp() *App {
	return &App{healthStop: make(chan struct{})}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.cfg = loadConfig()
	_ = runtime.InitializeNotifications(ctx)
	a.startLinkBridge()
	a.restoreWindowState(ctx)
}

func (a *App) domReady(ctx context.Context) {
	a.winCtx = ctx
	a.startBootstrap(ctx)
}

func (a *App) startBootstrap(ctx context.Context) {
	a.mu.Lock()
	if a.booting || a.booted {
		a.mu.Unlock()
		return
	}
	a.booting = true
	a.mu.Unlock()
	go a.bootstrap(ctx)
}

func (a *App) bootstrap(ctx context.Context) {
	defer func() {
		a.mu.Lock()
		a.booting = false
		a.mu.Unlock()
	}()

	if !a.portOpen() {
		a.maybeAutoUpdateHarness(ctx)
		if err := a.startDsh(); err != nil {
			a.failWithLog(ctx, "无法启动 DeepSeek Harness，请确认已安装：npm i -g @deepseek-ai/dsh\n\n"+err.Error())
			return
		}
		a.mu.Lock()
		a.owns = true
		a.mu.Unlock()
	}
	if !a.waitReady(waitTimeout) {
		a.failWithLog(ctx, "等待 DeepSeek Harness 启动超时（30 秒）\n\n请检查: "+a.cfg.URL())
		return
	}
	a.mu.Lock()
	a.booted = true
	a.mu.Unlock()
	runtime.WindowExecJS(ctx, "window.location.href = '"+a.cfg.URL()+"';")
	a.startHealthMonitor(ctx)
	a.startLinkInterceptor(ctx)
}

func (a *App) portOpen() bool {
	conn, err := net.DialTimeout("tcp", a.cfg.addr(), 800*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func (a *App) waitReady(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 2 * time.Second}
	for time.Now().Before(deadline) {
		resp, err := client.Get(a.cfg.URL())
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return true
			}
		}
		time.Sleep(pollInterval)
	}
	return false
}

// failWithLog 在错误信息后追加 harness 日志尾部，便于诊断启动/崩溃问题。
func (a *App) failWithLog(ctx context.Context, msg string) {
	if tail := readLogTail(dshLogPath(), 2048); tail != "" {
		msg += "\n\n最近日志:\n" + tail
	}
	runtime.EventsEmit(ctx, "dsh-error", msg)
}

// notify 发送原生系统通知（失败静默忽略）。
func notify(ctx context.Context, id, title, body string) {
	_ = runtime.SendNotification(ctx, runtime.NotificationOptions{
		ID:    id,
		Title: title,
		Body:  body,
	})
}

func (a *App) Retry() {
	if a.winCtx != nil {
		a.startBootstrap(a.winCtx)
	}
}

// startHealthMonitor 在启动成功后监测 harness 存活；掉线时自动重启（仅限自拉起的实例）。
func (a *App) startHealthMonitor(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(healthCheckInterval)
		defer ticker.Stop()
		for {
			select {
			case <-a.healthStop:
				return
			case <-ticker.C:
				if !a.portOpen() {
					a.recoverHarness(ctx)
				}
			}
		}
	}()
}

// recoverHarness 检测到端口掉线后，尝试重启自拉起的 harness 并重连。
func (a *App) recoverHarness(ctx context.Context) {
	a.mu.Lock()
	if a.shuttingDown || a.recovering || !a.owns {
		a.mu.Unlock()
		if !a.owns && !a.shuttingDown {
			runtime.EventsEmit(ctx, "dsh-disconnected", "harness 连接已断开")
			notify(ctx, "dsh-disconnected", "DeepSeek Harness", "harness 连接已断开")
		}
		return
	}
	a.recovering = true
	a.mu.Unlock()
	defer func() {
		a.mu.Lock()
		a.recovering = false
		a.mu.Unlock()
	}()

	// 瞬时抖动缓冲，避免把「慢响应」误判成崩溃
	time.Sleep(2 * time.Second)
	if a.portOpen() {
		return
	}
	if err := a.startDsh(); err != nil {
		a.failWithLog(ctx, "harness 已断开，重启失败：\n"+err.Error())
		return
	}
	if !a.waitReady(waitTimeout) {
		a.failWithLog(ctx, "harness 重启超时，请重试")
		return
	}
	runtime.WindowExecJS(ctx, "window.location.href = '"+a.cfg.URL()+"';")
	notify(ctx, "dsh-restarted", "DeepSeek Harness", "harness 已崩溃并自动重启")
}

func (a *App) shutdown(ctx context.Context) {
	close(a.healthStop)
	a.mu.Lock()
	owns, cmd := a.owns, a.cmd
	a.mu.Unlock()
	if owns && cmd != nil && cmd.Process != nil {
		exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(cmd.Process.Pid)).Run()
	}
}
