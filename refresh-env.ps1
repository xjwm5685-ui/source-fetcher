# 刷新环境变量的辅助脚本
# 使用方式: . .\refresh-env.ps1

function Refresh-Environment {
    <#
    .SYNOPSIS
    刷新当前 PowerShell 会话的环境变量，无需重启终端
    
    .DESCRIPTION
    从注册表重新加载系统和用户的 PATH 环境变量到当前会话
    
    .EXAMPLE
    Refresh-Environment
    #>
    
    Write-Host "正在刷新环境变量..." -ForegroundColor Cyan
    
    # 获取系统 PATH
    $machinePath = [System.Environment]::GetEnvironmentVariable("Path", "Machine")
    
    # 获取用户 PATH
    $userPath = [System.Environment]::GetEnvironmentVariable("Path", "User")
    
    # 合并并设置到当前会话
    $env:Path = "$machinePath;$userPath"
    
    Write-Host "✓ 环境变量已刷新" -ForegroundColor Green
    Write-Host ""
    Write-Host "现在可以使用 'sfer' 命令了" -ForegroundColor White
}

# 自动执行
Refresh-Environment
