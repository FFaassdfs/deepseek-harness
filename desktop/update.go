package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	goruntime "runtime"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const harnessPackage = "@deepseek-ai/dsh"

// UpdateInfo 描述 harness 的更新状态。
type UpdateInfo struct {
	Current         string `json:"current"`
	Latest          string `json:"latest"`
	UpdateAvailable bool   `json:"updateAvailable"`
}

// runShell 以 shell 方式执行命令并返回输出，超时则终止进程。
func runShell(command string, timeout time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var cmd *exec.Cmd
	if goruntime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}
	return cmd.CombinedOutput()
}

// installedHarnessVersion 返回已安装 dsh 的版本（运行 `dsh --version`）。
func installedHarnessVersion() (string, error) {
	out, err := runShell("dsh --version", 10*time.Second)
	if err != nil {
		return "", err
	}
	return normalizeVersion(string(out)), nil
}

// latestHarnessVersion 返回 npm 上最新 dsh 版本（运行 `npm view ... version`）。
func latestHarnessVersion() (string, error) {
	out, err := runShell("npm view "+harnessPackage+" version", 5*time.Second)
	if err != nil {
		return "", err
	}
	return normalizeVersion(string(out)), nil
}

// checkHarnessUpdate 对比已安装版本与最新版本。
func checkHarnessUpdate() (*UpdateInfo, error) {
	current, err := installedHarnessVersion()
	if err != nil {
		return nil, err
	}
	latest, err := latestHarnessVersion()
	if err != nil {
		return nil, err
	}
	return &UpdateInfo{
		Current:         current,
		Latest:          latest,
		UpdateAvailable: versionLess(current, latest),
	}, nil
}

// updateHarness 把全局 dsh 更新到最新版。
func updateHarness() error {
	_, err := runShell("npm install -g "+harnessPackage+"@latest", 120*time.Second)
	return err
}

// maybeAutoUpdateHarness 在拉起新实例前自动更新 harness（如启用且存在新版）。
// 任何错误都只记录日志、不阻塞启动。
func (a *App) maybeAutoUpdateHarness(ctx context.Context) {
	if !a.cfg.AutoUpdateHarness {
		return
	}
	info, err := checkHarnessUpdate()
	if err != nil {
		log.Printf("harness update check failed: %v", err)
		return
	}
	if !info.UpdateAvailable {
		return
	}
	log.Printf("harness update: %s -> %s", info.Current, info.Latest)
	runtime.EventsEmit(ctx, "dsh-updating", info)
	if err := updateHarness(); err != nil {
		log.Printf("harness update failed: %v", err)
		return
	}
	log.Printf("harness updated to %s", info.Latest)
	runtime.EventsEmit(ctx, "dsh-update", info)
}

// normalizeVersion 清理版本串：去空格 / 去 "v" 前缀 / 去命令前缀。
func normalizeVersion(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.LastIndexByte(s, ' '); i >= 0 {
		s = s[i+1:]
	}
	s = strings.TrimPrefix(s, "v")
	return strings.TrimSpace(s)
}

type semverParts struct {
	major, minor, patch int
	pre                 string
}

func parseSemver(v string) semverParts {
	v = normalizeVersion(v)
	core, pre := v, ""
	if i := strings.IndexByte(v, '-'); i >= 0 {
		core, pre = v[:i], v[i+1:]
	}
	var p semverParts
	fmt.Sscanf(core, "%d.%d.%d", &p.major, &p.minor, &p.patch)
	p.pre = pre
	return p
}

// versionLess 比较两个 semver 版本串（支持 "x.y.z-rc.N" 预发布）。
func versionLess(a, b string) bool {
	pa, pb := parseSemver(a), parseSemver(b)
	if pa.major != pb.major {
		return pa.major < pb.major
	}
	if pa.minor != pb.minor {
		return pa.minor < pb.minor
	}
	if pa.patch != pb.patch {
		return pa.patch < pb.patch
	}
	return prereleaseLess(pa.pre, pb.pre)
}

// prereleaseLess 比较预发布标识：无预发布（release）最大；alpha < beta < rc；同前缀比数字。
func prereleaseLess(a, b string) bool {
	if a == b {
		return false
	}
	if a == "" {
		return false
	}
	if b == "" {
		return true
	}
	la, na := splitPrerelease(a)
	lb, nb := splitPrerelease(b)
	if la != lb {
		return la < lb
	}
	return na < nb
}

// splitPrerelease 把预发布串拆成「标签前缀」和「数字」两部分，如 "rc.6" -> ("rc.", 6)。
func splitPrerelease(s string) (string, int) {
	i := strings.IndexFunc(s, func(r rune) bool { return r >= '0' && r <= '9' })
	if i < 0 {
		return s, 0
	}
	var n int
	fmt.Sscanf(s[i:], "%d", &n)
	return s[:i], n
}
