# GitHub 推送和发布指南

本指南将帮助你将项目推送到 GitHub 并启用一键安装功能。

## 📋 前置检查

在推送之前，确保：

- [x] 已创建 cargo 安装功能
- [x] 已创建一键安装脚本
- [x] 本地测试通过
- [ ] 准备推送到 GitHub
- [ ] 创建 Release

## 🚀 步骤一：初始化 Git 仓库（如果尚未初始化）

```powershell
cd d:\dy\source-fetcher

# 初始化 Git（如果还没有）
git init

# 添加远程仓库
git remote add origin https://github.com/xjwm5685-ui/source-fetcher.git

# 或者如果已存在远程仓库
git remote set-url origin https://github.com/xjwm5685-ui/source-fetcher.git
```

## 📝 步骤二：提交新文件

```powershell
# 查看状态
git status

# 添加新创建的文件
git add install.ps1
git add install-local.ps1
git add test-install.ps1
git add INSTALLATION.md
git add QUICK_INSTALL.md
git add ONE_LINE_INSTALL_SUMMARY.md
git add CARGO_INSTALL_GUIDE.md
git add CARGO_FEATURE_SUMMARY.md
git add GITHUB_PUSH_GUIDE.md

# 添加修改的文件
git add README.md
git add CHANGELOG.md
git add install_native.go
git add main.go
git add install.go

# 查看将要提交的文件
git status

# 提交
git commit -m "feat: 添加 Cargo 安装功能和一键安装脚本

- 新增 Cargo 包下载和源码解压安装功能
- 新增一键安装脚本 (install.ps1)
- 新增本地安装脚本 (install-local.ps1)
- 新增完整安装文档 (INSTALLATION.md)
- 更新 README 添加一键安装说明
- 优化用户安装体验"
```

## 🌐 步骤三：推送到 GitHub

```powershell
# 推送到主分支
git push origin main

# 如果是首次推送，可能需要
git push -u origin main

# 如果遇到冲突，先拉取
git pull origin main --rebase
git push origin main
```

## 📦 步骤四：创建 Release

### 方式一：通过 GitHub 网页

1. 访问 https://github.com/xjwm5685-ui/source-fetcher/releases
2. 点击 "Create a new release"
3. 填写信息：
   - **Tag version**: `v1.0.2` 或 `v1.1.0`
   - **Release title**: `v1.0.2 - Cargo 安装支持`
   - **Description**: 

```markdown
## ✨ 新功能

### Cargo 包安装支持
- 无需 Rust 工具链即可下载和解压 Cargo crate 源码包
- 支持查看和学习 Rust 项目源码
- 适合离线环境源码分发

### 一键安装体验
- 新增一键安装脚本，3 秒快速安装
- 自动配置环境变量和全局命令
- 类似 Kimi Code 的安装体验

## 📥 快速安装

```powershell
irm https://raw.githubusercontent.com/xjwm5685-ui/source-fetcher/main/install.ps1 | iex
```

## 🔧 使用示例

```powershell
# Cargo 包安装
sfer install --source cargo --name serde
sfer install --source cargo --name tokio --version 1.35.0

# 其他功能
sfer search --source npm --query react
sfer gui
```

## 📚 完整文档

- [安装指南](./INSTALLATION.md)
- [Cargo 安装指南](./CARGO_INSTALL_GUIDE.md)
- [更新日志](./CHANGELOG.md)
```

4. 上传编译好的文件：

### 编译发布版本

```powershell
# 编译 64 位版本
$env:GOOS="windows"; $env:GOARCH="amd64"
go build -ldflags "-s -w -X main.version=1.0.2" -o source-fetcher-windows-amd64.exe

# 编译 32 位版本
$env:GOOS="windows"; $env:GOARCH="386"
go build -ldflags "-s -w -X main.version=1.0.2" -o source-fetcher-windows-386.exe

# 重置环境变量
$env:GOOS=""; $env:GOARCH=""
```

5. 将以下文件拖放到 Release 页面的 "Attach binaries" 区域：
   - `source-fetcher-windows-amd64.exe`
   - `source-fetcher-windows-386.exe`

6. 点击 "Publish release"

### 方式二：使用 GitHub CLI (gh)

```powershell
# 安装 GitHub CLI（如果还没有）
# winget install --id GitHub.cli

# 登录
gh auth login

# 编译版本
# ... (同上)

# 创建 Release
gh release create v1.0.2 `
  source-fetcher-windows-amd64.exe `
  source-fetcher-windows-386.exe `
  --title "v1.0.2 - Cargo 安装支持" `
  --notes "详见 CHANGELOG.md"
```

## ✅ 步骤五：验证安装

推送和发布完成后，测试一键安装：

```powershell
# 在新的 PowerShell 窗口测试
irm https://raw.githubusercontent.com/xjwm5685-ui/source-fetcher/main/install.ps1 | iex

# 验证
sfer version
sfer search --source cargo --query serde
```

## 🐛 常见问题

### Q: 推送时提示没有权限

**A:** 检查 GitHub 认证：

```powershell
# 配置用户名和邮箱
git config --global user.name "your-username"
git config --global user.email "your-email@example.com"

# 使用 HTTPS 时，使用 Personal Access Token
# 生成 Token: https://github.com/settings/tokens
```

### Q: install.ps1 仍然 404

**A:** 确认：

1. 文件已成功推送到 GitHub
   - 访问 https://github.com/xjwm5685-ui/source-fetcher/blob/main/install.ps1
   - 如果能看到文件，说明推送成功

2. 使用正确的 raw 链接
   - 正确：`https://raw.githubusercontent.com/xjwm5685-ui/source-fetcher/main/install.ps1`
   - 错误：`https://github.com/xjwm5685-ui/source-fetcher/blob/main/install.ps1`

3. 分支名称正确
   - 如果主分支是 `master` 而不是 `main`，需要修改链接

### Q: Release 文件下载失败

**A:** 检查：

1. 文件命名必须准确匹配脚本中的格式：
   - `source-fetcher-windows-amd64.exe`
   - `source-fetcher-windows-386.exe`

2. Release 必须是 Published 状态，不能是 Draft

### Q: 安装后命令找不到

**A:** 这是正常的，需要：

1. 关闭所有 PowerShell 窗口
2. 重新打开新的 PowerShell 窗口
3. 或者运行 `refreshenv`（如果安装了 Chocolatey）

## 📋 推送前检查清单

使用此清单确保一切就绪：

### 代码检查

- [ ] 所有新功能已测试
- [ ] 编译无错误无警告
- [ ] 本地安装脚本测试通过
- [ ] 版本号已更新

### 文档检查

- [ ] README.md 已更新
- [ ] CHANGELOG.md 已更新
- [ ] 新增文档完整
- [ ] 链接和路径正确

### Git 检查

- [ ] 所有新文件已添加
- [ ] 提交信息清晰
- [ ] 没有敏感信息（密码、Token 等）
- [ ] `.gitignore` 配置正确

### GitHub 检查

- [ ] 远程仓库 URL 正确
- [ ] 有推送权限
- [ ] 分支名称正确（main/master）

### Release 检查

- [ ] 编译了所有平台版本
- [ ] 文件命名格式正确
- [ ] Release 说明完整
- [ ] Tag 版本号正确

## 🎯 推送后的工作

1. **更新文档中的链接**
   - 确认所有 GitHub 链接可访问
   - 测试一键安装命令

2. **社区宣传**
   - 更新项目主页
   - 发布更新公告
   - 社交媒体分享

3. **收集反馈**
   - 关注 Issues
   - 回复用户问题
   - 记录改进建议

## 🔄 后续更新流程

当需要发布新版本时：

```powershell
# 1. 更新代码和文档
# 2. 更新版本号
# 3. 提交和推送
git add .
git commit -m "feat: 新功能描述"
git push origin main

# 4. 编译新版本
# 5. 创建新 Release
gh release create v1.0.3 ...

# 6. 测试安装
irm https://raw.githubusercontent.com/.../install.ps1 | iex
```

## 📞 需要帮助？

- GitHub 文档: https://docs.github.com
- Git 教程: https://git-scm.com/book/zh/v2
- PowerShell 文档: https://docs.microsoft.com/powershell

---

**准备好了吗？** 按照以上步骤开始推送吧！🚀
