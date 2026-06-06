# Source Fetcher 安装指南

本文档介绍 Source Fetcher 的多种安装方式。

## 📦 安装方式对比

| 方式 | 难度 | 速度 | 推荐场景 |
|------|------|------|----------|
| 一键安装脚本 | ⭐ | ⚡⚡⚡ | 首次安装、快速体验 |
| 本地安装脚本 | ⭐⭐ | ⚡⚡ | 本地开发、离线环境 |
| 手动安装 | ⭐⭐⭐ | ⚡ | 自定义安装位置 |
| 从源码构建 | ⭐⭐⭐⭐ | ⚡ | 开发贡献、定制修改 |

---

## 🚀 方式一：一键安装脚本（推荐）

### 特点
- ✅ 自动下载最新版本
- ✅ 自动配置环境变量
- ✅ 自动创建全局命令 `sfer`
- ✅ 包含卸载脚本

### Windows PowerShell

```powershell
irm https://raw.githubusercontent.com/xjwm5685-ui/source-fetcher/main/install.ps1 | iex
```

### 详细说明

1. **打开 PowerShell**
   - 按 `Win + X`，选择 "Windows PowerShell" 或 "终端"

2. **运行安装命令**
   ```powershell
   irm https://raw.githubusercontent.com/xjwm5685-ui/source-fetcher/main/install.ps1 | iex
   ```

3. **等待安装完成**
   - 脚本会自动检测系统架构
   - 从 GitHub Releases 下载最新版本
   - 安装到 `%LOCALAPPDATA%\source-fetcher`
   - 配置 PATH 环境变量
   - 创建 `sfer` 命令别名

4. **重启终端**
   ```powershell
   # 关闭当前终端，重新打开
   ```

5. **验证安装**
   ```powershell
   sfer version
   ```

### 安装位置

- **程序文件**: `%LOCALAPPDATA%\source-fetcher\source-fetcher.exe`
- **命令别名**: `%LOCALAPPDATA%\source-fetcher\sfer.bat`
- **卸载脚本**: `%LOCALAPPDATA%\source-fetcher\uninstall.ps1`

### 卸载

```powershell
# 运行卸载脚本
& "$env:LOCALAPPDATA\source-fetcher\uninstall.ps1"

# 或者手动删除
Remove-Item -Recurse -Force "$env:LOCALAPPDATA\source-fetcher"
# 手动从 PATH 中移除（可选）
```

---

## 💻 方式二：本地安装脚本

适用于已下载或编译好的可执行文件。

### 前提条件

确保已有 `source-fetcher.exe` 文件（通过下载或编译）。

### 步骤

1. **编译或下载可执行文件**
   ```powershell
   # 如果从源码编译
   go build -o source-fetcher.exe
   
   # 或从 Releases 下载
   ```

2. **运行本地安装脚本**
   ```powershell
   .\install-local.ps1
   
   # 或指定源文件路径
   .\install-local.ps1 -SourcePath ".\path\to\source-fetcher.exe"
   ```

3. **重启终端并验证**
   ```powershell
   sfer version
   ```

---

## 📋 方式三：手动安装

完全手动控制安装过程。

### 步骤

1. **下载可执行文件**
   - 访问 [Releases 页面](https://github.com/xjwm5685-ui/source-fetcher/releases)
   - 下载对应系统架构的版本：
     - `source-fetcher-windows-amd64.exe` (64位)
     - `source-fetcher-windows-386.exe` (32位)

2. **选择安装目录**
   ```powershell
   # 推荐位置
   mkdir "$env:LOCALAPPDATA\source-fetcher"
   
   # 或自定义位置
   mkdir "C:\Tools\source-fetcher"
   ```

3. **复制文件**
   ```powershell
   # 重命名并复制
   Copy-Item ".\source-fetcher-windows-amd64.exe" "$env:LOCALAPPDATA\source-fetcher\source-fetcher.exe"
   ```

4. **添加到 PATH**
   
   **方法 A: PowerShell**
   ```powershell
   $installDir = "$env:LOCALAPPDATA\source-fetcher"
   $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
   $newPath = "$userPath;$installDir"
   [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
   ```
   
   **方法 B: 图形界面**
   - 按 `Win + R`，输入 `sysdm.cpl` 回车
   - 选择"高级"标签 → "环境变量"
   - 在"用户变量"中找到 `Path`，点击"编辑"
   - 点击"新建"，添加安装目录路径
   - 点击"确定"保存

5. **创建命令别名（可选）**
   ```powershell
   $installDir = "$env:LOCALAPPDATA\source-fetcher"
   $batchContent = @"
@echo off
"$installDir\source-fetcher.exe" %*
"@
   Set-Content -Path "$installDir\sfer.bat" -Value $batchContent
   ```

6. **重启终端并验证**
   ```powershell
   source-fetcher version
   # 或使用别名
   sfer version
   ```

---

## 🛠️ 方式四：从源码构建

适合开发者和贡献者。

### 前提条件

- **Go 1.25+**: 从 [golang.org](https://golang.org/dl/) 下载
- **Git**: 从 [git-scm.com](https://git-scm.com/) 下载

### 步骤

1. **克隆仓库**
   ```powershell
   git clone https://github.com/xjwm5685-ui/source-fetcher.git
   cd source-fetcher
   ```

2. **查看依赖**
   ```powershell
   go mod download
   ```

3. **编译**
   ```powershell
   # 编译当前平台
   go build -o source-fetcher.exe
   
   # 编译带版本信息
   go build -ldflags "-X main.version=1.0.1" -o source-fetcher.exe
   
   # 编译压缩版本
   go build -ldflags "-s -w" -o source-fetcher.exe
   ```

4. **运行**
   ```powershell
   .\source-fetcher.exe version
   ```

5. **安装（可选）**
   ```powershell
   # 使用本地安装脚本
   .\install-local.ps1
   
   # 或使用别名脚本
   .\install-alias.ps1
   ```

### 交叉编译

```powershell
# Windows 64位
$env:GOOS="windows"; $env:GOARCH="amd64"
go build -o source-fetcher-windows-amd64.exe

# Windows 32位
$env:GOOS="windows"; $env:GOARCH="386"
go build -o source-fetcher-windows-386.exe

# Linux 64位
$env:GOOS="linux"; $env:GOARCH="amd64"
go build -o source-fetcher-linux-amd64

# macOS 64位
$env:GOOS="darwin"; $env:GOARCH="amd64"
go build -o source-fetcher-darwin-amd64
```

---

## 🔧 常见问题

### Q: 安装后提示 "找不到命令"

**A:** 可能的原因：

1. **环境变量未生效**
   - 关闭所有终端窗口
   - 重新打开新的终端窗口

2. **PATH 未正确配置**
   ```powershell
   # 检查 PATH
   $env:Path -split ';' | Select-String "source-fetcher"
   
   # 手动刷新环境变量
   $env:Path = [System.Environment]::GetEnvironmentVariable("Path", "Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path", "User")
   ```

3. **使用完整路径**
   ```powershell
   & "$env:LOCALAPPDATA\source-fetcher\source-fetcher.exe" version
   ```

### Q: 一键安装脚本失败

**A:** 可能的原因：

1. **网络问题**
   - 检查是否能访问 GitHub
   - 尝试使用代理或 VPN

2. **PowerShell 执行策略**
   ```powershell
   # 检查当前策略
   Get-ExecutionPolicy
   
   # 临时允许执行
   Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass
   ```

3. **手动下载安装**
   - 使用方式三：手动安装

### Q: 提示需要管理员权限

**A:** 一键安装脚本会安装到用户目录，通常不需要管理员权限。如果提示需要，可以：

```powershell
# 以管理员身份运行 PowerShell
Start-Process powershell -Verb runAs

# 然后运行安装命令
irm https://raw.githubusercontent.com/xjwm5685-ui/source-fetcher/main/install.ps1 | iex
```

### Q: 如何更新到最新版本

**A:** 重新运行一键安装脚本即可：

```powershell
irm https://raw.githubusercontent.com/xjwm5685-ui/source-fetcher/main/install.ps1 | iex
```

脚本会自动覆盖旧版本。

### Q: 如何卸载

**A:** 三种方式：

```powershell
# 方式一：运行卸载脚本
& "$env:LOCALAPPDATA\source-fetcher\uninstall.ps1"

# 方式二：手动删除
Remove-Item -Recurse -Force "$env:LOCALAPPDATA\source-fetcher"

# 方式三：使用别名卸载脚本
.\install-alias.ps1 -Uninstall
```

### Q: 可以安装到自定义位置吗

**A:** 可以，使用手动安装方式（方式三），自行选择安装目录。

---

## 📚 相关文档

- [README](./README.md) - 项目主页
- [QUICK_START](./QUICK_START.md) - 快速开始
- [SETUP_GUIDE](./SETUP_GUIDE.md) - 配置指南
- [CHANGELOG](./CHANGELOG.md) - 更新日志

---

## 💡 提示

- 首次使用推荐使用一键安装脚本
- 开发者建议从源码构建
- 企业环境可以使用手动安装配合内网镜像
- 定期运行一键安装脚本以获取最新版本
