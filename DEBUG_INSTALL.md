# 安装问题调试指南

## 🔍 添加的调试日志

我已经在代码中添加了详细的调试日志，现在可以看到：

### 服务器端日志会显示

1. **接收安装请求时**：
```
Install request received: indexes=[0]
Current search results count: 5
Adding install task for index 0: Source=npm, Identifier=openclaw, Version=2026.6.1
Added 1 install tasks to queue
```

2. **执行安装任务时**：
```
Executing install: Source=npm, Identifier=openclaw, Version=2026.6.1
InstallRequest: Name=openclaw, Version=2026.6.1, OutputDir=.
```

3. **如果失败**：
```
Task 1 failed: failed to resolve install plan: fetch npm metadata for openclaw@2026.6.1 via npmjs: http 404: {"error":"Not found"}
```

## 📝 测试步骤

### 1. 重新编译

```bash
cd d:\dy\source-fetcher
go build -o source-fetcher.exe
```

### 2. 启动服务（注意观察终端）

```bash
.\source-fetcher.exe gui
```

### 3. 在浏览器中测试

1. 打开 http://localhost:8765
2. 搜索 `openclaw`
3. 按 F12 打开浏览器控制台
4. 选择搜索结果
5. 点击安装

### 4. 观察两个地方的日志

#### A. 浏览器控制台
```javascript
Installing packages with indexes: [0]
Install response: {success: true, added: 1}
```

#### B. 服务器终端（重要！）
应该看到类似：
```
Install request received: indexes=[0]
Current search results count: X
Adding install task for index 0: Source=npm, Identifier=openclaw, Version=2026.6.1
Added 1 install tasks to queue
Executing install: Source=npm, Identifier=openclaw, Version=2026.6.1
InstallRequest: Name=openclaw, Version=2026.6.1, OutputDir=.
```

## 🔍 可能的问题

### 问题 1: Version 字段不正确

如果日志显示：
```
InstallRequest: Name=openclaw, Version=, OutputDir=.
```

说明 `Version` 字段为空，这可能是搜索结果没有正确设置版本。

**解决方法**：检查搜索结果的 `Version` 字段

### 问题 2: 包确实不存在

你说包存在，让我们验证一下：

```bash
# 方法 1: 使用 npm CLI
npm view openclaw

# 方法 2: 使用 curl
curl https://registry.npmjs.org/openclaw

# 方法 3: 访问网页
# 打开 https://www.npmjs.com/package/openclaw
```

### 问题 3: 搜索结果被覆盖

如果多次搜索，`searchResults` 可能被新的搜索覆盖。

**解决方法**：
- 搜索后立即安装
- 不要在安装前再次搜索

## 🎯 完整测试流程

```bash
# 1. 编译
go build -o source-fetcher.exe

# 2. 启动（保持终端可见）
.\source-fetcher.exe gui

# 3. 在浏览器中：
#    - 搜索: openclaw
#    - 源: npm
#    - 等待结果
#    - 立即勾选第一个
#    - 立即点击安装
#    - 切换到 Queue 标签

# 4. 观察服务器终端输出
```

## 📋 需要的信息

请提供以下信息帮助调试：

1. **搜索结果**
   - 搜索 `openclaw` 后看到几个结果？
   - 第一个结果的版本号是什么？

2. **服务器日志**
   - 重新编译后，启动服务
   - 执行安装操作
   - 复制终端中的完整日志

3. **包验证**
   ```bash
   npm view openclaw
   ```
   - 这个命令的输出是什么？

4. **浏览器控制台**
   - 安装时控制台显示什么？

## 🔧 临时解决方法

如果 `openclaw` 有问题，先用确定存在的包测试：

```
搜索: lodash
源: npm
选择第一个
点击安装
```

这样可以确认：
- ✅ 代码逻辑是否正确
- ✅ 安装流程是否正常
- ✅ 是包的问题还是代码的问题

---

**请重新编译并测试，然后提供服务器终端的完整日志！** 📝
