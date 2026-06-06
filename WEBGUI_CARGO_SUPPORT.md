# Web GUI Cargo 安装支持更新

## 📋 问题描述

用户在 Web GUI 中尝试安装 cargo 包（如 `openssl@0.10.80`）时遇到错误：

```
⚠️ This package source does not support installation via Web UI. 
Only npm packages can be installed.
```

## ✅ 解决方案

已更新 Web GUI 以支持 **cargo、choco、winget** 包的安装。

## 🔧 修改内容

### 1. 后端修改 (`webgui.go`)

**文件**: `webgui.go`  
**函数**: `executeInstall()`

**修改前**:
```go
// 只支持 npm 安装
if result.Source != "npm" {
    return fmt.Errorf("install is only supported for npm packages, got: %s", result.Source)
}
```

**修改后**:
```go
// 根据源类型执行安装
switch result.Source {
case "npm":
    // npm 使用完整的依赖安装
    _, err = executeInstallPlan(ctx, s.httpClient, plan, DownloadOptions{
        OutputDir: ".",
    })
case "choco", "winget", "cargo":
    // 原生安装（choco、winget、cargo）
    _, err = executeNativeInstallPlan(ctx, s.httpClient, plan, DownloadOptions{
        OutputDir: ".",
    })
default:
    return fmt.Errorf("install is not supported for source: %s (only npm, cargo, choco, winget are supported)", result.Source)
}
```

### 2. 前端修改 (`webui/app.js`)

**文件**: `webui/app.js`  
**函数**: `handleInstall()`

**修改前**:
```javascript
// 检查是否有非 npm 包
const nonNpmPackages = [];
indexes.forEach(idx => {
    if (idx >= 0 && idx < searchResults.length) {
        const result = searchResults[idx];
        if (result.Source !== 'npm') {
            nonNpmPackages.push(`${result.Identifier} (${result.Source})`);
        }
    }
});

// 如果有非 npm 包，警告用户
if (nonNpmPackages.length > 0) {
    const message = `Warning: The following packages cannot be installed via Web UI (only npm is supported)...`;
    ...
}
```

**修改后**:
```javascript
// 检查包源支持情况
const supportedSources = ['npm', 'cargo', 'choco', 'winget'];
const unsupportedPackages = [];
const supportedPackages = [];

indexes.forEach(idx => {
    if (idx >= 0 && idx < searchResults.length) {
        const result = searchResults[idx];
        if (supportedSources.includes(result.Source)) {
            supportedPackages.push(result);
        } else {
            unsupportedPackages.push(`${result.Identifier} (${result.Source})`);
        }
    }
});

// 如果有不支持的包，警告用户
if (unsupportedPackages.length > 0) {
    const message = `⚠️ The following packages cannot be installed:\n\n${unsupportedPackages.join('\n')}\n\nOnly npm, cargo, choco, and winget packages are supported...`;
    ...
}
```

### 3. 错误消息优化

**文件**: `webui/app.js`  
**函数**: `renderQueue()` - 错误处理部分

**添加**:
```javascript
} else if (errorMsg.includes('install is not supported for source')) {
    // 提取源名称
    const match = errorMsg.match(/source: (\w+)/);
    const source = match ? match[1] : 'this source';
    errorMsg = `⚠️ Installation not supported for ${source}. Supported sources: npm, cargo, choco, winget.`;
}
```

## 🎯 支持的包源

现在 Web GUI 支持以下包源的安装：

| 源 | 安装类型 | 说明 |
|---|---|---|
| **npm** | ✅ 完整依赖管理 | 递归解析依赖，安装到 node_modules |
| **cargo** | ✅ 源码解压 | 下载并解压 .crate 到本地，无需 Rust 工具链 |
| **choco** | ✅ 自动安装 | 调用 choco 命令安装（需要 Chocolatey） |
| **winget** | ✅ 自动安装 | 自动识别安装器类型并执行 |
| pip | ❌ 暂不支持 | 仅支持下载 |
| maven | ❌ 暂不支持 | 仅支持下载 |

## 🧪 测试步骤

### 1. 启动 Web GUI

```powershell
sfer gui
# 或
.\source-fetcher.exe gui
```

### 2. 搜索 Cargo 包

在 Search 标签页：
- Source: **cargo**
- Query: **serde** 或 **tokio**
- 点击 Search

### 3. 安装测试

- 选中搜索结果
- 点击 **Install** 按钮
- 切换到 Queue 标签页查看进度

### 4. 验证结果

成功的情况：
- 状态显示为 `COMPLETED`
- 源码解压到 `./cargo-crates/<name>-<version>/`

## 📊 预期行为

### Cargo 包安装

```
[Search] serde → Select → Install
↓
[Queue] Type: install, Status: running
↓
[Backend] 
  1. resolveInstallPlan (cargo)
  2. executeNativeInstallPlan
  3. executeCargoInstall
     - 下载 .crate 文件
     - 解压到 cargo-crates/serde-1.0.210/
↓
[Queue] Status: completed
[Output] ./cargo-crates/serde-1.0.210/
```

### Choco/Winget 包安装

```
[Search] git (choco) → Select → Install
↓
[Queue] Type: install, Status: running
↓
[Backend]
  1. resolveInstallPlan (choco)
  2. executeNativeInstallPlan
  3. executeChocoInstall
     - 下载 .nupkg 文件
     - 调用 choco install
↓
[Queue] Status: completed 或 failed (如果 choco 未安装)
```

## 🎨 用户体验改进

### 1. 智能提示

- 自动识别包源类型
- 区分支持和不支持的包
- 提供清晰的错误消息

### 2. 批量操作

- 可以同时选择多个不同源的包
- 自动过滤不支持的包
- 提示用户确认

### 3. 友好的错误消息

原始错误：
```
install is not supported for source: pip (only npm, cargo, choco, winget are supported)
```

优化后：
```
⚠️ Installation not supported for pip. Supported sources: npm, cargo, choco, winget.
```

## 🔄 更新日志

### 2024-06-06

**Added**:
- ✅ Web GUI 支持 cargo 包安装
- ✅ Web GUI 支持 choco 包安装  
- ✅ Web GUI 支持 winget 包安装
- ✅ 智能包源检查和过滤
- ✅ 优化错误消息显示

**Changed**:
- 🔄 `webgui.go::executeInstall()` - 支持多种包源
- 🔄 `webui/app.js::handleInstall()` - 更新包源检查逻辑
- 🔄 `webui/app.js::renderQueue()` - 优化错误消息

## 📚 相关文档

- [Cargo 安装指南](./CARGO_INSTALL_GUIDE.md)
- [Web GUI 指南](./WEBUI_GUIDE.md)
- [安装后使用指南](./POST_INSTALL_GUIDE.md)

## 💡 使用提示

### Cargo 包安装

1. **无需 Rust 工具链**
   - 直接下载和解压源码
   - 适合查看、学习、离线分发

2. **安装后位置**
   ```
   ./cargo-crates/<name>-<version>/
   ```

3. **如需编译**
   ```powershell
   cd cargo-crates/serde-1.0.210
   cargo build --release
   ```

### Choco/Winget 安装

1. **系统要求**
   - Choco: 需要安装 Chocolatey
   - Winget: Windows 10/11 自带

2. **权限要求**
   - 某些包可能需要管理员权限

3. **安装结果**
   - 查看 Queue 标签页的输出信息

## 🐛 故障排除

### Q: Cargo 包安装失败

**A**: 检查：
1. 网络连接是否正常
2. crates.io 是否可访问
3. 磁盘空间是否充足

### Q: Choco 包安装提示未找到命令

**A**: 需要先安装 Chocolatey：
```powershell
Set-ExecutionPolicy Bypass -Scope Process -Force
iwr https://chocolatey.org/install.ps1 -UseBasicParsing | iex
```

### Q: 安装队列一直显示 running

**A**: 
1. 检查后端日志
2. 刷新页面
3. 重启 GUI 服务

## ✅ 总结

Web GUI 现在完全支持 Cargo 包安装，与 npm、choco、winget 一起，为用户提供了跨平台、多生态的统一包管理体验。

**主要改进**:
- ✅ 扩展了安装源支持
- ✅ 优化了用户体验
- ✅ 改善了错误提示
- ✅ 保持了一致的操作流程

现在可以在浏览器中轻松安装 npm、cargo、choco、winget 包！🎉
