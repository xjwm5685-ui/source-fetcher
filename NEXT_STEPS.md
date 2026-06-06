# 下一步操作指南

## 🎯 立即执行的任务

### 1. 发布到 GitHub Release ⭐ 最重要

你的一键安装脚本需要从 GitHub Releases 下载二进制文件，所以这是**最优先**的任务。

#### 步骤 A: 编译可执行文件

```powershell
# 确保在项目根目录
cd d:\dy\source-fetcher

# 创建 dist 目录
mkdir -Force dist

# 编译 Windows x64 版本
go build -ldflags="-s -w" -o dist/source-fetcher-windows-amd64.exe

# 编译 Windows x86 版本
$env:GOARCH="386"
go build -ldflags="-s -w" -o dist/source-fetcher-windows-386.exe
$env:GOARCH="amd64"  # 恢复
```

#### 步骤 B: 推送代码到 GitHub

```powershell
# 检查当前状态
git status

# 添加所有新文件
git add .

# 提交
git commit -m "feat: add cargo install support and one-line installation

- Add cargo package installation without Rust toolchain
- Add one-line PowerShell installation script
- Add Web GUI support for cargo/choco/winget installation
- Update documentation with comprehensive guides

Version: 1.1.0"

# 推送到 main 分支
git push origin main
```

#### 步骤 C: 创建 GitHub Release

1. **访问 GitHub 仓库**
   ```
   https://github.com/YOUR_USERNAME/source-fetcher/releases/new
   ```

2. **创建新 Release**
   - Tag: `v1.1.0`
   - Release title: `v1.1.0 - Cargo Install Support & One-Line Installation`
   - Description:

```markdown
## 🎉 主要更新

### 新功能

#### ⭐ Cargo 包安装（无需 Rust 工具链）
- 直接下载并解压 .crate 源码包
- 支持最新版本和指定版本
- 集成到 CLI 和 Web GUI

#### ⭐ 一键安装脚本
```powershell
irm https://raw.githubusercontent.com/YOUR_USERNAME/source-fetcher/main/install.ps1 | iex
```

#### ⭐ Web GUI 多源支持
- npm, cargo, choco, winget 全部支持安装
- 智能包源检查和错误提示
- 批量操作和实时队列更新

### 改进
- 📝 完善的文档系统
- 🔒 安全性增强
- 🎨 更好的用户体验

### 安装方式

**一键安装（推荐）**:
```powershell
irm https://raw.githubusercontent.com/YOUR_USERNAME/source-fetcher/main/install.ps1 | iex
```

**手动下载**:
下载下方的可执行文件，解压后运行 `install-local.ps1`

### 完整更新日志
见 [CHANGELOG.md](https://github.com/YOUR_USERNAME/source-fetcher/blob/main/CHANGELOG.md)
```

3. **上传文件**
   - 上传 `dist/source-fetcher-windows-amd64.exe`
   - 上传 `dist/source-fetcher-windows-386.exe`

4. **发布 Release**

#### 步骤 D: 测试一键安装

```powershell
# 在新的 PowerShell 窗口测试
irm https://raw.githubusercontent.com/YOUR_USERNAME/source-fetcher/main/install.ps1 | iex

# 验证安装
sfer version
sfer search --source npm --query react
sfer gui
```

### 2. 更新 CHANGELOG.md

将 `[Unreleased]` 部分的内容移到新版本：

```powershell
# 编辑 CHANGELOG.md
```

将以下内容：

```markdown
## [Unreleased]

### Added
- **Cargo install support**: Download and extract .crate source packages without Rust toolchain
- One-line installation script (`install.ps1`)
- Web GUI support for cargo/choco/winget installation
...
```

改为：

```markdown
## [Unreleased]

<!-- 未来的更新 -->

## [1.1.0] - 2026-06-06

### Added
- **⭐ Cargo install support**: Download and extract .crate source packages without Rust toolchain
- **⭐ One-line installation script** (`install.ps1`) for quick setup
- **⭐ Web GUI support** for cargo/choco/winget package installation
- Comprehensive documentation (INSTALLATION.md, QUICK_INSTALL.md, etc.)
- Local installation script (`install-local.ps1`) for offline scenarios
- Environment variable refresh script (`refresh-env.ps1`)

### Improved
- Web GUI now supports multiple package sources (npm, cargo, choco, winget)
- Better error messages in Web GUI queue display
- Smart package source detection and filtering
- Enhanced installation guides

### Documentation
- Added CARGO_INSTALL_GUIDE.md - Cargo installation user guide
- Added CARGO_FEATURE_SUMMARY.md - Technical implementation summary
- Added WEBGUI_CARGO_SUPPORT.md - Web GUI update documentation
- Added INSTALLATION.md - Complete installation guide
- Added QUICK_INSTALL.md - Quick start guide
- Added ONE_LINE_INSTALL_SUMMARY.md - Implementation summary
- Added POST_INSTALL_GUIDE.md - Post-installation guide
- Added PROJECT_STATUS.md - Project status report
- Added NEXT_STEPS.md - Next steps guide

## [1.0.1] - 2026-06-06
...
```

### 3. 更新 README.md 中的链接

确保 README 中的链接指向正确的 GitHub 仓库：

```powershell
# 搜索并替换（使用编辑器）
# 查找: YOUR_USERNAME
# 替换为: 你的 GitHub 用户名

# 查找: xjwm5685-ui
# 如果不是你的用户名，替换为正确的用户名
```

## 📋 完成后的验证清单

### 发布验证

- [ ] GitHub Release 已创建
- [ ] 二进制文件已上传
- [ ] Release 描述清晰完整
- [ ] Tag 版本号正确 (v1.1.0)

### 安装验证

- [ ] 一键安装脚本可以运行
- [ ] 从 GitHub 下载成功
- [ ] `sfer` 命令可用
- [ ] 环境变量配置正确

### 功能验证

```powershell
# CLI 测试
sfer version
sfer search --source npm --query react
sfer download --source cargo --name serde
sfer install --source cargo --name tokio --plan

# GUI 测试
sfer gui
# 在浏览器中测试搜索、下载、安装功能
```

### 文档验证

- [ ] CHANGELOG.md 更新正确
- [ ] README.md 链接有效
- [ ] 所有文档无明显错误

## 🚀 可选的后续任务

### 短期改进

#### 1. 添加自动更新功能

创建 `update.go` 文件实现 `sfer update` 命令：

```go
// update.go
package main

import (
    "context"
    "fmt"
)

func runUpdate(args []string) error {
    // 1. 获取最新版本
    // 2. 比较当前版本
    // 3. 下载新版本
    // 4. 替换可执行文件
    // 5. 显示更新日志
    
    fmt.Println("Checking for updates...")
    fmt.Println("You are using the latest version: 1.1.0")
    return nil
}
```

#### 2. 改进 Web GUI

- 添加包详情页面
- 添加下载历史记录
- 添加设置页面
- 支持深色/浅色主题切换

#### 3. 性能优化

```powershell
# 添加性能测试
go test -bench=. -benchmem

# 分析 CPU 使用
go test -cpuprofile=cpu.prof
go tool pprof cpu.prof

# 分析内存使用
go test -memprofile=mem.prof
go tool pprof mem.prof
```

### 中期目标

#### 4. Linux/macOS 支持

创建 `install.sh`:

```bash
#!/bin/bash
# Source Fetcher Installation Script for Linux/macOS

REPO_OWNER="YOUR_USERNAME"
REPO_NAME="source-fetcher"
APP_NAME="source-fetcher"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Download and install
# ...
```

#### 5. 创建 GitHub Actions CI/CD

创建 `.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Build
        run: |
          go build -ldflags="-s -w" -o dist/source-fetcher-windows-amd64.exe
          $env:GOARCH="386"
          go build -ldflags="-s -w" -o dist/source-fetcher-windows-386.exe
      
      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: dist/*
```

## 💡 使用技巧

### 快速开发测试

```powershell
# 监控文件变化并自动编译（需要安装 watchexec）
watchexec -e go -r go build

# 或使用 air 进行热重载
go install github.com/cosmtrek/air@latest
air

# 快速测试安装
.\install-local.ps1
& "$env:LOCALAPPDATA\source-fetcher\uninstall.ps1"
.\install-local.ps1
```

### 调试技巧

```powershell
# 启用详细输出
$env:SFER_DEBUG="1"
sfer gui

# 查看 HTTP 请求
$env:SFER_HTTP_DEBUG="1"
sfer search --source npm --query react

# 查看 Web GUI 日志
sfer gui --verbose
```

## 📚 参考资源

### 学习资源

- [Go 官方文档](https://go.dev/doc/)
- [PowerShell 文档](https://docs.microsoft.com/powershell/)
- [GitHub Actions 文档](https://docs.github.com/actions)

### 类似项目

- [rustup](https://github.com/rust-lang/rustup) - Rust 工具链安装器
- [nvm](https://github.com/nvm-sh/nvm) - Node.js 版本管理
- [scoop](https://github.com/ScoopInstaller/Scoop) - Windows 包管理器

## 🎓 最佳实践

### Git 提交规范

```
feat: 新功能
fix: 修复 bug
docs: 文档更新
style: 代码格式
refactor: 重构
test: 测试
chore: 构建/工具

示例:
feat: add cargo install support
fix: resolve CORS security issue
docs: update installation guide
```

### 版本发布流程

1. 更新版本号
2. 更新 CHANGELOG
3. 提交代码
4. 创建 Git Tag
5. 推送 Tag
6. GitHub 自动构建 Release
7. 测试发布版本
8. 公告新版本

## ✅ 快速检查清单

使用这个清单确保一切就绪：

```markdown
## 发布前检查

- [ ] 代码已提交到 GitHub
- [ ] 版本号已更新
- [ ] CHANGELOG 已更新
- [ ] 文档已审核
- [ ] 编译通过
- [ ] 测试通过
- [ ] GitHub Release 已创建
- [ ] 二进制文件已上传
- [ ] 一键安装脚本可用
- [ ] 安装验证通过
- [ ] 功能验证通过

## 发布后

- [ ] 在 GitHub 添加 Release 公告
- [ ] 更新项目网站（如有）
- [ ] 社交媒体发布（可选）
- [ ] 收集用户反馈
- [ ] 规划下一个版本
```

## 🆘 遇到问题？

### 常见问题

**Q: 一键安装显示 404 错误**
A: 确保 GitHub Release 已创建，并且文件名匹配脚本中的命名

**Q: sfer 命令不可用**
A: 重启终端，或运行 `.\refresh-env.ps1`

**Q: Web GUI 无法启动**
A: 检查端口是否被占用，尝试 `sfer gui --port 8888`

**Q: cargo 安装失败**
A: 检查网络连接，确保可以访问 crates.io

### 获取帮助

- 查看文档: `README.md`, `INSTALLATION.md`
- 查看日志: 启用 debug 模式
- 提交 Issue: GitHub Issues
- 参与讨论: GitHub Discussions

## 🎉 完成后

当你完成所有步骤后，你将拥有：

✅ 一个功能完整的包管理工具
✅ 优秀的一键安装体验
✅ 完善的文档系统
✅ 专业的 GitHub 仓库
✅ 活跃的开源项目

**恭喜！继续加油！** 🚀

---

**提示**: 按照这个指南，你应该能在 1-2 小时内完成 GitHub 发布流程。如果遇到问题，随时查看相关文档或寻求帮助。
