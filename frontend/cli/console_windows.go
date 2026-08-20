//go:build windows

package main

import (
	"os"
	"syscall"
	"unsafe"
)

// init 在 Windows 下将控制台代码页切换为 UTF-8（65001），
// 保证表格边框与中文输出不乱码。
func init() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setConsoleOutputCP := kernel32.NewProc("SetConsoleOutputCP")
	setConsoleOutputCP.Call(uintptr(65001))
	// 启用虚拟终端处理，支持 ANSI 转义序列（Win10 1511+）
	handle := syscall.Handle(os.Stdout.Fd())
	var mode uint32
	getConsoleMode := kernel32.NewProc("GetConsoleMode")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")
	if r, _, _ := getConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode))); r != 0 {
		setConsoleMode.Call(uintptr(handle), uintptr(mode|0x0004)) // ENABLE_VIRTUAL_TERMINAL_PROCESSING
	}
}
