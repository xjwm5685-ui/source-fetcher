# Cargo Build & Install Feature

## 🎉 新功能：自动编译和安装 Cargo 包

现在 Source Fetcher 支持自动编译 Cargo crate 并安装可执行文件！

## ✨ 特性

### 三种模式

1. **源码模式**（默认）
   - 只下载和解压源码
   - 无需 Rust 工具链
   - 适合查看源码、学习

2. **编译模式** (`--cargo-build`)
   - 下载源码并自动编译
   - 需要 Rust 工具链
   - 生成可执行文件

3. **安装模式** (`--cargo-install`)
   - 下载、编译并安装到系统
   - 自动安装到 `~/.cargo/bin`
   - 可直接使用命令

## 📖 使用方法

### 方式一：只下载源码（默认）

```powershell
# 无需 Rust，只下载源码
sfer install --source cargo --name ripgrep

# 结果：源码在 cargo-crates/ripgrep-14.1.1/
```

### 方式二：下载并编译

```powershell
# 需要 Rust，会自动编译
sfer install --source cargo --name ripgrep --cargo-build

# 结果：
# - 源码：cargo-crates/ripgrep-14.1.1/
# - 二进制：cargo-crates/ripgrep-14.1.1/target/release/rg.exe
```

### 方式三：下载、编译并安装

```powershell
# 需要 Rust，编译后自动安装到系统
sfer install --source cargo --name ripgrep --cargo-install

# 结果：
# - 源码：cargo-crates/ripgrep-14.1.1/
# - 安装到：~/.cargo/bin/rg.exe
# - 可以直接使用：rg --version
```

### 指定二进制名称

有些 crate 包含多个二进制文件，可以指定要编译的：

```powershell
# 只编译指定的二进制
sfer install --source cargo --name tokio --cargo-build --cargo-bin tokio-console
```

## 🔧 参数说明

| 参数 | 说明 | 需要 Rust |
|------|------|-----------|
| `--cargo-build` | 编译二进制文件 | ✅ 是 |
| `--cargo-install` | 编译并安装到系统 | ✅ 是 |
| `--cargo-bin <name>` | 指定要编译的二进制名称 | ✅ 是 |

**注意**：
- `--cargo-install` 会自动启用 `--cargo-build`
- 不指定任何参数时，只下载源码（无需 Rust）

## 📋 前提条件

### 仅查看源码
- ✅ 无需任何工具
- ✅ 无需 Rust 工具链

### 编译和安装
- ✅ 需要安装 Rust 工具链
- ✅ 需要安装 cargo 命令

### 安装 Rust

如果还没有 Rust：

**Windows (推荐)**:
```powershell
# 使用 winget
winget install Rustlang.Rustup

# 或使用 scoop
scoop install rustup

# 或使用 chocolatey
choco install rustup
```

**验证安装**:
```powershell
rustc --version
cargo --version
```

## 💡 使用示例

### 示例 1: ripgrep（快速搜索工具）

```powershell
# 编译并安装
sfer install --source cargo --name ripgrep --cargo-install

# 使用
rg "search pattern" .
```

### 示例 2: bat（更好的 cat）

```powershell
# 编译并安装
sfer install --source cargo --name bat --cargo-install

# 使用
bat README.md
```

### 示例 3: exa（更好的 ls）

```powershell
# 编译并安装
sfer install --source cargo --name exa --cargo-install

# 使用
exa --long --header
```

### 示例 4: tokio-console（Tokio 调试工具）

```powershell
# 包含多个二进制，指定要编译的
sfer install --source cargo --name tokio-console --cargo-install --cargo-bin tokio-console

# 使用
tokio-console
```

### 示例 5: 只查看源码

```powershell
# 不编译，只下载源码学习
sfer install --source cargo --name serde

# 查看源码
cd cargo-crates/serde-1.0.210
code .
```

## 📊 工作流程

### 编译模式流程

```
下载 .crate 文件
    ↓
解压到 cargo-crates/<name>-<version>/
    ↓
检查 Rust 工具链是否可用
    ↓
在源码目录执行 cargo build --release
    ↓
编译成功 → target/release/<binary>
```

### 安装模式流程

```
[编译流程]
    ↓
复制二进制文件到 ~/.cargo/bin/
    ↓
设置执行权限（Unix）
    ↓
检查 PATH 是否包含 ~/.cargo/bin
    ↓
完成！可以直接使用命令
```

## 🎯 输出信息

### 只下载源码

```
Successfully extracted ripgrep@14.1.1 to cargo-crates\ripgrep-14.1.1

ℹ️  This is a source distribution. To build and use:
  cd cargo-crates\ripgrep-14.1.1
  cargo build --release
  # Binary will be in target/release/
```

### 编译模式

```
Successfully extracted ripgrep@14.1.1 to cargo-crates\ripgrep-14.1.1

--- Building binary ---
Using cargo 1.80.0
Building in cargo-crates\ripgrep-14.1.1...
Command: cargo build --release
✅ Build successful (45.23s)
Binary: cargo-crates\ripgrep-14.1.1\target\release\rg.exe
```

### 安装模式

```
Successfully extracted ripgrep@14.1.1 to cargo-crates\ripgrep-14.1.1

--- Building binary ---
Using cargo 1.80.0
Building in cargo-crates\ripgrep-14.1.1...
✅ Build successful (45.23s)

--- Installing binary ---
✅ Installed to: C:\Users\username\.cargo\bin\rg.exe
```

## ⚠️ 注意事项

### 1. 编译时间

某些大型 crate 编译可能需要较长时间：

| Package | 大约时间 |
|---------|----------|
| ripgrep | 30-60s |
| bat | 20-40s |
| exa | 15-30s |
| tokio | 60-120s |

### 2. 磁盘空间

编译会生成 `target/` 目录，占用较多空间：

```
ripgrep/
├── src/          # 源码 (~500 KB)
└── target/       # 编译产物 (~50 MB)
    ├── debug/
    └── release/
```

### 3. PATH 设置

安装后如果命令不可用，检查 PATH：

**Windows PowerShell**:
```powershell
# 检查 PATH
$env:Path -split ';' | Select-String ".cargo"

# 临时添加（当前会话）
$env:Path += ";$env:USERPROFILE\.cargo\bin"

# 永久添加（需要重启终端）
[Environment]::SetEnvironmentVariable(
    "Path",
    $env:Path + ";$env:USERPROFILE\.cargo\bin",
    "User"
)
```

### 4. 依赖项

某些 crate 可能需要额外的系统依赖：

- **OpenSSL 相关**: 需要 OpenSSL 开发库
- **系统库绑定**: 可能需要 pkg-config
- **C/C++ 工具链**: 某些情况下需要 MSVC 或 MinGW

## 🐛 故障排除

### 问题 1: cargo 命令找不到

**错误**:
```
⚠️  Cargo not found: cargo not found in PATH
```

**解决**:
```powershell
# 安装 Rust
winget install Rustlang.Rustup

# 重启终端
```

### 问题 2: 编译失败

**错误**:
```
❌ Build failed: cargo build failed
```

**解决**:
1. 检查 Rust 版本：`rustc --version`
2. 更新 Rust：`rustup update`
3. 查看完整编译输出
4. 检查是否缺少系统依赖

### 问题 3: 安装的命令不可用

**错误**:
```
rg : The term 'rg' is not recognized...
```

**解决**:
```powershell
# 添加到 PATH
$env:Path += ";$env:USERPROFILE\.cargo\bin"

# 或重启终端
```

### 问题 4: 权限错误

**错误**:
```
❌ Install failed: permission denied
```

**解决**:
```powershell
# 确保有写入权限
mkdir $env:USERPROFILE\.cargo\bin -Force
```

## 📚 常见 Cargo 工具推荐

### 命令行工具

```powershell
# 搜索工具
sfer install --source cargo --name ripgrep --cargo-install
sfer install --source cargo --name fd-find --cargo-install

# 文件查看
sfer install --source cargo --name bat --cargo-install
sfer install --source cargo --name hexyl --cargo-install

# 目录浏览
sfer install --source cargo --name exa --cargo-install
sfer install --source cargo --name lsd --cargo-install

# 系统监控
sfer install --source cargo --name bottom --cargo-install
sfer install --source cargo --name procs --cargo-install

# Git 工具
sfer install --source cargo --name gitui --cargo-install
sfer install --source cargo --name delta --cargo-install
```

### 开发工具

```powershell
# Rust 开发
sfer install --source cargo --name cargo-edit --cargo-install
sfer install --source cargo --name cargo-watch --cargo-install

# 性能分析
sfer install --source cargo --name cargo-flamegraph --cargo-install
sfer install --source cargo --name hyperfine --cargo-install
```

## 🆚 对比：三种模式

| 特性 | 源码模式 | 编译模式 | 安装模式 |
|------|---------|---------|---------|
| 下载源码 | ✅ | ✅ | ✅ |
| 需要 Rust | ❌ | ✅ | ✅ |
| 生成二进制 | ❌ | ✅ | ✅ |
| 安装到系统 | ❌ | ❌ | ✅ |
| 直接可用 | ❌ | ❌ | ✅ |
| 查看源码 | ✅ | ✅ | ✅ |
| 磁盘占用 | 小 | 大 | 大 |
| 速度 | 快 | 慢 | 慢 |

**建议**:
- 🔍 学习源码 → 使用源码模式
- 🔨 开发调试 → 使用编译模式
- 🚀 日常使用 → 使用安装模式

## 🎓 学习资源

- [Rust 官方文档](https://www.rust-lang.org/)
- [Cargo 文档](https://doc.rust-lang.org/cargo/)
- [crates.io](https://crates.io/) - Rust 包仓库

## ✅ 总结

Source Fetcher 现在支持：

1. ✅ **源码模式**（默认）
   - 无需 Rust
   - 快速下载
   - 适合学习

2. ✅ **编译模式** (`--cargo-build`)
   - 自动编译
   - 生成可执行文件
   - 适合开发

3. ✅ **安装模式** (`--cargo-install`)
   - 一键安装
   - 自动配置
   - 适合日常使用

**快速开始**:
```powershell
# 安装 ripgrep
sfer install --source cargo --name ripgrep --cargo-install

# 使用
rg "pattern" .
```

---

**需要帮助？** 查看：
- `CARGO_INSTALL_GUIDE.md` - 基础安装指南
- `INSTALLATION.md` - 完整安装说明
- GitHub Issues - 报告问题
