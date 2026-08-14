//go:build !windows

package main

import "os/exec"

func (a *App) startDsh() error {
	cmd := exec.Command("dsh", "web")
	if err := cmd.Start(); err != nil {
		return err
	}
	a.cmd = cmd
	return nil
}
