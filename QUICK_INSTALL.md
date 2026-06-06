# ⚡ Source Fetcher 快速安装

一行命令，3 秒安装！

## Windows

```powershell
irm https://raw.githubusercontent.com/xjwm5685-ui/source-fetcher/main/install.ps1 | iex
```

安装完成后：

```powershell
# 查看版本
sfer version

# 搜索 npm 包
sfer search --source npm --query react

# 启动 Web GUI
sfer gui
```

## 安装内容

✅ 自动下载最新版本  
✅ 安装到 `%LOCALAPPDATA%\source-fetcher`  
✅ 配置 PATH 环境变量  
✅ 创建 `sfer` 全局命令  
✅ 包含卸载脚本  

## 卸载

```powershell
& "$env:LOCALAPPDATA\source-fetcher\uninstall.ps1"
```

## 更多安装方式

查看 [完整安装指南](./INSTALLATION.md) 了解：
- 本地安装脚本
- 手动安装
- 从源码构建
- 常见问题解答

## 立即开始

```powershell
# 搜索包
sfer search --source all --query powertoys

# 下载 npm 包
sfer download --source npm --name react --version latest

# 安装 npm 包
sfer install --source npm --name react --version ^19

# 安装 cargo 包
sfer install --source cargo --name ripgrep

# 启动 Web GUI
sfer gui

# 启动 TUI 界面
sfer tui
```

## 支持的包源

- **npm** - Node.js 包
- **pip** - Python 包  
- **cargo** - Rust 包 ⭐ 新增
- **maven** - Java 包
- **choco** - Chocolatey 包
- **winget** - Windows 包

## 需要帮助？

- 📖 [完整文档](./README.md)
- 🚀 [快速开始](./QUICK_START.md)
- 💾 [安装指南](./INSTALLATION.md)
- ❓ [常见问题](./INSTALLATION.md#常见问题)

---

**遇到问题？** 查看 [安装指南](./INSTALLATION.md) 或提交 [Issue](https://github.com/xjwm5685-ui/source-fetcher/issues)
