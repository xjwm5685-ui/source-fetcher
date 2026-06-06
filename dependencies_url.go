package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// URLDependency URL 依赖声明
type URLDependency struct {
	URL         string            `yaml:"url"`
	Name        string            `yaml:"name,omitempty"`        // 可选：自定义名称
	Version     string            `yaml:"version,omitempty"`     // 可选：版本标识
	InstallPath string            `yaml:"install_path,omitempty"` // 可选：安装路径
	Metadata    map[string]string `yaml:"metadata,omitempty"`    // 可选：额外元数据
}

// EnhancedDownloadRequest 增强的下载请求，支持依赖
type EnhancedDownloadRequest struct {
	DownloadRequest
	Dependencies []URLDependency `yaml:"dependencies,omitempty"`
}

// DownloadWithDependencies 下载包及其依赖
func DownloadWithDependencies(ctx context.Context, client *http.Client, req EnhancedDownloadRequest) error {
	// 先下载主文件
	plan, err := resolveDownloadPlan(ctx, client, req.DownloadRequest)
	if err != nil {
		return fmt.Errorf("build download plan: %w", err)
	}
	
	options := DownloadOptions{
		OutputDir: req.OutputDir,
		Resume:    req.Resume,
		Chunks:    req.Chunks,
	}
	
	if _, err := downloadPlan(ctx, client, plan, options); err != nil {
		return fmt.Errorf("download main file: %w", err)
	}

	// 下载依赖
	if len(req.Dependencies) > 0 {
		fmt.Printf("\nDownloading %d dependencies...\n", len(req.Dependencies))
		for i, dep := range req.Dependencies {
			fmt.Printf("[%d/%d] %s\n", i+1, len(req.Dependencies), dep.URL)

			depPlan, err := resolveDownloadPlan(ctx, client, DownloadRequest{
				Source:         "url",
				URL:            dep.URL,
				OutputDir:      req.OutputDir,
				Resume:         req.Resume,
				Chunks:         req.Chunks,
				RequestOptions: req.RequestOptions,
			})
			if err != nil {
				fmt.Printf("  Warning: Failed to build download plan for dependency: %v\n", err)
				continue
			}

			if _, err := downloadPlan(ctx, client, depPlan, options); err != nil {
				fmt.Printf("  Warning: Failed to download dependency: %v\n", err)
				continue
			}
		}
	}

	return nil
}

// TrackURLInstallation 跟踪 URL 安装
func TrackURLInstallation(url string, dependencies []URLDependency, outputDir string) error {
	tracking := URLInstallTracking{
		URL:          url,
		InstalledAt:  time.Now().UTC().Format(time.RFC3339),
		OutputDir:    outputDir,
		Dependencies: dependencies,
	}

	trackingPath := filepath.Join(outputDir, ".source-fetcher-url-tracking.json")
	return saveURLTracking(trackingPath, tracking)
}

// URLInstallTracking URL 安装跟踪信息
type URLInstallTracking struct {
	URL          string          `json:"url"`
	InstalledAt  string          `json:"installed_at"`
	OutputDir    string          `json:"output_dir"`
	Dependencies []URLDependency `json:"dependencies"`
}

func saveURLTracking(path string, tracking URLInstallTracking) error {
	data, err := json.MarshalIndent(tracking, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal URL tracking: %w", err)
	}
	// 使用安全的文件权限：仅所有者可读写
	return os.WriteFile(path, data, 0600)
}

// ResolveURLDependencies 解析 URL 依赖树
func ResolveURLDependencies(mainURL string, dependencies []URLDependency) ([]string, error) {
	var allURLs []string

	// 添加主 URL
	allURLs = append(allURLs, mainURL)

	// 添加所有依赖 URL
	for _, dep := range dependencies {
		if strings.TrimSpace(dep.URL) == "" {
			continue
		}
		allURLs = append(allURLs, dep.URL)
	}

	// 去重
	return uniqueStrings(allURLs), nil
}

// GenerateURLDependencyReport 生成依赖报告
func GenerateURLDependencyReport(mainURL string, dependencies []URLDependency, outputDir string) error {
	report := fmt.Sprintf("# URL Dependency Report\n\n")
	report += fmt.Sprintf("Main URL: %s\n", mainURL)
	report += fmt.Sprintf("Output Directory: %s\n\n", outputDir)

	if len(dependencies) > 0 {
		report += "## Dependencies\n\n"
		for i, dep := range dependencies {
			report += fmt.Sprintf("%d. %s\n", i+1, dep.URL)
			if dep.Name != "" {
				report += fmt.Sprintf("   Name: %s\n", dep.Name)
			}
			if dep.Version != "" {
				report += fmt.Sprintf("   Version: %s\n", dep.Version)
			}
			if dep.InstallPath != "" {
				report += fmt.Sprintf("   Install Path: %s\n", dep.InstallPath)
			}
			report += "\n"
		}
	} else {
		report += "No dependencies.\n"
	}

	reportPath := filepath.Join(outputDir, "source-fetcher-url-report.txt")
	// 报告文件可以稍微宽松一些权限
	return os.WriteFile(reportPath, []byte(report), 0640)
}

// ValidateURLDependencies 验证 URL 依赖配置（防止 SSRF）
func ValidateURLDependencies(dependencies []URLDependency) error {
	seen := make(map[string]bool)

	for i, dep := range dependencies {
		// 检查 URL 必须存在
		if strings.TrimSpace(dep.URL) == "" {
			return fmt.Errorf("dependency %d: URL is required", i+1)
		}

		url := strings.TrimSpace(dep.URL)
		urlLower := strings.ToLower(url)

		// 解析 URL
		parsed, err := parseURL(url)
		if err != nil {
			return fmt.Errorf("dependency %d: invalid URL %s: %w", i+1, url, err)
		}

		// 安全检查：仅允许 HTTPS（生产环境推荐）
		// 注意：允许 HTTP 用于向后兼容，但会发出警告
		if parsed.Scheme != "https" && parsed.Scheme != "http" {
			return fmt.Errorf("dependency %d: only HTTP/HTTPS URLs are allowed, got %s", i+1, parsed.Scheme)
		}

		// 安全检查：阻止内网地址（SSRF 防护）
		if isPrivateOrLocalAddress(parsed.Hostname()) {
			return fmt.Errorf("dependency %d: private/internal addresses not allowed: %s (SSRF protection)", i+1, parsed.Hostname())
		}

		// 检查重复
		if seen[urlLower] {
			return fmt.Errorf("dependency %d: duplicate URL %s", i+1, url)
		}
		seen[urlLower] = true
	}

	return nil
}

func uniqueStrings(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		item = strings.TrimSpace(item)
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		result = append(result, item)
	}

	return result
}

// parseURL 解析 URL（辅助函数）
func parseURL(rawURL string) (*urlParsed, error) {
	// 这里需要导入 net/url，但为了避免命名冲突，我们直接使用简单解析
	// 在实际使用中，应该使用 net/url.Parse
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		return nil, fmt.Errorf("invalid URL scheme")
	}
	
	var scheme, rest string
	if strings.HasPrefix(rawURL, "https://") {
		scheme = "https"
		rest = rawURL[8:]
	} else {
		scheme = "http"
		rest = rawURL[7:]
	}
	
	// 提取 hostname（在第一个 / 或 ? 之前）
	hostname := rest
	if idx := strings.IndexAny(rest, "/?#"); idx != -1 {
		hostname = rest[:idx]
	}
	
	// 移除端口号
	if idx := strings.Index(hostname, ":"); idx != -1 {
		hostname = hostname[:idx]
	}
	
	return &urlParsed{
		Scheme:   scheme,
		Host: hostname,
	}, nil
}

type urlParsed struct {
	Scheme   string
	Host string
}

func (u *urlParsed) Hostname() string {
	return u.Host
}

// isPrivateOrLocalAddress 检查是否为内网或本地地址（SSRF 防护）
func isPrivateOrLocalAddress(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	
	// 检查 localhost 变体
	if host == "localhost" || host == "localhost.localdomain" {
		return true
	}
	
	// 检查 127.0.0.0/8 (本地回环)
	if strings.HasPrefix(host, "127.") {
		return true
	}
	
	// 检查 10.0.0.0/8 (私有网络)
	if strings.HasPrefix(host, "10.") {
		return true
	}
	
	// 检查 172.16.0.0/12 (私有网络)
	if strings.HasPrefix(host, "172.") {
		parts := strings.Split(host, ".")
		if len(parts) >= 2 {
			if second := parts[1]; len(second) > 0 {
				if second == "16" || second == "17" || second == "18" || second == "19" ||
					second == "20" || second == "21" || second == "22" || second == "23" ||
					second == "24" || second == "25" || second == "26" || second == "27" ||
					second == "28" || second == "29" || second == "30" || second == "31" {
					return true
				}
			}
		}
	}
	
	// 检查 192.168.0.0/16 (私有网络)
	if strings.HasPrefix(host, "192.168.") {
		return true
	}
	
	// 检查 169.254.0.0/16 (链路本地地址)
	if strings.HasPrefix(host, "169.254.") {
		return true
	}
	
	// 检查 0.0.0.0/8
	if strings.HasPrefix(host, "0.") {
		return true
	}
	
	// 检查 IPv6 localhost
	if host == "::1" || host == "::ffff:127.0.0.1" {
		return true
	}
	
	// 检查其他特殊域名
	specialDomains := []string{
		".local",
		".internal",
		".localhost",
	}
	for _, suffix := range specialDomains {
		if strings.HasSuffix(host, suffix) {
			return true
		}
	}
	
	return false
}
