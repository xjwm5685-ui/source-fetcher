# 自动安装功能测试脚本

$ErrorActionPreference = "Continue"

Write-Host "=== Source Fetcher - 自动安装功能测试 ===" -ForegroundColor Cyan
Write-Host ""

# 刷新 PATH
$env:Path = [System.Environment]::GetEnvironmentVariable("Path","User") + ";" + [System.Environment]::GetEnvironmentVariable("Path","Machine")

Write-Host "=== 1. 测试 Choco 安装计划 ===" -ForegroundColor Cyan
Write-Host ""

Write-Host "测试: Choco 安装计划解析" -ForegroundColor Yellow
Write-Host "命令: sfer install --source choco --name curl --plan" -ForegroundColor Gray
sfer install --source choco --name curl --plan
Write-Host ""

Write-Host "=== 2. 测试 Winget 安装计划 ===" -ForegroundColor Cyan
Write-Host ""

Write-Host "测试: Winget 安装计划解析" -ForegroundColor Yellow
Write-Host "命令: sfer install --source winget --name Microsoft.PowerToys --plan" -ForegroundColor Gray
Write-Host "注意: 如果遇到 GitHub API 限流，请稍后重试或使用 --mirror jsdelivr" -ForegroundColor Yellow
sfer install --source winget --name Microsoft.PowerToys --plan 2>&1 | Out-String | Write-Host
Write-Host ""

Write-Host "=== 3. 功能说明 ===" -ForegroundColor Cyan
Write-Host ""

Write-Host "✅ 已实现的功能：" -ForegroundColor Green
Write-Host "  1. Choco 安装计划解析" -ForegroundColor White
Write-Host "  2. Winget 安装计划解析" -ForegroundColor White
Write-Host "  3. 自动下载安装包" -ForegroundColor White
Write-Host "  4. 自动执行安装" -ForegroundColor White
Write-Host ""

Write-Host "📋 使用方法：" -ForegroundColor Cyan
Write-Host ""

Write-Host "Choco 安装（需要 Chocolatey 客户端）：" -ForegroundColor Yellow
Write-Host "  sfer install --source choco --name curl" -ForegroundColor White
Write-Host "  sfer install --source choco --name git --version 2.47.0" -ForegroundColor White
Write-Host ""

Write-Host "Winget 安装：" -ForegroundColor Yellow
Write-Host "  sfer install --source winget --name Microsoft.PowerToys" -ForegroundColor White
Write-Host "  sfer install --source winget --name Microsoft.VisualStudioCode" -ForegroundColor White
Write-Host ""

Write-Host "⚠️ 注意事项：" -ForegroundColor Yellow
Write-Host ""
Write-Host "1. Choco 安装需要：" -ForegroundColor White
Write-Host "   - 安装 Chocolatey 客户端" -ForegroundColor Gray
Write-Host "   - 管理员权限" -ForegroundColor Gray
Write-Host "   - 如果没有 choco 命令，会下载 .nupkg 但不会安装" -ForegroundColor Gray
Write-Host ""

Write-Host "2. Winget 安装需要：" -ForegroundColor White
Write-Host "   - Windows 10/11（内置 winget）" -ForegroundColor Gray
Write-Host "   - 某些安装器需要管理员权限" -ForegroundColor Gray
Write-Host "   - 自动尝试静默安装参数" -ForegroundColor Gray
Write-Host ""

Write-Host "3. 如果自动安装失败：" -ForegroundColor White
Write-Host "   - 程序会提供手动安装命令" -ForegroundColor Gray
Write-Host "   - 可以手动运行下载的安装器" -ForegroundColor Gray
Write-Host ""

Write-Host "=== 4. 实际安装测试（可选）===" -ForegroundColor Cyan
Write-Host ""

$doInstall = Read-Host "是否要测试实际安装？这将下载并尝试安装软件。(y/N)"

if ($doInstall -eq "y" -or $doInstall -eq "Y") {
    Write-Host ""
    Write-Host "选择测试类型：" -ForegroundColor Yellow
    Write-Host "1. 测试 Choco 安装（需要 Chocolatey）" -ForegroundColor White
    Write-Host "2. 测试 Winget 安装" -ForegroundColor White
    Write-Host "3. 跳过" -ForegroundColor White
    $choice = Read-Host "请选择 (1/2/3)"
    
    switch ($choice) {
        "1" {
            Write-Host ""
            Write-Host "测试 Choco 安装..." -ForegroundColor Yellow
            
            # 检查是否有 choco 命令
            $hasChoco = Get-Command choco -ErrorAction SilentlyContinue
            if (-not $hasChoco) {
                Write-Host "⚠️ 未检测到 Chocolatey 客户端" -ForegroundColor Yellow
                Write-Host "安装 Chocolatey：" -ForegroundColor Cyan
                Write-Host "Set-ExecutionPolicy Bypass -Scope Process -Force; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))" -ForegroundColor Gray
                Write-Host ""
                Write-Host "将只下载 .nupkg 文件，不会实际安装" -ForegroundColor Yellow
            }
            
            Write-Host ""
            Write-Host "执行: sfer install --source choco --name curl --output .\test-install" -ForegroundColor Gray
            sfer install --source choco --name curl --output .\test-install
        }
        "2" {
            Write-Host ""
            Write-Host "测试 Winget 安装..." -ForegroundColor Yellow
            Write-Host "注意: 这将下载并尝试安装 Microsoft PowerToys" -ForegroundColor Yellow
            Write-Host ""
            
            $confirm = Read-Host "确认继续？(y/N)"
            if ($confirm -eq "y" -or $confirm -eq "Y") {
                Write-Host ""
                Write-Host "执行: sfer install --source winget --name Microsoft.PowerToys --output .\test-install" -ForegroundColor Gray
                sfer install --source winget --name Microsoft.PowerToys --output .\test-install
            } else {
                Write-Host "已取消" -ForegroundColor Gray
            }
        }
        default {
            Write-Host "已跳过实际安装测试" -ForegroundColor Gray
        }
    }
} else {
    Write-Host "已跳过实际安装测试" -ForegroundColor Gray
}

Write-Host ""
Write-Host "=== 测试完成 ===" -ForegroundColor Green
Write-Host ""
Write-Host "📚 更多信息：" -ForegroundColor Cyan
Write-Host "  - 详细指南: AUTO_INSTALL_GUIDE.md" -ForegroundColor White
Write-Host "  - 快速上手: QUICK_START.md" -ForegroundColor White
Write-Host "  - 完整文档: README.md" -ForegroundColor White
Write-Host ""
Write-Host "💡 提示：" -ForegroundColor Cyan
Write-Host "  - 使用 --plan 参数可以查看安装计划而不实际安装" -ForegroundColor White
Write-Host "  - 使用 --output 参数可以指定下载目录" -ForegroundColor White
Write-Host "  - 如果自动安装失败，可以手动运行下载的安装器" -ForegroundColor White
Write-Host ""
