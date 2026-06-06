package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const defaultRuntimeConfigPath = "source-fetcher.yaml"

type BatchConfig struct {
	OutputDir       string                 `yaml:"output_dir"`
	Timeout         string                 `yaml:"timeout"`
	Mirrors         map[string]string      `yaml:"mirrors"`
	InstallDefaults InstallDefaultsConfig  `yaml:"install_defaults"`
	AuthProfiles    map[string]AuthProfile `yaml:"auth_profiles"`
	Downloads       []DownloadRequest      `yaml:"downloads"`
	Installs        []InstallRequest       `yaml:"installs"`
}

type InstallDefaultsConfig struct {
	ScriptsPolicy string   `yaml:"scripts_policy"`
	AllowScripts  []string `yaml:"allow_scripts"`
}

type AuthProfile struct {
	Headers          map[string]string `yaml:"headers"`
	BearerTokenEnv   string            `yaml:"bearer_token_env"`
	BasicUsername    string            `yaml:"basic_username"`
	BasicPasswordEnv string            `yaml:"basic_password_env"`
}

func loadBatchConfig(path string) (BatchConfig, error) {
	absPath, err := filepath.Abs(strings.TrimSpace(path))
	if err != nil {
		return BatchConfig{}, fmt.Errorf("resolve config path: %w", err)
	}

	raw, err := os.ReadFile(absPath)
	if err != nil {
		return BatchConfig{}, fmt.Errorf("read config file %s: %w", absPath, err)
	}

	cfg := BatchConfig{
		OutputDir:    ".",
		Timeout:      "30s",
		Mirrors:      map[string]string{},
		AuthProfiles: map[string]AuthProfile{},
	}
	unknownFieldWarnings, err := collectYAMLUnknownFieldWarnings(raw, &cfg)
	if err != nil {
		return BatchConfig{}, fmt.Errorf("parse config file %s: %w", absPath, err)
	}
	// 使用安全的 YAML 解析
	if err := unmarshalYAMLSafe(raw, &cfg); err != nil {
		return BatchConfig{}, fmt.Errorf("parse config file %s: %w", absPath, err)
	}
	for _, warning := range unknownFieldWarnings {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: config file %s %s\n", absPath, warning)
	}

	if strings.TrimSpace(cfg.OutputDir) == "" {
		cfg.OutputDir = "."
	}
	outputDir := strings.TrimSpace(cfg.OutputDir)
	if !filepath.IsAbs(outputDir) {
		outputDir = filepath.Join(filepath.Dir(absPath), outputDir)
	}
	cfg.OutputDir = filepath.Clean(outputDir)

	if cfg.Mirrors == nil {
		cfg.Mirrors = map[string]string{}
	}
	for key, value := range cfg.Mirrors {
		delete(cfg.Mirrors, key)
		cfg.Mirrors[strings.ToLower(strings.TrimSpace(key))] = strings.TrimSpace(value)
	}
	if cfg.AuthProfiles == nil {
		cfg.AuthProfiles = map[string]AuthProfile{}
	}
	for key, profile := range cfg.AuthProfiles {
		delete(cfg.AuthProfiles, key)
		cfg.AuthProfiles[strings.ToLower(strings.TrimSpace(key))] = normalizeAuthProfile(profile)
	}
	cfg.InstallDefaults = normalizeInstallDefaultsConfig(cfg.InstallDefaults)

	for index := range cfg.Downloads {
		cfg.Downloads[index].Source = strings.ToLower(strings.TrimSpace(cfg.Downloads[index].Source))
		cfg.Downloads[index].Name = strings.TrimSpace(cfg.Downloads[index].Name)
		cfg.Downloads[index].PackageID = strings.TrimSpace(cfg.Downloads[index].PackageID)
		cfg.Downloads[index].Version = strings.TrimSpace(cfg.Downloads[index].Version)
		cfg.Downloads[index].URL = strings.TrimSpace(cfg.Downloads[index].URL)
		cfg.Downloads[index].Mirror = strings.TrimSpace(cfg.Downloads[index].Mirror)
		cfg.Downloads[index].Arch = strings.TrimSpace(cfg.Downloads[index].Arch)
		cfg.Downloads[index].OutputDir = strings.TrimSpace(cfg.Downloads[index].OutputDir)
		cfg.Downloads[index].AuthProfile = strings.TrimSpace(cfg.Downloads[index].AuthProfile)
		if cfg.Downloads[index].Chunks <= 0 {
			cfg.Downloads[index].Chunks = 1
		}
		if cfg.Downloads[index].OutputDir != "" && !filepath.IsAbs(cfg.Downloads[index].OutputDir) {
			cfg.Downloads[index].OutputDir = filepath.Join(filepath.Dir(absPath), cfg.Downloads[index].OutputDir)
		}
	}
	for index := range cfg.Installs {
		cfg.Installs[index].Source = strings.ToLower(strings.TrimSpace(cfg.Installs[index].Source))
		cfg.Installs[index].Name = strings.TrimSpace(cfg.Installs[index].Name)
		cfg.Installs[index].Version = strings.TrimSpace(cfg.Installs[index].Version)
		cfg.Installs[index].Mirror = strings.TrimSpace(cfg.Installs[index].Mirror)
		cfg.Installs[index].OutputDir = strings.TrimSpace(cfg.Installs[index].OutputDir)
		cfg.Installs[index].AuthProfile = strings.TrimSpace(cfg.Installs[index].AuthProfile)
		cfg.Installs[index].ScriptsPolicy = normalizeScriptsPolicy(cfg.Installs[index].ScriptsPolicy)
		cfg.Installs[index].AllowScripts = normalizeAllowedScriptPackages(cfg.Installs[index].AllowScripts)
		if cfg.Installs[index].Chunks <= 0 {
			cfg.Installs[index].Chunks = 1
		}
		if cfg.Installs[index].OutputDir != "" && !filepath.IsAbs(cfg.Installs[index].OutputDir) {
			cfg.Installs[index].OutputDir = filepath.Join(filepath.Dir(absPath), cfg.Installs[index].OutputDir)
		}
	}

	return cfg, nil
}

func (cfg BatchConfig) TimeoutValue() (time.Duration, error) {
	value := strings.TrimSpace(cfg.Timeout)
	if value == "" {
		return 30 * time.Second, nil
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse timeout %q: %w", value, err)
	}
	if duration <= 0 {
		return 30 * time.Second, nil
	}
	return duration, nil
}

func (cfg BatchConfig) ApplyTaskDefaults(req DownloadRequest) DownloadRequest {
	if strings.TrimSpace(req.OutputDir) == "" {
		req.OutputDir = cfg.OutputDir
	}
	if strings.TrimSpace(req.Mirror) == "" && cfg.Mirrors != nil {
		if mirror, ok := cfg.Mirrors[strings.ToLower(strings.TrimSpace(req.Source))]; ok {
			req.Mirror = strings.TrimSpace(mirror)
		}
	}
	if req.Chunks <= 0 {
		req.Chunks = 1
	}
	return req
}

func (cfg BatchConfig) ApplyInstallDefaults(req InstallRequest) InstallRequest {
	if strings.TrimSpace(req.OutputDir) == "" {
		req.OutputDir = cfg.OutputDir
	}
	if strings.TrimSpace(req.Mirror) == "" && cfg.Mirrors != nil {
		if mirror, ok := cfg.Mirrors[strings.ToLower(strings.TrimSpace(req.Source))]; ok {
			req.Mirror = strings.TrimSpace(mirror)
		}
	}
	if req.Chunks <= 0 {
		req.Chunks = 1
	}
	if strings.TrimSpace(req.ScriptsPolicy) == "" {
		req.ScriptsPolicy = cfg.InstallDefaults.ScriptsPolicy
	}
	if len(req.AllowScripts) == 0 {
		req.AllowScripts = append([]string(nil), cfg.InstallDefaults.AllowScripts...)
	}
	return req
}

func loadOptionalRuntimeConfig(path string) (BatchConfig, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return emptyRuntimeConfig(), nil
	}
	return loadOptionalRuntimeConfigWithPolicy(trimmed, true)
}

func loadOptionalRuntimeConfigWithPolicy(path string, allowMissing bool) (BatchConfig, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return emptyRuntimeConfig(), nil
	}
	absPath, err := filepath.Abs(trimmed)
	if err != nil {
		return BatchConfig{}, fmt.Errorf("resolve config path: %w", err)
	}
	if _, err := os.Stat(absPath); err != nil {
		if os.IsNotExist(err) {
			if allowMissing {
				return emptyRuntimeConfig(), nil
			}
			return BatchConfig{}, fmt.Errorf("config file %s does not exist", absPath)
		}
		return BatchConfig{}, fmt.Errorf("stat config file %s: %w", absPath, err)
	}
	return loadBatchConfig(absPath)
}

func emptyRuntimeConfig() BatchConfig {
	return BatchConfig{
		OutputDir:       ".",
		Timeout:         "30s",
		Mirrors:         map[string]string{},
		InstallDefaults: InstallDefaultsConfig{},
		AuthProfiles:    map[string]AuthProfile{},
	}
}

func (cfg BatchConfig) ResolveRequestOptions(profileName string) (RequestOptions, error) {
	profileName = strings.TrimSpace(profileName)
	if profileName == "" {
		return RequestOptions{}, nil
	}
	if cfg.AuthProfiles == nil {
		return RequestOptions{}, fmt.Errorf("auth profile %q not found", profileName)
	}
	profile, ok := cfg.AuthProfiles[strings.ToLower(profileName)]
	if !ok {
		return RequestOptions{}, fmt.Errorf("auth profile %q not found", profileName)
	}
	headers := make(map[string]string, len(profile.Headers)+1)
	for key, value := range profile.Headers {
		headers[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	if envName := strings.TrimSpace(profile.BearerTokenEnv); envName != "" {
		token := strings.TrimSpace(os.Getenv(envName))
		if token == "" {
			return RequestOptions{}, fmt.Errorf("auth profile %q requires env %s", profileName, envName)
		}
		headers["Authorization"] = "Bearer " + token
	}
	if username := strings.TrimSpace(profile.BasicUsername); username != "" || strings.TrimSpace(profile.BasicPasswordEnv) != "" {
		passwordEnv := strings.TrimSpace(profile.BasicPasswordEnv)
		password := strings.TrimSpace(os.Getenv(passwordEnv))
		if passwordEnv != "" && password == "" {
			return RequestOptions{}, fmt.Errorf("auth profile %q requires env %s", profileName, passwordEnv)
		}
		token := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
		headers["Authorization"] = "Basic " + token
	}
	return RequestOptions{Headers: headers}, nil
}

func (cfg BatchConfig) BindDownloadRequest(req DownloadRequest) (DownloadRequest, error) {
	req = cfg.ApplyTaskDefaults(req)
	options, err := cfg.ResolveRequestOptions(req.AuthProfile)
	if err != nil {
		return DownloadRequest{}, err
	}
	req.RequestOptions = options
	return req, nil
}

func (cfg BatchConfig) BindInstallRequest(req InstallRequest) (InstallRequest, error) {
	req = cfg.ApplyInstallDefaults(req)
	options, err := cfg.ResolveRequestOptions(req.AuthProfile)
	if err != nil {
		return InstallRequest{}, err
	}
	req.RequestOptions = options
	return req, nil
}

func normalizeAuthProfile(profile AuthProfile) AuthProfile {
	if profile.Headers == nil {
		profile.Headers = map[string]string{}
	}
	normalized := make(map[string]string, len(profile.Headers))
	for key, value := range profile.Headers {
		normalized[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	profile.Headers = normalized
	profile.BearerTokenEnv = strings.TrimSpace(profile.BearerTokenEnv)
	profile.BasicUsername = strings.TrimSpace(profile.BasicUsername)
	profile.BasicPasswordEnv = strings.TrimSpace(profile.BasicPasswordEnv)
	return profile
}

func normalizeInstallDefaultsConfig(value InstallDefaultsConfig) InstallDefaultsConfig {
	value.ScriptsPolicy = normalizeScriptsPolicy(value.ScriptsPolicy)
	value.AllowScripts = normalizeAllowedScriptPackages(value.AllowScripts)
	return value
}

func collectYAMLUnknownFieldWarnings(raw []byte, target any) ([]string, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	decoder.KnownFields(true)
	if err := decoder.Decode(target); err != nil {
		unknown, remaining := splitYAMLUnknownFieldErrors(err)
		if len(remaining) > 0 {
			return nil, fmt.Errorf("%s", strings.Join(remaining, "\n"))
		}
		return unknown, nil
	}
	return nil, nil
}

func splitYAMLUnknownFieldErrors(err error) ([]string, []string) {
	if err == nil {
		return nil, nil
	}
	lines := strings.Split(strings.TrimSpace(err.Error()), "\n")
	unknown := make([]string, 0, len(lines))
	remaining := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line == "yaml: unmarshal errors:" {
			continue
		}
		if strings.Contains(line, ": field ") && strings.Contains(line, " not found in type ") {
			unknown = append(unknown, "contains unknown field "+line)
			continue
		}
		remaining = append(remaining, line)
	}
	return unknown, remaining
}

// unmarshalYAMLSafe 安全地解析 YAML（确保禁用不安全特性）
// gopkg.in/yaml.v3 默认已经禁用了 !!python/object 等不安全标签
// 但这里明确使用 decoder 以保证一致性和可控性
func unmarshalYAMLSafe(data []byte, v interface{}) error {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	// KnownFields(false) 允许未知字段，但我们在 collectYAMLUnknownFieldWarnings 中单独处理
	decoder.KnownFields(false)
	return decoder.Decode(v)
}
