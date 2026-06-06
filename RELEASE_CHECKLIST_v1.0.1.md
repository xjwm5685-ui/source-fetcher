# 🚀 Release Checklist v1.0.1

## 版本信息
- **版本号**: v1.0.1
- **发布类型**: Security Patch Release + Feature Enhancement
- **发布日期**: 2026-06-06
- **优先级**: HIGH (包含关键安全修复)

---

## ✅ 发布前检查清单

### 代码质量
- [x] 所有严重安全问题已修复 (3/3)
- [x] 编译成功无错误
- [x] 核心功能测试通过 (32/32)
- [x] 无明显的代码警告
- [x] golangci-lint 检查通过
- [ ] 运行完整测试套件: `go test -v -race -cover ./...`

### 安全修复验证
- [x] CORS 配置仅允许本地 Origin
- [x] 包名输入验证已添加
- [x] SSRF 防护已实现
- [x] 文件权限已加固 (0600/0640)
- [x] YAML 解析使用安全方法
- [x] Context 取消支持已添加
- [x] winget 错误处理已修复

### 功能增强验证
- [x] Web GUI 健康检查机制工作正常
- [x] 配置文件系统已添加
- [x] 清单版本兼容性检查已实现
- [ ] 配置文件示例已测试

### 文档
- [x] CHANGELOG.md 已更新
- [x] 安全修复已记录
- [x] 新功能已记录
- [x] 配置文件示例已提供
- [ ] README.md 需要添加配置文件说明
- [ ] 安全建议已添加到文档

### 构建和打包
- [ ] 本地构建成功 (Windows/Linux/macOS)
- [ ] 版本号已更新到 v1.0.1
- [ ] 二进制文件已测试
- [ ] Release notes 已准备

### Git 操作
- [ ] 所有更改已提交
- [ ] Commit message 符合规范
- [ ] 创建 v1.0.1 标签
- [ ] 推送到远程仓库

---

## 📝 Release Notes 草稿

### v1.0.1 - Security Patch Release

**发布日期**: 2026-06-06

#### 🚨 关键安全修复

**所有用户应立即升级到此版本**

1. **CORS 漏洞修复 (CVE-待分配)**
   - 修复了 Web GUI 接受任意 Origin 的 CSRF 漏洞
   - 现在仅允许 localhost 和 127.0.0.1
   - 影响: 防止远程攻击者通过恶意网站控制本地实例

2. **命令注入防护**
   - 为 choco/winget 卸载添加了包名验证
   - 阻止包含特殊字符的恶意包名
   - 影响: 防止命令注入攻击

3. **SSRF 防护**
   - URL 依赖功能现在阻止私有 IP 地址
   - 防护范围: localhost, 10.x, 172.16-31.x, 192.168.x, 169.254.x
   - 影响: 防止服务器端请求伪造攻击

#### ✨ 新功能

- 添加全局配置文件支持 (`.source-fetcher.yaml`)
- Web GUI 启动现在使用健康检查机制
- 卸载操作支持 Ctrl+C 取消
- 清单文件版本兼容性检查

#### 🐛 Bug 修复

- 修复 winget 卸载总是返回成功的问题
- 文件权限加固 (0644 → 0600/0640)
- YAML 解析明确使用安全 decoder

#### 📚 文档

- 添加配置文件示例
- 更新 CHANGELOG.md
- 添加安全审查报告

#### 💔 破坏性变更

无

#### ⬆️ 升级说明

从 v1.0.0 升级到 v1.0.1:

1. 下载新的二进制文件
2. 替换旧版本
3. (可选) 创建 `.source-fetcher.yaml` 配置文件

配置文件是可选的，不创建配置文件也能正常使用。

---

## 🔧 构建命令

### Windows
```powershell
$env:GOOS="windows"; $env:GOARCH="amd64"; go build -ldflags="-s -w -X main.version=v1.0.1" -o source-fetcher-v1.0.1-windows-amd64.exe .
```

### Linux
```bash
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.version=v1.0.1" -o source-fetcher-v1.0.1-linux-amd64 .
```

### macOS
```bash
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.version=v1.0.1" -o source-fetcher-v1.0.1-darwin-amd64 .
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X main.version=v1.0.1" -o source-fetcher-v1.0.1-darwin-arm64 .
```

---

## 📦 发布流程

1. **准备阶段**
   ```bash
   # 运行所有检查
   go test -v -race -cover ./...
   go vet ./...
   golangci-lint run
   
   # 构建所有平台
   .\build.ps1
   ```

2. **Git 操作**
   ```bash
   git add .
   git commit -m "chore: release v1.0.1 - security patch"
   git tag -a v1.0.1 -m "Security patch release
   
   - Fix CORS vulnerability
   - Fix command injection risks
   - Fix SSRF vulnerability
   - Add configuration file support
   - Improve error handling"
   
   git push origin main
   git push origin v1.0.1
   ```

3. **GitHub Release**
   - 转到 GitHub Releases 页面
   - 点击 "Draft a new release"
   - 选择 tag v1.0.1
   - 标题: "v1.0.1 - Security Patch Release"
   - 复制上面的 Release Notes
   - 上传构建的二进制文件
   - 勾选 "Set as latest release"
   - 点击 "Publish release"

4. **发布后验证**
   - 下载发布的二进制文件
   - 验证版本号: `./source-fetcher-v1.0.1 version`
   - 测试基本功能
   - 验证安全修复生效

---

## 🔔 通知

发布后需要通知:
- [ ] GitHub 用户 (通过 Release)
- [ ] 更新 README badges
- [ ] 社交媒体发布 (可选)
- [ ] 安全公告 (如果需要 CVE)

---

## 📊 发布后监控

- [ ] 监控下载量
- [ ] 收集用户反馈
- [ ] 关注 GitHub Issues
- [ ] 检查是否有新的安全报告

---

## 🎯 v1.1.0 准备

在 v1.0.1 稳定后:
- [ ] 完成 v1.1.0 CLI 集成
- [ ] 添加 v1.1.0 功能测试
- [ ] 准备 v1.1.0 发布

---

**准备完成**: ⏰ 待完成最终检查
**预计发布时间**: 完成所有检查项后立即发布
