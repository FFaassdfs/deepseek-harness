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
	waitTimeout  = 30 * time.Second
	pollInterval = 300 * time.Millisecond
)

type App struct {
	ctx     context.Context
	winCtx  context.Context
	cfg     *DesktopConfig
	cmd     *exec.Cmd
	owns    bool
	mu      sync.Mutex
	booting bool
	booted  bool
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.cfg = loadConfig()
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
		if err := a.startDsh(); err != nil {
			a.fail(ctx, "无法启动 DeepSeek Harness，请确认已安装：npm i -g @deepseek-ai/dsh\n\n"+err.Error())
			return
		}
		a.owns = true
	}
	if !a.waitReady(waitTimeout) {
		a.fail(ctx, "等待 DeepSeek Harness 启动超时（30 秒）\n\n请检查: "+a.cfg.URL())
		return
	}
	a.mu.Lock()
	a.booted = true
	a.mu.Unlock()
	runtime.WindowExecJS(ctx, "window.location.href = '"+a.cfg.URL()+"';")
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

func (a *App) fail(ctx context.Context, msg string) {
	runtime.EventsEmit(ctx, "dsh-error", msg)
}

func (a *App) Retry() {
	if a.winCtx != nil {
		a.startBootstrap(a.winCtx)
	}
}

func (a *App) shutdown(ctx context.Context) {
	if a.owns && a.cmd != nil && a.cmd.Process != nil {
		exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(a.cmd.Process.Pid)).Run()
	}
}
