//go:build windows

package main

import "syscall"

var kernel32 = syscall.NewLazyDLL("kernel32.dll")

// Windows 下尽力将控制台输出切换为 UTF-8 代码页，避免中文乱码（失败不影响运行）
func init() {
	if p := kernel32.NewProc("SetConsoleOutputCP"); p.Find() == nil {
		_, _, _ = p.Call(65001)
	}
}
