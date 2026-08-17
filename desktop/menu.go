package main

import (
	"os/exec"
	goruntime "runtime"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// openPath 用系统默认程序打开文件。
func openPath(path string) error {
	var cmd *exec.Cmd
	if goruntime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", "start", "", path)
	} else {
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start()
}

// buildMenu 构建原生菜单。
func (a *App) buildMenu() *menu.Menu {
	root := menu.NewMenu()
	app := root.AddSubmenu("应用")
	app.AddText("重新加载", keys.CmdOrCtrl("r"), func(_ *menu.CallbackData) {
		if a.winCtx != nil {
			runtime.WindowReload(a.winCtx)
		}
	})
	app.AddSeparator()
	app.AddText("打开日志", nil, func(_ *menu.CallbackData) {
		_ = openPath(dshLogPath())
	})
	app.AddText("打开配置", nil, func(_ *menu.CallbackData) {
		if ensureConfigFile() == nil {
			_ = openPath(configPath())
		}
	})
	app.AddSeparator()
	app.AddText("退出", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
		if a.ctx != nil {
			runtime.Quit(a.ctx)
		}
	})
	return root
}
