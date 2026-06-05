<# 
.SYNOPSIS
    Build script for Source Fetcher on Windows

.DESCRIPTION
    This script builds Source Fetcher for multiple platforms and architectures.
    Supports single platform build or multi-platform builds.

.PARAMETER Platform
    Target platform: windows, linux, darwin, or all

.PARAMETER Arch
    Target architecture: amd64, arm64, or all

.PARAMETER Version
    Version string (default: dev)

.PARAMETER Clean
    Clean build artifacts before building

.PARAMETER Output
    Output directory (default: ./bin)

.EXAMPLE
    .\build.ps1
    Build for current platform (Windows amd64)

.EXAMPLE
    .\build.ps1 -Platform all -Arch all
    Build for all platforms and architectures

.EXAMPLE
    .\build.ps1 -Platform linux -Arch amd64 -Version v1.0.0
    Build Linux AMD64 binary with version v1.0.0

.EXAMPLE
    .\build.ps1 -Clean
    Clean and build for current platform
#>

param(
    [Parameter()]
    [ValidateSet("windows", "linux", "darwin", "all")]
    [string]$Platform = "windows",

    [Parameter()]
    [ValidateSet("amd64", "arm64", "all")]
    [string]$Arch = "amd64",

    [Parameter()]
    [string]$Version = "dev",

    [Parameter()]
    [switch]$Clean,

    [Parameter()]
    [string]$Output = ".\bin"
)

# Configuration
$BinaryName = "source-fetcher"
$BuildTime = (Get-Date -Format "yyyy-MM-dd_HH:mm:ss")
$GitCommit = (git rev-parse --short HEAD 2>$null) ?? "unknown"

# LDFLAGS for build
$LDFlags = "-s -w -X main.Version=$Version -X main.BuildTime=$BuildTime -X main.GitCommit=$GitCommit"

# Platform and architecture combinations
$Platforms = if ($Platform -eq "all") { @("windows", "linux", "darwin") } else { @($Platform) }
$Architectures = if ($Arch -eq "all") { @("amd64", "arm64") } else { @($Arch) }

# Colors
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

function Write-Success { param([string]$Message) Write-ColorOutput "✅ $Message" "Green" }
function Write-Info { param([string]$Message) Write-ColorOutput "ℹ️  $Message" "Cyan" }
function Write-Warning { param([string]$Message) Write-ColorOutput "⚠️  $Message" "Yellow" }
function Write-Error { param([string]$Message) Write-ColorOutput "❌ $Message" "Red" }

# Clean build artifacts
function Clean-BuildArtifacts {
    Write-Info "Cleaning build artifacts..."
    
    $ItemsToClean = @(
        ".\$BinaryName.exe",
        ".\$BinaryName",
        "$Output",
        ".\dist",
        ".\*.out",
        ".\*.test",
        ".\*.prof"
    )
    
    foreach ($item in $ItemsToClean) {
        if (Test-Path $item) {
            Remove-Item $item -Recurse -Force -ErrorAction SilentlyContinue
            Write-Info "Removed: $item"
        }
    }
    
    Write-Success "Clean complete"
}

# Build for specific platform and architecture
function Build-Binary {
    param(
        [string]$OS,
        [string]$Architecture
    )
    
    $OutputFile = "$Output\$BinaryName-$OS-$Architecture"
    if ($OS -eq "windows") {
        $OutputFile += ".exe"
    }
    
    Write-Info "Building $BinaryName for $OS/$Architecture..."
    
    # Create output directory if it doesn't exist
    if (-not (Test-Path $Output)) {
        New-Item -ItemType Directory -Path $Output -Force | Out-Null
    }
    
    # Set environment variables for cross-compilation
    $env:GOOS = $OS
    $env:GOARCH = $Architecture
    $env:CGO_ENABLED = "0"
    
    # Build command
    $BuildCommand = "go build -ldflags `"$LDFlags`" -o `"$OutputFile`" ."
    
    # Execute build
    try {
        Invoke-Expression $BuildCommand
        
        if ($LASTEXITCODE -eq 0) {
            $FileSize = (Get-Item $OutputFile).Length
            $FileSizeMB = [math]::Round($FileSize / 1MB, 2)
            Write-Success "Built: $OutputFile ($FileSizeMB MB)"
            return $true
        } else {
            Write-Error "Build failed for $OS/$Architecture"
            return $false
        }
    } catch {
        Write-Error "Build error: $_"
        return $false
    }
}

# Main build process
function Start-Build {
    Write-ColorOutput "`n🔨 Source Fetcher Build Script`n" "Magenta"
    
    # Clean if requested
    if ($Clean) {
        Clean-BuildArtifacts
        Write-Host ""
    }
    
    # Display build configuration
    Write-Info "Build Configuration:"
    Write-Host "  Version:    $Version"
    Write-Host "  Build Time: $BuildTime"
    Write-Host "  Git Commit: $GitCommit"
    Write-Host "  Platform:   $($Platforms -join ', ')"
    Write-Host "  Arch:       $($Architectures -join ', ')"
    Write-Host "  Output:     $Output"
    Write-Host ""
    
    # Build for each platform and architecture
    $SuccessCount = 0
    $FailCount = 0
    $TotalBuilds = $Platforms.Count * $Architectures.Count
    
    Write-Info "Starting build for $TotalBuilds configuration(s)..."
    Write-Host ""
    
    $StartTime = Get-Date
    
    foreach ($plat in $Platforms) {
        foreach ($arch in $Architectures) {
            if (Build-Binary -OS $plat -Architecture $arch) {
                $SuccessCount++
            } else {
                $FailCount++
            }
        }
    }
    
    $EndTime = Get-Date
    $Duration = ($EndTime - $StartTime).TotalSeconds
    
    # Summary
    Write-Host ""
    Write-ColorOutput "📊 Build Summary" "Magenta"
    Write-Host "  Total:     $TotalBuilds"
    Write-Host "  Success:   $SuccessCount" -ForegroundColor Green
    Write-Host "  Failed:    $FailCount" -ForegroundColor $(if ($FailCount -gt 0) { "Red" } else { "Gray" })
    Write-Host "  Duration:  $([math]::Round($Duration, 2))s"
    Write-Host ""
    
    if ($FailCount -eq 0) {
        Write-Success "All builds completed successfully! 🎉"
        
        # List all built binaries
        if (Test-Path $Output) {
            Write-Host ""
            Write-Info "Built binaries:"
            Get-ChildItem $Output | ForEach-Object {
                $size = [math]::Round($_.Length / 1MB, 2)
                Write-Host "  $($_.Name) - $size MB"
            }
        }
    } else {
        Write-Error "Some builds failed. Please check the output above."
        exit 1
    }
}

# Run tests
function Run-Tests {
    Write-Info "Running tests..."
    go test -v -race ./...
    
    if ($LASTEXITCODE -eq 0) {
        Write-Success "All tests passed!"
    } else {
        Write-Error "Tests failed!"
        exit 1
    }
}

# Check Go installation
function Test-GoInstalled {
    try {
        $null = go version
        return $true
    } catch {
        Write-Error "Go is not installed or not in PATH"
        Write-Info "Download Go from: https://golang.org/dl/"
        return $false
    }
}

# Main execution
if (-not (Test-GoInstalled)) {
    exit 1
}

# Start the build
Start-Build

# Exit with appropriate code
exit $(if ($FailCount -eq 0) { 0 } else { 1 })
