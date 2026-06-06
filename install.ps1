# Source Fetcher 一键安装脚本
# 使用方式: irm https://raw.githubusercontent.com/YOUR_USERNAME/source-fetcher/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

# 配置
$RepoOwner = "xjwm5685-ui"
$RepoName = "source-fetcher"
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

function Write-Success {
    param([string]$Message)
    Write-ColorOutput "✓ $Message" "Green"
}

function Write-Info {
    param([string]$Message)
    Write-ColorOutput "ℹ $Message" "Cyan"
}

function Write-Error {
    param([string]$Message)
    Write-ColorOutput "✗ $Message" "Red"
}

function Write-Warning {
    param([string]$Message)
    Write-ColorOutput "⚠ $Message" "Yellow"
}

# 检测架构
function Get-SystemArchitecture {
    $arch = [System.Environment]::Is64BitOperatingSystem
    if ($arch) {
        return "amd64"
    } else {
        return "386"
    }
}

# 获取最新版本
function Get-LatestVersion {
    try {
        Write-Info "正在获取最新版本信息..."
        $apiUrl = "https://api.github.com/repos/$RepoOwner/$RepoName/releases/latest"
        $response = Invoke-RestMethod -Uri $apiUrl -Method Get -ErrorAction Stop
        return $response.tag_name
    } catch {
        Write-Warning "无法从 GitHub API 获取版本信息，使用备用方法..."
        # 备用：尝试从 releases 页面获取
        try {
            $releasesUrl = "https://github.com/$RepoOwner/$RepoName/releases/latest"
            $response = Invoke-WebRequest -Uri $releasesUrl -UseBasicParsing
            if ($response.BaseResponse.ResponseUri.AbsoluteUri -match '/releases/tag/(.+)$') {
                return $matches[1]
            }
        } catch {
            Write-Error "无法获取最新版本信息"
            throw
        }
    }
}

# 下载文件
function Download-File {
    param(
        [string]$Url,
        [string]$OutFile
    )
    
    Write-Info "正在下载: $Url"
    
    try {
        # 创建目标目录
        $dir = Split-Path -Parent $OutFile
        if (!(Test-Path $dir)) {
            New-Item -ItemType Directory -Path $dir -Force | Out-Null
        }
        
        # 下载文件
        $ProgressPreference = 'SilentlyContinue'
        Invoke-WebRequest -Uri $Url -OutFile $OutFile -UseBasicParsing
        $ProgressPreference = 'Continue'
        
        Write-Success "下载完成"
    } catch {
        Write-Error "下载失败: $_"
        throw
    }
}

# 安装主程序
function Install-Application {
    param(
        [string]$Version,
        [string]$Architecture
    )
    
    # 构建下载 URL
    $fileName = "${AppName}-windows-${Architecture}.exe"
    $downloadUrl = "https://github.com/$RepoOwner/$RepoName/releases/download/$Version/$fileName"
    
    # 安装路径
    $installDir = "$env:LOCALAPPDATA\$AppName"
    $exePath = Join-Path $installDir "$AppName.exe"
    $tempFile = Join-Path $env:TEMP "$fileName"
    
    Write-Info "安装目录: $installDir"
    
    # 下载
    Download-File -Url $downloadUrl -OutFile $tempFile
    
    # 创建安装目录
    if (!(Test-Path $installDir)) {
        New-Item -ItemType Directory -Path $installDir -Force | Out-Null
    }
    
    # 如果已存在旧版本，尝试停止进程
    if (Test-Path $exePath) {
        try {
            $processes = Get-Process -Name $AppName -ErrorAction SilentlyContinue
            if ($processes) {
                Write-Warning "检测到正在运行的 $AppName 进程，正在停止..."
                $processes | Stop-Process -Force
                Start-Sleep -Seconds 1
            }
        } catch {
            # 忽略错误
        }
    }
    
    # 移动文件
    try {
        Move-Item -Path $tempFile -Destination $exePath -Force
        Write-Success "已安装到: $exePath"
    } catch {
        Write-Error "安装失败: $_"
        throw
    }
    
    return $installDir
}

# 配置环境变量
function Add-ToPath {
    param(
        [string]$Directory
    )
    
    Write-Info "配置环境变量..."
    
    # 获取用户 PATH
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    
    # 检查是否已存在
    if ($userPath -split ';' | Where-Object { $_ -eq $Directory }) {
        Write-Info "PATH 已包含安装目录"
        return
    }
    
    # 添加到 PATH
    $newPath = "$userPath;$Directory"
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    
    # 更新当前会话
    $env:Path = "$env:Path;$Directory"
    
    Write-Success "已添加到 PATH"
}

# 创建命令别名（通过批处理文件）
function Create-CommandAlias {
    param(
        [string]$InstallDir
    )
    
    Write-Info "创建命令别名 '$CommandAlias'..."
    
    $batchFile = Join-Path $InstallDir "$CommandAlias.bat"
    $exePath = Join-Path $InstallDir "$AppName.exe"
    
    # 创建批处理文件
    $batchContent = @"
@echo off
"$exePath" %*
"@
    
    Set-Content -Path $batchFile -Value $batchContent -Encoding ASCII
    Write-Success "已创建别名: $CommandAlias"
}

# 创建卸载脚本
function Create-UninstallScript {
    param(
        [string]$InstallDir
    )
    
    $uninstallScript = Join-Path $InstallDir "uninstall.ps1"
    
    $scriptContent = @"
# Source Fetcher 卸载脚本
`$ErrorActionPreference = "Stop"

Write-Host "正在卸载 $AppName..." -ForegroundColor Cyan

# 停止进程
try {
    Get-Process -Name "$AppName" -ErrorAction SilentlyContinue | Stop-Process -Force
} catch {}

# 从 PATH 中移除
`$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
`$installDir = "$InstallDir"
`$newPath = (`$userPath -split ';' | Where-Object { `$_ -ne `$installDir }) -join ';'
[Environment]::SetEnvironmentVariable("Path", `$newPath, "User")

# 删除安装目录
Start-Sleep -Seconds 1
Remove-Item -Path "$InstallDir" -Recurse -Force -ErrorAction SilentlyContinue

Write-Host "✓ 卸载完成" -ForegroundColor Green
Write-Host ""
Write-Host "如需重新安装，请运行:" -ForegroundColor Cyan
Write-Host "  irm https://raw.githubusercontent.com/$RepoOwner/$RepoName/main/install.ps1 | iex" -ForegroundColor White
"@
    
    Set-Content -Path $uninstallScript -Value $scriptContent -Encoding UTF8
}

# 验证安装
function Test-Installation {
    param(
        [string]$InstallDir
    )
    
    Write-Info "验证安装..."
    
    $exePath = Join-Path $InstallDir "$AppName.exe"
    
    if (!(Test-Path $exePath)) {
        Write-Error "安装验证失败：可执行文件不存在"
        return $false
    }
    
    try {
        # 刷新环境变量
        $env:Path = [System.Environment]::GetEnvironmentVariable("Path", "Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path", "User")
        
        # 测试命令
        $version = & $exePath version 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Success "安装验证成功 (版本: $version)"
            return $true
        } else {
            Write-Warning "无法获取版本信息"
            return $false
        }
    } catch {
        Write-Warning "验证时出现异常: $_"
        return $false
    }
}

# 显示使用说明
function Show-Usage {
    param(
        [string]$Version
    )
    
    Write-Host ""
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor DarkGray
    Write-Host ""
    Write-ColorOutput "🎉 Source Fetcher $Version 安装成功！" "Green"
    Write-Host ""
    Write-ColorOutput "快速开始:" "Cyan"
    Write-Host "  $CommandAlias version                    # 查看版本"
    Write-Host "  $CommandAlias search --source npm --query react    # 搜索包"
    Write-Host "  $CommandAlias download --source npm --name react   # 下载包"
    Write-Host "  $CommandAlias install --source npm --name react    # 安装包"
    Write-Host "  $CommandAlias gui                        # 启动 Web GUI"
    Write-Host "  $CommandAlias tui                        # 启动 TUI 界面"
    Write-Host ""
    Write-ColorOutput "支持的包源:" "Cyan"
    Write-Host "  • npm       - Node.js 包"
    Write-Host "  • pip       - Python 包"
    Write-Host "  • cargo     - Rust 包"
    Write-Host "  • maven     - Java 包"
    Write-Host "  • choco     - Chocolatey 包"
    Write-Host "  • winget    - Windows 包"
    Write-Host ""
    Write-ColorOutput "文档和帮助:" "Cyan"
    Write-Host "  $CommandAlias --help                     # 查看帮助"
    Write-Host "  https://github.com/$RepoOwner/$RepoName/blob/main/README.md"
    Write-Host ""
    Write-ColorOutput "卸载:" "Cyan"
    Write-Host "  `$env:LOCALAPPDATA\$AppName\uninstall.ps1"
    Write-Host ""
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor DarkGray
    Write-Host ""
    Write-ColorOutput "💡 提示: 请重新启动终端或运行 'refreshenv' 以刷新环境变量" "Yellow"
    Write-Host ""
}

# 主安装流程
function Main {
    try {
        Write-Host ""
        Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor DarkGray
        Write-ColorOutput "      Source Fetcher 安装程序" "Cyan"
        Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor DarkGray
        Write-Host ""
        
        # 检测架构
        $arch = Get-SystemArchitecture
        Write-Info "系统架构: $arch"
        
        # 获取最新版本
        $version = Get-LatestVersion
        Write-Info "最新版本: $version"
        Write-Host ""
        
        # 安装
        $installDir = Install-Application -Version $version -Architecture $arch
        Write-Host ""
        
        # 配置 PATH
        Add-ToPath -Directory $installDir
        Write-Host ""
        
        # 创建别名
        Create-CommandAlias -InstallDir $installDir
        Write-Host ""
        
        # 创建卸载脚本
        Create-UninstallScript -InstallDir $installDir
        
        # 验证安装
        Write-Host ""
        $testResult = Test-Installation -InstallDir $installDir
        Write-Host ""
        
        # 显示使用说明
        Show-Usage -Version $version
        
        if (!$testResult) {
            Write-Warning "安装完成，但验证失败。请尝试重新启动终端后再次测试。"
        }
        
    } catch {
        Write-Host ""
        Write-Error "安装失败: $_"
        Write-Host ""
        Write-ColorOutput "请尝试手动安装:" "Yellow"
        Write-Host "  1. 访问 https://github.com/$RepoOwner/$RepoName/releases"
        Write-Host "  2. 下载适合你系统的版本"
        Write-Host "  3. 解压到任意目录"
        Write-Host "  4. 将目录添加到 PATH 环境变量"
        Write-Host ""
        exit 1
    }
}

# 检查管理员权限（可选，但推荐）
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (!$isAdmin) {
    Write-Warning "建议以管理员权限运行以获得最佳体验"
    Write-Host "当前将以用户权限安装（仅当前用户可用）"
    Write-Host ""
    $continue = Read-Host "是否继续？(Y/n)"
    if ($continue -eq 'n' -or $continue -eq 'N') {
        exit 0
    }
}

# 运行主程序
Main
