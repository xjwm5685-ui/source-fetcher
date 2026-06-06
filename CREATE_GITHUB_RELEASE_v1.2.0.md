# Create GitHub Release v1.2.0 Guide

## 📋 Pre-Release Checklist

- [x] Version updated to 1.2.0 in `version.go`
- [x] Version updated to 1.2.0 in `main.go`
- [x] CHANGELOG.md updated with v1.2.0 changes
- [x] README.md updated with cargo build feature
- [x] New feature compiled and tested
- [x] Release notes created (`RELEASE_v1.2.0.md`)
- [ ] Compile release binaries
- [ ] Test release binaries
- [ ] Create git tag
- [ ] Push to GitHub
- [ ] Create GitHub Release

## 🔨 Step 1: Compile Release Binaries

### Windows AMD64
```powershell
$env:GOOS="windows"
$env:GOARCH="amd64"
go build -ldflags "-s -w -X main.BuildTime=$(Get-Date -Format 'yyyy-MM-dd_HH:mm:ss') -X main.GitCommit=$(git rev-parse --short HEAD)" -o dist/source-fetcher-windows-amd64.exe
```

### Windows 386
```powershell
$env:GOOS="windows"
$env:GOARCH="386"
go build -ldflags "-s -w -X main.BuildTime=$(Get-Date -Format 'yyyy-MM-dd_HH:mm:ss') -X main.GitCommit=$(git rev-parse --short HEAD)" -o dist/source-fetcher-windows-386.exe
```

### Linux AMD64
```powershell
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -ldflags "-s -w -X main.BuildTime=$(Get-Date -Format 'yyyy-MM-dd_HH:mm:ss') -X main.GitCommit=$(git rev-parse --short HEAD)" -o dist/source-fetcher-linux-amd64
```

### macOS AMD64
```powershell
$env:GOOS="darwin"
$env:GOARCH="amd64"
go build -ldflags "-s -w -X main.BuildTime=$(Get-Date -Format 'yyyy-MM-dd_HH:mm:ss') -X main.GitCommit=$(git rev-parse --short HEAD)" -o dist/source-fetcher-darwin-amd64
```

### macOS ARM64
```powershell
$env:GOOS="darwin"
$env:GOARCH="arm64"
go build -ldflags "-s -w -X main.BuildTime=$(Get-Date -Format 'yyyy-MM-dd_HH:mm:ss') -X main.GitCommit=$(git rev-parse --short HEAD)" -o dist/source-fetcher-darwin-arm64
```

## ✅ Step 2: Test Binaries

```powershell
# Test Windows AMD64
.\dist\source-fetcher-windows-amd64.exe version

# Test help
.\dist\source-fetcher-windows-amd64.exe install --help | Select-String cargo

# Test plan mode
.\dist\source-fetcher-windows-amd64.exe install --source cargo --name bat --plan
```

## 📦 Step 3: Create ZIP Archives (Optional)

```powershell
# Windows AMD64
Compress-Archive -Path dist\source-fetcher-windows-amd64.exe -DestinationPath dist\source-fetcher-v1.2.0-windows-amd64.zip

# Windows 386
Compress-Archive -Path dist\source-fetcher-windows-386.exe -DestinationPath dist\source-fetcher-v1.2.0-windows-386.zip
```

## 🏷️ Step 4: Create Git Tag

```powershell
# Create annotated tag
git tag -a v1.2.0 -m "Release v1.2.0 - Cargo Build & Install Feature

Major Features:
- Cargo crate automatic compilation
- Cargo binary system installation
- Three installation modes: source/build/install
- Cross-platform support

See RELEASE_v1.2.0.md for details."

# Verify tag
git tag -l -n9 v1.2.0

# Push tag
git push origin v1.2.0
```

## 📤 Step 5: Commit and Push

```powershell
# Stage all changes
git add .

# Commit
git commit -m "Release v1.2.0 - Cargo Build & Install Feature

- Added cargo build and install functionality
- Updated documentation
- Updated version to 1.2.0
- Created release notes"

# Push to GitHub
git push origin main
```

## 🌐 Step 6: Create GitHub Release

### Option A: Using GitHub Web UI

1. Go to: https://github.com/xjwm5685-ui/source-fetcher/releases/new
2. **Tag**: Select `v1.2.0`
3. **Title**: `v1.2.0 - Cargo Build & Install Feature`
4. **Description**: Copy content from below
5. **Attachments**: Upload binaries from `dist/` folder
6. Click "Publish release"

### Option B: Using GitHub CLI

```powershell
gh release create v1.2.0 `
  --title "v1.2.0 - Cargo Build & Install Feature" `
  --notes-file RELEASE_v1.2.0.md `
  dist/source-fetcher-windows-amd64.exe `
  dist/source-fetcher-windows-386.exe
```

## 📝 GitHub Release Description

```markdown
# 🦀 Cargo Build & Install Feature

Version 1.2.0 introduces automatic compilation and installation for Rust crates!

## 🌟 Three Installation Modes

### 1️⃣ Source Mode (Default) - No Rust Required
```powershell
sfer install --source cargo --name ripgrep
```

### 2️⃣ Build Mode - Requires Rust
```powershell
sfer install --source cargo --name ripgrep --cargo-build
```

### 3️⃣ Install Mode - Requires Rust
```powershell
sfer install --source cargo --name ripgrep --cargo-install
```

## ⚡ Quick Examples

```powershell
# Install popular Rust tools
sfer install --source cargo --name ripgrep --cargo-install
sfer install --source cargo --name bat --cargo-install
sfer install --source cargo --name fd-find --cargo-install
```

## 📖 Documentation

- [Complete Release Notes](https://github.com/xjwm5685-ui/source-fetcher/blob/main/RELEASE_v1.2.0.md)
- [User Guide](https://github.com/xjwm5685-ui/source-fetcher/blob/main/CARGO_BUILD_FEATURE.md)
- [Technical Implementation](https://github.com/xjwm5685-ui/source-fetcher/blob/main/CARGO_BUILD_IMPLEMENTATION.md)

## 🚀 Installation

### One-Line Install (Windows)
```powershell
irm https://raw.githubusercontent.com/xjwm5685-ui/source-fetcher/main/install.ps1 | iex
```

### Manual Download
Download the appropriate binary for your platform below.

## 📦 Assets

- `source-fetcher-windows-amd64.exe` - Windows 64-bit
- `source-fetcher-windows-386.exe` - Windows 32-bit

## 🔄 Upgrade from v1.1.0

Simply download and replace the executable, or use the one-line installer:

```powershell
irm https://raw.githubusercontent.com/xjwm5685-ui/source-fetcher/main/install.ps1 | iex
sfer version  # Should show: 1.2.0
```

## 📋 Full Changelog

### Added
- ⭐ Cargo build and install feature with three modes
- ⭐ New CLI flags: `--cargo-build`, `--cargo-install`, `--cargo-bin`
- ⭐ Intelligent error handling with manual compilation instructions
- ⭐ Cross-platform compilation support
- ⭐ PATH configuration detection and warnings
- 📚 Comprehensive documentation (3 new guides)

### Technical
- New `cargo_build.go` module
- Updated `install_native.go` with optional compilation
- Added YAML configuration support for cargo build

See [CHANGELOG.md](https://github.com/xjwm5685-ui/source-fetcher/blob/main/CHANGELOG.md) for details.

## 🙏 Credits

Thanks to all contributors and the Rust community!

---

**Full Changelog**: https://github.com/xjwm5685-ui/source-fetcher/compare/v1.1.0...v1.2.0
```

## 🎯 Post-Release Tasks

- [ ] Announce on social media
- [ ] Update project homepage
- [ ] Notify users via GitHub Discussions
- [ ] Update documentation website (if exists)
- [ ] Close milestone v1.2.0
- [ ] Create milestone v1.3.0

## 📊 Release Statistics

Collect after release:
- Download counts
- Star increase
- Issue reports
- Feature requests
- User feedback

---

**Ready to release?** Follow the steps above in order!
