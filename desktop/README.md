# dsh-desktop

> DeepSeek Harness 的**独立桌面客户端**（Windows），把 Web GUI 包装成原生窗口 —— 本项目自有的桌面应用。

基于 [Wails v2](https://wails.io) + WebView2，把 DeepSeek Harness 的 Web 界面（默认 `http://127.0.0.1:3080`）包装成原生 Windows 窗口，使用体验类似 opencode 桌面版。

## 特性

- **自动编排**：检测 3080 端口是否已有 dsh server —— 没有则后台拉起 `dsh web` 并等待就绪，已有则直接复用
- **进程生命周期**：应用退出时只杀掉自己拉起的 dsh 进程树；复用外部已存在的 server 时不动它
- **窗口状态记忆**：记住上次窗口大小 / 位置 / 最大化状态，重启自动还原
- **单实例**：重复启动时聚焦已有窗口，不重复拉起
- **启动画面**：内嵌 loading 页，失败时显示原因并可一键重试
- **运行配置**：`config.json` 可配端口 / 启动命令 / 是否自动更新
- **崩溃自愈**：运行中 harness 掉线自动重启并重连；复用外部实例时仅提示
- **错误诊断**：启动失败 / 崩溃时把 `dsh.log` 尾部一并展示
- **系统通知**：崩溃自愈 / 断连时发原生通知
- **自动更新 harness**：拉起新实例前自动把 dsh 更新到最新版（可通过配置关闭）
- **外链处理**：harness UI 里的外部链接用系统浏览器打开，不在桌面壳内跳走
- **原生菜单**：菜单栏「重新加载 / 打开日志 / 打开配置 / 退出」+ 快捷键

## 工作原理

1. 启动时显示内嵌启动画面（spinner）
2. 探测 `127.0.0.1:3080`：
   - **未占用** → 后台拉起 `dsh web`（日志写入 `%APPDATA%\dsh-desktop\dsh.log`），轮询 HTTP 200 就绪后窗口跳转到 `http://127.0.0.1:3080`
   - **已占用** → 直接复用
3. 30 秒未就绪 → 启动画面显示错误提示，可重试

## 环境要求

- **运行**：Windows 10+（自带 WebView2 Runtime）、Node.js，以及可用的 `dsh web`：
  - 本仓库源码：在仓库根目录 `pnpm install && pnpm run build` 后，`pnpm dsh web` 即服务 Web UI（桌面壳检测到 3080 已占用会直接复用它）
  - 或全局安装：`npm i -g @deepseek-ai/dsh`
- **构建**：Go 1.26+、Wails CLI v2

## 构建

```powershell
cd desktop
wails build            # 产物: build\bin\dsh-desktop.exe
```

生成 NSIS 安装器：

```powershell
wails build -nsis      # 产物: build\bin\dsh-desktop-amd64-installer.exe
```

## 开发

```powershell
cd desktop
wails dev              # 热重载开发模式
```

## 目录结构

```
desktop/
  app.go             # 启动编排：端口检测 / 拉起 dsh / 等待就绪 / 窗口跳转 / 退出清理
  dsh_windows.go     # Windows 平台 spawn dsh（隐藏窗口 + 日志重定向）
  dsh_other.go       # 其他平台 spawn dsh
  windowstate.go     # 窗口状态持久化（大小/位置/最大化）
  config.go          # 运行配置（端口 / 启动命令 / 自动更新）
  update.go          # harness 版本检测与自动更新
  logtail.go         # dsh.log 尾部读取（错误诊断用）
  links.go           # 外链桥接（外部链接跳系统浏览器）
  menu.go            # 原生菜单与快捷键
  main.go            # Wails 入口（单实例锁、窗口参数）
  wails.json         # Wails 项目配置（含版本元数据）
  build/             # 图标 / manifest / NSIS 安装器 / macOS plist
  frontend/          # 启动画面页（Vite + 原生 JS）
```

## 配置

- 运行配置保存在 `%APPDATA%\dsh-desktop\config.json`：`port`（默认 3080）、`command`（默认 `dsh web`，可为 `pnpm dsh web` 等）、`autoUpdateHarness`（默认 `true`，拉起新实例前自动把 dsh 更新到最新版）、`workdir`（harness 进程工作目录，用于定位 `.env`/cordis 配置，默认空=继承桌面壳目录），缺失时用默认值
- 窗口状态保存在 `%APPDATA%\dsh-desktop\window.json`，关闭前写入（`OnBeforeClose`）、启动时还原（`OnStartup`）
- dsh 运行日志保存在 `%APPDATA%\dsh-desktop\dsh.log`

## 已知问题

- **exe 的「文件属性 → 详细信息」版本字段显示为空**：wails v2.14.0 所用 `tc-hib/winres v0.3.1` 的上游 bug（版本资源 StringTable 键名写成小写 `040904b0`，Windows 期望大写 `040904B0`），导致 `FileVersionInfo` 读不到字符串。字符串实际已嵌入 exe、二进制文件版本 `0.1.0.0` 正常，属纯外观问题、不影响运行，等 wails 升级 winres 后自动修复。

## 作者与许可

- 作者：FFaassdfs
- 许可证：MIT（随本仓库）
