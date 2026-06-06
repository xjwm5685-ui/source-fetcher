package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// UnifiedManifest 统一的安装清单，支持多生态
type UnifiedManifest struct {
	Version     int                     `json:"version"`
	GeneratedAt string                  `json:"generated_at"`
	Ecosystems  map[string][]UnifiedInstallRecord `json:"ecosystems"`
}

// UnifiedInstallRecord 统一的安装记录
type UnifiedInstallRecord struct {
	ID              string            `json:"id"`
	Source          string            `json:"source"`
	Package         string            `json:"package"`
	Version         string            `json:"version"`
	InstalledAt     string            `json:"installed_at"`
	InstallPaths    []string          `json:"install_paths,omitempty"`
	UninstallMethod string            `json:"uninstall_method"`
	UninstallData   map[string]string `json:"uninstall_data,omitempty"`
}

// UnifiedUninstallRequest 统一卸载请求
type UnifiedUninstallRequest struct {
	ManifestPath string
	Source       string   // 指定生态，空则全部
	PackageNames []string // 指定包名，空则全部
	DryRun       bool     // 只预览，不实际执行
	KeepManifest bool     // 保留清单文件
	Interactive  bool     // 交互式确认
}

// UnifiedUninstallResult 统一卸载结果
type UnifiedUninstallResult struct {
	Successes []UninstallSuccess
	Failures  []UninstallFailure
	Skipped   []string
	Duration  time.Duration
}

// UninstallSuccess 成功的卸载
type UninstallSuccess struct {
	Source       string
	Package      string
	Version      string
	Method       string
	RemovedPaths []string
}

// UninstallFailure 失败的卸载
type UninstallFailure struct {
	Source  string
	Package string
	Version string
	Error   error
}

// UninstallChoco 卸载 Chocolatey 包
func UninstallChoco(ctx context.Context, packageName string, dryRun bool) error {
	if runtime.GOOS != "windows" {
		return errors.New("chocolatey is only supported on Windows")
	}

	// 验证包名格式（字母数字、点、连字符、下划线）
	if !isValidPackageName(packageName) {
		return fmt.Errorf("invalid package name format: %s (allowed: alphanumeric, dots, hyphens, underscores)", packageName)
	}

	// 检查 choco 是否安装
	if _, err := exec.LookPath("choco"); err != nil {
		return fmt.Errorf("chocolatey not found: %w (install from https://chocolatey.org)", err)
	}

	if dryRun {
		fmt.Printf("[DRY RUN] Would run: choco uninstall %s --yes\n", packageName)
		return nil
	}

	cmd := exec.CommandContext(ctx, "choco", "uninstall", packageName, "--yes")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if errors.Is(err, context.Canceled) {
			return fmt.Errorf("uninstall canceled by user: %w", err)
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("uninstall timed out: %w", err)
		}
		return fmt.Errorf("choco uninstall failed: %w", err)
	}

	return nil
}

// UninstallWinget 卸载 Winget 包
func UninstallWinget(ctx context.Context, packageID string, dryRun bool) error {
	if runtime.GOOS != "windows" {
		return errors.New("winget is only supported on Windows")
	}

	// 验证 packageID 格式（字母数字、点、连字符、下划线）
	if !isValidPackageName(packageID) {
		return fmt.Errorf("invalid package ID format: %s (allowed: alphanumeric, dots, hyphens, underscores)", packageID)
	}

	// winget 在 Windows 10/11 默认安装
	if dryRun {
		fmt.Printf("[DRY RUN] Would run: winget uninstall --id %s --silent\n", packageID)
		return nil
	}

	cmd := exec.CommandContext(ctx, "winget", "uninstall", "--id", packageID, "--silent")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if errors.Is(err, context.Canceled) {
			return fmt.Errorf("uninstall canceled by user: %w", err)
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("uninstall timed out: %w", err)
		}
		// 检查是否是已知的"良性"错误码
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode := exitErr.ExitCode()
			if isWingetBenignExitCode(exitCode) {
				fmt.Printf("Info: winget returned exit code %d (considered success)\n", exitCode)
				return nil
			}
		}
		return fmt.Errorf("winget uninstall failed: %w", err)
	}

	return nil
}

// UninstallURLPaths 删除 URL 安装的文件路径
func UninstallURLPaths(paths []string, dryRun bool) error {
	if len(paths) == 0 {
		return nil
	}

	if dryRun {
		fmt.Println("[DRY RUN] Would remove paths:")
		for _, path := range paths {
			fmt.Printf("  - %s\n", path)
		}
		return nil
	}

	var errs []error
	for _, path := range paths {
		if err := os.RemoveAll(path); err != nil {
			errs = append(errs, fmt.Errorf("remove %s: %w", path, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to remove some paths: %v", errs)
	}

	return nil
}

// LoadUnifiedManifest 加载统一清单
func LoadUnifiedManifest(path string) (*UnifiedManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read unified manifest: %w", err)
	}

	var manifest UnifiedManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("parse unified manifest: %w", err)
	}

	// 版本检查
	const currentVersion = 1
	if manifest.Version == 0 {
		// 旧版本或未设置版本，尝试迁移
		manifest.Version = currentVersion
		fmt.Println("Warning: manifest version not set, assuming version 1")
	} else if manifest.Version > currentVersion {
		return nil, fmt.Errorf("manifest version %d is newer than supported version %d (please upgrade source-fetcher)", 
			manifest.Version, currentVersion)
	}

	return &manifest, nil
}

// SaveUnifiedManifest 保存统一清单
func SaveUnifiedManifest(path string, manifest *UnifiedManifest) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal unified manifest: %w", err)
	}

	// 使用安全的文件权限：仅所有者可读写
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write unified manifest: %w", err)
	}

	return nil
}

// ExecuteUnifiedUninstall 执行统一卸载
func ExecuteUnifiedUninstall(ctx context.Context, req UnifiedUninstallRequest) (UnifiedUninstallResult, error) {
	start := time.Now()
	result := UnifiedUninstallResult{
		Successes: []UninstallSuccess{},
		Failures:  []UninstallFailure{},
		Skipped:   []string{},
	}

	// 加载清单
	manifest, err := LoadUnifiedManifest(req.ManifestPath)
	if err != nil {
		return result, err
	}

	// 收集要卸载的记录
	toUninstall := collectUninstallRecords(manifest, req.Source, req.PackageNames)

	if len(toUninstall) == 0 {
		return result, errors.New("no packages found to uninstall")
	}

	// 交互式确认
	if req.Interactive && !req.DryRun {
		if !confirmUninstall(toUninstall) {
			return result, errors.New("uninstall cancelled by user")
		}
	}

	// 执行卸载
	for _, record := range toUninstall {
		success, failure := uninstallSingleRecord(ctx, record, req.DryRun)
		if success != nil {
			result.Successes = append(result.Successes, *success)
		}
		if failure != nil {
			result.Failures = append(result.Failures, *failure)
		}
	}

	// 更新清单
	if !req.DryRun && !req.KeepManifest {
		updateManifestAfterUninstall(manifest, result.Successes)
		if err := SaveUnifiedManifest(req.ManifestPath, manifest); err != nil {
			fmt.Printf("Warning: failed to update manifest: %v\n", err)
		}
	}

	result.Duration = time.Since(start)
	return result, nil
}

func collectUninstallRecords(manifest *UnifiedManifest, source string, packageNames []string) []UnifiedInstallRecord {
	var records []UnifiedInstallRecord

	for eco, ecoRecords := range manifest.Ecosystems {
		// 过滤生态
		if source != "" && !strings.EqualFold(eco, source) {
			continue
		}

		for _, record := range ecoRecords {
			// 过滤包名
			if len(packageNames) > 0 {
				found := false
				for _, name := range packageNames {
					if strings.EqualFold(record.Package, name) {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			records = append(records, record)
		}
	}

	return records
}

func confirmUninstall(records []UnifiedInstallRecord) bool {
	fmt.Println("\n=== Packages to Uninstall ===")
	for _, record := range records {
		fmt.Printf("  [%s] %s@%s\n", record.Source, record.Package, record.Version)
	}
	fmt.Printf("\nTotal: %d packages\n", len(records))
	fmt.Print("\nProceed with uninstall? (yes/no): ")

	var response string
	fmt.Scanln(&response)
	return strings.ToLower(strings.TrimSpace(response)) == "yes"
}

func uninstallSingleRecord(ctx context.Context, record UnifiedInstallRecord, dryRun bool) (*UninstallSuccess, *UninstallFailure) {
	var err error

	switch strings.ToLower(record.Source) {
	case "npm":
		// npm 使用现有的卸载逻辑
		err = errors.New("npm uninstall should use existing uninstall command")

	case "choco":
		err = UninstallChoco(ctx, record.Package, dryRun)

	case "winget":
		err = UninstallWinget(ctx, record.Package, dryRun)

	case "url":
		err = UninstallURLPaths(record.InstallPaths, dryRun)

	default:
		err = fmt.Errorf("unsupported source: %s", record.Source)
	}

	if err != nil {
		return nil, &UninstallFailure{
			Source:  record.Source,
			Package: record.Package,
			Version: record.Version,
			Error:   err,
		}
	}

	return &UninstallSuccess{
		Source:       record.Source,
		Package:      record.Package,
		Version:      record.Version,
		Method:       record.UninstallMethod,
		RemovedPaths: record.InstallPaths,
	}, nil
}

func updateManifestAfterUninstall(manifest *UnifiedManifest, successes []UninstallSuccess) {
	for _, success := range successes {
		source := strings.ToLower(success.Source)
		if records, ok := manifest.Ecosystems[source]; ok {
			filtered := make([]UnifiedInstallRecord, 0, len(records))
			for _, record := range records {
				if !(strings.EqualFold(record.Package, success.Package) &&
					record.Version == success.Version) {
					filtered = append(filtered, record)
				}
			}
			manifest.Ecosystems[source] = filtered
		}
	}
}

// PrintUnifiedUninstallResult 打印卸载结果
func PrintUnifiedUninstallResult(result UnifiedUninstallResult) {
	fmt.Println("\n=== Uninstall Results ===")

	if len(result.Successes) > 0 {
		fmt.Println("\n✅ Successfully Uninstalled:")
		for _, success := range result.Successes {
			fmt.Printf("  [%s] %s@%s\n", success.Source, success.Package, success.Version)
		}
	}

	if len(result.Failures) > 0 {
		fmt.Println("\n❌ Failed:")
		for _, failure := range result.Failures {
			fmt.Printf("  [%s] %s@%s: %v\n", failure.Source, failure.Package, failure.Version, failure.Error)
		}
	}

	if len(result.Skipped) > 0 {
		fmt.Println("\n⏭️  Skipped:")
		for _, skipped := range result.Skipped {
			fmt.Printf("  %s\n", skipped)
		}
	}

	fmt.Printf("\nTotal: %d successful, %d failed, %d skipped\n",
		len(result.Successes), len(result.Failures), len(result.Skipped))
	fmt.Printf("Duration: %v\n", result.Duration)
}

// MigrateNPMManifestToUnified 将 npm 清单迁移到统一格式
func MigrateNPMManifestToUnified(npmManifestPath string) (*UnifiedManifest, error) {
	data, err := os.ReadFile(npmManifestPath)
	if err != nil {
		return nil, err
	}

	var npmManifest InstallManifest
	if err := json.Unmarshal(data, &npmManifest); err != nil {
		return nil, err
	}

	unified := &UnifiedManifest{
		Version:     1,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Ecosystems:  make(map[string][]UnifiedInstallRecord),
	}

	npmRecords := make([]UnifiedInstallRecord, len(npmManifest.Packages))
	for i, pkg := range npmManifest.Packages {
		npmRecords[i] = UnifiedInstallRecord{
			ID:              fmt.Sprintf("npm-%s-%s", pkg.Name, pkg.Version),
			Source:          "npm",
			Package:         pkg.Name,
			Version:         pkg.Version,
			InstalledAt:     npmManifest.GeneratedAt,
			InstallPaths:    pkg.InstallPaths,
			UninstallMethod: "manifest",
			UninstallData: map[string]string{
				"manifest_path": npmManifestPath,
			},
		}
	}

	unified.Ecosystems["npm"] = npmRecords
	return unified, nil
}

// isValidPackageName 验证包名/ID 格式
// 允许：字母数字、点、连字符、下划线，长度 1-256
func isValidPackageName(name string) bool {
	if len(name) == 0 || len(name) > 256 {
		return false
	}
	
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '.' || c == '-' || c == '_') {
			return false
		}
	}
	
	return true
}

// isWingetBenignExitCode 检查 winget 退出码是否为"良性"
// 参考：https://learn.microsoft.com/en-us/windows/package-manager/winget/returnCodes
func isWingetBenignExitCode(code int) bool {
	benignCodes := map[int]string{
		0:           "SUCCESS",
		-1978335189: "ERROR_NO_APPLICABLE_UPDATE (package not found/already uninstalled)",
	}
	_, ok := benignCodes[code]
	return ok
}
