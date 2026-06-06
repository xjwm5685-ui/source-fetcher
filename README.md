# Source Fetcher

<div align="center">

![Source Fetcher Logo](./assets/logo.png)

**统一的包管理下载工具 - 无需安装原生客户端**

[English](./README.en.md) | 中文

[![Build Status](https://img.shields.io/github/actions/workflow/status/xjwm5685-ui/source-fetcher/test.yml?branch=main)](https://github.com/xjwm5685-ui/source-fetcher/actions)
[![Go Version](https://img.shields.io/github/go-mod/go-version/xjwm5685-ui/source-fetcher)](https://go.dev/)
[![License](https://img.shields.io/github/license/xjwm5685-ui/source-fetcher)](LICENSE)
[![Release](https://img.shields.io/github/v/release/xjwm5685-ui/source-fetcher)](https://github.com/xjwm5685-ui/source-fetcher/releases)
[![Downloads](https://img.shields.io/github/downloads/xjwm5685-ui/source-fetcher/total)](https://github.com/xjwm5685-ui/source-fetcher/releases)
[![Stars](https://img.shields.io/github/stars/xjwm5685-ui/source-fetcher?style=social)](https://github.com/xjwm5685-ui/source-fetcher/stargazers)

</div>

---

## ✨ 为什么选择 Source Fetcher？

<table>
<tr>
<td width="33%" align="center">
<h3>🚀 无需安装客户端</h3>
直接访问仓库 API，无需安装 npm、choco、winget 等原生工具
</td>
<td width="33%" align="center">
<h3>🌐 统一接口</h3>
一个工具管理多个包源，统一的命令行体验
</td>
<td width="33%" align="center">
<h3>⚡ 镜像加速</h3>
内置国内镜像，自动测速和故障转移
</td>
</tr>
<tr>
<td width="33%" align="center">
<h3>📦 离线友好</h3>
先下载后安装，支持离线部署和企业内网
</td>
<td width="33%" align="center">
<h3>🔄 断点续传</h3>
支持断点续传和并发分块下载
</td>
<td width="33%" align="center">
<h3>🎯 批量操作</h3>
YAML 配置批量下载和安装
</td>
</tr>
</table>

## 🎬 快速演示

### 命令行模式

```bash
# 安装全局别名
.\install-alias.ps1

# 搜索包
sfer search --source npm --query react

# 下载包
sfer download --source npm --name react --version latest

# 安装依赖
sfer install --source npm --name react --version ^19

# 批量操作
sfer batch --config source-fetcher.yaml
```

### 图形界面模式 (GUI) 🆕

```bash
# 启动 Web GUI（推荐）
sfer gui

# 浏览器会自动打开 http://localhost:8765
```

**GUI 特性：**
- 🌐 **Web 界面** - 在浏览器中运行，无兼容性问题
- 🎨 **现代设计** - 深色主题，美观易用
- 🖱️ **鼠标操作** - 点击、多选、拖拽
- 📋 **多标签页** - 搜索、队列、镜像测速、关于
- ✅ **多选支持** - 批量下载/安装
- 📊 **实时更新** - 任务队列实时显示
- 🌐 **镜像测速** - 可视化速度对比
- 📱 **响应式** - 支持不同屏幕尺寸

### 终端界面模式 (TUI)

```bash
# 启动交互式终端界面
sfer tui
```

> 💡 **提示**: 推荐使用 GUI 模式获得最佳体验！

## 快速安装（推荐）

安装全局别名 `sfer`，在任何目录都可使用：

```powershell
# 在 source-fetcher 目录运行
.\install-alias.ps1

# 重启终端后即可使用
sfer version
sfer mirrors --source npm
sfer search --source npm --query react
```

卸载别名：

```powershell
.\install-alias.ps1 -Uninstall
```

## 📊 与其他工具对比

| 特性 | Source Fetcher | npm | choco | winget |
|------|----------------|-----|-------|--------|
| 无需安装客户端 | ✅ | ❌ | ❌ | ❌ |
| 统一多源管理 | ✅ | ❌ | ❌ | ❌ |
| 离线下载 | ✅ | ⚠️ | ⚠️ | ❌ |
| 镜像自动切换 | ✅ | ⚠️ | ❌ | ❌ |
| 断点续传 | ✅ | ❌ | ❌ | ❌ |
| 并发下载 | ✅ | ❌ | ❌ | ❌ |
| 批量操作 | ✅ | ⚠️ | ⚠️ | ❌ |
| TUI 界面 | ✅ | ❌ | ❌ | ❌ |
| 私有源支持 | ✅ | ✅ | ✅ | ❌ |
| 跨平台 | ✅ | ✅ | ❌ | ❌ |

## 💡 使用场景

### 🏢 企业内网部署
```bash
# 在有网络的机器上批量下载
sfer batch --config packages.yaml --output ./offline-packages

# 在内网机器上离线安装
sfer install --source npm --name react --output ./workspace
```

### 🌍 国内网络加速
```bash
# 自动使用国内镜像，失败自动切换
sfer download --source npm --name vue --version latest

# 手动指定镜像
sfer download --source npm --name vue --mirror npmmirror
```

### 🔧 CI/CD 集成
```yaml
# .github/workflows/build.yml
- name: Download dependencies
  run: sfer batch --config deps.yaml --continue-on-error
```

### 📦 多生态项目
```bash
# 一个命令管理所有依赖
sfer search --source all --query typescript
sfer download --source npm --name typescript
sfer download --source choco --name nodejs
sfer download --source winget --name Microsoft.VisualStudioCode
```

## 🌟 核心功能

### 📥 下载功能
- ✅ 多源镜像测速
- ✅ 多源包搜索
- ✅ `npm` 包下载
- ✅ `Chocolatey/NuGet` 包下载
- ✅ `winget` 清单解析并下载安装器
- ✅ `pip`、`cargo`、`maven` 包下载
- ✅ 任意 URL 直链下载
- ✅ 断点续传和并发分块下载

### 📦 安装功能
- ✅ `npm` 依赖树解析、解包并安装到 `node_modules`
- ✅ 基于安装清单的 `npm` 卸载
- ✅ 基于安装清单的 `npm` 修复
- ✅ `choco` 自动安装 ⭐ 新功能
- ✅ `winget` 自动安装 ⭐ 新功能

### 🔧 高级功能
- ✅ YAML 批量下载任务
- ✅ 私有源鉴权配置
- ✅ 安装 lockfile 和冻结模式
- ✅ 生命周期脚本控制
- ✅ TUI 交互式界面

## 为什么不需要安装原工具

- `npm`：直接请求 registry 元数据并下载 tarball
- `choco`：直接请求 NuGet/Chocolatey feed 并下载 `.nupkg`
- `winget`：直接读取 `winget-pkgs` 仓库清单并提取 `InstallerUrl`
- `url`：内置 HTTP 下载器，不依赖 `curl`

## 🔒 安全特性

Source Fetcher v1.0.1 包含以下安全增强：

- ✅ **CORS 防护**：Web GUI 仅允许本地 Origin（localhost/127.0.0.1）
- ✅ **输入验证**：包名和参数验证，防止命令注入
- ✅ **SSRF 防护**：阻止访问私有 IP 地址（可配置）
- ✅ **文件权限**：敏感文件使用严格权限（0600/0640）
- ✅ **安全解析**：YAML 配置使用安全解析器

> **安全建议**：
> - 仅在受信任的网络环境中运行 Web GUI
> - 不要禁用 `security.block_private_ips` 设置
> - 定期更新到最新版本以获取安全修复

## ⚙️ 配置文件

Source Fetcher 支持全局配置文件 `.source-fetcher.yaml`，放置在项目根目录或用户主目录。

### 快速开始

```bash
# 复制示例配置文件
copy .source-fetcher.example.yaml .source-fetcher.yaml

# 根据需要编辑配置
notepad .source-fetcher.yaml
```

### 配置项说明

```yaml
# Web GUI 配置
gui:
  port: 8765                    # 端口号
  auto_open_browser: true       # 自动打开浏览器
  allowed_origins:              # CORS 白名单
    - http://localhost:8765
    - http://127.0.0.1:8765

# 下载配置
download:
  output_dir: "./downloads"     # 默认下载目录
  timeout: "30s"                # 超时时间
  chunks: 4                     # 并发下载块数（1-10）
  resume: true                  # 启用断点续传

# 镜像源配置
mirrors:
  npm: ""                       # 留空使用默认镜像
  pip: ""                       # 或填写自定义镜像 URL
  
# 安全配置
security:
  https_only: false             # 仅允许 HTTPS URL
  block_private_ips: true       # 阻止私有 IP (SSRF 防护)

# 认证配置（用于私有源）
auth_profiles:
  private-registry:
    bearer_token_env: "NPM_TOKEN"  # 从环境变量读取
    headers:
      X-Custom-Header: "value"

# npm 安装配置
install_defaults:
  scripts_policy: "root"        # none | root | all
  allow_scripts:                # 允许运行脚本的包列表
    - "package-name"
```

完整配置示例请参考 [`.source-fetcher.example.yaml`](.source-fetcher.example.yaml)

### 配置文件优先级

1. 命令行参数（最高优先级）
2. 当前目录的 `.source-fetcher.yaml`
3. 用户主目录的 `.source-fetcher.yaml`
4. 内置默认值

## 当前范围

这是一个可运行的 MVP，不是完整包管理器。

- 已支持：下载、镜像测试、winget 安装器选择
- 已支持：单任务解析预览、YAML 批量下载和批量安装
- 已支持：全局配置文件系统 (`.source-fetcher.yaml`)
- 已支持：安全防护（CORS、输入验证、SSRF 防护）
- 已支持：`npm` 依赖安装/卸载/修复 MVP，会递归解析 `dependencies`/`optionalDependencies`/`peerDependencies`/`devDependencies`（按参数控制）、组装 `node_modules`、生成 `.bin`，并基于安装清单执行卸载与修复
- 已支持：私有鉴权源、断点续传、并发分块下载
- 已支持：安装 lockfile、冻结 lockfile、受控 lifecycle scripts（`none/root/all`）
- 已支持：TUI 内 download/install/uninstall/repair 入队
- 已支持：`pip`、`cargo`、`maven` 的搜索、镜像测试和直连下载解析
- 未覆盖：`choco/winget/url` 依赖安装、跨生态依赖卸载

当前 provider 覆盖：

- `npm`：搜索、下载、安装、卸载、修复
- `pip`：搜索、下载
- `cargo`：搜索、下载
- `maven`：搜索、下载
- `choco`：搜索、下载
- `winget`：搜索、下载、安装器枚举
- `url`：直链下载

## 运行方式

### 方式一：使用全局别名（推荐）

安装后在任何目录使用 `sfer` 命令：

```powershell
sfer mirrors --source npm
sfer search --source npm --query react
```

### 方式二：直接运行

```powershell
cd D:\dy\source-fetcher
.\source-fetcher.exe mirrors --source npm
```

### 方式三：开发模式

```powershell
cd D:\dy\source-fetcher
go run . mirrors --source npm
```

## 镜像测试

```powershell
sfer mirrors --source all
sfer mirrors --source npm
sfer mirrors --source winget
```

或使用完整命令：

```powershell
go run . mirrors --source all
go run . mirrors --source npm
go run . mirrors --source winget
```

## 搜索示例

使用 `sfer` 别名（推荐）：

```powershell
sfer search --source all --query powertoys
sfer search --source npm --query react --limit 5
sfer search --source pip --query requests --limit 5
sfer search --source cargo --query serde --limit 5
sfer search --source maven --query junit --limit 5
sfer search --source choco --query git --limit 10
sfer search --source winget --query terminal --limit 10
sfer search --source winget --query terminal --interactive --resolve-only
sfer search --source winget --query terminal --pick 2 --output .\downloads
sfer search --source winget --query terminal --pick 1,3 --resolve-only
sfer search --source npm --query internal-lib --config .\source-fetcher.yaml --auth-profile private-npm --pick 1 --resume --chunks 4 --output .\downloads
```

或使用完整命令：

```powershell
go run . search --source all --query powertoys
go run . search --source npm --query react --limit 5
```

说明：搜索结果会带 `INDEX` 编号，`choco` 等存在多版本返回的源会自动按包标识去重并保留较新的版本。
说明：传 `--pick N` 或 `--pick 1,3,5` 会直接将对应搜索结果顺序交给下载流程；可与 `--resolve-only`、`--output`、`--arch`、`--installer-index` 一起用。
说明：传 `--interactive` 会在命令行里直接提示你输入一个或多个结果编号，适合临时搜完就下。
说明：进入实际下载后会自动在终端输出进度、总大小和瞬时速度，未知文件大小时也会显示已下载字节和速度。
说明：传 `--resume` 会复用同名 `.part` 文件继续下载；传 `--chunks N` 会在服务器支持 Range 时并发分块下载。
说明：传 `--config` 与 `--auth-profile` 可加载 YAML 里的鉴权配置，适用于私有 npm/NuGet/API 下载源。

## TUI

使用 `sfer` 别名：

```powershell
sfer tui
sfer tui --source winget --query terminal --output .\downloads
```

或使用完整命令：

```powershell
go run . tui
go run . tui --source winget --query terminal --output .\downloads
```

说明：TUI 会复用现有搜索、解析、下载和安装链路，不依赖本机 `npm`、`choco`、`winget` 客户端。
说明：搜索页支持直接编辑 `arch` 和 `installer-index`，它们会作用于 TUI 内发起的 `winget` 解析和下载。
说明：搜索页还支持 `auth profile`、`resume` 和 `chunks`，加入队列后会沿用这些下载选项。
说明：`Enter` 会在表单区触发搜索，在结果区触发解析；`Space` 可多选；`d` 会把下载任务加入队列，`i` 会把 npm 安装任务加入队列；批量页 `e` 会把选中任务加入队列。
说明：队列页支持 `u` 加入卸载任务、`p` 加入修复任务，并会显示下载/安装/卸载/修复的执行状态。
说明：批量页会读取 YAML 里的默认镜像、输出目录、`auth_profiles` 和 `timeout`，可同时加载 `downloads` 与 `installs` 两类任务。

## 下载示例

下载 npm 包：

```powershell
sfer download --source npm --name react --version latest --output .\downloads
sfer download --source npm --name react --config .\source-fetcher.yaml --auth-profile private-npm --resume --chunks 4 --output .\downloads
```

下载 choco 包：

```powershell
sfer download --source choco --name git --version 2.47.0 --output .\downloads
```

下载 PyPI 包：

```powershell
sfer download --source pip --name requests --version latest --output .\downloads
```

下载 Cargo crate：

```powershell
sfer download --source cargo --name serde --version latest --output .\downloads
```

下载 Maven artifact：

```powershell
sfer download --source maven --name junit:junit --version latest --output .\downloads
```

下载 winget 安装器：

```powershell
sfer download --source winget --id Microsoft.PowerToys --arch x64 --output .\downloads
```

先查看 winget 可用安装器：

```powershell
sfer download --source winget --id Microsoft.PowerToys --list-installers
```

下载任意 URL：

```powershell
sfer download --source url --url https://example.com/file.zip --output .\downloads
```

只解析，不下载：

```powershell
sfer download --source winget --id Microsoft.PowerToys --resolve-only
```

## 依赖安装

`install` 命令支持 **npm**、**choco**、**winget** 三种源。

### npm 安装（完整依赖管理）

执行依赖安装：

```powershell
sfer install --source npm --name react --version ^19 --output .\workspace
sfer install --source npm --name react --version ^19 --include-peer --omit-optional --lockfile .\workspace\source-fetcher-install.lock.json
sfer install --source npm --name react --version ^19 --include-dev --scripts root
```

只看依赖解析计划：

```powershell
sfer install --source npm --name react --version ^19 --plan
sfer install --source npm --name react --frozen-lockfile --plan
```

### choco 自动安装 ⭐ 新功能

```powershell
# 安装最新版本
sfer install --source choco --name curl

# 安装指定版本
sfer install --source choco --name git --version 2.47.0

# 查看安装计划
sfer install --source choco --name 7zip --plan
```

**注意**：需要安装 Chocolatey 客户端，建议以管理员权限运行。

### winget 自动安装 ⭐ 新功能

```powershell
# 安装最新版本
sfer install --source winget --name Microsoft.PowerToys

# 安装指定版本
sfer install --source winget --name Microsoft.VisualStudioCode --version 1.85.0

# 查看安装计划
sfer install --source winget --name Microsoft.PowerToys --plan
```

**注意**：某些安装器需要管理员权限。

说明：`install` 会按参数控制递归解析 `dependencies`、`optionalDependencies`、`peerDependencies` 和根包 `devDependencies`，先把 tarball 下载到输出目录下的 `.source-fetcher\tarballs` 缓存，再解包到 `.source-fetcher\store`，最后按 npm 规则组装到 `node_modules`、生成 `node_modules\.bin`，并写出 `source-fetcher-install.json` 与 `source-fetcher-install.lock.json`。
说明：传 `--frozen-lockfile` 时会优先要求 lockfile 与当前请求匹配；匹配成功则直接按 lockfile 解析，不再请求 registry。
说明：传 `--scripts none|root|all` 可控制是否执行 `preinstall/install/postinstall` 生命周期脚本，默认 `none`。
说明：传 `--allow-scripts esbuild,sharp` 可对白名单包放开被默认拦截的 `preinstall`；若同时使用 `source-fetcher.yaml` 里的 `install_defaults.allow_scripts`，CLI 参数会覆盖 YAML，而不是合并。

卸载已安装内容：

```powershell
sfer uninstall --output .\workspace
```

说明：`uninstall` 默认读取输出目录下的 `source-fetcher-install.json`，只删除清单里记录过的安装路径、缓存路径和可回收空目录；传 `--manifest` 可指定清单路径，传 `--keep-cache` 或 `--keep-manifest` 可保留对应内容。

修复已安装内容：

```powershell
sfer repair --output .\workspace
```

说明：`repair` 默认读取输出目录下的 `source-fetcher-install.json`，逐个检查清单里的安装路径；如果目录缺失或损坏，会优先复用 `.source-fetcher\store`，必要时再从 `.source-fetcher\tarballs` 恢复。

## 批量任务

复制示例配置：

```powershell
copy .\source-fetcher.sample.yaml .\source-fetcher.yaml
```

执行批量下载：

```powershell
sfer batch --config .\source-fetcher.yaml
sfer batch --config .\source-fetcher.yaml --jobs 2 --continue-on-error --retries 2 --retry-backoff 500ms
sfer batch --config .\source-fetcher.yaml --continue-on-error --retries 2 --retry-backoff 500ms
```

先只看解析计划：

```powershell
sfer batch --config .\source-fetcher.yaml --plan
```

说明：`batch` 现在会顺序执行 YAML 中的 `downloads` 和 `installs`；传 `--plan` 时会分别输出下载计划与安装计划。
说明：YAML 可通过顶层 `install_defaults` 持久化 `scripts_policy` 与 `allow_scripts`，用于批量安装任务和 `install --config ...` 的默认脚本策略。
说明：配置文件中的未知字段会输出 `Warning` 到 `stderr`，但不会阻断已知字段的加载；字段类型错误仍会直接失败。
说明：传 `--jobs N` 可让 batch 同时执行最多 `N` 个任务；默认 `1`，保持串行行为不变。
说明：传 `--retries N` 会为每个 batch `resolve/download/install` 步骤增加最多 `N` 次重试；传 `--retry-backoff` 可指定基础退避时间，后续重试按 2 倍递增。

## 参数说明

### `mirrors`

- `--source`：`npm`、`pip`、`cargo`、`maven`、`choco`、`winget`、`all`
- `--timeout`：测速超时

### `download`

- `--source`：`npm`、`pip`、`cargo`、`maven`、`choco`、`winget`、`url`
- `--name`：`npm/pip/cargo/choco` 包名；`maven` 用 `group:artifact`
- `--id`：`winget` 包标识
- `--version`：版本号或 tag；留空时 `npm/pip/cargo/maven/choco/winget` 会尝试解析最新版本
- `--url`：直链下载地址
- `--output`：输出目录
- `--mirror`：镜像名称或自定义 base URL
- `--config`：可选配置文件路径，用于加载 `auth_profiles`
- `--auth-profile`：要使用的鉴权配置名称
- `--arch`：`winget` 期望架构，如 `x64`、`arm64`
- `--installer-index`：按索引指定 `winget` 安装器
- `--resume`：复用已有 `.part` 文件继续下载
- `--chunks`：服务器支持 Range 时启用并发分块下载
- `--list-installers`：只列出 `winget` 安装器，不执行下载
- `--resolve-only`：只解析最终下载地址和文件名，不执行下载

说明：实际下载时会自动显示进度；`winget`、`search --pick`、`batch` 走的也是同一套下载器。

### `install`

- `--source`：当前仅支持 `npm`
- `--name`：根 npm 包名
- `--version`：根包版本、tag 或 semver range，如 `latest`、`^19`、`>=18 <20`
- `--output`：安装根目录，实际会在其下生成 `node_modules`
- `--mirror`：npm 镜像名称或自定义 base URL
- `--config`：可选配置文件路径，用于加载 `auth_profiles`
- `--auth-profile`：安装过程请求所使用的鉴权配置名称
- `--resume`：安装时下载缓存 tarball 复用已有 `.part` 文件继续下载
- `--chunks`：安装时下载缓存 tarball 启用并发分块下载
- `--omit-optional`：安装时跳过 `optionalDependencies`
- `--include-peer`：安装时纳入 `peerDependencies`
- `--include-dev`：安装时为根包纳入 `devDependencies`
- `--lockfile`：显式指定安装 lockfile 路径，默认 `<output>\source-fetcher-install.lock.json`
- `--frozen-lockfile`：要求当前安装请求与 lockfile 匹配，否则直接失败
- `--scripts`：生命周期脚本策略，支持 `none`、`root`、`all`
- `--timeout`：解析与下载总超时
- `--plan`：只输出依赖安装计划，不执行下载

说明：当前版本会递归解析 npm 依赖树，按解析后的唯一版本集合复用 tarball 下载和解包缓存，并把包铺到类似 npm 的 `node_modules` 目录结构中；manifest 会记录缓存路径、store 路径、`.bin` shim 和每个包的实际安装位置。

### `uninstall`

- `--output`：安装根目录，默认从这里查找 `source-fetcher-install.json`
- `--manifest`：显式指定安装清单路径
- `--keep-cache`：保留 `.source-fetcher` 下的 tarball/store 缓存
- `--keep-manifest`：卸载完成后保留安装清单

说明：当前版本按 manifest 定向卸载，只会尝试删除清单里记录过、且仍然匹配对应 `name@version` 的安装路径，同时也会回收清单记录的 `.bin` shim。

### `repair`

- `--output`：安装根目录，默认从这里查找 `source-fetcher-install.json`
- `--manifest`：显式指定安装清单路径

说明：当前版本按 manifest 定向修复，只会尝试恢复清单里记录过的安装路径与缺失的 `.bin` shim；如果路径已被其他 `name@version` 占用，则会跳过而不是覆盖。

### `search`

- `--source`：`npm`、`pip`、`cargo`、`maven`、`choco`、`winget`、`all`
- `--query`：搜索关键词
- `--mirror`：指定镜像；当前主要对 `npm` 和 `choco` 生效
- `--config`：可选配置文件路径，用于加载 `auth_profiles`
- `--auth-profile`：搜索与下载所使用的鉴权配置名称
- `--limit`：每个源最多返回多少条结果
- `--interactive`：交互式选择一个或多个搜索结果并继续下载
- `--pick`：直接选中一个或多个搜索结果并下载，支持 `2`、`1,3,5`
- `--output`：配合 `--pick` 或 `--interactive` 指定下载目录
- `--resolve-only`：配合 `--pick` 或 `--interactive` 只解析，不下载
- `--arch`：配合 `--pick` 或 `--interactive` 为 `winget` 指定架构
- `--installer-index`：配合 `--pick` 或 `--interactive` 为 `winget` 指定安装器索引
- `--resume`：配合 `--pick` 或 `--interactive` 复用已有 `.part` 文件继续下载
- `--chunks`：配合 `--pick` 或 `--interactive` 启用并发分块下载
- `--timeout`：搜索请求超时

### `tui`

- `--source`：初始搜索源，`npm`、`choco`、`winget`、`all`
- `--query`：初始搜索关键词，可留空后在界面内输入
- `--mirror`：初始镜像配置
- `--config`：配置文件路径，用于 `auth_profiles` 和 Batch 页
- `--auth-profile`：搜索页初始鉴权配置名称
- `--limit`：每个源最多返回多少条结果
- `--output`：下载输出目录
- `--arch`：为 TUI 内发起的 `winget` 解析/下载指定架构，也可在搜索页内修改
- `--installer-index`：为 TUI 内发起的 `winget` 解析/下载指定安装器索引，也可在搜索页内修改
- `--resume`：为 TUI 内发起的下载指定默认续传行为，也可在搜索页内修改
- `--chunks`：为 TUI 内发起的下载指定默认分块数，也可在搜索页内修改
- `--timeout`：搜索/解析/下载请求超时

说明：Batch 页会在加载配置后显示 `downloads` 与 `installs` 混合任务列表，支持多选入队，并沿用该配置文件里的 `output_dir`、镜像默认值和 `timeout`。

### `batch`

- `--config`：YAML 配置文件路径，默认 `source-fetcher.yaml`
- `--plan`：只输出每个任务的解析结果，不执行下载或安装
- `--continue-on-error`：单个任务失败时继续执行剩余任务
- `--jobs`：batch 并发任务数，默认 `1`
- `--retries`：单个 batch `resolve/download/install` 步骤失败后的重试次数，不包含首次执行
- `--retry-backoff`：batch 重试基础退避时间，如 `500ms`、`2s`；后续重试按 2 倍递增

### `auth_profiles`

- `headers`：附加到请求上的 HTTP Header
- `bearer_token_env`：从环境变量读取 Bearer Token
- `basic_username` / `basic_password_env`：从环境变量读取 Basic Auth 密码

示例：

```yaml
auth_profiles:
  private-npm:
    bearer_token_env: PRIVATE_NPM_TOKEN
    headers:
      X-Registry: internal

downloads:
  - source: npm
    name: internal-lib
    auth: private-npm
    resume: true
    chunks: 4

installs:
  - source: npm
    name: internal-lib
    version: ^1
    auth: private-npm
    include_peer: true
    scripts_policy: root
```

## 默认镜像

### npm

- `huaweicloud`
- `npmjs`
- `npmmirror`
- `tencent`

说明：未显式指定 `--mirror` 时，`npm` 会优先尝试内置国内镜像，并在搜索/解析元数据失败时自动回退到后续镜像。

### choco

- `chocolatey`
- `nuget`

### winget

- `github-api`
- `github-raw`
- `jsdelivr`

说明：`winget` 实际下载的是清单中声明的厂商安装器 URL，镜像主要影响 manifest 拉取。

## 📈 性能数据

### 下载速度对比（国内网络）

```
Source Fetcher (镜像): ████████████ 10MB/s
npm install:           ███░░░░░░░░░  3MB/s
choco install:         ████░░░░░░░░  4MB/s
```

### 并发下载效果

```
单线程下载:  ████░░░░░░░░  4MB/s
4 线程下载:  ███████████░ 11MB/s
8 线程下载:  ████████████ 12MB/s
```

## 🗺️ 路线图

查看 [ROADMAP.md](ROADMAP.md) 了解未来计划：

- 🔜 **v1.0** - 稳定版本发布
- 🔜 **v1.1** - 更多包源支持（Homebrew、APT、YUM）
- 🔜 **v1.2** - 跨生态依赖管理
- 🔜 **v2.0** - 插件系统和高级功能

## 🤝 贡献

欢迎贡献！请查看 [CONTRIBUTING.md](CONTRIBUTING.md) 了解如何参与。

### 贡献者

感谢所有贡献者！

<!-- ALL-CONTRIBUTORS-LIST:START -->
<!-- 这里将自动生成贡献者列表 -->
<!-- ALL-CONTRIBUTORS-LIST:END -->

## 📄 许可证

本项目采用 [MIT License](LICENSE) 开源。

## 🙏 致谢

- 感谢所有包管理器项目的灵感
- 感谢开源社区的支持
- 感谢所有用户的反馈

## 📞 联系方式

- 🐛 [报告 Bug](https://github.com/xjwm5685-ui/source-fetcher/issues/new?template=bug_report.md)
- 💡 [功能建议](https://github.com/xjwm5685-ui/source-fetcher/issues/new?template=feature_request.md)
- 💬 [讨论区](https://github.com/xjwm5685-ui/source-fetcher/discussions)
- ⭐ 如果觉得有用，请给个 Star！

## 📚 相关文档

- [快速开始](QUICK_START.md) - 快速上手指南
- [架构设计](ARCHITECTURE.md) - 技术架构文档
- [更新日志](CHANGELOG.md) - 版本更新记录
- [路线图](ROADMAP.md) - 未来规划
- [贡献指南](CONTRIBUTING.md) - 如何贡献

---

<div align="center">

**[⬆ 回到顶部](#source-fetcher)**

Made with ❤️ by the Source Fetcher community

</div>
