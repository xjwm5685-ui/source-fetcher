# GitHub Release 创建指南

## 📋 准备工作（已完成 ✅）

- ✅ 编译可执行文件
  - `dist/source-fetcher-windows-amd64.exe` (9.55 MB)
  - `dist/source-fetcher-windows-386.exe` (9.05 MB)
- ✅ 更新 CHANGELOG.md
- ✅ 创建 RELEASE_v1.1.0.md

## 🚀 创建 Release 步骤

### 方式一：使用 GitHub CLI (推荐) 🔥

如果你已安装 GitHub CLI (`gh`)，可以直接运行：

```powershell
# 1. 确保已登录
gh auth status

# 如果未登录，先登录
gh auth login

# 2. 创建并推送代码
git add .
git commit -m "feat: v1.1.0 - cargo install support and one-line installation"
git push origin main

# 3. 创建 Release
gh release create v1.1.0 `
  dist/source-fetcher-windows-amd64.exe `
  dist/source-fetcher-windows-386.exe `
  --title "v1.1.0 - Cargo Install Support & One-Line Installation" `
  --notes-file RELEASE_v1.1.0.md `
  --latest

# 完成！
```

### 方式二：使用 GitHub 网页界面

#### 步骤 1: 推送代码

```powershell
# 添加所有更改
git add .

# 提交
git commit -m "feat: v1.1.0 - cargo install support and one-line installation

- Add cargo package installation without Rust toolchain
- Add one-line PowerShell installation script  
- Add Web GUI support for cargo/choco/winget installation
- Add comprehensive documentation (11+ guides)

This release includes:
- Cargo install feature (no Rust required)
- One-line installation script (install.ps1)
- Web GUI multi-source support
- Complete documentation system"

# 推送
git push origin main
```

#### 步骤 2: 在 GitHub 创建 Release

1. **打开浏览器访问**:
   ```
   https://github.com/YOUR_USERNAME/source-fetcher/releases/new
   ```
   
   > 注意：将 `YOUR_USERNAME` 替换为你的 GitHub 用户名

2. **填写 Release 信息**:

   - **Tag version**: `v1.1.0`
     - 选择 "Create new tag: v1.1.0 on publish"
   
   - **Release title**: 
     ```
     v1.1.0 - Cargo Install Support & One-Line Installation
     ```
   
   - **Describe this release**: 复制 `RELEASE_v1.1.0.md` 的内容
     - 或使用下面的精简版本

3. **上传文件**:
   - 拖拽或点击 "Attach binaries" 上传:
     - `dist/source-fetcher-windows-amd64.exe`
     - `dist/source-fetcher-windows-386.exe`

4. **设置选项**:
   - ✅ 勾选 "Set as the latest release"
   - ✅ 勾选 "Create a discussion for this release" (可选)

5. **发布**:
   - 点击 "Publish release" 按钮

## 📝 Release 描述（精简版）

如果你想要一个简短的 Release 描述，可以使用这个：

```markdown
## 🎉 主要更新

### ⭐ Cargo 包安装（无需 Rust 工具链）
直接下载并解压 Cargo crate 源码包，无需安装 Rust！

\```powershell
sfer install --source cargo --name serde
\```

### ⭐ 一键安装脚本
```powershell
irm https://raw.githubusercontent.com/YOUR_USERNAME/source-fetcher/main/install.ps1 | iex
\```

### ⭐ Web GUI 多源支持
Web GUI 现在支持 npm、cargo、choco、winget 四种包源安装！

\```powershell
sfer gui
\```

## 📦 安装

**一键安装**:
\```powershell
irm https://raw.githubusercontent.com/YOUR_USERNAME/source-fetcher/main/install.ps1 | iex
\```

**手动安装**: 下载下方对应架构的可执行文件

## 📚 文档

- [完整更新日志](https://github.com/YOUR_USERNAME/source-fetcher/blob/main/CHANGELOG.md)
- [安装指南](https://github.com/YOUR_USERNAME/source-fetcher/blob/main/INSTALLATION.md)
- [快速开始](https://github.com/YOUR_USERNAME/source-fetcher/blob/main/QUICK_INSTALL.md)

## 🎯 支持的包源

| 源 | 安装 (CLI) | 安装 (GUI) |
|---|------------|------------|
| npm | ✅ | ✅ |
| cargo | ✅ | ✅ |
| choco | ✅ | ✅ |
| winget | ✅ | ✅ |

完整功能列表见 [README](https://github.com/YOUR_USERNAME/source-fetcher/blob/main/README.md)
```

> **重要**: 记得替换 `YOUR_USERNAME` 为你的 GitHub 用户名！

## ✅ 发布后验证

### 1. 检查 Release 页面

访问：`https://github.com/YOUR_USERNAME/source-fetcher/releases`

确认：
- ✅ Tag 为 v1.1.0
- ✅ 标记为 "Latest"
- ✅ 两个可执行文件都已上传
- ✅ Release 描述显示正确

### 2. 测试一键安装

在**新的 PowerShell 窗口**中测试：

```powershell
# 测试一键安装
irm https://raw.githubusercontent.com/YOUR_USERNAME/source-fetcher/main/install.ps1 | iex

# 验证安装
sfer version
# 应该显示: 1.0.1 或 1.1.0

# 测试功能
sfer search --source npm --query react
sfer gui
```

> 注意：如果版本显示 1.0.1，需要在代码中更新版本号

### 3. 检查下载链接

确认以下链接可访问：

- `https://github.com/YOUR_USERNAME/source-fetcher/releases/download/v1.1.0/source-fetcher-windows-amd64.exe`
- `https://github.com/YOUR_USERNAME/source-fetcher/releases/download/v1.1.0/source-fetcher-windows-386.exe`

## 🐛 常见问题

### Q: 一键安装报 404 错误

**A**: 确认：
1. Release 已成功创建
2. 文件已上传
3. Tag 为 `v1.1.0`（带 v）
4. `install.ps1` 中的 URL 正确

### Q: 安装后 sfer 命令不可用

**A**: 
1. 关闭并重新打开终端
2. 或运行: `$env:Path = [System.Environment]::GetEnvironmentVariable('Path','Machine') + ';' + [System.Environment]::GetEnvironmentVariable('Path','User')`

### Q: 版本号显示错误

**A**: 需要在代码中更新版本常量：

在 `main.go` 中查找并更新：
```go
const version = "1.1.0"
```

然后重新编译并上传。

## 📊 Release 检查清单

完成后，使用这个清单验证：

```markdown
## GitHub Release 检查清单

- [ ] 代码已推送到 main 分支
- [ ] Release 已创建（v1.1.0）
- [ ] Tag 正确（v1.1.0）
- [ ] 标记为 Latest release
- [ ] source-fetcher-windows-amd64.exe 已上传
- [ ] source-fetcher-windows-386.exe 已上传
- [ ] Release 描述清晰完整
- [ ] 一键安装脚本可用
- [ ] sfer version 显示正确版本
- [ ] sfer gui 可以启动
- [ ] Cargo 安装功能正常
- [ ] Web GUI 多源支持正常
```

## 🎊 完成后

Release 创建成功后：

1. **更新 README.md**（如需要）
   - 确保一键安装命令正确
   - 更新版本号和徽章

2. **分享消息**（可选）
   - 在项目 README 添加 Release 公告
   - 社交媒体分享
   - 通知用户和贡献者

3. **收集反馈**
   - 关注 GitHub Issues
   - 查看 Discussions
   - 记录用户反馈

4. **规划下一个版本**
   - 查看 `PROJECT_STATUS.md` 中的待办事项
   - 根据用户反馈调整优先级

## 🚀 下一步

Release 发布后，可以考虑：

- [ ] 添加 GitHub Actions 自动构建
- [ ] 创建 Linux/macOS 版本
- [ ] 添加自动更新功能 (`sfer update`)
- [ ] 完善测试覆盖
- [ ] 收集用户反馈
- [ ] 规划 v1.2.0

---

**祝发布顺利！** 🎉

如果遇到问题，查看：
- `POST_INSTALL_GUIDE.md` - 安装后故障排除
- `NEXT_STEPS.md` - 详细操作指南
- GitHub Issues - 寻求帮助
