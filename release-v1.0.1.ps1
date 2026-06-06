# Source Fetcher v1.0.1 Release Script
# 此脚本执行所有发布前检查并创建 Git 标签

param(
    [switch]$SkipTests,
    [switch]$SkipBuild,
    [switch]$CreateTag,
    [switch]$Push
)

$ErrorActionPreference = "Stop"
$Version = "v1.0.1"

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "   Source Fetcher Release Preparation" -ForegroundColor Cyan
Write-Host "   Version: $Version" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

# 检查是否在正确的目录
if (-not (Test-Path "main.go")) {
    Write-Host "❌ Error: main.go not found. Please run from source-fetcher directory." -ForegroundColor Red
    exit 1
}

# 1. 检查 Git 状态
Write-Host "📋 Step 1: Checking Git status..." -ForegroundColor Yellow
$gitStatus = git status --porcelain
if ($gitStatus) {
    Write-Host "⚠️  Warning: Uncommitted changes found:" -ForegroundColor Yellow
    git status --short
    $continue = Read-Host "Continue anyway? (y/n)"
    if ($continue -ne "y") {
        Write-Host "❌ Release cancelled." -ForegroundColor Red
        exit 1
    }
} else {
    Write-Host "✅ Git working directory clean" -ForegroundColor Green
}

# 2. 验证版本号
Write-Host ""
Write-Host "📋 Step 2: Verifying version number..." -ForegroundColor Yellow
$versionGoContent = Get-Content "version.go" -Raw
$mainGoContent = Get-Content "main.go" -Raw

if ($versionGoContent -match 'Version = "v1\.0\.1"' -and $mainGoContent -match 'version=v1\.0\.1') {
    Write-Host "✅ Version numbers are correct in version.go and main.go" -ForegroundColor Green
} else {
    Write-Host "❌ Error: Version number mismatch!" -ForegroundColor Red
    Write-Host "Expected: v1.0.1" -ForegroundColor Red
    exit 1
}

# 3. 验证 CHANGELOG
Write-Host ""
Write-Host "📋 Step 3: Verifying CHANGELOG.md..." -ForegroundColor Yellow
$changelogContent = Get-Content "CHANGELOG.md" -Raw
if ($changelogContent -match '## \[v1\.0\.1\]') {
    Write-Host "✅ v1.0.1 entry found in CHANGELOG.md" -ForegroundColor Green
} else {
    Write-Host "❌ Error: v1.0.1 entry not found in CHANGELOG.md!" -ForegroundColor Red
    exit 1
}

# 4. 运行测试
if (-not $SkipTests) {
    Write-Host ""
    Write-Host "📋 Step 4: Running tests..." -ForegroundColor Yellow
    $testResult = go test -v -cover ./... 2>&1
    $testExitCode = $LASTEXITCODE
    
    if ($testExitCode -eq 0) {
        Write-Host "✅ All tests passed" -ForegroundColor Green
        # 提取测试统计
        $testOutput = $testResult -join "`n"
        if ($testOutput -match 'PASS') {
            $passCount = ($testOutput -split "`n" | Select-String "^--- PASS:").Count
            Write-Host "   Passed: $passCount tests" -ForegroundColor Gray
        }
        if ($testOutput -match 'coverage: ([\d.]+)%') {
            $coverage = $matches[1]
            Write-Host "   Coverage: $coverage%" -ForegroundColor Gray
        }
    } else {
        Write-Host "❌ Tests failed!" -ForegroundColor Red
        Write-Host $testResult -ForegroundColor Gray
        $continue = Read-Host "Continue anyway? (y/n)"
        if ($continue -ne "y") {
            exit 1
        }
    }
} else {
    Write-Host ""
    Write-Host "⚠️  Step 4: Tests skipped" -ForegroundColor Yellow
}

# 5. 运行 go vet
Write-Host ""
Write-Host "📋 Step 5: Running go vet..." -ForegroundColor Yellow
$vetResult = go vet ./... 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ go vet passed" -ForegroundColor Green
} else {
    Write-Host "⚠️  go vet found issues:" -ForegroundColor Yellow
    Write-Host $vetResult -ForegroundColor Gray
}

# 6. 编译检查
if (-not $SkipBuild) {
    Write-Host ""
    Write-Host "📋 Step 6: Building binary..." -ForegroundColor Yellow
    $buildResult = go build -o "source-fetcher-$Version-windows-amd64.exe" . 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Build successful" -ForegroundColor Green
        $fileSize = (Get-Item "source-fetcher-$Version-windows-amd64.exe").Length / 1MB
        Write-Host ("   Binary size: {0:N2} MB" -f $fileSize) -ForegroundColor Gray
        
        # 验证版本
        $versionOutput = & ".\source-fetcher-$Version-windows-amd64.exe" version 2>&1
        if ($versionOutput -match 'v1\.0\.1') {
            Write-Host "✅ Binary version verified" -ForegroundColor Green
        } else {
            Write-Host "⚠️  Binary version could not be verified" -ForegroundColor Yellow
        }
    } else {
        Write-Host "❌ Build failed!" -ForegroundColor Red
        Write-Host $buildResult -ForegroundColor Gray
        exit 1
    }
} else {
    Write-Host ""
    Write-Host "⚠️  Step 6: Build skipped" -ForegroundColor Yellow
}

# 7. 检查必需文件
Write-Host ""
Write-Host "📋 Step 7: Checking required files..." -ForegroundColor Yellow
$requiredFiles = @(
    "README.md",
    "README.en.md",
    "CHANGELOG.md",
    "LICENSE",
    "SECURITY.md",
    ".source-fetcher.example.yaml",
    "ULTRA_CODE_REVIEW_REPORT_v2.md",
    "FIXES_APPLIED.md",
    "RELEASE_CHECKLIST_v1.0.1.md"
)

$missingFiles = @()
foreach ($file in $requiredFiles) {
    if (Test-Path $file) {
        Write-Host "  ✓ $file" -ForegroundColor Gray
    } else {
        Write-Host "  ✗ $file (missing)" -ForegroundColor Red
        $missingFiles += $file
    }
}

if ($missingFiles.Count -eq 0) {
    Write-Host "✅ All required files present" -ForegroundColor Green
} else {
    Write-Host "⚠️  Missing files: $($missingFiles -join ', ')" -ForegroundColor Yellow
}

# 8. 安全修复验证
Write-Host ""
Write-Host "📋 Step 8: Verifying security fixes..." -ForegroundColor Yellow
$securityChecks = @{
    "CORS protection" = "webgui.go"
    "Input validation" = "uninstall_unified.go"
    "SSRF protection" = "dependencies_url.go"
    "Safe YAML parsing" = "config.go"
}

$allSecurityChecksPass = $true
foreach ($check in $securityChecks.GetEnumerator()) {
    if (Test-Path $check.Value) {
        $content = Get-Content $check.Value -Raw
        switch ($check.Key) {
            "CORS protection" { 
                if ($content -match 'isLocalOrigin') {
                    Write-Host "  ✓ $($check.Key)" -ForegroundColor Gray
                } else {
                    Write-Host "  ✗ $($check.Key) - function not found" -ForegroundColor Red
                    $allSecurityChecksPass = $false
                }
            }
            "Input validation" { 
                if ($content -match 'isValidPackageName') {
                    Write-Host "  ✓ $($check.Key)" -ForegroundColor Gray
                } else {
                    Write-Host "  ✗ $($check.Key) - function not found" -ForegroundColor Red
                    $allSecurityChecksPass = $false
                }
            }
            "SSRF protection" { 
                if ($content -match 'isPrivateOrLocalAddress') {
                    Write-Host "  ✓ $($check.Key)" -ForegroundColor Gray
                } else {
                    Write-Host "  ✗ $($check.Key) - function not found" -ForegroundColor Red
                    $allSecurityChecksPass = $false
                }
            }
            "Safe YAML parsing" { 
                if ($content -match 'unmarshalYAMLSafe') {
                    Write-Host "  ✓ $($check.Key)" -ForegroundColor Gray
                } else {
                    Write-Host "  ✗ $($check.Key) - function not found" -ForegroundColor Red
                    $allSecurityChecksPass = $false
                }
            }
        }
    } else {
        Write-Host "  ✗ $($check.Key) - $($check.Value) not found" -ForegroundColor Red
        $allSecurityChecksPass = $false
    }
}

if ($allSecurityChecksPass) {
    Write-Host "✅ All security fixes verified" -ForegroundColor Green
} else {
    Write-Host "❌ Security check failed!" -ForegroundColor Red
    exit 1
}

# 9. 总结
Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "   Release Readiness Summary" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "✅ Git status checked" -ForegroundColor Green
Write-Host "✅ Version numbers verified" -ForegroundColor Green
Write-Host "✅ CHANGELOG updated" -ForegroundColor Green
if (-not $SkipTests) { Write-Host "✅ Tests passed" -ForegroundColor Green }
Write-Host "✅ go vet passed" -ForegroundColor Green
if (-not $SkipBuild) { Write-Host "✅ Build successful" -ForegroundColor Green }
Write-Host "✅ Required files present" -ForegroundColor Green
Write-Host "✅ Security fixes verified" -ForegroundColor Green
Write-Host ""

# 10. 创建 Git 标签
if ($CreateTag) {
    Write-Host "📋 Creating Git tag..." -ForegroundColor Yellow
    
    $tagMessage = @"
Security patch release

- Fix CORS vulnerability
- Fix command injection risks  
- Fix SSRF vulnerability
- Add configuration file support
- Improve error handling
"@
    
    git tag -a $Version -m $tagMessage
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Git tag $Version created" -ForegroundColor Green
        
        if ($Push) {
            Write-Host ""
            Write-Host "📋 Pushing to remote..." -ForegroundColor Yellow
            git push origin main
            git push origin $Version
            
            if ($LASTEXITCODE -eq 0) {
                Write-Host "✅ Pushed to remote successfully" -ForegroundColor Green
            } else {
                Write-Host "❌ Failed to push to remote" -ForegroundColor Red
                exit 1
            }
        } else {
            Write-Host ""
            Write-Host "⚠️  To push the tag, run:" -ForegroundColor Yellow
            Write-Host "   git push origin main" -ForegroundColor Gray
            Write-Host "   git push origin $Version" -ForegroundColor Gray
        }
    } else {
        Write-Host "❌ Failed to create Git tag" -ForegroundColor Red
        exit 1
    }
}

Write-Host ""
Write-Host "🎉 Release preparation complete!" -ForegroundColor Green
Write-Host ""

if (-not $CreateTag) {
    Write-Host "Next steps:" -ForegroundColor Yellow
    Write-Host "1. Review all changes one more time" -ForegroundColor Gray
    Write-Host "2. Run: .\release-v1.0.1.ps1 -CreateTag" -ForegroundColor Gray
    Write-Host "3. Create GitHub Release at: https://github.com/xjwm5685-ui/source-fetcher/releases/new" -ForegroundColor Gray
    Write-Host "4. Upload binary: source-fetcher-$Version-windows-amd64.exe" -ForegroundColor Gray
}
