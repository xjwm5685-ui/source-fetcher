package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// CargoBuildOptions 编译选项
type CargoBuildOptions struct {
	SourceDir    string   // 源码目录
	Release      bool     // 是否 release 模式
	Target       string   // 目标平台（如 x86_64-pc-windows-msvc）
	Features     []string // 启用的特性
	AllFeatures  bool     // 启用所有特性
	NoDefault    bool     // 禁用默认特性
	Jobs         int      // 并行编译任务数
	Verbose      bool     // 详细输出
	BinName      string   // 指定要编译的二进制名称
}

// CargoBuildResult 编译结果
type CargoBuildResult struct {
	Success      bool          // 是否成功
	BinaryPath   string        // 编译后的二进制文件路径
	Output       string        // 编译输出
	Duration     time.Duration // 编译耗时
	Installed    bool          // 是否已安装到系统
	InstallPath  string        // 安装路径
}

// checkCargoAvailable 检查 Rust/Cargo 是否可用
func checkCargoAvailable() (bool, string, error) {
	cmd := exec.Command("cargo", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, "", fmt.Errorf("cargo not found: %w", err)
	}
	
	version := strings.TrimSpace(string(output))
	return true, version, nil
}

// buildCargoCrate 编译 Cargo crate
func buildCargoCrate(ctx context.Context, options CargoBuildOptions) (CargoBuildResult, error) {
	start := time.Now()
	
	result := CargoBuildResult{
		Success: false,
	}
	
	// 1. 检查 cargo 是否可用
	available, cargoVersion, err := checkCargoAvailable()
	if !available {
		return result, fmt.Errorf("cargo is not installed or not in PATH: %w", err)
	}
	
	fmt.Printf("Using %s\n", cargoVersion)
	
	// 2. 检查源码目录
	if _, err := os.Stat(options.SourceDir); os.IsNotExist(err) {
		return result, fmt.Errorf("source directory not found: %s", options.SourceDir)
	}
	
	cargoToml := filepath.Join(options.SourceDir, "Cargo.toml")
	if _, err := os.Stat(cargoToml); os.IsNotExist(err) {
		return result, fmt.Errorf("Cargo.toml not found in: %s", options.SourceDir)
	}
	
	// 3. 构建命令参数
	args := []string{"build"}
	
	if options.Release {
		args = append(args, "--release")
	}
	
	if options.Target != "" {
		args = append(args, "--target", options.Target)
	}
	
	if options.AllFeatures {
		args = append(args, "--all-features")
	} else if options.NoDefault {
		args = append(args, "--no-default-features")
		if len(options.Features) > 0 {
			args = append(args, "--features", strings.Join(options.Features, ","))
		}
	} else if len(options.Features) > 0 {
		args = append(args, "--features", strings.Join(options.Features, ","))
	}
	
	if options.BinName != "" {
		args = append(args, "--bin", options.BinName)
	}
	
	if options.Jobs > 0 {
		args = append(args, "--jobs", fmt.Sprintf("%d", options.Jobs))
	}
	
	if options.Verbose {
		args = append(args, "--verbose")
	}
	
	// 4. 执行编译
	fmt.Printf("Building in %s...\n", options.SourceDir)
	fmt.Printf("Command: cargo %s\n", strings.Join(args, " "))
	
	cmd := exec.CommandContext(ctx, "cargo", args...)
	cmd.Dir = options.SourceDir
	cmd.Env = os.Environ()
	
	output, err := cmd.CombinedOutput()
	result.Output = string(output)
	result.Duration = time.Since(start)
	
	if err != nil {
		return result, fmt.Errorf("cargo build failed: %w\n%s", err, result.Output)
	}
	
	result.Success = true
	
	// 5. 查找编译后的二进制文件
	buildProfile := "debug"
	if options.Release {
		buildProfile = "release"
	}
	
	targetDir := filepath.Join(options.SourceDir, "target")
	if options.Target != "" {
		targetDir = filepath.Join(targetDir, options.Target, buildProfile)
	} else {
		targetDir = filepath.Join(targetDir, buildProfile)
	}
	
	// 查找可执行文件
	binaries, err := findExecutables(targetDir)
	if err != nil || len(binaries) == 0 {
		return result, fmt.Errorf("no binary found in: %s", targetDir)
	}
	
	// 如果指定了二进制名称，查找匹配的
	if options.BinName != "" {
		for _, bin := range binaries {
			base := filepath.Base(bin)
			name := strings.TrimSuffix(base, filepath.Ext(base))
			if name == options.BinName {
				result.BinaryPath = bin
				break
			}
		}
		if result.BinaryPath == "" {
			return result, fmt.Errorf("binary %s not found", options.BinName)
		}
	} else {
		// 使用第一个找到的二进制
		result.BinaryPath = binaries[0]
	}
	
	fmt.Printf("Binary built: %s\n", result.BinaryPath)
	
	return result, nil
}

// findExecutables 查找目录中的可执行文件
func findExecutables(dir string) ([]string, error) {
	var executables []string
	
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		name := entry.Name()
		
		// Windows: 查找 .exe 文件
		// Unix: 查找有执行权限的文件
		if runtime.GOOS == "windows" {
			if strings.HasSuffix(strings.ToLower(name), ".exe") {
				executables = append(executables, filepath.Join(dir, name))
			}
		} else {
			// Unix 系统检查执行权限
			info, err := entry.Info()
			if err != nil {
				continue
			}
			if info.Mode()&0111 != 0 { // 有执行权限
				executables = append(executables, filepath.Join(dir, name))
			}
		}
	}
	
	return executables, nil
}

// installCargoBinary 安装编译后的二进制文件
func installCargoBinary(binaryPath string, installDir string) (string, error) {
	// 1. 创建安装目录
	if err := os.MkdirAll(installDir, 0o755); err != nil {
		return "", fmt.Errorf("create install directory: %w", err)
	}
	
	// 2. 确定目标路径
	binaryName := filepath.Base(binaryPath)
	targetPath := filepath.Join(installDir, binaryName)
	
	// 3. 复制二进制文件
	sourceFile, err := os.Open(binaryPath)
	if err != nil {
		return "", fmt.Errorf("open source binary: %w", err)
	}
	defer sourceFile.Close()
	
	targetFile, err := os.Create(targetPath)
	if err != nil {
		return "", fmt.Errorf("create target binary: %w", err)
	}
	defer targetFile.Close()
	
	if _, err := sourceFile.WriteTo(targetFile); err != nil {
		return "", fmt.Errorf("copy binary: %w", err)
	}
	
	// 4. 设置执行权限（Unix）
	if runtime.GOOS != "windows" {
		if err := os.Chmod(targetPath, 0o755); err != nil {
			return "", fmt.Errorf("set executable permission: %w", err)
		}
	}
	
	fmt.Printf("Installed binary to: %s\n", targetPath)
	
	return targetPath, nil
}

// getCargoInstallDir 获取 cargo 二进制文件的安装目录
func getCargoInstallDir() string {
	// 优先使用 CARGO_INSTALL_ROOT
	if root := os.Getenv("CARGO_INSTALL_ROOT"); root != "" {
		return filepath.Join(root, "bin")
	}
	
	// 默认使用 ~/.cargo/bin
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	
	return filepath.Join(home, ".cargo", "bin")
}

// checkPathContains 检查 PATH 是否包含指定目录
func checkPathContains(dir string) bool {
	pathEnv := os.Getenv("PATH")
	paths := filepath.SplitList(pathEnv)
	
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return false
	}
	
	for _, p := range paths {
		absPath, err := filepath.Abs(p)
		if err != nil {
			continue
		}
		if absPath == absDir {
			return true
		}
	}
	
	return false
}
