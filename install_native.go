package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// NativeInstallRequest 用于 choco/winget 的安装请求
type NativeInstallRequest struct {
	Source      string
	Name        string
	PackageID   string
	Version     string
	Arch        string
	Silent      bool
	Scope       string // user/machine (winget)
	Force       bool
	SkipConfirm bool
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

// executeNativeInstall 执行 choco/winget 的自动安装
func executeNativeInstall(ctx context.Context, client *http.Client, req NativeInstallRequest, downloadOptions DownloadOptions) (NativeInstallResult, error) {
	switch strings.ToLower(strings.TrimSpace(req.Source)) {
	case "choco":
		return executeChocoInstall(ctx, client, req, downloadOptions)
	case "winget":
		return executeWingetInstall(ctx, client, req, downloadOptions)
	default:
		return NativeInstallResult{}, fmt.Errorf("native install only supports choco and winget, got %q", req.Source)
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

// isCommandAvailable 检查命令是否可用
func isCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// resolveNativeInstallPlan 解析 choco/winget 安装计划
func resolveNativeInstallPlan(ctx context.Context, client *http.Client, req InstallRequest) (InstallPlan, error) {
	switch strings.ToLower(strings.TrimSpace(req.Source)) {
	case "choco":
		return resolveChocoInstallPlan(ctx, client, req)
	case "winget":
		return resolveWingetInstallPlan(ctx, client, req)
	default:
		return InstallPlan{}, fmt.Errorf("native install plan only supports choco and winget")
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

// executeNativeInstallPlan 执行 choco/winget 安装计划
func executeNativeInstallPlan(ctx context.Context, client *http.Client, plan InstallPlan, options DownloadOptions) (InstallResult, error) {
	if len(plan.Packages) == 0 {
		return InstallResult{}, errors.New("install plan contains no packages")
	}
	
	// 对于 choco/winget，我们只处理单个包
	pkg := plan.Packages[0]
	
	// 构建 native install request
	nativeReq := NativeInstallRequest{
		Source:      plan.Source,
		Name:        pkg.Name,
		PackageID:   pkg.Name,
		Version:     pkg.Version,
		Silent:      true,
		SkipConfirm: true,
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
