package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
)

var (
	stdin       = bufio.NewReader(os.Stdin)
	stdinClosed = false
)

// readLine 读取一行输入并去除首尾空白；输入流结束时退出程序
func readLine(prompt string) string {
	if stdinClosed {
		os.Exit(0)
	}
	fmt.Print(prompt)
	line, err := stdin.ReadString('\n')
	if err != nil && strings.TrimSpace(line) == "" {
		stdinClosed = true
		fmt.Println()
		os.Exit(0)
	}
	return strings.TrimSpace(line)
}

// readPassword 读取密码（不回显；非终端环境下回退为普通输入）
func readPassword(prompt string) string {
	fmt.Print(prompt)
	if b, err := term.ReadPassword(int(os.Stdin.Fd())); err == nil {
		fmt.Println()
		return strings.TrimSpace(string(b))
	}
	return readLine("")
}

// readChoice 读取菜单序号（min-max 之间，含边界）
func readChoice(prompt string, min, max int) int {
	for {
		s := readLine(prompt)
		n, err := strconv.Atoi(s)
		if err != nil || n < min || n > max {
			fmt.Printf("  输入无效，请输入 %d-%d 的序号\n", min, max)
			continue
		}
		return n
	}
}

// readIntRange 读取整数（min<=v<=max，max<=0 表示不限制上界）
func readIntRange(prompt string, min, max int64) int64 {
	for {
		s := readLine(prompt)
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil || n < min || (max > 0 && n > max) {
			if max > 0 {
				fmt.Printf("  输入无效，请输入 %d-%d 之间的整数\n", min, max)
			} else {
				fmt.Printf("  输入无效，请输入 >= %d 的整数\n", min)
			}
			continue
		}
		return n
	}
}

// readStorageOptional 读取仓库类型（1=普通 2=冷藏，回车=全部），用于卖出
func readStorageOptional() int {
	for {
		s := readLine("存储仓库（1=普通 2=冷藏，直接回车=全部仓库）: ")
		switch s {
		case "":
			return 0
		case "1":
			return StorageNormal
		case "2":
			return StorageCold
		default:
			fmt.Println("  输入无效，请输入 1、2 或直接回车")
		}
	}
}

// confirm 确认操作（y/yes 确认）
func confirm(prompt string) bool {
	s := strings.ToLower(readLine(prompt))
	return s == "y" || s == "yes"
}

// ---------- Linux 风格表格绘制（支持中文对齐） ----------

// runeWidth 返回字符在终端中的显示宽度（中日韩等全角字符占 2 列）
func runeWidth(r rune) int {
	switch {
	case r >= 0x1100 && r <= 0x115F, // 朝鲜文
		r >= 0x2E80 && r <= 0x303E, // CJK 部首、符号
		r >= 0x3041 && r <= 0x33FF, // 假名、注音、CJK 补充
		r >= 0x3400 && r <= 0x4DBF, // CJK 扩展 A
		r >= 0x4E00 && r <= 0x9FFF, // CJK 基本
		r >= 0xA000 && r <= 0xA4CF, // 彝文
		r >= 0xAC00 && r <= 0xD7A3, // 朝鲜文音节
		r >= 0xF900 && r <= 0xFAFF, // CJK 兼容
		r >= 0xFE30 && r <= 0xFE4F, // CJK 兼容形式
		r >= 0xFF00 && r <= 0xFF60, // 全角形式
		r >= 0xFFE0 && r <= 0xFFE6: // 全角符号
		return 2
	}
	return 1
}

// displayWidth 计算字符串的终端显示宽度
func displayWidth(s string) int {
	w := 0
	for _, r := range s {
		w += runeWidth(r)
	}
	return w
}

// Table Linux 风格表格（+---+ 分隔线）
type Table struct {
	headers []string
	rows    [][]string
}

func NewTable(headers ...string) *Table {
	return &Table{headers: headers}
}

func (t *Table) AddRow(cells ...string) *Table {
	t.rows = append(t.rows, cells)
	return t
}

// Print 输出表格到终端
func (t *Table) Print() {
	widths := make([]int, len(t.headers))
	for i, h := range t.headers {
		widths[i] = displayWidth(h)
	}
	for _, row := range t.rows {
		for i, c := range row {
			if w := displayWidth(c); w > widths[i] {
				widths[i] = w
			}
		}
	}
	sep := t.separator(widths)
	fmt.Println(sep)
	fmt.Println(t.formatRow(t.headers, widths))
	fmt.Println(sep)
	for _, row := range t.rows {
		fmt.Println(t.formatRow(row, widths))
	}
	fmt.Println(sep)
}

func (t *Table) separator(widths []int) string {
	var b strings.Builder
	b.WriteString("+")
	for _, w := range widths {
		b.WriteString(strings.Repeat("-", w+2))
		b.WriteString("+")
	}
	return b.String()
}

func (t *Table) formatRow(cells []string, widths []int) string {
	var b strings.Builder
	b.WriteString("|")
	for i, c := range cells {
		pad := widths[i] - displayWidth(c)
		if pad < 0 {
			pad = 0
		}
		b.WriteString(" ")
		b.WriteString(c)
		b.WriteString(strings.Repeat(" ", pad))
		b.WriteString(" |")
	}
	return b.String()
}
