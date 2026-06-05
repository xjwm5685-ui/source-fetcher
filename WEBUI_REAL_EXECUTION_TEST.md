# Web UI 真实执行测试指南

## 🎯 验证下载和安装功能确实执行

### 前置条件

1. 确保已编译最新代码：
```bash
go build -o source-fetcher.exe
```

2. 启动 Web GUI：
```bash
.\source-fetcher.exe gui
```

3. 打开浏览器开发者工具（F12），切换到 Console 标签

## ✅ 测试步骤

### 测试 1: 下载功能

1. **搜索包**
   - 在搜索框输入 `lodash`
   - 选择源：`npm`
   - 等待搜索结果

2. **选择并下载**
   - 勾选第一个结果（lodash）
   - 点击"下载"按钮（绿色下载图标）
   - 查看控制台输出：
     ```
     Downloading packages with indexes: [0]
     Download response: {success: true, added: 1}
     Queue updated: 1 tasks
     ```

3. **验证队列**
   - 自动跳转到"队列"标签页
   - 应该看到任务：`Download lodash@x.x.x`
   - 任务状态变化：`等待中` → `运行中` → `已完成` 或 `失败`
   - 控制台应显示：
     ```
     Running tasks: ["[1] Download lodash@x.x.x"]
     ```

4. **验证文件**
   - 检查当前目录是否有下载的文件
   - 对于 npm 包，应该有 `.tgz` 文件
   ```bash
   dir *.tgz
   ```

### 测试 2: 安装功能（仅 npm）

1. **搜索 npm 包**
   - 搜索框输入 `react`
   - 源选择 `npm`
   - 等待结果

2. **选择并安装**
   - 勾选第一个结果
   - 点击"安装"按钮（蓝色安装图标）
   - 控制台输出：
     ```
     Installing packages with indexes: [0]
     Install response: {success: true, added: 1}
     ```

3. **验证安装**
   - 切换到"队列"标签页
   - 看到任务：`Install react@x.x.x`
   - 任务状态更新
   - 控制台显示：
     ```
     Running tasks: ["[2] Install react@x.x.x"]
     ```

4. **验证安装结果**
   - 检查是否生成了 `node_modules` 文件夹
   - 检查是否有 `.source-fetcher` 缓存文件夹
   ```bash
   dir node_modules
   dir .source-fetcher
   ```

### 测试 3: 批量操作

1. **搜索并多选**
   - 搜索 `lodash`
   - 按 `Ctrl+A` 全选结果（或手动多选 3-5 个）

2. **批量下载**
   - 点击下载按钮
   - 控制台应显示：
     ```
     Downloading packages with indexes: [0, 1, 2, 3, 4]
     Download response: {success: true, added: 5}
     Queue updated: 5 tasks
     ```

3. **观察队列执行**
   - 队列中应该有 5 个任务
   - 任务按顺序执行（串行）
   - 第一个任务状态：`运行中`
   - 其他任务状态：`等待中`
   - 控制台持续显示执行进度

### 测试 4: 错误处理

1. **测试不支持的安装**
   - 搜索 `winget` 包（非 npm）
   - 选择一个结果
   - 点击"安装"
   - 应该看到任务失败
   - 错误信息：`install is only supported for npm packages, got: winget`

2. **测试网络错误**
   - 断开网络
   - 尝试搜索
   - 应该看到错误 Toast 提示

## 🔍 调试检查点

### 浏览器控制台

查找这些日志输出：

```javascript
// 下载请求
Downloading packages with indexes: Array(1)
Download response: {success: true, added: 1}

// 安装请求
Installing packages with indexes: Array(1)
Install response: {success: true, added: 1}

// 队列更新
Queue updated: 1 tasks

// 运行任务
Running tasks: Array(1)
```

### 服务器日志

在运行 `.\source-fetcher.exe gui` 的终端中查看：

```
Task 1 completed: Download lodash@4.17.21
Task 2 completed: Install react@19.0.0
Task 3 failed: install is only supported for npm packages
```

### 文件系统

检查这些文件/文件夹：

```
当前目录/
├── lodash-4.17.21.tgz          # 下载的包
├── node_modules/               # 安装目录
│   └── react/
└── .source-fetcher/            # 缓存目录
    ├── store/                  # 解包缓存
    └── tarballs/               # tarball 缓存
```

## ❌ 常见问题

### 问题 1: 队列任务一直"等待中"

**可能原因**：
- 队列处理器未启动
- 后端任务执行卡住

**解决方法**：
```bash
# 重启服务器
Ctrl+C
.\source-fetcher.exe gui
```

### 问题 2: 任务显示"失败"但无错误信息

**可能原因**：
- 网络连接问题
- 权限不足
- 磁盘空间不足

**解决方法**：
- 查看服务器终端的详细错误
- 检查网络连接
- 以管理员权限运行

### 问题 3: 控制台无日志输出

**可能原因**：
- 开发者工具未打开
- Console 被过滤

**解决方法**：
- 按 F12 打开开发者工具
- 切换到 Console 标签
- 确保 Console 过滤器设置为 "All levels"

### 问题 4: 文件下载到哪里了？

**当前行为**：
- 下载到执行 `source-fetcher.exe` 的当前目录
- 安装到当前目录的 `node_modules`

**修改下载路径**：
需要在后端代码中修改 `OutputDir: "."` 为期望路径。

## 🎨 自定义下拉框测试

### 验证自定义 UI

1. **点击下拉框**
   - 点击"Source"下拉框
   - 应该展开自定义选项列表
   - 不是浏览器原生的 `<select>`

2. **选择选项**
   - 点击任意选项
   - 下拉框关闭
   - 显示值更新

3. **键盘导航**
   - 点击下拉框打开
   - 按 `↓` 向下选择
   - 按 `↑` 向上选择
   - 按 `Enter` 确认
   - 按 `Esc` 取消

4. **点击外部关闭**
   - 打开下拉框
   - 点击页面其他地方
   - 下拉框自动关闭

## 📊 成功标准

✅ **下载功能正常**：
- 能添加到队列
- 任务能执行
- 文件能下载

✅ **安装功能正常**：
- npm 包能安装
- 生成 `node_modules`
- 非 npm 包正确提示错误

✅ **队列功能正常**：
- 任务状态正确更新
- 串行执行
- 错误能正确显示

✅ **UI 功能正常**：
- 自定义下拉框工作
- 所有交互响应
- 控制台有日志

## 🚀 下一步

如果所有测试通过：
- ✅ Web UI 功能完整
- ✅ 真实执行下载/安装
- ✅ 自定义 UI 组件工作

如果有问题：
1. 查看本文档的"常见问题"部分
2. 检查控制台和服务器日志
3. 确保 Go 代码已重新编译

---

**测试完成后，请报告结果！** 📝
