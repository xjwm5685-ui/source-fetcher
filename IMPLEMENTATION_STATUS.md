# 实现状态 - 未覆盖功能

**更新时间**: 2026-06-06  
**当前版本**: v1.0.0  
**计划版本**: v1.1.0

---

## 📊 功能实现状态

### ✅ v1.0.0 已实现

| 功能 | 状态 | 说明 |
|------|------|------|
| npm 下载 | ✅ 完整 | 支持镜像、认证、断点续传 |
| npm 安装 | ✅ 完整 | 完整依赖树解析和安装 |
| npm 卸载 | ✅ 完整 | 基于清单的精确卸载 |
| npm 修复 | ✅ 完整 | 检测和修复损坏安装 |
| choco 下载 | ✅ 完整 | 下载 .nupkg 文件 |
| choco 安装 | ✅ 基础 | 调用 choco 安装单个包 |
| winget 下载 | ✅ 完整 | 解析清单下载安装器 |
| winget 安装 | ✅ 基础 | 静默安装 exe/msi/msix |
| pip/cargo/maven 下载 | ✅ 完整 | 支持下载 |
| URL 下载 | ✅ 完整 | 任意 URL 直链下载 |
| 批量操作 | ✅ 完整 | YAML 批量下载和安装 |
| TUI 界面 | ✅ 完整 | 终端交互界面 |
| Web GUI | ✅ 完整 | 浏览器界面 |

### 🚧 v1.1.0 开发中

| 功能 | 状态 | 实现进度 |
|------|------|---------|
| choco 依赖安装 | 🚧 设计完成 | 让 choco 自动处理依赖 |
| winget 依赖处理 | 🚧 设计完成 | 安装器自行处理依赖 |
| URL 依赖支持 | ✅ 代码完成 | YAML 声明式依赖 |
| 统一卸载 (choco) | ✅ 代码完成 | 调用 `choco uninstall` |
| 统一卸载 (winget) | ✅ 代码完成 | 调用 `winget uninstall` |
| 统一卸载 (url) | ✅ 代码完成 | 删除跟踪的路径 |
| 跨生态卸载 | ✅ 代码完成 | 统一清单和卸载接口 |
| CLI 集成 | ⏳ 待开始 | 添加命令行参数 |
| 文档更新 | ⏳ 待开始 | README 和示例 |
| 测试 | ⏳ 待开始 | 单元和集成测试 |

---

## 📂 新增文件

### 核心实现
1. **`uninstall_unified.go`** (✅ 已创建)
   - 统一卸载接口
   - 支持 choco/winget/url 卸载
   - 统一清单格式
   - 交互式确认
   - Dry-run 模式

2. **`dependencies_url.go`** (✅ 已创建)
   - URL 依赖声明
   - 依赖下载
   - 依赖跟踪
   - 依赖报告生成

### 文档
3. **`FEATURE_ENHANCEMENT_PLAN.md`** (✅ 已创建)
   - 详细的增强计划
   - 设计决策
   - 实现选项对比
   - 测试策略

4. **`V1.1_ROADMAP.md`** (✅ 已创建)
   - v1.1.0 路线图
   - 使用示例
   - 清单格式说明
   - 开发进度跟踪

5. **`IMPLEMENTATION_STATUS.md`** (✅ 已创建)
   - 本文档
   - 实现状态跟踪

---

## 🎯 README 中提到的"未覆盖"功能

### 原文描述
> **未覆盖**: choco/winget/url 依赖安装、跨生态依赖卸载

### 实现方案

#### 1. choco 依赖安装
**方案**: 让 Chocolatey 自动处理依赖

```bash
# 下载包
sfer download --source choco --name git

# 安装（choco 自动安装依赖）
sfer install --source choco --name git
# 内部执行: choco install git.nupkg --yes
```

**优点**:
- ✅ 简单可靠
- ✅ 与 choco 行为一致
- ✅ 无需维护依赖解析器

**限制**:
- ⚠️ 需要安装 Chocolatey
- ⚠️ 无法预览依赖树
- ⚠️ 依赖记录有限

#### 2. winget 依赖处理
**方案**: 让 Windows 安装器自行处理

```bash
# 下载安装器
sfer download --source winget --name Microsoft.PowerToys

# 安装（安装器处理依赖）
sfer install --source winget --name Microsoft.PowerToys
# 内部执行: PowerToysSetup.exe /silent
```

**优点**:
- ✅ 符合 Windows 生态习惯
- ✅ 安装器自包含所需组件
- ✅ 无需额外依赖解析

**限制**:
- ⚠️ 无法跟踪依赖
- ⚠️ 卸载时可能残留依赖

#### 3. URL 依赖支持
**方案**: YAML 声明式依赖

```yaml
downloads:
  - source: url
    url: https://example.com/app.exe
    dependencies:
      - url: https://example.com/runtime.exe
        name: "Runtime"
        version: "2.0"
      - url: https://example.com/lib.dll
        name: "Library"
```

**优点**:
- ✅ 灵活可控
- ✅ 支持任意依赖关系
- ✅ 可跟踪所有文件

**限制**:
- ⚠️ 需要手动声明
- ⚠️ 无自动发现

#### 4. 跨生态统一卸载
**方案**: 统一清单 + 各生态专用卸载器

```bash
# 卸载 choco 包
sfer uninstall --source choco --name git

# 卸载 winget 包
sfer uninstall --source winget --name Microsoft.PowerToys

# 卸载所有 choco 包
sfer uninstall --all --source choco

# 卸载所有包（所有生态）
sfer uninstall --all

# 交互式卸载
sfer uninstall --all --interactive
```

**清单格式**:
```json
{
  "version": 1,
  "ecosystems": {
    "npm": [...],
    "choco": [...],
    "winget": [...],
    "url": [...]
  }
}
```

**优点**:
- ✅ 统一接口
- ✅ 跨生态管理
- ✅ 精确跟踪

---

## 🔄 从 "未覆盖" 到 "已覆盖"

### 更新前 (v1.0.0 README)
```markdown
未覆盖：choco/winget/url 依赖安装、跨生态依赖卸载
```

### 更新后 (v1.1.0 README)
```markdown
✅ choco 依赖安装 - 由 Chocolatey 自动处理
✅ winget 依赖处理 - 由 Windows 安装器处理
✅ URL 依赖支持 - YAML 声明式依赖
✅ 跨生态统一卸载 - 支持 npm/choco/winget/url
```

---

## 📋 v1.1.0 发布前检查清单

### 代码完成度
- [x] 统一卸载核心代码
- [x] URL 依赖核心代码
- [ ] CLI 命令集成
- [ ] 错误处理完善
- [ ] 日志输出优化

### 测试
- [ ] 单元测试 (目标 80%+)
- [ ] 集成测试
- [ ] 手动测试所有场景
- [ ] Windows 10/11 测试
- [ ] 不同版本 choco/winget 测试

### 文档
- [ ] README.md 更新
- [ ] 创建 UNINSTALL_GUIDE.md
- [ ] 创建 DEPENDENCY_GUIDE.md
- [ ] examples/ 新示例
- [ ] CHANGELOG.md 更新

### 发布准备
- [ ] 版本号更新
- [ ] Git tag v1.1.0
- [ ] Release notes
- [ ] 二进制文件构建
- [ ] 迁移指南 (v1.0 → v1.1)

---

## 🚀 下一步行动

### 本周任务
1. **集成到 main.go**
   - 添加 `uninstall --all` 命令
   - 添加 `--source` 过滤
   - 添加 `--interactive` 模式

2. **URL 依赖集成**
   - 修改 batch 命令支持依赖
   - 更新 YAML 解析
   - 添加验证逻辑

3. **测试框架**
   - 创建测试用例
   - Mock choco/winget 调用
   - 添加 CI 测试

### 下周任务
1. 文档编写
2. 示例创建
3. Bug 修复
4. 性能优化

### 发布前
1. 完整测试通过
2. 文档审核
3. 社区反馈
4. Release 准备

---

## 💡 设计考虑

### 为什么不自己实现 choco/winget 依赖解析？

**原因**:
1. **复杂度高** - NuSpec 解析、版本约束求解
2. **维护成本** - 需要跟踪上游变化
3. **行为差异** - 可能与官方工具不一致
4. **收益有限** - 用户通常已安装这些工具

**v1.1.0 策略**: 使用原生工具
**未来考虑**: v1.2.0 可选高级模式

### 为什么 URL 依赖需要手动声明？

**原因**:
1. **无标准格式** - URL 文件无统一依赖声明
2. **场景多样** - 不同应用依赖关系不同
3. **用户最清楚** - 用户知道需要哪些文件

**未来可能**: 支持读取元数据文件（如 .deps.json）

---

## 📞 反馈

如有疑问或建议，请：
- 开 GitHub Issue
- 发起 Discussion
- 邮件: ckkhua89@gmail.com

---

**文档状态**: ✅ 最新  
**下次更新**: 集成完成后  
**目标**: v1.1.0 Q3 2026
