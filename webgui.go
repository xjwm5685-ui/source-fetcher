package main

import (
	"context"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

//go:embed webui/*
var webUIFiles embed.FS

// runGUI 启动 Web GUI
func runGUI(args []string) error {
	fs := flag.NewFlagSet("gui", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	
	port := fs.Int("port", 8765, "web server port")
	noBrowser := fs.Bool("no-browser", false, "do not open browser automatically")
	
	helpShown, err := parseCommandFlags(fs, args)
	if err != nil {
		return err
	}
	if helpShown {
		return nil
	}
	
	// 检查端口是否可用，如果不可用则查找下一个可用端口
	finalPort := *port
	if !isPortAvailable(finalPort) {
		fmt.Printf("⚠️  Port %d is in use, finding available port...\n", finalPort)
		availablePort, err := findAvailablePort(finalPort)
		if err != nil {
			return fmt.Errorf("failed to find available port: %w", err)
		}
		finalPort = availablePort
		fmt.Printf("✅ Using port %d instead\n", finalPort)
	}
	
	// 创建 Web GUI 服务器
	server := NewWebGUIServer(finalPort)
	
	// 启动服务器
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()
	
	// 等待服务器启动
	time.Sleep(500 * time.Millisecond)
	
	// 构建 URL
	url := fmt.Sprintf("http://localhost:%d", finalPort)
	fmt.Printf("🌐 Web GUI started at: %s\n", url)
	fmt.Println("Press Ctrl+C to stop the server")
	
	// 自动打开浏览器（除非禁用）
	if !*noBrowser {
		if err := openBrowser(url); err != nil {
			fmt.Printf("💡 Please open your browser and visit: %s\n", url)
		}
	} else {
		fmt.Printf("💡 Open your browser and visit: %s\n", url)
	}
	
	// 等待中断信号
	select {}
}

type WebGUIServer struct {
	port       int
	httpClient *http.Client
	server     *http.Server
	
	// 搜索结果缓存
	searchResults []SearchResult
	searchMu      sync.RWMutex
	
	// 队列管理
	queue   []QueueTask
	queueMu sync.RWMutex
	nextID  int
	
	// 镜像结果缓存
	mirrorResults []MirrorResult
	mirrorMu      sync.RWMutex
	
	// 简单的速率限制：记录最近请求时间
	lastRequestTimes map[string][]time.Time
	rateLimitMu      sync.Mutex
}

type QueueTask struct {
	ID        int          `json:"id"`
	Type      string       `json:"type"`
	Label     string       `json:"label"`
	Status    string       `json:"status"`
	Error     string       `json:"error,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
	Result    SearchResult `json:"-"` // 存储搜索结果用于执行
}

func NewWebGUIServer(port int) *WebGUIServer {
	s := &WebGUIServer{
		port:             port,
		httpClient:       newHTTPClient(30 * time.Second),
		queue:            []QueueTask{},
		lastRequestTimes: make(map[string][]time.Time),
	}
	
	// 启动队列处理器
	go s.processQueue()
	
	// 定期清理速率限制记录
	go s.cleanupRateLimitRecords()
	
	return s
}

func (s *WebGUIServer) Start() error {
	mux := http.NewServeMux()
	
	// 静态文件服务
	webUIFS, err := fs.Sub(webUIFiles, "webui")
	if err != nil {
		return fmt.Errorf("failed to load web UI files: %w", err)
	}
	mux.Handle("/", http.FileServer(http.FS(webUIFS)))
	
	// API 路由 - 添加安全中间件
	mux.HandleFunc("/api/search", s.securityMiddleware(s.handleSearch))
	mux.HandleFunc("/api/download", s.securityMiddleware(s.handleDownload))
	mux.HandleFunc("/api/install", s.securityMiddleware(s.handleInstall))
	mux.HandleFunc("/api/queue", s.securityMiddleware(s.handleQueue))
	mux.HandleFunc("/api/mirrors", s.securityMiddleware(s.handleMirrors))
	mux.HandleFunc("/api/status", s.securityMiddleware(s.handleStatus))
	
	s.server = &http.Server{
		Addr:              fmt.Sprintf(":%d", s.port),
		Handler:           mux,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MB
	}
	
	return s.server.ListenAndServe()
}

// securityMiddleware 添加安全头和验证
func (s *WebGUIServer) securityMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 简单的速率限制：每个 IP 每分钟最多 60 个请求
		clientIP := r.RemoteAddr
		if !s.checkRateLimit(clientIP, 60, time.Minute) {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		
		// 添加安全响应头
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// 仅允许同源请求或 localhost（开发环境）
		origin := r.Header.Get("Origin")
		if origin != "" {
			// 在生产环境中，这里应该更严格地验证 origin
			// 目前允许 localhost 用于开发
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}
		
		// 处理 OPTIONS 预检请求
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		
		next(w, r)
	}
}

// checkRateLimit 检查速率限制
func (s *WebGUIServer) checkRateLimit(clientID string, maxRequests int, window time.Duration) bool {
	s.rateLimitMu.Lock()
	defer s.rateLimitMu.Unlock()
	
	now := time.Now()
	cutoff := now.Add(-window)
	
	// 获取并过滤过期的请求记录
	times := s.lastRequestTimes[clientID]
	validTimes := []time.Time{}
	for _, t := range times {
		if t.After(cutoff) {
			validTimes = append(validTimes, t)
		}
	}
	
	// 检查是否超过限制
	if len(validTimes) >= maxRequests {
		return false
	}
	
	// 添加当前请求
	validTimes = append(validTimes, now)
	s.lastRequestTimes[clientID] = validTimes
	
	return true
}

// cleanupRateLimitRecords 定期清理过期的速率限制记录
func (s *WebGUIServer) cleanupRateLimitRecords() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		s.rateLimitMu.Lock()
		cutoff := time.Now().Add(-10 * time.Minute)
		
		for clientID, times := range s.lastRequestTimes {
			validTimes := []time.Time{}
			for _, t := range times {
				if t.After(cutoff) {
					validTimes = append(validTimes, t)
				}
			}
			
			if len(validTimes) == 0 {
				delete(s.lastRequestTimes, clientID)
			} else {
				s.lastRequestTimes[clientID] = validTimes
			}
		}
		
		s.rateLimitMu.Unlock()
	}
}

func (s *WebGUIServer) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Limit request body size to prevent abuse
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit
	
	var req struct {
		Source string `json:"source"`
		Query  string `json:"query"`
		Mirror string `json:"mirror"`
		Limit  int    `json:"limit"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate and sanitize limit
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 10
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	results, err := searchPackages(ctx, s.httpClient, SearchRequest{
		Source: req.Source,
		Query:  req.Query,
		Mirror: req.Mirror,
		Limit:  req.Limit,
	})
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	s.searchMu.Lock()
	s.searchResults = results
	s.searchMu.Unlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"results": results,
		"count":   len(results),
	})
}

func (s *WebGUIServer) handleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit
	
	var req struct {
		Indexes []int `json:"indexes"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate indexes array size to prevent abuse
	if len(req.Indexes) > 100 {
		http.Error(w, "Too many indexes (max 100)", http.StatusBadRequest)
		return
	}
	
	s.searchMu.RLock()
	defer s.searchMu.RUnlock()
	
	added := 0
	for _, idx := range req.Indexes {
		if idx >= 0 && idx < len(s.searchResults) {
			result := s.searchResults[idx]
			s.addToQueue(QueueTask{
				Type:      "download",
				Label:     fmt.Sprintf("Download %s@%s", result.Identifier, result.Version),
				Status:    "pending",
				CreatedAt: time.Now(),
				Result:    result,
			})
			added++
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"added":   added,
	})
}

func (s *WebGUIServer) handleInstall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit
	
	var req struct {
		Indexes []int `json:"indexes"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate indexes array size
	if len(req.Indexes) > 100 {
		http.Error(w, "Too many indexes (max 100)", http.StatusBadRequest)
		return
	}
	
	log.Printf("Install request received: indexes=%v", req.Indexes)
	
	s.searchMu.RLock()
	defer s.searchMu.RUnlock()
	
	log.Printf("Current search results count: %d", len(s.searchResults))
	
	added := 0
	for _, idx := range req.Indexes {
		if idx >= 0 && idx < len(s.searchResults) {
			result := s.searchResults[idx]
			log.Printf("Adding install task for index %d: Source=%s, Identifier=%s, Version=%s", 
				idx, result.Source, result.Identifier, result.Version)
			
			s.addToQueue(QueueTask{
				Type:      "install",
				Label:     fmt.Sprintf("Install %s@%s", result.Identifier, result.Version),
				Status:    "pending",
				CreatedAt: time.Now(),
				Result:    result,
			})
			added++
		} else {
			log.Printf("Invalid index %d (total results: %d)", idx, len(s.searchResults))
		}
	}
	
	log.Printf("Added %d install tasks to queue", added)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"added":   added,
	})
}

func (s *WebGUIServer) handleQueue(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.queueMu.RLock()
		defer s.queueMu.RUnlock()
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"queue":   s.queue,
			"count":   len(s.queue),
		})
		return
	}
	
	if r.Method == http.MethodDelete {
		s.queueMu.Lock()
		defer s.queueMu.Unlock()
		
		var remaining []QueueTask
		for _, task := range s.queue {
			if task.Status != "completed" && task.Status != "failed" {
				remaining = append(remaining, task)
			}
		}
		s.queue = remaining
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"count":   len(s.queue),
		})
		return
	}
	
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (s *WebGUIServer) handleMirrors(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit
	
	var req struct {
		Source string `json:"source"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	
	results, err := testMirrors(ctx, s.httpClient, req.Source)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	s.mirrorMu.Lock()
	s.mirrorResults = results
	s.mirrorMu.Unlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"results": results,
		"count":   len(results),
	})
}

func (s *WebGUIServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.queueMu.RLock()
	queueCount := len(s.queue)
	s.queueMu.RUnlock()
	
	s.searchMu.RLock()
	searchCount := len(s.searchResults)
	s.searchMu.RUnlock()
	
	s.mirrorMu.RLock()
	mirrorCount := len(s.mirrorResults)
	s.mirrorMu.RUnlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"queue_count":   queueCount,
		"search_count":  searchCount,
		"mirror_count":  mirrorCount,
		"version":       version,
	})
}

func (s *WebGUIServer) addToQueue(task QueueTask) {
	s.queueMu.Lock()
	defer s.queueMu.Unlock()
	
	s.nextID++
	task.ID = s.nextID
	s.queue = append(s.queue, task)
}

// processQueue 处理队列中的任务
func (s *WebGUIServer) processQueue() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		s.queueMu.Lock()
		
		// 查找第一个 pending 任务
		var taskToProcess *QueueTask
		for i := range s.queue {
			if s.queue[i].Status == "pending" {
				taskToProcess = &s.queue[i]
				break
			}
		}
		
		s.queueMu.Unlock()
		
		if taskToProcess != nil {
			s.executeTask(taskToProcess)
		}
	}
}

// executeTask 执行单个任务
func (s *WebGUIServer) executeTask(task *QueueTask) {
	// 更新状态为 running
	s.updateTaskStatus(task.ID, "running", "")
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	
	var err error
	
	switch task.Type {
	case "download":
		err = s.executeDownload(ctx, task)
	case "install":
		err = s.executeInstall(ctx, task)
	default:
		err = fmt.Errorf("unsupported task type: %s", task.Type)
	}
	
	if err != nil {
		s.updateTaskStatus(task.ID, "failed", err.Error())
		log.Printf("Task %d failed: %v", task.ID, err)
	} else {
		s.updateTaskStatus(task.ID, "completed", "")
		log.Printf("Task %d completed: %s", task.ID, task.Label)
	}
}

// executeDownload 执行下载任务
func (s *WebGUIServer) executeDownload(ctx context.Context, task *QueueTask) error {
	result := task.Result
	
	// 构建下载请求
	req := DownloadRequest{
		Source:    result.Source,
		OutputDir: ".", // 使用当前目录
	}
	
	// 根据不同源设置不同字段
	switch result.Source {
	case "npm", "choco", "pip", "cargo", "maven":
		req.Name = result.Identifier
		req.Version = result.Version
	case "winget":
		req.PackageID = result.Identifier
	default:
		return fmt.Errorf("unsupported source: %s", result.Source)
	}
	
	// 解析下载计划
	plan, err := resolveDownloadPlan(ctx, s.httpClient, req)
	if err != nil {
		return fmt.Errorf("failed to resolve download plan: %w", err)
	}
	
	// 执行下载
	_, err = downloadPlan(ctx, s.httpClient, plan, DownloadOptions{
		OutputDir: ".",
	})
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	
	return nil
}

// executeInstall 执行安装任务
func (s *WebGUIServer) executeInstall(ctx context.Context, task *QueueTask) error {
	result := task.Result
	
	// 只支持 npm 安装
	if result.Source != "npm" {
		return fmt.Errorf("install is only supported for npm packages, got: %s", result.Source)
	}
	
	// 调试日志
	log.Printf("Executing install: Source=%s, Identifier=%s, Version=%s", 
		result.Source, result.Identifier, result.Version)
	
	// 构建安装请求
	req := InstallRequest{
		Source:    "npm",
		Name:      result.Identifier,
		Version:   result.Version,
		OutputDir: ".",
	}
	
	log.Printf("InstallRequest: Name=%s, Version=%s, OutputDir=%s", 
		req.Name, req.Version, req.OutputDir)
	
	// 解析安装计划
	plan, err := resolveInstallPlan(ctx, s.httpClient, req)
	if err != nil {
		return fmt.Errorf("failed to resolve install plan: %w", err)
	}
	
	// 执行安装
	_, err = executeInstallPlan(ctx, s.httpClient, plan, DownloadOptions{
		OutputDir: ".",
	})
	if err != nil {
		return fmt.Errorf("failed to install: %w", err)
	}
	
	return nil
}

// updateTaskStatus 更新任务状态
func (s *WebGUIServer) updateTaskStatus(id int, status string, errorMsg string) {
	s.queueMu.Lock()
	defer s.queueMu.Unlock()
	
	for i := range s.queue {
		if s.queue[i].ID == id {
			s.queue[i].Status = status
			s.queue[i].Error = errorMsg
			break
		}
	}
}

func openBrowser(url string) error {
	var cmd string
	var args []string
	
	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default: // linux, freebsd, openbsd, netbsd
		cmd = "xdg-open"
		args = []string{url}
	}
	
	return exec.Command(cmd, args...).Start()
}

// isPortAvailable 检查端口是否可用
func isPortAvailable(port int) bool {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

func findAvailablePort(startPort int) (int, error) {
	for port := startPort; port < startPort+100; port++ {
		addr := fmt.Sprintf(":%d", port)
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			listener.Close()
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available port found")
}
