# Source Fetcher 本地测试安装脚本
# 用于在推送到 GitHub 前测试安装流程
# 使用方式: 
#   1. 在项目目录运行: python -m http.server 8000
#   2. 在另一个终端运行: irm http://localhost:8000/test-install.ps1 | iex

$ErrorActionPreference = "Stop"

Write-Host ""
Write-Host "🧪 这是一个本地测试版本的安装脚本" -ForegroundColor Yellow
Write-Host ""

# 配置 - 使用本地路径
$LocalExePath = "d:\dy\source-fetcher\source-fetcher.exe"
$AppName = "source-fetcher"
$CommandAlias = "sfer"

# 颜色输出函数
function Write-ColorOutput {
    param([string]$Message, [string]$Color = "White")
    Write-Host $Message -ForegroundColor $Color
}

function Write-Success { param([string]$Message) Write-ColorOutput "✓ $Message" "Green" }
function Write-Info { param([string]$Message) Write-ColorOutput "ℹ $Message" "Cyan" }
function Write-Error { param([string]$Message) Write-ColorOutput "✗ $Message" "Red" }
function Write-Warning { param([string]$Message) Write-ColorOutput "⚠ $Message" "Yellow" }

Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor DarkGray
Write-ColorOutput "      Source Fetcher 测试安装程序" "Cyan"
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor DarkGray
Write-Host ""

# 检查本地文件
if (!(Test-Path $LocalExePath)) {
    Write-Error "找不到本地可执行文件: $LocalExePath"
    Write-Info "请先编译项目: go build -o source-fetcher.exe"
    exit 1
}

Write-Info "使用本地文件: $LocalExePath"

# 获取版本
try {
    $version = & $LocalExePath version 2>&1
    Write-Info "版本: $version"
} catch {
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
        Write-Warning "停止运行中的进程..."
        $processes | Stop-Process -Force
        Start-Sleep -Seconds 1
    }
} catch {}

# 复制文件
Copy-Item -Path $LocalExePath -Destination $exePath -Force
Write-Success "已安装到: $exePath"
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

# 验证
Write-Info "验证安装..."
try {
    $env:Path = [System.Environment]::GetEnvironmentVariable("Path", "Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path", "User")
    $testVersion = & $exePath version 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Success "安装验证成功"
    }
} catch {}

Write-Host ""
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor DarkGray
Write-ColorOutput "🎉 测试安装完成！" "Green"
Write-Host ""
Write-ColorOutput "测试命令:" "Cyan"
Write-Host "  sfer version"
Write-Host "  sfer search --source npm --query react"
Write-Host ""
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor DarkGray
Write-Host ""
