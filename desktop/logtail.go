package main

import (
	"os"
	"path/filepath"
)

// dshLogPath 返回 harness 运行日志路径。
func dshLogPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = os.TempDir()
	}
	return filepath.Join(dir, "dsh-desktop", "dsh.log")
}

// readLogTail 读取文件末尾最多 maxBytes 字节；文件缺失、为空或读失败返回空串。
func readLogTail(path string, maxBytes int) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return ""
	}
	size := info.Size()
	if size == 0 {
		return ""
	}
	start := size - int64(maxBytes)
	if start < 0 {
		start = 0
	}
	buf := make([]byte, size-start)
	if _, err := f.ReadAt(buf, start); err != nil {
		return ""
	}
	return string(buf)
}
