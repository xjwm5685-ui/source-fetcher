# Source Fetcher v1.1.0 Release 推送脚本
# 用于推送代码到 GitHub

$ErrorActionPreference = "Stop"

Write-Host "`n=== Source Fetcher v1.1.0 Release 推送 ===" -ForegroundColor Cyan
Write-Host ""

# 检查 git 状态
Write-Host "检查 Git 状态..." -ForegroundColor Yellow
git status

Write-Host ""
$continue = Read-Host "是否继续推送？(y/n)"
if ($continue -ne "y") {
    Write-Host "已取消" -ForegroundColor Red
    exit 0
}

Write-Host ""
Write-Host "添加所有更改..." -ForegroundColor Yellow
git add .

Write-Host ""
Write-Host "创建提交..." -ForegroundColor Yellow
git commit -m "release: v1.1.0 - cargo install support and one-line installation

Major updates:
- Add cargo package installation without Rust toolchain
- Add one-line PowerShell installation script
- Add Web GUI support for cargo/choco/winget installation
- Add 11+ comprehensive documentation guides

Features:
- Cargo install: Download and extract .crate without Rust
- One-line install: irm install.ps1 | iex
- Web GUI: Support npm/cargo/choco/winget installation

Documentation:
- CARGO_INSTALL_GUIDE.md - Cargo installation user guide
- INSTALLATION.md - Complete installation guide
- QUICK_INSTALL.md - Quick start guide
- PROJECT_STATUS.md - Project status report
- And 7 more comprehensive guides

Modified files:
- install_native.go - Cargo install implementation
- webgui.go - Web GUI multi-source support
- webui/app.js - Frontend source checking
- version.go, main.go - Version update to 1.1.0
- CHANGELOG.md - v1.1.0 updates

Version: 1.1.0
Release Date: 2026-06-06"

if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ 提交成功" -ForegroundColor Green
} else {
    Write-Host "✗ 提交失败" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "推送到 GitHub..." -ForegroundColor Yellow
git push origin main

if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ 推送成功！" -ForegroundColor Green
    Write-Host ""
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "下一步：创建 GitHub Release" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "方式 1: 使用 GitHub CLI（推荐）" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "gh release create v1.1.0 ``" -ForegroundColor White
    Write-Host "  dist/source-fetcher-windows-amd64.exe ``" -ForegroundColor White
    Write-Host "  dist/source-fetcher-windows-386.exe ``" -ForegroundColor White
    Write-Host "  --title 'v1.1.0 - Cargo Install Support & One-Line Installation' ``" -ForegroundColor White
    Write-Host "  --notes-file RELEASE_v1.1.0.md ``" -ForegroundColor White
    Write-Host "  --latest" -ForegroundColor White
    Write-Host ""
    Write-Host "方式 2: 使用 GitHub 网页" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "打开浏览器访问（替换 YOUR_USERNAME）:" -ForegroundColor White
    Write-Host "https://github.com/YOUR_USERNAME/source-fetcher/releases/new" -ForegroundColor Gray
    Write-Host ""
    Write-Host "详细步骤查看: RELEASE_READY.md" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
    Write-Host ""
} else {
    Write-Host "✗ 推送失败" -ForegroundColor Red
    Write-Host ""
    Write-Host "可能的原因：" -ForegroundColor Yellow
    Write-Host "  1. 网络连接问题" -ForegroundColor White
    Write-Host "  2. 没有推送权限" -ForegroundColor White
    Write-Host "  3. 远程分支不存在" -ForegroundColor White
    Write-Host ""
    Write-Host "请检查后重试：git push origin main" -ForegroundColor Yellow
    exit 1
}
