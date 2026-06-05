# Source Fetcher Architecture

This document describes the architecture and design of Source Fetcher.

## 🏗️ Overview

Source Fetcher is a unified package download and management tool written in Go. It provides a consistent interface for interacting with multiple package ecosystems without requiring their native clients.

## 📐 Design Principles

1. **No Native Client Dependencies** - Direct API access to package repositories
2. **Unified Interface** - Consistent commands across all package sources
3. **Offline-First** - Download packages for offline installation
4. **Mirror Support** - Built-in mirror testing and fallback
5. **Extensibility** - Easy to add new package sources
6. **Performance** - Concurrent downloads and efficient caching

## 🎯 Core Components

### 1. CLI Layer

**Location:** `main.go`, command handlers

**Responsibilities:**
- Parse command-line arguments
- Validate user input
- Coordinate between components
- Format output for users

**Key Commands:**
- `mirrors` - Test mirror speeds
- `search` - Search packages across sources
- `download` - Download packages
- `install` - Install packages with dependencies
- `uninstall` - Remove installed packages
- `repair` - Fix damaged installations
- `batch` - Execute batch operations
- `tui` - Launch terminal UI

### 2. Provider System

**Location:** `providers.go`, provider implementations

**Responsibilities:**
- Abstract package source differences
- Implement source-specific logic
- Handle API communication
- Parse package metadata

**Supported Providers:**
- **npm** - Node.js packages
- **pip** - Python packages
- **cargo** - Rust crates
- **maven** - Java artifacts
- **choco** - Chocolatey packages
- **winget** - Windows Package Manager
- **url** - Direct URL downloads

**Provider Interface:**
```go
type Provider interface {
    Search(query string, limit int) ([]Package, error)
    Resolve(name, version string) (*ResolvedPackage, error)
    Download(resolved *ResolvedPackage, output string) error
    Install(pkg *Package, options InstallOptions) error
}
```

### 3. Download Engine

**Location:** `download.go`, downloader implementations

**Responsibilities:**
- HTTP/HTTPS downloads
- Resume capability (Range requests)
- Concurrent chunked downloads
- Progress reporting
- Checksum verification

**Features:**
- Automatic retry with exponential backoff
- Bandwidth throttling (optional)
- Proxy support
- Custom headers and authentication

### 4. Dependency Resolver

**Location:** `install.go`, resolver implementations

**Responsibilities:**
- Parse dependency trees
- Resolve version constraints
- Detect conflicts
- Generate installation plans
- Handle peer dependencies

**npm Resolver:**
- Supports semver ranges
- Handles optional dependencies
- Manages peer dependencies
- Generates lockfiles
- Deduplicates packages

### 5. Mirror System

**Location:** `mirrors.go`

**Responsibilities:**
- Test mirror availability
- Measure response times
- Automatic fallback
- Custom mirror configuration

**Built-in Mirrors:**
- npm: npmjs, npmmirror, huaweicloud, tencent
- choco: chocolatey, nuget
- winget: github-api, github-raw, jsdelivr
- pip: pypi, tuna, aliyun
- cargo: crates.io, ustc, sjtu
- maven: central, aliyun, huawei

### 6. Authentication System

**Location:** `config.go`, auth handlers

**Responsibilities:**
- Manage authentication profiles
- Support multiple auth methods
- Secure credential storage
- Environment variable integration

**Supported Methods:**
- Bearer tokens
- Basic authentication
- Custom headers
- API keys

### 7. Cache System

**Location:** `.source-fetcher/` directory

**Structure:**
```
.source-fetcher/
├── tarballs/          # Downloaded package archives
├── store/             # Extracted package contents
└── metadata/          # Cached metadata
```

**Responsibilities:**
- Cache downloaded packages
- Store extracted contents
- Metadata caching
- Cache invalidation

### 8. TUI (Terminal User Interface)

**Location:** `tui.go`

**Responsibilities:**
- Interactive package management
- Visual feedback
- Multi-page navigation
- Queue management

**Pages:**
- Search - Find and select packages
- Queue - Manage download/install tasks
- Batch - Load and execute YAML tasks
- Logs - View operation logs

## 🔄 Data Flow

### Search Flow

```
User Input → CLI Parser → Provider.Search() → API Request → Parse Results → Format Output
```

### Download Flow

```
User Input → CLI Parser → Provider.Resolve() → Get Download URL → 
Download Engine → Progress Reporting → Save to Disk → Verify Checksum
```

### Install Flow (npm)

```
User Input → CLI Parser → Provider.Install() → Resolve Dependencies →
Download Tarballs → Extract to Store → Link to node_modules →
Generate .bin Shims → Write Manifest → Execute Lifecycle Scripts
```

## 🗂️ File Structure

```
source-fetcher/
├── main.go                 # Entry point, CLI setup
├── providers.go            # Provider interface and registry
├── npm_provider.go         # npm implementation
├── choco_provider.go       # Chocolatey implementation
├── winget_provider.go      # winget implementation
├── pip_provider.go         # pip implementation
├── cargo_provider.go       # cargo implementation
├── maven_provider.go       # maven implementation
├── url_provider.go         # URL download implementation
├── download.go             # Download engine
├── install.go              # Installation logic
├── mirrors.go              # Mirror testing
├── config.go               # Configuration management
├── tui.go                  # Terminal UI
├── main_test.go            # Tests
└── install_test.go         # Installation tests
```

## 🔌 Extension Points

### Adding a New Provider

1. Implement the `Provider` interface
2. Register in provider registry
3. Add mirror configurations
4. Implement search logic
5. Implement download logic
6. Add tests

Example:
```go
type MyProvider struct {
    mirror string
    client *http.Client
}

func (p *MyProvider) Search(query string, limit int) ([]Package, error) {
    // Implementation
}

func (p *MyProvider) Resolve(name, version string) (*ResolvedPackage, error) {
    // Implementation
}

func (p *MyProvider) Download(resolved *ResolvedPackage, output string) error {
    // Implementation
}
```

### Adding a New Mirror

1. Add mirror configuration to `mirrors.go`
2. Implement mirror-specific URL transformation
3. Add to mirror testing
4. Update documentation

## 🔒 Security Considerations

### Authentication
- Credentials stored in environment variables
- No plaintext passwords in config files
- Support for token-based auth

### Downloads
- HTTPS by default
- Checksum verification
- Signature validation (planned)

### Script Execution
- Lifecycle scripts disabled by default
- Whitelist-based script execution
- Sandboxing (planned)

### Input Validation
- Sanitize user input
- Validate version strings
- Check file paths

## 📊 Performance Optimizations

### Concurrent Operations
- Parallel dependency resolution
- Concurrent chunk downloads
- Batch operations

### Caching
- Metadata caching
- Package caching
- HTTP response caching

### Network
- Connection pooling
- Keep-alive connections
- Compression support

## 🧪 Testing Strategy

### Unit Tests
- Provider implementations
- Dependency resolver
- Download engine
- Mirror system

### Integration Tests
- End-to-end workflows
- Multi-provider scenarios
- Error handling

### Manual Testing
- TUI functionality
- Cross-platform compatibility
- Real-world package installations

## 🔮 Future Architecture

### Plugin System
- Dynamic provider loading
- Custom provider plugins
- Hook system for extensions

### Distributed Caching
- Shared cache across machines
- P2P package distribution
- CDN integration

### Advanced Dependency Resolution
- SAT solver for conflicts
- Machine learning for optimization
- Predictive caching

## 📚 Related Documents

- [CONTRIBUTING.md](CONTRIBUTING.md) - Development guidelines
- [ROADMAP.md](ROADMAP.md) - Future plans
- [README.md](README.md) - User documentation

---

**Last Updated:** May 31, 2026
