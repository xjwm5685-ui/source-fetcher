package main

import (
	"bytes"
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptrace"
	neturl "net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

const downloadProgressInterval = 200 * time.Millisecond

type Mirror struct {
	Source    string
	Name      string
	BaseURL   string
	ProbePath string
}

type MirrorResult struct {
	Source     string
	Name       string
	BaseURL    string
	StatusCode int
	OK         bool
	FirstByte  time.Duration
	Total      time.Duration
	Detail     string
}

type RequestOptions struct {
	Headers map[string]string
	Verbose *VerboseLogger
}

type DownloadRequest struct {
	Source         string `yaml:"source"`
	Name           string `yaml:"name"`
	PackageID      string `yaml:"id"`
	Version        string `yaml:"version"`
	URL            string `yaml:"url"`
	OutputDir      string `yaml:"output_dir"`
	Mirror         string `yaml:"mirror"`
	Arch           string `yaml:"arch"`
	InstallerIndex int    `yaml:"installer_index"`
	AuthProfile    string `yaml:"auth"`
	Resume         bool   `yaml:"resume"`
	Chunks         int    `yaml:"chunks"`

	RequestOptions RequestOptions `yaml:"-"`
}

type DownloadPlan struct {
	Source      string
	Identifier  string
	Version     string
	URL         string
	Filename    string
	MirrorName  string
	ManifestURL string
	Integrity   string
	Shasum      string
	Installers  []WingetInstaller
}

type DownloadResult struct {
	Path     string
	Size     int64
	SHA256   string
	Duration time.Duration
}

type DownloadProgress struct {
	Written int64
	Total   int64
	Elapsed time.Duration
}

type DownloadOptions struct {
	OutputDir      string
	Resume         bool
	Chunks         int
	RequestOptions RequestOptions
}

type downloadProgressWriter struct {
	mu          sync.Mutex
	out         io.Writer
	enabled     bool
	name        string
	total       int64
	start       time.Time
	lastLineLen int
	lastPrint   time.Time
	written     int64
	onProgress  func(DownloadProgress)
}

type SearchRequest struct {
	Source         string
	Query          string
	Mirror         string
	Limit          int
	AuthProfile    string
	RequestOptions RequestOptions
}

type SearchResult struct {
	Source      string
	MirrorName  string
	Identifier  string
	Version     string
	Description string
	Detail      string
}

type WingetInstaller struct {
	Architecture  string `yaml:"Architecture"`
	InstallerType string `yaml:"InstallerType"`
	InstallerURL  string `yaml:"InstallerUrl"`
	Scope         string `yaml:"Scope"`
}

type wingetInstallerManifest struct {
	PackageIdentifier string            `yaml:"PackageIdentifier"`
	PackageVersion    string            `yaml:"PackageVersion"`
	Installers        []WingetInstaller `yaml:"Installers"`
}

type npmPackageMetadata struct {
	Name     string                       `json:"name"`
	DistTags map[string]string            `json:"dist-tags"`
	Versions map[string]npmVersionDetails `json:"versions"`
}

type npmStringList []string

func (v *npmStringList) UnmarshalJSON(data []byte) error {
	if v == nil {
		return errors.New("npm string list target is nil")
	}
	trimmed := bytes.TrimSpace(data)
	switch {
	case len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")):
		*v = nil
		return nil
	case len(trimmed) >= 2 && trimmed[0] == '"' && trimmed[len(trimmed)-1] == '"':
		var value string
		if err := json.Unmarshal(trimmed, &value); err != nil {
			return err
		}
		value = strings.TrimSpace(value)
		if value == "" {
			*v = nil
		} else {
			*v = npmStringList{value}
		}
		return nil
	case len(trimmed) >= 2 && trimmed[0] == '[' && trimmed[len(trimmed)-1] == ']':
		var values []string
		if err := json.Unmarshal(trimmed, &values); err != nil {
			return err
		}
		normalized := make(npmStringList, 0, len(values))
		for _, value := range values {
			value = strings.TrimSpace(value)
			if value != "" {
				normalized = append(normalized, value)
			}
		}
		if len(normalized) == 0 {
			*v = nil
		} else {
			*v = normalized
		}
		return nil
	default:
		return fmt.Errorf("unsupported npm string list shape: %s", string(trimmed))
	}
}

type npmVersionDetails struct {
	Version              string            `json:"version"`
	Dependencies         map[string]string `json:"dependencies"`
	OptionalDependencies map[string]string `json:"optionalDependencies"`
	PeerDependencies     map[string]string `json:"peerDependencies"`
	DevDependencies      map[string]string `json:"devDependencies"`
	OS                   npmStringList     `json:"os"`
	CPU                  npmStringList     `json:"cpu"`
	Dist                 struct {
		Tarball   string `json:"tarball"`
		Shasum    string `json:"shasum"`
		Integrity string `json:"integrity"`
	} `json:"dist"`
}

type pypiProjectResponse struct {
	Info struct {
		Name        string `json:"name"`
		Version     string `json:"version"`
		Summary     string `json:"summary"`
		HomePage    string `json:"home_page"`
		ProjectURL  string `json:"project_url"`
		Description string `json:"description"`
	} `json:"info"`
	Releases map[string][]pypiFile `json:"releases"`
	URLs     []pypiFile            `json:"urls"`
}

type pypiFile struct {
	Filename    string `json:"filename"`
	URL         string `json:"url"`
	PackageType string `json:"packagetype"`
	Digests     struct {
		SHA256 string `json:"sha256"`
	} `json:"digests"`
}

type cargoCrateResponse struct {
	Crate struct {
		ID            string `json:"id"`
		Name          string `json:"name"`
		MaxVersion    string `json:"max_version"`
		NewestVersion string `json:"newest_version"`
		Description   string `json:"description"`
		Homepage      string `json:"homepage"`
		Repository    string `json:"repository"`
	} `json:"crate"`
	Versions []struct {
		Num string `json:"num"`
	} `json:"versions"`
}

type cargoSearchResponse struct {
	Crates []struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		MaxVersion  string `json:"max_version"`
		Description string `json:"description"`
		Homepage    string `json:"homepage"`
		Repository  string `json:"repository"`
	} `json:"crates"`
}

type mavenMetadata struct {
	GroupID    string `xml:"groupId"`
	ArtifactID string `xml:"artifactId"`
	Versioning struct {
		Latest   string   `xml:"latest"`
		Release  string   `xml:"release"`
		Versions []string `xml:"versions>version"`
	} `xml:"versioning"`
}

type mavenSearchResponse struct {
	Response struct {
		Docs []struct {
			GroupID    string `json:"g"`
			ArtifactID string `json:"a"`
			Latest     string `json:"latestVersion"`
			Packaging  string `json:"p"`
			Timestamp  int64  `json:"timestamp"`
		} `json:"docs"`
	} `json:"response"`
}

type githubContentItem struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"`
	DownloadURL string `json:"download_url"`
}

type chocoFeed struct {
	Entries []chocoEntry `xml:"entry"`
}

type chocoEntry struct {
	Properties chocoProperties `xml:"properties"`
}

type chocoProperties struct {
	ID          string `xml:"Id"`
	Version     string `xml:"Version"`
	Title       string `xml:"Title"`
	Description string `xml:"Description"`
	Summary     string `xml:"Summary"`
	ProjectURL  string `xml:"ProjectUrl"`
}

type npmSearchResponse struct {
	Objects []struct {
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
	} `json:"objects"`
}

type githubCodeSearchResponse struct {
	Items []struct {
		Name string `json:"name"`
		Path string `json:"path"`
		URL  string `json:"html_url"`
	} `json:"items"`
}

type wingetRunSearchResponse struct {
	Packages []struct {
		ID     string `json:"Id"`
		Latest struct {
			Name        string   `json:"Name"`
			Publisher   string   `json:"Publisher"`
			Tags        []string `json:"Tags"`
			Description string   `json:"Description"`
			Homepage    string   `json:"Homepage"`
		} `json:"Latest"`
		Versions    []string `json:"Versions"`
		SearchScore float64  `json:"SearchScore"`
	} `json:"Packages"`
	Total int `json:"Total"`
}

func downloadOptionsFromRequest(req DownloadRequest) DownloadOptions {
	chunks := req.Chunks
	if chunks <= 0 {
		chunks = 1
	}
	return DownloadOptions{
		OutputDir:      req.OutputDir,
		Resume:         req.Resume,
		Chunks:         chunks,
		RequestOptions: req.RequestOptions,
	}
}

var (
	wingetRunSearchAPIURL     = "https://api.winget.run/v2/packages"
	wingetGitHubCodeSearchURL = "https://api.github.com/search/code"
)

var builtinMirrors = map[string][]Mirror{
	"npm": {
		{Source: "npm", Name: "huaweicloud", BaseURL: "https://repo.huaweicloud.com/repository/npm", ProbePath: "/react"},
		{Source: "npm", Name: "npmmirror", BaseURL: "https://registry.npmmirror.com", ProbePath: "/react"},
		{Source: "npm", Name: "tencent", BaseURL: "https://mirrors.cloud.tencent.com/npm", ProbePath: "/react"},
		{Source: "npm", Name: "npmjs", BaseURL: "https://registry.npmjs.org", ProbePath: "/react"},
	},
	"pip": {
		{Source: "pip", Name: "pypi", BaseURL: "https://pypi.org/pypi", ProbePath: "/pip/json"},
	},
	"cargo": {
		{Source: "cargo", Name: "crates.io", BaseURL: "https://crates.io/api/v1/crates", ProbePath: "/serde"},
	},
	"maven": {
		{Source: "maven", Name: "central", BaseURL: "https://repo1.maven.org/maven2", ProbePath: "/junit/junit/maven-metadata.xml"},
	},
	"choco": {
		{Source: "choco", Name: "chocolatey", BaseURL: "https://community.chocolatey.org/api/v2", ProbePath: "/Packages()?$top=1"},
		{Source: "choco", Name: "nuget", BaseURL: "https://www.nuget.org/api/v2", ProbePath: "/Packages()?$top=1"},
	},
	"winget": {
		{Source: "winget", Name: "github-api", BaseURL: "https://api.github.com", ProbePath: "/rate_limit"},
		{Source: "winget", Name: "github-raw", BaseURL: "https://raw.githubusercontent.com", ProbePath: "/microsoft/winget-pkgs/master/README.md"},
		{Source: "winget", Name: "jsdelivr", BaseURL: "https://cdn.jsdelivr.net/gh/microsoft/winget-pkgs@master", ProbePath: "/README.md"},
	},
}

func testMirrors(ctx context.Context, client *http.Client, source string) ([]MirrorResult, error) {
	sources := []string{"npm", "pip", "cargo", "maven", "choco", "winget"}
	if source != "" && source != "all" {
		sources = []string{source}
	}

	var mirrors []Mirror
	for _, item := range sources {
		current, ok := builtinMirrors[item]
		if !ok {
			return nil, fmt.Errorf("unsupported source %q", item)
		}
		mirrors = append(mirrors, current...)
	}

	// Test mirrors concurrently for better performance
	results := make([]MirrorResult, len(mirrors))
	var wg sync.WaitGroup
	
	for i, mirror := range mirrors {
		wg.Add(1)
		go func(idx int, m Mirror) {
			defer wg.Done()
			results[idx] = probeMirror(ctx, client, m)
		}(i, mirror)
	}
	
	wg.Wait()

	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Source != results[j].Source {
			return results[i].Source < results[j].Source
		}
		if results[i].OK != results[j].OK {
			return results[i].OK
		}
		return results[i].Total < results[j].Total
	})

	return results, nil
}

func searchPackages(ctx context.Context, client *http.Client, req SearchRequest) ([]SearchResult, error) {
	sources := []string{"npm", "pip", "cargo", "maven", "choco", "winget"}
	if req.Source != "" && req.Source != "all" {
		sources = []string{req.Source}
	}

	// Use concurrent search for better performance
	type searchResult struct {
		results []SearchResult
		err     error
	}
	
	resultChan := make(chan searchResult, len(sources))
	var wg sync.WaitGroup
	
	for _, source := range sources {
		wg.Add(1)
		go func(src string) {
			defer wg.Done()
			
			var found []SearchResult
			var err error
			
			switch src {
			case "npm":
				found, err = searchNPM(ctx, client, req)
			case "pip":
				found, err = searchPyPI(ctx, client, req)
			case "cargo":
				found, err = searchCargo(ctx, client, req)
			case "maven":
				found, err = searchMaven(ctx, client, req)
			case "choco":
				found, err = searchChoco(ctx, client, req)
			case "winget":
				found, err = searchWinget(ctx, client, req)
			default:
				err = fmt.Errorf("unsupported search source %q", src)
			}
			
			resultChan <- searchResult{results: found, err: err}
		}(source)
	}
	
	// Wait for all searches to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	
	// Collect results
	results := make([]SearchResult, 0, req.Limit*max(1, len(sources)))
	for res := range resultChan {
		if res.err != nil {
			// Log error but continue with other sources
			log.Printf("Search error: %v", res.err)
			continue
		}
		results = append(results, uniqueLatestSearchResults(res.results, req.Limit)...)
	}
	
	return results, nil
}

func probeMirror(ctx context.Context, client *http.Client, mirror Mirror) MirrorResult {
	start := time.Now()
	firstByte := time.Duration(0)
	trace := &httptrace.ClientTrace{
		GotFirstResponseByte: func() {
			firstByte = time.Since(start)
		},
	}

	url := joinURL(mirror.BaseURL, mirror.ProbePath)
	req, err := http.NewRequestWithContext(httptrace.WithClientTrace(ctx, trace), http.MethodGet, url, nil)
	if err != nil {
		return MirrorResult{Source: mirror.Source, Name: mirror.Name, BaseURL: mirror.BaseURL, Detail: err.Error()}
	}
	applyRequestHeaders(req, "", RequestOptions{})

	resp, err := client.Do(req)
	if err != nil {
		return MirrorResult{
			Source:    mirror.Source,
			Name:      mirror.Name,
			BaseURL:   mirror.BaseURL,
			FirstByte: firstByte,
			Total:     time.Since(start),
			Detail:    err.Error(),
		}
	}
	defer resp.Body.Close()

	_, _ = io.CopyN(io.Discard, resp.Body, 1024)
	total := time.Since(start)
	detail := fmt.Sprintf("http %d", resp.StatusCode)

	return MirrorResult{
		Source:     mirror.Source,
		Name:       mirror.Name,
		BaseURL:    mirror.BaseURL,
		StatusCode: resp.StatusCode,
		OK:         resp.StatusCode >= 200 && resp.StatusCode < 400,
		FirstByte:  firstByte,
		Total:      total,
		Detail:     detail,
	}
}

func searchNPM(ctx context.Context, client *http.Client, req SearchRequest) ([]SearchResult, error) {
	mirrors, err := resolveNPMSearchMirrors(req.Mirror)
	if err != nil {
		return nil, err
	}

	var lastErr error
	for _, mirror := range mirrors {
		searchURL := fmt.Sprintf(
			"%s/-/v1/search?text=%s&size=%d&from=0",
			strings.TrimRight(mirror.BaseURL, "/"),
			neturl.QueryEscape(req.Query),
			req.Limit,
		)
		var payload npmSearchResponse
		if err := getJSON(ctx, client, searchURL, req.RequestOptions, &payload); err != nil {
			req.RequestOptions.Verbose.Logf("resolve", "npm search %q via %s failed: %v", req.Query, mirror.Name, err)
			lastErr = fmt.Errorf("search npm via %s: %w", mirror.Name, err)
			continue
		}
		if len(payload.Objects) == 0 && strings.TrimSpace(req.Query) != "" {
			req.RequestOptions.Verbose.Logf("resolve", "npm search %q via %s returned no results", req.Query, mirror.Name)
			continue
		}

		results := make([]SearchResult, 0, len(payload.Objects))
		for _, item := range payload.Objects {
			pkg := item.Package
			detail := pkg.Links.NPM
			if strings.TrimSpace(detail) == "" {
				detail = pkg.Links.Homepage
			}
			if strings.TrimSpace(detail) == "" {
				detail = pkg.Links.Repository
			}
			results = append(results, SearchResult{
				Source:      "npm",
				MirrorName:  mirror.Name,
				Identifier:  pkg.Name,
				Version:     pkg.Version,
				Description: pkg.Description,
				Detail:      detail,
			})
		}
		req.RequestOptions.Verbose.Logf("resolve", "npm search %q resolved via %s with %d result(s)", req.Query, mirror.Name, len(results))
		return results, nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, nil
}

func resolveNPMSearchMirrors(requested string) ([]Mirror, error) {
	requested = strings.TrimSpace(requested)
	if requested == "" {
		return resolveMirrors("npm", "")
	}
	mirror, err := resolveMirror("npm", requested)
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(mirror.Name, "custom") {
		return []Mirror{mirror}, nil
	}
	mirrors := []Mirror{mirror}
	for _, candidate := range builtinMirrors["npm"] {
		if strings.EqualFold(candidate.Name, mirror.Name) {
			continue
		}
		mirrors = append(mirrors, candidate)
	}
	return mirrors, nil
}

func searchPyPI(ctx context.Context, client *http.Client, req SearchRequest) ([]SearchResult, error) {
	mirror, err := resolveMirror("pip", req.Mirror)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Query) == "" {
		return nil, nil
	}

	projectURL := joinURL(mirror.BaseURL, "/"+neturl.PathEscape(req.Query)+"/json")
	var payload pypiProjectResponse
	if err := getJSON(ctx, client, projectURL, req.RequestOptions, &payload); err != nil {
		return nil, fmt.Errorf("search pip exact package: %w", err)
	}
	name := strings.TrimSpace(payload.Info.Name)
	if name == "" {
		name = strings.TrimSpace(req.Query)
	}
	version := strings.TrimSpace(payload.Info.Version)
	detail := firstNonEmpty(payload.Info.ProjectURL, payload.Info.HomePage)
	return []SearchResult{{
		Source:      "pip",
		MirrorName:  mirror.Name,
		Identifier:  name,
		Version:     version,
		Description: payload.Info.Summary,
		Detail:      detail,
	}}, nil
}

func searchCargo(ctx context.Context, client *http.Client, req SearchRequest) ([]SearchResult, error) {
	mirror, err := resolveMirror("cargo", req.Mirror)
	if err != nil {
		return nil, err
	}

	searchURL := fmt.Sprintf(
		"%s?q=%s&per_page=%d",
		strings.TrimRight(mirror.BaseURL, "/"),
		neturl.QueryEscape(req.Query),
		req.Limit,
	)
	var payload cargoSearchResponse
	if err := getJSON(ctx, client, searchURL, req.RequestOptions, &payload); err != nil {
		return nil, fmt.Errorf("search cargo: %w", err)
	}

	results := make([]SearchResult, 0, len(payload.Crates))
	for _, item := range payload.Crates {
		detail := firstNonEmpty(item.Homepage, item.Repository)
		identifier := firstNonEmpty(item.ID, item.Name)
		results = append(results, SearchResult{
			Source:      "cargo",
			MirrorName:  mirror.Name,
			Identifier:  identifier,
			Version:     item.MaxVersion,
			Description: item.Description,
			Detail:      detail,
		})
	}
	return results, nil
}

func searchMaven(ctx context.Context, client *http.Client, req SearchRequest) ([]SearchResult, error) {
	searchURL := fmt.Sprintf(
		"https://search.maven.org/solrsearch/select?q=%s&rows=%d&wt=json",
		neturl.QueryEscape(req.Query),
		req.Limit,
	)
	var payload mavenSearchResponse
	if err := getJSON(ctx, client, searchURL, req.RequestOptions, &payload); err != nil {
		return nil, fmt.Errorf("search maven: %w", err)
	}

	results := make([]SearchResult, 0, len(payload.Response.Docs))
	for _, item := range payload.Response.Docs {
		identifier := mavenCoordinate(item.GroupID, item.ArtifactID)
		results = append(results, SearchResult{
			Source:      "maven",
			Identifier:  identifier,
			Version:     item.Latest,
			Description: item.Packaging,
			Detail:      "https://search.maven.org/artifact/" + neturl.PathEscape(item.GroupID) + "/" + neturl.PathEscape(item.ArtifactID),
		})
	}
	return results, nil
}

func searchChoco(ctx context.Context, client *http.Client, req SearchRequest) ([]SearchResult, error) {
	mirror, err := resolveMirror("choco", req.Mirror)
	if err != nil {
		return nil, err
	}

	searchURL := fmt.Sprintf(
		"%s/Search()?searchTerm='%s'&targetFramework=''&includePrerelease=false&$top=%d",
		strings.TrimRight(mirror.BaseURL, "/"),
		neturl.QueryEscape(req.Query),
		req.Limit,
	)
	body, err := getBytesWithAccept(ctx, client, searchURL, "application/atom+xml, application/xml", req.RequestOptions)
	if err != nil {
		return nil, fmt.Errorf("search choco: %w", err)
	}

	var feed chocoFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("parse choco search feed: %w", err)
	}

	results := make([]SearchResult, 0, len(feed.Entries))
	for _, entry := range feed.Entries {
		props := entry.Properties
		description := strings.TrimSpace(props.Summary)
		if description == "" {
			description = strings.TrimSpace(props.Description)
		}
		identifier := strings.TrimSpace(props.ID)
		if identifier == "" {
			identifier = strings.TrimSpace(props.Title)
		}
		results = append(results, SearchResult{
			Source:      "choco",
			MirrorName:  mirror.Name,
			Identifier:  identifier,
			Version:     props.Version,
			Description: description,
			Detail:      props.ProjectURL,
		})
	}
	return results, nil
}

func searchWinget(ctx context.Context, client *http.Client, req SearchRequest) ([]SearchResult, error) {
	searchURL := fmt.Sprintf(
		"%s?query=%s&take=%d",
		strings.TrimRight(wingetRunSearchAPIURL, "/"),
		neturl.QueryEscape(req.Query),
		req.Limit,
	)

	var payload wingetRunSearchResponse
	primaryErr := getJSON(ctx, client, searchURL, req.RequestOptions, &payload)
	if primaryErr == nil {
		results := make([]SearchResult, 0, len(payload.Packages))
		for _, item := range payload.Packages {
			description := strings.TrimSpace(item.Latest.Description)
			if description == "" {
				description = strings.TrimSpace(item.Latest.Publisher)
			}
			version := ""
			if len(item.Versions) > 0 {
				version = item.Versions[0]
			}
			results = append(results, SearchResult{
				Source:      "winget",
				MirrorName:  "github-api",
				Identifier:  item.ID,
				Version:     version,
				Description: description,
				Detail:      item.Latest.Homepage,
			})
		}
		return results, nil
	}

	results, err := searchWingetViaGitHub(ctx, client, req)
	if err == nil {
		return results, nil
	}
	return nil, fmt.Errorf("search winget: primary=%v; fallback=%w", primaryErr, err)
}

func searchWingetViaGitHub(ctx context.Context, client *http.Client, req SearchRequest) ([]SearchResult, error) {
	searchURL := fmt.Sprintf(
		"%s?q=%s+repo:microsoft/winget-pkgs+path:manifests&per_page=%d",
		strings.TrimRight(wingetGitHubCodeSearchURL, "/"),
		neturl.QueryEscape(req.Query),
		req.Limit*4,
	)

	var payload githubCodeSearchResponse
	if err := getJSON(ctx, client, searchURL, req.RequestOptions, &payload); err != nil {
		return nil, fmt.Errorf("search winget: %w", err)
	}

	merged := map[string]SearchResult{}
	for _, item := range payload.Items {
		identifier, version, ok := wingetSearchInfoFromPath(item.Path)
		if !ok {
			continue
		}
		current, exists := merged[identifier]
		next := SearchResult{
			Source:      "winget",
			MirrorName:  "github-api",
			Identifier:  identifier,
			Version:     version,
			Description: item.Name,
			Detail:      item.URL,
		}
		if !exists || compareVersionish(next.Version, current.Version) > 0 {
			merged[identifier] = next
		}
	}

	results := make([]SearchResult, 0, len(merged))
	for _, result := range merged {
		results = append(results, result)
	}
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Identifier == results[j].Identifier {
			return compareVersionish(results[i].Version, results[j].Version) > 0
		}
		return strings.ToLower(results[i].Identifier) < strings.ToLower(results[j].Identifier)
	})
	if len(results) > req.Limit {
		results = results[:req.Limit]
	}
	return results, nil
}

func resolveDownloadPlan(ctx context.Context, client *http.Client, req DownloadRequest) (DownloadPlan, error) {
	switch req.Source {
	case "npm":
		return resolveNPM(ctx, client, req)
	case "pip":
		return resolvePyPI(ctx, client, req)
	case "cargo":
		return resolveCargo(ctx, client, req)
	case "maven":
		return resolveMaven(ctx, client, req)
	case "choco":
		return resolveChoco(ctx, client, req)
	case "winget":
		return resolveWinget(ctx, client, req)
	case "url":
		return resolveDirectURL(req)
	default:
		return DownloadPlan{}, fmt.Errorf("unsupported source %q", req.Source)
	}
}

func resolveDirectURL(req DownloadRequest) (DownloadPlan, error) {
	if req.URL == "" {
		return DownloadPlan{}, errors.New("--url is required when --source url")
	}
	return DownloadPlan{
		Source:     "url",
		Identifier: req.URL,
		URL:        req.URL,
		Filename:   filenameFromURLOrFallback(req.URL, "download.bin"),
	}, nil
}

func resolveNPM(ctx context.Context, client *http.Client, req DownloadRequest) (DownloadPlan, error) {
	if req.Name == "" {
		return DownloadPlan{}, errors.New("--name is required when --source npm")
	}

	mirror, metadata, err := fetchNPMMetadataWithFallback(ctx, client, req.Mirror, req.Name, req.RequestOptions)
	if err != nil {
		return DownloadPlan{}, err
	}

	resolvedVersion, details, err := pickNPMVersion(metadata, req.Version)
	if err != nil {
		return DownloadPlan{}, err
	}
	if strings.TrimSpace(details.Dist.Tarball) == "" {
		return DownloadPlan{}, fmt.Errorf("npm version %s has no tarball url", resolvedVersion)
	}

	tarballURL := rewriteMirrorURL(details.Dist.Tarball, mirror.BaseURL)
	return DownloadPlan{
		Source:     "npm",
		Identifier: req.Name,
		Version:    resolvedVersion,
		URL:        tarballURL,
		Filename:   filenameFromURLOrFallback(tarballURL, sanitizeFileName(req.Name+"-"+resolvedVersion+".tgz")),
		MirrorName: mirror.Name,
		Integrity:  details.Dist.Integrity,
		Shasum:     details.Dist.Shasum,
	}, nil
}

func resolvePyPI(ctx context.Context, client *http.Client, req DownloadRequest) (DownloadPlan, error) {
	if strings.TrimSpace(req.Name) == "" {
		return DownloadPlan{}, errors.New("--name is required when --source pip")
	}
	mirror, err := resolveMirror("pip", req.Mirror)
	if err != nil {
		return DownloadPlan{}, err
	}

	projectURL := joinURL(mirror.BaseURL, "/"+neturl.PathEscape(req.Name)+"/json")
	var metadata pypiProjectResponse
	if err := getJSON(ctx, client, projectURL, req.RequestOptions, &metadata); err != nil {
		return DownloadPlan{}, fmt.Errorf("fetch pip metadata: %w", err)
	}
	version := strings.TrimSpace(req.Version)
	if version == "" || strings.EqualFold(version, "latest") {
		version = strings.TrimSpace(metadata.Info.Version)
	}
	if version == "" {
		return DownloadPlan{}, fmt.Errorf("pip package %q has no resolvable version", req.Name)
	}
	files := metadata.Releases[version]
	if len(files) == 0 && strings.EqualFold(version, metadata.Info.Version) {
		files = metadata.URLs
	}
	file, ok := selectPyPIFile(files)
	if !ok {
		return DownloadPlan{}, fmt.Errorf("pip package %s@%s has no downloadable distribution", req.Name, version)
	}
	return DownloadPlan{
		Source:     "pip",
		Identifier: firstNonEmpty(metadata.Info.Name, req.Name),
		Version:    version,
		URL:        file.URL,
		Filename:   filenameFromURLOrFallback(file.URL, file.Filename),
		MirrorName: mirror.Name,
		Integrity:  formatDigestIntegrity("sha256", strings.TrimSpace(file.Digests.SHA256)),
	}, nil
}

func formatDigestIntegrity(algorithm string, hexDigest string) string {
	algorithm = strings.ToLower(strings.TrimSpace(algorithm))
	hexDigest = strings.TrimSpace(hexDigest)
	if algorithm == "" || hexDigest == "" {
		return ""
	}
	raw, err := hex.DecodeString(hexDigest)
	if err != nil {
		return ""
	}
	return algorithm + "-" + base64.StdEncoding.EncodeToString(raw)
}

func selectPyPIFile(files []pypiFile) (pypiFile, bool) {
	if len(files) == 0 {
		return pypiFile{}, false
	}
	for _, file := range files {
		if strings.EqualFold(file.PackageType, "sdist") && strings.TrimSpace(file.URL) != "" {
			return file, true
		}
	}
	for _, file := range files {
		if strings.EqualFold(file.PackageType, "bdist_wheel") && strings.TrimSpace(file.URL) != "" {
			return file, true
		}
	}
	for _, file := range files {
		if strings.TrimSpace(file.URL) != "" {
			return file, true
		}
	}
	return pypiFile{}, false
}

func resolveCargo(ctx context.Context, client *http.Client, req DownloadRequest) (DownloadPlan, error) {
	if strings.TrimSpace(req.Name) == "" {
		return DownloadPlan{}, errors.New("--name is required when --source cargo")
	}
	mirror, err := resolveMirror("cargo", req.Mirror)
	if err != nil {
		return DownloadPlan{}, err
	}

	metadataURL := joinURL(mirror.BaseURL, "/"+neturl.PathEscape(req.Name))
	var metadata cargoCrateResponse
	if err := getJSON(ctx, client, metadataURL, req.RequestOptions, &metadata); err != nil {
		return DownloadPlan{}, fmt.Errorf("fetch cargo metadata: %w", err)
	}
	version := strings.TrimSpace(req.Version)
	if version == "" || strings.EqualFold(version, "latest") {
		version = firstNonEmpty(metadata.Crate.MaxVersion, metadata.Crate.NewestVersion)
	}
	if version == "" {
		for _, item := range metadata.Versions {
			if item.Num != "" && (version == "" || compareVersionish(item.Num, version) > 0) {
				version = item.Num
			}
		}
	}
	if version == "" {
		return DownloadPlan{}, fmt.Errorf("cargo crate %q has no resolvable version", req.Name)
	}
	identifier := firstNonEmpty(metadata.Crate.ID, metadata.Crate.Name, req.Name)
	downloadURL := joinURL(mirror.BaseURL, "/"+neturl.PathEscape(identifier)+"/"+neturl.PathEscape(version)+"/download")
	return DownloadPlan{
		Source:     "cargo",
		Identifier: identifier,
		Version:    version,
		URL:        downloadURL,
		Filename:   sanitizeFileName(identifier + "-" + version + ".crate"),
		MirrorName: mirror.Name,
	}, nil
}

func resolveMaven(ctx context.Context, client *http.Client, req DownloadRequest) (DownloadPlan, error) {
	groupID, artifactID, err := parseMavenCoordinate(firstNonEmpty(req.Name, req.PackageID))
	if err != nil {
		return DownloadPlan{}, err
	}
	mirror, err := resolveMirror("maven", req.Mirror)
	if err != nil {
		return DownloadPlan{}, err
	}

	version := strings.TrimSpace(req.Version)
	artifactPath := strings.ReplaceAll(groupID, ".", "/") + "/" + artifactID
	if version == "" || strings.EqualFold(version, "latest") || strings.EqualFold(version, "release") {
		metadataURL := joinURL(mirror.BaseURL, "/"+artifactPath+"/maven-metadata.xml")
		body, err := getBytesWithAccept(ctx, client, metadataURL, "application/xml, text/xml", req.RequestOptions)
		if err != nil {
			return DownloadPlan{}, fmt.Errorf("fetch maven metadata: %w", err)
		}
		var metadata mavenMetadata
		if err := xml.Unmarshal(body, &metadata); err != nil {
			return DownloadPlan{}, fmt.Errorf("parse maven metadata: %w", err)
		}
		version = firstNonEmpty(metadata.Versioning.Release, metadata.Versioning.Latest)
		if version == "" {
			for _, item := range metadata.Versioning.Versions {
				if item != "" && (version == "" || compareVersionish(item, version) > 0) {
					version = item
				}
			}
		}
	}
	if version == "" {
		return DownloadPlan{}, fmt.Errorf("maven artifact %s has no resolvable version", mavenCoordinate(groupID, artifactID))
	}
	filename := artifactID + "-" + version + ".jar"
	return DownloadPlan{
		Source:     "maven",
		Identifier: mavenCoordinate(groupID, artifactID),
		Version:    version,
		URL:        joinURL(mirror.BaseURL, "/"+artifactPath+"/"+version+"/"+filename),
		Filename:   sanitizeFileName(filename),
		MirrorName: mirror.Name,
	}, nil
}

func pickNPMVersion(metadata npmPackageMetadata, requested string) (string, npmVersionDetails, error) {
	if len(metadata.Versions) == 0 {
		return "", npmVersionDetails{}, fmt.Errorf("package %s has no published versions", metadata.Name)
	}

	requested = strings.TrimSpace(requested)
	if requested == "" || strings.EqualFold(requested, "latest") {
		if latest := strings.TrimSpace(metadata.DistTags["latest"]); latest != "" {
			requested = latest
		}
	}
	if tagVersion, ok := metadata.DistTags[requested]; ok && strings.TrimSpace(tagVersion) != "" {
		requested = tagVersion
	}

	details, ok := metadata.Versions[requested]
	if !ok {
		if version, rangedDetails, matched := pickNPMVersionFromRange(metadata, requested); matched {
			return version, rangedDetails, nil
		}
		return "", npmVersionDetails{}, fmt.Errorf("npm version/range %q not found", requested)
	}
	return requested, details, nil
}

func resolveChoco(ctx context.Context, client *http.Client, req DownloadRequest) (DownloadPlan, error) {
	if req.Name == "" {
		return DownloadPlan{}, errors.New("--name is required when --source choco")
	}

	mirror, err := resolveMirror("choco", req.Mirror)
	if err != nil {
		return DownloadPlan{}, err
	}

	version := strings.TrimSpace(req.Version)
	if version == "" || strings.EqualFold(version, "latest") {
		version, err = resolveLatestChocoVersion(ctx, client, mirror.BaseURL, req.Name, req.RequestOptions)
		if err != nil {
			return DownloadPlan{}, err
		}
	}

	packageURL := joinURL(mirror.BaseURL, "/package/"+neturl.PathEscape(req.Name)+"/"+neturl.PathEscape(version))
	return DownloadPlan{
		Source:     "choco",
		Identifier: req.Name,
		Version:    version,
		URL:        packageURL,
		Filename:   sanitizeFileName(req.Name + "." + version + ".nupkg"),
		MirrorName: mirror.Name,
	}, nil
}

func resolveLatestChocoVersion(ctx context.Context, client *http.Client, baseURL string, packageName string, options RequestOptions) (string, error) {
	feedURL := joinURL(baseURL, "/FindPackagesById()?id='"+neturl.QueryEscape(packageName)+"'")
	body, err := getBytesWithAccept(ctx, client, feedURL, "application/atom+xml, application/xml", options)
	if err != nil {
		return "", fmt.Errorf("resolve latest choco version: %w", err)
	}

	var feed chocoFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return "", fmt.Errorf("parse choco feed: %w", err)
	}
	if len(feed.Entries) == 0 {
		return "", fmt.Errorf("package %q not found in %s", packageName, baseURL)
	}

	version := strings.TrimSpace(feed.Entries[0].Properties.Version)
	for _, entry := range feed.Entries[1:] {
		current := strings.TrimSpace(entry.Properties.Version)
		if compareVersionish(current, version) > 0 {
			version = current
		}
	}
	if version == "" {
		return "", fmt.Errorf("package %q has no resolvable version", packageName)
	}
	return version, nil
}

func resolveWinget(ctx context.Context, client *http.Client, req DownloadRequest) (DownloadPlan, error) {
	if req.PackageID == "" {
		return DownloadPlan{}, errors.New("--id is required when --source winget")
	}

	rawMirror, err := resolveWingetRawMirror(req.Mirror)
	if err != nil {
		return DownloadPlan{}, err
	}

	packagePath, err := wingetPackagePath(req.PackageID)
	if err != nil {
		return DownloadPlan{}, err
	}

	versionDir, version, err := resolveWingetVersionDir(ctx, client, packagePath, req.Version, req.RequestOptions)
	if err != nil {
		return DownloadPlan{}, err
	}

	manifestItem, err := resolveWingetManifestItem(ctx, client, versionDir, req.RequestOptions)
	if err != nil {
		return DownloadPlan{}, err
	}
	if manifestItem.DownloadURL == "" {
		return DownloadPlan{}, fmt.Errorf("winget manifest download url is missing for %s", manifestItem.Path)
	}

	manifestURL, mirrorName, body, err := fetchWingetManifestWithFallback(ctx, client, rawMirror, manifestItem.DownloadURL, req.RequestOptions)
	if err != nil {
		return DownloadPlan{}, err
	}

	var manifest wingetInstallerManifest
	if err := yaml.Unmarshal(body, &manifest); err != nil {
		return DownloadPlan{}, fmt.Errorf("parse winget manifest: %w", err)
	}
	if len(manifest.Installers) == 0 {
		return DownloadPlan{}, fmt.Errorf("winget manifest %s does not contain installers", manifestURL)
	}

	index, installer, err := selectWingetInstaller(manifest.Installers, req.Arch, req.InstallerIndex)
	if err != nil {
		return DownloadPlan{}, err
	}

	planVersion := strings.TrimSpace(version)
	if strings.TrimSpace(manifest.PackageVersion) != "" {
		planVersion = strings.TrimSpace(manifest.PackageVersion)
	}

	return DownloadPlan{
		Source:      "winget",
		Identifier:  req.PackageID,
		Version:     planVersion,
		URL:         installer.InstallerURL,
		Filename:    filenameFromURLOrFallback(installer.InstallerURL, sanitizeFileName(req.PackageID+"-"+planVersion+"-"+installer.Architecture)),
		MirrorName:  mirrorName,
		ManifestURL: manifestURL,
		Installers:  manifest.Installers,
	}.withSelectedIndex(index), nil
}

func (plan DownloadPlan) withSelectedIndex(index int) DownloadPlan {
	if index < 0 || index >= len(plan.Installers) {
		return plan
	}
	installer := plan.Installers[index]
	if installer.InstallerURL != "" {
		plan.URL = installer.InstallerURL
	}
	if installer.Architecture != "" {
		plan.Filename = filenameFromURLOrFallback(plan.URL, sanitizeFileName(plan.Identifier+"-"+plan.Version+"-"+installer.Architecture))
	}
	return plan
}

func resolveWingetVersionDir(ctx context.Context, client *http.Client, packagePath string, requested string, options RequestOptions) (string, string, error) {
	items, err := getGitHubContents(ctx, client, packagePath, options)
	if err != nil {
		return "", "", fmt.Errorf("list winget package versions: %w", err)
	}

	requested = strings.TrimSpace(requested)
	if requested != "" && !strings.EqualFold(requested, "latest") {
		for _, item := range items {
			if item.Type == "dir" && strings.EqualFold(item.Name, requested) {
				return item.Path, item.Name, nil
			}
		}
		return "", "", fmt.Errorf("winget version %q not found for %s", requested, packagePath)
	}

	best := githubContentItem{}
	for _, item := range items {
		if item.Type != "dir" {
			continue
		}
		if best.Name == "" || compareVersionish(item.Name, best.Name) > 0 {
			best = item
		}
	}
	if best.Name == "" {
		return "", "", fmt.Errorf("no winget version directory found for %s", packagePath)
	}
	return best.Path, best.Name, nil
}

func resolveWingetManifestItem(ctx context.Context, client *http.Client, versionDir string, options RequestOptions) (githubContentItem, error) {
	items, err := getGitHubContents(ctx, client, versionDir, options)
	if err != nil {
		return githubContentItem{}, err
	}

	best := githubContentItem{}
	for _, item := range items {
		if item.Type != "file" || !strings.HasSuffix(strings.ToLower(item.Name), ".yaml") {
			continue
		}
		if strings.HasSuffix(strings.ToLower(item.Name), ".installer.yaml") {
			return item, nil
		}
		if best.Name == "" {
			best = item
		}
	}
	if best.Name == "" {
		return githubContentItem{}, fmt.Errorf("no yaml manifest found in %s", versionDir)
	}
	return best, nil
}

func getGitHubContents(ctx context.Context, client *http.Client, repoPath string, options RequestOptions) ([]githubContentItem, error) {
	url := joinURL("https://api.github.com", "/repos/microsoft/winget-pkgs/contents/"+strings.TrimLeft(repoPath, "/"))
	body, err := getBytes(ctx, client, url, options)
	if err != nil {
		return nil, err
	}

	var items []githubContentItem
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, fmt.Errorf("parse github contents response: %w", err)
	}
	return items, nil
}

func resolveMirror(source string, requested string) (Mirror, error) {
	requested = strings.TrimSpace(requested)
	mirrors := builtinMirrors[source]
	if len(mirrors) == 0 {
		return Mirror{}, fmt.Errorf("no mirrors configured for %s", source)
	}

	if requested == "" {
		return mirrors[0], nil
	}
	if strings.HasPrefix(strings.ToLower(requested), "http://") || strings.HasPrefix(strings.ToLower(requested), "https://") {
		return Mirror{Source: source, Name: "custom", BaseURL: strings.TrimRight(requested, "/")}, nil
	}

	for _, mirror := range mirrors {
		if strings.EqualFold(mirror.Name, requested) {
			return mirror, nil
		}
	}
	return Mirror{}, fmt.Errorf("unknown %s mirror %q", source, requested)
}

func resolveMirrors(source string, requested string) ([]Mirror, error) {
	requested = strings.TrimSpace(requested)
	mirrors := builtinMirrors[source]
	if len(mirrors) == 0 {
		return nil, fmt.Errorf("no mirrors configured for %s", source)
	}
	if requested == "" {
		copied := make([]Mirror, len(mirrors))
		copy(copied, mirrors)
		return copied, nil
	}
	mirror, err := resolveMirror(source, requested)
	if err != nil {
		return nil, err
	}
	return []Mirror{mirror}, nil
}

func fetchNPMMetadataWithFallback(ctx context.Context, client *http.Client, requestedMirror string, packageName string, options RequestOptions) (Mirror, npmPackageMetadata, error) {
	mirrors, err := resolveMirrors("npm", requestedMirror)
	if err != nil {
		return Mirror{}, npmPackageMetadata{}, err
	}
	var lastErr error
	for _, mirror := range mirrors {
		metadataURL := joinURL(mirror.BaseURL, "/"+neturl.PathEscape(strings.TrimSpace(packageName)))
		var metadata npmPackageMetadata
		if err := getJSON(ctx, client, metadataURL, options, &metadata); err != nil {
			options.Verbose.Logf("resolve", "npm metadata %s via %s failed: %v", packageName, mirror.Name, err)
			lastErr = fmt.Errorf("fetch npm metadata for %s via %s: %w", packageName, mirror.Name, err)
			continue
		}
		options.Verbose.Logf("resolve", "npm metadata %s resolved via %s", packageName, mirror.Name)
		return mirror, metadata, nil
	}
	if lastErr != nil {
		return Mirror{}, npmPackageMetadata{}, lastErr
	}
	return Mirror{}, npmPackageMetadata{}, fmt.Errorf("fetch npm metadata for %s: no mirrors available", packageName)
}

func resolveWingetRawMirror(requested string) (Mirror, error) {
	if strings.TrimSpace(requested) == "" {
		return builtinMirrors["winget"][1], nil
	}
	if strings.EqualFold(strings.TrimSpace(requested), "github-api") {
		return Mirror{}, errors.New("winget download must use a raw manifest mirror, not github-api")
	}
	return resolveMirror("winget", requested)
}

func wingetSearchInfoFromPath(itemPath string) (string, string, bool) {
	parts := strings.Split(strings.Trim(strings.ReplaceAll(itemPath, "\\", "/"), "/"), "/")
	if len(parts) < 6 || parts[0] != "manifests" {
		return "", "", false
	}
	version := parts[len(parts)-2]
	if version == "" {
		return "", "", false
	}
	identifier := strings.Join(parts[2:len(parts)-2], ".")
	if identifier == "" {
		return "", "", false
	}
	return identifier, version, true
}

func fetchWingetManifestWithFallback(ctx context.Context, client *http.Client, primary Mirror, rawDownloadURL string, options RequestOptions) (string, string, []byte, error) {
	candidates := []Mirror{primary}
	for _, mirror := range builtinMirrors["winget"] {
		if strings.EqualFold(mirror.Name, "github-api") {
			continue
		}
		if strings.EqualFold(mirror.Name, primary.Name) {
			continue
		}
		candidates = append(candidates, mirror)
	}

	var lastErr error
	for _, mirror := range candidates {
		manifestURL := rewriteMirrorURL(rawDownloadURL, mirror.BaseURL)
		body, err := getBytes(ctx, client, manifestURL, options)
		if err == nil {
			return manifestURL, mirror.Name, body, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = errors.New("no winget manifest mirrors available")
	}
	return "", "", nil, fmt.Errorf("download winget manifest: %w", lastErr)
}

func wingetPackagePath(packageID string) (string, error) {
	packageID = strings.TrimSpace(packageID)
	if packageID == "" {
		return "", errors.New("winget package id is empty")
	}

	first := strings.ToLower(string([]rune(packageID)[0]))
	parts := strings.Split(packageID, ".")
	return path.Join(append([]string{"manifests", first}, parts...)...), nil
}

func selectWingetInstaller(installers []WingetInstaller, requestedArch string, requestedIndex int) (int, WingetInstaller, error) {
	if len(installers) == 0 {
		return -1, WingetInstaller{}, errors.New("winget manifest has no installers")
	}
	if requestedIndex >= 0 {
		if requestedIndex >= len(installers) {
			return -1, WingetInstaller{}, fmt.Errorf("installer index %d out of range", requestedIndex)
		}
		return requestedIndex, installers[requestedIndex], nil
	}
	if requestedArch != "" {
		for index, installer := range installers {
			if strings.EqualFold(installer.Architecture, requestedArch) {
				return index, installer, nil
			}
		}
		return -1, WingetInstaller{}, fmt.Errorf("no installer matches arch %q", requestedArch)
	}

	bestIndex := 0
	bestScore := -1
	for index, installer := range installers {
		score := 0
		switch strings.ToLower(strings.TrimSpace(installer.Architecture)) {
		case "x64":
			score += 5
		case "arm64":
			score += 4
		case "neutral":
			score += 3
		case "x86":
			score += 2
		default:
			score += 1
		}
		if strings.EqualFold(strings.TrimSpace(installer.Scope), "machine") {
			score++
		}
		if score > bestScore {
			bestScore = score
			bestIndex = index
		}
	}
	return bestIndex, installers[bestIndex], nil
}

type remoteDownloadInfo struct {
	Size           int64
	SupportsRanges bool
}

func downloadPlan(ctx context.Context, client *http.Client, plan DownloadPlan, options DownloadOptions) (DownloadResult, error) {
	return downloadPlanWithProgress(ctx, client, plan, options, os.Stderr)
}

func downloadPlanWithProgress(ctx context.Context, client *http.Client, plan DownloadPlan, options DownloadOptions, progressOut io.Writer) (DownloadResult, error) {
	return downloadPlanWithProgressCallback(ctx, client, plan, options, progressOut, nil)
}

func downloadPlanWithProgressCallback(ctx context.Context, client *http.Client, plan DownloadPlan, options DownloadOptions, progressOut io.Writer, onProgress func(DownloadProgress)) (DownloadResult, error) {
	if strings.TrimSpace(plan.URL) == "" {
		return DownloadResult{}, errors.New("resolved download url is empty")
	}
	if strings.TrimSpace(options.OutputDir) == "" {
		options.OutputDir = "."
	}
	if options.Chunks <= 0 {
		options.Chunks = 1
	}
	if err := os.MkdirAll(options.OutputDir, 0o755); err != nil {
		return DownloadResult{}, fmt.Errorf("create output dir: %w", err)
	}

	filename := sanitizeFileName(strings.TrimSpace(plan.Filename))
	if filename == "" {
		filename = filenameFromURLOrFallback(plan.URL, "download.bin")
	}
	targetPath, err := ensureUniquePath(filepath.Join(options.OutputDir, filename))
	if err != nil {
		return DownloadResult{}, err
	}
	tempPath := targetPath + ".part"
	info, err := probeRemoteDownload(ctx, client, plan.URL, options.RequestOptions)
	if err != nil {
		info = remoteDownloadInfo{}
	}

	start := time.Now()
	initialWritten := int64(0)
	if options.Resume {
		if stat, statErr := os.Stat(tempPath); statErr == nil {
			initialWritten = stat.Size()
		}
	}
	progress := newDownloadProgressWriter(progressOut, filepath.Base(targetPath), info.Size, start, initialWritten, onProgress)
	defer progress.finish()

	if info.Size > 0 && initialWritten >= info.Size {
		if err := resetTempDownload(tempPath, progress, info.Size); err != nil {
			return DownloadResult{}, err
		}
		initialWritten = 0
	}

	switch {
	case options.Chunks > 1 && info.SupportsRanges && info.Size > 0 && initialWritten == 0:
		if err := downloadChunked(ctx, client, plan.URL, tempPath, info.Size, options, progress); err != nil {
			cleanupPartialDownload(tempPath, options.Resume)
			return DownloadResult{}, err
		}
	default:
		if err := downloadSingleStream(ctx, client, plan.URL, tempPath, info, options, initialWritten, progress); err != nil {
			cleanupPartialDownload(tempPath, options.Resume)
			return DownloadResult{}, err
		}
	}

	digests, err := verifyDownloadedFile(tempPath, plan)
	if err != nil {
		options.RequestOptions.Verbose.Logf("integrity", "download %s failed verification: %v", downloadPlanLabel(plan, targetPath), err)
		_ = os.Remove(tempPath)
		return DownloadResult{}, err
	}
	options.RequestOptions.Verbose.Logf("integrity", "verified download %s using %s", downloadPlanLabel(plan, targetPath), integrityCheckSummary(plan))
	if err := syncFileBeforeRename(tempPath); err != nil {
		_ = os.Remove(tempPath)
		return DownloadResult{}, err
	}
	if err := os.Rename(tempPath, targetPath); err != nil {
		_ = os.Remove(tempPath)
		return DownloadResult{}, fmt.Errorf("rename temp file: %w", err)
	}
	return DownloadResult{
		Path:     targetPath,
		Size:     digests.Size,
		SHA256:   digests.SHA256,
		Duration: time.Since(start),
	}, nil
}

func downloadPlanLabel(plan DownloadPlan, fallbackPath string) string {
	identifier := strings.TrimSpace(plan.Identifier)
	if identifier == "" {
		identifier = filepath.Base(strings.TrimSpace(fallbackPath))
	}
	version := strings.TrimSpace(plan.Version)
	if version == "" {
		return identifier
	}
	return identifier + "@" + version
}

func integrityCheckSummary(plan DownloadPlan) string {
	checks := make([]string, 0, 2)
	if strings.TrimSpace(plan.Shasum) != "" {
		checks = append(checks, "shasum")
	}
	if strings.TrimSpace(plan.Integrity) != "" {
		checks = append(checks, "integrity")
	}
	if len(checks) == 0 {
		return "sha256 digest"
	}
	return strings.Join(checks, " + ")
}

func resetTempDownload(tempPath string, progress *downloadProgressWriter, total int64) error {
	if err := os.Remove(tempPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("reset temp file: %w", err)
	}
	progress.setWritten(0, total)
	return nil
}

func cleanupPartialDownload(tempPath string, keep bool) {
	if keep {
		return
	}
	_ = os.Remove(tempPath)
}

func downloadSingleStream(ctx context.Context, client *http.Client, url string, tempPath string, info remoteDownloadInfo, options DownloadOptions, resumeOffset int64, progress *downloadProgressWriter) error {
	flags := os.O_CREATE | os.O_WRONLY
	if resumeOffset > 0 && options.Resume && info.SupportsRanges {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
		resumeOffset = 0
		progress.setWritten(0, info.Size)
	}
	file, err := os.OpenFile(tempPath, flags, 0o644)
	if err != nil {
		return fmt.Errorf("open temp file: %w", err)
	}
	defer file.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create download request: %w", err)
	}
	if resumeOffset > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", resumeOffset))
	}
	applyRequestHeaders(req, "", options.RequestOptions)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()

	switch {
	case resumeOffset > 0 && resp.StatusCode != http.StatusPartialContent:
		return fmt.Errorf("resume download failed with http %d", resp.StatusCode)
	case resumeOffset == 0 && (resp.StatusCode < 200 || resp.StatusCode >= 400):
		message, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("download failed with http %d: %s", resp.StatusCode, strings.TrimSpace(string(message)))
	}

	writer := io.MultiWriter(file, progress)
	if _, err := io.Copy(writer, resp.Body); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

func downloadChunked(ctx context.Context, client *http.Client, url string, tempPath string, size int64, options DownloadOptions, progress *downloadProgressWriter) error {
	file, err := os.OpenFile(tempPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	if err := file.Truncate(size); err != nil {
		_ = file.Close()
		return fmt.Errorf("allocate temp file: %w", err)
	}
	defer file.Close()

	chunks := options.Chunks
	if chunks < 2 {
		chunks = 2
	}
	const minChunkSize = int64(1 << 20)
	maxChunks := int((size + minChunkSize - 1) / minChunkSize)
	if maxChunks < 1 {
		maxChunks = 1
	}
	if chunks > maxChunks {
		chunks = maxChunks
	}
	if chunks <= 1 {
		progress.setWritten(0, size)
		return downloadSingleStream(ctx, client, url, tempPath, remoteDownloadInfo{Size: size, SupportsRanges: true}, options, 0, progress)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	chunkSize := size / int64(chunks)
	if chunkSize <= 0 {
		chunkSize = size
	}
	errCh := make(chan error, chunks)
	var wg sync.WaitGroup
	for index := 0; index < chunks; index++ {
		start := int64(index) * chunkSize
		end := start + chunkSize - 1
		if index == chunks-1 {
			end = size - 1
		}
		wg.Add(1)
		go func(start int64, end int64) {
			defer wg.Done()
			if err := downloadChunkRange(ctx, client, url, file, start, end, options.RequestOptions, progress); err != nil {
				select {
				case errCh <- err:
				default:
				}
				cancel()
			}
		}(start, end)
	}
	wg.Wait()
	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

func downloadChunkRange(ctx context.Context, client *http.Client, url string, file *os.File, start int64, end int64, options RequestOptions, progress *downloadProgressWriter) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create chunk request: %w", err)
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	applyRequestHeaders(req, "", options)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download chunk %d-%d: %w", start, end, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusPartialContent {
		message, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("chunk download failed with http %d: %s", resp.StatusCode, strings.TrimSpace(string(message)))
	}

	offset := start
	buf := make([]byte, 128*1024)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, err := file.WriteAt(buf[:n], offset); err != nil {
				return fmt.Errorf("write chunk: %w", err)
			}
			offset += int64(n)
			progress.add(int64(n))
		}
		if errors.Is(readErr, io.EOF) {
			return nil
		}
		if readErr != nil {
			return fmt.Errorf("read chunk: %w", readErr)
		}
	}
}

func probeRemoteDownload(ctx context.Context, client *http.Client, url string, options RequestOptions) (remoteDownloadInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return remoteDownloadInfo{}, err
	}
	applyRequestHeaders(req, "", options)
	resp, err := client.Do(req)
	if err != nil {
		return remoteDownloadInfo{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return remoteDownloadInfo{}, fmt.Errorf("http %d", resp.StatusCode)
	}
	return remoteDownloadInfo{
		Size:           resp.ContentLength,
		SupportsRanges: strings.Contains(strings.ToLower(resp.Header.Get("Accept-Ranges")), "bytes"),
	}, nil
}

func verifyDownloadedFile(path string, plan DownloadPlan) (archiveDigests, error) {
	return verifyFileIntegrity(
		path,
		plan.Shasum,
		plan.Integrity,
		fmt.Sprintf("download for %s@%s", blankIfEmpty(plan.Identifier, filepath.Base(path)), blankIfEmpty(plan.Version, "unknown")),
	)
}

func verifyFileIntegrity(path string, shasum string, integrity string, label string) (archiveDigests, error) {
	digests, err := fileDigests(path)
	if err != nil {
		return archiveDigests{}, err
	}
	label = strings.TrimSpace(label)
	if label == "" {
		label = filepath.Base(path)
	}
	if shasum = strings.TrimSpace(shasum); shasum != "" && !strings.EqualFold(digests.SHA1, shasum) {
		return archiveDigests{}, fmt.Errorf("%s", formatIntegrityMismatchLabel(label, "shasum"))
	}
	if integrity = strings.TrimSpace(integrity); integrity != "" {
		if err := verifySubresourceIntegrity(integrity, digests); err != nil {
			return archiveDigests{}, fmt.Errorf("%s: %w", formatIntegrityMismatchLabel(label, "integrity"), err)
		}
	}
	return digests, nil
}

func fileDigests(path string) (archiveDigests, error) {
	file, err := os.Open(path)
	if err != nil {
		return archiveDigests{}, fmt.Errorf("open file %s: %w", path, err)
	}
	defer file.Close()
	sha1Hasher := sha1.New()
	sha256Hasher := sha256.New()
	sha512Hasher := sha512.New()
	size, err := io.Copy(io.MultiWriter(sha1Hasher, sha256Hasher, sha512Hasher), file)
	if err != nil {
		return archiveDigests{}, fmt.Errorf("hash file %s: %w", path, err)
	}
	return archiveDigests{
		Size:   size,
		SHA1:   hex.EncodeToString(sha1Hasher.Sum(nil)),
		SHA256: hex.EncodeToString(sha256Hasher.Sum(nil)),
		SHA512: hex.EncodeToString(sha512Hasher.Sum(nil)),
	}, nil
}

func syncFileBeforeRename(path string) error {
	file, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("open temp file for sync %s: %w", path, err)
	}
	defer file.Close()
	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync temp file %s: %w", path, err)
	}
	return nil
}

func formatIntegrityMismatchLabel(label string, kind string) string {
	label = strings.TrimSpace(label)
	kind = strings.TrimSpace(kind)
	if label == "" || kind == "" {
		return strings.TrimSpace(label + " " + kind + " mismatch")
	}
	if prefix, suffix, ok := strings.Cut(label, " for "); ok {
		return strings.TrimSpace(prefix) + " " + kind + " mismatch for " + strings.TrimSpace(suffix)
	}
	return label + " " + kind + " mismatch"
}

func newDownloadProgressWriter(out io.Writer, name string, total int64, start time.Time, initialWritten int64, onProgress func(DownloadProgress)) *downloadProgressWriter {
	return &downloadProgressWriter{
		out:        out,
		enabled:    shouldRenderDownloadProgress(out),
		name:       strings.TrimSpace(name),
		total:      total,
		start:      start,
		written:    initialWritten,
		onProgress: onProgress,
	}
}

func (w *downloadProgressWriter) Write(p []byte) (int, error) {
	if w == nil {
		return len(p), nil
	}
	w.add(int64(len(p)))
	return len(p), nil
}

func (w *downloadProgressWriter) add(delta int64) {
	if w == nil || delta == 0 {
		return
	}
	w.mu.Lock()
	w.written += delta
	w.emitLocked(false)
	w.mu.Unlock()
}

func (w *downloadProgressWriter) setWritten(value int64, total int64) {
	if w == nil {
		return
	}
	w.mu.Lock()
	w.written = value
	if total > 0 {
		w.total = total
	}
	w.emitLocked(false)
	w.mu.Unlock()
}

func (w *downloadProgressWriter) finish() {
	if w == nil {
		return
	}
	w.mu.Lock()
	w.emitLocked(true)
	w.mu.Unlock()
	if !w.enabled {
		return
	}
	_, _ = fmt.Fprintln(w.out)
}

func (w *downloadProgressWriter) emitLocked(force bool) {
	if w == nil {
		return
	}
	now := time.Now()
	if !force && !w.lastPrint.IsZero() && now.Sub(w.lastPrint) < downloadProgressInterval {
		return
	}
	if w.onProgress != nil {
		w.onProgress(DownloadProgress{
			Written: w.written,
			Total:   w.total,
			Elapsed: now.Sub(w.start),
		})
	}
	if !w.enabled {
		w.lastPrint = now
		return
	}
	line := formatDownloadProgressLine(w.name, w.written, w.total, now.Sub(w.start))
	padding := ""
	if diff := w.lastLineLen - len(line); diff > 0 {
		padding = strings.Repeat(" ", diff)
	}
	_, _ = fmt.Fprintf(w.out, "\r%s%s", line, padding)
	w.lastLineLen = len(line)
	w.lastPrint = now
}

func shouldRenderDownloadProgress(out io.Writer) bool {
	file, ok := out.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func formatDownloadProgressLine(name string, written int64, total int64, elapsed time.Duration) string {
	label := strings.TrimSpace(name)
	if label == "" {
		label = "download"
	}
	if elapsed <= 0 {
		elapsed = time.Millisecond
	}
	speed := formatByteSize(int64(float64(written) / elapsed.Seconds()))
	if total > 0 {
		percent := float64(written) * 100 / float64(total)
		if percent > 100 {
			percent = 100
		}
		return fmt.Sprintf(
			"Downloading %-24s %8s / %8s %6.1f%% %8s/s",
			truncateText(label, 24),
			formatByteSize(written),
			formatByteSize(total),
			percent,
			speed,
		)
	}
	return fmt.Sprintf(
		"Downloading %-24s %8s %8s/s",
		truncateText(label, 24),
		formatByteSize(written),
		speed,
	)
}

func formatByteSize(size int64) string {
	if size < 0 {
		size = 0
	}
	units := []string{"B", "KB", "MB", "GB", "TB"}
	value := float64(size)
	unit := units[0]
	for index := 1; index < len(units) && value >= 1024; index++ {
		value /= 1024
		unit = units[index]
	}
	if unit == "B" {
		return fmt.Sprintf("%d%s", size, unit)
	}
	return fmt.Sprintf("%.1f%s", value, unit)
}

func getJSON(ctx context.Context, client *http.Client, url string, options RequestOptions, target any) error {
	body, err := getBytesWithAccept(ctx, client, url, "application/json", options)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("parse json: %w", err)
	}
	return nil
}

func getBytes(ctx context.Context, client *http.Client, url string, options RequestOptions) ([]byte, error) {
	return getBytesWithAccept(ctx, client, url, "*/*", options)
}

func getBytesWithAccept(ctx context.Context, client *http.Client, url string, accept string, options RequestOptions) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(accept) == "" {
		accept = "*/*"
	}
	applyRequestHeaders(req, accept, options)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return body, nil
}

func applyRequestHeaders(req *http.Request, accept string, options RequestOptions) {
	if req == nil {
		return
	}
	req.Header.Set("User-Agent", "source-fetcher/"+version)
	if strings.TrimSpace(accept) != "" {
		req.Header.Set("Accept", accept)
	}
	for key, value := range options.Headers {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey == "" {
			continue
		}
		req.Header.Set(trimmedKey, strings.TrimSpace(value))
	}
	applyGitHubAuth(req)
}

func applyGitHubAuth(req *http.Request) {
	if req == nil || req.URL == nil {
		return
	}
	if !strings.EqualFold(req.URL.Hostname(), "api.github.com") {
		return
	}
	if strings.TrimSpace(req.Header.Get("Authorization")) != "" {
		return
	}
	token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
	if token == "" {
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)
}

func joinURL(baseURL string, suffix string) string {
	return strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(suffix, "/")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func parseMavenCoordinate(value string) (string, string, error) {
	value = strings.TrimSpace(value)
	parts := strings.Split(value, ":")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", "", errors.New("--name must be a Maven coordinate in group:artifact form when --source maven")
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
}

func mavenCoordinate(groupID string, artifactID string) string {
	return strings.TrimSpace(groupID) + ":" + strings.TrimSpace(artifactID)
}

func rewriteMirrorURL(rawURL string, targetBase string) string {
	if strings.TrimSpace(rawURL) == "" || strings.TrimSpace(targetBase) == "" {
		return rawURL
	}

	target, err := neturl.Parse(strings.TrimSpace(targetBase))
	if err != nil || target.Scheme == "" || target.Host == "" {
		return rawURL
	}
	parsed, err := neturl.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Host == "" {
		return rawURL
	}

	parsed.Scheme = target.Scheme
	parsed.Host = target.Host
	if strings.TrimSpace(target.Path) != "" && strings.TrimSpace(target.Path) != "/" {
		if strings.HasPrefix(strings.Trim(target.Path, "/"), "gh/") {
			parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
			if len(parts) >= 4 {
				parsed.Path = strings.TrimRight(target.Path, "/") + "/" + strings.Join(parts[3:], "/")
				return parsed.String()
			}
		}
		if parsed.Host == target.Host && strings.HasPrefix(parsed.Path, strings.TrimRight(target.Path, "/")+"/") {
			return parsed.String()
		}
		parsed.Path = joinURL(target.Path, parsed.Path)
	}
	return parsed.String()
}

func filenameFromURLOrFallback(rawURL string, fallback string) string {
	parsed, err := neturl.Parse(rawURL)
	if err != nil {
		return sanitizeFileName(fallback)
	}
	name := path.Base(parsed.Path)
	if name == "." || name == "/" || name == "" {
		return sanitizeFileName(fallback)
	}
	return sanitizeFileName(name)
}

func sanitizeFileName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	replacer := strings.NewReplacer(
		"<", "_",
		">", "_",
		":", "_",
		"\"", "_",
		"/", "_",
		"\\", "_",
		"|", "_",
		"?", "_",
		"*", "_",
	)
	return replacer.Replace(name)
}

func ensureUniquePath(target string) (string, error) {
	if _, err := os.Stat(target); errors.Is(err, os.ErrNotExist) {
		return target, nil
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return "", err
	}

	ext := filepath.Ext(target)
	base := strings.TrimSuffix(filepath.Base(target), ext)
	dir := filepath.Dir(target)
	for index := 1; index < 1000; index++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s(%d)%s", base, index, ext))
		if _, err := os.Stat(candidate); errors.Is(err, os.ErrNotExist) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("unable to allocate unique file path for %s", target)
}

func compareVersionish(left string, right string) int {
	leftVersion := parseVersionish(left)
	rightVersion := parseVersionish(right)

	if cmp := compareVersionCore(leftVersion.core, rightVersion.core); cmp != 0 {
		return cmp
	}
	switch {
	case len(leftVersion.prerelease) == 0 && len(rightVersion.prerelease) == 0:
		return 0
	case len(leftVersion.prerelease) == 0:
		return 1
	case len(rightVersion.prerelease) == 0:
		return -1
	default:
		return comparePrereleaseTokens(leftVersion.prerelease, rightVersion.prerelease)
	}
}

type versionish struct {
	core       []string
	prerelease []string
}

func parseVersionish(value string) versionish {
	value = strings.TrimSpace(value)
	if value == "" {
		return versionish{}
	}

	if plus := strings.IndexRune(value, '+'); plus >= 0 {
		value = value[:plus]
	}

	coreText := value
	prereleaseText := ""
	if dash := strings.IndexRune(value, '-'); dash >= 0 {
		coreText = value[:dash]
		prereleaseText = value[dash+1:]
	}

	return versionish{
		core:       splitVersionish(coreText),
		prerelease: splitVersionish(prereleaseText),
	}
}

func compareVersionCore(left []string, right []string) int {
	length := len(left)
	if len(right) > length {
		length = len(right)
	}

	for index := 0; index < length; index++ {
		leftPart, leftOK := versionPartAt(left, index)
		rightPart, rightOK := versionPartAt(right, index)
		if !leftOK {
			leftPart = "0"
		}
		if !rightOK {
			rightPart = "0"
		}
		if cmp := compareVersionToken(leftPart, rightPart, false); cmp != 0 {
			return cmp
		}
	}
	return 0
}

func comparePrereleaseTokens(left []string, right []string) int {
	for index := 0; ; index++ {
		leftPart, leftOK := versionPartAt(left, index)
		rightPart, rightOK := versionPartAt(right, index)
		switch {
		case !leftOK && !rightOK:
			return 0
		case !leftOK:
			return -1
		case !rightOK:
			return 1
		}
		if cmp := compareVersionToken(leftPart, rightPart, true); cmp != 0 {
			return cmp
		}
	}
}

func compareVersionToken(left string, right string, prerelease bool) int {
	leftNum, leftErr := strconv.Atoi(left)
	rightNum, rightErr := strconv.Atoi(right)
	switch {
	case leftErr == nil && rightErr == nil:
		if leftNum > rightNum {
			return 1
		}
		if leftNum < rightNum {
			return -1
		}
		return 0
	case prerelease && leftErr == nil && rightErr != nil:
		return -1
	case prerelease && leftErr != nil && rightErr == nil:
		return 1
	default:
		if left == right {
			return 0
		}
		if left > right {
			return 1
		}
		return -1
	}
}

func splitVersionish(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	var (
		parts   []string
		current strings.Builder
		lastNum bool
	)

	flush := func() {
		if current.Len() == 0 {
			return
		}
		parts = append(parts, current.String())
		current.Reset()
	}

	for _, r := range value {
		if r == '.' || r == '-' || r == '_' {
			flush()
			lastNum = false
			continue
		}
		isNum := r >= '0' && r <= '9'
		if current.Len() > 0 && isNum != lastNum {
			flush()
		}
		current.WriteRune(r)
		lastNum = isNum
	}
	flush()
	return parts
}

func versionPartAt(parts []string, index int) (string, bool) {
	if index >= 0 && index < len(parts) {
		return parts[index], true
	}
	return "", false
}

func uniqueLatestSearchResults(results []SearchResult, limit int) []SearchResult {
	if len(results) == 0 {
		return nil
	}

	unique := make([]SearchResult, 0, len(results))
	indexByKey := make(map[string]int, len(results))
	for _, result := range results {
		key := strings.ToLower(strings.TrimSpace(result.Source) + "\x00" + strings.TrimSpace(result.Identifier))
		index, exists := indexByKey[key]
		if !exists {
			indexByKey[key] = len(unique)
			unique = append(unique, result)
			continue
		}

		current := unique[index]
		if compareVersionish(result.Version, current.Version) > 0 {
			unique[index] = result
			continue
		}
		if strings.TrimSpace(current.Description) == "" && strings.TrimSpace(result.Description) != "" {
			current.Description = result.Description
		}
		if strings.TrimSpace(current.Detail) == "" && strings.TrimSpace(result.Detail) != "" {
			current.Detail = result.Detail
		}
		unique[index] = current
	}

	if limit > 0 && len(unique) > limit {
		return unique[:limit]
	}
	return unique
}
