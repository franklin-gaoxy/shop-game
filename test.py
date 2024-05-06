import shutil

# 获取当前窗口宽度
window_width = shutil.get_terminal_size().columns

# 输出 "-"
print("-" * window_width)
