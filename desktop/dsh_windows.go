//go:build windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func (a *App) startDsh() error {
	logPath := dshLogPath()
	if mkErr := os.MkdirAll(filepath.Dir(logPath), 0o755); mkErr != nil {
		logPath = os.DevNull
	}
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		logFile = nil
	}

	cmd := exec.Command("cmd", "/C", a.cfg.Command)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: 0x08000000 | 0x00000008,
	}
	if logFile != nil {
		cmd.Stdout = logFile
		cmd.Stderr = logFile
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	if logFile != nil {
		logFile.Close()
	}
	a.mu.Lock()
	a.cmd = cmd
	a.mu.Unlock()
	return nil
}
