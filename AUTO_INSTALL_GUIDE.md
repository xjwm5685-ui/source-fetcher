# 自动安装功能使用指南

## 🎉 新功能：Choco 和 Winget 自动安装

现在 Source Fetcher 支持 **choco** 和 **winget** 的自动安装功能！

## 📋 功能说明

### 支持的安装源

| 源 | 下载 | 自动安装 | 说明 |
|-----|------|----------|------|
| npm | ✅ | ✅ | 完整的依赖管理 |
| choco | ✅ | ✅ | 需要 Chocolatey 客户端 |
| winget | ✅ | ✅ | 直接运行安装器 |
| pip | ✅ | ❌ | 仅下载 |
| cargo | ✅ | ❌ | 仅下载 |
| maven | ✅ | ❌ | 仅下载 |

## 🚀 使用方法

### 1. Choco 自动安装

#### 前提条件
- 需要安装 Chocolatey 客户端
- 建议以管理员权限运行

#### 基本用法

```powershell
# 安装最新版本
sfer install --source choco --name curl

# 安装指定版本
sfer install --source choco --name git --version 2.47.0

# 查看安装计划（不实际安装）
sfer install --source choco --name 7zip --plan
```

#### 工作原理

1. **下载** .nupkg 文件到输出目录
2. **检测** 是否有 `choco` 命令
3. **执行** `choco install <package>.nupkg -y`
4. **报告** 安装结果

#### 示例输出

```
Source: choco
Package: curl
Version: 8.20.0
Installer URL: https://community.chocolatey.org/api/v2/package/curl/8.20.0
Downloaded To: D:\downloads\curl.8.20.0.nupkg
Status: ✓ Installed Successfully
Duration: 15.3s

Installation Output:
Chocolatey v2.0.0
Installing the following packages:
curl
...
The install of curl was successful.
```

### 2. Winget 自动安装

#### 前提条件
- Windows 10/11（内置 winget）
- 建议以管理员权限运行（某些安装器需要）

#### 基本用法

```powershell
# 安装最新版本
sfer install --source winget --name Microsoft.PowerToys

# 安装指定版本
sfer install --source winget --name Microsoft.VisualStudioCode --version 1.85.0

# 查看安装计划
sfer install --source winget --name Microsoft.PowerToys --plan
```

#### 工作原理

1. **下载** 安装器文件（.exe/.msi/.msix）到输出目录
2. **检测** 安装器类型
3. **执行** 静默安装命令：
   - `.msi`: `msiexec /i <file> /quiet /norestart`
   - `.exe`: `<file> /S /silent /quiet /verysilent`
   - `.msix/.appx`: `Add-AppxPackage -Path <file>`
4. **报告** 安装结果

#### 示例输出

```
Source: winget
Package: Microsoft.PowerToys
Version: 0.99.1
Installer URL: https://github.com/microsoft/PowerToys/releases/download/v0.99.1/PowerToysSetup-0.99.1-x64.exe
Downloaded To: D:\downloads\PowerToysSetup-0.99.1-x64.exe
Status: ✓ Installed Successfully
Duration: 45.2s

Installation Output:
PowerToys Setup
Installing...
Installation completed successfully.
```

## 📝 高级用法

### 批量安装

创建配置文件 `install-tools.yaml`：

```yaml
output_dir: ./downloads
timeout: 60s

installs:
  # Choco 包
  - source: choco
    name: git
  - source: choco
    name: 7zip
  - source: choco
    name: curl
  
  # Winget 包
  - source: winget
    name: Microsoft.PowerToys
  - source: winget
    name: Microsoft.VisualStudioCode
```

执行批量安装：

```powershell
sfer batch --config install-tools.yaml
```

### 指定输出目录

```powershell
# 下载到指定目录
sfer install --source choco --name git --output D:\software

# Winget 同样支持
sfer install --source winget --name Microsoft.PowerToys --output D:\software
```

### 使用镜像加速

```powershell
# Choco 使用 NuGet 镜像
sfer install --source choco --name git --mirror nuget

# Winget 使用 jsdelivr CDN
sfer install --source winget --name Microsoft.PowerToys --mirror jsdelivr
```

## ⚠️ 注意事项

### 1. 权限要求

**Choco 安装：**
- 需要管理员权限
- 如果没有权限，会下载文件但安装失败
- 可以手动运行：`choco install <downloaded-file>.nupkg`

**Winget 安装：**
- 大多数安装器需要管理员权限
- 某些用户级安装器不需要管理员权限
- 如果自动安装失败，会提示手动安装命令

### 2. Chocolatey 客户端

如果没有安装 Chocolatey：

```powershell
# 安装 Chocolatey（需要管理员权限）
Set-ExecutionPolicy Bypass -Scope Process -Force
[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072
iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))
```

或者，下载后手动安装：

```powershell
# 1. 下载 .nupkg
sfer download --source choco --name git --output .\downloads

# 2. 解压 .nupkg（它是 ZIP 文件）
Expand-Archive .\downloads\git.2.54.0.nupkg -DestinationPath .\git-extracted

# 3. 运行安装脚本
cd .\git-extracted\tools
.\chocolateyInstall.ps1
```

### 3. 静默安装参数

不同的 EXE 安装器使用不同的静默参数：

| 安装器类型 | 常见参数 |
|-----------|---------|
| NSIS | `/S` |
| Inno Setup | `/VERYSILENT` |
| InstallShield | `/s /v"/qn"` |
| MSI | `/quiet /norestart` |

Source Fetcher 会尝试多种常见参数，但不保证所有安装器都能静默安装。

### 4. 安装失败处理

如果自动安装失败，程序会提供手动安装命令：

```
Auto-install failed: exit status 1

Manual install command:
"D:\downloads\PowerToysSetup-0.99.1-x64.exe" /S
or try: "D:\downloads\PowerToysSetup-0.99.1-x64.exe" /silent
```

## 🔧 故障排除

### 问题 1：choco 命令未找到

**错误信息：**
```
choco command not found. Please install Chocolatey first
```

**解决方案：**
1. 安装 Chocolatey（见上文）
2. 或者手动安装下载的 .nupkg 文件

### 问题 2：权限不足

**错误信息：**
```
Installation failed: exit status 5
Access is denied
```

**解决方案：**
1. 以管理员身份运行 PowerShell
2. 右键点击 PowerShell → "以管理员身份运行"
3. 重新执行安装命令

### 问题 3：安装器静默参数不正确

**错误信息：**
```
Auto-install failed: installer requires user interaction
```

**解决方案：**
1. 使用提示的手动安装命令
2. 尝试不同的静默参数
3. 或者直接双击安装器手动安装

### 问题 4：GitHub API 限流

**错误信息：**
```
API rate limit exceeded
```

**解决方案：**
1. 等待一小时后重试
2. 使用 jsdelivr 镜像：`--mirror jsdelivr`
3. 设置 GitHub Token（见下文）

## 🔐 GitHub Token 配置（可选）

为了避免 API 限流，可以配置 GitHub Token：

### 1. 创建 Token

1. 访问 https://github.com/settings/tokens
2. 点击 "Generate new token (classic)"
3. 勾选 `public_repo` 权限
4. 生成并复制 Token

### 2. 配置环境变量

```powershell
# 临时设置（当前会话）
$env:GITHUB_TOKEN = "your_token_here"

# 永久设置（用户级）
[System.Environment]::SetEnvironmentVariable("GITHUB_TOKEN", "your_token_here", "User")
```

### 3. 使用

```powershell
# Token 会自动使用
sfer install --source winget --name Microsoft.PowerToys
```

## 📊 功能对比

### 与原生工具对比

| 功能 | sfer | choco | winget |
|------|------|-------|--------|
| 下载包 | ✅ | ✅ | ❌ |
| 离线安装 | ✅ | ✅ | ❌ |
| 批量操作 | ✅ | ❌ | ❌ |
| 镜像加速 | ✅ | ❌ | ❌ |
| 断点续传 | ✅ | ❌ | ❌ |
| 并发下载 | ✅ | ❌ | ❌ |
| 自动安装 | ✅ | ✅ | ✅ |

### 优势

1. **统一接口** - 一个工具管理多种包源
2. **离线支持** - 先下载，后安装
3. **批量操作** - YAML 配置批量安装
4. **镜像加速** - 国内网络友好
5. **断点续传** - 大文件下载不怕中断

## 🎯 实用场景

### 场景 1：准备离线安装包

```powershell
# 1. 在有网络的机器上下载
sfer install --source choco --name git --output .\offline-packages --plan
sfer download --source choco --name git --output .\offline-packages

# 2. 复制到离线机器
Copy-Item .\offline-packages \\offline-pc\software\

# 3. 在离线机器上安装
choco install \\offline-pc\software\git.2.54.0.nupkg -y
```

### 场景 2：企业软件批量部署

```yaml
# deploy-tools.yaml
output_dir: \\fileserver\software
installs:
  - source: choco
    name: git
  - source: choco
    name: nodejs
  - source: winget
    name: Microsoft.VisualStudioCode
  - source: winget
    name: Microsoft.PowerToys
```

```powershell
# 下载到文件服务器
sfer batch --config deploy-tools.yaml

# 在各台机器上安装
sfer batch --config deploy-tools.yaml
```

### 场景 3：开发环境快速搭建

```powershell
# 一键安装开发工具
sfer install --source choco --name git
sfer install --source choco --name nodejs
sfer install --source winget --name Microsoft.VisualStudioCode
sfer install --source npm --name typescript --output .\workspace
```

## 📚 更多信息

- **完整文档**：`README.md`
- **快速上手**：`QUICK_START.md`
- **功能审查**：`CHOCO_WINGET_REVIEW.md`
- **项目总结**：`SUMMARY.md`

## 🆘 获取帮助

```powershell
# 查看帮助
sfer install --help

# 查看版本
sfer version

# 测试功能
sfer install --source choco --name curl --plan
```

---

**提示**：自动安装功能需要相应的权限和环境。如果遇到问题，可以先使用 `download` 命令下载，然后手动安装。
