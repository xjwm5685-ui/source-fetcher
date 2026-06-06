# Release v1.1.0 - Cargo Install Support & One-Line Installation

发布日期: 2026-06-06

## 🎉 主要更新

### ⭐ Cargo 包安装（无需 Rust 工具链）

现在可以直接下载并解压 Cargo crate 源码包，**无需安装 Rust 或 Cargo**！

```powershell
# 安装最新版本
sfer install --source cargo --name serde

# 安装指定版本
sfer install --source cargo --name tokio --version 1.35.0

# 查看安装计划
sfer install --source cargo --name ripgrep --plan
```

**特性**:
- ✅ 无需 Rust 工具链
- ✅ 直接下载 .crate 源码包
- ✅ 自动解压到 `cargo-crates/<name>-<version>/`
- ✅ 支持 CLI 和 Web GUI

### ⭐ 一键安装脚本

类似 rustup 和 Kimi Code 的便捷安装体验：

```powershell
irm https://raw.githubusercontent.com/xjwm5685-ui/source-fetcher/main/install.ps1 | iex
```

**特性**:
- ✅ 自动检测系统架构
- ✅ 从 GitHub Releases 下载
- ✅ 自动配置 PATH 环境变量
- ✅ 创建全局命令 `sfer`
- ✅ 生成卸载脚本
- ✅ 彩色输出和进度提示

### ⭐ Web GUI 多源支持

Web GUI 现在完全支持 **npm、cargo、choco、winget** 四种包源的安装！

**改进**:
- ✅ 智能包源检查
- ✅ 友好的错误提示
- ✅ 批量操作支持
- ✅ 实时队列更新

启动 Web GUI:
```powershell
sfer gui
```

## 📦 安装方式

### 一键安装（推荐） 🚀

```powershell
irm https://raw.githubusercontent.com/xjwm5685-ui/source-fetcher/main/install.ps1 | iex
```

安装后立即可用：
```powershell
sfer version
sfer search --source npm --query react
sfer gui
```

### 手动下载

1. 下载下方对应架构的可执行文件
2. 解压到任意目录
3. 运行 `install-local.ps1` 进行配置

## 📝 完整更新日志

### Added
- **⭐ Cargo install support**: 无需 Rust 工具链即可下载和解压 .crate 源码包
- **⭐ One-line installation script** (`install.ps1`): 一键安装脚本，类似 rustup
- **⭐ Web GUI multi-source support**: Web GUI 支持 npm、cargo、choco、winget 安装
- Local installation script (`install-local.ps1`): 本地安装脚本，用于离线场景
- Environment variable refresh script (`refresh-env.ps1`): 环境变量刷新工具
- 11+ comprehensive documentation files: 完整的文档系统

### Improved
- Web GUI backend now handles multiple package sources intelligently
- Better error messages in Web GUI queue display
- Smart package source detection and filtering
- Enhanced user experience with clear installation instructions

### Documentation
新增文档（11+ 个）:
- `CARGO_INSTALL_GUIDE.md` - Cargo 安装用户指南
- `CARGO_FEATURE_SUMMARY.md` - Cargo 功能技术总结
- `WEBGUI_CARGO_SUPPORT.md` - Web GUI 更新说明
- `INSTALLATION.md` - 完整安装指南
- `QUICK_INSTALL.md` - 快速开始
- `ONE_LINE_INSTALL_SUMMARY.md` - 一键安装实现总结
- `POST_INSTALL_GUIDE.md` - 安装后指南
- `PROJECT_STATUS.md` - 项目状态报告
- `NEXT_STEPS.md` - 下一步操作指南
- `CONTINUATION_SUMMARY.md` - 对话总结
- `DOCUMENTS_CREATED.md` - 文档清单

### Modified Files
核心代码修改:
- `install_native.go` - 新增 Cargo 安装实现
- `webgui.go` - Web GUI 多源支持
- `webui/app.js` - 前端包源检查
- `install.go` - 安装计划解析
- `main.go` - CLI 集成
- `README.md` - 更新功能说明
- `CHANGELOG.md` - 记录更新日志

## 🎯 支持的包源

| 源 | 搜索 | 下载 | 安装 (CLI) | 安装 (GUI) |
|---|------|------|------------|------------|
| **npm** | ✅ | ✅ | ✅ | ✅ |
| **cargo** | ✅ | ✅ | ✅ | ✅ |
| **choco** | ✅ | ✅ | ✅ | ✅ |
| **winget** | ✅ | ✅ | ✅ | ✅ |
| pip | ✅ | ✅ | ❌ | ❌ |
| maven | ✅ | ✅ | ❌ | ❌ |
| url | ❌ | ✅ | ❌ | ❌ |

## 📚 使用示例

### Cargo 安装
```powershell
# 搜索包
sfer search --source cargo --query serde

# 安装最新版本（无需 Rust）
sfer install --source cargo --name serde

# 安装指定版本
sfer install --source cargo --name tokio --version 1.35.0

# 安装后的源码位置
# ./cargo-crates/serde-1.0.210/
```

### npm 安装
```powershell
# 完整依赖管理
sfer install --source npm --name react --version ^19

# 包括开发依赖
sfer install --source npm --name react --include-dev
```

### Web GUI
```powershell
# 启动 GUI
sfer gui

# 在浏览器中:
# 1. 搜索包（支持 npm/cargo/choco/winget）
# 2. 选择包
# 3. 点击 Install
# 4. 在 Queue 标签查看进度
```

## 🔧 卸载

```powershell
# 使用安装时生成的卸载脚本
& "$env:LOCALAPPDATA\source-fetcher\uninstall.ps1"
```

## 🐛 已知问题

无严重问题。

如果遇到问题，请：
1. 查看文档: `INSTALLATION.md`, `POST_INSTALL_GUIDE.md`
2. 提交 Issue: https://github.com/xjwm5685-ui/source-fetcher/issues

## 🙏 致谢

感谢所有用户的反馈和支持！

## 📄 许可证

MIT License

---

**完整文档**: https://github.com/xjwm5685-ui/source-fetcher/blob/main/README.md  
**更新日志**: https://github.com/xjwm5685-ui/source-fetcher/blob/main/CHANGELOG.md  
**安装指南**: https://github.com/xjwm5685-ui/source-fetcher/blob/main/INSTALLATION.md
