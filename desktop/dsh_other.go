//go:build !windows

package main

import "os/exec"

func (a *App) startDsh() error {
	cmd := exec.Command("sh", "-c", a.cfg.Command)
	if a.cfg.WorkDir != "" {
		cmd.Dir = a.cfg.WorkDir
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	a.mu.Lock()
	a.cmd = cmd
	a.mu.Unlock()
	return nil
}
