# dsh-desktop

DeepSeek Harness 的桌面应用壳（Wails v2 + WebView2），把 dsh 的 Web UI 包装成原生桌面窗口，使用方式类似 opencode 桌面版。

本目录是本仓库的**独立桌面部分**：一个独立的 Go + Vite 子项目（不参与 pnpm workspace 构建），把本仓库 fork 出的 Web GUI（默认 `http://127.0.0.1:3080`）包装成 Windows 原生窗口。

## 原理

- 启动时显示内嵌启动画面（loading spinner）
- 检测 `127.0.0.1:3080` 是否已有 dsh server：
  - **没有** → 自动后台拉起 `dsh web`（日志写入 `%APPDATA%\dsh-desktop\dsh.log`），轮询 HTTP 200 就绪后窗口跳转到 `http://127.0.0.1:3080`；**应用退出时杀掉自己拉起的 dsh 进程树**
  - **已有** → 直接复用，退出时不动它
- 30 秒未就绪 → 启动画面显示错误提示

## 依赖

- **运行**：Windows 10+（自带 WebView2 Runtime）、Node.js，以及可用的 `dsh web`：
  - 本仓库源码：在仓库根目录 `pnpm install && pnpm run build` 后，`pnpm dsh web` 即服务本 fork 的 Web UI（3080 端口已被占用时，桌面壳会直接复用它）
  - 或全局安装：`npm i -g @deepseek-ai/dsh`
- **构建**：Go 1.26+、Wails CLI v2

## 构建

```powershell
wails build          # 产物: build\bin\dsh-desktop.exe
```

## 开发

```powershell
wails dev            # 热重载开发模式
```

## 结构

```
app.go             # 启动编排：端口检测 / 拉起 dsh / 等待就绪 / 窗口跳转 / 退出清理
dsh_windows.go     # Windows 平台 spawn dsh（隐藏窗口 + 日志重定向）
dsh_other.go       # 其他平台 spawn dsh
windowstate.go     # 窗口状态持久化（记住大小/位置/最大化，重启还原）
main.go            # Wails 入口（单实例锁、窗口参数）
frontend/          # 启动画面页（Vite + 原生 JS）
```

窗口大小/位置/最大化状态保存在 `%APPDATA%\dsh-desktop\window.json`，关闭前写入（`OnBeforeClose`）、启动时还原（`OnStartup`）。

## 已知问题

- **exe 的「文件属性 → 详细信息」版本字段显示为空**：这是 wails v2.14.0 所用 `tc-hib/winres v0.3.1` 的上游 bug —— 它把版本资源的 StringTable 键名写成小写（`040904b0`），而 Windows 期望大写（`040904B0`），导致 `FileVersionInfo` 读不到字符串。字符串实际已嵌入 exe，二进制文件版本 `0.1.0.0` 也正常，属纯外观问题、不影响运行，等 wails 升级 winres 后自动修复。
