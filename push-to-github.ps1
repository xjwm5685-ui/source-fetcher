<#
.SYNOPSIS
    One-click script to push source-fetcher to GitHub

.DESCRIPTION
    This script helps you push the project to GitHub and trigger the release.
    
.PARAMETER Username
    Your GitHub username (default: jiahe)

.PARAMETER SkipRemoteCheck
    Skip checking if remote already exists

.EXAMPLE
    .\push-to-github.ps1
    
.EXAMPLE
    .\push-to-github.ps1 -Username yourusername
#>

param(
    [string]$Username = "jiahe",
    [switch]$SkipRemoteCheck
)

$ErrorActionPreference = "Stop"

function Write-Status {
    param([string]$Message, [string]$Type = "Info")
    
    $color = switch ($Type) {
        "Success" { "Green" }
        "Warning" { "Yellow" }
        "Error"   { "Red" }
        "Info"    { "Cyan" }
        default   { "White" }
    }
    
    $icon = switch ($Type) {
        "Success" { "✅" }
        "Warning" { "⚠️ " }
        "Error"   { "❌" }
        "Info"    { "ℹ️ " }
        default   { "  " }
    }
    
    Write-Host "$icon $Message" -ForegroundColor $color
}

Write-Host "`n🚀 Source Fetcher - Push to GitHub Script`n" -ForegroundColor Magenta

# Check if we're in the right directory
if (-not (Test-Path "go.mod")) {
    Write-Status "Error: Not in source-fetcher directory!" "Error"
    Write-Status "Please run this script from d:\dy\source-fetcher" "Info"
    exit 1
}

# Check git status
Write-Status "Checking Git status..." "Info"
$gitStatus = git status --porcelain 2>&1

if ($gitStatus) {
    Write-Status "Warning: You have uncommitted changes!" "Warning"
    Write-Host ""
    git status --short
    Write-Host ""
    
    $response = Read-Host "Do you want to commit these changes? (y/n)"
    if ($response -eq "y" -or $response -eq "Y") {
        git add .
        $commitMsg = Read-Host "Enter commit message"
        git commit -m $commitMsg
        Write-Status "Changes committed" "Success"
    } else {
        Write-Status "Proceeding with uncommitted changes..." "Warning"
    }
}

# Check remote
Write-Status "Checking remote repository..." "Info"
$remoteUrl = git remote get-url origin 2>&1

if ($LASTEXITCODE -ne 0) {
    Write-Status "No remote configured. Adding remote..." "Info"
    
    $repoUrl = "https://github.com/$Username/source-fetcher.git"
    Write-Host "  Remote URL: $repoUrl" -ForegroundColor Gray
    
    if (-not $SkipRemoteCheck) {
        Write-Host ""
        Write-Status "Before continuing, make sure you've created the repository on GitHub:" "Warning"
        Write-Host "  1. Go to https://github.com/new" -ForegroundColor Yellow
        Write-Host "  2. Repository name: source-fetcher" -ForegroundColor Yellow
        Write-Host "  3. Make it Public" -ForegroundColor Yellow
        Write-Host "  4. DO NOT add README, .gitignore, or license" -ForegroundColor Yellow
        Write-Host ""
        
        $response = Read-Host "Have you created the repository on GitHub? (y/n)"
        if ($response -ne "y" -and $response -ne "Y") {
            Write-Status "Please create the repository first, then run this script again." "Info"
            exit 0
        }
    }
    
    git remote add origin $repoUrl
    Write-Status "Remote added: $repoUrl" "Success"
} else {
    Write-Status "Remote already configured: $remoteUrl" "Success"
}

# Push main branch
Write-Host ""
Write-Status "Pushing main branch..." "Info"
try {
    git push -u origin main 2>&1 | Out-Host
    
    if ($LASTEXITCODE -eq 0) {
        Write-Status "Main branch pushed successfully!" "Success"
    } else {
        Write-Status "Failed to push main branch" "Error"
        Write-Status "You may need to authenticate with GitHub" "Info"
        Write-Status "Or check if the repository exists and you have access" "Info"
        exit 1
    }
} catch {
    Write-Status "Error pushing: $_" "Error"
    exit 1
}

# Push tags
Write-Host ""
Write-Status "Pushing tags..." "Info"
try {
    git push origin v1.0.0 2>&1 | Out-Host
    
    if ($LASTEXITCODE -eq 0) {
        Write-Status "Tag v1.0.0 pushed successfully!" "Success"
    } else {
        Write-Status "Failed to push tag" "Error"
        exit 1
    }
} catch {
    Write-Status "Error pushing tag: $_" "Error"
    exit 1
}

# Success summary
Write-Host ""
Write-Host "═══════════════════════════════════════════════" -ForegroundColor Green
Write-Status "🎉 Successfully pushed to GitHub!" "Success"
Write-Host "═══════════════════════════════════════════════" -ForegroundColor Green
Write-Host ""

Write-Status "Next steps:" "Info"
Write-Host "  1. Visit: https://github.com/$Username/source-fetcher" -ForegroundColor Cyan
Write-Host "  2. Check GitHub Actions: https://github.com/$Username/source-fetcher/actions" -ForegroundColor Cyan
Write-Host "  3. Wait for builds to complete (~5-10 minutes)" -ForegroundColor Cyan
Write-Host "  4. Check Release: https://github.com/$Username/source-fetcher/releases" -ForegroundColor Cyan
Write-Host ""

Write-Status "GitHub Actions will automatically:" "Info"
Write-Host "  ✅ Run all tests" -ForegroundColor Gray
Write-Host "  ✅ Build binaries for all platforms" -ForegroundColor Gray
Write-Host "  ✅ Create GitHub Release" -ForegroundColor Gray
Write-Host "  ✅ Upload binaries to Release" -ForegroundColor Gray
Write-Host ""

Write-Status "Configure your repository:" "Info"
Write-Host "  1. Add description: 'Unified package download tool - No native clients required'" -ForegroundColor Gray
Write-Host "  2. Add topics: package-manager, npm, chocolatey, winget, go, cli" -ForegroundColor Gray
Write-Host "  3. Enable Discussions (Settings → Features)" -ForegroundColor Gray
Write-Host ""

Write-Host "📖 For detailed instructions, see: RELEASE_READY.md" -ForegroundColor Yellow
Write-Host ""
Write-Host "🎊 Congratulations on your release! 🎊" -ForegroundColor Magenta
Write-Host ""
