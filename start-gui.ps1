#!/usr/bin/env pwsh
# Source Fetcher Web GUI 启动脚本

Write-Host "🚀 Starting Source Fetcher Web GUI..." -ForegroundColor Cyan
Write-Host ""

# 检查可执行文件
if (-not (Test-Path ".\source-fetcher.exe")) {
    Write-Host "❌ Error: source-fetcher.exe not found!" -ForegroundColor Red
    Write-Host "Please build the project first: go build -o source-fetcher.exe ." -ForegroundColor Yellow
    exit 1
}

# 启动 GUI
Write-Host "🌐 Starting Web GUI server..." -ForegroundColor Green
Write-Host "📱 Browser will open automatically at http://localhost:8765" -ForegroundColor Green
Write-Host ""
Write-Host "💡 Tips:" -ForegroundColor Yellow
Write-Host "  - Press Ctrl+C to stop the server" -ForegroundColor Gray
Write-Host "  - If browser doesn't open, visit http://localhost:8765 manually" -ForegroundColor Gray
Write-Host ""

.\source-fetcher.exe gui
