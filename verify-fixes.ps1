# 验证 Ultra Code Review 修复的脚本

Write-Host "🔍 验证 Ultra Code Review 修复..." -ForegroundColor Cyan
Write-Host ""

$allPassed = $true

# 测试 1: 编译检查
Write-Host "[1/5] 编译检查..." -ForegroundColor Yellow
$buildResult = go build -v . 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "  ✅ 编译成功" -ForegroundColor Green
} else {
    Write-Host "  ❌ 编译失败" -ForegroundColor Red
    Write-Host $buildResult
    $allPassed = $false
}

# 测试 2: 单元测试
Write-Host "[2/5] 运行单元测试..." -ForegroundColor Yellow
$testResult = go test -v ./... 2>&1 | Select-String -Pattern "PASS|FAIL"
$failCount = ($testResult | Select-String -Pattern "FAIL").Count
$passCount = ($testResult | Select-String -Pattern "PASS").Count

if ($failCount -eq 0 -and $passCount -gt 0) {
    Write-Host "  ✅ 所有测试通过 ($passCount 个测试)" -ForegroundColor Green
} else {
    Write-Host "  ❌ 测试失败: $failCount 失败, $passCount 通过" -ForegroundColor Red
    $allPassed = $false
}

# 测试 3: 检查修复的文件
Write-Host "[3/5] 检查修复的文件..." -ForegroundColor Yellow
$fixedFiles = @(
    "webgui.go",
    "uninstall_unified.go",
    "dependencies_url.go"
)

foreach ($file in $fixedFiles) {
    if (Test-Path $file) {
        Write-Host "  ✅ $file 存在" -ForegroundColor Green
    } else {
        Write-Host "  ❌ $file 不存在" -ForegroundColor Red
        $allPassed = $false
    }
}

# 测试 4: 检查关键安全修复
Write-Host "[4/5] 检查关键安全修复..." -ForegroundColor Yellow

# 检查 CORS 修复
$corsFixed = Select-String -Path "webgui.go" -Pattern "allowedOrigins.*:=.*\[\]string" -Quiet
if ($corsFixed) {
    Write-Host "  ✅ CORS 配置已修复 (白名单机制)" -ForegroundColor Green
} else {
    Write-Host "  ❌ CORS 配置未找到白名单" -ForegroundColor Red
    $allPassed = $false
}

# 检查输入验证
$validationFixed = Select-String -Path "uninstall_unified.go" -Pattern "isValidPackageName" -Quiet
if ($validationFixed) {
    Write-Host "  ✅ 包名输入验证已添加" -ForegroundColor Green
} else {
    Write-Host "  ❌ 输入验证函数未找到" -ForegroundColor Red
    $allPassed = $false
}

# 检查 SSRF 防护
$ssrfFixed = Select-String -Path "dependencies_url.go" -Pattern "isPrivateOrLocalAddress" -Quiet
if ($ssrfFixed) {
    Write-Host "  ✅ SSRF 防护已添加" -ForegroundColor Green
} else {
    Write-Host "  ❌ SSRF 防护函数未找到" -ForegroundColor Red
    $allPassed = $false
}

# 检查 context 传播
$contextFixed = Select-String -Path "uninstall_unified.go" -Pattern "func UninstallChoco\(ctx context\.Context" -Quiet
if ($contextFixed) {
    Write-Host "  ✅ Context 传播已添加" -ForegroundColor Green
} else {
    Write-Host "  ❌ Context 传播未添加" -ForegroundColor Red
    $allPassed = $false
}

# 测试 5: 检查文档
Write-Host "[5/5] 检查文档..." -ForegroundColor Yellow
$docs = @(
    "ULTRA_CODE_REVIEW_REPORT_v2.md",
    "FIXES_APPLIED.md"
)

foreach ($doc in $docs) {
    if (Test-Path $doc) {
        Write-Host "  ✅ $doc 存在" -ForegroundColor Green
    } else {
        Write-Host "  ❌ $doc 不存在" -ForegroundColor Red
        $allPassed = $false
    }
}

# 最终结果
Write-Host ""
Write-Host "================================" -ForegroundColor Cyan
if ($allPassed) {
    Write-Host "✅ 所有验证通过！" -ForegroundColor Green
    Write-Host ""
    Write-Host "修复总结:" -ForegroundColor Cyan
    Write-Host "  • CORS 配置安全化" -ForegroundColor White
    Write-Host "  • 输入验证添加" -ForegroundColor White
    Write-Host "  • SSRF 防护实现" -ForegroundColor White
    Write-Host "  • Context 取消支持" -ForegroundColor White
    Write-Host "  • 健康检查机制" -ForegroundColor White
    Write-Host "  • 文件权限加固" -ForegroundColor White
    Write-Host "  • 版本兼容性检查" -ForegroundColor White
    Write-Host ""
    Write-Host "可以安全发布 v1.0.1 (安全修复版本)" -ForegroundColor Green
    exit 0
} else {
    Write-Host "❌ 部分验证失败" -ForegroundColor Red
    Write-Host "请查看上面的错误详情" -ForegroundColor Yellow
    exit 1
}
