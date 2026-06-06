# 安装后使用指南

安装完成后，按照以下步骤开始使用 Source Fetcher。

## 🔄 第一步：刷新环境变量

安装后需要刷新环境变量才能使用 `sfer` 命令。

### 方式一：重启终端（推荐）✨

最简单可靠的方式：

1. 关闭当前的 PowerShell 窗口
2. 重新打开一个新的 PowerShell 窗口
3. `sfer` 命令立即可用

### 方式二：在当前终端刷新

如果不想关闭终端，在当前 PowerShell 中运行：

```powershell
# 方法 A: 使用刷新脚本
. .\refresh-env.ps1

# 方法 B: 手动刷新（复制整行运行）
$env:Path = [System.Environment]::GetEnvironmentVariable('Path','Machine') + ';' + [System.Environment]::GetEnvironmentVariable('Path','User')
```

### 验证是否生效

```powershell
# 查看版本
sfer version

# 如果显示版本号（如 1.0.1），说明安装成功！
```

## 🚀 第二步：快速开始

### 1. 查看帮助

```powershell
sfer --help
```

### 2. 搜索包

```powershell
# 搜索 npm 包
sfer search --source npm --query react

# 搜索 Cargo 包
sfer search --source cargo --query serde

# 搜索所有源
sfer search --source all --query http
```

### 3. 下载包

```powershell
# 下载 npm 包
sfer download --source npm --name react --version latest

# 下载 Cargo 包
sfer download --source cargo --name ripgrep
```

### 4. 安装包

```powershell
# 安装 npm 包
sfer install --source npm --name react --version ^19

# 安装 Cargo 包（下载并解压源码）
sfer install --source cargo --name serde
```

### 5. 启动 GUI

```powershell
# 启动 Web GUI（推荐）
sfer gui

# 浏览器会自动打开 http://localhost:8765
```

## 🎯 常用命令速查

```powershell
# 版本信息
sfer version

# 查看帮助
sfer --help
sfer download --help
sfer install --help

# 镜像测速
sfer mirrors --source npm
sfer mirrors --source all

# 搜索包
sfer search --source npm --query <package>
sfer search --source cargo --query <crate>
sfer search --source winget --query <app>

# 下载包
sfer download --source npm --name <package> --version <version>
sfer download --source cargo --name <crate>

# 安装包
sfer install --source npm --name <package>
sfer install --source cargo --name <crate>

# 批量操作
sfer batch --config source-fetcher.yaml

# 启动界面
sfer gui     # Web GUI
sfer tui     # 终端 TUI
```

## 📦 支持的包源

| 源 | 搜索 | 下载 | 安装 | 说明 |
|---|---|---|---|---|
| **npm** | ✅ | ✅ | ✅ | Node.js 包，完整依赖管理 |
| **pip** | ✅ | ✅ | ❌ | Python 包 |
| **cargo** | ✅ | ✅ | ✅ | Rust 包，源码解压 |
| **maven** | ✅ | ✅ | ❌ | Java 包 |
| **choco** | ✅ | ✅ | ✅ | Chocolatey 包 |
| **winget** | ✅ | ✅ | ✅ | Windows 包 |
| **url** | ❌ | ✅ | ❌ | 直链下载 |

## 🆕 Cargo 安装功能

**特点**：无需安装 Rust 工具链！

```powershell
# 安装最新版本
sfer install --source cargo --name serde

# 安装指定版本
sfer install --source cargo --name tokio --version 1.35.0

# 查看安装计划
sfer install --source cargo --name ripgrep --plan
```

安装后，源码会解压到：
```
./cargo-crates/<name>-<version>/
```

详细说明：[CARGO_INSTALL_GUIDE.md](./CARGO_INSTALL_GUIDE.md)

## ⚙️ 配置文件

创建 `.source-fetcher.yaml` 配置文件：

```powershell
# 复制示例配置
copy .source-fetcher.example.yaml .source-fetcher.yaml

# 编辑配置
notepad .source-fetcher.yaml
```

配置示例：

```yaml
# GUI 配置
gui:
  port: 8765
  auto_open_browser: true

# 下载配置
download:
  output_dir: "./downloads"
  timeout: "30s"
  chunks: 4
  resume: true

# 镜像配置
mirrors:
  npm: "npmmirror"  # 或留空使用默认
```

## 🐛 常见问题

### Q: 提示 "找不到命令 'sfer'"

**A:** 环境变量未生效，尝试：

1. **方法一**：关闭终端，重新打开
2. **方法二**：运行刷新命令
   ```powershell
   $env:Path = [System.Environment]::GetEnvironmentVariable('Path','Machine') + ';' + [System.Environment]::GetEnvironmentVariable('Path','User')
   ```
3. **方法三**：使用完整路径
   ```powershell
   & "$env:LOCALAPPDATA\source-fetcher\source-fetcher.exe" version
   ```

### Q: 如何查看安装位置

```powershell
# 查看安装目录
echo $env:LOCALAPPDATA\source-fetcher

# 打开安装目录
explorer $env:LOCALAPPDATA\source-fetcher

# 查看 PATH
$env:Path -split ';' | Select-String "source-fetcher"
```

### Q: 如何更新

```powershell
# 方法一：重新运行安装脚本
.\install-local.ps1

# 方法二：手动替换
copy source-fetcher.exe $env:LOCALAPPDATA\source-fetcher\
```

### Q: 如何卸载

```powershell
# 运行卸载脚本
& "$env:LOCALAPPDATA\source-fetcher\uninstall.ps1"

# 或手动删除
Remove-Item -Recurse -Force "$env:LOCALAPPDATA\source-fetcher"
```

## 💡 使用技巧

### 1. 使用配置文件简化命令

```yaml
# source-fetcher.yaml
download:
  output_dir: "./my-packages"
  resume: true
  chunks: 4
```

然后：
```powershell
sfer download --source npm --name react --config source-fetcher.yaml
```

### 2. 批量下载

```yaml
# packages.yaml
downloads:
  - source: npm
    name: react
    version: latest
  
  - source: cargo
    name: serde
    version: latest
```

```powershell
sfer batch --config packages.yaml
```

### 3. 使用镜像加速

```powershell
# 自动使用国内镜像
sfer download --source npm --name vue

# 指定镜像
sfer download --source npm --name vue --mirror npmmirror
```

### 4. 交互式搜索

```powershell
# 搜索并交互选择下载
sfer search --source npm --query react --interactive
```

## 📚 更多文档

- [README.md](./README.md) - 项目主页
- [INSTALLATION.md](./INSTALLATION.md) - 完整安装指南
- [CARGO_INSTALL_GUIDE.md](./CARGO_INSTALL_GUIDE.md) - Cargo 安装指南
- [QUICK_START.md](./QUICK_START.md) - 快速开始
- [SETUP_GUIDE.md](./SETUP_GUIDE.md) - 配置指南

## 🎉 开始使用吧！

现在你已经准备好使用 Source Fetcher 了！

```powershell
# 第一个命令
sfer version

# 试试搜索
sfer search --source npm --query lodash

# 启动 GUI
sfer gui
```

需要帮助？运行 `sfer --help` 或查看文档。

---

**享受使用 Source Fetcher！** 🚀
