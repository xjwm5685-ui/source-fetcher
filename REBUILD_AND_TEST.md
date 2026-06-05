# 🔧 重新编译和测试指南

## ⚡ 快速操作

### 1. 停止当前服务
```bash
# 在运行 GUI 的终端按 Ctrl+C
```

### 2. 重新编译（已添加调试日志）
```bash
cd d:\dy\source-fetcher
go build -o source-fetcher.exe
```

### 3. 启动服务
```bash
.\source-fetcher.exe gui
```

**重要**：保持这个终端窗口可见，以便看到调试日志！

### 4. 浏览器测试
```
1. 打开 http://localhost:8765
2. 按 F12 打开开发者工具
3. 搜索: openclaw
4. 源: npm
5. 等待结果显示
6. 勾选第一个结果
7. 点击安装按钮（蓝色图标）
8. 观察两个地方：
   - 浏览器控制台的日志
   - 服务器终端的日志
```

## 📋 观察重点

### A. 服务器终端应该显示：

```
Install request received: indexes=[0]
Current search results count: X
Adding install task for index 0: Source=npm, Identifier=openclaw, Version=XXXXX
Added 1 install tasks to queue
Executing install: Source=npm, Identifier=openclaw, Version=XXXXX
InstallRequest: Name=openclaw, Version=XXXXX, OutputDir=.
```

### B. 浏览器控制台应该显示：

```javascript
Installing packages with indexes: [0]
Install response: {success: true, added: 1}
Queue updated: 1 tasks
```

## 🔍 关键问题

请注意服务器日志中的 **Version** 值：

### 如果显示：
```
Version=2026.6.1  ✅ 正常
Version=          ❌ 版本为空！
Version=latest    ⚠️  使用了 latest 标签
```

## 🎯 如果版本为空

这说明搜索结果中的 `Version` 字段没有正确设置。

需要检查：
1. 搜索 API 返回的数据
2. `SearchResult` 结构体的映射
3. npm registry 的响应格式

## 📊 收集信息

测试后请提供：

### 1. 服务器终端的完整输出
从启动到安装失败的所有日志

### 2. 搜索结果的详细信息
浏览器控制台执行：
```javascript
console.log('Search results:', searchResults);
```

### 3. 验证包是否存在
在命令行执行：
```bash
npm view openclaw
```

### 4. 检查版本字段
在浏览器控制台执行：
```javascript
// 搜索完成后
searchResults.forEach((r, i) => {
    console.log(`[${i}] ${r.Identifier} - Version: "${r.Version}" - Source: ${r.Source}`);
});
```

## 🔄 替代测试

如果 `openclaw` 持续有问题，用这些包测试：

```bash
# 测试 1: lodash（肯定存在）
搜索: lodash
源: npm
结果: 应该有版本号如 4.17.21
操作: 安装第一个

# 测试 2: react（流行包）
搜索: react  
源: npm
结果: 应该有版本号如 18.2.0 或 19.0.0
操作: 安装第一个

# 测试 3: express（常用包）
搜索: express
源: npm
结果: 应该有版本号如 4.18.x
操作: 安装第一个
```

## 🐛 常见问题

### Q1: 编译失败
```bash
# 确保 Go 环境正确
go version

# 清理缓存重新编译
go clean
go build -o source-fetcher.exe
```

### Q2: 端口被占用
```bash
# 如果 8765 被占用，手动指定端口
# （需要修改代码，暂不支持命令行参数）
```

### Q3: 浏览器缓存
```bash
# 强制刷新页面
Ctrl + Shift + R

# 或清除缓存
F12 -> Network -> Disable cache
```

## ✅ 成功标志

安装成功会看到：

### 服务器终端：
```
Task X completed: Install openclaw@2026.6.1
```

### 文件系统：
```bash
dir node_modules\openclaw
dir .source-fetcher\tarballs
```

### 队列状态：
```
状态: COMPLETED ✅
```

---

**现在就开始测试，记录所有日志！** 🚀
