# Source Fetcher

<div align="center">

![Source Fetcher Logo](./assets/logo.png)

**Unified Package Download Tool - No Native Clients Required**

English | [中文](./README.md)

[![Build Status](https://img.shields.io/github/actions/workflow/status/jiahe/source-fetcher/test.yml?branch=main)](https://github.com/jiahe/source-fetcher/actions)
[![Go Version](https://img.shields.io/github/go-mod/go-version/jiahe/source-fetcher)](https://go.dev/)
[![License](https://img.shields.io/github/license/jiahe/source-fetcher)](LICENSE)
[![Release](https://img.shields.io/github/v/release/jiahe/source-fetcher)](https://github.com/jiahe/source-fetcher/releases)
[![Downloads](https://img.shields.io/github/downloads/jiahe/source-fetcher/total)](https://github.com/jiahe/source-fetcher/releases)
[![Stars](https://img.shields.io/github/stars/jiahe/source-fetcher?style=social)](https://github.com/jiahe/source-fetcher/stargazers)

</div>

---

## ✨ Why Source Fetcher?

<table>
<tr>
<td width="33%" align="center">
<h3>🚀 No Client Installation</h3>
Direct API access to repositories, no need to install npm, choco, winget, etc.
</td>
<td width="33%" align="center">
<h3>🌐 Unified Interface</h3>
One tool for multiple package sources with consistent CLI experience
</td>
<td width="33%" align="center">
<h3>⚡ Mirror Acceleration</h3>
Built-in mirrors with automatic speed testing and failover
</td>
</tr>
<tr>
<td width="33%" align="center">
<h3>📦 Offline Friendly</h3>
Download first, install later - perfect for offline deployment
</td>
<td width="33%" align="center">
<h3>🔄 Resume Downloads</h3>
Support for resume and concurrent chunked downloads
</td>
<td width="33%" align="center">
<h3>🎯 Batch Operations</h3>
YAML-based batch download and installation
</td>
</tr>
</table>

## 🎬 Quick Demo

```bash
# Install global alias
.\install-alias.ps1

# Search packages
sfer search --source npm --query react

# Download packages
sfer download --source npm --name react --version latest

# Install dependencies
sfer install --source npm --name react --version ^19

# Batch operations
sfer batch --config source-fetcher.yaml
```

> 💡 **Tip**: Use `sfer tui` to launch the interactive interface for a better experience!

## Quick Installation (Recommended)

Install the global alias `sfer` to use it from any directory:

```powershell
# Run in the source-fetcher directory
.\install-alias.ps1

# After restarting the terminal, you can use:
sfer version
sfer mirrors --source npm
sfer search --source npm --query react
```

Uninstall the alias:

```powershell
.\install-alias.ps1 -Uninstall
```

## 📊 Comparison with Other Tools

| Feature | Source Fetcher | npm | choco | winget |
|---------|----------------|-----|-------|--------|
| No Client Required | ✅ | ❌ | ❌ | ❌ |
| Unified Multi-Source | ✅ | ❌ | ❌ | ❌ |
| Offline Download | ✅ | ⚠️ | ⚠️ | ❌ |
| Auto Mirror Switch | ✅ | ⚠️ | ❌ | ❌ |
| Resume Download | ✅ | ❌ | ❌ | ❌ |
| Concurrent Download | ✅ | ❌ | ❌ | ❌ |
| Batch Operations | ✅ | ⚠️ | ⚠️ | ❌ |
| TUI Interface | ✅ | ❌ | ❌ | ❌ |
| Private Registry | ✅ | ✅ | ✅ | ❌ |
| Cross-Platform | ✅ | ✅ | ❌ | ❌ |

## 💡 Use Cases

### 🏢 Enterprise Intranet Deployment
```bash
# Batch download on internet-connected machine
sfer batch --config packages.yaml --output ./offline-packages

# Offline install on intranet machine
sfer install --source npm --name react --output ./workspace
```

### 🌍 Network Acceleration
```bash
# Auto use mirrors with automatic failover
sfer download --source npm --name vue --version latest

# Manually specify mirror
sfer download --source npm --name vue --mirror npmmirror
```

### 🔧 CI/CD Integration
```yaml
# .github/workflows/build.yml
- name: Download dependencies
  run: sfer batch --config deps.yaml --continue-on-error
```

### 📦 Multi-Ecosystem Projects
```bash
# Manage all dependencies with one command
sfer search --source all --query typescript
sfer download --source npm --name typescript
sfer download --source choco --name nodejs
sfer download --source winget --name Microsoft.VisualStudioCode
```

## 🌟 Core Features

### 📥 Download Features
- ✅ Multi-source mirror speed testing
- ✅ Multi-source package search
- ✅ `npm` package download
- ✅ `Chocolatey/NuGet` package download
- ✅ `winget` manifest parsing and installer download
- ✅ `pip`, `cargo`, `maven` package download
- ✅ Direct URL download
- ✅ Resume and concurrent chunked downloads

### 📦 Installation Features
- ✅ `npm` dependency tree resolution and installation to `node_modules`
- ✅ `npm` uninstall based on installation manifest
- ✅ `npm` repair based on installation manifest
- ✅ `choco` auto-install ⭐ New Feature
- ✅ `winget` auto-install ⭐ New Feature

### 🔧 Advanced Features
- ✅ YAML batch download tasks
- ✅ Private source authentication
- ✅ Install lockfile and frozen mode
- ✅ Lifecycle scripts control
- ✅ TUI interactive interface

## Why You Don't Need to Install the Original Tools

- `npm`: Directly requests registry metadata and downloads tarballs
- `choco`: Directly requests NuGet/Chocolatey feed and downloads `.nupkg`
- `winget`: Directly reads `winget-pkgs` repository manifests and extracts `InstallerUrl`
- `url`: Built-in HTTP downloader, no dependency on `curl`

## Current Scope

This is a runnable MVP, not a complete package manager.

- Supported: Download, mirror testing, winget installer selection
- Supported: Single task resolution preview, YAML batch download and batch install
- Supported: `npm` dependency install/uninstall/repair MVP, recursively resolves `dependencies`/`optionalDependencies`/`peerDependencies`/`devDependencies` (controlled by parameters), assembles `node_modules`, generates `.bin`, and performs uninstall and repair based on installation manifest
- Supported: Private authenticated sources, resume download, concurrent chunked download
- Supported: Install lockfile, frozen lockfile, controlled lifecycle scripts (`none/root/all`)
- Supported: TUI enqueue for download/install/uninstall/repair
- Supported: Search, mirror testing, and direct download resolution for `pip`, `cargo`, `maven`
- Not covered: `choco/winget/url` dependency installation, cross-ecosystem dependency uninstall

Current provider coverage:

- `npm`: Search, download, install, uninstall, repair
- `pip`: Search, download
- `cargo`: Search, download
- `maven`: Search, download
- `choco`: Search, download, **auto-install** ⭐
- `winget`: Search, download, installer enumeration, **auto-install** ⭐
- `url`: Direct download

## Usage Methods

### Method 1: Using Global Alias (Recommended)

After installation, use the `sfer` command from any directory:

```powershell
sfer mirrors --source npm
sfer search --source npm --query react
```

### Method 2: Direct Execution

```powershell
cd D:\dy\source-fetcher
.\source-fetcher.exe mirrors --source npm
```

### Method 3: Development Mode

```powershell
cd D:\dy\source-fetcher
go run . mirrors --source npm
```

## Mirror Testing

```powershell
sfer mirrors --source all
sfer mirrors --source npm
sfer mirrors --source winget
```

Or use the full command:

```powershell
go run . mirrors --source all
go run . mirrors --source npm
go run . mirrors --source winget
```

## Search Examples

Using `sfer` alias (recommended):

```powershell
sfer search --source all --query powertoys
sfer search --source npm --query react --limit 5
sfer search --source pip --query requests --limit 5
sfer search --source cargo --query serde --limit 5
sfer search --source maven --query junit --limit 5
sfer search --source choco --query git --limit 10
sfer search --source winget --query terminal --limit 10
sfer search --source winget --query terminal --interactive --resolve-only
sfer search --source winget --query terminal --pick 2 --output .\downloads
sfer search --source winget --query terminal --pick 1,3 --resolve-only
sfer search --source npm --query internal-lib --config .\source-fetcher.yaml --auth-profile private-npm --pick 1 --resume --chunks 4 --output .\downloads
```

Or use the full command:

```powershell
go run . search --source all --query powertoys
go run . search --source npm --query react --limit 5
```

Notes: Search results come with `INDEX` numbers. Sources like `choco` that return multiple versions will automatically deduplicate by package identifier and keep the newer version.
Notes: Passing `--pick N` or `--pick 1,3,5` will directly pass the corresponding search results to the download process; can be used with `--resolve-only`, `--output`, `--arch`, `--installer-index`.
Notes: Passing `--interactive` will prompt you to enter one or more result numbers in the command line, suitable for quick search and download.
Notes: After entering actual download, progress, total size, and instantaneous speed will be automatically displayed in the terminal. When file size is unknown, downloaded bytes and speed will also be shown.
Notes: Passing `--resume` will reuse the same-named `.part` file to continue downloading; passing `--chunks N` will enable concurrent chunked download when the server supports Range.
Notes: Passing `--config` and `--auth-profile` can load authentication configuration from YAML, suitable for private npm/NuGet/API download sources.

## TUI

Using `sfer` alias:

```powershell
sfer tui
sfer tui --source winget --query terminal --output .\downloads
```

Or use the full command:

```powershell
go run . tui
go run . tui --source winget --query terminal --output .\downloads
```

Notes: TUI reuses existing search, resolution, download, and install pipelines, without depending on local `npm`, `choco`, or `winget` clients.
Notes: The search page supports direct editing of `arch` and `installer-index`, which will apply to `winget` resolution and download initiated within TUI.
Notes: The search page also supports `auth profile`, `resume`, and `chunks`, which will be carried over to download options after adding to queue.
Notes: `Enter` triggers search in the form area and resolution in the result area; `Space` for multi-select; `d` adds download tasks to queue, `i` adds npm install tasks to queue; in batch page, `e` adds selected tasks to queue.
Notes: The queue page supports `u` to add uninstall tasks, `p` to add repair tasks, and displays execution status of download/install/uninstall/repair.
Notes: The batch page reads default mirrors, output directory, `auth_profiles`, and `timeout` from YAML, and can load both `downloads` and `installs` tasks simultaneously.

## Download Examples

Download npm package:

```powershell
sfer download --source npm --name react --version latest --output .\downloads
sfer download --source npm --name react --config .\source-fetcher.yaml --auth-profile private-npm --resume --chunks 4 --output .\downloads
```

Download choco package:

```powershell
sfer download --source choco --name git --version 2.47.0 --output .\downloads
```

Download PyPI package:

```powershell
sfer download --source pip --name requests --version latest --output .\downloads
```

Download Cargo crate:

```powershell
sfer download --source cargo --name serde --version latest --output .\downloads
```

Download Maven artifact:

```powershell
sfer download --source maven --name junit:junit --version latest --output .\downloads
```

Download winget installer:

```powershell
sfer download --source winget --id Microsoft.PowerToys --arch x64 --output .\downloads
```

First check available winget installers:

```powershell
sfer download --source winget --id Microsoft.PowerToys --list-installers
```

Download any URL:

```powershell
sfer download --source url --url https://example.com/file.zip --output .\downloads
```

Resolve only, no download:

```powershell
sfer download --source winget --id Microsoft.PowerToys --resolve-only
```

## Dependency Installation

The `install` command supports **npm**, **choco**, and **winget** sources.

### npm Installation (Full Dependency Management)

Execute dependency installation:

```powershell
sfer install --source npm --name react --version ^19 --output .\workspace
sfer install --source npm --name react --version ^19 --include-peer --omit-optional --lockfile .\workspace\source-fetcher-install.lock.json
sfer install --source npm --name react --version ^19 --include-dev --scripts root
```

View dependency resolution plan only:

```powershell
sfer install --source npm --name react --version ^19 --plan
sfer install --source npm --name react --frozen-lockfile --plan
```

### choco Auto-Install ⭐ New Feature

```powershell
# Install latest version
sfer install --source choco --name curl

# Install specific version
sfer install --source choco --name git --version 2.47.0

# View installation plan
sfer install --source choco --name 7zip --plan
```

**Note**: Requires Chocolatey client installed, recommended to run with administrator privileges.

### winget Auto-Install ⭐ New Feature

```powershell
# Install latest version
sfer install --source winget --name Microsoft.PowerToys

# Install specific version
sfer install --source winget --name Microsoft.VisualStudioCode --version 1.85.0

# View installation plan
sfer install --source winget --name Microsoft.PowerToys --plan
```

**Note**: Some installers require administrator privileges.

Notes: `install` recursively resolves `dependencies`, `optionalDependencies`, `peerDependencies`, and root package `devDependencies` according to parameters, first downloads tarballs to `.source-fetcher\tarballs` cache under the output directory, then unpacks to `.source-fetcher\store`, and finally assembles to `node_modules` according to npm rules, generates `node_modules\.bin`, and writes out `source-fetcher-install.json` and `source-fetcher-install.lock.json`.
Notes: When passing `--frozen-lockfile`, it will prioritize requiring the lockfile to match the current request; if matched successfully, it will resolve directly according to the lockfile without requesting the registry.
Notes: Passing `--scripts none|root|all` controls whether to execute `preinstall/install/postinstall` lifecycle scripts, default is `none`.
Notes: Passing `--allow-scripts esbuild,sharp` allows whitelisted packages to run `preinstall` that is blocked by default; if using `install_defaults.allow_scripts` in `source-fetcher.yaml` simultaneously, CLI parameters will override YAML instead of merging.

Uninstall installed content:

```powershell
sfer uninstall --output .\workspace
```

Notes: `uninstall` reads `source-fetcher-install.json` from the output directory by default, only deletes installation paths, cache paths, and recyclable empty directories recorded in the manifest; passing `--manifest` can specify the manifest path, passing `--keep-cache` or `--keep-manifest` can keep corresponding content.

Repair installed content:

```powershell
sfer repair --output .\workspace
```

Notes: `repair` reads `source-fetcher-install.json` from the output directory by default, checks installation paths in the manifest one by one; if a directory is missing or damaged, it will prioritize reusing `.source-fetcher\store`, and restore from `.source-fetcher\tarballs` if necessary.

## Batch Tasks

Copy sample configuration:

```powershell
copy .\source-fetcher.sample.yaml .\source-fetcher.yaml
```

Execute batch download:

```powershell
sfer batch --config .\source-fetcher.yaml
sfer batch --config .\source-fetcher.yaml --jobs 2 --continue-on-error --retries 2 --retry-backoff 500ms
sfer batch --config .\source-fetcher.yaml --continue-on-error --retries 2 --retry-backoff 500ms
```

View resolution plan only:

```powershell
sfer batch --config .\source-fetcher.yaml --plan
```

Notes: `batch` now executes `downloads` and `installs` in YAML sequentially; when passing `--plan`, it will output download plans and install plans separately.
Notes: YAML can persist `scripts_policy` and `allow_scripts` through top-level `install_defaults`, used for batch install tasks and default script policy for `install --config ...`.
Notes: Unknown fields in the configuration file will output `Warning` to `stderr`, but will not block loading of known fields; field type errors will still fail directly.
Notes: Passing `--jobs N` allows batch to execute up to `N` tasks simultaneously; default is `1`, maintaining serial behavior.
Notes: Passing `--retries N` adds up to `N` retries for each batch `resolve/download/install` step; passing `--retry-backoff` can specify base backoff time, subsequent retries increase by 2x.

## Parameter Description

### `mirrors`

- `--source`: `npm`, `pip`, `cargo`, `maven`, `choco`, `winget`, `all`
- `--timeout`: Speed test timeout

### `download`

- `--source`: `npm`, `pip`, `cargo`, `maven`, `choco`, `winget`, `url`
- `--name`: Package name for `npm/pip/cargo/choco`; `maven` uses `group:artifact`
- `--id`: `winget` package identifier
- `--version`: Version number or tag; when left empty, `npm/pip/cargo/maven/choco/winget` will try to resolve the latest version
- `--url`: Direct download URL
- `--output`: Output directory
- `--mirror`: Mirror name or custom base URL
- `--config`: Optional configuration file path for loading `auth_profiles`
- `--auth-profile`: Authentication configuration name to use
- `--arch`: Expected architecture for `winget`, such as `x64`, `arm64`
- `--installer-index`: Specify `winget` installer by index
- `--resume`: Reuse existing `.part` file to continue download
- `--chunks`: Enable concurrent chunked download when server supports Range
- `--list-installers`: Only list `winget` installers, no download
- `--resolve-only`: Only resolve final download URL and filename, no download

Notes: Progress will be automatically displayed during actual download; `winget`, `search --pick`, `batch` all use the same downloader.

### `install`

- `--source`: Currently supports `npm`, `choco`, `winget`
- `--name`: Root npm package name, or package ID for choco/winget
- `--version`: Root package version, tag, or semver range, such as `latest`, `^19`, `>=18 <20`
- `--output`: Installation root directory, will generate `node_modules` underneath
- `--mirror`: npm mirror name or custom base URL
- `--config`: Optional configuration file path for loading `auth_profiles`
- `--auth-profile`: Authentication configuration name used for installation requests
- `--resume`: Reuse existing `.part` file to continue download when downloading cache tarball during installation
- `--chunks`: Enable concurrent chunked download when downloading cache tarball during installation
- `--omit-optional`: Skip `optionalDependencies` during installation
- `--include-peer`: Include `peerDependencies` during installation
- `--include-dev`: Include `devDependencies` for root package during installation
- `--lockfile`: Explicitly specify install lockfile path, default is `<output>\source-fetcher-install.lock.json`
- `--frozen-lockfile`: Require current install request to match lockfile, otherwise fail directly
- `--scripts`: Lifecycle script policy, supports `none`, `root`, `all`
- `--timeout`: Total timeout for resolution and download
- `--plan`: Only output dependency install plan, no download

Notes: Current version recursively resolves npm dependency tree, reuses tarball download and unpack cache according to the resolved unique version set, and lays out packages in a `node_modules` directory structure similar to npm; manifest records cache paths, store paths, `.bin` shims, and actual installation locations for each package.

### `uninstall`

- `--output`: Installation root directory, by default looks for `source-fetcher-install.json` here
- `--manifest`: Explicitly specify installation manifest path
- `--keep-cache`: Keep tarball/store cache under `.source-fetcher`
- `--keep-manifest`: Keep installation manifest after uninstall completes

Notes: Current version performs targeted uninstall according to manifest, only attempts to delete installation paths recorded in the manifest and still matching the corresponding `name@version`, and also recycles `.bin` shims recorded in the manifest.

### `repair`

- `--output`: Installation root directory, by default looks for `source-fetcher-install.json` here
- `--manifest`: Explicitly specify installation manifest path

Notes: Current version performs targeted repair according to manifest, only attempts to restore installation paths and missing `.bin` shims recorded in the manifest; if a path is already occupied by another `name@version`, it will be skipped instead of overwritten.

### `search`

- `--source`: `npm`, `pip`, `cargo`, `maven`, `choco`, `winget`, `all`
- `--query`: Search keyword
- `--mirror`: Specify mirror; currently mainly effective for `npm` and `choco`
- `--config`: Optional configuration file path for loading `auth_profiles`
- `--auth-profile`: Authentication configuration name used for search and download
- `--limit`: Maximum number of results per source
- `--interactive`: Interactively select one or more search results and continue download
- `--pick`: Directly select one or more search results and download, supports `2`, `1,3,5`
- `--output`: Specify download directory with `--pick` or `--interactive`
- `--resolve-only`: With `--pick` or `--interactive`, only resolve, no download
- `--arch`: With `--pick` or `--interactive`, specify architecture for `winget`
- `--installer-index`: With `--pick` or `--interactive`, specify installer index for `winget`
- `--resume`: With `--pick` or `--interactive`, reuse existing `.part` file to continue download
- `--chunks`: With `--pick` or `--interactive`, enable concurrent chunked download
- `--timeout`: Search request timeout

### `tui`

- `--source`: Initial search source, `npm`, `choco`, `winget`, `all`
- `--query`: Initial search keyword, can be left empty and entered in the interface
- `--mirror`: Initial mirror configuration
- `--config`: Configuration file path for `auth_profiles` and Batch page
- `--auth-profile`: Initial authentication configuration name for search page
- `--limit`: Maximum number of results per source
- `--output`: Download output directory
- `--arch`: Specify architecture for `winget` resolution/download initiated within TUI, can also be modified in search page
- `--installer-index`: Specify installer index for `winget` resolution/download initiated within TUI, can also be modified in search page
- `--resume`: Specify default resume behavior for downloads initiated within TUI, can also be modified in search page
- `--chunks`: Specify default chunk count for downloads initiated within TUI, can also be modified in search page
- `--timeout`: Search/resolution/download request timeout

Notes: Batch page displays mixed task list of `downloads` and `installs` after loading configuration, supports multi-select enqueue, and carries over `output_dir`, mirror defaults, and `timeout` from that configuration file.

### `batch`

- `--config`: YAML configuration file path, default is `source-fetcher.yaml`
- `--plan`: Only output resolution results for each task, no download or install
- `--continue-on-error`: Continue executing remaining tasks when a single task fails
- `--jobs`: Batch concurrent task count, default is `1`
- `--retries`: Retry count after a single batch `resolve/download/install` step fails, not including initial execution
- `--retry-backoff`: Batch retry base backoff time, such as `500ms`, `2s`; subsequent retries increase by 2x

### `auth_profiles`

- `headers`: HTTP Headers attached to requests
- `bearer_token_env`: Read Bearer Token from environment variable
- `basic_username` / `basic_password_env`: Read Basic Auth password from environment variable

Example:

```yaml
auth_profiles:
  private-npm:
    bearer_token_env: PRIVATE_NPM_TOKEN
    headers:
      X-Registry: internal

downloads:
  - source: npm
    name: internal-lib
    auth: private-npm
    resume: true
    chunks: 4

installs:
  - source: npm
    name: internal-lib
    version: ^1
    auth: private-npm
    include_peer: true
    scripts_policy: root
```

## Default Mirrors

### npm

- `huaweicloud`
- `npmjs`
- `npmmirror`
- `tencent`

Notes: When `--mirror` is not explicitly specified, `npm` will prioritize trying built-in domestic mirrors and automatically fall back to subsequent mirrors when search/resolution metadata fails.

### choco

- `chocolatey`
- `nuget`

### winget

- `github-api`
- `github-raw`
- `jsdelivr`

Notes: `winget` actually downloads the vendor installer URL declared in the manifest, mirrors mainly affect manifest fetching.

## Auto-Install Feature ⭐

Source Fetcher now supports **automatic installation** for Choco and Winget packages!

### How It Works

**Choco Auto-Install:**
1. Downloads the `.nupkg` file
2. Checks if `choco` command is available
3. Executes `choco install <package>.nupkg -y`
4. Reports installation results

**Winget Auto-Install:**
1. Downloads the installer file (.exe/.msi/.msix/.appx)
2. Automatically detects installer type
3. Executes installation with appropriate silent parameters
   - MSI: `msiexec /i <file> /quiet /norestart`
   - EXE: Tries `/S`, `/silent`, `/quiet`, `/verysilent`
   - MSIX/APPX: `Add-AppxPackage -Path <file>`
4. Provides manual install command if auto-install fails

### Requirements

**For Choco:**
- Chocolatey client must be installed
- Administrator privileges recommended

**For Winget:**
- Windows 10/11 (winget is built-in)
- Some installers require administrator privileges

### Batch Installation Example

Create `install-tools.yaml`:

```yaml
output_dir: ./downloads
timeout: 60s

installs:
  # Choco packages
  - source: choco
    name: git
  - source: choco
    name: 7zip
  
  # Winget packages
  - source: winget
    name: Microsoft.PowerToys
  - source: winget
    name: Microsoft.VisualStudioCode
```

Execute batch install:

```powershell
sfer batch --config install-tools.yaml
```

### Documentation

For detailed auto-install documentation, see:
- **English Guide**: `AUTO_INSTALL_GUIDE.md` (Chinese, comprehensive)
- **Implementation Details**: `IMPLEMENTATION_COMPLETE.md` (Chinese)

## 📈 Performance Data

### Download Speed Comparison

```
Source Fetcher (Mirror): ████████████ 10MB/s
npm install:             ███░░░░░░░░░  3MB/s
choco install:           ████░░░░░░░░  4MB/s
```

### Concurrent Download Effect

```
Single Thread:  ████░░░░░░░░  4MB/s
4 Threads:      ███████████░ 11MB/s
8 Threads:      ████████████ 12MB/s
```

## 🗺️ Roadmap

See [ROADMAP.md](ROADMAP.md) for future plans:

- 🔜 **v1.0** - Stable release
- 🔜 **v1.1** - More package sources (Homebrew, APT, YUM)
- 🔜 **v1.2** - Cross-ecosystem dependency management
- 🔜 **v2.0** - Plugin system and advanced features

## 🤝 Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for how to get involved.

### Contributors

Thanks to all contributors!

<!-- ALL-CONTRIBUTORS-LIST:START -->
<!-- Contributors list will be auto-generated here -->
<!-- ALL-CONTRIBUTORS-LIST:END -->

## 📄 License

This project is licensed under the [MIT License](LICENSE).

## 🙏 Acknowledgments

- Thanks to all package manager projects for inspiration
- Thanks to the open source community for support
- Thanks to all users for feedback

## 📞 Contact

- 🐛 [Report Bug](https://github.com/jiahe/source-fetcher/issues/new?template=bug_report.md)
- 💡 [Feature Request](https://github.com/jiahe/source-fetcher/issues/new?template=feature_request.md)
- 💬 [Discussions](https://github.com/jiahe/source-fetcher/discussions)
- ⭐ Star us if you find it useful!

## 📚 Documentation

- [Quick Start](QUICK_START.md) - Getting started guide
- [Architecture](ARCHITECTURE.md) - Technical architecture
- [Changelog](CHANGELOG.md) - Version history
- [Roadmap](ROADMAP.md) - Future plans
- [Contributing](CONTRIBUTING.md) - How to contribute

---

<div align="center">

**[⬆ Back to Top](#source-fetcher)**

Made with ❤️ by the Source Fetcher community

</div>
