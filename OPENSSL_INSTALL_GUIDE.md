# OpenSSL 安装指南

## 问题说明

你通过 Web GUI 安装了 `openssl@3.5.0`（cargo 包），但是 `openssl` 命令不可用。

这是因为：
- **cargo 的 openssl 包**：是 OpenSSL 的 Rust 语言绑定库（源码）
- **openssl 命令**：是 OpenSSL 的可执行程序

它们是不同的东西！

## 解决方案

### 选项 1: 使用 Chocolatey 安装 OpenSSL（推荐）⭐

如果你想要 `openssl.exe` 命令行工具：

```powershell
# 搜索 OpenSSL
sfer search --source choco --query openssl

# 安装 OpenSSL（需要 Chocolatey）
choco install openssl

# 或使用 source-fetcher（需要 Chocolatey）
sfer install --source choco --name openssl
```

**注意**：需要先安装 Chocolatey：
```powershell
Set-ExecutionPolicy Bypass -Scope Process -Force
iwr https://chocolatey.org/install.ps1 -UseBasicParsing | iex
```

### 选项 2: 使用 winget 安装 OpenSSL

```powershell
# 搜索 OpenSSL
sfer search --source winget --query openssl

# 使用 winget 安装
winget install OpenSSL.Light

# 或使用 source-fetcher
sfer install --source winget --name OpenSSL.Light
```

### 选项 3: 手动下载

1. 访问：https://slproweb.com/products/Win32OpenSSL.html
2. 下载 Win64 OpenSSL v3.x.x
3. 运行安装程序
4. 添加到 PATH

### 选项 4: 使用 scoop

```powershell
scoop install openssl
```

## Cargo OpenSSL 包的用途

cargo 的 openssl 包（你刚才安装的）是用于：

1. **Rust 项目开发**
   - 在 Rust 代码中使用 OpenSSL 功能
   - 编译需要 OpenSSL 的 Rust 项目

2. **查看源码**
   - 学习 OpenSSL Rust 绑定的实现
   - 参考 API 使用方法

**位置**（如果安装成功）:
```
d:\dy\source-fetcher\cargo-crates\openssl-3.5.0\
```

## 检查 Cargo 包安装

如果想确认 cargo 包是否安装成功：

```powershell
# 检查安装目录
cd d:\dy\source-fetcher
dir cargo-crates

# 如果存在 openssl-3.5.0 目录
cd cargo-crates\openssl-3.5.0
dir

# 应该看到：
# - Cargo.toml
# - src/
# - README.md
# - 等等
```

## 重新安装 Cargo OpenSSL（如果需要）

如果 cargo 包安装失败，可以重试：

```powershell
# 使用 CLI
sfer install --source cargo --name openssl --version 0.10.68

# 或启动 GUI 重新安装
sfer gui
```

## 常见问题

### Q: 为什么 cargo 包没有可执行文件？

**A**: Cargo 包是 Rust 语言的库（library），不是可执行程序。它们是：
- 源代码
- 用于其他 Rust 项目的依赖
- 需要编译才能使用

### Q: 我应该安装哪个？

**A**: 取决于你的需求：

| 需求 | 安装方式 |
|------|---------|
| 使用 `openssl` 命令行工具 | Chocolatey / winget / 手动下载 |
| 开发 Rust 项目（需要 OpenSSL） | Cargo 包 |
| 查看 OpenSSL Rust 绑定源码 | Cargo 包 |

### Q: cargo 包安装到哪里了？

**A**: 默认安装到：
```
<工作目录>/cargo-crates/<包名>-<版本>/
```

例如：
```
d:\dy\source-fetcher\cargo-crates\openssl-3.5.0\
```

### Q: 如何使用 cargo 包？

**A**: 
1. 如果你在开发 Rust 项目，在 `Cargo.toml` 中添加：
   ```toml
   [dependencies]
   openssl = "0.10.68"
   ```

2. 如果只是想查看源码，直接浏览目录即可。

## 推荐方案

如果你想要 `openssl.exe` 命令：

### Windows 10/11（推荐）

```powershell
# 1. 使用 winget（系统自带）
winget install OpenSSL.Light

# 2. 验证
openssl version
```

### 有 Chocolatey

```powershell
# 1. 使用 choco
choco install openssl

# 2. 验证
openssl version
```

### 使用 Source Fetcher GUI

```powershell
# 1. 启动 GUI
sfer gui

# 2. 在浏览器中：
#    - Source: winget 或 choco
#    - Query: openssl
#    - 选择 OpenSSL.Light 或 openssl
#    - 点击 Install

# 3. 验证
openssl version
```

## 总结

- ✅ **Cargo openssl 包**：Rust 语言的 OpenSSL 绑定库（源码）
- ✅ **OpenSSL 可执行程序**：命令行工具（openssl.exe）
- ⚠️ 它们是不同的东西！
- 💡 根据需求选择合适的安装方式

---

**需要帮助？** 查看：
- `CARGO_INSTALL_GUIDE.md` - Cargo 包安装详解
- `INSTALLATION.md` - 完整安装指南
