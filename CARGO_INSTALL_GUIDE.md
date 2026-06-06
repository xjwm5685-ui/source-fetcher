# Cargo 包安装指南

Source Fetcher 现在支持直接下载和解压 Cargo crate 源码包，**无需安装 Rust 工具链**。

## 功能特点

✅ **无需 Rust/Cargo**：直接下载 .crate 文件并解压，不依赖 Rust 工具链  
✅ **获取源码**：快速获取 Rust 项目的源代码用于学习和查看  
✅ **离线友好**：支持离线环境下的源码分发  
✅ **版本控制**：可以精确指定要下载的 crate 版本  

## 基本使用

### 安装最新版本

```powershell
sfer install --source cargo --name serde
```

### 安装指定版本

```powershell
sfer install --source cargo --name tokio --version 1.35.0
```

### 查看安装计划

```powershell
sfer install --source cargo --name ripgrep --plan
```

### 指定输出目录

```powershell
sfer install --source cargo --name clap --output ./my-crates
```

## 安装后的文件结构

安装后，源码包会被解压到以下目录：

```
<output-dir>/
└── cargo-crates/
    └── <name>-<version>/
        ├── Cargo.toml
        ├── Cargo.toml.orig
        ├── README.md
        ├── LICENSE-*
        ├── src/
        │   └── lib.rs (或 main.rs)
        └── ... (其他源文件)
```

## 典型使用场景

### 1. 查看和学习源码

```powershell
# 下载热门 crate 的源码学习
sfer install --source cargo --name serde
sfer install --source cargo --name tokio
sfer install --source cargo --name reqwest

# 解压后可以直接查看源代码
code ./cargo-crates/serde-1.0.210
```

### 2. 离线源码分发

```powershell
# 在有网络的机器上批量下载
sfer install --source cargo --name serde --output ./offline-crates
sfer install --source cargo --name tokio --output ./offline-crates
sfer install --source cargo --name clap --output ./offline-crates

# 将 ./offline-crates 目录复制到离线环境使用
```

### 3. 特定版本的源码

```powershell
# 下载指定版本用于调试或对比
sfer install --source cargo --name actix-web --version 4.4.0
sfer install --source cargo --name actix-web --version 4.5.0

# 对比不同版本的代码差异
```

## 注意事项

### 关于编译

下载的是 **源码包**，不是编译好的二进制文件。如果需要编译和运行：

1. **仍然需要安装 Rust 工具链**（从 https://rustup.rs/ 安装）
2. 进入解压目录
3. 运行 `cargo build --release`
4. 编译后的二进制文件位于 `target/release/` 目录

```powershell
# 下载源码
sfer install --source cargo --name ripgrep

# 编译（需要 Rust）
cd ./cargo-crates/ripgrep-15.1.0
cargo build --release

# 使用编译后的二进制
./target/release/rg --version
```

### 文件格式

- `.crate` 文件本质上是 **gzip 压缩的 tar 归档**
- 格式类似 npm 的 tarball
- Source Fetcher 会自动处理解压，无需手动操作

### 依赖处理

- 当前版本只下载指定的 crate 本身
- **不会递归下载依赖**
- 如需完整编译，`cargo build` 会自动下载所需依赖

## 高级用法

### 结合搜索功能

```powershell
# 搜索 crate
sfer search --source cargo --query http

# 交互式选择并安装
sfer search --source cargo --query http --interactive
```

### 批量下载配置

可以在 `source-fetcher.yaml` 中配置批量下载：

```yaml
downloads:
  - source: cargo
    name: serde
    version: latest
    output: ./crates
  
  - source: cargo
    name: tokio
    version: 1.35.0
    output: ./crates
```

然后执行：

```powershell
sfer batch --config source-fetcher.yaml
```

## 常见问题

### Q: 为什么不能直接运行下载的包？

A: 下载的是源码包，不是编译后的二进制。如需运行，请安装 Rust 工具链后编译。

### Q: 如何查看已安装的 crate？

A: 检查 `<output-dir>/cargo-crates/` 目录，每个子目录对应一个 crate。

### Q: 支持私有 crate registry 吗？

A: 当前版本仅支持 crates.io 官方源，未来版本计划支持私有 registry。

### Q: 可以像 cargo install 一样直接安装二进制吗？

A: 不可以。Source Fetcher 的定位是**无客户端的源码下载工具**，不替代 cargo 的完整功能。如需编译后的二进制，请使用官方的 cargo 工具链。

## 相关命令

```powershell
# 搜索 cargo 包
sfer search --source cargo --query <keyword>

# 下载 cargo 包（不安装，仅下载 .crate 文件）
sfer download --source cargo --name <package>

# 安装 cargo 包（下载并解压）
sfer install --source cargo --name <package>

# 查看镜像状态
sfer mirrors --source cargo
```

## 更多信息

- [完整文档](./README.md)
- [快速开始](./QUICK_START.md)
- [配置文件指南](./SETUP_GUIDE.md)
- [更新日志](./CHANGELOG.md)
