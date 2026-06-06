# Source Fetcher v1.2.0 Release Push Script
# This script helps push the v1.2.0 release to GitHub

Write-Host "=" -NoNewline -ForegroundColor Cyan
Write-Host "=" * 60 -ForegroundColor Cyan
Write-Host "  Source Fetcher v1.2.0 Release Push Script" -ForegroundColor Yellow
Write-Host "=" * 61 -ForegroundColor Cyan
Write-Host ""

# Check if we're in the right directory
if (-not (Test-Path "version.go")) {
    Write-Host "❌ Error: Must run from source-fetcher root directory" -ForegroundColor Red
    exit 1
}

# Verify version
Write-Host "📋 Checking version..." -ForegroundColor Cyan
$versionContent = Get-Content "version.go" -Raw
if ($versionContent -match 'Version = "1\.2\.0"') {
    Write-Host "✅ Version confirmed: 1.2.0" -ForegroundColor Green
} else {
    Write-Host "❌ Error: Version not set to 1.2.0" -ForegroundColor Red
    exit 1
}

# Check git status
Write-Host ""
Write-Host "📋 Checking git status..." -ForegroundColor Cyan
$status = git status --porcelain
if ($status) {
    Write-Host "📝 Uncommitted changes found:" -ForegroundColor Yellow
    git status --short
    Write-Host ""
    $commit = Read-Host "Commit these changes? (y/n)"
    if ($commit -eq "y") {
        Write-Host "📝 Committing changes..." -ForegroundColor Cyan
        git add .
        git commit -m "Release v1.2.0 - Cargo Build & Install Feature

- Added cargo build and install functionality
- Three installation modes: source/build/install
- Updated documentation
- Updated version to 1.2.0
- Created release notes and binaries"
        Write-Host "✅ Changes committed" -ForegroundColor Green
    } else {
        Write-Host "❌ Aborted: Please commit changes first" -ForegroundColor Red
        exit 1
    }
} else {
    Write-Host "✅ No uncommitted changes" -ForegroundColor Green
}

# Check binaries
Write-Host ""
Write-Host "📋 Checking release binaries..." -ForegroundColor Cyan
if ((Test-Path "dist/source-fetcher-windows-amd64.exe") -and (Test-Path "dist/source-fetcher-windows-386.exe")) {
    Write-Host "✅ Release binaries found:" -ForegroundColor Green
    Get-ChildItem dist\*.exe | ForEach-Object {
        $sizeMB = [math]::Round($_.Length/1MB, 2)
        Write-Host "   - $($_.Name) ($sizeMB MB)" -ForegroundColor Gray
    }
} else {
    Write-Host "⚠️  Warning: Release binaries not found in dist/" -ForegroundColor Yellow
    Write-Host "   Run the build commands from CREATE_GITHUB_RELEASE_v1.2.0.md" -ForegroundColor Gray
}

# Show tag info
Write-Host ""
Write-Host "📋 Tag information:" -ForegroundColor Cyan
Write-Host "   Tag name: v1.2.0" -ForegroundColor Gray
Write-Host "   Tag message: Release v1.2.0 - Cargo Build & Install Feature" -ForegroundColor Gray

# Confirm push
Write-Host ""
Write-Host "🚀 Ready to push v1.2.0 release!" -ForegroundColor Green
Write-Host ""
Write-Host "This will:" -ForegroundColor Yellow
Write-Host "   1. Push commits to origin/main" -ForegroundColor Gray
Write-Host "   2. Create and push tag v1.2.0" -ForegroundColor Gray
Write-Host ""
$confirm = Read-Host "Continue? (y/n)"

if ($confirm -ne "y") {
    Write-Host "❌ Aborted by user" -ForegroundColor Yellow
    exit 0
}

# Push commits
Write-Host ""
Write-Host "📤 Pushing commits to origin/main..." -ForegroundColor Cyan
try {
    git push origin main
    Write-Host "✅ Commits pushed successfully" -ForegroundColor Green
} catch {
    Write-Host "❌ Error pushing commits: $_" -ForegroundColor Red
    exit 1
}

# Create and push tag
Write-Host ""
Write-Host "🏷️  Creating tag v1.2.0..." -ForegroundColor Cyan
try {
    git tag -a v1.2.0 -m "Release v1.2.0 - Cargo Build & Install Feature

Major Features:
- Cargo crate automatic compilation
- Cargo binary system installation  
- Three installation modes: source/build/install
- Cross-platform support

New CLI Flags:
- --cargo-build: Compile binary from source
- --cargo-install: Compile and install to system
- --cargo-bin: Specify binary name to build

Documentation:
- CARGO_BUILD_FEATURE.md: Complete user guide
- CARGO_BUILD_IMPLEMENTATION.md: Technical details
- Updated README.md and CHANGELOG.md

See RELEASE_v1.2.0.md for complete release notes."
    
    Write-Host "✅ Tag created successfully" -ForegroundColor Green
} catch {
    Write-Host "⚠️  Tag might already exist, continuing..." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "📤 Pushing tag to origin..." -ForegroundColor Cyan
try {
    git push origin v1.2.0
    Write-Host "✅ Tag pushed successfully" -ForegroundColor Green
} catch {
    Write-Host "❌ Error pushing tag: $_" -ForegroundColor Red
    exit 1
}

# Success message
Write-Host ""
Write-Host "=" * 61 -ForegroundColor Green
Write-Host "  ✅ v1.2.0 Release Pushed Successfully!" -ForegroundColor Green
Write-Host "=" * 61 -ForegroundColor Green
Write-Host ""
Write-Host "📋 Next Steps:" -ForegroundColor Cyan
Write-Host ""
Write-Host "1. Create GitHub Release:" -ForegroundColor White
Write-Host "   Go to: https://github.com/YOUR_USERNAME/source-fetcher/releases/new" -ForegroundColor Gray
Write-Host "   - Select tag: v1.2.0" -ForegroundColor Gray
Write-Host "   - Title: v1.2.0 - Cargo Build & Install Feature" -ForegroundColor Gray
Write-Host "   - Description: Copy from RELEASE_v1.2.0.md" -ForegroundColor Gray
Write-Host "   - Attach binaries from dist/ folder" -ForegroundColor Gray
Write-Host ""
Write-Host "2. Or use GitHub CLI:" -ForegroundColor White
Write-Host '   gh release create v1.2.0 \' -ForegroundColor Gray
Write-Host '     --title "v1.2.0 - Cargo Build & Install Feature" \' -ForegroundColor Gray
Write-Host '     --notes-file RELEASE_v1.2.0.md \' -ForegroundColor Gray
Write-Host '     dist/source-fetcher-windows-amd64.exe \' -ForegroundColor Gray
Write-Host '     dist/source-fetcher-windows-386.exe' -ForegroundColor Gray
Write-Host ""
Write-Host "3. Announce the release:" -ForegroundColor White
Write-Host "   - GitHub Discussions" -ForegroundColor Gray
Write-Host "   - Social media" -ForegroundColor Gray
Write-Host "   - Documentation site" -ForegroundColor Gray
Write-Host ""
Write-Host "📖 Documentation:" -ForegroundColor Cyan
Write-Host "   - See CREATE_GITHUB_RELEASE_v1.2.0.md for detailed instructions" -ForegroundColor Gray
Write-Host "   - See v1.2.0_RELEASE_READY.md for complete checklist" -ForegroundColor Gray
Write-Host ""
Write-Host "🎉 Great work! Time to celebrate! 🚀" -ForegroundColor Yellow
Write-Host ""
