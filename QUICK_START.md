# Source Fetcher 快速上手指南

## 一、安装全局别名（5秒完成）

```powershell
# 1. 进入 source-fetcher 目录
cd D:\dy\source-fetcher

# 2. 运行安装脚本
.\install-alias.ps1

# 3. 重启终端
# 关闭当前终端，重新打开一个新的 PowerShell 或 CMD 窗口
```

## 二、验证安装

```powershell
# 在任意目录运行
sfer version
# 输出: dev

# 测试镜像速度
sfer mirrors --source npm
```

## 三、常用命令速查

### 1. 搜索包

```powershell
# 搜索 npm 包
sfer search --source npm --query react

# 搜索 winget 应用
sfer search --source winget --query vscode

# 搜索所有源
sfer search --source all --query git
```

### 2. 下载包

```powershell
# 下载 npm 包
sfer download --source npm --name react --output .\downloads

# 下载 winget 应用
sfer download --source winget --id Microsoft.PowerToys --output .\downloads

# 下载任意 URL
sfer download --source url --url https://example.com/file.zip --output .\downloads
```

### 3. 安装 npm 依赖

```powershell
# 安装到当前目录
sfer install --source npm --name react --version ^19

# 安装到指定目录
sfer install --source npm --name express --output .\my-project

# 查看安装计划（不实际安装）
sfer install --source npm --name lodash --plan
```

### 4. 安装 choco 包 ⭐ 新功能

```powershell
# 安装最新版本（需要 Chocolatey 客户端）
sfer install --source choco --name curl

# 安装指定版本
sfer install --source choco --name git --version 2.47.0

# 查看安装计划
sfer install --source choco --name 7zip --plan
```

### 5. 安装 winget 包 ⭐ 新功能

```powershell
# 安装最新版本
sfer install --source winget --name Microsoft.PowerToys

# 安装指定版本
sfer install --source winget --name Microsoft.VisualStudioCode --version 1.85.0

# 查看安装计划
sfer install --source winget --name Microsoft.PowerToys --plan
```

### 6. 卸载和修复

```powershell
# 卸载已安装的包（仅 npm）
sfer uninstall --output .\my-project

# 修复损坏的安装（仅 npm）
sfer repair --output .\my-project
```

**注意**：卸载和修复功能目前仅支持 npm 包。

### 7. 批量任务

```powershell
# 创建配置文件
copy source-fetcher.sample.yaml my-config.yaml

# 编辑 my-config.yaml，添加要下载的包

# 执行批量下载
sfer batch --config my-config.yaml

# 并发执行（2个任务同时运行）
sfer batch --config my-config.yaml --jobs 2
```

### 8. TUI 图形界面

```powershell
# 启动交互式界面
sfer tui

# 带初始搜索启动
sfer tui --source npm --query react
```

## 四、高级功能

### 断点续传

```powershell
# 下载大文件时支持断点续传
sfer download --source npm --name typescript --resume --output .\downloads
```

### 并发分块下载

```powershell
# 使用 4 个线程并发下载
sfer download --source npm --name webpack --chunks 4 --output .\downloads
```

### 私有源鉴权

```powershell
# 1. 编辑 source-fetcher.yaml，添加鉴权配置
# auth_profiles:
#   my-private-npm:
#     bearer_token_env: MY_NPM_TOKEN

# 2. 设置环境变量
$env:MY_NPM_TOKEN = "your-token-here"

# 3. 使用鉴权配置下载
sfer download --source npm --name private-package --auth-profile my-private-npm --output .\downloads
```

### 镜像加速

```powershell
# 测试镜像速度
sfer mirrors --source npm

# 使用指定镜像下载
sfer download --source npm --name react --mirror npmmirror --output .\downloads
```

## 五、卸载别名

如果不再需要全局别名：

```powershell
cd D:\dy\source-fetcher
.\install-alias.ps1 -Uninstall
```

## 六、常见问题

### Q1: 提示 "无法识别 sfer 命令"

**A:** 请确保：
1. 已运行 `.\install-alias.ps1`
2. 已重启终端
3. 检查 PATH 环境变量是否包含 `D:\dy\source-fetcher`

### Q2: PowerShell 提示执行策略错误

**A:** 运行以下命令：
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### Q3: 如何查看所有可用命令？

**A:** 运行：
```powershell
sfer help
```

### Q4: 如何查看某个命令的详细参数？

**A:** 运行：
```powershell
sfer download --help
sfer install --help
sfer search --help
```

## 七、实用示例

### 示例 1：快速下载 React 全家桶

```powershell
# 创建项目目录
mkdir my-react-app
cd my-react-app

# 安装 React
sfer install --source npm --name react --version ^18

# 安装 React DOM
sfer install --source npm --name react-dom --version ^18

# 安装开发工具
sfer install --source npm --name vite --include-dev
```

### 示例 2：批量下载常用工具

创建 `tools.yaml`：
```yaml
output_dir: ./tools
downloads:
  - source: winget
    id: Microsoft.PowerToys
  - source: winget
    id: Microsoft.VisualStudioCode
  - source: choco
    name: git
```

执行下载：
```powershell
sfer batch --config tools.yaml
```

### 示例 3：搜索并交互式下载

```powershell
# 搜索包
sfer search --source npm --query axios

# 交互式选择下载
sfer search --source npm --query axios --interactive --output .\downloads
```

## 八、更多帮助

- 查看完整文档：`README.md`
- 查看示例配置：`source-fetcher.sample.yaml`
- 运行测试：`go test ./...`
- 查看版本：`sfer version`

---

**提示**：所有 `sfer` 命令都可以用 `.\source-fetcher.exe` 或 `go run .` 替代，效果完全相同。
