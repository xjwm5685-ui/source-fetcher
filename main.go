// Package main implements source-fetcher, a unified package manager for downloading
// and installing packages from multiple sources including npm, Chocolatey, winget,
// pip, cargo, and Maven. It provides both CLI and Web GUI interfaces.
//
// The tool supports:
//   - Multi-source package search and download
//   - Dependency resolution and installation
//   - Mirror support for faster downloads
//   - Batch operations via YAML configuration
//   - Interactive TUI for package selection
//   - Web GUI for browser-based management
//
// Basic usage:
//
//	source-fetcher download --source npm --name react --version latest
//	source-fetcher install --source npm --name react --version ^19
//	source-fetcher search --source npm --query typescript
//	source-fetcher gui --port 8765
package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var version = "1.0.1"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage(os.Stdout)
		return nil
	}

	switch strings.ToLower(strings.TrimSpace(args[0])) {
	case "download":
		return runDownload(args[1:])
	case "install":
		return runInstall(args[1:])
	case "uninstall":
		return runUninstall(args[1:])
	case "repair":
		return runRepair(args[1:])
	case "batch":
		return runBatch(args[1:])
	case "search":
		return runSearch(args[1:])
	case "tui":
		return runTUI(args[1:])
	case "gui":
		return runGUI(args[1:])
	case "mirrors":
		return runMirrors(args[1:])
	case "version", "--version", "-v":
		fmt.Println(version)
		return nil
	case "help", "--help", "-h":
		printUsage(os.Stdout)
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runMirrors(args []string) error {
	fs := flag.NewFlagSet("mirrors", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	source := fs.String("source", "all", "source to test: npm, pip, cargo, maven, choco, winget, all")
	timeout := fs.Duration("timeout", 8*time.Second, "request timeout")

	helpShown, err := parseCommandFlags(fs, args)
	if err != nil {
		return err
	}
	if helpShown {
		return nil
	}

	client := newHTTPClient(*timeout)
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	results, err := testMirrors(ctx, client, strings.ToLower(strings.TrimSpace(*source)))
	if err != nil {
		return err
	}

	printMirrorResults(results)
	return nil
}

func runDownload(args []string) error {
	fs := flag.NewFlagSet("download", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	source := fs.String("source", "", "source: npm, pip, cargo, maven, choco, winget, url")
	name := fs.String("name", "", "package name for npm/choco")
	packageID := fs.String("id", "", "package id for winget")
	versionFlag := fs.String("version", "", "package version or tag")
	rawURL := fs.String("url", "", "direct download url when source=url")
	outputDir := fs.String("output", ".", "output directory")
	mirror := fs.String("mirror", "", "mirror name or base url override")
	configPath := fs.String("config", defaultRuntimeConfigPath, "optional config file for auth profiles")
	authProfile := fs.String("auth-profile", "", "auth profile name from config")
	arch := fs.String("arch", "", "preferred installer architecture for winget")
	installerIndex := fs.Int("installer-index", -1, "installer index for winget")
	resume := fs.Bool("resume", false, "resume an existing .part file when supported by the server")
	chunks := fs.Int("chunks", 1, "number of parallel download chunks when the server supports ranges")
	timeout := fs.Duration("timeout", 30*time.Second, "request timeout")
	listInstallers := fs.Bool("list-installers", false, "list winget installers without downloading")
	resolveOnly := fs.Bool("resolve-only", false, "resolve package metadata and show the final download plan without downloading")
	verbose := fs.Bool("verbose", false, "print verbose resolve and integrity logs to stderr")

	helpShown, err := parseCommandFlags(fs, args)
	if err != nil {
		return err
	}
	if helpShown {
		return nil
	}

	if strings.TrimSpace(*source) == "" {
		return errors.New("--source is required")
	}
	if *chunks <= 0 {
		return errors.New("--chunks must be greater than 0")
	}

	absOutput, err := filepath.Abs(*outputDir)
	if err != nil {
		return fmt.Errorf("resolve output dir: %w", err)
	}
	cfg, err := loadOptionalRuntimeConfigWithPolicy(*configPath, shouldAllowMissingRuntimeConfig(fs, *configPath))
	if err != nil {
		return err
	}
	requestOptions, err := cfg.ResolveRequestOptions(*authProfile)
	if err != nil {
		return err
	}
	requestOptions.Verbose = newVerboseLogger(*verbose, os.Stderr)

	req := DownloadRequest{
		Source:         strings.ToLower(strings.TrimSpace(*source)),
		Name:           strings.TrimSpace(*name),
		PackageID:      strings.TrimSpace(*packageID),
		Version:        strings.TrimSpace(*versionFlag),
		URL:            strings.TrimSpace(*rawURL),
		OutputDir:      absOutput,
		Mirror:         strings.TrimSpace(*mirror),
		Arch:           strings.TrimSpace(*arch),
		InstallerIndex: *installerIndex,
		AuthProfile:    strings.TrimSpace(*authProfile),
		Resume:         *resume,
		Chunks:         *chunks,
		RequestOptions: requestOptions,
	}

	client := newHTTPClient(*timeout)
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	plan, err := resolveDownloadPlan(ctx, client, req)
	if err != nil {
		return err
	}

	if *listInstallers {
		if req.Source != "winget" {
			return errors.New("--list-installers is only supported for --source winget")
		}
		printWingetInstallers(plan)
		return nil
	}
	if *resolveOnly {
		printDownloadPlan(os.Stdout, plan)
		return nil
	}

	result, err := downloadPlan(ctx, client, plan, downloadOptionsFromRequest(req))
	if err != nil {
		return err
	}

	printDownloadResult(plan, result)
	return nil
}

func runInstall(args []string) error {
	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	source := fs.String("source", "npm", "source to install: npm, choco, winget")
	name := fs.String("name", "", "package name (npm/choco) or package ID (winget)")
	versionFlag := fs.String("version", "", "package version, tag, or range")
	outputDir := fs.String("output", ".", "install root directory")
	mirror := fs.String("mirror", "", "mirror name or base url override")
	configPath := fs.String("config", defaultRuntimeConfigPath, "optional config file for auth profiles")
	authProfile := fs.String("auth-profile", "", "auth profile name from config")
	resume := fs.Bool("resume", false, "resume an existing .part file when supported by the server")
	chunks := fs.Int("chunks", 1, "number of parallel download chunks when the server supports ranges")
	timeout := fs.Duration("timeout", 30*time.Second, "request timeout")
	omitOptional := fs.Bool("omit-optional", false, "skip optionalDependencies when resolving npm installs")
	includePeer := fs.Bool("include-peer", false, "include peerDependencies when resolving npm installs")
	includeDev := fs.Bool("include-dev", false, "include devDependencies for the root package")
	lockfilePath := fs.String("lockfile", "", "install lockfile path; defaults to <output>\\source-fetcher-install.lock.json")
	frozenLockfile := fs.Bool("frozen-lockfile", false, "fail when the resolved install does not match the lockfile")
	scriptsPolicy := fs.String("scripts", "none", "lifecycle script policy: none, root, all (preinstall is always blocked)")
	allowScripts := fs.String("allow-scripts", "", "comma-separated package names allowed to run blocked lifecycle scripts")
	planOnly := fs.Bool("plan", false, "resolve dependency tree without installing")
	verbose := fs.Bool("verbose", false, "print verbose resolve, lockfile, integrity, and script logs to stderr")

	helpShown, err := parseCommandFlags(fs, args)
	if err != nil {
		return err
	}
	if helpShown {
		return nil
	}
	if *chunks <= 0 {
		return errors.New("--chunks must be greater than 0")
	}
	if normalized := normalizeScriptsPolicy(*scriptsPolicy); normalized != strings.ToLower(strings.TrimSpace(*scriptsPolicy)) && strings.TrimSpace(*scriptsPolicy) != "" {
		return errors.New("--scripts must be one of: none, root, all")
	}
	if strings.TrimSpace(*name) == "" {
		return errors.New("--name is required")
	}

	absOutput, err := filepath.Abs(*outputDir)
	if err != nil {
		return fmt.Errorf("resolve output dir: %w", err)
	}
	cfg, err := loadOptionalRuntimeConfigWithPolicy(*configPath, shouldAllowMissingRuntimeConfig(fs, *configPath))
	if err != nil {
		return err
	}
	requestOptions, err := cfg.ResolveRequestOptions(*authProfile)
	if err != nil {
		return err
	}
	requestOptions.Verbose = newVerboseLogger(*verbose, os.Stderr)

	effectiveScriptsPolicy, effectiveAllowScripts := resolveInstallScriptSettings(
		fs,
		cfg.InstallDefaults,
		*scriptsPolicy,
		*allowScripts,
	)
	warnIgnoredAllowScripts(os.Stderr, effectiveScriptsPolicy, effectiveAllowScripts)

	client := newHTTPClient(*timeout)
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	plan, err := resolveInstallPlan(ctx, client, InstallRequest{
		Source:         strings.ToLower(strings.TrimSpace(*source)),
		Name:           strings.TrimSpace(*name),
		Version:        strings.TrimSpace(*versionFlag),
		OutputDir:      absOutput,
		Mirror:         strings.TrimSpace(*mirror),
		AuthProfile:    strings.TrimSpace(*authProfile),
		Resume:         *resume,
		Chunks:         *chunks,
		OmitOptional:   *omitOptional,
		IncludePeer:    *includePeer,
		IncludeDev:     *includeDev,
		LockfilePath:   strings.TrimSpace(*lockfilePath),
		FrozenLockfile: *frozenLockfile,
		ScriptsPolicy:  effectiveScriptsPolicy,
		AllowScripts:   effectiveAllowScripts,
		RequestOptions: requestOptions,
	})
	if err != nil {
		return err
	}
	if *planOnly {
		printInstallPlan(os.Stdout, plan)
		return nil
	}

	// 根据 source 类型选择执行函数
	var result InstallResult
	switch plan.Source {
	case "npm":
		result, err = executeInstallPlan(ctx, client, plan, DownloadOptions{
			OutputDir:      absOutput,
			Resume:         *resume,
			Chunks:         *chunks,
			RequestOptions: requestOptions,
		})
	case "choco", "winget":
		result, err = executeNativeInstallPlan(ctx, client, plan, DownloadOptions{
			OutputDir:      absOutput,
			Resume:         *resume,
			Chunks:         *chunks,
			RequestOptions: requestOptions,
		})
	default:
		return fmt.Errorf("unsupported install source: %s", plan.Source)
	}
	
	if err != nil {
		return err
	}
	printInstallResult(os.Stdout, plan, result)
	return nil
}

func resolveInstallScriptSettings(fs *flag.FlagSet, defaults InstallDefaultsConfig, cliScriptsPolicy string, cliAllowScripts string) (string, []string) {
	effectiveScriptsPolicy := defaults.ScriptsPolicy
	if flagWasSet(fs, "scripts") {
		effectiveScriptsPolicy = normalizeScriptsPolicy(cliScriptsPolicy)
	}
	effectiveAllowScripts := append([]string(nil), defaults.AllowScripts...)
	if flagWasSet(fs, "allow-scripts") {
		effectiveAllowScripts = parseAllowedScriptPackagesFlag(cliAllowScripts)
	}
	return effectiveScriptsPolicy, effectiveAllowScripts
}

func warnIgnoredAllowScripts(out io.Writer, scriptsPolicy string, allowScripts []string) {
	if out == nil {
		return
	}
	if normalizeScriptsPolicy(scriptsPolicy) != "none" {
		return
	}
	allowed := normalizeAllowedScriptPackages(allowScripts)
	if len(allowed) == 0 {
		return
	}
	_, _ = fmt.Fprintf(out, "Warning: --allow-scripts has no effect when --scripts=none (ignored: %s)\n", strings.Join(allowed, ", "))
}

func runUninstall(args []string) error {
	fs := flag.NewFlagSet("uninstall", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	outputDir := fs.String("output", ".", "install root directory")
	manifestPath := fs.String("manifest", "", "explicit install manifest path; defaults to <output>\\source-fetcher-install.json")
	keepCache := fs.Bool("keep-cache", false, "keep cached tarballs and extracted store contents")
	keepManifest := fs.Bool("keep-manifest", false, "keep the install manifest after uninstall completes")

	helpShown, err := parseCommandFlags(fs, args)
	if err != nil {
		return err
	}
	if helpShown {
		return nil
	}

	result, err := executeUninstall(UninstallRequest{
		OutputDir:    strings.TrimSpace(*outputDir),
		ManifestPath: strings.TrimSpace(*manifestPath),
		KeepCache:    *keepCache,
		KeepManifest: *keepManifest,
	})
	if err != nil {
		return err
	}
	printUninstallResult(os.Stdout, result)
	return nil
}

func runRepair(args []string) error {
	fs := flag.NewFlagSet("repair", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	outputDir := fs.String("output", ".", "install root directory")
	manifestPath := fs.String("manifest", "", "explicit install manifest path; defaults to <output>\\source-fetcher-install.json")

	helpShown, err := parseCommandFlags(fs, args)
	if err != nil {
		return err
	}
	if helpShown {
		return nil
	}

	result, err := executeRepair(RepairRequest{
		OutputDir:    strings.TrimSpace(*outputDir),
		ManifestPath: strings.TrimSpace(*manifestPath),
	})
	if err != nil {
		return err
	}
	printRepairResult(os.Stdout, result)
	return nil
}

func runBatch(args []string) error {
	fs := flag.NewFlagSet("batch", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	configPath := fs.String("config", "source-fetcher.yaml", "batch config file path")
	planOnly := fs.Bool("plan", false, "resolve all configured tasks and print plans without downloading")
	continueOnError := fs.Bool("continue-on-error", false, "continue remaining tasks when one task fails")
	jobs := fs.Int("jobs", 1, "number of batch tasks to run concurrently")
	retries := fs.Int("retries", 0, "retry each batch resolve/download/install step this many times after the first failure")
	retryBackoff := fs.Duration("retry-backoff", 0, "base delay between batch retries; doubles after each failed attempt")
	verbose := fs.Bool("verbose", false, "print verbose resolve, lockfile, integrity, and script logs to stderr")

	helpShown, err := parseCommandFlags(fs, args)
	if err != nil {
		return err
	}
	if helpShown {
		return nil
	}
	if *retries < 0 {
		return errors.New("--retries must be greater than or equal to 0")
	}
	if *retryBackoff < 0 {
		return errors.New("--retry-backoff must be greater than or equal to 0")
	}
	if *jobs <= 0 {
		return errors.New("--jobs must be greater than 0")
	}

	cfg, err := loadBatchConfig(*configPath)
	if err != nil {
		return err
	}
	if len(cfg.Downloads) == 0 && len(cfg.Installs) == 0 {
		return errors.New("batch config contains no downloads or installs")
	}

	timeout, err := cfg.TimeoutValue()
	if err != nil {
		return err
	}
	client := newHTTPClient(timeout)
	logger := newVerboseLogger(*verbose, os.Stderr)
	retryPolicy := batchRetryPolicy{Retries: *retries, Backoff: *retryBackoff}
	if err := runBatchWithOps(client, cfg, timeout, *jobs, retryPolicy, *planOnly, *continueOnError, os.Stdout, logger, resolveDownloadPlan, downloadPlan); err != nil {
		if len(cfg.Installs) == 0 {
			return err
		}
		if !*continueOnError {
			return err
		}
		fmt.Fprintf(os.Stdout, "download batch failed: %v\n", err)
	}
	if len(cfg.Installs) == 0 {
		return nil
	}
	return runBatchInstallsWithOps(client, cfg, timeout, *jobs, retryPolicy, *planOnly, *continueOnError, os.Stdout, logger, resolveInstallPlan, executeInstallPlan)
}

type batchResolver func(ctx context.Context, client *http.Client, req DownloadRequest) (DownloadPlan, error)
type batchDownloader func(ctx context.Context, client *http.Client, plan DownloadPlan, options DownloadOptions) (DownloadResult, error)
type batchRetryPolicy struct {
	Retries int
	Backoff time.Duration
}

func batchRetryAttempts(policy batchRetryPolicy) int {
	if policy.Retries < 0 {
		return 1
	}
	return policy.Retries + 1
}

func batchRetryDelay(base time.Duration, retryIndex int) time.Duration {
	if base <= 0 || retryIndex <= 0 {
		return 0
	}
	delay := base
	for step := 1; step < retryIndex; step++ {
		if delay > time.Duration((1<<63-1)/2) {
			return time.Duration(1<<63 - 1)
		}
		delay *= 2
	}
	return delay
}

func runBatchStepWithRetry(
	parent context.Context,
	timeout time.Duration,
	policy batchRetryPolicy,
	out io.Writer,
	step string,
	run func(context.Context) error,
) error {
	attempts := batchRetryAttempts(policy)
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		if err := parent.Err(); err != nil {
			if lastErr != nil {
				return err
			}
			return err
		}
		taskCtx, cancel := context.WithTimeout(parent, timeout)
		err := run(taskCtx)
		cancel()
		if err == nil {
			return nil
		}
		lastErr = err
		if attempt == attempts {
			break
		}
		if out != nil {
			fmt.Fprintf(out, "  %s attempt %d/%d failed: %v\n", step, attempt, attempts, err)
		}
		delay := batchRetryDelay(policy.Backoff, attempt)
		if out != nil {
			if delay > 0 {
				fmt.Fprintf(out, "  retrying %s in %s\n", step, roundDuration(delay))
			} else {
				fmt.Fprintf(out, "  retrying %s immediately\n", step)
			}
		}
		if delay > 0 {
			timer := time.NewTimer(delay)
			select {
			case <-parent.Done():
				if !timer.Stop() {
					<-timer.C
				}
				return parent.Err()
			case <-timer.C:
			}
		}
	}
	return lastErr
}

func flushBatchTaskOutput(mu *sync.Mutex, out io.Writer, data []byte) {
	if mu == nil || out == nil || len(data) == 0 {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	_, _ = out.Write(data)
}

func runBatchWorkers(
	jobs int,
	total int,
	continueOnError bool,
	runTask func(context.Context, int) error,
) (int, error) {
	if total == 0 {
		return 0, nil
	}
	if jobs <= 0 {
		jobs = 1
	}
	if jobs > total {
		jobs = total
	}
	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	taskIndexes := make(chan int, jobs)
	var wg sync.WaitGroup
	var stateMu sync.Mutex
	failed := 0
	var firstErr error

	for worker := 0; worker < jobs; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for index := range taskIndexes {
				if !continueOnError && rootCtx.Err() != nil {
					continue
				}
				err := runTask(rootCtx, index)
				if err == nil {
					continue
				}
				stateMu.Lock()
				failed++
				if firstErr == nil {
					firstErr = err
				}
				stateMu.Unlock()
				if !continueOnError {
					cancel()
				}
			}
		}()
	}

enqueueLoop:
	for index := 0; index < total; index++ {
		if !continueOnError {
			select {
			case <-rootCtx.Done():
				break enqueueLoop
			default:
			}
		}
		select {
		case <-rootCtx.Done():
			break enqueueLoop
		case taskIndexes <- index:
		}
	}
	close(taskIndexes)
	wg.Wait()

	return failed, firstErr
}

func runBatchWithOps(
	client *http.Client,
	cfg BatchConfig,
	timeout time.Duration,
	jobs int,
	retryPolicy batchRetryPolicy,
	planOnly bool,
	continueOnError bool,
	out io.Writer,
	logger *VerboseLogger,
	resolve batchResolver,
	download batchDownloader,
) error {
	if len(cfg.Downloads) == 0 {
		return nil
	}
	absOutput, err := filepath.Abs(cfg.OutputDir)
	if err != nil {
		return fmt.Errorf("resolve batch output dir: %w", err)
	}

	var outputMu sync.Mutex
	failed, firstErr := runBatchWorkers(jobs, len(cfg.Downloads), continueOnError, func(rootCtx context.Context, index int) (taskErr error) {
		task := cfg.Downloads[index]
		req, err := cfg.BindDownloadRequest(task)
		if err != nil {
			return err
		}
		if strings.TrimSpace(req.OutputDir) == "" {
			req.OutputDir = absOutput
		}
		req.RequestOptions.Verbose = logger

		var taskOut bytes.Buffer
		defer func() {
			flushBatchTaskOutput(&outputMu, out, taskOut.Bytes())
		}()
		defer func() {
			if recovered := recover(); recovered != nil {
				taskErr = fmt.Errorf("batch download task %s panicked: %v", taskDisplayName(req), recovered)
				fmt.Fprintf(&taskOut, "  task failed: %v\n", taskErr)
			}
		}()

		fmt.Fprintf(&taskOut, "[%d/%d] %s\n", index+1, len(cfg.Downloads), taskDisplayName(req))
		var plan DownloadPlan
		err = runBatchStepWithRetry(rootCtx, timeout, retryPolicy, &taskOut, "resolve", func(taskCtx context.Context) error {
			resolvedPlan, resolveErr := resolve(taskCtx, client, req)
			if resolveErr != nil {
				return resolveErr
			}
			plan = resolvedPlan
			return nil
		})
		if err != nil {
			fmt.Fprintf(&taskOut, "  resolve failed: %v\n", err)
			return err
		}

		if planOnly {
			printDownloadPlan(&taskOut, plan)
			fmt.Fprintln(&taskOut, "")
			return nil
		}

		var result DownloadResult
		err = runBatchStepWithRetry(rootCtx, timeout, retryPolicy, &taskOut, "download", func(taskCtx context.Context) error {
			downloadResult, downloadErr := download(taskCtx, client, plan, downloadOptionsFromRequest(req))
			if downloadErr != nil {
				return downloadErr
			}
			result = downloadResult
			return nil
		})
		if err != nil {
			fmt.Fprintf(&taskOut, "  download failed: %v\n", err)
			return err
		}

		printDownloadResultTo(&taskOut, plan, result)
		fmt.Fprintln(&taskOut, "")
		return nil
	})
	if !continueOnError && firstErr != nil {
		return firstErr
	}
	if failed > 0 {
		return fmt.Errorf("batch completed with %d failed task(s)", failed)
	}
	return nil
}

type batchInstallResolver func(ctx context.Context, client *http.Client, req InstallRequest) (InstallPlan, error)
type batchInstaller func(ctx context.Context, client *http.Client, plan InstallPlan, options DownloadOptions) (InstallResult, error)

func runBatchInstallsWithOps(
	client *http.Client,
	cfg BatchConfig,
	timeout time.Duration,
	jobs int,
	retryPolicy batchRetryPolicy,
	planOnly bool,
	continueOnError bool,
	out io.Writer,
	logger *VerboseLogger,
	resolve batchInstallResolver,
	install batchInstaller,
) error {
	if len(cfg.Installs) == 0 {
		return nil
	}
	absOutput, err := filepath.Abs(cfg.OutputDir)
	if err != nil {
		return fmt.Errorf("resolve batch output dir: %w", err)
	}

	var outputMu sync.Mutex
	failed, firstErr := runBatchWorkers(jobs, len(cfg.Installs), continueOnError, func(rootCtx context.Context, index int) (taskErr error) {
		task := cfg.Installs[index]
		req, err := cfg.BindInstallRequest(task)
		if err != nil {
			return err
		}
		if strings.TrimSpace(req.OutputDir) == "" {
			req.OutputDir = absOutput
		}
		req.RequestOptions.Verbose = logger
		warnIgnoredAllowScripts(os.Stderr, req.ScriptsPolicy, req.AllowScripts)

		var taskOut bytes.Buffer
		defer func() {
			flushBatchTaskOutput(&outputMu, out, taskOut.Bytes())
		}()
		defer func() {
			if recovered := recover(); recovered != nil {
				taskErr = fmt.Errorf("batch install task %s@%s panicked: %v", req.Name, blankIfEmpty(req.Version, "latest"), recovered)
				fmt.Fprintf(&taskOut, "  task failed: %v\n", taskErr)
			}
		}()

		fmt.Fprintf(&taskOut, "[install %d/%d] %s@%s\n", index+1, len(cfg.Installs), req.Name, blankIfEmpty(req.Version, "latest"))
		var plan InstallPlan
		err = runBatchStepWithRetry(rootCtx, timeout, retryPolicy, &taskOut, "resolve", func(taskCtx context.Context) error {
			resolvedPlan, resolveErr := resolve(taskCtx, client, req)
			if resolveErr != nil {
				return resolveErr
			}
			plan = resolvedPlan
			return nil
		})
		if err != nil {
			fmt.Fprintf(&taskOut, "  resolve failed: %v\n", err)
			return err
		}
		if planOnly {
			printInstallPlan(&taskOut, plan)
			fmt.Fprintln(&taskOut, "")
			return nil
		}

		var result InstallResult
		err = runBatchStepWithRetry(rootCtx, timeout, retryPolicy, &taskOut, "install", func(taskCtx context.Context) error {
			installResult, installErr := install(taskCtx, client, plan, DownloadOptions{
				OutputDir:      req.OutputDir,
				Resume:         req.Resume,
				Chunks:         req.Chunks,
				RequestOptions: req.RequestOptions,
			})
			if installErr != nil {
				return installErr
			}
			result = installResult
			return nil
		})
		if err != nil {
			fmt.Fprintf(&taskOut, "  install failed: %v\n", err)
			return err
		}
		printInstallResult(&taskOut, plan, result)
		fmt.Fprintln(&taskOut, "")
		return nil
	})
	if !continueOnError && firstErr != nil {
		return firstErr
	}

	if failed > 0 {
		return fmt.Errorf("batch completed with %d failed task(s)", failed)
	}
	return nil
}

func runSearch(args []string) error {
	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	source := fs.String("source", "all", "source to search: npm, pip, cargo, maven, choco, winget, all")
	query := fs.String("query", "", "search query")
	mirror := fs.String("mirror", "", "mirror name or base url override when supported")
	configPath := fs.String("config", defaultRuntimeConfigPath, "optional config file for auth profiles")
	authProfile := fs.String("auth-profile", "", "auth profile name from config")
	limit := fs.Int("limit", 10, "max results per source")
	pick := fs.String("pick", "", "pick search result indexes like 2 or 1,3,5 and download them directly")
	interactive := fs.Bool("interactive", false, "interactively choose one or more search results to download")
	outputDir := fs.String("output", ".", "output directory when using --pick or --interactive")
	resolveOnly := fs.Bool("resolve-only", false, "resolve the picked search result(s) without downloading")
	arch := fs.String("arch", "", "preferred installer architecture for winget when using --pick or --interactive")
	installerIndex := fs.Int("installer-index", -1, "installer index for winget when using --pick or --interactive")
	resume := fs.Bool("resume", false, "resume an existing .part file when supported by the server")
	chunks := fs.Int("chunks", 1, "number of parallel download chunks when the server supports ranges")
	timeout := fs.Duration("timeout", 20*time.Second, "request timeout")
	verbose := fs.Bool("verbose", false, "print verbose resolve logs to stderr")

	helpShown, err := parseCommandFlags(fs, args)
	if err != nil {
		return err
	}
	if helpShown {
		return nil
	}
	if strings.TrimSpace(*query) == "" {
		return errors.New("--query is required")
	}
	if *limit <= 0 {
		return errors.New("--limit must be greater than 0")
	}
	if *chunks <= 0 {
		return errors.New("--chunks must be greater than 0")
	}
	pickSpec := strings.TrimSpace(*pick)
	if *interactive && pickSpec != "" {
		return errors.New("--interactive cannot be used together with --pick")
	}

	cfg, err := loadOptionalRuntimeConfigWithPolicy(*configPath, shouldAllowMissingRuntimeConfig(fs, *configPath))
	if err != nil {
		return err
	}
	requestOptions, err := cfg.ResolveRequestOptions(*authProfile)
	if err != nil {
		return err
	}
	requestOptions.Verbose = newVerboseLogger(*verbose, os.Stderr)

	client := newHTTPClient(*timeout)
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	results, err := searchPackages(ctx, client, SearchRequest{
		Source:         strings.ToLower(strings.TrimSpace(*source)),
		Query:          strings.TrimSpace(*query),
		Mirror:         strings.TrimSpace(*mirror),
		Limit:          *limit,
		AuthProfile:    strings.TrimSpace(*authProfile),
		RequestOptions: requestOptions,
	})
	if err != nil {
		return err
	}
	if warning := searchFallbackWarning(strings.TrimSpace(*source), strings.TrimSpace(*mirror), results); warning != "" {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: %s\n", warning)
	}
	selectedPicks, err := parsePickSpec(pickSpec, len(results))
	if err != nil {
		return fmt.Errorf("parse --pick: %w", err)
	}
	if *interactive {
		selected, err := promptSearchPicks(os.Stdin, os.Stdout, results)
		if err != nil {
			return err
		}
		if len(selected) == 0 {
			return nil
		}
		selectedPicks = selected
	}
	if len(selectedPicks) == 0 {
		printSearchResults(os.Stdout, results)
		return nil
	}

	absOutput, err := filepath.Abs(*outputDir)
	if err != nil {
		return fmt.Errorf("resolve output dir: %w", err)
	}
	requests, err := downloadRequestsFromSearchResults(results, selectedPicks, SearchPickOptions{
		Mirror:         strings.TrimSpace(*mirror),
		OutputDir:      absOutput,
		ResolveOnly:    *resolveOnly,
		Arch:           strings.TrimSpace(*arch),
		InstallerIndex: *installerIndex,
		AuthProfile:    strings.TrimSpace(*authProfile),
		Resume:         *resume,
		Chunks:         *chunks,
		RequestOptions: requestOptions,
	})
	if err != nil {
		return err
	}
	for index, req := range requests {
		if len(requests) > 1 {
			fmt.Fprintf(os.Stdout, "[%d/%d] %s\n", index+1, len(requests), taskDisplayName(req))
		}

		actionCtx, actionCancel := context.WithTimeout(context.Background(), *timeout)
		plan, err := resolveDownloadPlan(actionCtx, client, req)
		if err != nil {
			actionCancel()
			return fmt.Errorf("resolve pick %d: %w", selectedPicks[index], err)
		}
		if *resolveOnly {
			printDownloadPlan(os.Stdout, plan)
			actionCancel()
		} else {
			result, err := downloadPlan(actionCtx, client, plan, downloadOptionsFromRequest(req))
			actionCancel()
			if err != nil {
				return fmt.Errorf("download pick %d: %w", selectedPicks[index], err)
			}
			printDownloadResult(plan, result)
		}

		if len(requests) > 1 && index < len(requests)-1 {
			fmt.Fprintln(os.Stdout, "")
		}
	}
	return nil
}

type SearchPickOptions struct {
	Mirror         string
	OutputDir      string
	ResolveOnly    bool
	Arch           string
	InstallerIndex int
	AuthProfile    string
	Resume         bool
	Chunks         int
	RequestOptions RequestOptions
}

func downloadRequestFromSearchResult(results []SearchResult, pick int, options SearchPickOptions) (DownloadRequest, error) {
	if len(results) == 0 {
		return DownloadRequest{}, errors.New("no search results to pick from")
	}
	if pick <= 0 || pick > len(results) {
		return DownloadRequest{}, fmt.Errorf("--pick %d is out of range; available results: 1-%d", pick, len(results))
	}

	selected := results[pick-1]
	req := DownloadRequest{
		Source:         strings.TrimSpace(selected.Source),
		Version:        strings.TrimSpace(selected.Version),
		OutputDir:      strings.TrimSpace(options.OutputDir),
		Mirror:         strings.TrimSpace(options.Mirror),
		Arch:           strings.TrimSpace(options.Arch),
		InstallerIndex: options.InstallerIndex,
		AuthProfile:    strings.TrimSpace(options.AuthProfile),
		Resume:         options.Resume,
		Chunks:         options.Chunks,
		RequestOptions: options.RequestOptions,
	}
	switch req.Source {
	case "npm", "choco":
		req.Name = strings.TrimSpace(selected.Identifier)
	case "winget":
		req.PackageID = strings.TrimSpace(selected.Identifier)
	default:
		return DownloadRequest{}, fmt.Errorf("search result source %q cannot be downloaded directly", selected.Source)
	}
	return req, nil
}

func downloadRequestsFromSearchResults(results []SearchResult, picks []int, options SearchPickOptions) ([]DownloadRequest, error) {
	if len(picks) == 0 {
		return nil, errors.New("no search results selected")
	}

	requests := make([]DownloadRequest, 0, len(picks))
	for _, pick := range picks {
		req, err := downloadRequestFromSearchResult(results, pick, options)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	return requests, nil
}

func parsePickSpec(spec string, max int) ([]int, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil, nil
	}
	if max == 0 {
		return nil, errors.New("no search results to pick from")
	}

	parts := strings.Split(spec, ",")
	picks := make([]int, 0, len(parts))
	seen := make(map[int]struct{}, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, errors.New("empty index in pick list")
		}

		pick, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid index %q", part)
		}
		if pick <= 0 || pick > max {
			return nil, fmt.Errorf("index %d is out of range; available results: 1-%d", pick, max)
		}
		if _, ok := seen[pick]; ok {
			continue
		}

		seen[pick] = struct{}{}
		picks = append(picks, pick)
	}
	return picks, nil
}

func promptSearchPicks(in io.Reader, out io.Writer, results []SearchResult) ([]int, error) {
	if len(results) == 0 {
		_, _ = fmt.Fprintln(out, "No results.")
		return nil, nil
	}

	printSearchResults(out, results)
	_, _ = fmt.Fprintln(out, "")
	_, _ = fmt.Fprintf(out, "Choose result indexes like 1 or 1,3,5 (1-%d), or enter q to cancel: ", len(results))

	reader := bufio.NewReader(in)
	for {
		line, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("read interactive selection: %w", err)
		}
		choice := strings.TrimSpace(line)
		switch {
		case choice == "" && errors.Is(err, io.EOF):
			return nil, nil
		case choice == "":
			_, _ = fmt.Fprintf(out, "Enter indexes like 1 or 1,3,5, or q to cancel: ")
		case strings.EqualFold(choice, "q"), strings.EqualFold(choice, "quit"), strings.EqualFold(choice, "exit"):
			_, _ = fmt.Fprintln(out, "Cancelled.")
			return nil, nil
		default:
			picks, parseErr := parsePickSpec(choice, len(results))
			if parseErr == nil {
				return picks, nil
			}
			_, _ = fmt.Fprintf(out, "Invalid choice %q. Enter indexes like 1 or 1,3,5, or q to cancel: ", choice)
		}
		if errors.Is(err, io.EOF) {
			return nil, nil
		}
	}
}

func printUsage(out *os.File) {
	_, _ = fmt.Fprintln(out, "source-fetcher")
	_, _ = fmt.Fprintln(out, "")
	_, _ = fmt.Fprintln(out, "Commands:")
	_, _ = fmt.Fprintln(out, "  mirrors   test built-in mirrors without requiring npm/choco/winget to be installed")
	_, _ = fmt.Fprintln(out, "  download  resolve a package and download it directly")
	_, _ = fmt.Fprintln(out, "  install   resolve npm dependencies and assemble a node_modules install tree")
	_, _ = fmt.Fprintln(out, "  uninstall remove an install tree using source-fetcher-install.json")
	_, _ = fmt.Fprintln(out, "  repair    repair an install tree using source-fetcher-install.json")
	_, _ = fmt.Fprintln(out, "  batch     run multiple download jobs from a yaml config")
	_, _ = fmt.Fprintln(out, "  search    search packages across supported sources")
	_, _ = fmt.Fprintln(out, "  tui       open an interactive terminal UI")
	_, _ = fmt.Fprintln(out, "")
	_, _ = fmt.Fprintln(out, "Examples:")
	_, _ = fmt.Fprintln(out, "  source-fetcher mirrors --source all")
	_, _ = fmt.Fprintln(out, "  source-fetcher search --source all --query powertoys")
	_, _ = fmt.Fprintln(out, "  source-fetcher search --source pip --query requests")
	_, _ = fmt.Fprintln(out, "  source-fetcher search --source cargo --query serde")
	_, _ = fmt.Fprintln(out, "  source-fetcher search --source winget --query terminal --interactive --resolve-only")
	_, _ = fmt.Fprintln(out, "  source-fetcher search --source winget --query terminal --pick 2 --output .\\downloads")
	_, _ = fmt.Fprintln(out, "  source-fetcher search --source winget --query terminal --pick 1,3 --resolve-only")
	_, _ = fmt.Fprintln(out, "  source-fetcher install --source npm --name react --version ^19 --output .\\workspace")
	_, _ = fmt.Fprintln(out, "  source-fetcher uninstall --output .\\workspace")
	_, _ = fmt.Fprintln(out, "  source-fetcher repair --output .\\workspace")
	_, _ = fmt.Fprintln(out, "  source-fetcher tui --source winget --query terminal --output .\\downloads")
	_, _ = fmt.Fprintln(out, "  source-fetcher download --source npm --name react --version latest --output .\\downloads")
	_, _ = fmt.Fprintln(out, "  source-fetcher download --source pip --name requests --version latest --output .\\downloads")
	_, _ = fmt.Fprintln(out, "  source-fetcher download --source maven --name junit:junit --version latest --output .\\downloads")
	_, _ = fmt.Fprintln(out, "  source-fetcher download --source npm --name react --resolve-only")
	_, _ = fmt.Fprintln(out, "  source-fetcher download --source choco --name git --version 2.47.0 --output .\\downloads")
	_, _ = fmt.Fprintln(out, "  source-fetcher download --source winget --id Microsoft.PowerToys --arch x64 --output .\\downloads")
	_, _ = fmt.Fprintln(out, "  source-fetcher batch --config .\\source-fetcher.yaml")
	_, _ = fmt.Fprintln(out, "  source-fetcher batch --config .\\source-fetcher.yaml --jobs 2 --continue-on-error --retries 2 --retry-backoff 500ms")
	_, _ = fmt.Fprintln(out, "  source-fetcher batch --config .\\source-fetcher.yaml --continue-on-error --retries 2 --retry-backoff 500ms")
	_, _ = fmt.Fprintln(out, "  source-fetcher download --source url --url https://example.com/file.zip --output .\\downloads")
}

func printMirrorResults(results []MirrorResult) {
	fmt.Printf("%-8s %-14s %-5s %-10s %-10s %s\n", "SOURCE", "MIRROR", "OK", "FIRSTBYTE", "TOTAL", "DETAIL")
	for _, result := range results {
		okLabel := "no"
		detail := result.Detail
		if result.OK {
			okLabel = "yes"
			if detail == "" {
				detail = fmt.Sprintf("http %d", result.StatusCode)
			}
		}
		fmt.Printf(
			"%-8s %-14s %-5s %-10s %-10s %s\n",
			result.Source,
			result.Name,
			okLabel,
			roundDuration(result.FirstByte),
			roundDuration(result.Total),
			detail,
		)
	}
}

func printWingetInstallers(plan DownloadPlan) {
	fmt.Printf("Package: %s\n", plan.Identifier)
	fmt.Printf("Version: %s\n", plan.Version)
	fmt.Printf("Manifest: %s\n", plan.ManifestURL)
	fmt.Println("")
	fmt.Printf("%-5s %-8s %-10s %-10s %s\n", "INDEX", "ARCH", "TYPE", "SCOPE", "URL")
	for index, installer := range plan.Installers {
		fmt.Printf(
			"%-5d %-8s %-10s %-10s %s\n",
			index,
			blankIfEmpty(installer.Architecture, "-"),
			blankIfEmpty(installer.InstallerType, "-"),
			blankIfEmpty(installer.Scope, "-"),
			installer.InstallerURL,
		)
	}
}

func printDownloadResult(plan DownloadPlan, result DownloadResult) {
	printDownloadResultTo(os.Stdout, plan, result)
}

func printDownloadResultTo(out io.Writer, plan DownloadPlan, result DownloadResult) {
	fmt.Fprintf(out, "Source: %s\n", plan.Source)
	fmt.Fprintf(out, "Identifier: %s\n", plan.Identifier)
	fmt.Fprintf(out, "Version: %s\n", blankIfEmpty(plan.Version, "-"))
	if plan.MirrorName != "" {
		fmt.Fprintf(out, "Mirror: %s\n", plan.MirrorName)
	}
	if plan.ManifestURL != "" {
		fmt.Fprintf(out, "Manifest: %s\n", plan.ManifestURL)
	}
	fmt.Fprintf(out, "Resolved URL: %s\n", plan.URL)
	fmt.Fprintf(out, "Saved To: %s\n", result.Path)
	fmt.Fprintf(out, "Size: %d bytes\n", result.Size)
	fmt.Fprintf(out, "SHA256: %s\n", result.SHA256)
	fmt.Fprintf(out, "Duration: %s\n", roundDuration(result.Duration))
}

func printDownloadPlan(out io.Writer, plan DownloadPlan) {
	_, _ = fmt.Fprintf(out, "Source: %s\n", plan.Source)
	_, _ = fmt.Fprintf(out, "Identifier: %s\n", plan.Identifier)
	_, _ = fmt.Fprintf(out, "Version: %s\n", blankIfEmpty(plan.Version, "-"))
	if plan.MirrorName != "" {
		_, _ = fmt.Fprintf(out, "Mirror: %s\n", plan.MirrorName)
	}
	if plan.ManifestURL != "" {
		_, _ = fmt.Fprintf(out, "Manifest: %s\n", plan.ManifestURL)
	}
	_, _ = fmt.Fprintf(out, "Resolved URL: %s\n", plan.URL)
	_, _ = fmt.Fprintf(out, "Filename: %s\n", plan.Filename)
}

func printSearchResults(out io.Writer, results []SearchResult) {
	if len(results) == 0 {
		_, _ = fmt.Fprintln(out, "No results.")
		return
	}

	_, _ = fmt.Fprintf(out, "%-5s %-8s %-14s %-28s %-14s %-48s %s\n", "INDEX", "SOURCE", "MIRROR", "IDENTIFIER", "VERSION", "DESCRIPTION", "DETAIL")
	for index, result := range results {
		_, _ = fmt.Fprintf(
			out,
			"%-5d %-8s %-14s %-28s %-14s %-48s %s\n",
			index+1,
			result.Source,
			blankIfEmpty(result.MirrorName, "-"),
			truncateText(result.Identifier, 28),
			blankIfEmpty(result.Version, "-"),
			truncateText(result.Description, 48),
			truncateText(result.Detail, 80),
		)
	}
}

func searchFallbackWarning(source string, requestedMirror string, results []SearchResult) string {
	source = strings.ToLower(strings.TrimSpace(source))
	requestedMirror = strings.TrimSpace(requestedMirror)
	if requestedMirror == "" {
		return ""
	}
	for _, result := range results {
		if source != "" && source != "all" && !strings.EqualFold(result.Source, source) {
			continue
		}
		if !strings.EqualFold(result.Source, "npm") {
			continue
		}
		if strings.TrimSpace(result.MirrorName) == "" || strings.EqualFold(result.MirrorName, requestedMirror) {
			continue
		}
		return fmt.Sprintf("search mirror %s returned no usable npm results, fell back to %s", requestedMirror, result.MirrorName)
	}
	return ""
}

func roundDuration(value time.Duration) string {
	if value <= 0 {
		return "-"
	}
	return value.Round(time.Millisecond).String()
}

func newHTTPClient(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = (&net.Dialer{
		Timeout:   minDuration(15*time.Second, timeout),
		KeepAlive: 30 * time.Second,
	}).DialContext
	transport.TLSHandshakeTimeout = minDuration(20*time.Second, timeout)
	transport.ResponseHeaderTimeout = minDuration(20*time.Second, timeout)
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

func minDuration(a time.Duration, b time.Duration) time.Duration {
	if a <= 0 {
		return b
	}
	if b <= 0 {
		return a
	}
	if a < b {
		return a
	}
	return b
}

func blankIfEmpty(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func taskDisplayName(req DownloadRequest) string {
	switch req.Source {
	case "winget":
		return req.Source + ":" + req.PackageID
	case "url":
		return req.Source + ":" + req.URL
	default:
		return req.Source + ":" + req.Name
	}
}

func installTaskDisplayName(req InstallRequest) string {
	source := blankIfEmpty(strings.TrimSpace(req.Source), "npm")
	return source + ":" + strings.TrimSpace(req.Name)
}

func truncateText(value string, max int) string {
	value = strings.TrimSpace(value)
	if max <= 0 || len([]rune(value)) <= max {
		return value
	}
	runes := []rune(value)
	if max <= 1 {
		return string(runes[:max])
	}
	return string(runes[:max-1]) + "..."
}

func parseCommandFlags(fs *flag.FlagSet, args []string) (bool, error) {
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

func flagWasSet(fs *flag.FlagSet, name string) bool {
	wasSet := false
	fs.Visit(func(current *flag.Flag) {
		if current.Name == name {
			wasSet = true
		}
	})
	return wasSet
}

func shouldAllowMissingRuntimeConfig(fs *flag.FlagSet, path string) bool {
	return !flagWasSet(fs, "config") && strings.EqualFold(strings.TrimSpace(path), defaultRuntimeConfigPath)
}
