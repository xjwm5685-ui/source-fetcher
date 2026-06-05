# Source Fetcher Examples

This directory contains practical examples of using Source Fetcher in various scenarios.

## 📂 Examples

### 1. Basic Usage
- [basic-download.yaml](basic-download.yaml) - Simple package downloads
- [basic-install.yaml](basic-install.yaml) - Basic npm installation

### 2. Enterprise Scenarios
- [offline-deployment.yaml](offline-deployment.yaml) - Offline package deployment
- [private-registry.yaml](private-registry.yaml) - Using private npm registries
- [ci-cd-integration.yaml](ci-cd-integration.yaml) - CI/CD pipeline integration

### 3. Advanced Features
- [batch-operations.yaml](batch-operations.yaml) - Batch download and install
- [mirror-optimization.yaml](mirror-optimization.yaml) - Mirror selection and failover
- [multi-source.yaml](multi-source.yaml) - Working with multiple package sources

### 4. Platform-Specific
- [windows-tools.yaml](windows-tools.yaml) - Windows software installation
- [dev-environment.yaml](dev-environment.yaml) - Development environment setup

## 🚀 Quick Start

### Run an example:

```powershell
# Using the alias
sfer batch --config examples/basic-download.yaml

# Or directly
.\source-fetcher.exe batch --config examples/basic-download.yaml
```

### Customize an example:

1. Copy an example file
2. Modify for your needs
3. Run with source-fetcher

## 📖 Learn More

- [QUICK_START.md](../QUICK_START.md) - Getting started guide
- [README.md](../README.md) - Full documentation
- [CONTRIBUTING.md](../CONTRIBUTING.md) - How to contribute examples

## 💡 Contributing Examples

Have a useful example? Contributions are welcome!

1. Create a new YAML file with a descriptive name
2. Add comments explaining the configuration
3. Test it thoroughly
4. Submit a pull request
5. Update this README

## ⚠️ Notes

- Replace placeholder values (URLs, tokens, paths) with your actual values
- Some examples require authentication - see comments in files
- Examples are tested on Windows - adjust paths for other platforms
