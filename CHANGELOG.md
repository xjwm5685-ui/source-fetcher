# Changelog

All notable changes to Source Fetcher will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Configuration file support (`.source-fetcher.yaml`)
- Global configuration example (`.source-fetcher.example.yaml`)
- Unified uninstall system for choco/winget/url (v1.1.0)
- URL dependency tracking and installation (v1.1.0)
- Context propagation for cancellable uninstall operations
- Health check mechanism for Web GUI startup
- Package name input validation (security)
- SSRF protection for URL dependencies (security)
- Manifest version compatibility checking

### Changed
- CORS configuration now restricts to localhost origins only (security fix)
- YAML parsing now uses explicit safe unmarshaling
- File permissions hardened (0644 → 0600/0640) for sensitive files
- Web GUI startup uses health check instead of fixed delay
- winget uninstall error handling now distinguishes benign exit codes

### Security
- **CRITICAL**: Fixed CORS configuration to prevent CSRF attacks
- **CRITICAL**: Added input validation to prevent command injection
- **CRITICAL**: Implemented SSRF protection blocking private IP ranges
- Hardened file permissions for manifest and tracking files

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
