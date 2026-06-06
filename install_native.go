package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// NativeInstallRequest 用于 choco/winget/cargo 的安装请求
type NativeInstallRequest struct {
	Source       string
	Name         string
	PackageID    string
	Version      string
	Arch         string
	Silent       bool
	Scope        string // user/machine (winget)
	Force        bool
	SkipConfirm  bool
	
	// Cargo 特定选项
	BuildBinary  bool   // 是否编译二进制文件
	InstallBinary bool  // 是否安装二进制文件到系统
	BinName      string // 指定要编译的二进制名称（可选）
}

// NativeInstallResult 安装结果
type NativeInstallResult struct {
	Source       string
	Identifier   string
	Version      string
	InstallerURL string
	LocalPath    string
	Installed    bool
	Output       string
	Duration     time.Duration
}

// executeNativeInstall 执行 choco/winget/cargo 的自动安装
func executeNativeInstall(ctx context.Context, client *http.Client, req NativeInstallRequest, downloadOptions DownloadOptions) (NativeInstallResult, error) {
	switch strings.ToLower(strings.TrimSpace(req.Source)) {
	case "choco":
		return executeChocoInstall(ctx, client, req, downloadOptions)
	case "winget":
		return executeWingetInstall(ctx, client, req, downloadOptions)
	case "cargo":
		return executeCargoInstall(ctx, client, req, downloadOptions)
	default:
		return NativeInstallResult{}, fmt.Errorf("native install only supports choco, winget, and cargo, got %q", req.Source)
	}
}

// executeChocoInstall 执行 Chocolatey 包安装
func executeChocoInstall(ctx context.Context, client *http.Client, req NativeInstallRequest, downloadOptions DownloadOptions) (NativeInstallResult, error) {
	start := time.Now()
	
	// 1. 下载 .nupkg 文件
	downloadReq := DownloadRequest{
		Source:         "choco",
		Name:           req.Name,
		Version:        req.Version,
		OutputDir:      downloadOptions.OutputDir,
		Mirror:         "",
		Resume:         downloadOptions.Resume,
		Chunks:         downloadOptions.Chunks,
		RequestOptions: downloadOptions.RequestOptions,
	}
	
	plan, err := resolveDownloadPlan(ctx, client, downloadReq)
	if err != nil {
		return NativeInstallResult{}, fmt.Errorf("resolve choco package: %w", err)
	}
	
	result, err := downloadPlan(ctx, client, plan, downloadOptions)
	if err != nil {
		return NativeInstallResult{}, fmt.Errorf("download choco package: %w", err)
	}
	
	// 2. 检查是否有 choco 命令
	if !isCommandAvailable("choco") {
		return NativeInstallResult{
			Source:       "choco",
			Identifier:   req.Name,
			Version:      plan.Version,
			InstallerURL: plan.URL,
			LocalPath:    result.Path,
			Installed:    false,
			Output:       "choco command not found. Please install Chocolatey first or install manually: choco install " + result.Path,
			Duration:     time.Since(start),
		}, nil
	}
	
	// 3. 构建安装命令
	args := []string{"install", result.Path}
	if req.Silent || req.SkipConfirm {
		args = append(args, "-y")
	}
	if req.Force {
		args = append(args, "--force")
	}
	
	// 4. 执行安装
	cmd := exec.CommandContext(ctx, "choco", args...)
	output, err := cmd.CombinedOutput()
	
	installed := err == nil
	outputStr := string(output)
	
	if err != nil {
		outputStr = fmt.Sprintf("Installation failed: %v\n%s", err, output)
	}
	
	return NativeInstallResult{
		Source:       "choco",
		Identifier:   req.Name,
		Version:      plan.Version,
		InstallerURL: plan.URL,
		LocalPath:    result.Path,
		Installed:    installed,
		Output:       outputStr,
		Duration:     time.Since(start),
	}, nil
}

// executeWingetInstall 执行 Winget 包安装
func executeWingetInstall(ctx context.Context, client *http.Client, req NativeInstallRequest, downloadOptions DownloadOptions) (NativeInstallResult, error) {
	start := time.Now()
	
	// 1. 下载安装器
	downloadReq := DownloadRequest{
		Source:         "winget",
		PackageID:      req.PackageID,
		Version:        req.Version,
		Arch:           req.Arch,
		OutputDir:      downloadOptions.OutputDir,
		Resume:         downloadOptions.Resume,
		Chunks:         downloadOptions.Chunks,
		RequestOptions: downloadOptions.RequestOptions,
	}
	
	plan, err := resolveDownloadPlan(ctx, client, downloadReq)
	if err != nil {
		return NativeInstallResult{}, fmt.Errorf("resolve winget package: %w", err)
	}
	
	result, err := downloadPlan(ctx, client, plan, downloadOptions)
	if err != nil {
		return NativeInstallResult{}, fmt.Errorf("download winget installer: %w", err)
	}
	
	// 2. 检测安装器类型并执行
	installerType := detectInstallerType(result.Path)
	
	var cmd *exec.Cmd
	var outputStr string
	var installed bool
	
	switch installerType {
	case "msi":
		// MSI 安装器
		args := []string{"/i", result.Path}
		if req.Silent {
			args = append(args, "/quiet", "/norestart")
		}
		cmd = exec.CommandContext(ctx, "msiexec", args...)
		
	case "exe":
		// EXE 安装器 - 尝试常见的静默参数
		args := []string{}
		if req.Silent {
			// 尝试多种静默参数
			args = append(args, "/S", "/silent", "/quiet", "/verysilent")
		}
		cmd = exec.CommandContext(ctx, result.Path, args...)
		
	case "msix", "appx":
		// MSIX/APPX 包 - 使用 PowerShell
		psScript := fmt.Sprintf("Add-AppxPackage -Path '%s'", result.Path)
		cmd = exec.CommandContext(ctx, "powershell", "-Command", psScript)
		
	default:
		// 未知类型，尝试直接运行
		cmd = exec.CommandContext(ctx, result.Path)
	}
	
	// 3. 执行安装
	output, err := cmd.CombinedOutput()
	outputStr = string(output)
	installed = err == nil
	
	if err != nil {
		// 如果自动安装失败，提供手动安装提示
		outputStr = fmt.Sprintf("Auto-install failed: %v\n%s\n\nManual install command:\n", err, output)
		switch installerType {
		case "msi":
			outputStr += fmt.Sprintf("msiexec /i \"%s\" /quiet\n", result.Path)
		case "exe":
			outputStr += fmt.Sprintf("\"%s\" /S\n", result.Path)
			outputStr += fmt.Sprintf("or try: \"%s\" /silent\n", result.Path)
		case "msix", "appx":
			outputStr += fmt.Sprintf("Add-AppxPackage -Path \"%s\"\n", result.Path)
		default:
			outputStr += fmt.Sprintf("\"%s\"\n", result.Path)
		}
	}
	
	return NativeInstallResult{
		Source:       "winget",
		Identifier:   req.PackageID,
		Version:      plan.Version,
		InstallerURL: plan.URL,
		LocalPath:    result.Path,
		Installed:    installed,
		Output:       outputStr,
		Duration:     time.Since(start),
	}, nil
}

// detectInstallerType 检测安装器类型
func detectInstallerType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".msi":
		return "msi"
	case ".exe":
		return "exe"
	case ".msix":
		return "msix"
	case ".appx":
		return "appx"
	default:
		return "unknown"
	}
}

// executeCargoInstall 执行 Cargo crate 安装（支持可选编译）
// .crate 文件是 gzip 压缩的 tar 归档，直接解压到目标目录
// 如果指定了编译选项，会自动编译并安装二进制文件
func executeCargoInstall(ctx context.Context, client *http.Client, req NativeInstallRequest, downloadOptions DownloadOptions) (NativeInstallResult, error) {
	start := time.Now()
	
	// 1. 下载 .crate 文件
	downloadReq := DownloadRequest{
		Source:         "cargo",
		Name:           req.Name,
		Version:        req.Version,
		OutputDir:      downloadOptions.OutputDir,
		Mirror:         "",
		Resume:         downloadOptions.Resume,
		Chunks:         downloadOptions.Chunks,
		RequestOptions: downloadOptions.RequestOptions,
	}
	
	plan, err := resolveDownloadPlan(ctx, client, downloadReq)
	if err != nil {
		return NativeInstallResult{}, fmt.Errorf("resolve cargo crate: %w", err)
	}
	
	result, err := downloadPlan(ctx, client, plan, downloadOptions)
	if err != nil {
		return NativeInstallResult{}, fmt.Errorf("download cargo crate: %w", err)
	}
	
	// 2. 确定安装目标目录
	installDir := filepath.Join(downloadOptions.OutputDir, "cargo-crates", sanitizeFileName(req.Name+"-"+plan.Version))
	if err := os.MkdirAll(installDir, 0o755); err != nil {
		return NativeInstallResult{}, fmt.Errorf("create install directory: %w", err)
	}
	
	// 3. 解压 .crate 文件
	if err := extractCargoArchive(result.Path, installDir); err != nil {
		return NativeInstallResult{}, fmt.Errorf("extract cargo crate: %w", err)
	}
	
	outputMsg := fmt.Sprintf("Successfully extracted %s@%s to %s\n", req.Name, plan.Version, installDir)
	
	// 4. 如果启用了编译选项，尝试编译
	var binaryPath string
	var buildSuccess bool
	
	if req.BuildBinary {
		outputMsg += "\n--- Building binary ---\n"
		
		// 检查 cargo 是否可用
		available, cargoVersion, err := checkCargoAvailable()
		if !available {
			outputMsg += fmt.Sprintf("⚠️  Cargo not found: %v\n", err)
			outputMsg += "   Binary not built. Source code is available in the directory above.\n"
			outputMsg += "   To build manually:\n"
			outputMsg += fmt.Sprintf("     cd %s\n", installDir)
			outputMsg += "     cargo build --release\n"
		} else {
			outputMsg += fmt.Sprintf("Using %s\n", cargoVersion)
			
			// 构建选项
			buildOptions := CargoBuildOptions{
				SourceDir:   installDir,
				Release:     true,  // 默认 release 模式
				Verbose:     false,
			}
			
			// 如果指定了二进制名称
			if req.BinName != "" {
				buildOptions.BinName = req.BinName
			}
			
			// 执行编译
			buildResult, err := buildCargoCrate(ctx, buildOptions)
			if err != nil {
				outputMsg += fmt.Sprintf("❌ Build failed: %v\n", err)
				outputMsg += fmt.Sprintf("\nBuild output:\n%s\n", buildResult.Output)
			} else {
				buildSuccess = true
				binaryPath = buildResult.BinaryPath
				outputMsg += fmt.Sprintf("✅ Build successful (%.2fs)\n", buildResult.Duration.Seconds())
				outputMsg += fmt.Sprintf("Binary: %s\n", binaryPath)
				
				// 5. 如果启用了安装选项，安装二进制文件
				if req.InstallBinary && binaryPath != "" {
					outputMsg += "\n--- Installing binary ---\n"
					
					installBinDir := getCargoInstallDir()
					installedPath, err := installCargoBinary(binaryPath, installBinDir)
					if err != nil {
						outputMsg += fmt.Sprintf("❌ Install failed: %v\n", err)
					} else {
						outputMsg += fmt.Sprintf("✅ Installed to: %s\n", installedPath)
						binaryPath = installedPath
						
						// 检查 PATH
						if !checkPathContains(installBinDir) {
							outputMsg += fmt.Sprintf("\n⚠️  Warning: %s is not in PATH\n", installBinDir)
							outputMsg += "   Add it to PATH to use the binary globally:\n"
							if runtime.GOOS == "windows" {
								outputMsg += fmt.Sprintf("     $env:Path += \";%s\"\n", installBinDir)
							} else {
								outputMsg += fmt.Sprintf("     export PATH=\"%s:$PATH\"\n", installBinDir)
							}
						}
					}
				}
			}
		}
	} else {
		// 不编译，提供使用说明
		outputMsg += "\nℹ️  This is a source distribution. To build and use:\n"
		outputMsg += fmt.Sprintf("  cd %s\n", installDir)
		outputMsg += "  cargo build --release\n"
		outputMsg += "  # Binary will be in target/release/\n"
	}
	
	return NativeInstallResult{
		Source:       "cargo",
		Identifier:   req.Name,
		Version:      plan.Version,
		InstallerURL: plan.URL,
		LocalPath:    binaryPath,  // 如果编译了，返回二进制路径
		Installed:    buildSuccess, // 如果编译成功，标记为已安装
		Output:       outputMsg,
		Duration:     time.Since(start),
	}, nil
}

// extractCargoArchive 解压 .crate 文件到目标目录
// .crate 文件是 gzip 压缩的 tar 归档，类似 npm tarball
func extractCargoArchive(archivePath string, destDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("open gzip stream: %w", err)
	}
	defer gzipReader.Close()

	reader := tar.NewReader(gzipReader)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read archive entry: %w", err)
		}
		
		// Cargo crate 文件格式：包名-版本/文件路径
		// 例如：serde-1.0.0/Cargo.toml, serde-1.0.0/src/lib.rs
		relativePath, include, err := trimCargoArchivePath(header.Name)
		if err != nil {
			return err
		}
		if !include {
			continue
		}
		
		targetPath := filepath.Join(destDir, filepath.FromSlash(relativePath))
		if !isPathWithinBase(targetPath, destDir) && !sameFilePath(targetPath, destDir) {
			return fmt.Errorf("archive entry escapes destination: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return fmt.Errorf("create dir %s: %w", targetPath, err)
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
				return fmt.Errorf("create parent dir for %s: %w", targetPath, err)
			}
			if err := writeCargoArchiveFile(reader, targetPath, header.FileInfo().Mode().Perm()); err != nil {
				return err
			}
		case tar.TypeSymlink:
			// 跳过符号链接以保证安全
			continue
		default:
			continue
		}
	}
}

// trimCargoArchivePath 处理 cargo crate 归档路径
// cargo crate 格式：包名-版本/文件路径，我们需要去掉顶层目录
func trimCargoArchivePath(name string) (string, bool, error) {
	cleaned := strings.TrimSpace(strings.ReplaceAll(name, "\\", "/"))
	cleaned = strings.TrimPrefix(cleaned, "./")
	if cleaned == "" {
		return "", false, nil
	}
	
	// 去掉路径分隔符后缀
	cleaned = strings.TrimSuffix(cleaned, "/")
	if cleaned == "" || cleaned == "." {
		return "", false, nil
	}
	
	// 找到第一个斜杠，去掉顶层目录（包名-版本）
	firstSlash := strings.Index(cleaned, "/")
	if firstSlash < 0 {
		// 顶层目录本身，跳过
		return "", false, nil
	}
	
	// 提取顶层目录之后的路径
	cleaned = cleaned[firstSlash+1:]
	if cleaned == "" || cleaned == "." {
		return "", false, nil
	}
	
	// 检查路径遍历攻击
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.Contains(cleaned, "/../") {
		return "", false, fmt.Errorf("archive entry escapes package root: %s", name)
	}
	
	return cleaned, true, nil
}

// writeCargoArchiveFile 写入归档文件
func writeCargoArchiveFile(reader io.Reader, targetPath string, mode os.FileMode) error {
	// 确保权限合理
	if mode == 0 {
		mode = 0o644
	}
	file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return fmt.Errorf("create file %s: %w", targetPath, err)
	}
	defer file.Close()
	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("write file %s: %w", targetPath, err)
	}
	return nil
}

// isCommandAvailable 检查命令是否可用
func isCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// resolveNativeInstallPlan 解析 choco/winget/cargo 安装计划
func resolveNativeInstallPlan(ctx context.Context, client *http.Client, req InstallRequest) (InstallPlan, error) {
	switch strings.ToLower(strings.TrimSpace(req.Source)) {
	case "choco":
		return resolveChocoInstallPlan(ctx, client, req)
	case "winget":
		return resolveWingetInstallPlan(ctx, client, req)
	case "cargo":
		return resolveCargoInstallPlan(ctx, client, req)
	default:
		return InstallPlan{}, fmt.Errorf("native install plan only supports choco, winget, and cargo")
	}
}

// resolveChocoInstallPlan 解析 Choco 安装计划
func resolveChocoInstallPlan(ctx context.Context, client *http.Client, req InstallRequest) (InstallPlan, error) {
	if strings.TrimSpace(req.Name) == "" {
		return InstallPlan{}, errors.New("--name is required when --source choco")
	}
	
	// 解析包信息
	downloadReq := DownloadRequest{
		Source:         "choco",
		Name:           req.Name,
		Version:        req.Version,
		Mirror:         req.Mirror,
		RequestOptions: req.RequestOptions,
	}
	
	plan, err := resolveDownloadPlan(ctx, client, downloadReq)
	if err != nil {
		return InstallPlan{}, err
	}
	
	// 构建安装计划
	pkg := InstallPackage{
		Name:       req.Name,
		Requested:  req.Version,
		Version:    plan.Version,
		MirrorName: plan.MirrorName,
		URL:        plan.URL,
		Filename:   plan.Filename,
	}
	
	return InstallPlan{
		Source:      "choco",
		Root:        req.Name,
		Requested:   req.Version,
		RootVersion: plan.Version,
		MirrorName:  plan.MirrorName,
		Packages:    []InstallPackage{pkg},
	}, nil
}

// resolveWingetInstallPlan 解析 Winget 安装计划
func resolveWingetInstallPlan(ctx context.Context, client *http.Client, req InstallRequest) (InstallPlan, error) {
	if strings.TrimSpace(req.Name) == "" {
		return InstallPlan{}, errors.New("--name (package ID) is required when --source winget")
	}
	
	// 解析包信息
	downloadReq := DownloadRequest{
		Source:         "winget",
		PackageID:      req.Name,
		Version:        req.Version,
		Mirror:         req.Mirror,
		RequestOptions: req.RequestOptions,
	}
	
	plan, err := resolveDownloadPlan(ctx, client, downloadReq)
	if err != nil {
		return InstallPlan{}, err
	}
	
	// 构建安装计划
	pkg := InstallPackage{
		Name:       req.Name,
		Requested:  req.Version,
		Version:    plan.Version,
		MirrorName: plan.MirrorName,
		URL:        plan.URL,
		Filename:   plan.Filename,
	}
	
	return InstallPlan{
		Source:      "winget",
		Root:        req.Name,
		Requested:   req.Version,
		RootVersion: plan.Version,
		MirrorName:  plan.MirrorName,
		Packages:    []InstallPackage{pkg},
	}, nil
}

// resolveCargoInstallPlan 解析 Cargo 安装计划
func resolveCargoInstallPlan(ctx context.Context, client *http.Client, req InstallRequest) (InstallPlan, error) {
	if strings.TrimSpace(req.Name) == "" {
		return InstallPlan{}, errors.New("--name is required when --source cargo")
	}
	
	// 解析包信息
	downloadReq := DownloadRequest{
		Source:         "cargo",
		Name:           req.Name,
		Version:        req.Version,
		Mirror:         req.Mirror,
		RequestOptions: req.RequestOptions,
	}
	
	plan, err := resolveDownloadPlan(ctx, client, downloadReq)
	if err != nil {
		return InstallPlan{}, err
	}
	
	// 构建安装计划
	pkg := InstallPackage{
		Name:       req.Name,
		Requested:  req.Version,
		Version:    plan.Version,
		MirrorName: plan.MirrorName,
		URL:        plan.URL,
		Filename:   plan.Filename,
	}
	
	return InstallPlan{
		Source:      "cargo",
		Root:        req.Name,
		Requested:   req.Version,
		RootVersion: plan.Version,
		MirrorName:  plan.MirrorName,
		Packages:    []InstallPackage{pkg},
	}, nil
}

// executeNativeInstallPlan 执行 choco/winget/cargo 安装计划
func executeNativeInstallPlan(ctx context.Context, client *http.Client, plan InstallPlan, options DownloadOptions, req InstallRequest) (InstallResult, error) {
	if len(plan.Packages) == 0 {
		return InstallResult{}, errors.New("install plan contains no packages")
	}
	
	// 对于 choco/winget/cargo，我们只处理单个包
	pkg := plan.Packages[0]
	
	// 构建 native install request
	nativeReq := NativeInstallRequest{
		Source:        plan.Source,
		Name:          pkg.Name,
		PackageID:     pkg.Name,
		Version:       pkg.Version,
		Silent:        true,
		SkipConfirm:   true,
		BuildBinary:   req.CargoBuild,
		InstallBinary: req.CargoInstall,
		BinName:       req.CargoBinName,
	}
	
	// 执行安装
	result, err := executeNativeInstall(ctx, client, nativeReq, options)
	if err != nil {
		return InstallResult{}, err
	}
	
	// 转换为 InstallResult
	installedPkg := InstalledPackage{
		Name:       result.Identifier,
		Version:    result.Version,
		MirrorName: plan.MirrorName,
		ArchivePath: result.LocalPath,
		StorePath:  result.LocalPath,
		InstallPaths: []string{result.LocalPath},
		Size:       0, // 未知大小
		SHA256:     "", // 未计算
	}
	
	return InstallResult{
		ManifestPath:   "",
		LockfilePath:   "",
		RootPath:       result.LocalPath,
		NodeModulesDir: "",
		Packages:       []InstalledPackage{installedPkg},
		Duration:       result.Duration,
	}, nil
}

// printNativeInstallResult 打印安装结果
func printNativeInstallResult(out io.Writer, result NativeInstallResult) {
	_, _ = fmt.Fprintf(out, "Source: %s\n", result.Source)
	_, _ = fmt.Fprintf(out, "Package: %s\n", result.Identifier)
	_, _ = fmt.Fprintf(out, "Version: %s\n", result.Version)
	_, _ = fmt.Fprintf(out, "Installer URL: %s\n", result.InstallerURL)
	_, _ = fmt.Fprintf(out, "Downloaded To: %s\n", result.LocalPath)
	
	if result.Installed {
		_, _ = fmt.Fprintf(out, "Status: ✓ Installed Successfully\n")
	} else {
		_, _ = fmt.Fprintf(out, "Status: ✗ Installation Failed or Skipped\n")
	}
	
	_, _ = fmt.Fprintf(out, "Duration: %s\n", roundDuration(result.Duration))
	
	if result.Output != "" {
		_, _ = fmt.Fprintf(out, "\nInstallation Output:\n%s\n", result.Output)
	}
}
