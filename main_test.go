package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPickNPMVersionUsesLatestTag(t *testing.T) {
	metadata := npmPackageMetadata{
		Name: "react",
		DistTags: map[string]string{
			"latest": "19.1.0",
			"next":   "20.0.0-rc.1",
		},
		Versions: map[string]npmVersionDetails{
			"19.1.0":      {Version: "19.1.0"},
			"20.0.0-rc.1": {Version: "20.0.0-rc.1"},
		},
	}

	version, _, err := pickNPMVersion(metadata, "latest")
	if err != nil {
		t.Fatalf("pickNPMVersion returned error: %v", err)
	}
	if version != "19.1.0" {
		t.Fatalf("expected latest to resolve to 19.1.0, got %q", version)
	}
}

func TestPickNPMVersionUsesDistTagAlias(t *testing.T) {
	metadata := npmPackageMetadata{
		Name: "react",
		DistTags: map[string]string{
			"next": "20.0.0-rc.1",
		},
		Versions: map[string]npmVersionDetails{
			"20.0.0-rc.1": {Version: "20.0.0-rc.1"},
		},
	}

	version, _, err := pickNPMVersion(metadata, "next")
	if err != nil {
		t.Fatalf("pickNPMVersion returned error: %v", err)
	}
	if version != "20.0.0-rc.1" {
		t.Fatalf("expected next tag to resolve to 20.0.0-rc.1, got %q", version)
	}
}

func TestRewriteMirrorURLReplacesHost(t *testing.T) {
	got := rewriteMirrorURL(
		"https://registry.npmjs.org/react/-/react-19.1.0.tgz",
		"https://registry.npmmirror.com",
	)
	want := "https://registry.npmmirror.com/react/-/react-19.1.0.tgz"
	if got != want {
		t.Fatalf("unexpected rewritten url: got %q want %q", got, want)
	}
}

func TestRewriteMirrorURLSupportsPathBasedMirror(t *testing.T) {
	got := rewriteMirrorURL(
		"https://raw.githubusercontent.com/microsoft/winget-pkgs/master/manifests/m/Microsoft/WindowsTerminal/1.16.10261.0/Microsoft.WindowsTerminal.installer.yaml",
		"https://cdn.jsdelivr.net/gh/microsoft/winget-pkgs@master",
	)
	want := "https://cdn.jsdelivr.net/gh/microsoft/winget-pkgs@master/manifests/m/Microsoft/WindowsTerminal/1.16.10261.0/Microsoft.WindowsTerminal.installer.yaml"
	if got != want {
		t.Fatalf("unexpected path-based rewrite: got %q want %q", got, want)
	}
}

func TestRewriteMirrorURLDoesNotDuplicateMirrorPath(t *testing.T) {
	got := rewriteMirrorURL(
		"https://repo.huaweicloud.com/repository/npm/react/-/react-19.2.0.tgz",
		"https://repo.huaweicloud.com/repository/npm",
	)
	want := "https://repo.huaweicloud.com/repository/npm/react/-/react-19.2.0.tgz"
	if got != want {
		t.Fatalf("unexpected duplicated path rewrite: got %q want %q", got, want)
	}
}

func TestFetchNPMMetadataWithFallbackUsesNextMirror(t *testing.T) {
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "timeout", http.StatusGatewayTimeout)
	}))
	defer primary.Close()

	secondary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/react" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(npmPackageMetadata{
			Name:     "react",
			DistTags: map[string]string{"latest": "19.2.0"},
			Versions: map[string]npmVersionDetails{"19.2.0": {Version: "19.2.0"}},
		})
	}))
	defer secondary.Close()

	original := builtinMirrors["npm"]
	builtinMirrors["npm"] = []Mirror{
		{Source: "npm", Name: "primary", BaseURL: primary.URL},
		{Source: "npm", Name: "secondary", BaseURL: secondary.URL},
	}
	defer func() { builtinMirrors["npm"] = original }()

	mirror, metadata, err := fetchNPMMetadataWithFallback(context.Background(), &http.Client{}, "", "react", RequestOptions{})
	if err != nil {
		t.Fatalf("fetchNPMMetadataWithFallback returned error: %v", err)
	}
	if mirror.Name != "secondary" || metadata.Name != "react" {
		t.Fatalf("expected fallback mirror to succeed, got mirror=%q metadata=%+v", mirror.Name, metadata)
	}
}

func TestResolvePyPIMapsSHA256DigestToIntegrity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/demo/json" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"info": map[string]any{
				"name":    "demo",
				"version": "1.2.3",
				"summary": "demo package",
			},
			"releases": map[string]any{
				"1.2.3": []map[string]any{
					{
						"filename":    "demo-1.2.3.tar.gz",
						"url":         "https://files.pythonhosted.org/packages/demo-1.2.3.tar.gz",
						"packagetype": "sdist",
						"digests": map[string]any{
							"sha256": strings.Repeat("ab", 32),
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	original := builtinMirrors["pip"]
	builtinMirrors["pip"] = []Mirror{{Source: "pip", Name: "test-pypi", BaseURL: server.URL}}
	defer func() { builtinMirrors["pip"] = original }()

	plan, err := resolvePyPI(context.Background(), server.Client(), DownloadRequest{
		Source: "pip",
		Name:   "demo",
	})
	if err != nil {
		t.Fatalf("resolvePyPI returned error: %v", err)
	}
	if plan.Shasum != "" {
		t.Fatalf("expected PyPI sha256 not to be stored as shasum, got %q", plan.Shasum)
	}
	want := "sha256-" + base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0xab}, 32))
	if plan.Integrity != want {
		t.Fatalf("expected integrity %q, got %q", want, plan.Integrity)
	}
}

func TestSearchNPMFallsBackWhenFirstMirrorReturnsNoResults(t *testing.T) {
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(npmSearchResponse{})
	}))
	defer primary.Close()

	secondary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload npmSearchResponse
		payload.Objects = make([]struct {
			Package struct {
				Name        string `json:"name"`
				Version     string `json:"version"`
				Description string `json:"description"`
				Links       struct {
					NPM        string `json:"npm"`
					Homepage   string `json:"homepage"`
					Repository string `json:"repository"`
				} `json:"links"`
			} `json:"package"`
		}, 1)
		payload.Objects[0].Package.Name = "react"
		payload.Objects[0].Package.Version = "19.2.0"
		payload.Objects[0].Package.Description = "ui"
		payload.Objects[0].Package.Links.NPM = "https://www.npmjs.com/package/react"
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer secondary.Close()

	original := builtinMirrors["npm"]
	builtinMirrors["npm"] = []Mirror{
		{Source: "npm", Name: "primary", BaseURL: primary.URL},
		{Source: "npm", Name: "secondary", BaseURL: secondary.URL},
	}
	defer func() { builtinMirrors["npm"] = original }()

	results, err := searchNPM(context.Background(), &http.Client{}, SearchRequest{Source: "npm", Query: "react", Limit: 5})
	if err != nil {
		t.Fatalf("searchNPM returned error: %v", err)
	}
	if len(results) != 1 || results[0].Identifier != "react" {
		t.Fatalf("expected fallback search result, got %+v", results)
	}
}

func TestSearchNPMFallsBackWhenExplicitMirrorReturnsNoResults(t *testing.T) {
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(npmSearchResponse{})
	}))
	defer primary.Close()

	secondary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload npmSearchResponse
		payload.Objects = make([]struct {
			Package struct {
				Name        string `json:"name"`
				Version     string `json:"version"`
				Description string `json:"description"`
				Links       struct {
					NPM        string `json:"npm"`
					Homepage   string `json:"homepage"`
					Repository string `json:"repository"`
				} `json:"links"`
			} `json:"package"`
		}, 1)
		payload.Objects[0].Package.Name = "react"
		payload.Objects[0].Package.Version = "19.2.0"
		payload.Objects[0].Package.Description = "ui"
		payload.Objects[0].Package.Links.NPM = "https://www.npmjs.com/package/react"
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer secondary.Close()

	original := builtinMirrors["npm"]
	builtinMirrors["npm"] = []Mirror{
		{Source: "npm", Name: "primary", BaseURL: primary.URL},
		{Source: "npm", Name: "secondary", BaseURL: secondary.URL},
	}
	defer func() { builtinMirrors["npm"] = original }()

	results, err := searchNPM(context.Background(), &http.Client{}, SearchRequest{
		Source: "npm",
		Query:  "react",
		Limit:  5,
		Mirror: "primary",
	})
	if err != nil {
		t.Fatalf("searchNPM returned error: %v", err)
	}
	if len(results) != 1 || results[0].Identifier != "react" {
		t.Fatalf("expected explicit mirror fallback search result, got %+v", results)
	}
}

func TestWingetPackagePath(t *testing.T) {
	got, err := wingetPackagePath("Microsoft.PowerToys")
	if err != nil {
		t.Fatalf("wingetPackagePath returned error: %v", err)
	}
	want := "manifests/m/Microsoft/PowerToys"
	if got != want {
		t.Fatalf("unexpected winget path: got %q want %q", got, want)
	}
}

func TestSelectWingetInstallerPrefersX64(t *testing.T) {
	installers := []WingetInstaller{
		{Architecture: "x86", InstallerURL: "https://example.com/x86.exe"},
		{Architecture: "x64", InstallerURL: "https://example.com/x64.exe", Scope: "machine"},
		{Architecture: "neutral", InstallerURL: "https://example.com/neutral.exe"},
	}

	index, installer, err := selectWingetInstaller(installers, "", -1)
	if err != nil {
		t.Fatalf("selectWingetInstaller returned error: %v", err)
	}
	if index != 1 {
		t.Fatalf("expected x64 installer index 1, got %d", index)
	}
	if installer.InstallerURL != "https://example.com/x64.exe" {
		t.Fatalf("unexpected installer selected: %+v", installer)
	}
}

func TestSelectWingetInstallerRespectsArchOverride(t *testing.T) {
	installers := []WingetInstaller{
		{Architecture: "x64", InstallerURL: "https://example.com/x64.exe"},
		{Architecture: "arm64", InstallerURL: "https://example.com/arm64.exe"},
	}

	index, installer, err := selectWingetInstaller(installers, "arm64", -1)
	if err != nil {
		t.Fatalf("selectWingetInstaller returned error: %v", err)
	}
	if index != 1 || installer.InstallerURL != "https://example.com/arm64.exe" {
		t.Fatalf("unexpected arm64 selection: index=%d installer=%+v", index, installer)
	}
}

func TestCompareVersionish(t *testing.T) {
	if compareVersionish("1.10.0", "1.2.0") <= 0 {
		t.Fatal("expected 1.10.0 to be greater than 1.2.0")
	}
	if compareVersionish("2025.11.1", "2025.9.9") <= 0 {
		t.Fatal("expected 2025.11.1 to be greater than 2025.9.9")
	}
	if compareVersionish("1.0.0", "1.0.0") != 0 {
		t.Fatal("expected same versions to compare equal")
	}
	if compareVersionish("1.0.0", "1.0.0-preview") <= 0 {
		t.Fatal("expected stable version to be greater than prerelease")
	}
	if compareVersionish("1.0.0-rc.10", "1.0.0-rc.2") <= 0 {
		t.Fatal("expected rc.10 to be greater than rc.2")
	}
}

func TestFilenameFromURLOrFallback(t *testing.T) {
	got := filenameFromURLOrFallback("https://example.com/files/demo.zip", "fallback.bin")
	if got != "demo.zip" {
		t.Fatalf("expected demo.zip, got %q", got)
	}
}

func TestEnsureUniquePathAddsSuffix(t *testing.T) {
	dir := t.TempDir()
	first := filepath.Join(dir, "demo.zip")
	if err := os.WriteFile(first, []byte("data"), 0o644); err != nil {
		t.Fatalf("write first file: %v", err)
	}

	got, err := ensureUniquePath(first)
	if err != nil {
		t.Fatalf("ensureUniquePath returned error: %v", err)
	}
	want := filepath.Join(dir, "demo(1).zip")
	if got != want {
		t.Fatalf("unexpected unique path: got %q want %q", got, want)
	}
}

func TestLoadBatchConfigResolvesRelativePathsAndMirrors(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "source-fetcher.yaml")
	content := `
output_dir: downloads
timeout: 45s
mirrors:
  npm: npmmirror
install_defaults:
  scripts_policy: root
  allow_scripts:
    - esbuild
    - sharp
downloads:
  - source: npm
    name: react
  - source: url
    url: https://example.com/a.zip
    output_dir: custom
`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := loadBatchConfig(configPath)
	if err != nil {
		t.Fatalf("loadBatchConfig returned error: %v", err)
	}
	if cfg.OutputDir != filepath.Join(dir, "downloads") {
		t.Fatalf("unexpected output dir: %q", cfg.OutputDir)
	}
	if cfg.Mirrors["npm"] != "npmmirror" {
		t.Fatalf("expected npm mirror default, got %+v", cfg.Mirrors)
	}
	if cfg.InstallDefaults.ScriptsPolicy != "root" {
		t.Fatalf("expected install default scripts policy, got %+v", cfg.InstallDefaults)
	}
	if !equalStringSlices(cfg.InstallDefaults.AllowScripts, []string{"esbuild", "sharp"}) {
		t.Fatalf("expected normalized allow_scripts, got %+v", cfg.InstallDefaults.AllowScripts)
	}
	if cfg.Downloads[1].OutputDir != filepath.Join(dir, "custom") {
		t.Fatalf("unexpected task output dir: %q", cfg.Downloads[1].OutputDir)
	}
}

func TestLoadBatchConfigWarnsOnUnknownFields(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "source-fetcher.yaml")
	content := `
output_dir: downloads
extra_root: true
downloads:
  - source: npm
    name: react
    extra_task: ignored
`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	warnings := captureStderrMain(t, func() {
		cfg, err := loadBatchConfig(configPath)
		if err != nil {
			t.Fatalf("loadBatchConfig returned error: %v", err)
		}
		if len(cfg.Downloads) != 1 || cfg.Downloads[0].Name != "react" {
			t.Fatalf("expected config to still load known fields, got %+v", cfg)
		}
	})
	if !strings.Contains(warnings, "contains unknown field") {
		t.Fatalf("expected unknown field warning, got %q", warnings)
	}
	if !strings.Contains(warnings, "extra_root") || !strings.Contains(warnings, "extra_task") {
		t.Fatalf("expected warning to mention unknown fields, got %q", warnings)
	}
}

func TestLoadBatchConfigStillFailsOnInvalidKnownFieldType(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "source-fetcher.yaml")
	content := `
downloads:
  - source: npm
    name: react
    chunks: nope
`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := loadBatchConfig(configPath)
	if err == nil || !strings.Contains(err.Error(), "cannot unmarshal") {
		t.Fatalf("expected invalid type error, got %v", err)
	}
}

func TestBatchConfigApplyTaskDefaults(t *testing.T) {
	cfg := BatchConfig{
		OutputDir: `D:\downloads`,
		Mirrors: map[string]string{
			"npm": "npmmirror",
		},
	}

	got := cfg.ApplyTaskDefaults(DownloadRequest{
		Source: "npm",
		Name:   "react",
	})
	if got.OutputDir != `D:\downloads` {
		t.Fatalf("expected default output dir, got %q", got.OutputDir)
	}
	if got.Mirror != "npmmirror" {
		t.Fatalf("expected default npm mirror, got %q", got.Mirror)
	}
}

func TestBatchConfigApplyInstallDefaultsUsesInstallDefaults(t *testing.T) {
	cfg := BatchConfig{
		OutputDir: `D:\downloads`,
		Mirrors: map[string]string{
			"npm": "npmmirror",
		},
		InstallDefaults: InstallDefaultsConfig{
			ScriptsPolicy: "root",
			AllowScripts:  []string{"esbuild", "sharp"},
		},
	}

	got := cfg.ApplyInstallDefaults(InstallRequest{
		Source: "npm",
		Name:   "react",
	})
	if got.OutputDir != `D:\downloads` {
		t.Fatalf("expected default output dir, got %q", got.OutputDir)
	}
	if got.Mirror != "npmmirror" {
		t.Fatalf("expected default npm mirror, got %q", got.Mirror)
	}
	if got.ScriptsPolicy != "root" {
		t.Fatalf("expected default scripts policy, got %q", got.ScriptsPolicy)
	}
	if !equalStringSlices(got.AllowScripts, []string{"esbuild", "sharp"}) {
		t.Fatalf("expected default allow scripts, got %+v", got.AllowScripts)
	}
}

func TestResolveInstallScriptSettingsCLIOverridesYAML(t *testing.T) {
	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	fs.String("scripts", "none", "")
	fs.String("allow-scripts", "", "")
	if err := fs.Parse([]string{"--scripts", "all", "--allow-scripts", "sharp,esbuild"}); err != nil {
		t.Fatalf("parse flags: %v", err)
	}

	scriptsPolicy, allowScripts := resolveInstallScriptSettings(fs, InstallDefaultsConfig{
		ScriptsPolicy: "root",
		AllowScripts:  []string{"vite"},
	}, "all", "sharp,esbuild")
	if scriptsPolicy != "all" {
		t.Fatalf("expected cli scripts policy override, got %q", scriptsPolicy)
	}
	if !equalStringSlices(allowScripts, []string{"esbuild", "sharp"}) {
		t.Fatalf("expected cli allow scripts override, got %+v", allowScripts)
	}
}

func TestResolveInstallScriptSettingsUsesYAMLWhenCLIUnset(t *testing.T) {
	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	fs.String("scripts", "none", "")
	fs.String("allow-scripts", "", "")
	if err := fs.Parse(nil); err != nil {
		t.Fatalf("parse flags: %v", err)
	}

	scriptsPolicy, allowScripts := resolveInstallScriptSettings(fs, InstallDefaultsConfig{
		ScriptsPolicy: "root",
		AllowScripts:  []string{"esbuild"},
	}, "none", "")
	if scriptsPolicy != "root" {
		t.Fatalf("expected yaml scripts policy default, got %q", scriptsPolicy)
	}
	if !equalStringSlices(allowScripts, []string{"esbuild"}) {
		t.Fatalf("expected yaml allow scripts default, got %+v", allowScripts)
	}
}

func TestResolveInstallScriptSettingsEmptyCLIAllowScriptsClearsYAML(t *testing.T) {
	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	fs.String("scripts", "none", "")
	fs.String("allow-scripts", "", "")
	if err := fs.Parse([]string{"--allow-scripts="}); err != nil {
		t.Fatalf("parse flags: %v", err)
	}

	_, allowScripts := resolveInstallScriptSettings(fs, InstallDefaultsConfig{
		ScriptsPolicy: "root",
		AllowScripts:  []string{"esbuild"},
	}, "none", "")
	if len(allowScripts) != 0 {
		t.Fatalf("expected explicit empty cli allow-scripts to clear yaml defaults, got %+v", allowScripts)
	}
}

func TestWarnIgnoredAllowScriptsWarnsWhenScriptsNone(t *testing.T) {
	var out bytes.Buffer
	warnIgnoredAllowScripts(&out, "none", []string{"sharp", "esbuild", "sharp"})
	text := out.String()
	if !strings.Contains(text, "--allow-scripts has no effect when --scripts=none") {
		t.Fatalf("expected ignored allow-scripts warning, got %q", text)
	}
	if !strings.Contains(text, "esbuild, sharp") {
		t.Fatalf("expected normalized package list in warning, got %q", text)
	}
}

func TestWarnIgnoredAllowScriptsStaysSilentWhenScriptsEnabled(t *testing.T) {
	var out bytes.Buffer
	warnIgnoredAllowScripts(&out, "root", []string{"esbuild"})
	if out.Len() != 0 {
		t.Fatalf("expected no warning when scripts are enabled, got %q", out.String())
	}
}

func TestTaskDisplayName(t *testing.T) {
	if got := taskDisplayName(DownloadRequest{Source: "winget", PackageID: "Microsoft.PowerToys"}); got != "winget:Microsoft.PowerToys" {
		t.Fatalf("unexpected winget display name: %q", got)
	}
	if got := taskDisplayName(DownloadRequest{Source: "npm", Name: "react"}); got != "npm:react" {
		t.Fatalf("unexpected npm display name: %q", got)
	}
}

func TestWingetSearchInfoFromPath(t *testing.T) {
	identifier, version, ok := wingetSearchInfoFromPath("manifests/m/Microsoft/PowerToys/0.99.1/Microsoft.PowerToys.installer.yaml")
	if !ok {
		t.Fatal("expected winget path to parse")
	}
	if identifier != "Microsoft.PowerToys" {
		t.Fatalf("unexpected identifier: %q", identifier)
	}
	if version != "0.99.1" {
		t.Fatalf("unexpected version: %q", version)
	}
}

func TestTruncateText(t *testing.T) {
	got := truncateText("abcdefghijklmnopqrstuvwxyz", 10)
	if got != "abcdefghi..." {
		t.Fatalf("unexpected truncated text: %q", got)
	}
}

func TestFormatByteSize(t *testing.T) {
	if got := formatByteSize(999); got != "999B" {
		t.Fatalf("expected 999B, got %q", got)
	}
	if got := formatByteSize(1536); got != "1.5KB" {
		t.Fatalf("expected 1.5KB, got %q", got)
	}
	if got := formatByteSize(5 * 1024 * 1024); got != "5.0MB" {
		t.Fatalf("expected 5.0MB, got %q", got)
	}
}

func TestFormatDownloadProgressLineWithKnownSize(t *testing.T) {
	line := formatDownloadProgressLine("Microsoft.WindowsTerminal_super_long_name.exe", 512*1024, 1024*1024, time.Second)
	if !strings.Contains(line, "50.0%") {
		t.Fatalf("expected percentage in progress line, got %q", line)
	}
	if !strings.Contains(line, "512.0KB") || !strings.Contains(line, "1.0MB") {
		t.Fatalf("expected transferred and total sizes, got %q", line)
	}
	if !strings.Contains(line, "Downloading Microsoft.WindowsTermin...") {
		t.Fatalf("expected truncated file name, got %q", line)
	}
}

func TestFormatDownloadProgressLineWithoutKnownSize(t *testing.T) {
	line := formatDownloadProgressLine("demo.zip", 2048, -1, 2*time.Second)
	if strings.Contains(line, "%") {
		t.Fatalf("did not expect percentage for unknown total, got %q", line)
	}
	if !strings.Contains(line, "2.0KB") || !strings.Contains(line, "1.0KB/s") {
		t.Fatalf("expected size and speed in progress line, got %q", line)
	}
}

func TestUniqueLatestSearchResultsKeepsNewestVersionPerPackage(t *testing.T) {
	results := []SearchResult{
		{Source: "choco", Identifier: "Git", Version: "2.49.0", Description: "old"},
		{Source: "choco", Identifier: "Git", Version: "2.51.0", Description: "new"},
		{Source: "choco", Identifier: "NodeJS", Version: "22.0.0"},
	}

	got := uniqueLatestSearchResults(results, 10)
	if len(got) != 2 {
		t.Fatalf("expected 2 unique results, got %d", len(got))
	}
	if got[0].Identifier != "Git" || got[0].Version != "2.51.0" || got[0].Description != "new" {
		t.Fatalf("unexpected deduped git result: %+v", got[0])
	}
	if got[1].Identifier != "NodeJS" || got[1].Version != "22.0.0" {
		t.Fatalf("unexpected second result: %+v", got[1])
	}
}

func TestUniqueLatestSearchResultsAppliesLimitAfterDedup(t *testing.T) {
	results := []SearchResult{
		{Source: "npm", Identifier: "react", Version: "19.2.0"},
		{Source: "npm", Identifier: "react", Version: "19.1.0"},
		{Source: "npm", Identifier: "react-dom", Version: "19.2.0"},
	}

	got := uniqueLatestSearchResults(results, 1)
	if len(got) != 1 {
		t.Fatalf("expected limited result set, got %d", len(got))
	}
	if got[0].Identifier != "react" || got[0].Version != "19.2.0" {
		t.Fatalf("unexpected limited result: %+v", got[0])
	}
}

func TestDownloadRequestFromSearchResultMapsNPM(t *testing.T) {
	req, err := downloadRequestFromSearchResult([]SearchResult{
		{Source: "npm", Identifier: "react", Version: "19.2.6"},
	}, 1, SearchPickOptions{
		Mirror:    "npmmirror",
		OutputDir: `D:\downloads`,
	})
	if err != nil {
		t.Fatalf("downloadRequestFromSearchResult returned error: %v", err)
	}
	if req.Source != "npm" || req.Name != "react" || req.Version != "19.2.6" {
		t.Fatalf("unexpected npm request: %+v", req)
	}
	if req.Mirror != "npmmirror" || req.OutputDir != `D:\downloads` {
		t.Fatalf("unexpected npm options: %+v", req)
	}
}

func TestDownloadRequestFromSearchResultMapsWinget(t *testing.T) {
	req, err := downloadRequestFromSearchResult([]SearchResult{
		{Source: "winget", Identifier: "Microsoft.WindowsTerminal", Version: "1.16.10261.0"},
	}, 1, SearchPickOptions{
		OutputDir:      `D:\downloads`,
		Arch:           "x64",
		InstallerIndex: 2,
	})
	if err != nil {
		t.Fatalf("downloadRequestFromSearchResult returned error: %v", err)
	}
	if req.Source != "winget" || req.PackageID != "Microsoft.WindowsTerminal" {
		t.Fatalf("unexpected winget request: %+v", req)
	}
	if req.Arch != "x64" || req.InstallerIndex != 2 {
		t.Fatalf("unexpected winget pick options: %+v", req)
	}
}

func TestDownloadRequestFromSearchResultRejectsOutOfRangePick(t *testing.T) {
	_, err := downloadRequestFromSearchResult([]SearchResult{
		{Source: "npm", Identifier: "react", Version: "19.2.6"},
	}, 2, SearchPickOptions{})
	if err == nil || !strings.Contains(err.Error(), "out of range") {
		t.Fatalf("expected out of range error, got %v", err)
	}
}

func TestDownloadRequestsFromSearchResultsMapsMultiplePicks(t *testing.T) {
	requests, err := downloadRequestsFromSearchResults([]SearchResult{
		{Source: "npm", Identifier: "react", Version: "19.2.6"},
		{Source: "winget", Identifier: "Microsoft.WindowsTerminal", Version: "1.16.10261.0"},
	}, []int{2, 1}, SearchPickOptions{
		OutputDir:      `D:\downloads`,
		Mirror:         "npmmirror",
		Arch:           "x64",
		InstallerIndex: 1,
	})
	if err != nil {
		t.Fatalf("downloadRequestsFromSearchResults returned error: %v", err)
	}
	if len(requests) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(requests))
	}
	if requests[0].Source != "winget" || requests[0].PackageID != "Microsoft.WindowsTerminal" {
		t.Fatalf("unexpected first request: %+v", requests[0])
	}
	if requests[1].Source != "npm" || requests[1].Name != "react" {
		t.Fatalf("unexpected second request: %+v", requests[1])
	}
}

func TestParsePickSpecDeduplicatesAndPreservesOrder(t *testing.T) {
	picks, err := parsePickSpec("2, 1,2,3", 4)
	if err != nil {
		t.Fatalf("parsePickSpec returned error: %v", err)
	}
	if len(picks) != 3 {
		t.Fatalf("expected 3 picks, got %d", len(picks))
	}
	if picks[0] != 2 || picks[1] != 1 || picks[2] != 3 {
		t.Fatalf("unexpected picks: %+v", picks)
	}
}

func TestParsePickSpecRejectsInvalidIndex(t *testing.T) {
	_, err := parsePickSpec("1,foo", 3)
	if err == nil || !strings.Contains(err.Error(), `invalid index "foo"`) {
		t.Fatalf("expected invalid index error, got %v", err)
	}
}

func TestPrintSearchResultsIncludesIndexColumn(t *testing.T) {
	var out bytes.Buffer
	printSearchResults(&out, []SearchResult{
		{Source: "npm", MirrorName: "npmmirror", Identifier: "react", Version: "19.2.6", Description: "React"},
		{Source: "winget", MirrorName: "github-api", Identifier: "Microsoft.WindowsTerminal", Version: "1.16.10261.0", Description: "Windows Terminal"},
	})
	text := out.String()
	if !strings.Contains(text, "INDEX") || !strings.Contains(text, "MIRROR") {
		t.Fatalf("expected INDEX header, got %q", text)
	}
	if !strings.Contains(text, "1     npm      npmmirror") || !strings.Contains(text, "2     winget   github-api") {
		t.Fatalf("expected numbered rows, got %q", text)
	}
}

func TestSearchFallbackWarningReportsMirrorFallback(t *testing.T) {
	warning := searchFallbackWarning("npm", "huaweicloud", []SearchResult{
		{Source: "npm", MirrorName: "npmjs", Identifier: "react"},
	})
	if !strings.Contains(warning, "huaweicloud") || !strings.Contains(warning, "npmjs") {
		t.Fatalf("expected fallback warning to mention mirrors, got %q", warning)
	}
}

func TestPromptSearchPicksReturnsSelectedIndexes(t *testing.T) {
	var out bytes.Buffer
	picks, err := promptSearchPicks(strings.NewReader("2,1\n"), &out, []SearchResult{
		{Source: "npm", Identifier: "react", Version: "19.2.6"},
		{Source: "winget", Identifier: "Microsoft.WindowsTerminal", Version: "1.16.10261.0"},
	})
	if err != nil {
		t.Fatalf("promptSearchPicks returned error: %v", err)
	}
	if len(picks) != 2 || picks[0] != 2 || picks[1] != 1 {
		t.Fatalf("expected picks [2 1], got %+v", picks)
	}
	if !strings.Contains(out.String(), "Choose result indexes like 1 or 1,3,5") {
		t.Fatalf("expected prompt output, got %q", out.String())
	}
}

func TestPromptSearchPicksAllowsRetryAndCancel(t *testing.T) {
	var out bytes.Buffer
	picks, err := promptSearchPicks(strings.NewReader("9\nq\n"), &out, []SearchResult{
		{Source: "npm", Identifier: "react", Version: "19.2.6"},
	})
	if err != nil {
		t.Fatalf("promptSearchPicks returned error: %v", err)
	}
	if len(picks) != 0 {
		t.Fatalf("expected cancel to return no picks, got %+v", picks)
	}
	text := out.String()
	if !strings.Contains(text, "Invalid choice") || !strings.Contains(text, "Cancelled.") {
		t.Fatalf("expected retry and cancel output, got %q", text)
	}
}

func TestSearchWingetReturnsNoResultsWithoutFallbackError(t *testing.T) {
	primaryHits := 0
	fallbackHits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/packages":
			primaryHits++
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"Packages":[],"Total":0}`))
		case "/search/code":
			fallbackHits++
			http.Error(w, "should not call fallback for empty primary result", http.StatusUnauthorized)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	oldPrimary := wingetRunSearchAPIURL
	oldFallback := wingetGitHubCodeSearchURL
	wingetRunSearchAPIURL = server.URL + "/v2/packages"
	wingetGitHubCodeSearchURL = server.URL + "/search/code"
	defer func() {
		wingetRunSearchAPIURL = oldPrimary
		wingetGitHubCodeSearchURL = oldFallback
	}()

	results, err := searchWinget(context.Background(), server.Client(), SearchRequest{
		Query: "does-not-exist",
		Limit: 5,
	})
	if err != nil {
		t.Fatalf("searchWinget returned error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected no results, got %+v", results)
	}
	if primaryHits != 1 {
		t.Fatalf("expected primary api to be called once, got %d", primaryHits)
	}
	if fallbackHits != 0 {
		t.Fatalf("expected fallback api not to be called, got %d", fallbackHits)
	}
}

func TestRunBatchWithOpsUsesFreshContextPerTask(t *testing.T) {
	cfg := BatchConfig{
		OutputDir: t.TempDir(),
		Downloads: []DownloadRequest{
			{Source: "url", URL: "https://example.com/1.zip"},
			{Source: "url", URL: "https://example.com/2.zip"},
		},
	}

	var (
		resolveCalls       int
		secondDeadlineLeft time.Duration
		out                bytes.Buffer
	)
	timeout := 20 * time.Millisecond
	err := runBatchWithOps(
		newHTTPClient(timeout),
		cfg,
		timeout,
		1,
		batchRetryPolicy{},
		true,
		true,
		&out,
		nil,
		func(ctx context.Context, _ *http.Client, req DownloadRequest) (DownloadPlan, error) {
			resolveCalls++
			deadline, ok := ctx.Deadline()
			if !ok {
				return DownloadPlan{}, errors.New("missing deadline")
			}
			switch resolveCalls {
			case 1:
				<-ctx.Done()
				return DownloadPlan{}, ctx.Err()
			case 2:
				if ctx.Err() != nil {
					return DownloadPlan{}, fmt.Errorf("second task context already expired: %w", ctx.Err())
				}
				secondDeadlineLeft = time.Until(deadline)
				return DownloadPlan{
					Source:     "url",
					Identifier: req.URL,
					URL:        req.URL,
					Filename:   "ok.zip",
				}, nil
			default:
				return DownloadPlan{}, fmt.Errorf("unexpected resolve call %d", resolveCalls)
			}
		},
		func(context.Context, *http.Client, DownloadPlan, DownloadOptions) (DownloadResult, error) {
			return DownloadResult{}, nil
		},
	)
	if err == nil || !strings.Contains(err.Error(), "1 failed task") {
		t.Fatalf("expected aggregated batch failure, got %v", err)
	}
	if resolveCalls != 2 {
		t.Fatalf("expected 2 resolve calls, got %d", resolveCalls)
	}
	if secondDeadlineLeft <= 0 {
		t.Fatalf("expected fresh timeout for second task, got %s", secondDeadlineLeft)
	}
}

func TestApplyRequestHeadersAddsGitHubTokenForGitHubAPI(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "secret-token")
	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/rate_limit", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	applyRequestHeaders(req, "application/json", RequestOptions{})

	if got := req.Header.Get("Authorization"); got != "Bearer secret-token" {
		t.Fatalf("expected github auth header, got %q", got)
	}
	if got := req.Header.Get("Accept"); got != "application/json" {
		t.Fatalf("expected accept header, got %q", got)
	}
	if got := req.Header.Get("User-Agent"); !strings.HasPrefix(got, "source-fetcher/") {
		t.Fatalf("expected user-agent prefix, got %q", got)
	}
}

func TestApplyRequestHeadersSkipsGitHubTokenForOtherHosts(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "secret-token")
	req, err := http.NewRequest(http.MethodGet, "https://example.com/file.zip", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	applyRequestHeaders(req, "", RequestOptions{})

	if got := req.Header.Get("Authorization"); got != "" {
		t.Fatalf("did not expect auth header for non-github host, got %q", got)
	}
}

func TestRunBatchWithOpsWritesDownloadResultToProvidedWriter(t *testing.T) {
	cfg := BatchConfig{
		OutputDir: t.TempDir(),
		Downloads: []DownloadRequest{
			{Source: "url", URL: "https://example.com/file.zip"},
		},
	}
	var out bytes.Buffer

	err := runBatchWithOps(
		newHTTPClient(time.Second),
		cfg,
		time.Second,
		1,
		batchRetryPolicy{},
		false,
		false,
		&out,
		nil,
		func(context.Context, *http.Client, DownloadRequest) (DownloadPlan, error) {
			return DownloadPlan{
				Source:     "url",
				Identifier: "https://example.com/file.zip",
				URL:        "https://example.com/file.zip",
				Filename:   "file.zip",
			}, nil
		},
		func(context.Context, *http.Client, DownloadPlan, DownloadOptions) (DownloadResult, error) {
			return DownloadResult{
				Path:     `D:\downloads\file.zip`,
				Size:     123,
				SHA256:   "abc123",
				Duration: time.Second,
			}, nil
		},
	)
	if err != nil {
		t.Fatalf("runBatchWithOps returned error: %v", err)
	}
	text := out.String()
	if !strings.Contains(text, "Saved To: D:\\downloads\\file.zip") {
		t.Fatalf("expected download result in provided writer, got %q", text)
	}
	if !strings.Contains(text, "SHA256: abc123") {
		t.Fatalf("expected hash in provided writer, got %q", text)
	}
}

func TestRunBatchInstallsWithOpsWritesInstallResultToProvidedWriter(t *testing.T) {
	cfg := BatchConfig{
		OutputDir: t.TempDir(),
		Installs: []InstallRequest{
			{Source: "npm", Name: "react", Version: "latest"},
		},
	}
	var out bytes.Buffer

	err := runBatchInstallsWithOps(
		newHTTPClient(time.Second),
		cfg,
		time.Second,
		1,
		batchRetryPolicy{},
		false,
		false,
		&out,
		nil,
		func(context.Context, *http.Client, InstallRequest) (InstallPlan, error) {
			return InstallPlan{
				Source:      "npm",
				Root:        "react",
				Requested:   "latest",
				RootVersion: "19.2.0",
				Packages:    []InstallPackage{{Name: "react", Version: "19.2.0"}},
			}, nil
		},
		func(context.Context, *http.Client, InstallPlan, DownloadOptions) (InstallResult, error) {
			return InstallResult{
				ManifestPath:   filepath.Join(cfg.OutputDir, "source-fetcher-install.json"),
				LockfilePath:   filepath.Join(cfg.OutputDir, "source-fetcher-install.lock.json"),
				RootPath:       filepath.Join(cfg.OutputDir, "node_modules", "react"),
				NodeModulesDir: filepath.Join(cfg.OutputDir, "node_modules"),
				Packages:       []InstalledPackage{{Name: "react", Version: "19.2.0"}},
				Duration:       time.Second,
			}, nil
		},
	)
	if err != nil {
		t.Fatalf("runBatchInstallsWithOps returned error: %v", err)
	}
	text := out.String()
	if !strings.Contains(text, "Manifest:") || !strings.Contains(text, "Lockfile:") {
		t.Fatalf("expected install result in provided writer, got %q", text)
	}
}

func TestRunBatchInstallsWithOpsWarnsWhenAllowScriptsIgnored(t *testing.T) {
	cfg := BatchConfig{
		OutputDir: t.TempDir(),
		Installs: []InstallRequest{
			{Source: "npm", Name: "react", Version: "latest", ScriptsPolicy: "none", AllowScripts: []string{"esbuild"}},
		},
	}
	var out bytes.Buffer

	warnings := captureStderrMain(t, func() {
		err := runBatchInstallsWithOps(
			newHTTPClient(time.Second),
			cfg,
			time.Second,
			1,
			batchRetryPolicy{},
			true,
			false,
			&out,
			nil,
			func(context.Context, *http.Client, InstallRequest) (InstallPlan, error) {
				return InstallPlan{
					Source:        "npm",
					Root:          "react",
					Requested:     "latest",
					RootVersion:   "19.2.0",
					ScriptsPolicy: "none",
					AllowScripts:  []string{"esbuild"},
					Packages:      []InstallPackage{{Name: "react", Version: "19.2.0"}},
				}, nil
			},
			func(context.Context, *http.Client, InstallPlan, DownloadOptions) (InstallResult, error) {
				return InstallResult{}, nil
			},
		)
		if err != nil {
			t.Fatalf("runBatchInstallsWithOps returned error: %v", err)
		}
	})
	if !strings.Contains(warnings, "--allow-scripts has no effect when --scripts=none") {
		t.Fatalf("expected ignored allow-scripts warning, got %q", warnings)
	}
}

func TestRunBatchWithOpsRetriesResolveAndEventuallySucceeds(t *testing.T) {
	cfg := BatchConfig{
		OutputDir: t.TempDir(),
		Downloads: []DownloadRequest{
			{Source: "url", URL: "https://example.com/file.zip"},
		},
	}
	var (
		out          bytes.Buffer
		resolveCalls int
	)

	err := runBatchWithOps(
		newHTTPClient(time.Second),
		cfg,
		time.Second,
		1,
		batchRetryPolicy{Retries: 1},
		false,
		false,
		&out,
		nil,
		func(context.Context, *http.Client, DownloadRequest) (DownloadPlan, error) {
			resolveCalls++
			if resolveCalls == 1 {
				return DownloadPlan{}, errors.New("temporary metadata error")
			}
			return DownloadPlan{
				Source:     "url",
				Identifier: "https://example.com/file.zip",
				URL:        "https://example.com/file.zip",
				Filename:   "file.zip",
			}, nil
		},
		func(context.Context, *http.Client, DownloadPlan, DownloadOptions) (DownloadResult, error) {
			return DownloadResult{
				Path:     `D:\downloads\file.zip`,
				Size:     123,
				SHA256:   "abc123",
				Duration: time.Second,
			}, nil
		},
	)
	if err != nil {
		t.Fatalf("runBatchWithOps returned error: %v", err)
	}
	if resolveCalls != 2 {
		t.Fatalf("expected resolve to be retried once, got %d calls", resolveCalls)
	}
	text := out.String()
	if !strings.Contains(text, "resolve attempt 1/2 failed: temporary metadata error") {
		t.Fatalf("expected retry log in output, got %q", text)
	}
	if !strings.Contains(text, "Saved To: D:\\downloads\\file.zip") {
		t.Fatalf("expected final download output after retry, got %q", text)
	}
}

func TestRunBatchWithOpsRetriesDownloadAndAggregatesFailure(t *testing.T) {
	cfg := BatchConfig{
		OutputDir: t.TempDir(),
		Downloads: []DownloadRequest{
			{Source: "url", URL: "https://example.com/file.zip"},
		},
	}
	var (
		out           bytes.Buffer
		downloadCalls int
	)

	err := runBatchWithOps(
		newHTTPClient(time.Second),
		cfg,
		time.Second,
		1,
		batchRetryPolicy{Retries: 2},
		false,
		true,
		&out,
		nil,
		func(context.Context, *http.Client, DownloadRequest) (DownloadPlan, error) {
			return DownloadPlan{
				Source:     "url",
				Identifier: "https://example.com/file.zip",
				URL:        "https://example.com/file.zip",
				Filename:   "file.zip",
			}, nil
		},
		func(context.Context, *http.Client, DownloadPlan, DownloadOptions) (DownloadResult, error) {
			downloadCalls++
			return DownloadResult{}, errors.New("temporary download error")
		},
	)
	if err == nil || !strings.Contains(err.Error(), "1 failed task") {
		t.Fatalf("expected aggregated failure after retries, got %v", err)
	}
	if downloadCalls != 3 {
		t.Fatalf("expected 3 download attempts, got %d", downloadCalls)
	}
	text := out.String()
	if !strings.Contains(text, "download attempt 2/3 failed: temporary download error") {
		t.Fatalf("expected retry output for second failed attempt, got %q", text)
	}
	if !strings.Contains(text, "download failed: temporary download error") {
		t.Fatalf("expected final failure output, got %q", text)
	}
}

func TestRunBatchInstallsWithOpsRetriesInstallAndContinuesToNextTask(t *testing.T) {
	cfg := BatchConfig{
		OutputDir: t.TempDir(),
		Installs: []InstallRequest{
			{Source: "npm", Name: "broken", Version: "latest"},
			{Source: "npm", Name: "react", Version: "latest"},
		},
	}
	var (
		out           bytes.Buffer
		installCalls  = map[string]int{}
		resolvedTasks []string
	)

	err := runBatchInstallsWithOps(
		newHTTPClient(time.Second),
		cfg,
		time.Second,
		1,
		batchRetryPolicy{Retries: 1},
		false,
		true,
		&out,
		nil,
		func(_ context.Context, _ *http.Client, req InstallRequest) (InstallPlan, error) {
			resolvedTasks = append(resolvedTasks, req.Name)
			return InstallPlan{
				Source:      "npm",
				Root:        req.Name,
				Requested:   req.Version,
				RootVersion: "1.0.0",
				Packages:    []InstallPackage{{Name: req.Name, Version: "1.0.0"}},
			}, nil
		},
		func(_ context.Context, _ *http.Client, plan InstallPlan, _ DownloadOptions) (InstallResult, error) {
			installCalls[plan.Root]++
			if plan.Root == "broken" {
				return InstallResult{}, errors.New("temporary install error")
			}
			return InstallResult{
				ManifestPath:   filepath.Join(cfg.OutputDir, "source-fetcher-install.json"),
				LockfilePath:   filepath.Join(cfg.OutputDir, "source-fetcher-install.lock.json"),
				RootPath:       filepath.Join(cfg.OutputDir, "node_modules", plan.Root),
				NodeModulesDir: filepath.Join(cfg.OutputDir, "node_modules"),
				Packages:       []InstalledPackage{{Name: plan.Root, Version: "1.0.0"}},
				Duration:       time.Second,
			}, nil
		},
	)
	if err == nil || !strings.Contains(err.Error(), "1 failed task") {
		t.Fatalf("expected aggregated install failure, got %v", err)
	}
	if installCalls["broken"] != 2 {
		t.Fatalf("expected broken install to retry once, got %d calls", installCalls["broken"])
	}
	if installCalls["react"] != 1 {
		t.Fatalf("expected second task to continue and install once, got %d calls", installCalls["react"])
	}
	if len(resolvedTasks) != 2 || resolvedTasks[0] != "broken" || resolvedTasks[1] != "react" {
		t.Fatalf("expected both install tasks to be resolved in order, got %+v", resolvedTasks)
	}
	text := out.String()
	if !strings.Contains(text, "install attempt 1/2 failed: temporary install error") {
		t.Fatalf("expected install retry output, got %q", text)
	}
	if !strings.Contains(text, "[install 2/2] react@latest") {
		t.Fatalf("expected second install task to continue after failure, got %q", text)
	}
}

func TestRunBatchWithOpsHonorsJobsConcurrencyLimit(t *testing.T) {
	cfg := BatchConfig{
		OutputDir: t.TempDir(),
		Downloads: []DownloadRequest{
			{Source: "url", URL: "https://example.com/1.zip"},
			{Source: "url", URL: "https://example.com/2.zip"},
			{Source: "url", URL: "https://example.com/3.zip"},
			{Source: "url", URL: "https://example.com/4.zip"},
		},
	}
	var (
		out           bytes.Buffer
		current       int32
		maxConcurrent int32
	)
	started := make(chan string, len(cfg.Downloads))
	release := make(chan struct{})
	done := make(chan error, 1)

	go func() {
		done <- runBatchWithOps(
			newHTTPClient(time.Second),
			cfg,
			time.Second,
			2,
			batchRetryPolicy{},
			true,
			true,
			&out,
			nil,
			func(_ context.Context, _ *http.Client, req DownloadRequest) (DownloadPlan, error) {
				active := atomic.AddInt32(&current, 1)
				defer atomic.AddInt32(&current, -1)
				for {
					seen := atomic.LoadInt32(&maxConcurrent)
					if active <= seen || atomic.CompareAndSwapInt32(&maxConcurrent, seen, active) {
						break
					}
				}
				started <- req.URL
				<-release
				return DownloadPlan{
					Source:     "url",
					Identifier: req.URL,
					URL:        req.URL,
					Filename:   filepath.Base(req.URL),
				}, nil
			},
			func(context.Context, *http.Client, DownloadPlan, DownloadOptions) (DownloadResult, error) {
				return DownloadResult{}, nil
			},
		)
	}()

	for i := 0; i < 2; i++ {
		select {
		case <-started:
		case <-time.After(200 * time.Millisecond):
			t.Fatal("expected first two tasks to start")
		}
	}
	select {
	case third := <-started:
		t.Fatalf("expected jobs=2 to block the third task before release, got %s", third)
	case <-time.After(50 * time.Millisecond):
	}

	close(release)
	if err := <-done; err != nil {
		t.Fatalf("runBatchWithOps returned error: %v", err)
	}
	if got := atomic.LoadInt32(&maxConcurrent); got != 2 {
		t.Fatalf("expected max concurrency 2, got %d", got)
	}
}

func TestRunBatchWithOpsRecoversWorkerPanicAndContinues(t *testing.T) {
	cfg := BatchConfig{
		OutputDir: t.TempDir(),
		Downloads: []DownloadRequest{
			{Source: "url", URL: "https://example.com/panic.zip"},
			{Source: "url", URL: "https://example.com/ok.zip"},
		},
	}
	var (
		out            bytes.Buffer
		downloadedURLs []string
		mu             sync.Mutex
	)

	err := runBatchWithOps(
		newHTTPClient(time.Second),
		cfg,
		time.Second,
		2,
		batchRetryPolicy{},
		false,
		true,
		&out,
		nil,
		func(_ context.Context, _ *http.Client, req DownloadRequest) (DownloadPlan, error) {
			if strings.Contains(req.URL, "panic.zip") {
				panic("boom")
			}
			return DownloadPlan{
				Source:     "url",
				Identifier: req.URL,
				URL:        req.URL,
				Filename:   filepath.Base(req.URL),
			}, nil
		},
		func(_ context.Context, _ *http.Client, plan DownloadPlan, _ DownloadOptions) (DownloadResult, error) {
			mu.Lock()
			downloadedURLs = append(downloadedURLs, plan.URL)
			mu.Unlock()
			return DownloadResult{
				Path:     filepath.Join(`D:\downloads`, filepath.Base(plan.URL)),
				Duration: time.Second,
			}, nil
		},
	)
	if err == nil || !strings.Contains(err.Error(), "1 failed task") {
		t.Fatalf("expected aggregated panic failure, got %v", err)
	}
	if len(downloadedURLs) != 1 || downloadedURLs[0] != "https://example.com/ok.zip" {
		t.Fatalf("expected non-panicking task to still complete, got %+v", downloadedURLs)
	}
	text := out.String()
	if !strings.Contains(text, "batch download task url:https://example.com/panic.zip panicked: boom") {
		t.Fatalf("expected panic to be reported in output, got %q", text)
	}
	if !strings.Contains(text, "Saved To: D:\\downloads\\ok.zip") {
		t.Fatalf("expected successful task output after panic, got %q", text)
	}
}

func TestRunBatchWithOpsJobsReduceElapsedTime(t *testing.T) {
	cfg := BatchConfig{
		OutputDir: t.TempDir(),
		Downloads: []DownloadRequest{
			{Source: "url", URL: "https://example.com/1.zip"},
			{Source: "url", URL: "https://example.com/2.zip"},
			{Source: "url", URL: "https://example.com/3.zip"},
			{Source: "url", URL: "https://example.com/4.zip"},
		},
	}
	taskDelay := 80 * time.Millisecond
	runBatch := func(jobs int) time.Duration {
		start := time.Now()
		err := runBatchWithOps(
			newHTTPClient(time.Second),
			cfg,
			time.Second,
			jobs,
			batchRetryPolicy{},
			true,
			true,
			io.Discard,
			nil,
			func(ctx context.Context, _ *http.Client, req DownloadRequest) (DownloadPlan, error) {
				select {
				case <-ctx.Done():
					return DownloadPlan{}, ctx.Err()
				case <-time.After(taskDelay):
				}
				return DownloadPlan{
					Source:     "url",
					Identifier: req.URL,
					URL:        req.URL,
					Filename:   filepath.Base(req.URL),
				}, nil
			},
			func(context.Context, *http.Client, DownloadPlan, DownloadOptions) (DownloadResult, error) {
				return DownloadResult{}, nil
			},
		)
		if err != nil {
			t.Fatalf("runBatchWithOps returned error: %v", err)
		}
		return time.Since(start)
	}

	serial := runBatch(1)
	parallel := runBatch(2)
	if parallel >= serial {
		t.Fatalf("expected jobs=2 to be faster than jobs=1, got serial=%s parallel=%s", serial, parallel)
	}
	if serial-parallel < taskDelay/2 {
		t.Fatalf("expected a meaningful concurrency win, got serial=%s parallel=%s", serial, parallel)
	}
}

func TestRunBatchWithOpsStopsStartingQueuedTasksAfterFirstFailure(t *testing.T) {
	cfg := BatchConfig{
		OutputDir: t.TempDir(),
		Downloads: []DownloadRequest{
			{Source: "url", URL: "https://example.com/1.zip"},
			{Source: "url", URL: "https://example.com/2.zip"},
			{Source: "url", URL: "https://example.com/3.zip"},
			{Source: "url", URL: "https://example.com/4.zip"},
		},
	}
	var (
		out        bytes.Buffer
		mu         sync.Mutex
		startedSet = map[string]bool{}
	)
	secondStarted := make(chan struct{})
	var secondStartedOnce sync.Once
	err := runBatchWithOps(
		newHTTPClient(time.Second),
		cfg,
		time.Second,
		2,
		batchRetryPolicy{Retries: 2, Backoff: 200 * time.Millisecond},
		true,
		false,
		&out,
		nil,
		func(ctx context.Context, _ *http.Client, req DownloadRequest) (DownloadPlan, error) {
			mu.Lock()
			startedSet[req.URL] = true
			mu.Unlock()
			switch filepath.Base(req.URL) {
			case "1.zip":
				<-secondStarted
				return DownloadPlan{}, errors.New("first failure")
			case "2.zip":
				secondStartedOnce.Do(func() { close(secondStarted) })
				<-ctx.Done()
				return DownloadPlan{}, ctx.Err()
			default:
				return DownloadPlan{
					Source:     "url",
					Identifier: req.URL,
					URL:        req.URL,
					Filename:   filepath.Base(req.URL),
				}, nil
			}
		},
		func(context.Context, *http.Client, DownloadPlan, DownloadOptions) (DownloadResult, error) {
			return DownloadResult{}, nil
		},
	)
	if err == nil || !strings.Contains(err.Error(), "first failure") {
		t.Fatalf("expected first failure to be returned, got %v", err)
	}
	if len(startedSet) != 2 || !startedSet["https://example.com/1.zip"] || !startedSet["https://example.com/2.zip"] {
		t.Fatalf("expected only the first two tasks to start, got %+v", startedSet)
	}
	text := out.String()
	if strings.Contains(text, "[3/4]") || strings.Contains(text, "[4/4]") {
		t.Fatalf("expected queued tasks not to start after first failure, got %q", text)
	}
	if strings.Contains(text, "resolve attempt 2/3 failed: context canceled") {
		t.Fatalf("expected canceled sibling task to stop retrying promptly, got %q", text)
	}
}

func TestRunBatchWithOpsFlushesTaskOutputAsContiguousBlocks(t *testing.T) {
	cfg := BatchConfig{
		OutputDir: t.TempDir(),
		Downloads: []DownloadRequest{
			{Source: "url", URL: "https://example.com/slow.zip"},
			{Source: "url", URL: "https://example.com/fast.zip"},
		},
	}
	var out bytes.Buffer

	err := runBatchWithOps(
		newHTTPClient(time.Second),
		cfg,
		time.Second,
		2,
		batchRetryPolicy{},
		false,
		true,
		&out,
		nil,
		func(ctx context.Context, _ *http.Client, req DownloadRequest) (DownloadPlan, error) {
			delay := 10 * time.Millisecond
			if strings.Contains(req.URL, "slow.zip") {
				delay = 60 * time.Millisecond
			}
			select {
			case <-ctx.Done():
				return DownloadPlan{}, ctx.Err()
			case <-time.After(delay):
			}
			return DownloadPlan{
				Source:     "url",
				Identifier: req.URL,
				URL:        req.URL,
				Filename:   filepath.Base(req.URL),
			}, nil
		},
		func(_ context.Context, _ *http.Client, plan DownloadPlan, _ DownloadOptions) (DownloadResult, error) {
			return DownloadResult{
				Path:     filepath.Join(`D:\downloads`, filepath.Base(plan.URL)),
				Size:     123,
				SHA256:   "hash-" + filepath.Base(plan.URL),
				Duration: time.Second,
			}, nil
		},
	)
	if err != nil {
		t.Fatalf("runBatchWithOps returned error: %v", err)
	}
	text := out.String()
	fastBlock := strings.Join([]string{
		"[2/2] url:https://example.com/fast.zip",
		"Source: url",
		"Identifier: https://example.com/fast.zip",
		"Version: -",
		"Resolved URL: https://example.com/fast.zip",
		"Saved To: D:\\downloads\\fast.zip",
		"Size: 123 bytes",
		"SHA256: hash-fast.zip",
		"Duration: 1s",
		"",
	}, "\n")
	slowBlock := strings.Join([]string{
		"[1/2] url:https://example.com/slow.zip",
		"Source: url",
		"Identifier: https://example.com/slow.zip",
		"Version: -",
		"Resolved URL: https://example.com/slow.zip",
		"Saved To: D:\\downloads\\slow.zip",
		"Size: 123 bytes",
		"SHA256: hash-slow.zip",
		"Duration: 1s",
		"",
	}, "\n")
	if !strings.Contains(text, fastBlock) {
		t.Fatalf("expected fast task output to remain contiguous, got %q", text)
	}
	if !strings.Contains(text, slowBlock) {
		t.Fatalf("expected slow task output to remain contiguous, got %q", text)
	}
}

func TestResolveRequestOptionsUsesBearerProfile(t *testing.T) {
	t.Setenv("PRIVATE_TOKEN", "top-secret")
	cfg := BatchConfig{
		AuthProfiles: map[string]AuthProfile{
			"private": {
				BearerTokenEnv: "PRIVATE_TOKEN",
				Headers: map[string]string{
					"X-Registry": "internal",
				},
			},
		},
	}

	options, err := cfg.ResolveRequestOptions("private")
	if err != nil {
		t.Fatalf("ResolveRequestOptions returned error: %v", err)
	}
	if got := options.Headers["Authorization"]; got != "Bearer top-secret" {
		t.Fatalf("expected bearer auth header, got %q", got)
	}
	if got := options.Headers["X-Registry"]; got != "internal" {
		t.Fatalf("expected custom header, got %q", got)
	}
}

func TestLoadOptionalRuntimeConfigWithPolicyAllowsMissingConfig(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "missing.yaml")

	cfg, err := loadOptionalRuntimeConfigWithPolicy(missingPath, true)
	if err != nil {
		t.Fatalf("loadOptionalRuntimeConfigWithPolicy returned error: %v", err)
	}
	if cfg.OutputDir != "." {
		t.Fatalf("expected default output dir, got %q", cfg.OutputDir)
	}
	if len(cfg.AuthProfiles) != 0 {
		t.Fatalf("expected empty auth profiles, got %+v", cfg.AuthProfiles)
	}
}

func TestLoadOptionalRuntimeConfigWithPolicyRejectsMissingExplicitConfig(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "missing.yaml")

	_, err := loadOptionalRuntimeConfigWithPolicy(missingPath, false)
	if err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("expected missing config error, got %v", err)
	}
}

func TestRunHelpSubcommandsReturnNil(t *testing.T) {
	commands := []string{"download", "search", "batch", "mirrors"}
	for _, command := range commands {
		t.Run(command, func(t *testing.T) {
			if err := run([]string{command, "--help"}); err != nil {
				t.Fatalf("run returned error for %s --help: %v", command, err)
			}
		})
	}
}

func TestDownloadPlanWithProgressCallbackResumesPartialFile(t *testing.T) {
	content := []byte(strings.Repeat("resume-me-", 1024))
	server := newRangeTestServer(content, nil)
	defer server.Close()

	dir := t.TempDir()
	tempPath := filepath.Join(dir, "demo.bin.part")
	prefix := content[:2048]
	if err := os.WriteFile(tempPath, prefix, 0o644); err != nil {
		t.Fatalf("write partial file: %v", err)
	}

	result, err := downloadPlanWithProgressCallback(
		context.Background(),
		server.Client(),
		DownloadPlan{URL: server.URL + "/demo.bin", Filename: "demo.bin"},
		DownloadOptions{OutputDir: dir, Resume: true, Chunks: 1},
		io.Discard,
		nil,
	)
	if err != nil {
		t.Fatalf("downloadPlanWithProgressCallback returned error: %v", err)
	}
	raw, err := os.ReadFile(result.Path)
	if err != nil {
		t.Fatalf("read resumed file: %v", err)
	}
	if !bytes.Equal(raw, content) {
		t.Fatal("resumed download content mismatch")
	}
}

func TestDownloadPlanWithProgressCallbackKeepsPartialFileOnResumeError(t *testing.T) {
	content := []byte(strings.Repeat("resume-me-", 1024))
	dir := t.TempDir()
	tempPath := filepath.Join(dir, "demo.bin.part")

	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.Method {
			case http.MethodHead:
				return &http.Response{
					StatusCode:    http.StatusOK,
					Header:        http.Header{"Accept-Ranges": []string{"bytes"}},
					ContentLength: int64(len(content)),
					Body:          io.NopCloser(bytes.NewReader(nil)),
					Request:       req,
				}, nil
			case http.MethodGet:
				return &http.Response{
					StatusCode:    http.StatusOK,
					ContentLength: int64(len(content)),
					Body: &failingReadCloser{
						data: content[:2048],
						err:  io.ErrUnexpectedEOF,
					},
					Request: req,
				}, nil
			default:
				return nil, fmt.Errorf("unexpected method %s", req.Method)
			}
		}),
	}

	_, err := downloadPlanWithProgressCallback(
		context.Background(),
		client,
		DownloadPlan{URL: "https://example.com/demo.bin", Filename: "demo.bin"},
		DownloadOptions{OutputDir: dir, Resume: true, Chunks: 1},
		io.Discard,
		nil,
	)
	if err == nil {
		t.Fatal("expected download error")
	}
	stat, statErr := os.Stat(tempPath)
	if statErr != nil {
		t.Fatalf("expected partial file to remain, got %v", statErr)
	}
	if stat.Size() != 2048 {
		t.Fatalf("expected partial file size 2048, got %d", stat.Size())
	}
}

func TestDownloadPlanWithProgressCallbackRedownloadsCompletedTempFile(t *testing.T) {
	content := []byte(strings.Repeat("resume-me-", 1024))
	var getHits int
	server := newRangeTestServer(content, func(r *http.Request) {
		if r.Method == http.MethodGet {
			getHits++
		}
	})
	defer server.Close()

	dir := t.TempDir()
	tempPath := filepath.Join(dir, "demo.bin.part")
	if err := os.WriteFile(tempPath, bytes.Repeat([]byte("x"), len(content)), 0o644); err != nil {
		t.Fatalf("write stale temp file: %v", err)
	}

	result, err := downloadPlanWithProgressCallback(
		context.Background(),
		server.Client(),
		DownloadPlan{URL: server.URL + "/demo.bin", Filename: "demo.bin"},
		DownloadOptions{OutputDir: dir, Resume: true, Chunks: 1},
		io.Discard,
		nil,
	)
	if err != nil {
		t.Fatalf("downloadPlanWithProgressCallback returned error: %v", err)
	}
	raw, err := os.ReadFile(result.Path)
	if err != nil {
		t.Fatalf("read downloaded file: %v", err)
	}
	if !bytes.Equal(raw, content) {
		t.Fatal("expected stale temp file to be replaced with fresh download")
	}
	if getHits == 0 {
		t.Fatal("expected at least one GET request to redownload completed temp file")
	}
}

func TestDownloadPlanWithProgressCallbackUsesParallelChunks(t *testing.T) {
	content := []byte(strings.Repeat("chunk-me-", 200000))
	var (
		mu          sync.Mutex
		rangeValues []string
	)
	server := newRangeTestServer(content, func(r *http.Request) {
		if value := strings.TrimSpace(r.Header.Get("Range")); value != "" {
			mu.Lock()
			rangeValues = append(rangeValues, value)
			mu.Unlock()
		}
	})
	defer server.Close()

	dir := t.TempDir()
	result, err := downloadPlanWithProgressCallback(
		context.Background(),
		server.Client(),
		DownloadPlan{URL: server.URL + "/demo.bin", Filename: "demo.bin"},
		DownloadOptions{OutputDir: dir, Resume: false, Chunks: 4},
		io.Discard,
		nil,
	)
	if err != nil {
		t.Fatalf("downloadPlanWithProgressCallback returned error: %v", err)
	}
	raw, err := os.ReadFile(result.Path)
	if err != nil {
		t.Fatalf("read chunked file: %v", err)
	}
	if !bytes.Equal(raw, content) {
		t.Fatal("chunked download content mismatch")
	}
	mu.Lock()
	defer mu.Unlock()
	if len(rangeValues) < 2 {
		t.Fatalf("expected multiple range requests, got %+v", rangeValues)
	}
}

func TestDownloadPlanWithProgressCallbackRejectsIntegrityMismatch(t *testing.T) {
	content := []byte(strings.Repeat("integrity-me-", 512))
	server := newRangeTestServer(content, nil)
	defer server.Close()

	dir := t.TempDir()
	_, err := downloadPlanWithProgressCallback(
		context.Background(),
		server.Client(),
		DownloadPlan{
			URL:        server.URL + "/demo.bin",
			Filename:   "demo.bin",
			Identifier: "demo",
			Version:    "1.0.0",
			Integrity:  sha512Integrity([]byte("bad")),
		},
		DownloadOptions{OutputDir: dir, Resume: false, Chunks: 1},
		io.Discard,
		nil,
	)
	if err == nil || !strings.Contains(err.Error(), "download integrity mismatch") {
		t.Fatalf("expected integrity mismatch error, got %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(dir, "demo.bin.part")); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("expected corrupt temp file to be removed, got %v", statErr)
	}
}

func TestDownloadPlanWithProgressCallbackVerifiesResumedShasum(t *testing.T) {
	content := []byte(strings.Repeat("resume-check-", 1024))
	server := newRangeTestServer(content, nil)
	defer server.Close()

	dir := t.TempDir()
	tempPath := filepath.Join(dir, "demo.bin.part")
	if err := os.WriteFile(tempPath, content[:4096], 0o644); err != nil {
		t.Fatalf("write partial file: %v", err)
	}

	result, err := downloadPlanWithProgressCallback(
		context.Background(),
		server.Client(),
		DownloadPlan{
			URL:        server.URL + "/demo.bin",
			Filename:   "demo.bin",
			Identifier: "demo",
			Version:    "1.0.0",
			Shasum:     sha1Hex(content),
		},
		DownloadOptions{OutputDir: dir, Resume: true, Chunks: 1},
		io.Discard,
		nil,
	)
	if err != nil {
		t.Fatalf("expected resumed download to pass shasum verification, got %v", err)
	}
	raw, err := os.ReadFile(result.Path)
	if err != nil {
		t.Fatalf("read resumed file: %v", err)
	}
	if !bytes.Equal(raw, content) {
		t.Fatal("resumed download content mismatch")
	}
}

func TestDownloadPlanWithProgressCallbackAcceptsMultiAlgorithmIntegrity(t *testing.T) {
	content := []byte(strings.Repeat("multi-integrity-", 512))
	server := newRangeTestServer(content, nil)
	defer server.Close()

	dir := t.TempDir()
	result, err := downloadPlanWithProgressCallback(
		context.Background(),
		server.Client(),
		DownloadPlan{
			URL:        server.URL + "/demo.bin",
			Filename:   "demo.bin",
			Identifier: "demo",
			Version:    "1.0.0",
			Integrity:  "sha256-" + sha512Integrity([]byte("bad"))[7:] + " " + sha512Integrity(content),
		},
		DownloadOptions{OutputDir: dir, Resume: false, Chunks: 1},
		io.Discard,
		nil,
	)
	if err != nil {
		t.Fatalf("expected multi-algorithm integrity to accept matching supported digest, got %v", err)
	}
	if _, err := os.Stat(result.Path); err != nil {
		t.Fatalf("expected downloaded file to exist, got %v", err)
	}
}

func TestVerifyFileIntegrityAcceptsMatchingChecks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "demo.tgz")
	content := []byte(strings.Repeat("integrity-ok-", 256))
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write archive: %v", err)
	}

	digests, err := verifyFileIntegrity(path, sha1Hex(content), formatDigestIntegrity("sha256", sha256Hex(content)), "archive for demo@1.0.0")
	if err != nil {
		t.Fatalf("verifyFileIntegrity returned error: %v", err)
	}
	if digests.SHA1 != sha1Hex(content) {
		t.Fatalf("expected sha1 digest to match, got %q", digests.SHA1)
	}
	if digests.SHA256 != sha256Hex(content) {
		t.Fatalf("expected sha256 digest to match, got %q", digests.SHA256)
	}
	if digests.Size != int64(len(content)) {
		t.Fatalf("expected size %d, got %d", len(content), digests.Size)
	}
}

func TestVerifyFileIntegrityRejectsShasumMismatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "demo.tgz")
	content := []byte(strings.Repeat("integrity-bad-sha1-", 128))
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write archive: %v", err)
	}

	_, err := verifyFileIntegrity(path, sha1Hex([]byte("wrong")), "", "archive for demo@1.0.0")
	if err == nil || !strings.Contains(err.Error(), "archive shasum mismatch for demo@1.0.0") {
		t.Fatalf("expected shasum mismatch error, got %v", err)
	}
}

func TestVerifyFileIntegrityRejectsIntegrityMismatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "demo.tgz")
	content := []byte(strings.Repeat("integrity-bad-sri-", 128))
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write archive: %v", err)
	}

	_, err := verifyFileIntegrity(path, "", sha512Integrity([]byte("wrong")), "archive for demo@1.0.0")
	if err == nil || !strings.Contains(err.Error(), "archive integrity mismatch for demo@1.0.0") {
		t.Fatalf("expected integrity mismatch error, got %v", err)
	}
}

func TestVerifyFileIntegrityAllowsNoChecks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "demo.tgz")
	content := []byte(strings.Repeat("integrity-none-", 64))
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write archive: %v", err)
	}

	digests, err := verifyFileIntegrity(path, "", "", "")
	if err != nil {
		t.Fatalf("verifyFileIntegrity returned error without checks: %v", err)
	}
	if digests.SHA256 != sha256Hex(content) {
		t.Fatalf("expected digest calculation without checks, got %q", digests.SHA256)
	}
}

func TestSearchNPMVerboseLogsFallbackAndSuccess(t *testing.T) {
	original := builtinMirrors["npm"]
	builtinMirrors["npm"] = []Mirror{
		{Source: "npm", Name: "primary", BaseURL: "https://primary.example.com"},
		{Source: "npm", Name: "fallback", BaseURL: "https://fallback.example.com"},
	}
	defer func() {
		builtinMirrors["npm"] = original
	}()

	var log bytes.Buffer
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Host {
			case "primary.example.com":
				return jsonResponse(req, http.StatusOK, `{"objects":[]}`), nil
			case "fallback.example.com":
				return jsonResponse(req, http.StatusOK, `{"objects":[{"package":{"name":"react","version":"19.2.0","description":"React","links":{"npm":"https://npmjs.com/package/react"}}}]}`), nil
			default:
				return nil, fmt.Errorf("unexpected host %s", req.URL.Host)
			}
		}),
	}

	results, err := searchNPM(context.Background(), client, SearchRequest{
		Query:  "react",
		Mirror: "primary",
		Limit:  5,
		RequestOptions: RequestOptions{
			Verbose: newVerboseLogger(true, &log),
		},
	})
	if err != nil {
		t.Fatalf("searchNPM returned error: %v", err)
	}
	if len(results) != 1 || results[0].MirrorName != "fallback" {
		t.Fatalf("expected fallback result, got %+v", results)
	}
	text := log.String()
	if !strings.Contains(text, `[resolve] npm search "react" via primary returned no results`) {
		t.Fatalf("expected empty primary verbose log, got %q", text)
	}
	if !strings.Contains(text, `[resolve] npm search "react" resolved via fallback with 1 result(s)`) {
		t.Fatalf("expected fallback success verbose log, got %q", text)
	}
}

func TestDownloadPlanWithProgressCallbackVerboseLogsIntegrityResult(t *testing.T) {
	content := []byte(strings.Repeat("verbose-integrity-", 256))
	server := newRangeTestServer(content, nil)
	defer server.Close()

	dir := t.TempDir()
	var log bytes.Buffer
	_, err := downloadPlanWithProgressCallback(
		context.Background(),
		server.Client(),
		DownloadPlan{
			URL:        server.URL + "/demo.bin",
			Filename:   "demo.bin",
			Identifier: "demo",
			Version:    "1.0.0",
			Shasum:     sha1Hex(content),
		},
		DownloadOptions{
			OutputDir: dir,
			Resume:    false,
			Chunks:    1,
			RequestOptions: RequestOptions{
				Verbose: newVerboseLogger(true, &log),
			},
		},
		io.Discard,
		nil,
	)
	if err != nil {
		t.Fatalf("downloadPlanWithProgressCallback returned error: %v", err)
	}
	if !strings.Contains(log.String(), "[integrity] verified download demo@1.0.0 using shasum") {
		t.Fatalf("expected integrity verbose log, got %q", log.String())
	}
}

func TestDownloadPlanWithProgressCallbackAcceptsPyPISHA256Integrity(t *testing.T) {
	content := []byte(strings.Repeat("pypi-integrity-", 256))
	server := newRangeTestServer(content, nil)
	defer server.Close()

	dir := t.TempDir()
	_, err := downloadPlanWithProgressCallback(
		context.Background(),
		server.Client(),
		DownloadPlan{
			Source:     "pip",
			URL:        server.URL + "/demo.tar.gz",
			Filename:   "demo.tar.gz",
			Identifier: "demo",
			Version:    "1.2.3",
			Integrity:  formatDigestIntegrity("sha256", sha256Hex(content)),
			MirrorName: "pypi",
		},
		DownloadOptions{OutputDir: dir, Resume: false, Chunks: 1},
		io.Discard,
		nil,
	)
	if err != nil {
		t.Fatalf("expected PyPI sha256 integrity verification to pass, got %v", err)
	}
}

func TestVerboseLoggerStaysSilentWhenDisabled(t *testing.T) {
	var out bytes.Buffer
	logger := newVerboseLogger(false, &out)
	logger.Logf("resolve", "should stay silent")
	if out.Len() != 0 {
		t.Fatalf("expected disabled logger to stay silent, got %q", out.String())
	}
}

func jsonResponse(req *http.Request, status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}
}

func captureStderrMain(t *testing.T, fn func()) string {
	t.Helper()
	original := os.Stderr
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}
	os.Stderr = writer
	defer func() {
		os.Stderr = original
	}()

	done := make(chan string, 1)
	go func() {
		raw, _ := io.ReadAll(reader)
		done <- string(raw)
	}()

	fn()

	_ = writer.Close()
	return <-done
}

func sha256Hex(value []byte) string {
	sum := sha256.Sum256(value)
	return hex.EncodeToString(sum[:])
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

type failingReadCloser struct {
	data []byte
	err  error
	done bool
}

func (r *failingReadCloser) Read(p []byte) (int, error) {
	if r.done {
		if r.err == nil {
			return 0, io.EOF
		}
		return 0, r.err
	}
	r.done = true
	n := copy(p, r.data)
	if r.err == nil {
		return n, io.EOF
	}
	return n, r.err
}

func (r *failingReadCloser) Close() error {
	return nil
}

func newRangeTestServer(content []byte, onRequest func(*http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if onRequest != nil {
			onRequest(r)
		}
		switch r.Method {
		case http.MethodHead:
			w.Header().Set("Content-Length", strconv.Itoa(len(content)))
			w.Header().Set("Accept-Ranges", "bytes")
			w.WriteHeader(http.StatusOK)
			return
		case http.MethodGet:
			rangeHeader := strings.TrimSpace(r.Header.Get("Range"))
			if rangeHeader == "" {
				w.Header().Set("Content-Length", strconv.Itoa(len(content)))
				w.Header().Set("Accept-Ranges", "bytes")
				_, _ = w.Write(content)
				return
			}
			spec := strings.TrimPrefix(rangeHeader, "bytes=")
			parts := strings.SplitN(spec, "-", 2)
			if len(parts) != 2 {
				http.Error(w, "invalid range", http.StatusRequestedRangeNotSatisfiable)
				return
			}
			start, err := strconv.Atoi(parts[0])
			if err != nil {
				http.Error(w, "invalid range", http.StatusRequestedRangeNotSatisfiable)
				return
			}
			end := len(content) - 1
			if strings.TrimSpace(parts[1]) != "" {
				end, err = strconv.Atoi(parts[1])
				if err != nil {
					http.Error(w, "invalid range", http.StatusRequestedRangeNotSatisfiable)
					return
				}
			}
			if start < 0 || start >= len(content) {
				http.Error(w, "range out of bounds", http.StatusRequestedRangeNotSatisfiable)
				return
			}
			if end <= 0 || end >= len(content) {
				end = len(content) - 1
			}
			if end < start {
				http.Error(w, "invalid range", http.StatusRequestedRangeNotSatisfiable)
				return
			}
			w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, len(content)))
			w.Header().Set("Content-Length", strconv.Itoa(end-start+1))
			w.Header().Set("Accept-Ranges", "bytes")
			w.WriteHeader(http.StatusPartialContent)
			_, _ = w.Write(content[start : end+1])
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))
}
