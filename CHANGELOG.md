# Changelog

All notable changes to Source Fetcher will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **⭐ Cargo build and install feature**: Automatic compilation and installation of Rust crates
  - Three modes: source (default), build (`--cargo-build`), and install (`--cargo-install`)
  - Source mode: Download and extract .crate source without Rust toolchain (v1.1.0 feature)
  - Build mode: Automatically compile binaries with Rust toolchain
  - Install mode: Compile and install to system (`~/.cargo/bin/`)
  - Support for specifying binary name with `--cargo-bin` flag
  - Intelligent error handling with manual compilation instructions
  - Cross-platform support (Windows/Unix)
  - PATH configuration detection and warnings

### Technical Details
- New `cargo_build.go` module with core compilation logic
- `checkCargoAvailable()` - Check Rust toolchain availability
- `buildCargoCrate()` - Execute cargo build with options
- `installCargoBinary()` - Install compiled binary to system
- `getCargoInstallDir()` - Get installation directory (~/.cargo/bin)
- `findExecutables()` - Find compiled binaries
- Updated `install_native.go` with optional compilation support
- Added CLI flags: `--cargo-build`, `--cargo-install`, `--cargo-bin`
- YAML configuration support for cargo build options

### Documentation
- Added `CARGO_BUILD_FEATURE.md` - Comprehensive user guide for cargo build feature
- Added `CARGO_BUILD_IMPLEMENTATION.md` - Technical implementation summary
- Added `CARGO_BUILD_SUMMARY.md` - Quick reference guide

<!-- Future changes will be documented here -->

## [1.1.0] - 2026-06-06

### Added
- **⭐ Cargo install support**: Download and extract .crate source packages without Rust toolchain required
- **⭐ One-line installation script** (`install.ps1`) for quick setup, similar to popular tools like rustup
- **⭐ Web GUI support** for cargo/choco/winget package installation (previously only npm was supported)
- Local installation script (`install-local.ps1`) for offline scenarios and development
- Environment variable refresh script (`refresh-env.ps1`) for PATH updates
- Comprehensive documentation system with 11+ detailed guides
- Smart package source detection and filtering in Web GUI
- Friendly error messages in Web GUI queue display

### Improved
- Web GUI now supports multiple package sources (npm, cargo, choco, winget) for installation
- Backend `executeInstall()` uses switch statement to handle different package sources
- Frontend `handleInstall()` checks supported sources before installation
- Error messages are more user-friendly and informative

### Documentation
- Added `CARGO_INSTALL_GUIDE.md` - Comprehensive Cargo installation user guide
- Added `CARGO_FEATURE_SUMMARY.md` - Technical implementation summary of Cargo feature
- Added `WEBGUI_CARGO_SUPPORT.md` - Web GUI multi-source support documentation
- Added `INSTALLATION.md` - Complete installation guide with multiple methods
- Added `QUICK_INSTALL.md` - Quick start guide for impatient users
- Added `ONE_LINE_INSTALL_SUMMARY.md` - Implementation summary of one-line installation
- Added `POST_INSTALL_GUIDE.md` - Post-installation guide and troubleshooting
- Added `PROJECT_STATUS.md` - Comprehensive project status report
- Added `NEXT_STEPS.md` - Detailed next steps and development guide
- Added `CONTINUATION_SUMMARY.md` - Conversation continuation summary
- Added `DOCUMENTS_CREATED.md` - Documentation index and quick reference

## [1.0.1] - 2026-06-06

### Security
- Fixed CORS vulnerability allowing arbitrary origins
- Added package name validation to prevent command injection
- Implemented SSRF protection for URL dependencies
- Hardened file permissions for sensitive data

### Fixed
- winget uninstall now correctly reports actual failures
- Web GUI startup reliability improved with health checks
- Context cancellation support for long-running uninstall operations

### Changed
- YAML parsing explicitly uses safe decoder
- Manifest loading includes version compatibility check

## [1.0.0] - 2026-05-31

### Added
- Initial release of Source Fetcher
- Multi-source package manager support (npm, pip, cargo, maven, choco, winget)
- Mirror speed testing for all supported sources
- Package search across multiple sources
- Direct package download without installing native clients
- TUI (Terminal User Interface) for interactive package management
- npm dependency installation with full dependency tree resolution
- npm uninstall and repair functionality
- Choco auto-install feature
- Winget auto-install feature
- YAML batch download and installation tasks
- Private source authentication support
- Resume download capability
- Concurrent chunked download
- Install lockfile support
- Frozen lockfile mode
- Lifecycle scripts control (none/root/all)
- Global alias installation (`sfer` command)

### Features by Source

#### npm
- Search packages
- Download tarballs
- Install with dependency resolution
- Uninstall with manifest tracking
- Repair damaged installations
- Support for dependencies, optionalDependencies, peerDependencies, devDependencies
- Lockfile generation and frozen mode
- Lifecycle scripts with whitelist support

#### pip
- Search packages
- Download from PyPI
- Mirror support

#### cargo
- Search crates
- Download from crates.io
- Mirror support

#### maven
- Search artifacts
- Download JARs
- Mirror support

#### choco
- Search packages
- Download .nupkg files
- Auto-install with Chocolatey client

#### winget
- Search packages
- Parse manifests from winget-pkgs repository
- Download installers
- Auto-install with silent parameters
- Architecture selection
- Installer type detection

#### url
- Direct URL download
- Resume support
- Chunked download

### Documentation
- Comprehensive README in English and Chinese
- Quick Start Guide
- Auto-Install Guide
- Implementation documentation
- Contributing guidelines
- MIT License

### Infrastructure
- Go 1.21+ support
- Cross-platform compatibility (Windows primary)
- Comprehensive test coverage
- PowerShell installation script

## [1.0.0] - TBD

### Planned
- First stable release
- GitHub Actions CI/CD
- Pre-built binaries for releases
- Enhanced documentation
- Community feedback integration

---

## Version History

### Versioning Scheme

We use [Semantic Versioning](https://semver.org/):
- **MAJOR** version for incompatible API changes
- **MINOR** version for new functionality in a backward compatible manner
- **PATCH** version for backward compatible bug fixes

### Release Notes

Release notes for each version will include:
- New features
- Bug fixes
- Breaking changes
- Deprecations
- Security updates
- Performance improvements

---

For more details, see the [commit history](https://github.com/jiahe/source-fetcher/commits/main).
