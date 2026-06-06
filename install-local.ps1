# Source Fetcher 本地安装脚本
# 用于测试安装流程或本地构建版本
# 使用方式: .\install-local.ps1

param(
    [string]$SourcePath = ".\source-fetcher.exe"
)

$ErrorActionPreference = "Stop"

# 配置
$AppName = "source-fetcher"
$CommandAlias = "sfer"

# 颜色输出函数
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

function Write-Success { param([string]$Message) Write-ColorOutput "✓ $Message" "Green" }
function Write-Info { param([string]$Message) Write-ColorOutput "ℹ $Message" "Cyan" }
function Write-Error { param([string]$Message) Write-ColorOutput "✗ $Message" "Red" }
function Write-Warning { param([string]$Message) Write-ColorOutput "⚠ $Message" "Yellow" }

Write-Host ""
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor DarkGray
Write-ColorOutput "      Source Fetcher 本地安装程序" "Cyan"
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor DarkGray
Write-Host ""

# 检查源文件
if (!(Test-Path $SourcePath)) {
    Write-Error "找不到源文件: $SourcePath"
    Write-Host ""
    Write-Info "请先编译项目:"
    Write-Host "  go build -o source-fetcher.exe"
    Write-Host ""
    exit 1
}

$SourcePath = Resolve-Path $SourcePath
Write-Info "源文件: $SourcePath"

# 获取版本
try {
    $version = & $SourcePath version 2>&1
    Write-Info "版本: $version"
} catch {
    Write-Warning "无法获取版本信息"
    $version = "unknown"
}

Write-Host ""

# 安装路径
$installDir = "$env:LOCALAPPDATA\$AppName"
$exePath = Join-Path $installDir "$AppName.exe"

Write-Info "安装目录: $installDir"

# 创建安装目录
if (!(Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
}

# 停止运行中的进程
try {
    $processes = Get-Process -Name $AppName -ErrorAction SilentlyContinue
    if ($processes) {
        Write-Warning "检测到正在运行的进程，正在停止..."
        $processes | Stop-Process -Force
        Start-Sleep -Seconds 1
    }
} catch {}

# 复制文件
try {
    Copy-Item -Path $SourcePath -Destination $exePath -Force
    Write-Success "已安装到: $exePath"
} catch {
    Write-Error "安装失败: $_"
    exit 1
}

Write-Host ""

# 配置 PATH
Write-Info "配置环境变量..."
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")

if (!($userPath -split ';' | Where-Object { $_ -eq $installDir })) {
    $newPath = "$userPath;$installDir"
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    $env:Path = "$env:Path;$installDir"
    Write-Success "已添加到 PATH"
} else {
    Write-Info "PATH 已包含安装目录"
}

Write-Host ""

# 创建命令别名
Write-Info "创建命令别名 '$CommandAlias'..."
$batchFile = Join-Path $installDir "$CommandAlias.bat"
$batchContent = @"
@echo off
"$exePath" %*
"@
Set-Content -Path $batchFile -Value $batchContent -Encoding ASCII
Write-Success "已创建别名: $CommandAlias"

Write-Host ""

# 创建卸载脚本
$uninstallScript = Join-Path $installDir "uninstall.ps1"
$scriptContent = @"
# Source Fetcher 卸载脚本
`$ErrorActionPreference = "Stop"

Write-Host "正在卸载 $AppName..." -ForegroundColor Cyan

try {
    Get-Process -Name "$AppName" -ErrorAction SilentlyContinue | Stop-Process -Force
} catch {}

`$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
`$installDir = "$installDir"
`$newPath = (`$userPath -split ';' | Where-Object { `$_ -ne `$installDir }) -join ';'
[Environment]::SetEnvironmentVariable("Path", `$newPath, "User")

Start-Sleep -Seconds 1
Remove-Item -Path "$installDir" -Recurse -Force -ErrorAction SilentlyContinue

Write-Host "✓ 卸载完成" -ForegroundColor Green
"@
Set-Content -Path $uninstallScript -Value $scriptContent -Encoding UTF8

# 验证安装
Write-Info "验证安装..."
try {
    $env:Path = [System.Environment]::GetEnvironmentVariable("Path", "Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path", "User")
    $testVersion = & $exePath version 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Success "安装验证成功 (版本: $testVersion)"
    }
} catch {
    Write-Warning "验证时出现异常"
}

Write-Host ""

# 显示使用说明
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor DarkGray
Write-Host ""
Write-ColorOutput "🎉 Source Fetcher 安装成功！" "Green"
Write-Host ""
Write-ColorOutput "快速开始:" "Cyan"
Write-Host "  $CommandAlias version                    # 查看版本"
Write-Host "  $CommandAlias search --source npm --query react    # 搜索包"
Write-Host "  $CommandAlias download --source npm --name react   # 下载包"
Write-Host "  $CommandAlias install --source npm --name react    # 安装包"
Write-Host "  $CommandAlias gui                        # 启动 Web GUI"
Write-Host ""
Write-ColorOutput "卸载:" "Cyan"
Write-Host "  $installDir\uninstall.ps1"
Write-Host ""
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor DarkGray
Write-Host ""
Write-ColorOutput "💡 刷新环境变量的方式：" "Yellow"
Write-Host ""
Write-ColorOutput "方式一（推荐）：关闭并重新打开终端" "White"
Write-Host ""
Write-ColorOutput "方式二：在当前终端运行：" "White"
Write-Host "  . .\refresh-env.ps1" -ForegroundColor Cyan
Write-Host "  # 或者"
Write-Host "  `$env:Path = [System.Environment]::GetEnvironmentVariable('Path','Machine') + ';' + [System.Environment]::GetEnvironmentVariable('Path','User')" -ForegroundColor Cyan
Write-Host ""
