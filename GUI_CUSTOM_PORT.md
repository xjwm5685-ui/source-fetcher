# GUI 端口自定义功能 / GUI Custom Port Feature

## 中文说明

### 功能概述
GUI 现在支持自定义端口启动，并提供自动端口冲突检测和回退功能。

### 使用方法

#### 1. 使用默认端口（8765）
```bash
.\source-fetcher.exe gui
```

#### 2. 使用自定义端口
```bash
.\source-fetcher.exe gui --port 8080
```

#### 3. 禁用自动打开浏览器
```bash
.\source-fetcher.exe gui --no-browser
```

#### 4. 组合使用
```bash
.\source-fetcher.exe gui --port 9000 --no-browser
```

### 端口冲突自动处理
如果指定的端口已被占用，系统会自动：
1. 检测到端口冲突
2. 显示警告信息
3. 自动查找下一个可用端口（在指定端口 +100 范围内）
4. 使用找到的可用端口启动服务

**示例输出：**
```
⚠️  Port 8765 is in use, finding available port...
✅ Using port 8766 instead
🌐 Web GUI started at: http://localhost:8766
Press Ctrl+C to stop the server
```

### 命令行参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--port` | int | 8765 | Web 服务器监听端口 |
| `--no-browser` | bool | false | 不自动打开浏览器 |

### 查看帮助
```bash
.\source-fetcher.exe gui --help
```

---

## English Documentation

### Feature Overview
The GUI now supports custom port configuration with automatic port conflict detection and fallback.

### Usage

#### 1. Use Default Port (8765)
```bash
.\source-fetcher.exe gui
```

#### 2. Use Custom Port
```bash
.\source-fetcher.exe gui --port 8080
```

#### 3. Disable Auto-Open Browser
```bash
.\source-fetcher.exe gui --no-browser
```

#### 4. Combined Options
```bash
.\source-fetcher.exe gui --port 9000 --no-browser
```

### Automatic Port Conflict Handling
If the specified port is already in use, the system will automatically:
1. Detect the port conflict
2. Display a warning message
3. Find the next available port (within +100 range)
4. Start the service on the available port

**Example Output:**
```
⚠️  Port 8765 is in use, finding available port...
✅ Using port 8766 instead
🌐 Web GUI started at: http://localhost:8766
Press Ctrl+C to stop the server
```

### Command Line Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `--port` | int | 8765 | Web server listening port |
| `--no-browser` | bool | false | Do not open browser automatically |

### Show Help
```bash
.\source-fetcher.exe gui --help
```

---

## 技术实现 / Technical Implementation

### 端口检测函数 / Port Detection Function

```go
// isPortAvailable checks if a port is available
func isPortAvailable(port int) bool {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}
```

### 端口回退函数 / Port Fallback Function

```go
// findAvailablePort finds an available port starting from startPort
func findAvailablePort(startPort int) (int, error) {
	for port := startPort; port < startPort+100; port++ {
		addr := fmt.Sprintf(":%d", port)
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			listener.Close()
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available port found")
}
```

### 启动逻辑 / Startup Logic

1. 解析命令行参数（端口和浏览器选项）
2. 检查指定端口是否可用
3. 如果不可用，查找下一个可用端口
4. 使用最终确定的端口启动服务器
5. 根据 `--no-browser` 选项决定是否打开浏览器

---

## 常见问题 / FAQ

### Q: 如何知道 GUI 使用的是哪个端口？
**A:** 启动时会在控制台显示：`🌐 Web GUI started at: http://localhost:端口号`

### Q: 端口被占用会发生什么？
**A:** 系统会自动查找下一个可用端口并使用，你会看到警告和新端口的提示。

### Q: 如何在后台运行 GUI？
**A:** 使用 `--no-browser` 参数，并将进程放到后台：
```bash
# Windows (PowerShell)
Start-Process .\source-fetcher.exe -ArgumentList "gui --no-browser" -WindowStyle Hidden

# Linux/macOS
.\source-fetcher.exe gui --no-browser &
```

### Q: How do I know which port the GUI is using?
**A:** The console will display on startup: `🌐 Web GUI started at: http://localhost:port`

### Q: What happens if the port is in use?
**A:** The system automatically finds the next available port and uses it. You'll see a warning and the new port number.

### Q: How do I run the GUI in the background?
**A:** Use the `--no-browser` flag and background the process:
```bash
# Windows (PowerShell)
Start-Process .\source-fetcher.exe -ArgumentList "gui --no-browser" -WindowStyle Hidden

# Linux/macOS
.\source-fetcher.exe gui --no-browser &
```

---

## 测试步骤 / Testing Steps

### 1. 测试默认端口 / Test Default Port
```bash
.\source-fetcher.exe gui
# 应该在 8765 端口启动
```

### 2. 测试自定义端口 / Test Custom Port
```bash
.\source-fetcher.exe gui --port 8080
# 应该在 8080 端口启动
```

### 3. 测试端口冲突 / Test Port Conflict
```bash
# 先启动一个实例
.\source-fetcher.exe gui --port 9000

# 再启动第二个实例（在另一个终端）
.\source-fetcher.exe gui --port 9000
# 应该自动使用 9001 或其他可用端口
```

### 4. 测试无浏览器模式 / Test No-Browser Mode
```bash
.\source-fetcher.exe gui --no-browser
# 不应该自动打开浏览器
```

---

## 更新日志 / Changelog

### v1.1.0 (2026-06-05)
- ✅ 新增：`--port` 参数支持自定义端口
- ✅ 新增：`--no-browser` 参数控制浏览器自动打开
- ✅ 新增：自动端口冲突检测
- ✅ 新增：端口占用时自动回退到可用端口
- ✅ 改进：更友好的启动提示信息

### v1.1.0 (2026-06-05)
- ✅ Added: `--port` flag for custom port configuration
- ✅ Added: `--no-browser` flag to control automatic browser opening
- ✅ Added: Automatic port conflict detection
- ✅ Added: Automatic fallback to available port when port is in use
- ✅ Improved: Better startup messages
