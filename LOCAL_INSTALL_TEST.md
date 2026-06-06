# 本地安装测试指南

在推送到 GitHub 之前，你可以在本地完整测试安装流程。

## 🧪 方式一：直接测试本地安装脚本（推荐）

最简单的方式，无需任何额外设置：

```powershell
# 1. 确保已编译
go build -o source-fetcher.exe

# 2. 运行本地安装脚本
.\install-local.ps1

# 3. 重启终端后测试
sfer version
sfer search --source npm --query react
```

### 卸载测试

```powershell
# 运行卸载
& "$env:LOCALAPPDATA\source-fetcher\uninstall.ps1"

# 或手动删除
Remove-Item -Recurse -Force "$env:LOCALAPPDATA\source-fetcher"
```

## 🌐 方式二：模拟 GitHub 安装流程

如果想测试完整的网络安装流程：

### 步骤 1: 启动本地 HTTP 服务器

```powershell
# 在 source-fetcher 目录打开第一个终端
cd d:\dy\source-fetcher

# 使用 Python 启动 HTTP 服务器
python -m http.server 8000

# 或使用 Node.js
# npx http-server -p 8000

# 或使用 PowerShell（需要 PS 5.1+）
# 在另一个脚本中实现简单 HTTP 服务器
```

### 步骤 2: 测试安装

在另一个 PowerShell 窗口：

```powershell
# 测试脚本下载（使用 test-install.ps1）
irm http://localhost:8000/test-install.ps1 | iex

# 验证
sfer version
```

## 📦 方式三：完整模拟 GitHub Release

### 步骤 1: 编译发布版本

```powershell
# 64位版本
$env:GOOS="windows"; $env:GOARCH="amd64"
go build -ldflags "-s -w -X main.version=1.0.2" -o source-fetcher-windows-amd64.exe

# 32位版本（可选）
$env:GOOS="windows"; $env:GOARCH="386"
go build -ldflags "-s -w -X main.version=1.0.2" -o source-fetcher-windows-386.exe

# 重置
$env:GOOS=""; $env:GOARCH=""
```

### 步骤 2: 创建本地 Release 目录

```powershell
# 创建模拟 GitHub Release 结构
mkdir releases\download\v1.0.2
copy source-fetcher-windows-amd64.exe releases\download\v1.0.2\
copy source-fetcher-windows-386.exe releases\download\v1.0.2\
```

### 步骤 3: 修改安装脚本测试版本

创建 `install-local-mock.ps1`（修改下载 URL 为本地）：

```powershell
# 在原 install.ps1 基础上修改
$downloadUrl = "http://localhost:8000/releases/download/v1.0.2/source-fetcher-windows-amd64.exe"
```

### 步骤 4: 启动 HTTP 服务器并测试

```powershell
# 终端 1: 启动服务器
python -m http.server 8000

# 终端 2: 测试安装
irm http://localhost:8000/install-local-mock.ps1 | iex
```

## ✅ 测试清单

完成以下测试确保一切正常：

### 基础功能测试

- [ ] `sfer version` - 显示版本号
- [ ] `sfer --help` - 显示帮助信息
- [ ] `sfer search --source npm --query react` - 搜索功能
- [ ] `sfer download --source npm --name react` - 下载功能
- [ ] `sfer install --source cargo --name serde` - Cargo 安装

### 安装测试

- [ ] 安装脚本正常运行
- [ ] 文件复制到正确位置
- [ ] PATH 环境变量配置成功
- [ ] `sfer` 别名创建成功
- [ ] 卸载脚本已生成

### 环境测试

- [ ] 重启终端后 `sfer` 命令可用
- [ ] 在任意目录都能使用 `sfer`
- [ ] 多个终端窗口都能使用

### 卸载测试

- [ ] 卸载脚本正常运行
- [ ] 程序文件已删除
- [ ] PATH 已清理
- [ ] 可以重新安装

## 🐛 常见问题

### Q: PowerShell 执行策略限制

```powershell
# 临时允许执行脚本
Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass

# 或永久设置（不推荐）
Set-ExecutionPolicy -Scope CurrentUser -ExecutionPolicy RemoteSigned
```

### Q: HTTP 服务器启动失败

```powershell
# 检查端口是否被占用
netstat -ano | findstr :8000

# 使用其他端口
python -m http.server 8080
```

### Q: 找不到 Python

```powershell
# 使用 Node.js
npx http-server -p 8000

# 或安装 Python
winget install Python.Python.3
```

### Q: 权限不足

```powershell
# 以管理员身份运行 PowerShell
Start-Process powershell -Verb runAs
```

## 📝 测试记录模板

使用此模板记录测试结果：

```
测试日期: 2024-XX-XX
测试人员: XXX
版本: 1.0.2

[ ] 编译成功
[ ] 本地安装成功
[ ] 命令可用
[ ] 功能正常
[ ] 卸载成功
[ ] 重新安装成功

问题记录:
- 

改进建议:
- 
```

## 🎯 测试成功后的下一步

如果所有测试都通过：

1. **准备推送到 GitHub**
   - 参考 [GITHUB_PUSH_GUIDE.md](./GITHUB_PUSH_GUIDE.md)

2. **创建 Release**
   - 使用测试通过的编译文件
   - 编写清晰的 Release Notes

3. **验证线上安装**
   ```powershell
   irm https://raw.githubusercontent.com/.../install.ps1 | iex
   ```

## 💡 提示

- 先在本地充分测试，再推送到 GitHub
- 使用虚拟机或沙箱环境测试更安全
- 记录所有测试过程和结果
- 收集用户反馈持续改进

---

**开始本地测试吧！** 🚀
