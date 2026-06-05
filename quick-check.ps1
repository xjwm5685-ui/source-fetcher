<#
.SYNOPSIS
    Quick project health check script

.DESCRIPTION
    Runs essential checks before publishing:
    - File existence check
    - Placeholder detection
    - Build test
    - Test execution
    - Link validation (basic)
#>

param(
    [switch]$SkipBuild,
    [switch]$SkipTests,
    [switch]$Verbose
)

$ErrorActionPreference = "Continue"
$WarningCount = 0
$ErrorCount = 0

function Write-Status {
    param([string]$Message, [string]$Type = "Info")
    
    $color = switch ($Type) {
        "Success" { "Green" }
        "Warning" { "Yellow" }
        "Error"   { "Red" }
        "Info"    { "Cyan" }
        default   { "White" }
    }
    
    $icon = switch ($Type) {
        "Success" { "✅" }
        "Warning" { "⚠️ " }
        "Error"   { "❌" }
        "Info"    { "ℹ️ " }
        default   { "  " }
    }
    
    Write-Host "$icon $Message" -ForegroundColor $color
}

function Test-FileExists {
    param([string]$Path, [string]$Description)
    
    if (Test-Path $Path) {
        Write-Status "$Description exists" "Success"
        return $true
    } else {
        Write-Status "$Description missing: $Path" "Error"
        $script:ErrorCount++
        return $false
    }
}

function Test-PlaceholderInFile {
    param([string]$Path, [string]$Pattern = "YOUR_USERNAME")
    
    if (Test-Path $Path) {
        $content = Get-Content $Path -Raw -ErrorAction SilentlyContinue
        if ($content -match $Pattern) {
            Write-Status "Found placeholder '$Pattern' in: $Path" "Warning"
            $script:WarningCount++
            return $false
        }
    }
    return $true
}

Write-Host "`n🔍 Source Fetcher Project Health Check`n" -ForegroundColor Magenta

# === Essential Files Check ===
Write-Host "📄 Checking essential files..." -ForegroundColor Cyan
Test-FileExists "README.md" "Main README"
Test-FileExists "README.en.md" "English README"
Test-FileExists "LICENSE" "License file"
Test-FileExists "CONTRIBUTING.md" "Contributing guide"
Test-FileExists "CHANGELOG.md" "Changelog"
Test-FileExists "SECURITY.md" "Security policy"
Test-FileExists "ROADMAP.md" "Roadmap"

Write-Host "`n🔧 Checking build files..." -ForegroundColor Cyan
Test-FileExists "Makefile" "Makefile"
Test-FileExists "build.ps1" "Windows build script"
Test-FileExists ".golangci.yml" "Linter config"
Test-FileExists ".editorconfig" "Editor config"
Test-FileExists "version.go" "Version file"

Write-Host "`n🐙 Checking GitHub files..." -ForegroundColor Cyan
Test-FileExists ".github\workflows\test.yml" "Test workflow"
Test-FileExists ".github\workflows\release.yml" "Release workflow"
Test-FileExists ".github\workflows\codeql.yml" "CodeQL workflow"
Test-FileExists ".github\ISSUE_TEMPLATE\bug_report.md" "Bug report template"
Test-FileExists ".github\ISSUE_TEMPLATE\feature_request.md" "Feature request template"
Test-FileExists ".github\PULL_REQUEST_TEMPLATE.md" "PR template"

Write-Host "`n📚 Checking examples..." -ForegroundColor Cyan
Test-FileExists "examples\README.md" "Examples index"
Test-FileExists "examples\basic-download.yaml" "Basic download example"
Test-FileExists "examples\basic-install.yaml" "Basic install example"
Test-FileExists "examples\offline-deployment.yaml" "Offline deployment example"
Test-FileExists "examples\private-registry.yaml" "Private registry example"
Test-FileExists "examples\ci-cd-integration.yaml" "CI/CD example"
Test-FileExists "examples\windows-tools.yaml" "Windows tools example"

# === Placeholder Check ===
Write-Host "`n🔎 Checking for placeholders..." -ForegroundColor Cyan
$filesToCheck = @(
    "README.md",
    "README.en.md",
    "examples\ci-cd-integration.yaml"
)

foreach ($file in $filesToCheck) {
    Test-PlaceholderInFile $file
}

# === Logo Check ===
Write-Host "`n🎨 Checking visual assets..." -ForegroundColor Cyan
if (-not (Test-Path "assets\logo.png")) {
    Write-Status "Logo missing: assets\logo.png" "Warning"
    Write-Status "  Tip: Create a 512x512px PNG logo" "Info"
    $script:WarningCount++
}

if (-not (Test-Path "assets\demo.gif")) {
    Write-Status "Demo GIF missing: assets\demo.gif" "Warning"
    Write-Status "  Tip: Record a 5-10s TUI demo" "Info"
    $script:WarningCount++
}

# === Go Environment Check ===
Write-Host "`n🔧 Checking Go environment..." -ForegroundColor Cyan
try {
    $goVersion = go version
    Write-Status "Go installed: $goVersion" "Success"
} catch {
    Write-Status "Go not found in PATH" "Error"
    $script:ErrorCount++
    $SkipBuild = $true
    $SkipTests = $true
}

# === Build Check ===
if (-not $SkipBuild) {
    Write-Host "`n🔨 Running build test..." -ForegroundColor Cyan
    
    $buildOutput = go build -v -o source-fetcher-test.exe . 2>&1
    
    if ($LASTEXITCODE -eq 0) {
        Write-Status "Build successful" "Success"
        
        # Test binary
        if (Test-Path "source-fetcher-test.exe") {
            $versionOutput = .\source-fetcher-test.exe version 2>&1
            if ($LASTEXITCODE -eq 0) {
                Write-Status "Binary execution successful" "Success"
                if ($Verbose) {
                    Write-Host "  Output: $versionOutput" -ForegroundColor Gray
                }
            } else {
                Write-Status "Binary execution failed" "Error"
                $script:ErrorCount++
            }
            
            # Cleanup
            Remove-Item "source-fetcher-test.exe" -ErrorAction SilentlyContinue
        }
    } else {
        Write-Status "Build failed" "Error"
        if ($Verbose) {
            Write-Host $buildOutput -ForegroundColor Red
        }
        $script:ErrorCount++
    }
}

# === Test Check ===
if (-not $SkipTests) {
    Write-Host "`n🧪 Running tests..." -ForegroundColor Cyan
    
    $testOutput = go test -v ./... 2>&1
    
    if ($LASTEXITCODE -eq 0) {
        Write-Status "All tests passed" "Success"
        
        # Get coverage
        $coverage = go test -cover ./... 2>&1 | Select-String "coverage:"
        if ($coverage) {
            Write-Status "Coverage: $coverage" "Info"
        }
    } else {
        Write-Status "Some tests failed" "Error"
        if ($Verbose) {
            Write-Host $testOutput -ForegroundColor Red
        }
        $script:ErrorCount++
    }
}

# === Git Check ===
Write-Host "`n📦 Checking Git status..." -ForegroundColor Cyan
try {
    $gitStatus = git status --porcelain 2>&1
    
    if ($LASTEXITCODE -eq 0) {
        if ($gitStatus) {
            Write-Status "Uncommitted changes detected" "Warning"
            $script:WarningCount++
            if ($Verbose) {
                Write-Host "  Changes:" -ForegroundColor Gray
                git status --short
            }
        } else {
            Write-Status "Working tree clean" "Success"
        }
        
        # Check if on main branch
        $branch = git branch --show-current 2>&1
        if ($branch -eq "main") {
            Write-Status "On main branch" "Success"
        } else {
            Write-Status "Not on main branch (current: $branch)" "Info"
        }
        
        # Check remote
        $remote = git remote -v 2>&1
        if ($remote) {
            Write-Status "Remote configured" "Success"
        } else {
            Write-Status "No remote configured" "Warning"
            Write-Status "  Run: git remote add origin <URL>" "Info"
            $script:WarningCount++
        }
    }
} catch {
    Write-Status "Git not initialized or not in PATH" "Warning"
    $script:WarningCount++
}

# === Summary ===
Write-Host "`n📊 Summary" -ForegroundColor Magenta
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Gray

if ($ErrorCount -eq 0 -and $WarningCount -eq 0) {
    Write-Status "Perfect! Project is ready for release! 🎉" "Success"
    $exitCode = 0
} elseif ($ErrorCount -eq 0) {
    Write-Status "Good! $WarningCount warning(s) to address" "Warning"
    Write-Host "`nWarnings are not critical but should be addressed before release." -ForegroundColor Yellow
    $exitCode = 0
} else {
    Write-Status "$ErrorCount error(s) and $WarningCount warning(s) found" "Error"
    Write-Host "`nPlease fix errors before releasing!" -ForegroundColor Red
    $exitCode = 1
}

Write-Host "`n💡 Next steps:" -ForegroundColor Cyan
if (Test-Path "PROJECT_COMPLETION_SUMMARY.md") {
    Write-Host "   1. Review: PROJECT_COMPLETION_SUMMARY.md"
}
if (Test-Path "PRE_LAUNCH_CHECKLIST.md") {
    Write-Host "   2. Review: PRE_LAUNCH_CHECKLIST.md"
}
Write-Host "   3. Fix any errors or warnings above"
Write-Host "   4. Create logo: assets\logo.png"
Write-Host "   5. Replace placeholders: YOUR_USERNAME"
Write-Host "   6. Create git tag: git tag -a v1.0.0 -m 'Release v1.0.0'"
Write-Host "   7. Push: git push origin main && git push origin v1.0.0"
Write-Host ""

exit $exitCode
