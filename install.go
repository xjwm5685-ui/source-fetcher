package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	neturl "net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

type InstallRequest struct {
	Source         string   `yaml:"source"`
	Name           string   `yaml:"name"`
	Version        string   `yaml:"version"`
	OutputDir      string   `yaml:"output_dir"`
	Mirror         string   `yaml:"mirror"`
	AuthProfile    string   `yaml:"auth"`
	Resume         bool     `yaml:"resume"`
	Chunks         int      `yaml:"chunks"`
	OmitOptional   bool     `yaml:"omit_optional"`
	IncludePeer    bool     `yaml:"include_peer"`
	IncludeDev     bool     `yaml:"include_dev"`
	LockfilePath   string   `yaml:"lockfile"`
	FrozenLockfile bool     `yaml:"frozen_lockfile"`
	ScriptsPolicy  string   `yaml:"scripts_policy"`
	AllowScripts   []string `yaml:"allow_scripts"`

	RequestOptions RequestOptions `yaml:"-"`
}

type InstallPackage struct {
	Name                 string            `json:"name"`
	Requested            string            `json:"requested"`
	Version              string            `json:"version"`
	MirrorName           string            `json:"mirror,omitempty"`
	URL                  string            `json:"url"`
	Filename             string            `json:"filename"`
	Integrity            string            `json:"integrity,omitempty"`
	Shasum               string            `json:"shasum,omitempty"`
	Dependencies         map[string]string `json:"dependencies,omitempty"`
	OptionalDependencies map[string]string `json:"optional_dependencies,omitempty"`
	PeerDependencies     map[string]string `json:"peer_dependencies,omitempty"`
	DevDependencies      map[string]string `json:"dev_dependencies,omitempty"`
	ResolvedDependencies map[string]string `json:"resolved_dependencies,omitempty"`
}

type InstallPlan struct {
	Source        string           `json:"source"`
	Root          string           `json:"root"`
	Requested     string           `json:"requested,omitempty"`
	RootVersion   string           `json:"root_version"`
	MirrorName    string           `json:"mirror"`
	LockfilePath  string           `json:"lockfile_path,omitempty"`
	OmitOptional  bool             `json:"omit_optional,omitempty"`
	IncludePeer   bool             `json:"include_peer,omitempty"`
	IncludeDev    bool             `json:"include_dev,omitempty"`
	ScriptsPolicy string           `json:"scripts_policy,omitempty"`
	AllowScripts  []string         `json:"allow_scripts,omitempty"`
	Packages      []InstallPackage `json:"packages"`
}

type InstalledPackage struct {
	Name                 string            `json:"name"`
	Version              string            `json:"version"`
	MirrorName           string            `json:"mirror,omitempty"`
	ArchivePath          string            `json:"archive_path"`
	StorePath            string            `json:"store_path"`
	InstallPaths         []string          `json:"install_paths,omitempty"`
	BinDir               string            `json:"bin_dir,omitempty"`
	BinLinks             []string          `json:"bin_links,omitempty"`
	Size                 int64             `json:"size"`
	SHA256               string            `json:"sha256"`
	Integrity            string            `json:"integrity,omitempty"`
	Shasum               string            `json:"shasum,omitempty"`
	Required             map[string]string `json:"dependencies,omitempty"`
	ResolvedDependencies map[string]string `json:"resolved_dependencies,omitempty"`
}

type InstallManifest struct {
	Source         string             `json:"source"`
	Root           string             `json:"root"`
	RootVersion    string             `json:"root_version"`
	MirrorName     string             `json:"mirror"`
	GeneratedAt    string             `json:"generated_at"`
	CacheDir       string             `json:"cache_dir"`
	StoreDir       string             `json:"store_dir"`
	NodeModulesDir string             `json:"node_modules_dir"`
	BinDir         string             `json:"bin_dir"`
	RootPath       string             `json:"root_path"`
	LockfilePath   string             `json:"lockfile_path"`
	Packages       []InstalledPackage `json:"packages"`
}

type InstallResult struct {
	ManifestPath   string
	LockfilePath   string
	RootPath       string
	NodeModulesDir string
	Packages       []InstalledPackage
	Duration       time.Duration
}

type InstallLockfile struct {
	Version      int              `json:"version"`
	Source       string           `json:"source"`
	Root         string           `json:"root"`
	Requested    string           `json:"requested,omitempty"`
	RootVersion  string           `json:"root_version"`
	MirrorName   string           `json:"mirror"`
	GeneratedAt  string           `json:"generated_at"`
	OmitOptional bool             `json:"omit_optional,omitempty"`
	IncludePeer  bool             `json:"include_peer,omitempty"`
	IncludeDev   bool             `json:"include_dev,omitempty"`
	AllowScripts []string         `json:"allow_scripts,omitempty"`
	Packages     []InstallPackage `json:"packages"`
}

type UninstallRequest struct {
	OutputDir    string
	ManifestPath string
	KeepCache    bool
	KeepManifest bool
}

type UninstallResult struct {
	ManifestPath string
	RemovedPaths []string
	MissingPaths []string
	SkippedPaths []string
	Duration     time.Duration
}

type RepairRequest struct {
	OutputDir    string
	ManifestPath string
}

type RepairResult struct {
	ManifestPath      string
	RepairedPaths     []string
	HealthyPaths      []string
	MissingCachePaths []string
	SkippedPaths      []string
	Duration          time.Duration
}

type uninstallPackagePath struct {
	Path   string
	Record InstalledPackage
}

type npmInstallPlanner struct {
	ctx           context.Context
	client        *http.Client
	mirrors       []Mirror
	rootMirror    Mirror
	options       RequestOptions
	omitOptional  bool
	includePeer   bool
	metadataCache map[string]npmMetadataCacheEntry
	packages      map[string]InstallPackage
	order         []string
}

type npmMetadataCacheEntry struct {
	Mirror   Mirror
	Metadata npmPackageMetadata
}

type npmDependencySpec struct {
	InstallName string
	SourceName  string
	Requested   string
	Optional    bool
}

type npmPackageArchiveManifest struct {
	Name    string            `json:"name"`
	Version string            `json:"version"`
	Bin     npmBinMap         `json:"bin"`
	Scripts map[string]string `json:"scripts"`
}

type npmBinMap map[string]string

func (m *npmBinMap) UnmarshalJSON(data []byte) error {
	if m == nil {
		return errors.New("npm bin map target is nil")
	}
	trimmed := bytes.TrimSpace(data)
	switch {
	case len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")):
		*m = nil
		return nil
	case len(trimmed) >= 2 && trimmed[0] == '"' && trimmed[len(trimmed)-1] == '"':
		var value string
		if err := json.Unmarshal(trimmed, &value); err != nil {
			return err
		}
		*m = npmBinMap{"": strings.TrimSpace(value)}
		return nil
	case len(trimmed) >= 2 && trimmed[0] == '{' && trimmed[len(trimmed)-1] == '}':
		values := map[string]string{}
		if err := json.Unmarshal(trimmed, &values); err != nil {
			return err
		}
		normalized := make(npmBinMap, len(values))
		for key, value := range values {
			normalized[strings.TrimSpace(key)] = strings.TrimSpace(value)
		}
		*m = normalized
		return nil
	default:
		return fmt.Errorf("unsupported npm bin shape: %s", string(trimmed))
	}
}

type npmInstallExecutor struct {
	outputDir      string
	nodeModulesDir string
	scriptsPolicy  string
	allowScripts   map[string]struct{}
	verbose        *VerboseLogger
	packageByKey   map[string]InstallPackage
	materialized   map[string]*InstalledPackage
	placedByPath   map[string]string
}

type archiveDigests struct {
	Size   int64
	SHA1   string
	SHA256 string
	SHA512 string
}

func resolveInstallPlan(ctx context.Context, client *http.Client, req InstallRequest) (InstallPlan, error) {
	switch strings.ToLower(strings.TrimSpace(req.Source)) {
	case "", "npm":
		req.Source = "npm"
		return resolveNPMInstallPlan(ctx, client, req)
	case "choco", "winget":
		return resolveNativeInstallPlan(ctx, client, req)
	default:
		return InstallPlan{}, fmt.Errorf("install currently supports: npm, choco, winget")
	}
}

func resolveNPMInstallPlan(ctx context.Context, client *http.Client, req InstallRequest) (InstallPlan, error) {
	if strings.TrimSpace(req.Name) == "" {
		return InstallPlan{}, errors.New("--name is required when --source npm")
	}
	req.ScriptsPolicy = normalizeScriptsPolicy(req.ScriptsPolicy)
	req.AllowScripts = normalizeAllowedScriptPackages(req.AllowScripts)

	mirrors, err := resolveMirrors("npm", req.Mirror)
	if err != nil {
		return InstallPlan{}, err
	}
	lockfilePath, err := resolveInstallLockfilePath(req.OutputDir, req.LockfilePath)
	if err != nil {
		return InstallPlan{}, err
	}
	lockfileMirrorName := ""
	if len(mirrors) > 0 {
		lockfileMirrorName = mirrors[0].Name
	}
	if lockfile, err := loadInstallLockfile(lockfilePath); err == nil {
		if reason := installLockfileMismatchReason(lockfile, req, lockfileMirrorName); reason == "" {
			req.RequestOptions.Verbose.Logf("lockfile", "matched %s for %s", lockfilePath, installRequestLabel(req))
			return installPlanFromLockfile(lockfile, lockfilePath, req.ScriptsPolicy), nil
		} else {
			req.RequestOptions.Verbose.Logf("lockfile", "ignored %s for %s: %s", lockfilePath, installRequestLabel(req), reason)
		}
		if req.FrozenLockfile {
			return InstallPlan{}, fmt.Errorf("frozen lockfile %s does not match install request", lockfilePath)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return InstallPlan{}, err
	} else if req.FrozenLockfile {
		return InstallPlan{}, fmt.Errorf("frozen lockfile %s does not exist", lockfilePath)
	}

	planner := &npmInstallPlanner{
		ctx:           ctx,
		client:        client,
		mirrors:       mirrors,
		options:       req.RequestOptions,
		omitOptional:  req.OmitOptional,
		includePeer:   req.IncludePeer,
		metadataCache: map[string]npmMetadataCacheEntry{},
		packages:      map[string]InstallPackage{},
	}

	rootName := strings.TrimSpace(req.Name)
	rootPackage, err := planner.planPackage(rootName, rootName, strings.TrimSpace(req.Version), req.IncludeDev)
	if err != nil {
		return InstallPlan{}, err
	}

	packages := make([]InstallPackage, 0, len(planner.order))
	for _, key := range planner.order {
		packages = append(packages, planner.packages[key])
	}

	return InstallPlan{
		Source:        "npm",
		Root:          rootPackage.Name,
		Requested:     strings.TrimSpace(req.Version),
		RootVersion:   rootPackage.Version,
		MirrorName:    blankIfEmpty(planner.rootMirror.Name, mirrors[0].Name),
		LockfilePath:  lockfilePath,
		OmitOptional:  req.OmitOptional,
		IncludePeer:   req.IncludePeer,
		IncludeDev:    req.IncludeDev,
		ScriptsPolicy: req.ScriptsPolicy,
		AllowScripts:  append([]string(nil), req.AllowScripts...),
		Packages:      packages,
	}, nil
}

func (p *npmInstallPlanner) planPackage(installName string, sourceName string, requested string, includeDev bool) (InstallPackage, error) {
	metadataMirror, metadata, err := p.fetchMetadata(sourceName)
	if err != nil {
		return InstallPackage{}, err
	}

	resolvedVersion, details, err := pickNPMVersion(metadata, requested)
	if err != nil {
		return InstallPackage{}, fmt.Errorf("resolve npm dependency %s@%s: %w", sourceName, blankIfEmpty(requested, "latest"), err)
	}
	key := strings.ToLower(strings.TrimSpace(installName)) + "@" + resolvedVersion
	if pkg, ok := p.packages[key]; ok {
		return pkg, nil
	}
	if !npmPackageMatchesCurrentPlatform(details) {
		return InstallPackage{}, fmt.Errorf("npm dependency %s@%s is not supported on %s/%s", installName, resolvedVersion, currentNPMOS(), currentNPMCPU())
	}
	if strings.TrimSpace(details.Dist.Tarball) == "" {
		return InstallPackage{}, fmt.Errorf("npm dependency %s@%s has no tarball url", sourceName, resolvedVersion)
	}

	pkg := InstallPackage{
		Name:                 strings.TrimSpace(installName),
		Requested:            strings.TrimSpace(requested),
		Version:              resolvedVersion,
		MirrorName:           metadataMirror.Name,
		URL:                  rewriteMirrorURL(details.Dist.Tarball, metadataMirror.BaseURL),
		Filename:             sanitizeFileName(strings.ReplaceAll(strings.TrimSpace(installName), "/", "_") + "-" + resolvedVersion + ".tgz"),
		Integrity:            strings.TrimSpace(details.Dist.Integrity),
		Shasum:               strings.TrimSpace(details.Dist.Shasum),
		Dependencies:         cloneStringMap(details.Dependencies),
		OptionalDependencies: cloneStringMap(details.OptionalDependencies),
		PeerDependencies:     cloneStringMap(details.PeerDependencies),
		DevDependencies:      cloneStringMap(details.DevDependencies),
	}
	p.options.Verbose.Logf("resolve", "npm package %s@%s resolved via %s", pkg.Name, pkg.Version, pkg.MirrorName)
	p.packages[key] = pkg
	p.order = append(p.order, key)

	selectedDependencies, err := p.selectedDependencySpecs(pkg, includeDev)
	if err != nil {
		return InstallPackage{}, err
	}
	resolvedDependencies := make(map[string]string, len(selectedDependencies))
	for _, spec := range selectedDependencies {
		dependencyPackage, err := p.planPackage(spec.InstallName, spec.SourceName, spec.Requested, false)
		if err != nil {
			if spec.Optional {
				continue
			}
			return InstallPackage{}, err
		}
		resolvedDependencies[spec.InstallName] = dependencyPackage.Version
	}
	pkg.ResolvedDependencies = resolvedDependencies
	p.packages[key] = pkg
	return pkg, nil
}

func (p *npmInstallPlanner) fetchMetadata(name string) (Mirror, npmPackageMetadata, error) {
	cacheKey := strings.ToLower(strings.TrimSpace(name))
	if entry, ok := p.metadataCache[cacheKey]; ok {
		p.options.Verbose.Logf("resolve", "npm metadata %s reused from cache via %s", name, entry.Mirror.Name)
		return entry.Mirror, entry.Metadata, nil
	}

	candidates := p.mirrors
	if len(candidates) == 0 {
		var err error
		candidates, err = resolveMirrors("npm", "")
		if err != nil {
			return Mirror{}, npmPackageMetadata{}, err
		}
	}
	var (
		mirror   Mirror
		metadata npmPackageMetadata
		lastErr  error
	)
	for _, candidate := range candidates {
		metadataURL := joinURL(candidate.BaseURL, "/"+neturl.PathEscape(strings.TrimSpace(name)))
		if err := getJSON(p.ctx, p.client, metadataURL, p.options, &metadata); err != nil {
			p.options.Verbose.Logf("resolve", "npm metadata %s via %s failed: %v", name, candidate.Name, err)
			lastErr = fmt.Errorf("fetch npm metadata for %s via %s: %w", name, candidate.Name, err)
			continue
		}
		mirror = candidate
		break
	}
	if mirror.Name == "" {
		return Mirror{}, npmPackageMetadata{}, lastErr
	}
	if p.rootMirror.Name == "" {
		p.rootMirror = mirror
	}
	p.metadataCache[cacheKey] = npmMetadataCacheEntry{
		Mirror:   mirror,
		Metadata: metadata,
	}
	return mirror, metadata, nil
}

func (p *npmInstallPlanner) selectedDependencySpecs(pkg InstallPackage, includeDev bool) ([]npmDependencySpec, error) {
	selected := map[string]npmDependencySpec{}
	mergeInstallDependencySpecs(selected, pkg.Dependencies, false, false)
	if !p.omitOptional {
		mergeInstallDependencySpecs(selected, pkg.OptionalDependencies, true, true)
	}
	if p.includePeer {
		mergeInstallDependencySpecs(selected, pkg.PeerDependencies, false, false)
	}
	if includeDev {
		mergeInstallDependencySpecs(selected, pkg.DevDependencies, false, false)
	}
	if len(selected) == 0 {
		return nil, nil
	}
	specs := make([]npmDependencySpec, 0, len(selected))
	for _, name := range sortedMapKeysFromSpecs(selected) {
		spec, err := parseNPMDependencySpec(selected[name].InstallName, selected[name].Requested)
		if err != nil {
			if selected[name].Optional {
				continue
			}
			return nil, err
		}
		spec.Optional = selected[name].Optional
		specs = append(specs, spec)
	}
	return specs, nil
}

func mergeInstallDependencySpecs(target map[string]npmDependencySpec, source map[string]string, optional bool, overwrite bool) {
	for key, value := range source {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if _, exists := target[key]; exists && !overwrite {
			continue
		}
		target[key] = npmDependencySpec{
			InstallName: key,
			SourceName:  key,
			Requested:   strings.TrimSpace(value),
			Optional:    optional,
		}
	}
}

func sortedMapKeysFromSpecs(values map[string]npmDependencySpec) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func parseNPMDependencySpec(installName string, requested string) (npmDependencySpec, error) {
	spec := npmDependencySpec{
		InstallName: strings.TrimSpace(installName),
		SourceName:  strings.TrimSpace(installName),
		Requested:   strings.TrimSpace(requested),
	}
	if spec.InstallName == "" {
		return npmDependencySpec{}, errors.New("npm dependency name is empty")
	}
	if !strings.HasPrefix(spec.Requested, "npm:") {
		return spec, nil
	}
	sourceName, sourceRequested, err := splitNPMAliasSpecifier(strings.TrimSpace(strings.TrimPrefix(spec.Requested, "npm:")))
	if err != nil {
		return npmDependencySpec{}, fmt.Errorf("parse npm alias dependency %s: %w", spec.InstallName, err)
	}
	spec.SourceName = sourceName
	spec.Requested = sourceRequested
	return spec, nil
}

func splitNPMAliasSpecifier(value string) (string, string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", "", errors.New("empty alias target")
	}
	lastAt := strings.LastIndex(value, "@")
	if lastAt > 0 {
		return strings.TrimSpace(value[:lastAt]), strings.TrimSpace(value[lastAt+1:]), nil
	}
	return value, "", nil
}

func npmPackageMatchesCurrentPlatform(details npmVersionDetails) bool {
	return npmPlatformListAllows(details.OS, currentNPMOS()) && npmPlatformListAllows(details.CPU, currentNPMCPU())
}

func npmPlatformListAllows(values []string, current string) bool {
	if len(values) == 0 {
		return true
	}
	hasPositive := false
	positiveMatch := false
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if strings.HasPrefix(value, "!") {
			if strings.TrimSpace(strings.TrimPrefix(value, "!")) == current {
				return false
			}
			continue
		}
		hasPositive = true
		if value == current {
			positiveMatch = true
		}
	}
	if hasPositive {
		return positiveMatch
	}
	return true
}

func currentNPMOS() string {
	switch runtime.GOOS {
	case "windows":
		return "win32"
	default:
		return runtime.GOOS
	}
}

func currentNPMCPU() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x64"
	case "386":
		return "ia32"
	default:
		return runtime.GOARCH
	}
}

func executeInstallPlan(ctx context.Context, client *http.Client, plan InstallPlan, options DownloadOptions) (InstallResult, error) {
	start := time.Now()
	if strings.TrimSpace(options.OutputDir) == "" {
		options.OutputDir = "."
	}
	absOutput, err := filepath.Abs(options.OutputDir)
	if err != nil {
		return InstallResult{}, fmt.Errorf("resolve install output dir: %w", err)
	}
	options.OutputDir = absOutput

	if err := os.MkdirAll(options.OutputDir, 0o755); err != nil {
		return InstallResult{}, fmt.Errorf("create install output dir: %w", err)
	}
	workDir := filepath.Join(options.OutputDir, ".source-fetcher")
	cacheDir := filepath.Join(workDir, "tarballs")
	storeDir := filepath.Join(workDir, "store")
	nodeModulesDir := filepath.Join(options.OutputDir, "node_modules")
	for _, dir := range []string{cacheDir, storeDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return InstallResult{}, fmt.Errorf("create install work dir %s: %w", dir, err)
		}
	}
	stagingRoot, err := os.MkdirTemp(filepath.Dir(options.OutputDir), filepath.Base(options.OutputDir)+".source-fetcher-install-")
	if err != nil {
		return InstallResult{}, fmt.Errorf("create install staging dir: %w", err)
	}
	defer os.RemoveAll(stagingRoot)
	stagingOutputDir := filepath.Join(stagingRoot, "output")
	stagingNodeModulesDir := filepath.Join(stagingOutputDir, "node_modules")
	if err := os.MkdirAll(stagingNodeModulesDir, 0o755); err != nil {
		return InstallResult{}, fmt.Errorf("create staging node_modules dir: %w", err)
	}

	packageByKey := make(map[string]InstallPackage, len(plan.Packages))
	materialized := make(map[string]*InstalledPackage, len(plan.Packages))
	for _, pkg := range plan.Packages {
		key := installPackageKey(pkg.Name, pkg.Version)
		packageByKey[key] = pkg
		downloaded, err := ensureInstallPackageArchive(ctx, client, plan, pkg, DownloadOptions{
			OutputDir:      cacheDir,
			Resume:         options.Resume,
			Chunks:         options.Chunks,
			RequestOptions: options.RequestOptions,
		})
		if err != nil {
			return InstallResult{}, fmt.Errorf("install %s@%s: %w", pkg.Name, pkg.Version, err)
		}
		storePath := filepath.Join(storeDir, installStoreDirectoryName(pkg.Name, pkg.Version))
		if err := materializeNPMPackage(downloaded.Path, storePath, pkg.Name, pkg.Version); err != nil {
			return InstallResult{}, fmt.Errorf("extract %s@%s: %w", pkg.Name, pkg.Version, err)
		}
		materialized[key] = &InstalledPackage{
			Name:                 pkg.Name,
			Version:              pkg.Version,
			MirrorName:           pkg.MirrorName,
			ArchivePath:          downloaded.Path,
			StorePath:            storePath,
			Size:                 downloaded.Size,
			SHA256:               downloaded.SHA256,
			Integrity:            pkg.Integrity,
			Shasum:               pkg.Shasum,
			Required:             cloneStringMap(pkg.Dependencies),
			ResolvedDependencies: cloneStringMap(pkg.ResolvedDependencies),
		}
	}

	rootKey := installPackageKey(plan.Root, plan.RootVersion)
	rootPackage, ok := packageByKey[rootKey]
	if !ok {
		return InstallResult{}, fmt.Errorf("root package %s@%s is missing from the install plan", plan.Root, plan.RootVersion)
	}
	rootPath := filepath.Join(stagingNodeModulesDir, packageRelativeInstallPath(rootPackage.Name))
	executor := npmInstallExecutor{
		outputDir:      stagingOutputDir,
		nodeModulesDir: stagingNodeModulesDir,
		scriptsPolicy:  normalizeScriptsPolicy(plan.ScriptsPolicy),
		allowScripts:   allowedScriptPackageSet(plan.AllowScripts),
		verbose:        options.RequestOptions.Verbose,
		packageByKey:   packageByKey,
		materialized:   materialized,
		placedByPath:   map[string]string{},
	}
	if err := executor.installPackageTree(rootPackage, rootPath); err != nil {
		return InstallResult{}, err
	}
	if err := executor.runLifecycleScripts(ctx, plan.Packages, rootKey); err != nil {
		return InstallResult{}, err
	}

	results := make([]InstalledPackage, 0, len(plan.Packages))
	for _, pkg := range plan.Packages {
		record := materialized[installPackageKey(pkg.Name, pkg.Version)]
		if record == nil {
			continue
		}
		sort.Strings(record.InstallPaths)
		results = append(results, *record)
	}
	results = rewriteInstalledPackagesBasePath(results, stagingOutputDir, options.OutputDir)
	rootPath = filepath.Join(nodeModulesDir, packageRelativeInstallPath(rootPackage.Name))

	manifestPath := filepath.Join(options.OutputDir, "source-fetcher-install.json")
	manifest := InstallManifest{
		Source:         plan.Source,
		Root:           plan.Root,
		RootVersion:    plan.RootVersion,
		MirrorName:     plan.MirrorName,
		GeneratedAt:    time.Now().UTC().Format(time.RFC3339),
		CacheDir:       cacheDir,
		StoreDir:       storeDir,
		NodeModulesDir: nodeModulesDir,
		RootPath:       rootPath,
		LockfilePath:   plan.LockfilePath,
		Packages:       results,
	}
	manifestTempPath, err := writeInstallManifestTemp(manifestPath, manifest)
	if err != nil {
		return InstallResult{}, err
	}
	commitItems := []installCommitItem{
		{StagedPath: stagingNodeModulesDir, TargetPath: nodeModulesDir},
		{StagedPath: manifestTempPath, TargetPath: manifestPath},
	}
	defer os.Remove(manifestTempPath)
	if plan.LockfilePath != "" {
		lockfileTempPath, err := writeInstallLockfileTemp(plan.LockfilePath, plan)
		if err != nil {
			return InstallResult{}, err
		}
		defer os.Remove(lockfileTempPath)
		commitItems = append(commitItems, installCommitItem{StagedPath: lockfileTempPath, TargetPath: plan.LockfilePath})
	}
	if err := commitInstallArtifacts(commitItems); err != nil {
		return InstallResult{}, err
	}

	return InstallResult{
		ManifestPath:   manifestPath,
		LockfilePath:   plan.LockfilePath,
		RootPath:       rootPath,
		NodeModulesDir: nodeModulesDir,
		Packages:       results,
		Duration:       time.Since(start),
	}, nil
}

type installCommitItem struct {
	StagedPath string
	TargetPath string
}

type installBackupItem struct {
	TargetPath string
	BackupPath string
}

func rewriteInstalledPackagesBasePath(packages []InstalledPackage, sourceBase string, targetBase string) []InstalledPackage {
	rewritten := make([]InstalledPackage, len(packages))
	for index, pkg := range packages {
		rewritten[index] = pkg
		rewritten[index].InstallPaths = rewriteInstallPathsBase(pkg.InstallPaths, sourceBase, targetBase)
		rewritten[index].BinLinks = rewriteInstallPathsBase(pkg.BinLinks, sourceBase, targetBase)
	}
	return rewritten
}

func rewriteInstallPathsBase(paths []string, sourceBase string, targetBase string) []string {
	rewritten := make([]string, len(paths))
	for index, path := range paths {
		rewritten[index] = rewriteInstallPathBase(path, sourceBase, targetBase)
	}
	return rewritten
}

func rewriteInstallPathBase(path string, sourceBase string, targetBase string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	relativePath, err := filepath.Rel(sourceBase, path)
	if err != nil || relativePath == "." || strings.HasPrefix(relativePath, ".."+string(os.PathSeparator)) || relativePath == ".." {
		return path
	}
	return filepath.Join(targetBase, relativePath)
}

func writeInstallManifestTemp(path string, manifest InstallManifest) (string, error) {
	raw, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal install manifest: %w", err)
	}
	return writeInstallTempFile(path, raw, "install manifest")
}

func writeInstallLockfileTemp(path string, plan InstallPlan) (string, error) {
	raw, err := marshalInstallLockfile(plan)
	if err != nil {
		return "", err
	}
	return writeInstallTempFile(path, raw, "install lockfile")
}

func writeInstallTempFile(path string, raw []byte, description string) (string, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("create parent dir for %s: %w", description, err)
	}
	file, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".source-fetcher-*")
	if err != nil {
		return "", fmt.Errorf("create temp %s: %w", description, err)
	}
	tempPath := file.Name()
	if _, err := file.Write(raw); err != nil {
		file.Close()
		_ = os.Remove(tempPath)
		return "", fmt.Errorf("write temp %s: %w", description, err)
	}
	if err := file.Sync(); err != nil {
		file.Close()
		_ = os.Remove(tempPath)
		return "", fmt.Errorf("sync temp %s: %w", description, err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(tempPath)
		return "", fmt.Errorf("close temp %s: %w", description, err)
	}
	return tempPath, nil
}

func commitInstallArtifacts(items []installCommitItem) (err error) {
	backups := make([]installBackupItem, 0, len(items))
	committed := make([]installCommitItem, 0, len(items))
	defer func() {
		if err == nil {
			for _, backup := range backups {
				_, _ = removePathIfPresent(backup.BackupPath)
			}
			return
		}
		for index := len(committed) - 1; index >= 0; index-- {
			_, _ = removePathIfPresent(committed[index].TargetPath)
		}
		for index := len(backups) - 1; index >= 0; index-- {
			_ = os.Rename(backups[index].BackupPath, backups[index].TargetPath)
		}
	}()

	for _, item := range items {
		if strings.TrimSpace(item.StagedPath) == "" || strings.TrimSpace(item.TargetPath) == "" {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(item.TargetPath), 0o755); err != nil {
			return fmt.Errorf("create install commit parent %s: %w", filepath.Dir(item.TargetPath), err)
		}
		if _, statErr := os.Lstat(item.TargetPath); statErr == nil {
			backupPath, backupErr := allocateInstallBackupPath(item.TargetPath)
			if backupErr != nil {
				return backupErr
			}
			if err := os.Rename(item.TargetPath, backupPath); err != nil {
				return fmt.Errorf("backup existing install path %s: %w", item.TargetPath, err)
			}
			backups = append(backups, installBackupItem{TargetPath: item.TargetPath, BackupPath: backupPath})
		} else if !os.IsNotExist(statErr) {
			return fmt.Errorf("stat install commit target %s: %w", item.TargetPath, statErr)
		}
		if err := os.Rename(item.StagedPath, item.TargetPath); err != nil {
			return fmt.Errorf("commit install path %s: %w", item.TargetPath, err)
		}
		committed = append(committed, item)
	}
	return nil
}

func allocateInstallBackupPath(targetPath string) (string, error) {
	directory := filepath.Dir(targetPath)
	base := filepath.Base(targetPath)
	for attempt := 0; attempt < 32; attempt++ {
		candidate := filepath.Join(directory, fmt.Sprintf(".%s.source-fetcher-backup-%d-%d", base, time.Now().UnixNano(), attempt))
		if _, err := os.Lstat(candidate); os.IsNotExist(err) {
			return candidate, nil
		} else if err != nil {
			return "", fmt.Errorf("stat install backup path %s: %w", candidate, err)
		}
	}
	return "", fmt.Errorf("allocate install backup path for %s", targetPath)
}

func printInstallPlan(out io.Writer, plan InstallPlan) {
	_, _ = fmt.Fprintf(out, "Source: %s\n", plan.Source)
	_, _ = fmt.Fprintf(out, "Root: %s\n", plan.Root)
	_, _ = fmt.Fprintf(out, "Requested: %s\n", blankIfEmpty(plan.Requested, "latest"))
	_, _ = fmt.Fprintf(out, "Version: %s\n", blankIfEmpty(plan.RootVersion, "-"))
	if plan.MirrorName != "" {
		_, _ = fmt.Fprintf(out, "Mirror: %s\n", plan.MirrorName)
	}
	if plan.LockfilePath != "" {
		_, _ = fmt.Fprintf(out, "Lockfile: %s\n", plan.LockfilePath)
	}
	_, _ = fmt.Fprintf(out, "Omit Optional: %t\n", plan.OmitOptional)
	_, _ = fmt.Fprintf(out, "Include Peer: %t\n", plan.IncludePeer)
	_, _ = fmt.Fprintf(out, "Include Dev: %t\n", plan.IncludeDev)
	_, _ = fmt.Fprintf(out, "Scripts: %s\n", blankIfEmpty(plan.ScriptsPolicy, "none"))
	_, _ = fmt.Fprintf(out, "Allow Scripts: %s\n", formatAllowedScriptPackages(plan.AllowScripts))
	_, _ = fmt.Fprintf(out, "Packages: %d\n", len(plan.Packages))
	for index, pkg := range plan.Packages {
		_, _ = fmt.Fprintf(out, "%d. %s@%s", index+1, pkg.Name, pkg.Version)
		if strings.TrimSpace(pkg.Requested) != "" {
			_, _ = fmt.Fprintf(out, " (requested %s)", pkg.Requested)
		}
		_, _ = fmt.Fprintln(out)
	}
}

func printInstallResult(out io.Writer, plan InstallPlan, result InstallResult) {
	_, _ = fmt.Fprintf(out, "Source: %s\n", plan.Source)
	_, _ = fmt.Fprintf(out, "Root: %s\n", plan.Root)
	_, _ = fmt.Fprintf(out, "Requested: %s\n", blankIfEmpty(plan.Requested, "latest"))
	_, _ = fmt.Fprintf(out, "Version: %s\n", blankIfEmpty(plan.RootVersion, "-"))
	_, _ = fmt.Fprintf(out, "Root Path: %s\n", result.RootPath)
	_, _ = fmt.Fprintf(out, "node_modules: %s\n", result.NodeModulesDir)
	_, _ = fmt.Fprintf(out, "Installed Packages: %d\n", len(result.Packages))
	_, _ = fmt.Fprintf(out, "Manifest: %s\n", result.ManifestPath)
	if result.LockfilePath != "" {
		_, _ = fmt.Fprintf(out, "Lockfile: %s\n", result.LockfilePath)
	}
	_, _ = fmt.Fprintf(out, "Duration: %s\n", roundDuration(result.Duration))
}

func normalizeScriptsPolicy(policy string) string {
	switch strings.ToLower(strings.TrimSpace(policy)) {
	case "", "none":
		return "none"
	case "root":
		return "root"
	case "all":
		return "all"
	default:
		return "none"
	}
}

func normalizeAllowedScriptPackages(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		name := strings.ToLower(strings.TrimSpace(value))
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		normalized = append(normalized, name)
	}
	sort.Strings(normalized)
	return normalized
}

func parseAllowedScriptPackagesFlag(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return normalizeAllowedScriptPackages(strings.Split(value, ","))
}

func allowedScriptPackageSet(values []string) map[string]struct{} {
	if len(values) == 0 {
		return nil
	}
	allowed := make(map[string]struct{}, len(values))
	for _, value := range normalizeAllowedScriptPackages(values) {
		allowed[value] = struct{}{}
	}
	return allowed
}

func formatAllowedScriptPackages(values []string) string {
	normalized := normalizeAllowedScriptPackages(values)
	if len(normalized) == 0 {
		return "-"
	}
	return strings.Join(normalized, ", ")
}

func equalStringSlices(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func resolveInstallLockfilePath(outputDir string, lockfilePath string) (string, error) {
	lockfilePath = strings.TrimSpace(lockfilePath)
	if lockfilePath != "" {
		return filepath.Abs(lockfilePath)
	}
	if strings.TrimSpace(outputDir) == "" {
		outputDir = "."
	}
	absOutput, err := filepath.Abs(outputDir)
	if err != nil {
		return "", fmt.Errorf("resolve install lockfile output dir: %w", err)
	}
	return filepath.Join(absOutput, "source-fetcher-install.lock.json"), nil
}

func loadInstallLockfile(path string) (InstallLockfile, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return InstallLockfile{}, err
	}
	var lockfile InstallLockfile
	if err := json.Unmarshal(raw, &lockfile); err != nil {
		return InstallLockfile{}, fmt.Errorf("parse install lockfile: %w", err)
	}
	if strings.TrimSpace(lockfile.Source) == "" || strings.TrimSpace(lockfile.Root) == "" || strings.TrimSpace(lockfile.RootVersion) == "" {
		return InstallLockfile{}, errors.New("install lockfile is missing source, root, or root_version")
	}
	return lockfile, nil
}

func matchesInstallLockfile(lockfile InstallLockfile, req InstallRequest, mirrorName string) bool {
	return installLockfileMismatchReason(lockfile, req, mirrorName) == ""
}

func installLockfileMismatchReason(lockfile InstallLockfile, req InstallRequest, mirrorName string) string {
	if !strings.EqualFold(strings.TrimSpace(lockfile.Source), "npm") {
		return fmt.Sprintf("source mismatch: lockfile=%q request=%q", lockfile.Source, req.Source)
	}
	if !strings.EqualFold(strings.TrimSpace(lockfile.Root), strings.TrimSpace(req.Name)) {
		return fmt.Sprintf("root mismatch: lockfile=%q request=%q", lockfile.Root, req.Name)
	}
	if lockfile.OmitOptional != req.OmitOptional || lockfile.IncludePeer != req.IncludePeer || lockfile.IncludeDev != req.IncludeDev {
		return "dependency mode flags differ"
	}
	if !equalStringSlices(normalizeAllowedScriptPackages(lockfile.AllowScripts), normalizeAllowedScriptPackages(req.AllowScripts)) {
		return "allow_scripts differ"
	}
	if strings.TrimSpace(req.Mirror) != "" && !strings.EqualFold(strings.TrimSpace(lockfile.MirrorName), strings.TrimSpace(mirrorName)) {
		return fmt.Sprintf("mirror mismatch: lockfile=%q request=%q", lockfile.MirrorName, mirrorName)
	}
	requested := strings.TrimSpace(req.Version)
	if requested == "" || strings.EqualFold(requested, "latest") {
		if len(lockfile.Packages) == 0 {
			return "lockfile has no packages"
		}
		return ""
	}
	if !matchesNPMVersionRange(lockfile.RootVersion, requested) {
		return fmt.Sprintf("version mismatch: lockfile=%q request=%q", lockfile.RootVersion, requested)
	}
	return ""
}

func installPlanFromLockfile(lockfile InstallLockfile, lockfilePath string, scriptsPolicy string) InstallPlan {
	packages := make([]InstallPackage, len(lockfile.Packages))
	copy(packages, lockfile.Packages)
	return InstallPlan{
		Source:        lockfile.Source,
		Root:          lockfile.Root,
		Requested:     lockfile.Requested,
		RootVersion:   lockfile.RootVersion,
		MirrorName:    lockfile.MirrorName,
		LockfilePath:  lockfilePath,
		OmitOptional:  lockfile.OmitOptional,
		IncludePeer:   lockfile.IncludePeer,
		IncludeDev:    lockfile.IncludeDev,
		ScriptsPolicy: normalizeScriptsPolicy(scriptsPolicy),
		AllowScripts:  normalizeAllowedScriptPackages(lockfile.AllowScripts),
		Packages:      packages,
	}
}

func writeInstallLockfile(path string, plan InstallPlan) error {
	tempPath, err := writeInstallLockfileTemp(path, plan)
	if err != nil {
		return err
	}
	defer os.Remove(tempPath)
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("commit install lockfile: %w", err)
	}
	return nil
}

func marshalInstallLockfile(plan InstallPlan) ([]byte, error) {
	lockfile := InstallLockfile{
		Version:      1,
		Source:       plan.Source,
		Root:         plan.Root,
		Requested:    plan.Requested,
		RootVersion:  plan.RootVersion,
		MirrorName:   plan.MirrorName,
		GeneratedAt:  time.Now().UTC().Format(time.RFC3339),
		OmitOptional: plan.OmitOptional,
		IncludePeer:  plan.IncludePeer,
		IncludeDev:   plan.IncludeDev,
		AllowScripts: normalizeAllowedScriptPackages(plan.AllowScripts),
		Packages:     plan.Packages,
	}
	raw, err := json.MarshalIndent(lockfile, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal install lockfile: %w", err)
	}
	return raw, nil
}

func ensureInstallPackageArchive(ctx context.Context, client *http.Client, plan InstallPlan, pkg InstallPackage, options DownloadOptions) (DownloadResult, error) {
	filename := sanitizeFileName(strings.TrimSpace(pkg.Filename))
	if filename == "" {
		filename = filenameFromURLOrFallback(pkg.URL, sanitizeFileName(pkg.Name+"-"+pkg.Version+".tgz"))
	}
	targetPath := filepath.Join(options.OutputDir, filename)
	if stat, err := os.Stat(targetPath); err == nil && !stat.IsDir() {
		digests, verifyErr := verifyInstallPackageArchive(targetPath, pkg)
		if verifyErr == nil {
			options.RequestOptions.Verbose.Logf("integrity", "verified cached archive %s@%s using %s", pkg.Name, pkg.Version, installIntegrityCheckSummary(pkg))
			return DownloadResult{
				Path:     targetPath,
				Size:     digests.Size,
				SHA256:   digests.SHA256,
				Duration: 0,
			}, nil
		}
		options.RequestOptions.Verbose.Logf("integrity", "cached archive %s@%s failed verification: %v", pkg.Name, pkg.Version, verifyErr)
		if err := os.Remove(targetPath); err != nil {
			return DownloadResult{}, fmt.Errorf("remove invalid cached archive %s: %w", targetPath, err)
		}
	}

	downloaded, err := downloadPlan(
		ctx,
		client,
		DownloadPlan{
			Source:     plan.Source,
			Identifier: pkg.Name,
			Version:    pkg.Version,
			URL:        pkg.URL,
			Filename:   filename,
			MirrorName: plan.MirrorName,
		},
		options,
	)
	if err != nil {
		return DownloadResult{}, err
	}
	if _, err := verifyInstallPackageArchive(downloaded.Path, pkg); err != nil {
		options.RequestOptions.Verbose.Logf("integrity", "downloaded archive %s@%s failed verification: %v", pkg.Name, pkg.Version, err)
		return DownloadResult{}, err
	}
	options.RequestOptions.Verbose.Logf("integrity", "verified downloaded archive %s@%s using %s", pkg.Name, pkg.Version, installIntegrityCheckSummary(pkg))
	return downloaded, nil
}

func installIntegrityCheckSummary(pkg InstallPackage) string {
	checks := make([]string, 0, 2)
	if strings.TrimSpace(pkg.Shasum) != "" {
		checks = append(checks, "shasum")
	}
	if strings.TrimSpace(pkg.Integrity) != "" {
		checks = append(checks, "integrity")
	}
	if len(checks) == 0 {
		return "sha256 digest"
	}
	return strings.Join(checks, " + ")
}

func verifyInstallPackageArchive(path string, pkg InstallPackage) (archiveDigests, error) {
	return verifyFileIntegrity(
		path,
		pkg.Shasum,
		pkg.Integrity,
		fmt.Sprintf("archive for %s@%s", pkg.Name, pkg.Version),
	)
}

func verifySubresourceIntegrity(integrity string, digests archiveDigests) error {
	tokens := strings.Fields(strings.TrimSpace(integrity))
	if len(tokens) == 0 {
		return nil
	}
	supported := false
	for _, token := range tokens {
		algorithm, expectedValue, ok := strings.Cut(strings.TrimSpace(token), "-")
		if !ok || strings.TrimSpace(expectedValue) == "" {
			continue
		}
		var actual string
		switch strings.ToLower(strings.TrimSpace(algorithm)) {
		case "sha1":
			actual = digests.SHA1
			supported = true
		case "sha256":
			actual = digests.SHA256
			supported = true
		case "sha512":
			actual = digests.SHA512
			supported = true
		default:
			continue
		}
		actualBytes, err := hex.DecodeString(actual)
		if err != nil {
			return fmt.Errorf("decode %s digest: %w", algorithm, err)
		}
		expectedBytes, err := decodeBase64Digest(expectedValue)
		if err != nil {
			return fmt.Errorf("decode integrity digest: %w", err)
		}
		if bytes.Equal(actualBytes, expectedBytes) {
			return nil
		}
	}
	if supported {
		return errors.New("no integrity digest matched")
	}
	return fmt.Errorf("unsupported integrity format %q", integrity)
}

func decodeBase64Digest(value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, errors.New("empty digest")
	}
	if decoded, err := base64.StdEncoding.DecodeString(value); err == nil {
		return decoded, nil
	}
	return base64.RawStdEncoding.DecodeString(value)
}

func resolveInstallManifestPath(outputDir string, manifestPath string, action string) (string, error) {
	manifestPath = strings.TrimSpace(manifestPath)
	if manifestPath != "" {
		return filepath.Abs(manifestPath)
	}
	outputDir = strings.TrimSpace(outputDir)
	if outputDir == "" {
		outputDir = "."
	}
	absOutput, err := filepath.Abs(outputDir)
	if err != nil {
		return "", fmt.Errorf("resolve %s output dir: %w", action, err)
	}
	return filepath.Join(absOutput, "source-fetcher-install.json"), nil
}

func loadInstallManifest(manifestPath string) (InstallManifest, error) {
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		return InstallManifest{}, fmt.Errorf("read install manifest: %w", err)
	}
	var manifest InstallManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return InstallManifest{}, fmt.Errorf("parse install manifest: %w", err)
	}
	if strings.TrimSpace(manifest.RootPath) == "" || strings.TrimSpace(manifest.NodeModulesDir) == "" {
		return InstallManifest{}, errors.New("install manifest is missing root_path or node_modules_dir")
	}
	return manifest, nil
}

func executeUninstall(req UninstallRequest) (UninstallResult, error) {
	start := time.Now()
	manifestPath, err := resolveInstallManifestPath(req.OutputDir, req.ManifestPath, "uninstall")
	if err != nil {
		return UninstallResult{}, err
	}
	manifest, err := loadInstallManifest(manifestPath)
	if err != nil {
		return UninstallResult{}, err
	}

	baseDir := filepath.Dir(manifestPath)
	result := UninstallResult{
		ManifestPath: manifestPath,
	}
	removed := map[string]struct{}{}
	missing := map[string]struct{}{}
	skipped := map[string]struct{}{}
	recordRemoved := func(target string) {
		if target == "" {
			return
		}
		target = filepath.Clean(target)
		if _, ok := removed[target]; ok {
			return
		}
		removed[target] = struct{}{}
		result.RemovedPaths = append(result.RemovedPaths, target)
	}
	recordMissing := func(target string) {
		if target == "" {
			return
		}
		target = filepath.Clean(target)
		if _, ok := missing[target]; ok {
			return
		}
		missing[target] = struct{}{}
		result.MissingPaths = append(result.MissingPaths, target)
	}
	recordSkipped := func(target string) {
		if target == "" {
			return
		}
		target = filepath.Clean(target)
		if _, ok := skipped[target]; ok {
			return
		}
		skipped[target] = struct{}{}
		result.SkippedPaths = append(result.SkippedPaths, target)
	}

	for _, entry := range uninstallPackagePaths(manifest.Packages) {
		installPath := filepath.Clean(entry.Path)
		if !isPathWithinBase(installPath, baseDir) && !sameFilePath(installPath, baseDir) {
			recordSkipped(installPath)
			continue
		}
		status, err := removeInstalledPackagePath(installPath, entry.Record)
		if err != nil {
			return UninstallResult{}, err
		}
		switch status {
		case "removed":
			recordRemoved(installPath)
			cleanupEmptyParentDirectories(filepath.Dir(installPath), baseDir, &result, recordRemoved)
		case "missing":
			recordMissing(installPath)
		case "skipped":
			recordSkipped(installPath)
		}
	}
	for _, pkg := range manifest.Packages {
		for _, binLink := range pkg.BinLinks {
			binLink = filepath.Clean(binLink)
			if !isManagedManifestPath(binLink, baseDir) {
				recordSkipped(binLink)
				continue
			}
			status, err := removePathIfPresent(binLink)
			if err != nil {
				return UninstallResult{}, err
			}
			switch status {
			case "removed":
				recordRemoved(binLink)
				cleanupEmptyParentDirectories(filepath.Dir(binLink), baseDir, &result, recordRemoved)
			case "missing":
				recordMissing(binLink)
			}
		}
	}

	if !req.KeepCache {
		for _, pkg := range manifest.Packages {
			for _, cachePath := range []string{pkg.StorePath, pkg.ArchivePath} {
				if !isManagedManifestPath(cachePath, baseDir) {
					recordSkipped(cachePath)
					continue
				}
				status, err := removePathIfPresent(cachePath)
				if err != nil {
					return UninstallResult{}, err
				}
				switch status {
				case "removed":
					recordRemoved(cachePath)
					cleanupEmptyParentDirectories(filepath.Dir(cachePath), baseDir, &result, recordRemoved)
				case "missing":
					recordMissing(cachePath)
				}
			}
		}
		for _, dir := range []string{manifest.StoreDir, manifest.CacheDir, filepath.Join(baseDir, ".source-fetcher")} {
			if !isManagedManifestPath(dir, baseDir) {
				recordSkipped(dir)
				continue
			}
			status, err := removeDirectoryIfEmpty(dir)
			if err != nil {
				return UninstallResult{}, err
			}
			if status == "removed" {
				recordRemoved(dir)
			}
		}
	}

	status, err := removeDirectoryIfEmpty(manifest.NodeModulesDir)
	if err != nil {
		return UninstallResult{}, err
	}
	if status == "removed" {
		recordRemoved(manifest.NodeModulesDir)
	}
	if !req.KeepManifest {
		status, err = removePathIfPresent(manifestPath)
		if err != nil {
			return UninstallResult{}, err
		}
		switch status {
		case "removed":
			recordRemoved(manifestPath)
		case "missing":
			recordMissing(manifestPath)
		}
	}

	sort.Strings(result.RemovedPaths)
	sort.Strings(result.MissingPaths)
	sort.Strings(result.SkippedPaths)
	result.Duration = time.Since(start)
	return result, nil
}

func executeRepair(req RepairRequest) (RepairResult, error) {
	start := time.Now()
	manifestPath, err := resolveInstallManifestPath(req.OutputDir, req.ManifestPath, "repair")
	if err != nil {
		return RepairResult{}, err
	}
	manifest, err := loadInstallManifest(manifestPath)
	if err != nil {
		return RepairResult{}, err
	}

	baseDir := filepath.Dir(manifestPath)
	result := RepairResult{ManifestPath: manifestPath}
	repaired := map[string]struct{}{}
	healthy := map[string]struct{}{}
	missingCache := map[string]struct{}{}
	skipped := map[string]struct{}{}
	record := func(target string, bucket map[string]struct{}, values *[]string) {
		if target == "" {
			return
		}
		target = filepath.Clean(target)
		if _, ok := bucket[target]; ok {
			return
		}
		bucket[target] = struct{}{}
		*values = append(*values, target)
	}

	for _, pkg := range manifest.Packages {
		for _, installPath := range pkg.InstallPaths {
			installPath = filepath.Clean(installPath)
			if !isPathWithinBase(installPath, baseDir) && !sameFilePath(installPath, baseDir) {
				record(installPath, skipped, &result.SkippedPaths)
				continue
			}
			status, err := ensureInstalledPackagePath(pkg, installPath, baseDir)
			if err != nil {
				return RepairResult{}, err
			}
			switch status {
			case "healthy":
				record(installPath, healthy, &result.HealthyPaths)
			case "repaired":
				record(installPath, repaired, &result.RepairedPaths)
			case "missing_cache":
				record(installPath, missingCache, &result.MissingCachePaths)
			case "skipped":
				record(installPath, skipped, &result.SkippedPaths)
			}
		}
	}

	sort.Strings(result.RepairedPaths)
	sort.Strings(result.HealthyPaths)
	sort.Strings(result.MissingCachePaths)
	sort.Strings(result.SkippedPaths)
	result.Duration = time.Since(start)
	return result, nil
}

func printUninstallResult(out io.Writer, result UninstallResult) {
	_, _ = fmt.Fprintf(out, "Manifest: %s\n", result.ManifestPath)
	_, _ = fmt.Fprintf(out, "Removed Paths: %d\n", len(result.RemovedPaths))
	_, _ = fmt.Fprintf(out, "Missing Paths: %d\n", len(result.MissingPaths))
	_, _ = fmt.Fprintf(out, "Skipped Paths: %d\n", len(result.SkippedPaths))
	_, _ = fmt.Fprintf(out, "Duration: %s\n", roundDuration(result.Duration))
}

func printRepairResult(out io.Writer, result RepairResult) {
	_, _ = fmt.Fprintf(out, "Manifest: %s\n", result.ManifestPath)
	_, _ = fmt.Fprintf(out, "Repaired Paths: %d\n", len(result.RepairedPaths))
	_, _ = fmt.Fprintf(out, "Healthy Paths: %d\n", len(result.HealthyPaths))
	_, _ = fmt.Fprintf(out, "Missing Cache Paths: %d\n", len(result.MissingCachePaths))
	_, _ = fmt.Fprintf(out, "Skipped Paths: %d\n", len(result.SkippedPaths))
	_, _ = fmt.Fprintf(out, "Duration: %s\n", roundDuration(result.Duration))
}

func uninstallPackagePaths(packages []InstalledPackage) []uninstallPackagePath {
	paths := []uninstallPackagePath{}
	for _, pkg := range packages {
		for _, installPath := range pkg.InstallPaths {
			paths = append(paths, uninstallPackagePath{
				Path:   filepath.Clean(installPath),
				Record: pkg,
			})
		}
	}
	sort.SliceStable(paths, func(i, j int) bool {
		leftDepth := installPathDepth(paths[i].Path)
		rightDepth := installPathDepth(paths[j].Path)
		if leftDepth != rightDepth {
			return leftDepth > rightDepth
		}
		return paths[i].Path > paths[j].Path
	})
	return paths
}

func installPathDepth(value string) int {
	value = filepath.Clean(value)
	if value == "." || value == string(os.PathSeparator) {
		return 0
	}
	depth := 0
	for _, part := range strings.Split(value, string(os.PathSeparator)) {
		if part != "" {
			depth++
		}
	}
	return depth
}

func sortedMapKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

func pickNPMVersionFromRange(metadata npmPackageMetadata, requested string) (string, npmVersionDetails, bool) {
	requested = strings.TrimSpace(requested)
	if requested == "" {
		return "", npmVersionDetails{}, false
	}
	for _, version := range sortedNPMVersions(metadata) {
		details := metadata.Versions[version]
		if matchesNPMVersionRange(version, requested) {
			return version, details, true
		}
	}
	return "", npmVersionDetails{}, false
}

func sortedNPMVersions(metadata npmPackageMetadata) []string {
	versions := make([]string, 0, len(metadata.Versions))
	for version := range metadata.Versions {
		versions = append(versions, version)
	}
	sort.SliceStable(versions, func(i, j int) bool {
		return compareVersionish(normalizeVersionConstraint(versions[i]), normalizeVersionConstraint(versions[j])) > 0
	})
	return versions
}

func matchesNPMVersionRange(version string, requested string) bool {
	requested = strings.TrimSpace(requested)
	if requested == "" || requested == "*" || strings.EqualFold(requested, "latest") {
		return true
	}
	if strings.Contains(requested, "||") {
		return false
	}
	version = normalizeVersionConstraint(version)
	requested = strings.ReplaceAll(requested, ",", " ")
	tokens := strings.Fields(requested)
	if len(tokens) == 0 {
		return false
	}
	for _, token := range tokens {
		if !matchesNPMSingleConstraint(version, token) {
			return false
		}
	}
	return true
}

func matchesNPMSingleConstraint(version string, token string) bool {
	token = strings.TrimSpace(token)
	switch {
	case token == "", token == "*", token == "x", token == "X":
		return true
	case strings.HasPrefix(token, "^"):
		return matchesNPMCaretRange(version, token[1:])
	case strings.HasPrefix(token, "~"):
		return matchesNPMTildeRange(version, token[1:])
	case strings.HasPrefix(token, ">="):
		return compareVersionish(version, normalizeVersionConstraint(token[2:])) >= 0
	case strings.HasPrefix(token, "<="):
		return compareVersionish(version, normalizeVersionConstraint(token[2:])) <= 0
	case strings.HasPrefix(token, ">"):
		return compareVersionish(version, normalizeVersionConstraint(token[1:])) > 0
	case strings.HasPrefix(token, "<"):
		return compareVersionish(version, normalizeVersionConstraint(token[1:])) < 0
	case strings.ContainsAny(token, "*xX"):
		return matchesNPMWildcardRange(version, token)
	default:
		return compareVersionish(version, normalizeVersionConstraint(token)) == 0
	}
}

func matchesNPMCaretRange(version string, base string) bool {
	base = normalizeVersionConstraint(base)
	if base == "" {
		return true
	}
	if compareVersionish(version, base) < 0 {
		return false
	}
	major, minor, patch := parseNPMVersionCore(base)
	switch {
	case major > 0:
		return compareVersionish(version, fmt.Sprintf("%d.0.0", major+1)) < 0
	case minor > 0:
		return compareVersionish(version, fmt.Sprintf("0.%d.0", minor+1)) < 0
	default:
		return compareVersionish(version, fmt.Sprintf("0.0.%d", patch+1)) < 0
	}
}

func matchesNPMTildeRange(version string, base string) bool {
	base = normalizeVersionConstraint(base)
	if base == "" {
		return true
	}
	if compareVersionish(version, base) < 0 {
		return false
	}
	major, minor, _ := parseNPMVersionCore(base)
	segments := parseVersionish(base).core
	if len(segments) <= 1 {
		return compareVersionish(version, fmt.Sprintf("%d.0.0", major+1)) < 0
	}
	return compareVersionish(version, fmt.Sprintf("%d.%d.0", major, minor+1)) < 0
}

func matchesNPMWildcardRange(version string, token string) bool {
	token = normalizeVersionConstraint(token)
	parts := strings.Split(token, ".")
	versionParts := parseVersionish(version).core
	for index, part := range parts {
		if part == "*" || strings.EqualFold(part, "x") {
			return true
		}
		current, ok := versionPartAt(versionParts, index)
		if !ok || current != part {
			return false
		}
	}
	return len(versionParts) == len(parts)
}

func normalizeVersionConstraint(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "=")
	value = strings.TrimPrefix(value, "v")
	value = strings.TrimPrefix(value, "V")
	return strings.TrimSpace(value)
}

func parseNPMVersionCore(value string) (int, int, int) {
	parts := parseVersionish(normalizeVersionConstraint(value)).core
	read := func(index int) int {
		part, ok := versionPartAt(parts, index)
		if !ok {
			return 0
		}
		number, err := strconv.Atoi(part)
		if err != nil {
			return 0
		}
		return number
	}
	return read(0), read(1), read(2)
}

func installPackageKey(name string, version string) string {
	return strings.ToLower(strings.TrimSpace(name)) + "@" + strings.TrimSpace(version)
}

func installStoreDirectoryName(name string, version string) string {
	return sanitizeFileName(strings.ReplaceAll(strings.TrimSpace(name), "/", "__") + "-" + strings.TrimSpace(version))
}

func packagePathElements(name string) []string {
	parts := strings.Split(strings.TrimSpace(name), "/")
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		filtered = append(filtered, part)
	}
	return filtered
}

func packageRelativeInstallPath(name string) string {
	return filepath.Join(packagePathElements(name)...)
}

func materializeNPMPackage(archivePath string, storePath string, expectedName string, expectedVersion string) error {
	if manifest, err := readNPMPackageArchiveManifest(storePath); err == nil {
		if strings.TrimSpace(manifest.Version) == strings.TrimSpace(expectedVersion) {
			return nil
		}
	}
	if err := os.MkdirAll(filepath.Dir(storePath), 0o755); err != nil {
		return fmt.Errorf("create store parent dir %s: %w", filepath.Dir(storePath), err)
	}
	stagingRoot, err := os.MkdirTemp(filepath.Dir(storePath), "."+filepath.Base(storePath)+".source-fetcher-store-*")
	if err != nil {
		return fmt.Errorf("create store staging dir for %s: %w", storePath, err)
	}
	defer os.RemoveAll(stagingRoot)
	stagedStorePath := filepath.Join(stagingRoot, "store")
	if err := os.MkdirAll(stagedStorePath, 0o755); err != nil {
		return fmt.Errorf("create staged store path %s: %w", stagedStorePath, err)
	}
	if err := extractNPMPackageArchive(archivePath, stagedStorePath); err != nil {
		return err
	}
	manifest, err := readNPMPackageArchiveManifest(stagedStorePath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(manifest.Version) != strings.TrimSpace(expectedVersion) {
		return fmt.Errorf("archive package version mismatch: expected %s, got %s", expectedVersion, manifest.Version)
	}
	if err := commitInstallArtifacts([]installCommitItem{{StagedPath: stagedStorePath, TargetPath: storePath}}); err != nil {
		return fmt.Errorf("commit store path %s: %w", storePath, err)
	}
	return nil
}

func readNPMPackageArchiveManifest(storePath string) (npmPackageArchiveManifest, error) {
	raw, err := os.ReadFile(filepath.Join(storePath, "package.json"))
	if err != nil {
		return npmPackageArchiveManifest{}, fmt.Errorf("read package manifest: %w", err)
	}
	var manifest npmPackageArchiveManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return npmPackageArchiveManifest{}, fmt.Errorf("parse package manifest: %w", err)
	}
	if strings.TrimSpace(manifest.Name) == "" || strings.TrimSpace(manifest.Version) == "" {
		return npmPackageArchiveManifest{}, errors.New("package manifest is missing name or version")
	}
	return manifest, nil
}

func extractNPMPackageArchive(archivePath string, destDir string) error {
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
		relativePath, include, err := trimNPMPackageArchivePath(header.Name)
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
			if err := writeArchiveFile(reader, targetPath, header.FileInfo().Mode().Perm()); err != nil {
				return err
			}
		case tar.TypeSymlink:
			return fmt.Errorf("archive symlinks are not supported: %s", header.Name)
		default:
			continue
		}
	}
}

func trimNPMPackageArchivePath(name string) (string, bool, error) {
	cleaned := strings.TrimSpace(strings.ReplaceAll(name, "\\", "/"))
	cleaned = strings.TrimPrefix(cleaned, "./")
	if cleaned == "" {
		return "", false, nil
	}
	cleaned = path.Clean(cleaned)
	if cleaned == "." || cleaned == "/" {
		return "", false, nil
	}
	if cleaned == "package" {
		return "", false, nil
	}
	if strings.HasPrefix(cleaned, "package/") {
		cleaned = strings.TrimPrefix(cleaned, "package/")
	}
	if cleaned == "" || cleaned == "." {
		return "", false, nil
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", false, fmt.Errorf("archive entry escapes package root: %s", name)
	}
	return cleaned, true, nil
}

func writeArchiveFile(reader io.Reader, targetPath string, mode fs.FileMode) error {
	file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fileModeOrDefault(mode, 0o644))
	if err != nil {
		return fmt.Errorf("create file %s: %w", targetPath, err)
	}
	defer file.Close()
	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("write file %s: %w", targetPath, err)
	}
	return nil
}

func (e *npmInstallExecutor) installPackageTree(pkg InstallPackage, installPath string) error {
	key := installPackageKey(pkg.Name, pkg.Version)
	if existingKey, ok := e.placedByPath[installPath]; ok {
		if existingKey == key {
			return nil
		}
		return fmt.Errorf("install path conflict at %s between %s and %s", installPath, existingKey, key)
	}
	record := e.materialized[key]
	if record == nil {
		return fmt.Errorf("package %s is missing extracted contents", key)
	}
	if err := copyDirectory(record.StorePath, installPath); err != nil {
		return fmt.Errorf("copy %s into %s: %w", key, installPath, err)
	}
	e.placedByPath[installPath] = key
	record.InstallPaths = appendUniqueString(record.InstallPaths, installPath)
	if err := e.ensurePackageBinLinks(record, installPath); err != nil {
		return err
	}

	for _, dependencyName := range sortedMapKeys(pkg.ResolvedDependencies) {
		dependencyVersion := pkg.ResolvedDependencies[dependencyName]
		dependencyKey := installPackageKey(dependencyName, dependencyVersion)
		dependencyPackage, ok := e.packageByKey[dependencyKey]
		if !ok {
			return fmt.Errorf("resolved dependency %s for %s is missing from the install plan", dependencyKey, key)
		}
		if reusePath := e.findReusableInstallPath(installPath, dependencyPackage.Name, dependencyPackage.Version); reusePath != "" {
			if reusedRecord := e.materialized[dependencyKey]; reusedRecord != nil {
				reusedRecord.InstallPaths = appendUniqueString(reusedRecord.InstallPaths, reusePath)
			}
			continue
		}
		dependencyPath := filepath.Join(installPath, "node_modules", packageRelativeInstallPath(dependencyPackage.Name))
		if err := e.installPackageTree(dependencyPackage, dependencyPath); err != nil {
			return err
		}
	}
	return nil
}

func (e *npmInstallExecutor) ensurePackageBinLinks(record *InstalledPackage, installPath string) error {
	manifest, err := readNPMPackageArchiveManifest(installPath)
	if err != nil {
		return fmt.Errorf("read installed manifest for %s: %w", installPath, err)
	}
	created, binDir, err := createNPMPackageBinLinks(installPath, manifest)
	if err != nil {
		return err
	}
	if binDir != "" {
		record.BinDir = binDir
	}
	for _, link := range created {
		record.BinLinks = appendUniqueString(record.BinLinks, link)
	}
	return nil
}

func (e *npmInstallExecutor) runLifecycleScripts(ctx context.Context, packages []InstallPackage, rootKey string) error {
	if e.scriptsPolicy == "none" {
		return nil
	}
	for index := len(packages) - 1; index >= 0; index-- {
		pkg := packages[index]
		key := installPackageKey(pkg.Name, pkg.Version)
		if e.scriptsPolicy == "root" && key != rootKey {
			continue
		}
		record := e.materialized[key]
		if record == nil {
			continue
		}
		paths := append([]string(nil), record.InstallPaths...)
		sort.Strings(paths)
		for _, installPath := range paths {
			manifest, err := readNPMPackageArchiveManifest(installPath)
			if err != nil {
				return fmt.Errorf("read lifecycle manifest for %s: %w", installPath, err)
			}
			for _, event := range []string{"preinstall", "install", "postinstall"} {
				command := strings.TrimSpace(manifest.Scripts[event])
				if command == "" {
					continue
				}
				if shouldBlockNPMLifecycleScript(event, record.Name, e.allowScripts) {
					e.verbose.Logf("scripts", "blocked %s for %s at %s", event, record.Name, installPath)
					warnBlockedNPMLifecycleScript(record.Name, event, installPath)
					continue
				}
				e.verbose.Logf("scripts", "running %s for %s at %s", event, record.Name, installPath)
				if err := runNPMLifecycleScript(ctx, installPath, event, command, e.outputDir); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func installRequestLabel(req InstallRequest) string {
	version := blankIfEmpty(strings.TrimSpace(req.Version), "latest")
	return strings.TrimSpace(req.Name) + "@" + version
}

func shouldBlockNPMLifecycleScript(event string, packageName string, allowed map[string]struct{}) bool {
	if !strings.EqualFold(strings.TrimSpace(event), "preinstall") {
		return false
	}
	if len(allowed) == 0 {
		return true
	}
	_, ok := allowed[strings.ToLower(strings.TrimSpace(packageName))]
	return !ok
}

func warnBlockedNPMLifecycleScript(packageName string, event string, installPath string) {
	_, _ = fmt.Fprintf(
		os.Stderr,
		"Warning: blocked %s script for %s at %s; package may not function correctly; allow with --allow-scripts=%s\n",
		strings.TrimSpace(event),
		blankIfEmpty(strings.TrimSpace(packageName), filepath.Base(installPath)),
		installPath,
		strings.ToLower(blankIfEmpty(strings.TrimSpace(packageName), filepath.Base(installPath))),
	)
}

func createNPMPackageBinLinks(installPath string, manifest npmPackageArchiveManifest) ([]string, string, error) {
	if len(manifest.Bin) == 0 {
		return nil, "", nil
	}
	nodeModulesDir := findNearestNodeModulesDir(installPath)
	if nodeModulesDir == "" {
		return nil, "", fmt.Errorf("unable to locate node_modules for %s", installPath)
	}
	binDir := filepath.Join(nodeModulesDir, ".bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return nil, "", fmt.Errorf("create bin dir %s: %w", binDir, err)
	}
	created := []string{}
	for _, binName := range sortedMapKeys(manifest.Bin) {
		targetValue := strings.TrimSpace(manifest.Bin[binName])
		if targetValue == "" {
			continue
		}
		actualName := normalizeBinName(binName, manifest.Name)
		targetPath := filepath.Clean(filepath.Join(installPath, filepath.FromSlash(targetValue)))
		if !isPathWithinBase(targetPath, installPath) && !sameFilePath(targetPath, installPath) {
			return nil, "", fmt.Errorf("bin target escapes package path for %s: %s", manifest.Name, targetValue)
		}
		links, err := writeNodeBinShims(binDir, actualName, targetPath)
		if err != nil {
			return nil, "", err
		}
		created = append(created, links...)
	}
	sort.Strings(created)
	return created, binDir, nil
}

func normalizeBinName(value string, packageName string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return value
	}
	parts := packagePathElements(packageName)
	if len(parts) == 0 {
		return "package"
	}
	return parts[len(parts)-1]
}

func findNearestNodeModulesDir(installPath string) string {
	current := filepath.Clean(installPath)
	for {
		if strings.EqualFold(filepath.Base(current), "node_modules") {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			return ""
		}
		current = parent
	}
}

func writeNodeBinShims(binDir string, name string, targetPath string) ([]string, error) {
	relativeTarget, err := filepath.Rel(binDir, targetPath)
	if err != nil {
		return nil, fmt.Errorf("compute bin target relative path for %s: %w", targetPath, err)
	}
	relativeTarget = filepath.Clean(relativeTarget)
	links := []string{}
	if runtime.GOOS == "windows" {
		cmdPath := filepath.Join(binDir, name+".cmd")
		ps1Path := filepath.Join(binDir, name+".ps1")
		cmdContent := "@ECHO off\r\nnode \"%~dp0" + strings.ReplaceAll(relativeTarget, "/", "\\") + "\" %*\r\n"
		ps1Content := "& node \"$PSScriptRoot\\" + strings.ReplaceAll(relativeTarget, "/", "\\") + "\" @args\r\n"
		for _, output := range []struct {
			path    string
			content string
		}{
			{path: cmdPath, content: cmdContent},
			{path: ps1Path, content: ps1Content},
		} {
			if err := os.WriteFile(output.path, []byte(output.content), 0o755); err != nil {
				return nil, fmt.Errorf("write bin shim %s: %w", output.path, err)
			}
			links = append(links, output.path)
		}
		return links, nil
	}
	shimPath := filepath.Join(binDir, name)
	content := "#!/bin/sh\nexec node \"$(dirname \"$0\")/" + filepath.ToSlash(relativeTarget) + "\" \"$@\"\n"
	if err := os.WriteFile(shimPath, []byte(content), 0o755); err != nil {
		return nil, fmt.Errorf("write bin shim %s: %w", shimPath, err)
	}
	return []string{shimPath}, nil
}

func expectedNodeBinShimPaths(binDir string, name string) []string {
	if runtime.GOOS == "windows" {
		return []string{
			filepath.Join(binDir, name+".cmd"),
			filepath.Join(binDir, name+".ps1"),
		}
	}
	return []string{filepath.Join(binDir, name)}
}

func runNPMLifecycleScript(ctx context.Context, installPath string, event string, command string, outputDir string) error {
	cmd := lifecycleScriptCommand(ctx, command)
	cmd.Dir = installPath
	cmd.Env = append(os.Environ(),
		"npm_lifecycle_event="+event,
		"npm_package_json="+filepath.Join(installPath, "package.json"),
		"PATH="+strings.Join(append(npmLifecyclePathEntries(installPath, outputDir), os.Getenv("PATH")), string(os.PathListSeparator)),
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message != "" {
			return fmt.Errorf("run %s script for %s: %w: %s", event, installPath, err, message)
		}
		return fmt.Errorf("run %s script for %s: %w", event, installPath, err)
	}
	return nil
}

func lifecycleScriptCommand(ctx context.Context, script string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.CommandContext(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	}
	return exec.CommandContext(ctx, "sh", "-c", script)
}

func npmLifecyclePathEntries(installPath string, outputDir string) []string {
	entries := []string{}
	current := filepath.Dir(filepath.Clean(installPath))
	for {
		if strings.EqualFold(filepath.Base(current), "node_modules") {
			entries = appendUniqueString(entries, filepath.Join(current, ".bin"))
		}
		if sameFilePath(current, outputDir) {
			break
		}
		parent := filepath.Dir(current)
		if parent == current || !isPathWithinBase(parent, outputDir) {
			break
		}
		current = parent
	}
	return appendUniqueString(entries, filepath.Join(outputDir, "node_modules", ".bin"))
}

func (e *npmInstallExecutor) findReusableInstallPath(parentPackagePath string, name string, version string) string {
	key := installPackageKey(name, version)
	for _, candidate := range npmDependencyLookupCandidates(parentPackagePath, name, e.outputDir) {
		if existingKey, ok := e.placedByPath[candidate]; ok && existingKey == key {
			return candidate
		}
	}
	return ""
}

func npmDependencyLookupCandidates(startPath string, dependencyName string, outputDir string) []string {
	relativePath := packageRelativeInstallPath(dependencyName)
	candidates := []string{}
	current := startPath
	for {
		if !strings.EqualFold(filepath.Base(current), "node_modules") {
			candidate := filepath.Join(current, "node_modules", relativePath)
			if len(candidates) == 0 || !sameFilePath(candidates[len(candidates)-1], candidate) {
				candidates = append(candidates, candidate)
			}
		}
		if sameFilePath(current, outputDir) {
			break
		}
		parent := filepath.Dir(current)
		if parent == current || !isPathWithinBase(parent, outputDir) {
			break
		}
		current = parent
	}
	rootCandidate := filepath.Join(outputDir, "node_modules", relativePath)
	if len(candidates) == 0 || !sameFilePath(candidates[len(candidates)-1], rootCandidate) {
		candidates = append(candidates, rootCandidate)
	}
	return candidates
}

func copyDirectory(sourceDir string, targetDir string) error {
	if err := os.RemoveAll(targetDir); err != nil {
		return fmt.Errorf("reset target dir %s: %w", targetDir, err)
	}
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("create target dir %s: %w", targetDir, err)
	}
	return filepath.WalkDir(sourceDir, func(currentPath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relativePath, err := filepath.Rel(sourceDir, currentPath)
		if err != nil {
			return fmt.Errorf("compute relative path for %s: %w", currentPath, err)
		}
		if relativePath == "." {
			return nil
		}
		targetPath := filepath.Join(targetDir, relativePath)
		if entry.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			linkTarget, err := os.Readlink(currentPath)
			if err != nil {
				return fmt.Errorf("read symlink %s: %w", currentPath, err)
			}
			if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
				return fmt.Errorf("create parent dir for symlink %s: %w", targetPath, err)
			}
			if err := os.Symlink(linkTarget, targetPath); err != nil {
				return fmt.Errorf("create symlink %s: %w", targetPath, err)
			}
			return nil
		}
		return copyFile(currentPath, targetPath, info.Mode().Perm())
	})
}

func copyFile(sourcePath string, targetPath string, mode fs.FileMode) error {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("open %s: %w", sourcePath, err)
	}
	defer sourceFile.Close()
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return fmt.Errorf("create parent dir for %s: %w", targetPath, err)
	}
	targetFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fileModeOrDefault(mode, 0o644))
	if err != nil {
		return fmt.Errorf("create %s: %w", targetPath, err)
	}
	defer targetFile.Close()
	if _, err := io.Copy(targetFile, sourceFile); err != nil {
		return fmt.Errorf("copy %s to %s: %w", sourcePath, targetPath, err)
	}
	return nil
}

func fileModeOrDefault(mode fs.FileMode, fallback fs.FileMode) fs.FileMode {
	if mode.Perm() == 0 {
		return fallback
	}
	return mode.Perm()
}

func appendUniqueString(values []string, value string) []string {
	for _, existing := range values {
		if sameFilePath(existing, value) {
			return values
		}
	}
	return append(values, value)
}

func isPathWithinBase(path string, base string) bool {
	relativePath, err := filepath.Rel(base, path)
	if err != nil {
		return false
	}
	if relativePath == "." {
		return true
	}
	relativePath = filepath.Clean(relativePath)
	return relativePath != ".." && !strings.HasPrefix(relativePath, ".."+string(os.PathSeparator))
}

func sameFilePath(left string, right string) bool {
	left = filepath.Clean(left)
	right = filepath.Clean(right)
	if os.PathSeparator == '\\' {
		return strings.EqualFold(left, right)
	}
	return left == right
}

func removeInstalledPackagePath(installPath string, pkg InstalledPackage) (string, error) {
	info, err := os.Stat(installPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "missing", nil
		}
		return "", fmt.Errorf("stat install path %s: %w", installPath, err)
	}
	if !info.IsDir() {
		return "skipped", nil
	}
	manifest, err := readNPMPackageArchiveManifest(installPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "skipped", nil
		}
		return "", fmt.Errorf("read installed package at %s: %w", installPath, err)
	}
	if !strings.EqualFold(strings.TrimSpace(manifest.Name), strings.TrimSpace(pkg.Name)) || strings.TrimSpace(manifest.Version) != strings.TrimSpace(pkg.Version) {
		return "skipped", nil
	}
	if err := os.RemoveAll(installPath); err != nil {
		return "", fmt.Errorf("remove install path %s: %w", installPath, err)
	}
	return "removed", nil
}

func removePathIfPresent(target string) (string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return "missing", nil
	}
	if _, err := os.Stat(target); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "missing", nil
		}
		return "", fmt.Errorf("stat path %s: %w", target, err)
	}
	if err := os.RemoveAll(target); err != nil {
		return "", fmt.Errorf("remove path %s: %w", target, err)
	}
	return "removed", nil
}

func removeDirectoryIfEmpty(target string) (string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return "missing", nil
	}
	entries, err := os.ReadDir(target)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "missing", nil
		}
		return "", fmt.Errorf("read dir %s: %w", target, err)
	}
	if len(entries) > 0 {
		return "skipped", nil
	}
	if err := os.Remove(target); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "missing", nil
		}
		return "", fmt.Errorf("remove dir %s: %w", target, err)
	}
	return "removed", nil
}

func cleanupEmptyParentDirectories(start string, stop string, result *UninstallResult, onRemoved func(string)) {
	current := filepath.Clean(start)
	stop = filepath.Clean(stop)
	for {
		if sameFilePath(current, stop) || current == "." || current == string(os.PathSeparator) {
			return
		}
		status, err := removeDirectoryIfEmpty(current)
		if err != nil || status != "removed" {
			return
		}
		onRemoved(current)
		parent := filepath.Dir(current)
		if parent == current {
			return
		}
		current = parent
	}
}

func ensureInstalledPackagePath(pkg InstalledPackage, installPath string, baseDir string) (string, error) {
	status, err := inspectInstalledPackagePath(installPath, pkg)
	if err != nil {
		return "", err
	}
	switch status {
	case "healthy":
		repairedBinLinks, err := ensureInstalledPackageBinLinks(installPath, baseDir)
		if err != nil {
			return "", err
		}
		if repairedBinLinks {
			return "repaired", nil
		}
		return "healthy", nil
	case "conflict":
		return "skipped", nil
	}
	storeStatus, err := ensureStoredPackage(pkg, baseDir)
	if err != nil {
		return "", err
	}
	if storeStatus == "missing_cache" {
		return "missing_cache", nil
	}
	if storeStatus == "invalid_cache_path" {
		return "skipped", nil
	}
	if err := copyDirectory(pkg.StorePath, installPath); err != nil {
		return "", fmt.Errorf("repair %s at %s: %w", installPackageKey(pkg.Name, pkg.Version), installPath, err)
	}
	if _, _, err := createNPMPackageBinLinksFromPath(installPath); err != nil {
		return "", err
	}
	return "repaired", nil
}

func inspectInstalledPackagePath(installPath string, pkg InstalledPackage) (string, error) {
	info, err := os.Stat(installPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "repair", nil
		}
		return "", fmt.Errorf("stat install path %s: %w", installPath, err)
	}
	if !info.IsDir() {
		return "repair", nil
	}
	manifest, err := readNPMPackageArchiveManifest(installPath)
	if err != nil {
		// Missing or malformed package.json means the path is damaged and can be rebuilt.
		return "repair", nil
	}
	if !strings.EqualFold(strings.TrimSpace(manifest.Name), strings.TrimSpace(pkg.Name)) || strings.TrimSpace(manifest.Version) != strings.TrimSpace(pkg.Version) {
		return "conflict", nil
	}
	return "healthy", nil
}

func ensureStoredPackage(pkg InstalledPackage, baseDir string) (string, error) {
	if !isManagedManifestPath(pkg.StorePath, baseDir) {
		return "invalid_cache_path", nil
	}
	if strings.TrimSpace(pkg.ArchivePath) != "" && !isManagedManifestPath(pkg.ArchivePath, baseDir) {
		return "invalid_cache_path", nil
	}
	manifest, err := readNPMPackageArchiveManifest(pkg.StorePath)
	if err == nil {
		if strings.EqualFold(strings.TrimSpace(manifest.Name), strings.TrimSpace(pkg.Name)) && strings.TrimSpace(manifest.Version) == strings.TrimSpace(pkg.Version) {
			return "healthy", nil
		}
	}
	if strings.TrimSpace(pkg.ArchivePath) == "" {
		return "missing_cache", nil
	}
	if _, err := os.Stat(pkg.ArchivePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "missing_cache", nil
		}
		return "", fmt.Errorf("stat archive cache %s: %w", pkg.ArchivePath, err)
	}
	if err := materializeNPMPackage(pkg.ArchivePath, pkg.StorePath, pkg.Name, pkg.Version); err != nil {
		return "", fmt.Errorf("restore store cache for %s@%s: %w", pkg.Name, pkg.Version, err)
	}
	return "restored", nil
}

func ensureInstalledPackageBinLinks(installPath string, baseDir string) (bool, error) {
	missing, created, err := createNPMPackageBinLinksFromPath(installPath)
	if err != nil {
		return false, err
	}
	for _, link := range created {
		if !isManagedManifestPath(link, baseDir) {
			return false, fmt.Errorf("bin link escapes base dir for %s: %s", installPath, link)
		}
	}
	return missing && len(created) > 0, nil
}

func createNPMPackageBinLinksFromPath(installPath string) (bool, []string, error) {
	manifest, err := readNPMPackageArchiveManifest(installPath)
	if err != nil {
		return false, nil, fmt.Errorf("read installed package manifest for %s: %w", installPath, err)
	}
	missing, err := npmPackageBinLinksMissing(installPath, manifest)
	if err != nil {
		return false, nil, err
	}
	created, _, err := createNPMPackageBinLinks(installPath, manifest)
	if err != nil {
		return false, nil, err
	}
	return missing, created, nil
}

func npmPackageBinLinksMissing(installPath string, manifest npmPackageArchiveManifest) (bool, error) {
	if len(manifest.Bin) == 0 {
		return false, nil
	}
	nodeModulesDir := findNearestNodeModulesDir(installPath)
	if nodeModulesDir == "" {
		return false, fmt.Errorf("unable to locate node_modules for %s", installPath)
	}
	binDir := filepath.Join(nodeModulesDir, ".bin")
	for _, binName := range sortedMapKeys(manifest.Bin) {
		for _, expected := range expectedNodeBinShimPaths(binDir, normalizeBinName(binName, manifest.Name)) {
			if _, err := os.Stat(expected); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return true, nil
				}
				return false, fmt.Errorf("stat bin shim %s: %w", expected, err)
			}
		}
	}
	return false, nil
}

func isManagedManifestPath(target string, baseDir string) bool {
	target = strings.TrimSpace(target)
	if target == "" {
		return false
	}
	target = filepath.Clean(target)
	baseDir = filepath.Clean(baseDir)
	return isPathWithinBase(target, baseDir) || sameFilePath(target, baseDir)
}
