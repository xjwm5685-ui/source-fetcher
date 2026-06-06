# Quick Reference: Cargo Build & Install (v1.2.0)

## 🚀 Quick Start

### Source Mode (Default - No Rust)
```powershell
sfer install --source cargo --name ripgrep
```

### Build Mode (Requires Rust)
```powershell
sfer install --source cargo --name ripgrep --cargo-build
```

### Install Mode (Requires Rust)
```powershell
sfer install --source cargo --name ripgrep --cargo-install
```

---

## 📋 New CLI Flags

| Flag | Description |
|------|-------------|
| `--cargo-build` | Compile binary from source |
| `--cargo-install` | Compile and install to system |
| `--cargo-bin <name>` | Specify binary name |

---

## 🔥 Popular Tools

```powershell
# Search Tools
sfer install --source cargo --name ripgrep --cargo-install
sfer install --source cargo --name fd-find --cargo-install

# File Viewers
sfer install --source cargo --name bat --cargo-install
sfer install --source cargo --name hexyl --cargo-install

# System Tools
sfer install --source cargo --name bottom --cargo-install
sfer install --source cargo --name procs --cargo-install
```

---

## 🛠️ Troubleshooting

### Cargo not found
```powershell
winget install Rustlang.Rustup
```

### Command not available
```powershell
$env:Path += ";$env:USERPROFILE\.cargo\bin"
```

---

## 📖 Documentation

- **User Guide**: `CARGO_BUILD_FEATURE.md`
- **Technical**: `CARGO_BUILD_IMPLEMENTATION.md`
- **Release Notes**: `RELEASE_v1.2.0.md`

---

## 🎯 Three Modes Comparison

| Feature | Source | Build | Install |
|---------|--------|-------|---------|
| Downloads source | ✅ | ✅ | ✅ |
| Compiles binary | ❌ | ✅ | ✅ |
| Installs to system | ❌ | ❌ | ✅ |
| Needs Rust | ❌ | ✅ | ✅ |

---

**Version**: 1.2.0  
**Date**: 2026-06-06  
**Status**: ✅ Production Ready
