# Feature Enhancement Plan - v1.1.0

**Date**: 2026-06-06  
**Target Version**: v1.1.0  
**Status**: Planning

---

## 🎯 Overview

This document outlines the plan to implement missing features mentioned in the README:
1. **Dependency installation for choco/winget/url sources**
2. **Cross-ecosystem dependency uninstall**

---

## 📋 Current Status

### ✅ Already Implemented (v1.0.0)

**Download Features**:
- ✅ Multi-source downloads (npm, pip, cargo, maven, choco, winget, url)
- ✅ Mirror speed testing and auto-failover
- ✅ Resume and chunked downloads
- ✅ Batch operations with YAML

**Install Features**:
- ✅ npm full dependency management (install/uninstall/repair)
- ✅ choco auto-install (basic, no dependency resolution)
- ✅ winget auto-install (basic, no dependency resolution)
- ✅ Private registry authentication
- ✅ Lockfile support

**Missing Features**:
- ❌ choco dependency chain installation
- ❌ winget dependency chain installation  
- ❌ url source dependency tracking
- ❌ Cross-ecosystem unified uninstall

---

## 🚀 Feature 1: Dependency Installation for Native Sources

### 1.1 Choco Dependency Installation

**Current State**:
- Can download `.nupkg` files
- Can invoke `choco install` for single package
- ❌ Does NOT resolve or install dependencies

**Enhancement Plan**:

#### Option A: Query Chocolatey API for Dependencies
```go
// Fetch package metadata from Chocolatey API
// Parse dependencies from NuSpec
// Recursively resolve dependency tree
// Download all packages
// Install in correct order
```

**Pros**:
- Full control over process
- Can create install manifest
- Can implement our own dependency resolver

**Cons**:
- Need to parse NuSpec XML
- Complex dependency resolution
- May not match choco's exact behavior

#### Option B: Let Chocolatey Handle Dependencies
```go
// Download main package
// Run: choco install package.nupkg --yes
// Chocolatey automatically installs dependencies
// Parse choco output to track what was installed
```

**Pros**:
- Simpler implementation
- Matches choco's behavior exactly
- Less maintenance burden

**Cons**:
- Less control
- Requires choco to be installed
- Harder to create precise manifest

**Recommendation**: **Option B** for v1.1.0
- Simpler and more reliable
- Can enhance to Option A in v1.2.0 if needed

### 1.2 Winget Dependency Installation

**Current State**:
- Can download installer files (.exe/.msi/.msix)
- Can invoke installer with silent parameters
- ❌ Does NOT resolve or install dependencies

**Enhancement Plan**:

#### Option A: Parse Winget Manifest Dependencies
```go
// Fetch full winget manifest from GitHub
// Parse Dependencies field
// Recursively resolve dependency tree
// Download and install all in order
```

**Pros**:
- Can track what's being installed
- Can create install manifest

**Cons**:
- Winget manifests don't always list all dependencies
- Windows installers often handle their own dependencies
- Complex to implement correctly

#### Option B: Let Windows Installer Handle Dependencies
```go
// Download main installer
// Run installer with silent parameters
// Installer handles its own dependencies
// Track main package only
```

**Pros**:
- Simpler implementation
- Matches actual install behavior
- Most Windows installers are self-contained

**Cons**:
- Can't track dependencies
- Can't uninstall dependencies separately

**Recommendation**: **Option B** for v1.1.0
- Most practical for Windows ecosystem
- Installers are designed to be self-contained
- Can document that dependency tracking is limited

### 1.3 URL Source Dependency Tracking

**Current State**:
- Can download any URL
- ❌ No concept of dependencies

**Enhancement Options**:

#### Option A: Manual Dependency Declaration
```yaml
downloads:
  - source: url
    url: https://example.com/app-v1.0.exe
    dependencies:
      - url: https://example.com/runtime.exe
      - url: https://example.com/lib.dll
```

#### Option B: Companion Metadata File
```yaml
# app-v1.0.deps.yaml
dependencies:
  - url: https://example.com/runtime.exe
    version: "2.0"
  - url: https://example.com/lib.dll
    version: "1.5"
```

**Recommendation**: **Option A** for v1.1.0
- Simpler to implement
- User has full control
- Fits existing YAML structure

---

## 🗑️ Feature 2: Cross-Ecosystem Unified Uninstall

### 2.1 Current Uninstall Support

**npm**:
- ✅ Full uninstall support
- ✅ Reads install manifest
- ✅ Removes packages, bins, and cache

**choco**:
- ❌ No uninstall support
- ❌ No install tracking

**winget**:
- ❌ No uninstall support
- ❌ No install tracking

**url**:
- ❌ No concept of uninstall

### 2.2 Unified Uninstall Design

#### Enhanced Install Manifest

```json
{
  "version": 1,
  "generated_at": "2026-06-06T10:00:00Z",
  "installations": [
    {
      "id": "install-1",
      "source": "npm",
      "package": "react",
      "version": "18.2.0",
      "install_paths": [...],
      "uninstall_method": "manifest"
    },
    {
      "id": "install-2",
      "source": "choco",
      "package": "git",
      "version": "2.47.0",
      "install_paths": ["C:\\Program Files\\Git"],
      "uninstall_method": "choco_uninstall",
      "uninstall_command": "choco uninstall git --yes"
    },
    {
      "id": "install-3",
      "source": "winget",
      "package": "Microsoft.PowerToys",
      "version": "0.75.0",
      "install_paths": ["C:\\Program Files\\PowerToys"],
      "uninstall_method": "windows_uninstaller",
      "uninstall_guid": "{GUID}"
    },
    {
      "id": "install-4",
      "source": "url",
      "url": "https://example.com/app.exe",
      "install_paths": ["C:\\MyApp"],
      "uninstall_method": "manual_paths"
    }
  ]
}
```

#### Uninstall Methods

**1. NPM Uninstall** (existing):
```bash
sfer uninstall --source npm --output ./workspace
```

**2. Choco Uninstall** (new):
```bash
sfer uninstall --source choco --name git
# Runs: choco uninstall git --yes
```

**3. Winget Uninstall** (new):
```bash
sfer uninstall --source winget --name Microsoft.PowerToys
# Runs: winget uninstall Microsoft.PowerToys
```

**4. Unified Uninstall** (new):
```bash
# Uninstall everything tracked in manifest
sfer uninstall --all --manifest ./source-fetcher-global.json

# Uninstall specific ecosystem
sfer uninstall --all --source choco --manifest ./source-fetcher-global.json

# Interactive mode
sfer uninstall --interactive
```

### 2.3 Implementation Plan

#### Phase 1: Enhanced Manifest (Week 1)
- [ ] Design unified manifest format
- [ ] Update install commands to write unified manifest
- [ ] Add install tracking for choco/winget

#### Phase 2: Individual Uninstallers (Week 2)
- [ ] Implement choco uninstall
- [ ] Implement winget uninstall
- [ ] Implement url path cleanup

#### Phase 3: Unified Uninstall (Week 3)
- [ ] Implement `--all` flag
- [ ] Implement ecosystem filtering
- [ ] Implement interactive mode
- [ ] Add dry-run/preview mode

---

## 📅 Development Timeline

### v1.1.0 - Basic Dependency Support (Target: Q3 2026)

**Week 1-2: Choco/Winget Dependency Handling**
- [ ] Implement choco dependency installation (let choco handle)
- [ ] Implement winget dependency handling (installer handles)
- [ ] Add dependency tracking to install output
- [ ] Update documentation

**Week 3-4: URL Dependencies**
- [ ] Implement URL dependency declaration in YAML
- [ ] Add URL dependency download and tracking
- [ ] Update batch command to handle URL dependencies
- [ ] Add examples

**Week 5-6: Cross-Ecosystem Uninstall**
- [ ] Design and implement unified manifest
- [ ] Implement choco uninstall
- [ ] Implement winget uninstall
- [ ] Implement unified `--all` uninstall

**Week 7-8: Testing and Documentation**
- [ ] Write comprehensive tests
- [ ] Update all documentation
- [ ] Create migration guide from v1.0.0
- [ ] Prepare release notes

### v1.2.0 - Advanced Features (Target: Q4 2026)

- [ ] Advanced choco dependency resolver (Optional)
- [ ] Dependency conflict detection
- [ ] Dependency version constraints
- [ ] Rollback support

---

## 🧪 Testing Strategy

### Unit Tests
- Choco dependency parsing
- Winget manifest parsing
- URL dependency resolution
- Uninstall manifest operations

### Integration Tests
- Full install → uninstall cycle
- Mixed ecosystem installations
- Dependency chain resolution
- Error handling and edge cases

### Manual Testing Checklist
- [ ] Install choco package with dependencies
- [ ] Install winget package with dependencies
- [ ] Install URL with declared dependencies
- [ ] Uninstall each ecosystem separately
- [ ] Uninstall all at once
- [ ] Verify clean removal (no leftover files)

---

## 📝 Documentation Updates

### README.md
- Update "Current Scope" section
- Remove "Not covered" mentions
- Add new feature descriptions
- Update examples

### New Documents
- [ ] DEPENDENCY_MANAGEMENT.md - How dependency resolution works
- [ ] UNINSTALL_GUIDE.md - Complete uninstall documentation
- [ ] MANIFEST_SPEC.md - Unified manifest specification

### Updated Examples
- [ ] `examples/choco-with-deps.yaml`
- [ ] `examples/url-dependencies.yaml`
- [ ] `examples/uninstall-all.yaml`

---

## 🚨 Breaking Changes

None planned. v1.1.0 will be backward compatible with v1.0.0.

Existing v1.0.0 install manifests will continue to work.

---

## 💡 Future Considerations (v1.2+)

### Advanced Dependency Features
- Dependency conflict resolution
- Version constraint solving
- Dependency graphs visualization
- Dependency update recommendations

### Enhanced Uninstall
- Orphan detection (unused dependencies)
- Cascade uninstall (with dependents)
- Partial uninstall (keep dependencies)
- Uninstall simulation mode

### Cross-Platform Support
- Linux package managers (apt, yum, dnf)
- macOS package managers (homebrew, macports)
- Universal package format support

---

## 📊 Success Metrics

### For v1.1.0 Release

**Functionality**:
- ✅ Choco packages install with dependencies
- ✅ Winget packages install correctly
- ✅ URL dependencies are tracked and installed
- ✅ All ecosystems can be uninstalled cleanly

**Quality**:
- ✅ 80%+ test coverage for new code
- ✅ Zero critical bugs
- ✅ Documentation complete and accurate

**User Experience**:
- ✅ Intuitive commands
- ✅ Clear error messages
- ✅ Good progress feedback

---

## 🤝 Contributing

This enhancement is tracked in GitHub Issues:
- Issue #1: Choco/Winget dependency installation
- Issue #2: Cross-ecosystem uninstall

Contributors welcome! See CONTRIBUTING.md for guidelines.

---

## 📞 Questions?

- Open an issue on GitHub
- Start a discussion
- Email: ckkhua89@gmail.com

---

**Last Updated**: 2026-06-06  
**Status**: 📋 Planning Phase  
**Target Release**: v1.1.0 (Q3 2026)
