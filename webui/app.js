// State
let searchResults = [];
let selectedIndexes = new Set();
let queueTasks = [];
let mirrorResults = [];

// DOM Elements
const navPills = document.querySelectorAll('.nav-pill');
const tabPanes = document.querySelectorAll('.tab-pane');
const searchForm = document.getElementById('searchForm');
const mirrorForm = document.getElementById('mirrorForm');
const resultsContainer = document.getElementById('resultsContainer');
const queueContainer = document.getElementById('queueContainer');
const mirrorResultsContainer = document.getElementById('mirrorResults');
const statusText = document.getElementById('statusText');
const statusTime = document.getElementById('statusTime');
const loadingOverlay = document.getElementById('loadingOverlay');
const loadingText = document.getElementById('loadingText');

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    initTabs();
    initForms();
    initButtons();
    initCustomSelects(); // 初始化自定义下拉框
    loadStatus();
    updateTime();
    setInterval(updateTime, 1000);
    setInterval(loadQueue, 5000); // Auto-refresh queue every 5s
    
    // Add real-time search
    const searchQueryInput = document.getElementById('searchQuery');
    if (searchQueryInput) {
        searchQueryInput.addEventListener('input', debounceSearch);
    }
    
    // Keyboard shortcuts
    initKeyboardShortcuts();
    
    // Show welcome message
    setTimeout(() => {
        showToast('👋 Welcome to Source Fetcher! Press ? for keyboard shortcuts.', 'info', 5000);
    }, 1000);
});

// Custom Select Component
function initCustomSelects() {
    const customSelects = document.querySelectorAll('.custom-select');
    
    customSelects.forEach(selectElement => {
        const trigger = selectElement.querySelector('.custom-select-trigger');
        const dropdown = selectElement.querySelector('.custom-select-dropdown');
        const options = selectElement.querySelectorAll('.custom-select-option');
        const valueDisplay = selectElement.querySelector('.custom-select-value');
        
        // 初始化选中值
        const selectedOption = selectElement.querySelector('.custom-select-option.selected');
        if (selectedOption) {
            selectElement.dataset.value = selectedOption.dataset.value;
        }
        
        // Toggle dropdown
        trigger.addEventListener('click', (e) => {
            e.stopPropagation();
            
            // Close other selects
            document.querySelectorAll('.custom-select.active').forEach(other => {
                if (other !== selectElement) {
                    other.classList.remove('active');
                }
            });
            
            selectElement.classList.toggle('active');
        });
        
        // Handle option selection
        options.forEach(option => {
            option.addEventListener('click', (e) => {
                e.stopPropagation();
                
                const value = option.dataset.value;
                const text = option.textContent;
                
                // Update UI
                options.forEach(opt => opt.classList.remove('selected'));
                option.classList.add('selected');
                valueDisplay.textContent = text;
                
                // Store value in data attribute
                selectElement.dataset.value = value;
                
                // Close dropdown
                selectElement.classList.remove('active');
            });
        });
    });
    
    // Close dropdowns when clicking outside
    document.addEventListener('click', () => {
        document.querySelectorAll('.custom-select.active').forEach(select => {
            select.classList.remove('active');
        });
    });
    
    // Keyboard navigation
    document.addEventListener('keydown', (e) => {
        const activeSelect = document.querySelector('.custom-select.active');
        if (!activeSelect) return;
        
        const options = Array.from(activeSelect.querySelectorAll('.custom-select-option'));
        const selectedOption = activeSelect.querySelector('.custom-select-option.selected');
        const currentIndex = options.indexOf(selectedOption);
        
        if (e.key === 'ArrowDown') {
            e.preventDefault();
            const nextIndex = Math.min(currentIndex + 1, options.length - 1);
            options[nextIndex].click();
        } else if (e.key === 'ArrowUp') {
            e.preventDefault();
            const prevIndex = Math.max(currentIndex - 1, 0);
            options[prevIndex].click();
        } else if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            activeSelect.classList.remove('active');
        } else if (e.key === 'Escape') {
            e.preventDefault();
            activeSelect.classList.remove('active');
        }
    });
}

// 辅助函数：获取自定义下拉框的值
function getCustomSelectValue(selectId) {
    const selectElement = document.getElementById(selectId);
    return selectElement ? selectElement.dataset.value : '';
}

// Keyboard shortcuts
function initKeyboardShortcuts() {
    document.addEventListener('keydown', (e) => {
        // Don't trigger shortcuts when typing in inputs
        if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA' || e.target.tagName === 'SELECT') {
            // Allow Ctrl+A in results
            if (e.ctrlKey && e.key === 'a' && document.activeElement.id === 'resultsContainer') {
                e.preventDefault();
                selectAll();
            }
            return;
        }
        
        // Cmd/Ctrl + K: Focus search
        if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
            e.preventDefault();
            document.getElementById('searchQuery').focus();
            return;
        }
        
        // Escape: Clear selection or close modals
        if (e.key === 'Escape') {
            if (selectedIndexes.size > 0) {
                clearSelection();
            }
            return;
        }
        
        // Ctrl/Cmd + A: Select all results
        if ((e.ctrlKey || e.metaKey) && e.key === 'a') {
            const activeTab = document.querySelector('.tab-pane.active');
            if (activeTab && activeTab.id === 'search' && searchResults.length > 0) {
                e.preventDefault();
                selectAll();
            }
            return;
        }
        
        // Numbers 1-4: Switch tabs
        if (e.key >= '1' && e.key <= '4' && !e.ctrlKey && !e.metaKey) {
            const tabs = ['search', 'queue', 'mirrors', 'about'];
            switchTab(tabs[parseInt(e.key) - 1]);
            return;
        }
        
        // D: Download selected
        if (e.key === 'd' || e.key === 'D') {
            if (selectedIndexes.size > 0) {
                handleDownload();
            }
            return;
        }
        
        // I: Install selected
        if (e.key === 'i' || e.key === 'I') {
            if (selectedIndexes.size > 0) {
                handleInstall();
            }
            return;
        }
        
        // R: Refresh queue
        if (e.key === 'r' || e.key === 'R') {
            const activeTab = document.querySelector('.tab-pane.active');
            if (activeTab && activeTab.id === 'queue') {
                loadQueue();
                showToast('🔄 Queue refreshed', 'info', 1500);
            }
            return;
        }
        
        // ?: Show keyboard shortcuts
        if (e.key === '?') {
            showKeyboardShortcuts();
            return;
        }
    });
}

// Show keyboard shortcuts modal
function showKeyboardShortcuts() {
    const shortcuts = [
        { key: 'Ctrl/Cmd + K', desc: 'Focus search box' },
        { key: '1-4', desc: 'Switch between tabs' },
        { key: 'Ctrl/Cmd + A', desc: 'Select all results' },
        { key: 'D', desc: 'Download selected packages' },
        { key: 'I', desc: 'Install selected packages' },
        { key: 'R', desc: 'Refresh queue' },
        { key: 'Esc', desc: 'Clear selection' },
        { key: '?', desc: 'Show this help' }
    ];
    
    const shortcutList = shortcuts.map(s => 
        `<div style="display: flex; justify-content: space-between; margin: 8px 0;">
            <span class="kbd">${s.key}</span>
            <span style="color: var(--gray-700);">${s.desc}</span>
        </div>`
    ).join('');
    
    const modal = document.createElement('div');
    modal.className = 'loading-overlay';
    modal.style.display = 'flex';
    modal.innerHTML = `
        <div style="background: white; padding: 32px; border-radius: 16px; max-width: 500px; box-shadow: 0 20px 60px rgba(0,0,0,0.3);">
            <h2 style="margin: 0 0 24px 0; color: var(--gray-900);">⌨️ Keyboard Shortcuts</h2>
            ${shortcutList}
            <button class="primary-button" style="margin-top: 24px; width: 100%;" onclick="this.closest('.loading-overlay').remove()">
                Got it!
            </button>
        </div>
    `;
    
    modal.addEventListener('click', (e) => {
        if (e.target === modal) {
            modal.remove();
        }
    });
    
    document.body.appendChild(modal);
}

// Tab Navigation
function initTabs() {
    navPills.forEach(pill => {
        pill.addEventListener('click', () => {
            const tabName = pill.dataset.tab;
            switchTab(tabName);
        });
    });
}

function switchTab(tabName) {
    navPills.forEach(p => p.classList.remove('active'));
    tabPanes.forEach(p => p.classList.remove('active'));
    
    document.querySelector(`[data-tab="${tabName}"]`).classList.add('active');
    document.getElementById(tabName).classList.add('active');
    
    if (tabName === 'queue') {
        loadQueue();
    }
}

// Forms
function initForms() {
    searchForm.addEventListener('submit', handleSearch);
    mirrorForm.addEventListener('submit', handleMirrorTest);
}

// Buttons
function initButtons() {
    document.getElementById('downloadBtn').addEventListener('click', handleDownload);
    document.getElementById('installBtn').addEventListener('click', handleInstall);
    document.getElementById('selectAllBtn').addEventListener('click', selectAll);
    document.getElementById('clearSelectionBtn').addEventListener('click', clearSelection);
    document.getElementById('refreshQueueBtn').addEventListener('click', loadQueue);
    document.getElementById('clearQueueBtn').addEventListener('click', clearQueue);
}

// Search
// Search with debounce and caching
let searchCache = new Map();
let searchAbortController = null;
let searchTimeout = null;

// Debounced search function
function debounceSearch() {
    clearTimeout(searchTimeout);
    searchTimeout = setTimeout(() => {
        const query = document.getElementById('searchQuery').value.trim();
        if (query.length >= 2) {
            handleSearch(new Event('submit'));
        }
    }, 800);
}

async function handleSearch(e) {
    e.preventDefault();
    
    const source = getCustomSelectValue('searchSourceSelect');
    const query = document.getElementById('searchQuery').value.trim();
    const mirror = document.getElementById('searchMirror').value.trim();
    const limit = parseInt(document.getElementById('searchLimit').value) || 10;
    
    if (!query) {
        showStatus('Please enter a search query', 'error');
        return;
    }
    
    // Check cache first
    const cacheKey = `${source}:${query}:${mirror}:${limit}`;
    if (searchCache.has(cacheKey)) {
        const cached = searchCache.get(cacheKey);
        searchResults = cached;
        selectedIndexes.clear();
        renderResults();
        showStatus(`Found ${cached.length} results (cached)`);
        return;
    }
    
    // Cancel previous search
    if (searchAbortController) {
        searchAbortController.abort();
    }
    searchAbortController = new AbortController();
    
    showLoading('Searching packages...');
    const startTime = Date.now();
    
    try {
        const response = await fetch('/api/search', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ source, query, mirror, limit }),
            signal: searchAbortController.signal
        });
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        const data = await response.json();
        const duration = ((Date.now() - startTime) / 1000).toFixed(2);
        
        if (data.success) {
            searchResults = data.results || [];
            selectedIndexes.clear();
            
            // Cache results
            searchCache.set(cacheKey, searchResults);
            
            // Limit cache size to 50 entries
            if (searchCache.size > 50) {
                const firstKey = searchCache.keys().next().value;
                searchCache.delete(firstKey);
            }
            
            renderResults();
            showStatus(`Found ${data.count} results in ${duration}s`, 'success');
        } else {
            throw new Error(data.error || 'Search failed');
        }
    } catch (error) {
        if (error.name === 'AbortError') {
            showStatus('Search cancelled');
        } else {
            showStatus('Search failed: ' + error.message, 'error');
            searchResults = [];
            renderResults();
        }
    } finally {
        hideLoading();
        searchAbortController = null;
    }
}

function renderResults() {
    const count = searchResults.length;
    document.getElementById('resultsCount').textContent = `${count} package${count !== 1 ? 's' : ''} found`;
    
    if (count === 0) {
        resultsContainer.innerHTML = `
            <div class="empty-state">
                <svg width="64" height="64" viewBox="0 0 64 64" fill="none">
                    <circle cx="28" cy="28" r="20" stroke="currentColor" stroke-width="4" opacity="0.2"/>
                    <path d="M42 42L54 54" stroke="currentColor" stroke-width="4" stroke-linecap="round" opacity="0.2"/>
                </svg>
                <p>No packages found. Try a different search query.</p>
            </div>
        `;
        updateActionButtons();
        return;
    }
    
    resultsContainer.innerHTML = searchResults.map((result, index) => {
        const isSelected = selectedIndexes.has(index);
        const sourceColors = {
            'npm': '#CB3837',
            'pip': '#3776AB',
            'cargo': '#000000',
            'maven': '#C71A36',
            'choco': '#80B5E3',
            'winget': '#0078D4'
        };
        const sourceColor = sourceColors[result.Source] || 'var(--google-blue)';
        
        return `
            <div class="result-item ${isSelected ? 'selected' : ''}" data-index="${index}" title="Click to view details">
                <input type="checkbox" class="result-checkbox" ${isSelected ? 'checked' : ''} title="Select package">
                <div class="result-content">
                    <div class="result-header">
                        <span class="result-source" style="background: ${sourceColor};">${result.Source.toUpperCase()}</span>
                        <span class="result-name">${escapeHtml(result.Identifier)}</span>
                        <span class="result-version" title="Version">v${escapeHtml(result.Version)}</span>
                    </div>
                    <div class="result-description" title="${escapeHtml(result.Description)}">${escapeHtml(truncate(result.Description, 100))}</div>
                </div>
            </div>
        `;
    }).join('');
    
    // Add event listeners
    resultsContainer.querySelectorAll('.result-item').forEach(item => {
        const index = parseInt(item.dataset.index);
        
        item.addEventListener('click', (e) => {
            if (e.target.classList.contains('result-checkbox')) {
                return;
            }
            toggleSelection(index);
            showResultDetail(index);
        });
        
        item.querySelector('.result-checkbox').addEventListener('change', (e) => {
            e.stopPropagation();
            toggleSelection(index);
        });
        
        // Add double-click to copy package name
        item.addEventListener('dblclick', (e) => {
            e.preventDefault();
            const result = searchResults[index];
            copyToClipboard(`${result.Identifier}@${result.Version}`);
        });
    });
    
    updateActionButtons();
}

function toggleSelection(index) {
    if (selectedIndexes.has(index)) {
        selectedIndexes.delete(index);
    } else {
        selectedIndexes.add(index);
    }
    renderResults();
}

function selectAll() {
    searchResults.forEach((_, index) => selectedIndexes.add(index));
    renderResults();
}

function clearSelection() {
    selectedIndexes.clear();
    renderResults();
}

function showResultDetail(index) {
    const result = searchResults[index];
    if (!result) return;
    
    const details = [
        `📦 Package: ${result.Identifier}`,
        `🏷️  Version: ${result.Version}`,
        `📚 Source: ${result.Source}`,
        ``,
        `📝 Description:`,
        result.Description || 'No description available',
        ``,
        `💡 Tip: Double-click on a result to copy package name`
    ];
    
    document.getElementById('detailText').textContent = details.join('\n');
    
    // Add copy button functionality to detail panel
    const detailCard = document.querySelector('.detail-card');
    const copyBtn = detailCard.querySelector('.copy-detail-btn');
    if (!copyBtn) {
        const copyButton = document.createElement('button');
        copyButton.className = 'icon-button copy-detail-btn';
        copyButton.innerHTML = `
            <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
                <rect x="7" y="7" width="10" height="10" rx="2" stroke="currentColor" stroke-width="2"/>
                <path d="M3 13V5a2 2 0 012-2h8" stroke="currentColor" stroke-width="2"/>
            </svg>
        `;
        copyButton.title = 'Copy package name';
        copyButton.addEventListener('click', () => {
            copyToClipboard(`${result.Identifier}@${result.Version}`);
        });
        detailCard.querySelector('.card-header .action-buttons')?.appendChild(copyButton);
    }
}

function updateActionButtons() {
    const hasSelection = selectedIndexes.size > 0;
    const hasResults = searchResults.length > 0;
    
    document.getElementById('downloadBtn').disabled = !hasSelection;
    document.getElementById('installBtn').disabled = !hasSelection;
    document.getElementById('selectAllBtn').disabled = !hasResults;
    document.getElementById('clearSelectionBtn').disabled = !hasSelection;
}

// Download & Install
async function handleDownload() {
    const indexes = Array.from(selectedIndexes);
    if (indexes.length === 0) {
        showToast('Please select packages to download', 'warning');
        return;
    }
    
    showLoading('Adding to download queue...');
    
    try {
        const response = await fetch('/api/download', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ indexes })
        });
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        const data = await response.json();
        if (data.success) {
            showStatus(`Added ${data.added} packages to download queue`, 'success');
            showToast(`✅ Added ${data.added} package(s) to download queue`, 'success');
            selectedIndexes.clear();
            renderResults();
            updateQueueBadge();
            
            // Switch to queue tab if multiple items
            if (data.added > 1) {
                setTimeout(() => switchTab('queue'), 500);
            }
        } else {
            throw new Error(data.error || 'Failed to add to queue');
        }
    } catch (error) {
        // Error logged to status
        showStatus('Failed to add to queue: ' + error.message, 'error');
        showToast('❌ Failed to add to queue: ' + error.message, 'error');
    } finally {
        hideLoading();
    }
}

async function handleInstall() {
    const indexes = Array.from(selectedIndexes);
    if (indexes.length === 0) {
        showToast('Please select packages to install', 'warning');
        return;
    }
    
    // 检查包源支持情况
    const supportedSources = ['npm', 'cargo', 'choco', 'winget'];
    const unsupportedPackages = [];
    const supportedPackages = [];
    
    indexes.forEach(idx => {
        if (idx >= 0 && idx < searchResults.length) {
            const result = searchResults[idx];
            if (supportedSources.includes(result.Source)) {
                supportedPackages.push(result);
            } else {
                unsupportedPackages.push(`${result.Identifier} (${result.Source})`);
            }
        }
    });
    
    // 如果有不支持的包，警告用户
    if (unsupportedPackages.length > 0) {
        const message = `⚠️ The following packages cannot be installed:\n\n${unsupportedPackages.join('\n')}\n\nOnly npm, cargo, choco, and winget packages are supported.\n\n${supportedPackages.length > 0 ? 'Continue with supported packages?' : ''}`;
        
        if (supportedPackages.length === 0) {
            showToast(message, 'warning', 5000);
            return;
        }
        
        if (!confirm(message)) {
            return;
        }
    }
    
    showLoading('Adding to install queue...');
    try {
        const response = await fetch('/api/install', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ indexes })
        });
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        const data = await response.json();
        if (data.success) {
            const addedCount = data.added || 0;
            if (addedCount > 0) {
                const sourceInfo = supportedPackages.length > 1 ? 'packages' : `${supportedPackages[0].Source} package`;
                showStatus(`Added ${addedCount} ${sourceInfo} to install queue`, 'success');
                showToast(`✅ Added ${addedCount} package(s) to install queue`, 'success');
                selectedIndexes.clear();
                renderResults();
                updateQueueBadge();
                
                // Switch to queue tab
                setTimeout(() => switchTab('queue'), 500);
            } else if (unsupportedPackages.length > 0) {
                showStatus('No supported packages selected', 'warning');
                showToast('⚠️ No installable packages selected', 'warning');
            }
        } else {
            throw new Error(data.error || 'Failed to add to queue');
        }
    } catch (error) {
        // Error logged to status
        showStatus('Failed to add to queue: ' + error.message, 'error');
        showToast('❌ Failed to add to queue: ' + error.message, 'error');
    } finally {
        hideLoading();
    }
}

// Queue
async function loadQueue() {
    try {
        const response = await fetch('/api/queue');
        const data = await response.json();
        
        if (data.success) {
            const oldCount = queueTasks.length;
            queueTasks = data.queue || [];
            
            // 如果队列有变化，更新界面
            if (queueTasks.length !== oldCount) {
                // Queue count changed
            }
            
            // 如果有正在运行的任务，更新状态
            const runningTasks = queueTasks.filter(t => t.status === 'running');
            if (runningTasks.length > 0) {
                // Tasks are running;
            }
            
            renderQueue();
            updateQueueBadge();
        }
    } catch (error) {
        // Error logged
    }
}

function renderQueue() {
    if (queueTasks.length === 0) {
        queueContainer.innerHTML = `
            <div class="empty-state">
                <svg width="64" height="64" viewBox="0 0 64 64" fill="none">
                    <rect x="12" y="16" width="40" height="8" rx="2" stroke="currentColor" stroke-width="4" opacity="0.2"/>
                    <rect x="12" y="28" width="40" height="8" rx="2" stroke="currentColor" stroke-width="4" opacity="0.2"/>
                    <rect x="12" y="40" width="40" height="8" rx="2" stroke="currentColor" stroke-width="4" opacity="0.2"/>
                </svg>
                <p>No tasks in queue</p>
            </div>
        `;
        return;
    }
    
    queueContainer.innerHTML = queueTasks.map(task => {
        const statusIcon = getStatusIcon(task.status);
        const statusClass = task.status === 'completed' ? 'badge-success' : 
                           task.status === 'failed' ? 'badge-error' : 
                           task.status === 'running' ? 'badge-warning' : 'badge-info';
        
        // 简化错误信息显示
        let errorDisplay = '';
        if (task.error) {
            let errorMsg = task.error;
            
            // 提取关键错误信息
            if (errorMsg.includes('404') || errorMsg.includes('Not found')) {
                errorMsg = '❌ Package not found (404). Please check the package name and version.';
            } else if (errorMsg.includes('timeout')) {
                errorMsg = '⏱️ Request timeout. Please check your network connection.';
            } else if (errorMsg.includes('network')) {
                errorMsg = '🌐 Network error. Please check your internet connection.';
            } else if (errorMsg.includes('install is not supported for source')) {
                // 提取源名称
                const match = errorMsg.match(/source: (\w+)/);
                const source = match ? match[1] : 'this source';
                errorMsg = `⚠️ Installation not supported for ${source}. Supported sources: npm, cargo, choco, winget.`;
            } else if (errorMsg.includes('only supported for npm')) {
                errorMsg = '⚠️ This package source does not support installation via Web UI. Only npm packages can be installed.';
            } else {
                // 截断过长的错误信息
                errorMsg = truncate(errorMsg, 200);
            }
            
            errorDisplay = `
                <div class="queue-error">
                    <strong>Error:</strong> ${escapeHtml(errorMsg)}
                </div>
            `;
        }
        
        return `
            <div class="queue-item ${task.status}" data-task-id="${task.id}">
                <div class="queue-status">${statusIcon}</div>
                <div class="queue-info">
                    <div class="queue-label">
                        <span class="badge ${statusClass}">${task.status.toUpperCase()}</span>
                        [${task.id}] ${escapeHtml(task.label)}
                    </div>
                    <div class="queue-meta">
                        Type: <strong>${task.type}</strong> • 
                        Created: ${formatTime(task.created_at)}
                    </div>
                    ${errorDisplay}
                    ${task.status === 'running' ? `
                        <div class="progress-bar">
                            <div class="progress-bar-fill" style="width: 100%;"></div>
                        </div>
                    ` : ''}
                </div>
            </div>
        `;
    }).join('');
}

async function clearQueue() {
    if (!confirm('Are you sure you want to clear all completed and failed tasks?')) {
        return;
    }
    
    showLoading('Clearing finished tasks...');
    
    try {
        const response = await fetch('/api/queue', { method: 'DELETE' });
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        const data = await response.json();
        
        if (data.success) {
            showStatus('Cleared finished tasks', 'success');
            showToast('✅ Queue cleaned up', 'success', 2000);
            loadQueue();
        }
    } catch (error) {
        showStatus('Failed to clear queue: ' + error.message, 'error');
        showToast('❌ Failed to clear queue', 'error');
    } finally {
        hideLoading();
    }
}

function updateQueueBadge() {
    const count = queueTasks.length;
    document.getElementById('queueBadge').textContent = count;
    document.getElementById('queueCount').textContent = count;
    
    // Update nav pill badge
    const queuePill = document.querySelector('[data-tab="queue"] .pill-badge');
    if (queuePill) {
        queuePill.textContent = count;
    }
}

// Mirrors
async function handleMirrorTest(e) {
    e.preventDefault();
    
    const source = getCustomSelectValue('mirrorSourceSelect');
    
    showLoading('Testing mirrors...');
    
    try {
        const response = await fetch('/api/mirrors', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ source })
        });
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        const data = await response.json();
        
        if (data.success) {
            mirrorResults = data.results || [];
            renderMirrors();
            showStatus(`Tested ${data.count} mirrors`, 'success');
            showToast(`✅ Tested ${data.count} mirrors`, 'success', 2000);
        }
    } catch (error) {
        showStatus('Mirror test failed: ' + error.message, 'error');
        showToast('❌ Mirror test failed', 'error');
        mirrorResults = [];
        renderMirrors();
    } finally {
        hideLoading();
    }
}

function renderMirrors() {
    if (mirrorResults.length === 0) {
        mirrorResultsContainer.innerHTML = `
            <div class="empty-state">
                <svg width="64" height="64" viewBox="0 0 64 64" fill="none" class="empty-icon">
                    <defs>
                        <linearGradient id="mirror-empty-grad" x1="0" y1="0" x2="64" y2="64">
                            <stop offset="0%" stop-color="currentColor" stop-opacity="0.3"/>
                            <stop offset="100%" stop-color="currentColor" stop-opacity="0.1"/>
                        </linearGradient>
                    </defs>
                    <circle cx="32" cy="32" r="24" stroke="url(#mirror-empty-grad)" stroke-width="4" fill="currentColor" fill-opacity="0.03">
                        <animate attributeName="r" values="24; 26; 24" dur="3s" repeatCount="indefinite"/>
                    </circle>
                    <path d="M32 8v48M8 32h48" stroke="currentColor" stroke-width="4" opacity="0.2"/>
                    <circle cx="32" cy="32" r="4" fill="currentColor" opacity="0.3">
                        <animate attributeName="opacity" values="0.3; 0.6; 0.3" dur="2s" repeatCount="indefinite"/>
                    </circle>
                    <circle cx="32" cy="32" r="16" stroke="currentColor" stroke-width="1" opacity="0.15">
                        <animate attributeName="r" values="16; 20; 16" dur="2s" repeatCount="indefinite"/>
                        <animate attributeName="opacity" values="0.15; 0; 0.15" dur="2s" repeatCount="indefinite"/>
                    </circle>
                </svg>
                <p>Click "Test Mirrors" to start</p>
            </div>
        `;
        return;
    }
    
    // Sort mirrors by latency (fastest first), failed ones last
    const sortedResults = [...mirrorResults].sort((a, b) => {
        if (a.OK && !b.OK) return -1;
        if (!a.OK && b.OK) return 1;
        if (!a.OK && !b.OK) return 0;
        return a.FirstByte - b.FirstByte;
    });
    
    // 计算最快的延迟用于进度条
    const fastestLatency = sortedResults.filter(r => r.OK).reduce((min, r) => 
        r.FirstByte < min ? r.FirstByte : min, Infinity
    );
    
    mirrorResultsContainer.innerHTML = sortedResults.map((result, index) => {
        const latencyMs = result.FirstByte / 1000000; // 转换为毫秒
        const isAvailable = result.OK;
        const latencyText = isAvailable ? `${latencyMs.toFixed(0)}ms` : 'Failed';
        
        // 计算速度等级
        let speedClass = '';
        let speedLabel = '';
        let speedEmoji = '';
        if (isAvailable) {
            if (latencyMs < 100) {
                speedClass = 'speed-excellent';
                speedLabel = 'Excellent';
                speedEmoji = '⚡';
            } else if (latencyMs < 300) {
                speedClass = 'speed-good';
                speedLabel = 'Good';
                speedEmoji = '🚀';
            } else if (latencyMs < 800) {
                speedClass = 'speed-fair';
                speedLabel = 'Fair';
                speedEmoji = '✈️';
            } else {
                speedClass = 'speed-slow';
                speedLabel = 'Slow';
                speedEmoji = '🐌';
            }
        }
        
        // 排名徽章
        let rankBadge = '';
        if (isAvailable && index < 3) {
            const rankColors = ['#FFD700', '#C0C0C0', '#CD7F32']; // 金银铜
            const rankEmojis = ['🥇', '🥈', '🥉'];
            rankBadge = `
                <div class="mirror-rank" style="background: linear-gradient(135deg, ${rankColors[index]}, ${rankColors[index]}99);">
                    ${rankEmojis[index]} TOP ${index + 1}
                </div>
            `;
        }
        
        // 计算进度条宽度（相对于最快的镜像）
        let progressWidth = 0;
        if (isAvailable && fastestLatency !== Infinity) {
            progressWidth = Math.min((fastestLatency / result.FirstByte) * 100, 100);
        }
        
        return `
            <div class="mirror-card ${speedClass} ${!isAvailable ? 'mirror-unavailable' : ''}" data-index="${index}">
                ${rankBadge}
                <div class="mirror-card-header">
                    <div class="mirror-status-indicator ${isAvailable ? 'status-online' : 'status-offline'}">
                        <svg width="12" height="12" viewBox="0 0 12 12" fill="none">
                            <circle cx="6" cy="6" r="5" fill="currentColor">
                                ${isAvailable ? '<animate attributeName="opacity" values="1; 0.5; 1" dur="2s" repeatCount="indefinite"/>' : ''}
                            </circle>
                        </svg>
                    </div>
                    <div class="mirror-source-badge">${result.Source.toUpperCase()}</div>
                    <div class="mirror-speed-badge ${speedClass}">
                        ${speedEmoji} ${speedLabel || 'Offline'}
                    </div>
                </div>
                
                <div class="mirror-card-body">
                    <h3 class="mirror-name">${escapeHtml(result.Name)}</h3>
                    <div class="mirror-url-container">
                        <div class="mirror-url" title="${escapeHtml(result.BaseURL)}">
                            ${escapeHtml(result.BaseURL)}
                        </div>
                        <button class="mirror-copy-btn" data-url="${escapeHtml(result.BaseURL)}" title="Copy URL">
                            <svg width="16" height="16" viewBox="0 0 20 20" fill="none">
                                <rect x="7" y="7" width="10" height="10" rx="2" stroke="currentColor" stroke-width="2"/>
                                <path d="M3 13V5a2 2 0 012-2h8" stroke="currentColor" stroke-width="2"/>
                            </svg>
                        </button>
                    </div>
                </div>
                
                <div class="mirror-card-footer">
                    <div class="mirror-latency-display">
                        <div class="latency-label">Response Time</div>
                        <div class="latency-value ${isAvailable ? 'text-success' : 'text-error'}">
                            ${latencyText}
                        </div>
                    </div>
                    ${isAvailable ? `
                        <div class="mirror-speed-bar">
                            <div class="mirror-speed-bar-bg">
                                <div class="mirror-speed-bar-fill ${speedClass}" style="width: ${progressWidth}%">
                                    <div class="speed-bar-shine"></div>
                                </div>
                            </div>
                        </div>
                    ` : `
                        <div class="mirror-error-msg">
                            <svg width="14" height="14" viewBox="0 0 20 20" fill="none">
                                <circle cx="10" cy="10" r="8" stroke="currentColor" stroke-width="2"/>
                                <path d="M10 6v4M10 14h.01" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
                            </svg>
                            Connection failed
                        </div>
                    `}
                </div>
            </div>
        `;
    }).join('');
    
    // 添加复制功能
    mirrorResultsContainer.querySelectorAll('.mirror-copy-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            e.stopPropagation();
            const url = btn.dataset.url;
            copyToClipboard(url);
        });
    });
    
    // 添加卡片点击复制功能
    mirrorResultsContainer.querySelectorAll('.mirror-card').forEach((card, index) => {
        const result = sortedResults[index];
        card.addEventListener('click', (e) => {
            if (!e.target.closest('.mirror-copy-btn')) {
                copyToClipboard(result.BaseURL);
            }
        });
    });
}

// Status
async function loadStatus() {
    try {
        const response = await fetch('/api/status');
        const data = await response.json();
        
        if (data.success) {
            document.getElementById('versionBadge').textContent = `v${data.version}`;
            document.getElementById('aboutVersion').textContent = data.version;
        }
    } catch (error) {
        // Error logged
    }
}

function showStatus(message, type = 'info') {
    statusText.textContent = message;
    
    // Set color based on type
    const colors = {
        'error': 'var(--google-red)',
        'success': 'var(--google-green)',
        'warning': 'var(--google-yellow)',
        'info': 'var(--gray-700)'
    };
    
    statusText.style.color = colors[type] || colors.info;
    
    // Add a subtle animation
    statusText.style.animation = 'none';
    setTimeout(() => {
        statusText.style.animation = 'statusPulse 0.5s ease-out';
    }, 10);
}

function updateTime() {
    const now = new Date();
    statusTime.textContent = now.toLocaleTimeString();
}

// Loading
function showLoading(message = 'Loading...') {
    loadingText.textContent = message;
    loadingOverlay.style.display = 'flex';
}

function hideLoading() {
    loadingOverlay.style.display = 'none';
}

// Utilities
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text || '';
    return div.innerHTML;
}

function truncate(text, maxLength) {
    if (!text) return '';
    if (text.length <= maxLength) return text;
    return text.substring(0, maxLength) + '...';
}

function getStatusIcon(status) {
    const icons = {
        pending: '⏳',
        running: '🔄',
        completed: '✅',
        failed: '❌'
    };
    return icons[status] || '❓';
}

function formatTime(timestamp) {
    if (!timestamp) return '';
    const date = new Date(timestamp);
    return date.toLocaleString();
}

function formatDuration(nanoseconds) {
    const ms = nanoseconds / 1000000;
    if (ms < 1000) {
        return `${Math.round(ms)}ms`;
    }
    return `${(ms / 1000).toFixed(2)}s`;
}

// Toast Notifications
function showToast(message, type = 'info', duration = 3000) {
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.textContent = message;
    
    document.body.appendChild(toast);
    
    if (duration > 0) {
        setTimeout(() => {
            toast.style.animation = 'toastSlideOut 0.4s cubic-bezier(0.16, 1, 0.3, 1)';
            setTimeout(() => {
                document.body.removeChild(toast);
            }, 400);
        }, duration);
    }
    
    return toast;
}

// Copy to Clipboard
function copyToClipboard(text) {
    if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(text).then(() => {
            showToast('Copied to clipboard!', 'success', 2000);
        }).catch(err => {
            console.error('Failed to copy:', err);
            showToast('Failed to copy to clipboard', 'error');
        });
    } else {
        // Fallback for older browsers
        const textarea = document.createElement('textarea');
        textarea.value = text;
        textarea.style.position = 'fixed';
        textarea.style.opacity = '0';
        document.body.appendChild(textarea);
        textarea.select();
        try {
            document.execCommand('copy');
            showToast('Copied to clipboard!', 'success', 2000);
        } catch (err) {
            console.error('Failed to copy:', err);
            showToast('Failed to copy to clipboard', 'error');
        }
        document.body.removeChild(textarea);
    }
}

// Add toast slide out animation to CSS
const style = document.createElement('style');
style.textContent = `
@keyframes toastSlideOut {
    from {
        opacity: 1;
        transform: translateX(0);
    }
    to {
        opacity: 0;
        transform: translateX(100px);
    }
}
`;
document.head.appendChild(style);
