# Source Fetcher 全局别名安装脚本
# 功能：将 source-fetcher.exe 添加到 PATH，并创建 sfer 别名

param(
    [switch]$Uninstall
)

$ErrorActionPreference = "Stop"

# 获取当前脚本所在目录（source-fetcher.exe 所在目录）
$SourceFetcherDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ExePath = Join-Path $SourceFetcherDir "source-fetcher.exe"

# 检查 source-fetcher.exe 是否存在
if (-not (Test-Path $ExePath)) {
    Write-Host "错误: 找不到 source-fetcher.exe 在 $SourceFetcherDir" -ForegroundColor Red
    Write-Host "请确保在 source-fetcher.exe 所在目录运行此脚本" -ForegroundColor Yellow
    exit 1
}

# 卸载模式
if ($Uninstall) {
    Write-Host "=== 卸载 Source Fetcher 别名 ===" -ForegroundColor Cyan
    
    # 从用户 PATH 中移除
    $UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($UserPath -like "*$SourceFetcherDir*") {
        $NewPath = ($UserPath -split ';' | Where-Object { $_ -ne $SourceFetcherDir }) -join ';'
        [Environment]::SetEnvironmentVariable("Path", $NewPath, "User")
        Write-Host "✓ 已从 PATH 中移除: $SourceFetcherDir" -ForegroundColor Green
    } else {
        Write-Host "- PATH 中未找到该目录" -ForegroundColor Yellow
    }
    
    # 删除 sfer.bat
    $SferBat = Join-Path $SourceFetcherDir "sfer.bat"
    if (Test-Path $SferBat) {
        Remove-Item $SferBat -Force
        Write-Host "✓ 已删除 sfer.bat" -ForegroundColor Green
    }
    
    # 删除 sfer.ps1
    $SferPs1 = Join-Path $SourceFetcherDir "sfer.ps1"
    if (Test-Path $SferPs1) {
        Remove-Item $SferPs1 -Force
        Write-Host "✓ 已删除 sfer.ps1" -ForegroundColor Green
    }
    
    Write-Host ""
    Write-Host "卸载完成！请重启终端使更改生效。" -ForegroundColor Green
    exit 0
}

# 安装模式
Write-Host "=== 安装 Source Fetcher 全局别名 ===" -ForegroundColor Cyan
Write-Host ""

# 1. 添加到用户 PATH
Write-Host "[1/3] 添加到 PATH 环境变量..." -ForegroundColor Yellow
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")

if ($UserPath -like "*$SourceFetcherDir*") {
    Write-Host "  - 目录已在 PATH 中: $SourceFetcherDir" -ForegroundColor Gray
} else {
    $NewPath = if ($UserPath) { "$UserPath;$SourceFetcherDir" } else { $SourceFetcherDir }
    [Environment]::SetEnvironmentVariable("Path", $NewPath, "User")
    Write-Host "  ✓ 已添加到 PATH: $SourceFetcherDir" -ForegroundColor Green
}

# 2. 创建 sfer.bat (CMD 别名)
Write-Host "[2/3] 创建 CMD 别名 (sfer.bat)..." -ForegroundColor Yellow
$SferBatPath = Join-Path $SourceFetcherDir "sfer.bat"
$SferBatContent = @"
@echo off
REM Source Fetcher 快捷别名
"%~dp0source-fetcher.exe" %*
"@
Set-Content -Path $SferBatPath -Value $SferBatContent -Encoding ASCII
Write-Host "  ✓ 已创建: $SferBatPath" -ForegroundColor Green

# 3. 创建 sfer.ps1 (PowerShell 别名)
Write-Host "[3/3] 创建 PowerShell 别名 (sfer.ps1)..." -ForegroundColor Yellow
$SferPs1Path = Join-Path $SourceFetcherDir "sfer.ps1"
$SferPs1Content = @"
# Source Fetcher 快捷别名
`$ExePath = Join-Path `$PSScriptRoot "source-fetcher.exe"
& `$ExePath @args
"@
Set-Content -Path $SferPs1Path -Value $SferPs1Content -Encoding UTF8
Write-Host "  ✓ 已创建: $SferPs1Path" -ForegroundColor Green

Write-Host ""
Write-Host "=== 安装完成！ ===" -ForegroundColor Green
Write-Host ""
Write-Host "现在你可以在任何目录使用以下命令：" -ForegroundColor Cyan
Write-Host "  sfer mirrors --source npm" -ForegroundColor White
Write-Host "  sfer search --source npm --query react" -ForegroundColor White
Write-Host "  sfer download --source npm --name react --output .\downloads" -ForegroundColor White
Write-Host ""
Write-Host "注意事项：" -ForegroundColor Yellow
Write-Host "  1. 请重启终端（CMD/PowerShell）使 PATH 更改生效" -ForegroundColor Gray
Write-Host "  2. 如果 PowerShell 提示执行策略错误，请运行：" -ForegroundColor Gray
Write-Host "     Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser" -ForegroundColor Gray
Write-Host "  3. 卸载别名请运行：" -ForegroundColor Gray
Write-Host "     .\install-alias.ps1 -Uninstall" -ForegroundColor Gray
Write-Host ""
Write-Host "测试命令：" -ForegroundColor Cyan
Write-Host "  sfer version" -ForegroundColor White
Write-Host ""
