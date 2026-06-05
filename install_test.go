package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha1"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	neturl "net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestPickNPMVersionSupportsRanges(t *testing.T) {
	metadata := npmPackageMetadata{
		Name: "demo",
		Versions: map[string]npmVersionDetails{
			"1.2.0": {},
			"1.3.0": {},
			"2.0.0": {},
		},
	}

	version, _, err := pickNPMVersion(metadata, "^1.2.0")
	if err != nil {
		t.Fatalf("pickNPMVersion returned error: %v", err)
	}
	if version != "1.3.0" {
		t.Fatalf("expected ^1.2.0 to resolve to 1.3.0, got %q", version)
	}

	version, _, err = pickNPMVersion(metadata, ">=1.2.0 <2.0.0")
	if err != nil {
		t.Fatalf("pickNPMVersion returned error: %v", err)
	}
	if version != "1.3.0" {
		t.Fatalf("expected comparator range to resolve to 1.3.0, got %q", version)
	}
}

func TestResolveInstallPlanRecursivelyResolvesNPMPackages(t *testing.T) {
	server := newNPMRegistryTestServer(t, defaultNPMRegistryFixture())
	defer server.Close()

	plan, err := resolveInstallPlan(context.Background(), server.Client(), InstallRequest{
		Source:    "npm",
		Name:      "root",
		Version:   "latest",
		OutputDir: t.TempDir(),
		Mirror:    server.URL,
	})
	if err != nil {
		t.Fatalf("resolveInstallPlan returned error: %v", err)
	}
	if plan.Root != "root" || plan.RootVersion != "1.0.0" {
		t.Fatalf("unexpected root package: %+v", plan)
	}
	if len(plan.Packages) != 3 {
		t.Fatalf("expected 3 unique packages, got %d", len(plan.Packages))
	}

	versions := map[string]string{}
	for _, pkg := range plan.Packages {
		versions[pkg.Name] = pkg.Version
	}
	if versions["dep-a"] != "1.2.0" {
		t.Fatalf("expected dep-a to resolve to 1.2.0, got %q", versions["dep-a"])
	}
	if versions["dep-b"] != "2.0.0" {
		t.Fatalf("expected dep-b to resolve to 2.0.0, got %q", versions["dep-b"])
	}
}

func TestExecuteInstallPlanBuildsNodeModulesTree(t *testing.T) {
	server := newNPMRegistryTestServer(t, defaultNPMRegistryFixture())
	defer server.Close()

	outputDir := t.TempDir()
	plan, err := resolveInstallPlan(context.Background(), server.Client(), InstallRequest{
		Source:    "npm",
		Name:      "root",
		Version:   "latest",
		OutputDir: outputDir,
		Mirror:    server.URL,
	})
	if err != nil {
		t.Fatalf("resolveInstallPlan returned error: %v", err)
	}

	result, err := executeInstallPlan(context.Background(), server.Client(), plan, DownloadOptions{
		OutputDir: outputDir,
	})
	if err != nil {
		t.Fatalf("executeInstallPlan returned error: %v", err)
	}

	rootPath := filepath.Join(outputDir, "node_modules", "root")
	depAPath := filepath.Join(rootPath, "node_modules", "dep-a")
	depBPath := filepath.Join(rootPath, "node_modules", "dep-b")
	if result.RootPath != rootPath {
		t.Fatalf("expected root path %q, got %q", rootPath, result.RootPath)
	}
	if result.NodeModulesDir != filepath.Join(outputDir, "node_modules") {
		t.Fatalf("unexpected node_modules dir: %q", result.NodeModulesDir)
	}
	for _, path := range []string{
		filepath.Join(rootPath, "package.json"),
		filepath.Join(depAPath, "package.json"),
		filepath.Join(depBPath, "package.json"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected installed file at %s: %v", path, err)
		}
	}
	if _, err := os.Stat(filepath.Join(depBPath, "node_modules", "dep-a")); !os.IsNotExist(err) {
		t.Fatalf("expected dep-a to be reused from ancestor, stat err=%v", err)
	}
}

func TestExecuteInstallPlanWritesManifestWithTopology(t *testing.T) {
	server := newNPMRegistryTestServer(t, defaultNPMRegistryFixture())
	defer server.Close()

	outputDir := t.TempDir()
	plan, err := resolveInstallPlan(context.Background(), server.Client(), InstallRequest{
		Source:    "npm",
		Name:      "root",
		Version:   "latest",
		OutputDir: outputDir,
		Mirror:    server.URL,
	})
	if err != nil {
		t.Fatalf("resolveInstallPlan returned error: %v", err)
	}

	result, err := executeInstallPlan(context.Background(), server.Client(), plan, DownloadOptions{OutputDir: outputDir})
	if err != nil {
		t.Fatalf("executeInstallPlan returned error: %v", err)
	}
	if _, err := os.Stat(result.ManifestPath); err != nil {
		t.Fatalf("expected manifest to exist: %v", err)
	}

	raw, err := os.ReadFile(result.ManifestPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var manifest InstallManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatalf("parse manifest: %v", err)
	}
	if manifest.RootPath != filepath.Join(outputDir, "node_modules", "root") {
		t.Fatalf("unexpected manifest root path: %q", manifest.RootPath)
	}
	if manifest.NodeModulesDir != filepath.Join(outputDir, "node_modules") {
		t.Fatalf("unexpected manifest node_modules dir: %q", manifest.NodeModulesDir)
	}
	rootPackage := findInstalledPackage(t, manifest.Packages, "root", "1.0.0")
	if rootPackage.ResolvedDependencies["dep-a"] != "1.2.0" || rootPackage.ResolvedDependencies["dep-b"] != "2.0.0" {
		t.Fatalf("unexpected resolved dependencies: %+v", rootPackage.ResolvedDependencies)
	}
	depA := findInstalledPackage(t, manifest.Packages, "dep-a", "1.2.0")
	if len(depA.InstallPaths) != 1 || depA.InstallPaths[0] != filepath.Join(outputDir, "node_modules", "root", "node_modules", "dep-a") {
		t.Fatalf("unexpected dep-a install paths: %+v", depA.InstallPaths)
	}
}

func TestExecuteInstallPlanSupportsScopedPackages(t *testing.T) {
	server := newNPMRegistryTestServer(t, scopedNPMRegistryFixture())
	defer server.Close()

	outputDir := t.TempDir()
	plan, err := resolveInstallPlan(context.Background(), server.Client(), InstallRequest{
		Source:    "npm",
		Name:      "@scope/root",
		Version:   "latest",
		OutputDir: outputDir,
		Mirror:    server.URL,
	})
	if err != nil {
		t.Fatalf("resolveInstallPlan returned error: %v", err)
	}

	if _, err := executeInstallPlan(context.Background(), server.Client(), plan, DownloadOptions{OutputDir: outputDir}); err != nil {
		t.Fatalf("executeInstallPlan returned error: %v", err)
	}

	rootPath := filepath.Join(outputDir, "node_modules", "@scope", "root")
	dependencyPath := filepath.Join(rootPath, "node_modules", "@scope", "dep")
	for _, path := range []string{
		filepath.Join(rootPath, "package.json"),
		filepath.Join(dependencyPath, "package.json"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected scoped package content at %s: %v", path, err)
		}
	}
}

func TestResolveInstallPlanSupportsNPMAliasAndSkipsIncompatibleOptionalDeps(t *testing.T) {
	server := newNPMRegistryTestServer(t, aliasNPMRegistryFixture())
	defer server.Close()

	plan, err := resolveInstallPlan(context.Background(), server.Client(), InstallRequest{
		Source:    "npm",
		Name:      "alias-root",
		Version:   "latest",
		OutputDir: t.TempDir(),
		Mirror:    server.URL,
	})
	if err != nil {
		t.Fatalf("resolveInstallPlan returned error: %v", err)
	}

	root := findInstallPackage(t, plan.Packages, "alias-root", "1.0.0")
	if root.ResolvedDependencies["alias-dep"] != "2.1.0" {
		t.Fatalf("expected alias dependency to resolve through dist-tag, got %+v", root.ResolvedDependencies)
	}
	if root.ResolvedDependencies["platform-match"] != "1.0.0-win32-x64" {
		t.Fatalf("expected current-platform optional dependency to resolve, got %+v", root.ResolvedDependencies)
	}
	if _, ok := root.ResolvedDependencies["platform-mismatch"]; ok {
		t.Fatalf("did not expect mismatched-platform optional dependency, got %+v", root.ResolvedDependencies)
	}
	if _, ok := root.ResolvedDependencies["optional-missing"]; ok {
		t.Fatalf("did not expect missing optional dependency, got %+v", root.ResolvedDependencies)
	}
	findInstallPackage(t, plan.Packages, "alias-dep", "2.1.0")
	findInstallPackage(t, plan.Packages, "platform-match", "1.0.0-win32-x64")
}

func TestExecuteInstallPlanPlacesAliasDependencyUnderAliasName(t *testing.T) {
	server := newNPMRegistryTestServer(t, aliasNPMRegistryFixture())
	defer server.Close()

	outputDir := t.TempDir()
	plan, err := resolveInstallPlan(context.Background(), server.Client(), InstallRequest{
		Source:    "npm",
		Name:      "alias-root",
		Version:   "latest",
		OutputDir: outputDir,
		Mirror:    server.URL,
	})
	if err != nil {
		t.Fatalf("resolveInstallPlan returned error: %v", err)
	}
	if _, err := executeInstallPlan(context.Background(), server.Client(), plan, DownloadOptions{OutputDir: outputDir}); err != nil {
		t.Fatalf("executeInstallPlan returned error: %v", err)
	}

	aliasPath := filepath.Join(outputDir, "node_modules", "alias-root", "node_modules", "alias-dep", "package.json")
	if _, err := os.Stat(aliasPath); err != nil {
		t.Fatalf("expected alias dependency to be installed under alias name: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "node_modules", "alias-root", "node_modules", "platform-match", "package.json")); err != nil {
		t.Fatalf("expected matching optional platform dependency to be installed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "node_modules", "alias-root", "node_modules", "platform-mismatch")); !os.IsNotExist(err) {
		t.Fatalf("expected mismatched platform dependency to be skipped, stat err=%v", err)
	}
}

func TestExecuteUninstallRemovesInstalledTreeAndManifest(t *testing.T) {
	server := newNPMRegistryTestServer(t, defaultNPMRegistryFixture())
	defer server.Close()

	outputDir, manifestPath := installFixtureForTest(t, server.URL, "root")
	result, err := executeUninstall(UninstallRequest{OutputDir: outputDir})
	if err != nil {
		t.Fatalf("executeUninstall returned error: %v", err)
	}
	if result.ManifestPath != manifestPath {
		t.Fatalf("expected manifest path %q, got %q", manifestPath, result.ManifestPath)
	}
	if len(result.MissingPaths) != 0 {
		t.Fatalf("expected clean uninstall to have no missing paths, got %+v", result.MissingPaths)
	}
	if len(result.SkippedPaths) != 0 {
		t.Fatalf("expected clean uninstall to have no skipped paths, got %+v", result.SkippedPaths)
	}
	for _, path := range []string{
		filepath.Join(outputDir, "node_modules"),
		filepath.Join(outputDir, ".source-fetcher"),
		manifestPath,
	} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("expected %s to be removed, stat err=%v", path, err)
		}
	}
}

func TestExecuteUninstallToleratesMissingInstalledPath(t *testing.T) {
	server := newNPMRegistryTestServer(t, defaultNPMRegistryFixture())
	defer server.Close()

	outputDir, _ := installFixtureForTest(t, server.URL, "root")
	missingPath := filepath.Join(outputDir, "node_modules", "root", "node_modules", "dep-a")
	if err := os.RemoveAll(missingPath); err != nil {
		t.Fatalf("remove dep-a before uninstall: %v", err)
	}

	result, err := executeUninstall(UninstallRequest{OutputDir: outputDir})
	if err != nil {
		t.Fatalf("executeUninstall returned error: %v", err)
	}
	if !containsPath(result.MissingPaths, missingPath) {
		t.Fatalf("expected missing paths to include %q, got %+v", missingPath, result.MissingPaths)
	}
}

func TestExecuteUninstallSupportsScopedPackages(t *testing.T) {
	server := newNPMRegistryTestServer(t, scopedNPMRegistryFixture())
	defer server.Close()

	outputDir, _ := installFixtureForTest(t, server.URL, "@scope/root")
	if _, err := executeUninstall(UninstallRequest{OutputDir: outputDir}); err != nil {
		t.Fatalf("executeUninstall returned error: %v", err)
	}

	for _, path := range []string{
		filepath.Join(outputDir, "node_modules", "@scope"),
		filepath.Join(outputDir, "node_modules"),
	} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("expected scoped uninstall to remove %s, stat err=%v", path, err)
		}
	}
}

func TestExecuteRepairRestoresMissingInstallPath(t *testing.T) {
	server := newNPMRegistryTestServer(t, defaultNPMRegistryFixture())
	defer server.Close()

	outputDir, manifestPath := installFixtureForTest(t, server.URL, "root")
	manifest := mustLoadInstallManifest(t, manifestPath)
	depA := findInstalledPackage(t, manifest.Packages, "dep-a", "1.2.0")
	repairPath := depA.InstallPaths[0]
	if err := os.RemoveAll(repairPath); err != nil {
		t.Fatalf("remove install path before repair: %v", err)
	}

	result, err := executeRepair(RepairRequest{OutputDir: outputDir})
	if err != nil {
		t.Fatalf("executeRepair returned error: %v", err)
	}
	if !containsPath(result.RepairedPaths, repairPath) {
		t.Fatalf("expected repaired paths to include %q, got %+v", repairPath, result.RepairedPaths)
	}
	if _, err := os.Stat(filepath.Join(repairPath, "package.json")); err != nil {
		t.Fatalf("expected repair to restore package.json: %v", err)
	}
}

func TestExecuteRepairRestoresStoreFromArchive(t *testing.T) {
	server := newNPMRegistryTestServer(t, defaultNPMRegistryFixture())
	defer server.Close()

	outputDir, manifestPath := installFixtureForTest(t, server.URL, "root")
	manifest := mustLoadInstallManifest(t, manifestPath)
	depA := findInstalledPackage(t, manifest.Packages, "dep-a", "1.2.0")
	repairPath := depA.InstallPaths[0]
	if err := os.RemoveAll(depA.StorePath); err != nil {
		t.Fatalf("remove store path before repair: %v", err)
	}
	if err := os.RemoveAll(repairPath); err != nil {
		t.Fatalf("remove install path before repair: %v", err)
	}

	result, err := executeRepair(RepairRequest{OutputDir: outputDir})
	if err != nil {
		t.Fatalf("executeRepair returned error: %v", err)
	}
	if !containsPath(result.RepairedPaths, repairPath) {
		t.Fatalf("expected repaired paths to include %q, got %+v", repairPath, result.RepairedPaths)
	}
	if _, err := os.Stat(filepath.Join(depA.StorePath, "package.json")); err != nil {
		t.Fatalf("expected repair to restore store cache: %v", err)
	}
}

func TestExecuteRepairSkipsConflictingInstallPath(t *testing.T) {
	server := newNPMRegistryTestServer(t, defaultNPMRegistryFixture())
	defer server.Close()

	outputDir, manifestPath := installFixtureForTest(t, server.URL, "root")
	manifest := mustLoadInstallManifest(t, manifestPath)
	depA := findInstalledPackage(t, manifest.Packages, "dep-a", "1.2.0")
	repairPath := depA.InstallPaths[0]
	if err := os.WriteFile(filepath.Join(repairPath, "package.json"), []byte(`{"name":"other","version":"9.9.9"}`), 0o644); err != nil {
		t.Fatalf("overwrite package.json before repair: %v", err)
	}

	result, err := executeRepair(RepairRequest{OutputDir: outputDir})
	if err != nil {
		t.Fatalf("executeRepair returned error: %v", err)
	}
	if !containsPath(result.SkippedPaths, repairPath) {
		t.Fatalf("expected skipped paths to include %q, got %+v", repairPath, result.SkippedPaths)
	}
	raw, err := os.ReadFile(filepath.Join(repairPath, "package.json"))
	if err != nil {
		t.Fatalf("read package.json after repair: %v", err)
	}
	if !strings.Contains(string(raw), `"name":"other"`) {
		t.Fatalf("expected conflicting package.json to remain untouched, got %q", string(raw))
	}
}

func TestExecuteRepairReportsMissingCache(t *testing.T) {
	server := newNPMRegistryTestServer(t, defaultNPMRegistryFixture())
	defer server.Close()

	outputDir, manifestPath := installFixtureForTest(t, server.URL, "root")
	manifest := mustLoadInstallManifest(t, manifestPath)
	depA := findInstalledPackage(t, manifest.Packages, "dep-a", "1.2.0")
	repairPath := depA.InstallPaths[0]
	if err := os.RemoveAll(depA.StorePath); err != nil {
		t.Fatalf("remove store path before repair: %v", err)
	}
	if err := os.Remove(depA.ArchivePath); err != nil {
		t.Fatalf("remove archive path before repair: %v", err)
	}
	if err := os.RemoveAll(repairPath); err != nil {
		t.Fatalf("remove install path before repair: %v", err)
	}

	result, err := executeRepair(RepairRequest{OutputDir: outputDir})
	if err != nil {
		t.Fatalf("executeRepair returned error: %v", err)
	}
	if !containsPath(result.MissingCachePaths, repairPath) {
		t.Fatalf("expected missing cache paths to include %q, got %+v", repairPath, result.MissingCachePaths)
	}
}

func TestExecuteInstallPlanReusesVerifiedCachedTarballs(t *testing.T) {
	var tarballHits int
	server := newNPMRegistryTestServerWithObserver(t, defaultNPMRegistryFixture(), func(r *http.Request) {
		if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/tarballs/") {
			tarballHits++
		}
	})
	defer server.Close()

	outputDir := t.TempDir()
	plan, err := resolveInstallPlan(context.Background(), server.Client(), InstallRequest{
		Source:    "npm",
		Name:      "root",
		Version:   "latest",
		OutputDir: outputDir,
		Mirror:    server.URL,
	})
	if err != nil {
		t.Fatalf("resolveInstallPlan returned error: %v", err)
	}
	if _, err := executeInstallPlan(context.Background(), server.Client(), plan, DownloadOptions{OutputDir: outputDir}); err != nil {
		t.Fatalf("first executeInstallPlan returned error: %v", err)
	}
	firstHits := tarballHits
	if firstHits != len(plan.Packages) {
		t.Fatalf("expected %d tarball downloads on first install, got %d", len(plan.Packages), firstHits)
	}
	if _, err := executeInstallPlan(context.Background(), server.Client(), plan, DownloadOptions{OutputDir: outputDir}); err != nil {
		t.Fatalf("second executeInstallPlan returned error: %v", err)
	}
	if tarballHits != firstHits {
		t.Fatalf("expected second install to reuse cached tarballs, hits before=%d after=%d", firstHits, tarballHits)
	}
}

func TestExecuteInstallPlanRejectsIntegrityMismatch(t *testing.T) {
	server := newNPMRegistryTestServer(t, defaultNPMRegistryFixture())
	defer server.Close()

	outputDir := t.TempDir()
	plan, err := resolveInstallPlan(context.Background(), server.Client(), InstallRequest{
		Source:    "npm",
		Name:      "root",
		Version:   "latest",
		OutputDir: outputDir,
		Mirror:    server.URL,
	})
	if err != nil {
		t.Fatalf("resolveInstallPlan returned error: %v", err)
	}
	plan.Packages[0].Integrity = "sha512-" + base64.StdEncoding.EncodeToString([]byte("bad-digest"))

	_, err = executeInstallPlan(context.Background(), server.Client(), plan, DownloadOptions{OutputDir: outputDir})
	if err == nil || !strings.Contains(err.Error(), "archive integrity mismatch") {
		t.Fatalf("expected integrity mismatch error, got %v", err)
	}
}

func TestExecuteInstallPlanLeavesNoCommittedInstallOnFailure(t *testing.T) {
	server := newNPMRegistryTestServer(t, defaultNPMRegistryFixture())
	defer server.Close()

	outputDir := t.TempDir()
	plan, err := resolveInstallPlan(context.Background(), server.Client(), InstallRequest{
		Source:    "npm",
		Name:      "root",
		Version:   "latest",
		OutputDir: outputDir,
		Mirror:    server.URL,
	})
	if err != nil {
		t.Fatalf("resolveInstallPlan returned error: %v", err)
	}
	plan.Packages[0].Integrity = "sha512-" + base64.StdEncoding.EncodeToString([]byte("bad-digest"))

	_, err = executeInstallPlan(context.Background(), server.Client(), plan, DownloadOptions{OutputDir: outputDir})
	if err == nil || !strings.Contains(err.Error(), "archive integrity mismatch") {
		t.Fatalf("expected integrity mismatch error, got %v", err)
	}

	if _, err := os.Stat(filepath.Join(outputDir, "node_modules")); !os.IsNotExist(err) {
		t.Fatalf("expected node_modules to remain absent after failed install, stat err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "source-fetcher-install.json")); !os.IsNotExist(err) {
		t.Fatalf("expected manifest to remain absent after failed install, stat err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "source-fetcher-install.lock.json")); !os.IsNotExist(err) {
		t.Fatalf("expected lockfile to remain absent after failed install, stat err=%v", err)
	}
}

func TestExecuteInstallPlanPreservesPreviousInstallOnFailure(t *testing.T) {
	server := newNPMRegistryTestServer(t, defaultNPMRegistryFixture())
	defer server.Close()

	outputDir := t.TempDir()
	plan, err := resolveInstallPlan(context.Background(), server.Client(), InstallRequest{
		Source:    "npm",
		Name:      "root",
		Version:   "latest",
		OutputDir: outputDir,
		Mirror:    server.URL,
	})
	if err != nil {
		t.Fatalf("resolveInstallPlan returned error: %v", err)
	}
	result, err := executeInstallPlan(context.Background(), server.Client(), plan, DownloadOptions{OutputDir: outputDir})
	if err != nil {
		t.Fatalf("initial executeInstallPlan returned error: %v", err)
	}
	manifestBefore, err := os.ReadFile(result.ManifestPath)
	if err != nil {
		t.Fatalf("read manifest before failed reinstall: %v", err)
	}
	lockfileBefore, err := os.ReadFile(result.LockfilePath)
	if err != nil {
		t.Fatalf("read lockfile before failed reinstall: %v", err)
	}

	plan.Packages[0].Integrity = "sha512-" + base64.StdEncoding.EncodeToString([]byte("bad-digest"))
	_, err = executeInstallPlan(context.Background(), server.Client(), plan, DownloadOptions{OutputDir: outputDir})
	if err == nil || !strings.Contains(err.Error(), "archive integrity mismatch") {
		t.Fatalf("expected integrity mismatch error on reinstall, got %v", err)
	}

	rootPackageJSON := filepath.Join(outputDir, "node_modules", "root", "package.json")
	if _, err := os.Stat(rootPackageJSON); err != nil {
		t.Fatalf("expected previous install to remain available: %v", err)
	}
	manifestAfter, err := os.ReadFile(result.ManifestPath)
	if err != nil {
		t.Fatalf("read manifest after failed reinstall: %v", err)
	}
	if !bytes.Equal(manifestBefore, manifestAfter) {
		t.Fatal("expected manifest to remain unchanged after failed reinstall")
	}
	lockfileAfter, err := os.ReadFile(result.LockfilePath)
	if err != nil {
		t.Fatalf("read lockfile after failed reinstall: %v", err)
	}
	if !bytes.Equal(lockfileBefore, lockfileAfter) {
		t.Fatal("expected lockfile to remain unchanged after failed reinstall")
	}
}

func TestResolveInstallPlanSupportsInstallStrategyFlags(t *testing.T) {
	server := newNPMRegistryTestServer(t, featureNPMRegistryFixture())
	defer server.Close()

	baseReq := InstallRequest{
		Source:    "npm",
		Name:      "feature-root",
		Version:   "latest",
		OutputDir: t.TempDir(),
		Mirror:    server.URL,
	}
	plan, err := resolveInstallPlan(context.Background(), server.Client(), baseReq)
	if err != nil {
		t.Fatalf("resolveInstallPlan returned error: %v", err)
	}
	root := findInstallPackage(t, plan.Packages, "feature-root", "1.0.0")
	if _, ok := root.ResolvedDependencies["feature-opt"]; !ok {
		t.Fatalf("expected optional dependency to be included by default, got %+v", root.ResolvedDependencies)
	}
	if _, ok := root.ResolvedDependencies["feature-peer"]; ok {
		t.Fatalf("did not expect peer dependency without flag, got %+v", root.ResolvedDependencies)
	}
	if _, ok := root.ResolvedDependencies["feature-dev"]; ok {
		t.Fatalf("did not expect dev dependency without flag, got %+v", root.ResolvedDependencies)
	}

	plan, err = resolveInstallPlan(context.Background(), server.Client(), InstallRequest{
		Source:       "npm",
		Name:         "feature-root",
		Version:      "latest",
		OutputDir:    t.TempDir(),
		Mirror:       server.URL,
		OmitOptional: true,
		IncludePeer:  true,
		IncludeDev:   true,
	})
	if err != nil {
		t.Fatalf("resolveInstallPlan with strategy flags returned error: %v", err)
	}
	root = findInstallPackage(t, plan.Packages, "feature-root", "1.0.0")
	if _, ok := root.ResolvedDependencies["feature-opt"]; ok {
		t.Fatalf("did not expect optional dependency when omitted, got %+v", root.ResolvedDependencies)
	}
	if _, ok := root.ResolvedDependencies["feature-peer"]; !ok {
		t.Fatalf("expected peer dependency with flag, got %+v", root.ResolvedDependencies)
	}
	if _, ok := root.ResolvedDependencies["feature-dev"]; !ok {
		t.Fatalf("expected dev dependency with flag, got %+v", root.ResolvedDependencies)
	}
}

func TestExecuteInstallPlanWritesAndUsesLockfile(t *testing.T) {
	var requests int
	server := newNPMRegistryTestServerWithObserver(t, defaultNPMRegistryFixture(), func(r *http.Request) {
		requests++
	})
	defer server.Close()

	outputDir := t.TempDir()
	plan, err := resolveInstallPlan(context.Background(), server.Client(), InstallRequest{
		Source:    "npm",
		Name:      "root",
		Version:   "latest",
		OutputDir: outputDir,
		Mirror:    server.URL,
	})
	if err != nil {
		t.Fatalf("resolveInstallPlan returned error: %v", err)
	}
	result, err := executeInstallPlan(context.Background(), server.Client(), plan, DownloadOptions{OutputDir: outputDir})
	if err != nil {
		t.Fatalf("executeInstallPlan returned error: %v", err)
	}
	if result.LockfilePath == "" {
		t.Fatal("expected install to write a lockfile path")
	}
	if _, err := os.Stat(result.LockfilePath); err != nil {
		t.Fatalf("expected lockfile to exist: %v", err)
	}

	requests = 0
	lockedPlan, err := resolveInstallPlan(context.Background(), server.Client(), InstallRequest{
		Source:         "npm",
		Name:           "root",
		Version:        "latest",
		OutputDir:      outputDir,
		Mirror:         server.URL,
		FrozenLockfile: true,
	})
	if err != nil {
		t.Fatalf("resolveInstallPlan with frozen lockfile returned error: %v", err)
	}
	if lockedPlan.LockfilePath != result.LockfilePath {
		t.Fatalf("expected lockfile path %q, got %q", result.LockfilePath, lockedPlan.LockfilePath)
	}
	if requests != 0 {
		t.Fatalf("expected frozen lockfile resolve to avoid registry requests, got %d", requests)
	}
}

func TestExecuteInstallPlanLockfileRecordsPerPackageMirror(t *testing.T) {
	primaryFixture := npmRegistryFixture{
		Packages: map[string]npmRegistryFixturePackage{
			"root": defaultNPMRegistryFixture().Packages["root"],
		},
	}
	primary := newNPMRegistryTestServer(t, primaryFixture)
	defer primary.Close()

	secondary := newNPMRegistryTestServer(t, defaultNPMRegistryFixture())
	defer secondary.Close()

	original := builtinMirrors["npm"]
	builtinMirrors["npm"] = []Mirror{
		{Source: "npm", Name: "primary", BaseURL: primary.URL},
		{Source: "npm", Name: "secondary", BaseURL: secondary.URL},
	}
	defer func() { builtinMirrors["npm"] = original }()

	outputDir := t.TempDir()
	plan, err := resolveInstallPlan(context.Background(), &http.Client{}, InstallRequest{
		Source:    "npm",
		Name:      "root",
		Version:   "latest",
		OutputDir: outputDir,
	})
	if err != nil {
		t.Fatalf("resolveInstallPlan returned error: %v", err)
	}
	result, err := executeInstallPlan(context.Background(), &http.Client{}, plan, DownloadOptions{OutputDir: outputDir})
	if err != nil {
		t.Fatalf("executeInstallPlan returned error: %v", err)
	}

	lockfile, err := loadInstallLockfile(result.LockfilePath)
	if err != nil {
		t.Fatalf("loadInstallLockfile returned error: %v", err)
	}
	rootPkg := findInstallPackage(t, lockfile.Packages, "root", "1.0.0")
	if rootPkg.MirrorName != "primary" {
		t.Fatalf("expected root package mirror to be primary, got %+v", rootPkg)
	}
	depPkg := findInstallPackage(t, lockfile.Packages, "dep-a", "1.2.0")
	if depPkg.MirrorName != "secondary" {
		t.Fatalf("expected dep-a mirror to be secondary, got %+v", depPkg)
	}
	if lockfile.MirrorName != "primary" {
		t.Fatalf("expected root lockfile mirror to remain primary, got %+v", lockfile)
	}
}

func TestResolveInstallPlanFrozenLockfileSupportsLegacyPackageMirrorShape(t *testing.T) {
	var requests int
	server := newNPMRegistryTestServerWithObserver(t, defaultNPMRegistryFixture(), func(r *http.Request) {
		requests++
	})
	defer server.Close()

	outputDir := t.TempDir()
	plan, err := resolveInstallPlan(context.Background(), server.Client(), InstallRequest{
		Source:    "npm",
		Name:      "root",
		Version:   "latest",
		OutputDir: outputDir,
		Mirror:    server.URL,
	})
	if err != nil {
		t.Fatalf("resolveInstallPlan returned error: %v", err)
	}
	result, err := executeInstallPlan(context.Background(), server.Client(), plan, DownloadOptions{OutputDir: outputDir})
	if err != nil {
		t.Fatalf("executeInstallPlan returned error: %v", err)
	}

	raw, err := os.ReadFile(result.LockfilePath)
	if err != nil {
		t.Fatalf("read lockfile: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("parse lockfile json: %v", err)
	}
	packages, ok := payload["packages"].([]any)
	if !ok {
		t.Fatalf("expected packages array, got %+v", payload["packages"])
	}
	for _, item := range packages {
		pkg, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("expected package object, got %+v", item)
		}
		delete(pkg, "mirror")
	}
	raw, err = json.MarshalIndent(payload, "", "  ")
	if err != nil {
		t.Fatalf("marshal legacy lockfile json: %v", err)
	}
	if err := os.WriteFile(result.LockfilePath, raw, 0o644); err != nil {
		t.Fatalf("rewrite legacy lockfile: %v", err)
	}

	requests = 0
	lockedPlan, err := resolveInstallPlan(context.Background(), server.Client(), InstallRequest{
		Source:         "npm",
		Name:           "root",
		Version:        "latest",
		OutputDir:      outputDir,
		Mirror:         server.URL,
		FrozenLockfile: true,
	})
	if err != nil {
		t.Fatalf("resolveInstallPlan with legacy frozen lockfile returned error: %v", err)
	}
	if requests != 0 {
		t.Fatalf("expected legacy frozen lockfile resolve to avoid registry requests, got %d", requests)
	}
	rootPkg := findInstallPackage(t, lockedPlan.Packages, "root", "1.0.0")
	if rootPkg.MirrorName != "" {
		t.Fatalf("expected legacy package mirror to remain empty, got %+v", rootPkg)
	}
}

func TestResolveInstallPlanFrozenLockfileRejectsMirrorMismatch(t *testing.T) {
	primary := newNPMRegistryTestServer(t, defaultNPMRegistryFixture())
	defer primary.Close()
	secondary := newNPMRegistryTestServer(t, defaultNPMRegistryFixture())
	defer secondary.Close()

	original := builtinMirrors["npm"]
	builtinMirrors["npm"] = []Mirror{
		{Source: "npm", Name: "primary", BaseURL: primary.URL},
		{Source: "npm", Name: "secondary", BaseURL: secondary.URL},
	}
	defer func() { builtinMirrors["npm"] = original }()

	outputDir := t.TempDir()
	plan, err := resolveInstallPlan(context.Background(), &http.Client{}, InstallRequest{
		Source:    "npm",
		Name:      "root",
		Version:   "latest",
		OutputDir: outputDir,
		Mirror:    "primary",
	})
	if err != nil {
		t.Fatalf("resolveInstallPlan returned error: %v", err)
	}
	if _, err := executeInstallPlan(context.Background(), &http.Client{}, plan, DownloadOptions{OutputDir: outputDir}); err != nil {
		t.Fatalf("executeInstallPlan returned error: %v", err)
	}

	_, err = resolveInstallPlan(context.Background(), &http.Client{}, InstallRequest{
		Source:         "npm",
		Name:           "root",
		Version:        "latest",
		OutputDir:      outputDir,
		Mirror:         "secondary",
		FrozenLockfile: true,
	})
	if err == nil || !strings.Contains(err.Error(), "does not match install request") {
		t.Fatalf("expected frozen lockfile mirror mismatch error, got %v", err)
	}
}

func TestResolveInstallPlanFrozenLockfileRejectsAllowScriptsMismatch(t *testing.T) {
	server := newNPMRegistryTestServer(t, featureNPMRegistryFixture())
	defer server.Close()

	outputDir := t.TempDir()
	plan, err := resolveInstallPlan(context.Background(), server.Client(), InstallRequest{
		Source:        "npm",
		Name:          "feature-root",
		Version:       "latest",
		OutputDir:     outputDir,
		Mirror:        server.URL,
		ScriptsPolicy: "all",
		AllowScripts:  []string{"feature-opt"},
	})
	if err != nil {
		t.Fatalf("resolveInstallPlan returned error: %v", err)
	}
	if _, err := executeInstallPlan(context.Background(), server.Client(), plan, DownloadOptions{OutputDir: outputDir}); err != nil {
		t.Fatalf("executeInstallPlan returned error: %v", err)
	}

	_, err = resolveInstallPlan(context.Background(), server.Client(), InstallRequest{
		Source:         "npm",
		Name:           "feature-root",
		Version:        "latest",
		OutputDir:      outputDir,
		Mirror:         server.URL,
		FrozenLockfile: true,
		ScriptsPolicy:  "all",
	})
	if err == nil || !strings.Contains(err.Error(), "does not match install request") {
		t.Fatalf("expected frozen lockfile allow-scripts mismatch error, got %v", err)
	}
}

func TestResolveInstallPlanVerboseLogsLockfileMismatch(t *testing.T) {
	server := newNPMRegistryTestServer(t, featureNPMRegistryFixture())
	defer server.Close()

	outputDir := t.TempDir()
	plan, err := resolveInstallPlan(context.Background(), server.Client(), InstallRequest{
		Source:        "npm",
		Name:          "feature-root",
		Version:       "latest",
		OutputDir:     outputDir,
		Mirror:        server.URL,
		ScriptsPolicy: "all",
		AllowScripts:  []string{"feature-opt"},
	})
	if err != nil {
		t.Fatalf("resolveInstallPlan returned error: %v", err)
	}
	if _, err := executeInstallPlan(context.Background(), server.Client(), plan, DownloadOptions{OutputDir: outputDir}); err != nil {
		t.Fatalf("executeInstallPlan returned error: %v", err)
	}

	var log bytes.Buffer
	_, err = resolveInstallPlan(context.Background(), server.Client(), InstallRequest{
		Source:         "npm",
		Name:           "feature-root",
		Version:        "latest",
		OutputDir:      outputDir,
		Mirror:         server.URL,
		FrozenLockfile: true,
		ScriptsPolicy:  "all",
		RequestOptions: RequestOptions{
			Verbose: newVerboseLogger(true, &log),
		},
	})
	if err == nil {
		t.Fatal("expected frozen lockfile mismatch error")
	}
	if !strings.Contains(log.String(), "[lockfile] ignored") || !strings.Contains(log.String(), "allow_scripts differ") {
		t.Fatalf("expected verbose lockfile mismatch log, got %q", log.String())
	}
}

func TestExecuteInstallPlanCreatesBinLinks(t *testing.T) {
	server := newNPMRegistryTestServer(t, featureNPMRegistryFixture())
	defer server.Close()

	outputDir := t.TempDir()
	plan, err := resolveInstallPlan(context.Background(), server.Client(), InstallRequest{
		Source:    "npm",
		Name:      "feature-root",
		Version:   "latest",
		OutputDir: outputDir,
		Mirror:    server.URL,
	})
	if err != nil {
		t.Fatalf("resolveInstallPlan returned error: %v", err)
	}
	result, err := executeInstallPlan(context.Background(), server.Client(), plan, DownloadOptions{OutputDir: outputDir})
	if err != nil {
		t.Fatalf("executeInstallPlan returned error: %v", err)
	}
	rootPkg := findInstalledPackage(t, result.Packages, "feature-root", "1.0.0")
	if len(rootPkg.BinLinks) == 0 {
		t.Fatalf("expected root package to record bin links, got %+v", rootPkg)
	}
	if _, err := os.Stat(expectedBinShimPath(filepath.Join(outputDir, "node_modules", ".bin"), "feature-root")); err != nil {
		t.Fatalf("expected root bin shim to exist: %v", err)
	}
}

func TestExecuteInstallPlanRunsRootScriptsWhenEnabled(t *testing.T) {
	server := newNPMRegistryTestServer(t, featureNPMRegistryFixture())
	defer server.Close()

	outputDir := t.TempDir()
	plan, err := resolveInstallPlan(context.Background(), server.Client(), InstallRequest{
		Source:        "npm",
		Name:          "feature-root",
		Version:       "latest",
		OutputDir:     outputDir,
		Mirror:        server.URL,
		ScriptsPolicy: "root",
	})
	if err != nil {
		t.Fatalf("resolveInstallPlan returned error: %v", err)
	}
	if _, err := executeInstallPlan(context.Background(), server.Client(), plan, DownloadOptions{OutputDir: outputDir}); err != nil {
		t.Fatalf("executeInstallPlan returned error: %v", err)
	}
	rootMarker := filepath.Join(outputDir, "node_modules", "feature-root", "root-script.txt")
	if _, err := os.Stat(rootMarker); err != nil {
		t.Fatalf("expected root install script marker: %v", err)
	}
	rootPostMarker := filepath.Join(outputDir, "node_modules", "feature-root", "root-postinstall.txt")
	if _, err := os.Stat(rootPostMarker); err != nil {
		t.Fatalf("expected root postinstall script marker: %v", err)
	}
	rootPreMarker := filepath.Join(outputDir, "node_modules", "feature-root", "root-preinstall.txt")
	if _, err := os.Stat(rootPreMarker); !os.IsNotExist(err) {
		t.Fatalf("expected root preinstall script to be blocked, stat err=%v", err)
	}
	depMarker := filepath.Join(outputDir, "node_modules", "feature-root", "node_modules", "feature-opt", "dep-script.txt")
	if _, err := os.Stat(depMarker); !os.IsNotExist(err) {
		t.Fatalf("expected dependency script to remain disabled under root policy, stat err=%v", err)
	}
}

func TestExecuteInstallPlanBlocksPreinstallButRunsPostinstallUnderAllPolicy(t *testing.T) {
	server := newNPMRegistryTestServer(t, featureNPMRegistryFixture())
	defer server.Close()

	outputDir := t.TempDir()
	plan, err := resolveInstallPlan(context.Background(), server.Client(), InstallRequest{
		Source:        "npm",
		Name:          "feature-root",
		Version:       "latest",
		OutputDir:     outputDir,
		Mirror:        server.URL,
		ScriptsPolicy: "all",
	})
	if err != nil {
		t.Fatalf("resolveInstallPlan returned error: %v", err)
	}

	warnings := captureStderr(t, func() {
		if _, err := executeInstallPlan(context.Background(), server.Client(), plan, DownloadOptions{OutputDir: outputDir}); err != nil {
			t.Fatalf("executeInstallPlan returned error: %v", err)
		}
	})
	if !strings.Contains(warnings, "blocked preinstall script for feature-root") {
		t.Fatalf("expected root preinstall warning, got %q", warnings)
	}
	if !strings.Contains(warnings, "blocked preinstall script for feature-opt") {
		t.Fatalf("expected dependency preinstall warning, got %q", warnings)
	}
	if !strings.Contains(warnings, "--allow-scripts=feature-root") {
		t.Fatalf("expected override hint in warning, got %q", warnings)
	}

	rootDir := filepath.Join(outputDir, "node_modules", "feature-root")
	if _, err := os.Stat(filepath.Join(rootDir, "root-script.txt")); err != nil {
		t.Fatalf("expected root install script marker: %v", err)
	}
	if _, err := os.Stat(filepath.Join(rootDir, "root-postinstall.txt")); err != nil {
		t.Fatalf("expected root postinstall marker: %v", err)
	}
	if _, err := os.Stat(filepath.Join(rootDir, "root-preinstall.txt")); !os.IsNotExist(err) {
		t.Fatalf("expected root preinstall marker to remain absent, stat err=%v", err)
	}

	depDir := filepath.Join(rootDir, "node_modules", "feature-opt")
	if _, err := os.Stat(filepath.Join(depDir, "dep-script.txt")); err != nil {
		t.Fatalf("expected dependency install script marker: %v", err)
	}
	if _, err := os.Stat(filepath.Join(depDir, "dep-postinstall.txt")); err != nil {
		t.Fatalf("expected dependency postinstall marker: %v", err)
	}
	if _, err := os.Stat(filepath.Join(depDir, "dep-preinstall.txt")); !os.IsNotExist(err) {
		t.Fatalf("expected dependency preinstall marker to remain absent, stat err=%v", err)
	}
}

func TestExecuteInstallPlanAllowsAuthorizedPreinstallUnderAllPolicy(t *testing.T) {
	server := newNPMRegistryTestServer(t, featureNPMRegistryFixture())
	defer server.Close()

	outputDir := t.TempDir()
	plan, err := resolveInstallPlan(context.Background(), server.Client(), InstallRequest{
		Source:        "npm",
		Name:          "feature-root",
		Version:       "latest",
		OutputDir:     outputDir,
		Mirror:        server.URL,
		ScriptsPolicy: "all",
		AllowScripts:  []string{"feature-opt"},
	})
	if err != nil {
		t.Fatalf("resolveInstallPlan returned error: %v", err)
	}

	warnings := captureStderr(t, func() {
		if _, err := executeInstallPlan(context.Background(), server.Client(), plan, DownloadOptions{OutputDir: outputDir}); err != nil {
			t.Fatalf("executeInstallPlan returned error: %v", err)
		}
	})
	if !strings.Contains(warnings, "blocked preinstall script for feature-root") {
		t.Fatalf("expected root preinstall warning, got %q", warnings)
	}
	if strings.Contains(warnings, "blocked preinstall script for feature-opt") {
		t.Fatalf("did not expect dependency preinstall warning for allowed package, got %q", warnings)
	}

	rootDir := filepath.Join(outputDir, "node_modules", "feature-root")
	if _, err := os.Stat(filepath.Join(rootDir, "root-preinstall.txt")); !os.IsNotExist(err) {
		t.Fatalf("expected root preinstall marker to remain absent, stat err=%v", err)
	}
	depDir := filepath.Join(rootDir, "node_modules", "feature-opt")
	if _, err := os.Stat(filepath.Join(depDir, "dep-preinstall.txt")); err != nil {
		t.Fatalf("expected dependency preinstall marker for allowed package: %v", err)
	}
}

func TestExecuteInstallPlanVerboseLogsScriptDecisions(t *testing.T) {
	server := newNPMRegistryTestServer(t, featureNPMRegistryFixture())
	defer server.Close()

	outputDir := t.TempDir()
	plan, err := resolveInstallPlan(context.Background(), server.Client(), InstallRequest{
		Source:        "npm",
		Name:          "feature-root",
		Version:       "latest",
		OutputDir:     outputDir,
		Mirror:        server.URL,
		ScriptsPolicy: "all",
		AllowScripts:  []string{"feature-opt"},
	})
	if err != nil {
		t.Fatalf("resolveInstallPlan returned error: %v", err)
	}

	var log bytes.Buffer
	if _, err := executeInstallPlan(context.Background(), server.Client(), plan, DownloadOptions{
		OutputDir: outputDir,
		RequestOptions: RequestOptions{
			Verbose: newVerboseLogger(true, &log),
		},
	}); err != nil {
		t.Fatalf("executeInstallPlan returned error: %v", err)
	}
	text := log.String()
	if !strings.Contains(text, "[scripts] blocked preinstall for feature-root") {
		t.Fatalf("expected blocked script verbose log, got %q", text)
	}
	if !strings.Contains(text, "[scripts] running preinstall for feature-opt") {
		t.Fatalf("expected allowed preinstall verbose log, got %q", text)
	}
	if !strings.Contains(text, "[scripts] running postinstall for feature-root") {
		t.Fatalf("expected postinstall verbose log, got %q", text)
	}
}

func TestExecuteUninstallSkipsCachePathsOutsideManifestBase(t *testing.T) {
	server := newNPMRegistryTestServer(t, defaultNPMRegistryFixture())
	defer server.Close()

	outputDir, manifestPath := installFixtureForTest(t, server.URL, "root")
	manifest := mustLoadInstallManifest(t, manifestPath)
	depAIndex := findInstalledPackageIndex(t, manifest.Packages, "dep-a", "1.2.0")
	externalDir := filepath.Join(t.TempDir(), "outside-store")
	externalFile := filepath.Join(t.TempDir(), "outside.tgz")
	if err := os.MkdirAll(externalDir, 0o755); err != nil {
		t.Fatalf("mkdir external dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(externalDir, "package.json"), []byte(`{"name":"dep-a","version":"1.2.0"}`), 0o644); err != nil {
		t.Fatalf("write external store manifest: %v", err)
	}
	if err := os.WriteFile(externalFile, []byte("archive"), 0o644); err != nil {
		t.Fatalf("write external archive: %v", err)
	}
	manifest.Packages[depAIndex].StorePath = externalDir
	manifest.Packages[depAIndex].ArchivePath = externalFile
	writeInstallManifestForTest(t, manifestPath, manifest)

	result, err := executeUninstall(UninstallRequest{OutputDir: outputDir})
	if err != nil {
		t.Fatalf("executeUninstall returned error: %v", err)
	}
	if !containsPath(result.SkippedPaths, externalDir) || !containsPath(result.SkippedPaths, externalFile) {
		t.Fatalf("expected skipped paths to include external cache paths, got %+v", result.SkippedPaths)
	}
	if _, err := os.Stat(externalDir); err != nil {
		t.Fatalf("expected external dir to remain untouched: %v", err)
	}
	if _, err := os.Stat(externalFile); err != nil {
		t.Fatalf("expected external archive to remain untouched: %v", err)
	}
}

func TestExecuteRepairSkipsCachePathsOutsideManifestBase(t *testing.T) {
	server := newNPMRegistryTestServer(t, defaultNPMRegistryFixture())
	defer server.Close()

	outputDir, manifestPath := installFixtureForTest(t, server.URL, "root")
	manifest := mustLoadInstallManifest(t, manifestPath)
	depAIndex := findInstalledPackageIndex(t, manifest.Packages, "dep-a", "1.2.0")
	repairPath := manifest.Packages[depAIndex].InstallPaths[0]
	externalDir := filepath.Join(t.TempDir(), "outside-store")
	externalFile := filepath.Join(t.TempDir(), "outside.tgz")
	if err := os.MkdirAll(externalDir, 0o755); err != nil {
		t.Fatalf("mkdir external dir: %v", err)
	}
	if err := os.WriteFile(externalFile, []byte("archive"), 0o644); err != nil {
		t.Fatalf("write external archive: %v", err)
	}
	manifest.Packages[depAIndex].StorePath = externalDir
	manifest.Packages[depAIndex].ArchivePath = externalFile
	writeInstallManifestForTest(t, manifestPath, manifest)
	if err := os.RemoveAll(repairPath); err != nil {
		t.Fatalf("remove install path before repair: %v", err)
	}

	result, err := executeRepair(RepairRequest{OutputDir: outputDir})
	if err != nil {
		t.Fatalf("executeRepair returned error: %v", err)
	}
	if !containsPath(result.SkippedPaths, repairPath) {
		t.Fatalf("expected skipped paths to include %q, got %+v", repairPath, result.SkippedPaths)
	}
	if _, err := os.Stat(externalDir); err != nil {
		t.Fatalf("expected external dir to remain untouched: %v", err)
	}
	if _, err := os.Stat(externalFile); err != nil {
		t.Fatalf("expected external archive to remain untouched: %v", err)
	}
}

func TestMaterializeNPMPackageRejectsArchiveSymlink(t *testing.T) {
	archivePath := filepath.Join(t.TempDir(), "evil.tgz")
	if err := os.WriteFile(archivePath, mustBuildNPMPackageArchiveWithSymlink(t), 0o644); err != nil {
		t.Fatalf("write archive: %v", err)
	}

	storePath := filepath.Join(t.TempDir(), "store")
	err := materializeNPMPackage(archivePath, storePath, "evil", "1.0.0")
	if err == nil || !strings.Contains(err.Error(), "archive symlinks are not supported") {
		t.Fatalf("expected symlink rejection error, got %v", err)
	}
	if _, err := os.Stat(storePath); !os.IsNotExist(err) {
		t.Fatalf("expected failed materialize not to leave store path behind, stat err=%v", err)
	}
}

func TestMaterializeNPMPackagePreservesExistingStoreOnFailure(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "store")
	goodArchivePath := filepath.Join(t.TempDir(), "dep-a-good.tgz")
	goodArchive := mustBuildNPMPackageArchive(t, "dep-a", "1.2.0", npmRegistryFixtureVersion{}, map[string]string{
		"index.js": `module.exports = "dep-a@1.2.0";`,
	})
	if err := os.WriteFile(goodArchivePath, goodArchive, 0o644); err != nil {
		t.Fatalf("write good archive: %v", err)
	}
	if err := materializeNPMPackage(goodArchivePath, storePath, "dep-a", "1.2.0"); err != nil {
		t.Fatalf("materialize good archive: %v", err)
	}
	packageJSONPath := filepath.Join(storePath, "package.json")
	before, err := os.ReadFile(packageJSONPath)
	if err != nil {
		t.Fatalf("read package.json before failed rematerialize: %v", err)
	}

	badArchivePath := filepath.Join(t.TempDir(), "dep-a-bad.tgz")
	if err := os.WriteFile(badArchivePath, mustBuildNPMPackageArchiveWithSymlink(t), 0o644); err != nil {
		t.Fatalf("write bad archive: %v", err)
	}
	err = materializeNPMPackage(badArchivePath, storePath, "dep-a", "1.0.0")
	if err == nil || !strings.Contains(err.Error(), "archive symlinks are not supported") {
		t.Fatalf("expected symlink rejection error, got %v", err)
	}
	after, err := os.ReadFile(packageJSONPath)
	if err != nil {
		t.Fatalf("read package.json after failed rematerialize: %v", err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("expected existing store contents to remain unchanged after failed rematerialize")
	}
}

func TestRunInstallHelpReturnsNil(t *testing.T) {
	if err := run([]string{"install", "--help"}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
}

func TestRunUninstallHelpReturnsNil(t *testing.T) {
	if err := run([]string{"uninstall", "--help"}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
}

func TestRunRepairHelpReturnsNil(t *testing.T) {
	if err := run([]string{"repair", "--help"}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
}

func TestPrintInstallPlanIncludesPackageCount(t *testing.T) {
	var builder strings.Builder
	printInstallPlan(&builder, InstallPlan{
		Source:       "npm",
		Root:         "root",
		RootVersion:  "1.0.0",
		AllowScripts: []string{"feature-opt", "feature-root"},
		Packages: []InstallPackage{
			{Name: "root", Version: "1.0.0"},
			{Name: "dep-a", Version: "1.2.0"},
		},
	})
	text := builder.String()
	if !strings.Contains(text, "Packages: 2") || !strings.Contains(text, "root@1.0.0") || !strings.Contains(text, "Allow Scripts: feature-opt, feature-root") {
		t.Fatalf("unexpected install plan output: %q", text)
	}
}

func TestPrintInstallResultIncludesManifest(t *testing.T) {
	var builder strings.Builder
	printInstallResult(&builder, InstallPlan{Source: "npm", Root: "root", RootVersion: "1.0.0"}, InstallResult{
		ManifestPath:   `D:\tmp\source-fetcher-install.json`,
		RootPath:       `D:\tmp\node_modules\root`,
		NodeModulesDir: `D:\tmp\node_modules`,
		Packages: []InstalledPackage{
			{Name: "root", Version: "1.0.0"},
		},
		Duration: time.Second,
	})
	text := builder.String()
	if !strings.Contains(text, "Installed Packages: 1") ||
		!strings.Contains(text, "Manifest: D:\\tmp\\source-fetcher-install.json") ||
		!strings.Contains(text, "Root Path: D:\\tmp\\node_modules\\root") {
		t.Fatalf("unexpected install result output: %q", text)
	}
}

func TestPrintUninstallResultIncludesCounts(t *testing.T) {
	var builder strings.Builder
	printUninstallResult(&builder, UninstallResult{
		ManifestPath: `D:\tmp\source-fetcher-install.json`,
		RemovedPaths: []string{`D:\tmp\node_modules\root`},
		MissingPaths: []string{`D:\tmp\node_modules\root\node_modules\dep-a`},
		Duration:     time.Second,
	})
	text := builder.String()
	if !strings.Contains(text, "Removed Paths: 1") ||
		!strings.Contains(text, "Missing Paths: 1") ||
		!strings.Contains(text, "Manifest: D:\\tmp\\source-fetcher-install.json") {
		t.Fatalf("unexpected uninstall result output: %q", text)
	}
}

func TestPrintRepairResultIncludesCounts(t *testing.T) {
	var builder strings.Builder
	printRepairResult(&builder, RepairResult{
		ManifestPath:      `D:\tmp\source-fetcher-install.json`,
		RepairedPaths:     []string{`D:\tmp\node_modules\root\node_modules\dep-a`},
		HealthyPaths:      []string{`D:\tmp\node_modules\root`},
		MissingCachePaths: []string{`D:\tmp\node_modules\root\node_modules\dep-b`},
		Duration:          time.Second,
	})
	text := builder.String()
	if !strings.Contains(text, "Repaired Paths: 1") ||
		!strings.Contains(text, "Healthy Paths: 1") ||
		!strings.Contains(text, "Missing Cache Paths: 1") {
		t.Fatalf("unexpected repair result output: %q", text)
	}
}

func TestResolveInstallPlanRejectsUnsupportedSource(t *testing.T) {
	_, err := resolveInstallPlan(context.Background(), &http.Client{}, InstallRequest{Source: "pip", Name: "demo"})
	if err == nil || !strings.Contains(err.Error(), "install currently supports") {
		t.Fatalf("expected unsupported source error, got %v", err)
	}
}

func TestMatchesNPMWildcardRange(t *testing.T) {
	if !matchesNPMVersionRange("2.3.4", "2.x") {
		t.Fatal("expected 2.3.4 to satisfy 2.x")
	}
	if matchesNPMVersionRange("3.0.0", "2.x") {
		t.Fatal("did not expect 3.0.0 to satisfy 2.x")
	}
}

type npmRegistryFixture struct {
	Packages map[string]npmRegistryFixturePackage
}

type npmRegistryFixturePackage struct {
	DistTags map[string]string
	Versions map[string]npmRegistryFixtureVersion
}

type npmRegistryFixtureVersion struct {
	Dependencies         map[string]string
	OptionalDependencies map[string]string
	PeerDependencies     map[string]string
	DevDependencies      map[string]string
	OS                   []string
	CPU                  []string
	Bin                  map[string]string
	Scripts              map[string]string
	Files                map[string]string
}

func defaultNPMRegistryFixture() npmRegistryFixture {
	return npmRegistryFixture{
		Packages: map[string]npmRegistryFixturePackage{
			"root": {
				DistTags: map[string]string{"latest": "1.0.0"},
				Versions: map[string]npmRegistryFixtureVersion{
					"1.0.0": {
						Dependencies: map[string]string{
							"dep-a": "^1.1.0",
							"dep-b": "2.x",
						},
						Files: map[string]string{
							"index.js": `module.exports = "root";`,
						},
					},
				},
			},
			"dep-a": {
				Versions: map[string]npmRegistryFixtureVersion{
					"1.1.0": {
						Files: map[string]string{
							"index.js": `module.exports = "dep-a@1.1.0";`,
						},
					},
					"1.2.0": {
						Files: map[string]string{
							"index.js": `module.exports = "dep-a@1.2.0";`,
						},
					},
				},
			},
			"dep-b": {
				Versions: map[string]npmRegistryFixtureVersion{
					"2.0.0": {
						Dependencies: map[string]string{
							"dep-a": "^1.0.0",
						},
						Files: map[string]string{
							"index.js": `module.exports = require("dep-a");`,
						},
					},
				},
			},
		},
	}
}

func scopedNPMRegistryFixture() npmRegistryFixture {
	return npmRegistryFixture{
		Packages: map[string]npmRegistryFixturePackage{
			"@scope/root": {
				DistTags: map[string]string{"latest": "1.0.0"},
				Versions: map[string]npmRegistryFixtureVersion{
					"1.0.0": {
						Dependencies: map[string]string{
							"@scope/dep": "^1.0.0",
						},
						Files: map[string]string{
							"index.js": `module.exports = "@scope/root";`,
						},
					},
				},
			},
			"@scope/dep": {
				DistTags: map[string]string{"latest": "1.0.0"},
				Versions: map[string]npmRegistryFixtureVersion{
					"1.0.0": {
						Files: map[string]string{
							"lib/main.js": `module.exports = "@scope/dep";`,
						},
					},
				},
			},
		},
	}
}

func featureNPMRegistryFixture() npmRegistryFixture {
	return npmRegistryFixture{
		Packages: map[string]npmRegistryFixturePackage{
			"feature-root": {
				DistTags: map[string]string{"latest": "1.0.0"},
				Versions: map[string]npmRegistryFixtureVersion{
					"1.0.0": {
						Dependencies:         map[string]string{"feature-base": "^1.0.0"},
						OptionalDependencies: map[string]string{"feature-opt": "^1.0.0"},
						PeerDependencies:     map[string]string{"feature-peer": "^1.0.0"},
						DevDependencies:      map[string]string{"feature-dev": "^1.0.0"},
						Bin:                  map[string]string{"": "bin/feature-root.js"},
						Scripts: map[string]string{
							"preinstall":  `Set-Content -Path root-preinstall.txt -Value blocked`,
							"install":     `Set-Content -Path root-script.txt -Value ok`,
							"postinstall": `Set-Content -Path root-postinstall.txt -Value ok`,
						},
						Files: map[string]string{
							"index.js":            `module.exports = "feature-root";`,
							"bin/feature-root.js": `console.log("feature-root");`,
						},
					},
				},
			},
			"feature-base": {
				DistTags: map[string]string{"latest": "1.0.0"},
				Versions: map[string]npmRegistryFixtureVersion{
					"1.0.0": {
						Files: map[string]string{"index.js": `module.exports = "feature-base";`},
					},
				},
			},
			"feature-opt": {
				DistTags: map[string]string{"latest": "1.0.0"},
				Versions: map[string]npmRegistryFixtureVersion{
					"1.0.0": {
						Scripts: map[string]string{
							"preinstall":  `Set-Content -Path dep-preinstall.txt -Value blocked`,
							"install":     `Set-Content -Path dep-script.txt -Value dep`,
							"postinstall": `Set-Content -Path dep-postinstall.txt -Value dep`,
						},
						Files: map[string]string{"index.js": `module.exports = "feature-opt";`},
					},
				},
			},
			"feature-peer": {
				DistTags: map[string]string{"latest": "1.0.0"},
				Versions: map[string]npmRegistryFixtureVersion{
					"1.0.0": {Files: map[string]string{"index.js": `module.exports = "feature-peer";`}},
				},
			},
			"feature-dev": {
				DistTags: map[string]string{"latest": "1.0.0"},
				Versions: map[string]npmRegistryFixtureVersion{
					"1.0.0": {Files: map[string]string{"index.js": `module.exports = "feature-dev";`}},
				},
			},
		},
	}
}

func aliasNPMRegistryFixture() npmRegistryFixture {
	return npmRegistryFixture{
		Packages: map[string]npmRegistryFixturePackage{
			"alias-root": {
				DistTags: map[string]string{"latest": "1.0.0"},
				Versions: map[string]npmRegistryFixtureVersion{
					"1.0.0": {
						Dependencies: map[string]string{
							"alias-dep": "npm:alias-source@stable",
						},
						OptionalDependencies: map[string]string{
							"platform-match":    "npm:platform-source@win32-x64",
							"platform-mismatch": "npm:platform-source@darwin-arm64",
							"optional-missing":  "^1.0.0",
						},
						Files: map[string]string{
							"index.js": `module.exports = "alias-root";`,
						},
					},
				},
			},
			"alias-source": {
				DistTags: map[string]string{"stable": "2.1.0"},
				Versions: map[string]npmRegistryFixtureVersion{
					"2.1.0": {
						Files: map[string]string{
							"index.js": `module.exports = "alias-source";`,
						},
					},
				},
			},
			"platform-source": {
				DistTags: map[string]string{
					"win32-x64":    "1.0.0-win32-x64",
					"darwin-arm64": "1.0.0-darwin-arm64",
				},
				Versions: map[string]npmRegistryFixtureVersion{
					"1.0.0-win32-x64": {
						OS:  []string{"win32"},
						CPU: []string{"x64"},
						Files: map[string]string{
							"index.js": `module.exports = "platform-win";`,
						},
					},
					"1.0.0-darwin-arm64": {
						OS:  []string{"darwin"},
						CPU: []string{"arm64"},
						Files: map[string]string{
							"index.js": `module.exports = "platform-darwin";`,
						},
					},
				},
			},
		},
	}
}

func newNPMRegistryTestServer(t *testing.T, fixture npmRegistryFixture) *httptest.Server {
	return newNPMRegistryTestServerWithObserver(t, fixture, nil)
}

func newNPMRegistryTestServerWithObserver(t *testing.T, fixture npmRegistryFixture, onRequest func(*http.Request)) *httptest.Server {
	t.Helper()

	tarballs := make(map[string][]byte)
	shasums := make(map[string]string)
	integrities := make(map[string]string)
	for packageName, pkg := range fixture.Packages {
		for version, details := range pkg.Versions {
			tarballPath := fixtureTarballPath(packageName, version)
			tarball := mustBuildNPMPackageArchive(t, packageName, version, details, details.Files)
			tarballs[tarballPath] = tarball
			shasums[tarballPath] = sha1Hex(tarball)
			integrities[tarballPath] = sha512Integrity(tarball)
		}
	}

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if onRequest != nil {
			onRequest(r)
		}
		if content, ok := tarballs[r.URL.Path]; ok {
			w.Header().Set("Content-Type", "application/octet-stream")
			_, _ = w.Write(content)
			return
		}

		name, err := neturl.PathUnescape(strings.TrimPrefix(r.URL.EscapedPath(), "/"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		pkg, ok := fixture.Packages[name]
		if !ok {
			http.NotFound(w, r)
			return
		}

		metadata := npmPackageMetadata{
			Name:     name,
			DistTags: cloneStringMap(pkg.DistTags),
			Versions: map[string]npmVersionDetails{},
		}
		for version, details := range pkg.Versions {
			metadata.Versions[version] = npmVersionDetails{
				Version: version,
				Dist: struct {
					Tarball   string `json:"tarball"`
					Shasum    string `json:"shasum"`
					Integrity string `json:"integrity"`
				}{
					Tarball:   server.URL + fixtureTarballPath(name, version),
					Shasum:    shasums[fixtureTarballPath(name, version)],
					Integrity: integrities[fixtureTarballPath(name, version)],
				},
				Dependencies:         cloneStringMap(details.Dependencies),
				OptionalDependencies: cloneStringMap(details.OptionalDependencies),
				PeerDependencies:     cloneStringMap(details.PeerDependencies),
				DevDependencies:      cloneStringMap(details.DevDependencies),
				OS:                   append(npmStringList(nil), details.OS...),
				CPU:                  append(npmStringList(nil), details.CPU...),
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(metadata)
	}))
	return server
}

func fixtureTarballPath(name string, version string) string {
	return "/tarballs/" + sanitizeFileName(strings.ReplaceAll(name, "/", "_")+"-"+version+".tgz")
}

func mustBuildNPMPackageArchive(t *testing.T, name string, version string, details npmRegistryFixtureVersion, files map[string]string) []byte {
	t.Helper()

	manifest := map[string]any{
		"name":    name,
		"version": version,
	}
	if len(details.Dependencies) > 0 {
		manifest["dependencies"] = details.Dependencies
	}
	if len(details.OptionalDependencies) > 0 {
		manifest["optionalDependencies"] = details.OptionalDependencies
	}
	if len(details.PeerDependencies) > 0 {
		manifest["peerDependencies"] = details.PeerDependencies
	}
	if len(details.DevDependencies) > 0 {
		manifest["devDependencies"] = details.DevDependencies
	}
	if len(details.Bin) == 1 {
		if value, ok := details.Bin[""]; ok {
			manifest["bin"] = value
		} else {
			manifest["bin"] = details.Bin
		}
	} else if len(details.Bin) > 0 {
		manifest["bin"] = details.Bin
	}
	if len(details.Scripts) > 0 {
		manifest["scripts"] = details.Scripts
	}
	manifestJSON, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshal package manifest: %v", err)
	}

	entries := map[string]string{
		"package/package.json": string(manifestJSON),
	}
	for name, content := range files {
		entries["package/"+strings.TrimPrefix(filepath.ToSlash(name), "/")] = content
	}

	var compressed bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressed)
	tarWriter := tar.NewWriter(gzipWriter)
	for entryName, content := range entries {
		header := &tar.Header{
			Name: entryName,
			Mode: 0o644,
			Size: int64(len(content)),
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			t.Fatalf("write tar header %s: %v", entryName, err)
		}
		if _, err := tarWriter.Write([]byte(content)); err != nil {
			t.Fatalf("write tar content %s: %v", entryName, err)
		}
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatalf("close tar writer: %v", err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatalf("close gzip writer: %v", err)
	}
	return compressed.Bytes()
}

func mustBuildNPMPackageArchiveWithSymlink(t *testing.T) []byte {
	t.Helper()

	var compressed bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressed)
	tarWriter := tar.NewWriter(gzipWriter)
	entries := []struct {
		header  *tar.Header
		content string
	}{
		{
			header: &tar.Header{
				Name: "package/package.json",
				Mode: 0o644,
				Size: int64(len(`{"name":"evil","version":"1.0.0"}`)),
			},
			content: `{"name":"evil","version":"1.0.0"}`,
		},
		{
			header: &tar.Header{
				Name:     "package/link",
				Typeflag: tar.TypeSymlink,
				Linkname: "../outside",
				Mode:     0o777,
			},
		},
	}

	for _, entry := range entries {
		if err := tarWriter.WriteHeader(entry.header); err != nil {
			t.Fatalf("write tar header %s: %v", entry.header.Name, err)
		}
		if entry.content != "" {
			if _, err := tarWriter.Write([]byte(entry.content)); err != nil {
				t.Fatalf("write tar content %s: %v", entry.header.Name, err)
			}
		}
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatalf("close tar writer: %v", err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatalf("close gzip writer: %v", err)
	}
	return compressed.Bytes()
}

func sha1Hex(value []byte) string {
	sum := sha1.Sum(value)
	return hex.EncodeToString(sum[:])
}

func sha512Integrity(value []byte) string {
	sum := sha512.Sum512(value)
	return "sha512-" + base64.StdEncoding.EncodeToString(sum[:])
}

func findInstalledPackage(t *testing.T, packages []InstalledPackage, name string, version string) InstalledPackage {
	t.Helper()
	for _, pkg := range packages {
		if pkg.Name == name && pkg.Version == version {
			return pkg
		}
	}
	t.Fatalf("package %s@%s not found in manifest", name, version)
	return InstalledPackage{}
}

func findInstallPackage(t *testing.T, packages []InstallPackage, name string, version string) InstallPackage {
	t.Helper()
	for _, pkg := range packages {
		if pkg.Name == name && pkg.Version == version {
			return pkg
		}
	}
	t.Fatalf("package %s@%s not found in plan", name, version)
	return InstallPackage{}
}

func findInstalledPackageIndex(t *testing.T, packages []InstalledPackage, name string, version string) int {
	t.Helper()
	for index, pkg := range packages {
		if pkg.Name == name && pkg.Version == version {
			return index
		}
	}
	t.Fatalf("package %s@%s not found in manifest", name, version)
	return -1
}

func installFixtureForTest(t *testing.T, mirrorURL string, packageName string) (string, string) {
	t.Helper()

	outputDir := t.TempDir()
	plan, err := resolveInstallPlan(context.Background(), &http.Client{}, InstallRequest{
		Source:    "npm",
		Name:      packageName,
		Version:   "latest",
		OutputDir: outputDir,
		Mirror:    mirrorURL,
	})
	if err != nil {
		t.Fatalf("resolveInstallPlan returned error: %v", err)
	}
	client := &http.Client{}
	if strings.TrimSpace(mirrorURL) != "" {
		client = newHTTPClient(30 * time.Second)
	}
	result, err := executeInstallPlan(context.Background(), client, plan, DownloadOptions{OutputDir: outputDir})
	if err != nil {
		t.Fatalf("executeInstallPlan returned error: %v", err)
	}
	return outputDir, result.ManifestPath
}

func containsPath(paths []string, expected string) bool {
	for _, path := range paths {
		if sameFilePath(path, expected) {
			return true
		}
	}
	return false
}

func mustLoadInstallManifest(t *testing.T, manifestPath string) InstallManifest {
	t.Helper()
	manifest, err := loadInstallManifest(manifestPath)
	if err != nil {
		t.Fatalf("load install manifest: %v", err)
	}
	return manifest
}

func writeInstallManifestForTest(t *testing.T, manifestPath string, manifest InstallManifest) {
	t.Helper()
	raw, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	if err := os.WriteFile(manifestPath, raw, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	file, err := os.CreateTemp(t.TempDir(), "stderr-*.log")
	if err != nil {
		t.Fatalf("create temp stderr file: %v", err)
	}
	defer os.Remove(file.Name())

	original := os.Stderr
	os.Stderr = file
	defer func() {
		os.Stderr = original
	}()

	fn()

	if err := file.Close(); err != nil {
		t.Fatalf("close temp stderr file: %v", err)
	}
	raw, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatalf("read temp stderr file: %v", err)
	}
	return string(raw)
}

func expectedBinShimPath(binDir string, name string) string {
	if strings.EqualFold(runtime.GOOS, "windows") {
		return filepath.Join(binDir, name+".cmd")
	}
	return filepath.Join(binDir, name)
}

// Tests for native install (choco/winget) functionality

func TestDetectInstallerType(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"installer.msi", "msi"},
		{"setup.exe", "exe"},
		{"package.msix", "msix"},
		{"app.appx", "appx"},
		{"INSTALLER.MSI", "msi"},
		{"Setup.EXE", "exe"},
		{"unknown.zip", "unknown"},
		{"noextension", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := detectInstallerType(tt.path)
			if result != tt.expected {
				t.Errorf("detectInstallerType(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestIsCommandAvailable(t *testing.T) {
	// Test with a command that should exist on Windows
	if !isCommandAvailable("cmd") {
		t.Error("expected cmd to be available on Windows")
	}

	// Test with a command that should not exist
	if isCommandAvailable("this-command-definitely-does-not-exist-12345") {
		t.Error("expected non-existent command to return false")
	}
}

func TestResolveChocoInstallPlan(t *testing.T) {
	server := newChocoTestServer(t)
	defer server.Close()

	// Override builtin mirrors for this test
	original := builtinMirrors["choco"]
	builtinMirrors["choco"] = []Mirror{
		{Source: "choco", Name: "test", BaseURL: server.URL},
	}
	defer func() { builtinMirrors["choco"] = original }()

	plan, err := resolveChocoInstallPlan(context.Background(), server.Client(), InstallRequest{
		Name:    "curl",
		Version: "8.20.0",
		Mirror:  server.URL,
	})
	if err != nil {
		t.Fatalf("resolveChocoInstallPlan returned error: %v", err)
	}

	if plan.Source != "choco" {
		t.Errorf("expected source 'choco', got %q", plan.Source)
	}
	if plan.Root != "curl" {
		t.Errorf("expected root 'curl', got %q", plan.Root)
	}
	if len(plan.Packages) != 1 {
		t.Errorf("expected 1 package, got %d", len(plan.Packages))
	}
	if plan.Packages[0].Name != "curl" {
		t.Errorf("expected package name 'curl', got %q", plan.Packages[0].Name)
	}
}

func TestResolveChocoInstallPlanRequiresName(t *testing.T) {
	_, err := resolveChocoInstallPlan(context.Background(), &http.Client{}, InstallRequest{
		Name: "",
	})
	if err == nil || !strings.Contains(err.Error(), "--name is required") {
		t.Errorf("expected error about missing name, got %v", err)
	}
}

func TestResolveWingetInstallPlan(t *testing.T) {
	t.Skip("Skipping winget test - requires GitHub API access")
	
	server := newWingetTestServer(t)
	defer server.Close()

	plan, err := resolveWingetInstallPlan(context.Background(), server.Client(), InstallRequest{
		Name:    "Microsoft.PowerToys",
		Version: "0.99.1",
		Mirror:  server.URL,
	})
	if err != nil {
		t.Fatalf("resolveWingetInstallPlan returned error: %v", err)
	}

	if plan.Source != "winget" {
		t.Errorf("expected source 'winget', got %q", plan.Source)
	}
	if plan.Root != "Microsoft.PowerToys" {
		t.Errorf("expected root 'Microsoft.PowerToys', got %q", plan.Root)
	}
	if len(plan.Packages) != 1 {
		t.Errorf("expected 1 package, got %d", len(plan.Packages))
	}
}

func TestResolveWingetInstallPlanRequiresName(t *testing.T) {
	_, err := resolveWingetInstallPlan(context.Background(), &http.Client{}, InstallRequest{
		Name: "",
	})
	if err == nil || !strings.Contains(err.Error(), "--name") {
		t.Errorf("expected error about missing name, got %v", err)
	}
}

func TestResolveNativeInstallPlanSupportsChocoAndWinget(t *testing.T) {
	tests := []struct {
		source      string
		shouldError bool
		errorMsg    string
	}{
		{"choco", false, ""},
		{"winget", false, ""},
		{"CHOCO", false, ""},
		{"WINGET", false, ""},
		{"npm", true, "native install plan only supports choco and winget"},
		{"pip", true, "native install plan only supports choco and winget"},
		{"", true, "native install plan only supports choco and winget"},
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			_, err := resolveNativeInstallPlan(context.Background(), &http.Client{}, InstallRequest{
				Source: tt.source,
				Name:   "test-package",
			})
			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error for source %q, got nil", tt.source)
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, got %v", tt.errorMsg, err)
				}
			} else {
				// For valid sources, we expect errors about missing name or network issues, not source validation errors
				if err != nil && strings.Contains(err.Error(), "native install plan only supports") {
					t.Errorf("unexpected source validation error for source %q: %v", tt.source, err)
				}
			}
		})
	}
}

// Helper function to create a test choco server
func newChocoTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/FindPackagesById") {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <entry>
    <id>curl</id>
    <title>curl</title>
    <content type="application/xml">
      <m:properties xmlns:m="http://schemas.microsoft.com/ado/2007/08/dataservices/metadata">
        <d:Version xmlns:d="http://schemas.microsoft.com/ado/2007/08/dataservices">8.20.0</d:Version>
      </m:properties>
    </content>
  </entry>
</feed>`))
		} else if strings.Contains(r.URL.Path, "/package/") {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("fake-nupkg-content"))
		}
	}))
}

// Helper function to create a test winget server
func newWingetTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/manifests/") {
			w.Header().Set("Content-Type", "application/x-yaml")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`PackageIdentifier: Microsoft.PowerToys
PackageVersion: 0.99.1
Installers:
  - Architecture: x64
    InstallerType: exe
    InstallerUrl: https://github.com/microsoft/PowerToys/releases/download/v0.99.1/PowerToysSetup-0.99.1-x64.exe
`))
		} else if strings.Contains(r.URL.Path, "/releases/") {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("fake-installer-content"))
		}
	}))
}
