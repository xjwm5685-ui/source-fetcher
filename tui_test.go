package main

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestVisibleWindowCentersAroundSelection(t *testing.T) {
	start, end := visibleWindow(10, 5, 4)
	if start != 3 || end != 7 {
		t.Fatalf("unexpected window: start=%d end=%d", start, end)
	}
}

func TestVisibleWindowClampsToEnd(t *testing.T) {
	start, end := visibleWindow(10, 9, 4)
	if start != 6 || end != 10 {
		t.Fatalf("unexpected window near end: start=%d end=%d", start, end)
	}
}

func TestFormatSearchRowIncludesKeyFields(t *testing.T) {
	row := formatSearchRow(2, SearchResult{
		Source:      "winget",
		Identifier:  "Microsoft.WindowsTerminal",
		Version:     "1.22.10731.0",
		Description: "Windows Terminal application",
	}, 120)
	if !strings.Contains(row, "[2]") || !strings.Contains(row, "WINGET") || !strings.Contains(row, "Microsoft.WindowsTerminal") {
		t.Fatalf("expected row to include key fields, got %q", row)
	}
}

func TestFormatDownloadPlanTextIncludesResolvedFields(t *testing.T) {
	text := formatDownloadPlanText(DownloadPlan{
		Source:     "npm",
		Identifier: "react",
		Version:    "19.2.0",
		URL:        "https://registry.npmjs.org/react/-/react-19.2.0.tgz",
		Filename:   "react-19.2.0.tgz",
	})
	if !strings.Contains(text, "Source: npm") || !strings.Contains(text, "Filename: react-19.2.0.tgz") {
		t.Fatalf("expected formatted plan text, got %q", text)
	}
}

func TestSearchPickOptionsIncludesWingetOverrides(t *testing.T) {
	outputDir := t.TempDir()
	model := newTUIModel(newHTTPClient(time.Second), time.Second, "winget", "", outputDir, "github-raw", 10, "", "arm64", 2, true, 4, "source-fetcher.yaml", true)

	options, err := model.searchPickOptions()
	if err != nil {
		t.Fatalf("searchPickOptions returned error: %v", err)
	}
	if options.OutputDir != outputDir {
		t.Fatalf("expected output dir %q, got %q", outputDir, options.OutputDir)
	}
	if options.Mirror != "github-raw" || options.Arch != "arm64" || options.InstallerIndex != 2 || !options.Resume || options.Chunks != 4 {
		t.Fatalf("unexpected pick options: %+v", options)
	}
}

func TestSelectedBatchRequestsCarryTimeoutAndOrigin(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "source-fetcher.yaml")
	model := newTUIModel(newHTTPClient(time.Second), time.Second, "all", "", t.TempDir(), "", 10, "", "", -1, false, 1, configPath, false)
	model.batchTasks = []tuiBatchTask{
		{
			Label: "npm:react",
			Item: tuiQueueItem{
				Action: "download",
				Request: DownloadRequest{Source: "npm", Name: "react", OutputDir: t.TempDir()},
			},
		},
	}
	model.batchTimeout = 45 * time.Second

	items, err := model.selectedBatchRequests()
	if err != nil {
		t.Fatalf("selectedBatchRequests returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 queued batch item, got %d", len(items))
	}
	if items[0].Timeout != 45*time.Second {
		t.Fatalf("expected batch timeout to carry into queue item, got %s", items[0].Timeout)
	}
	if items[0].Origin != "batch:source-fetcher.yaml" {
		t.Fatalf("unexpected queue origin %q", items[0].Origin)
	}
	if items[0].Action != "download" {
		t.Fatalf("expected download batch action, got %q", items[0].Action)
	}
	if items[0].Request.Name != "react" {
		t.Fatalf("unexpected queued request: %+v", items[0].Request)
	}
}

func TestSelectedSearchInstallRequestsCreateNPMInstallTasks(t *testing.T) {
	outputDir := t.TempDir()
	model := newTUIModel(newHTTPClient(time.Second), time.Second, "npm", "", outputDir, "npmmirror", 10, "", "", -1, false, 1, "source-fetcher.yaml", true)
	model.searchResults = []SearchResult{{Source: "npm", Identifier: "react", Version: "19.2.0"}}

	items, err := model.selectedSearchInstallRequests()
	if err != nil {
		t.Fatalf("selectedSearchInstallRequests returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 install item, got %d", len(items))
	}
	if items[0].Action != "install" || items[0].Install.Name != "react" || items[0].Install.OutputDir != outputDir {
		t.Fatalf("unexpected install queue item: %+v", items[0])
	}
}

func TestQueueEnqueueSupportsInstallTasks(t *testing.T) {
	queue := &tuiQueueState{}
	count := queue.enqueue([]tuiQueueItem{{
		Action: "install",
		Install: InstallRequest{
			Source:    "npm",
			Name:      "react",
			OutputDir: t.TempDir(),
		},
		Origin: "search",
	}})
	if count != 1 {
		t.Fatalf("expected 1 enqueued task, got %d", count)
	}
	snapshot := queue.snapshot()
	if len(snapshot) != 1 {
		t.Fatalf("expected 1 queued task, got %d", len(snapshot))
	}
	if snapshot[0].Action != "install" || snapshot[0].Label != "npm:react" {
		t.Fatalf("unexpected queued task: %+v", snapshot[0])
	}
}

func TestPageHelpUsesStepBasedGuidance(t *testing.T) {
	model := newTUIModel(newHTTPClient(time.Second), time.Second, "all", "", t.TempDir(), "", 10, "", "", -1, false, 1, "source-fetcher.yaml", true)
	model.page = tuiPageSearch

	help := model.pageHelp()
	if !strings.Contains(help, "当前页: Step1 配条件, Step2 选结果") {
		t.Fatalf("expected step-based help text, got %q", help)
	}
	if !strings.Contains(help, "i 安装入队") {
		t.Fatalf("expected install action in help text, got %q", help)
	}
}

func TestUpdateSearchDetailFromSelectionExplainsNextStep(t *testing.T) {
	model := newTUIModel(newHTTPClient(time.Second), time.Second, "npm", "", t.TempDir(), "", 10, "", "", -1, false, 1, "source-fetcher.yaml", true)
	model.searchResults = []SearchResult{
		{Source: "npm", Identifier: "react", Version: "19.2.0"},
		{Source: "npm", Identifier: "vite", Version: "6.0.0"},
	}
	model.searchSelected = map[int]bool{0: true, 1: true}

	model.updateSearchDetailFromSelection()

	if !strings.Contains(model.searchDetail, "Selected Results") {
		t.Fatalf("expected selection title, got %q", model.searchDetail)
	}
	if !strings.Contains(model.searchDetail, "下一步: d 下载入队，i 安装入队。") {
		t.Fatalf("expected next-step hint, got %q", model.searchDetail)
	}
}

func TestRenderHeaderBarIncludesPageAndQueueSummary(t *testing.T) {
	model := newTUIModel(newHTTPClient(time.Second), time.Second, "all", "", t.TempDir(), "", 10, "", "", -1, false, 1, "source-fetcher.yaml", true)
	model.width = 140
	model.page = tuiPageQueue
	model.queueTasks = []tuiQueueTask{{ID: 1}, {ID: 2}}

	header := model.renderHeaderBar()
	if !strings.Contains(header, "Source Fetcher") {
		t.Fatalf("expected app title in header, got %q", header)
	}
	if !strings.Contains(header, "QUEUE") || !strings.Contains(header, "Queue 2") {
		t.Fatalf("expected page and queue summary in header, got %q", header)
	}
}

func TestNewTUIModelStartsBusyWhenInitialQueryProvided(t *testing.T) {
	model := newTUIModel(newHTTPClient(time.Second), time.Second, "all", "terminal", t.TempDir(), "", 10, "", "", -1, false, 1, "source-fetcher.yaml", true)
	if !model.busy {
		t.Fatal("expected initial query to mark model busy")
	}
	if model.status != "正在搜索..." {
		t.Fatalf("expected initial status to reflect search, got %q", model.status)
	}
}

func TestQueueCompleteDoesNotResetRunningWhenTaskMissing(t *testing.T) {
	queue := &tuiQueueState{
		running: true,
		tasks:   []*tuiQueueTask{{ID: 1, Status: "downloading"}},
	}
	queue.complete(99, DownloadPlan{}, DownloadResult{})
	if !queue.running {
		t.Fatal("expected running to remain true when complete misses task")
	}
}

func TestQueueFailDoesNotResetRunningWhenTaskMissing(t *testing.T) {
	queue := &tuiQueueState{
		running: true,
		tasks:   []*tuiQueueTask{{ID: 1, Status: "downloading"}},
	}
	queue.fail(99, errExample("boom"))
	if !queue.running {
		t.Fatal("expected running to remain true when fail misses task")
	}
}

type errExample string

func (e errExample) Error() string {
	return string(e)
}
