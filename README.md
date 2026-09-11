# Flash Plus

[![Go](https://img.shields.io/badge/go-1.22-blue.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](LICENSE)

Gauge 执行进度实时报告插件，原版 [getgauge/Flash](https://github.com/getgauge/Flash) 的前后端分离实现。

- 通过 gRPC 接收 Gauge 执行事件
- HTTP + WebSocket 向前端推送进度
- Vue 3 实时报告页（规格书文件名 / 场景 / 步骤与概念、失败堆栈、Hide all）

## 安装

从 [GitHub Releases](https://github.com/Linhanmic/flash-plus/releases) 下载对应平台的 zip，在 Gauge 项目目录安装：

```bash
# Linux x86_64 示例，请按系统和架构替换文件名
gauge install flash --file flash-1.1.0-linux.x86_64.zip
```

各平台包命名为 `flash-1.1.0-<os>.<arch>.zip`，例如：

| 平台 | 文件 |
|------|------|
| Linux x86_64 | `flash-1.1.0-linux.x86_64.zip` |
| Linux arm64 | `flash-1.1.0-linux.arm64.zip` |
| macOS x86_64 | `flash-1.1.0-darwin.x86_64.zip` |
| macOS arm64 | `flash-1.1.0-darwin.arm64.zip` |
| Windows x86_64 | `flash-1.1.0-windows.x86_64.zip` |

或从源码生成当前平台分发包后安装：

```bash
go run build.go --distro
gauge install flash --file deploy/flash-1.1.0-linux.x86_64.zip
```

也可直接安装到 `~/.gauge/plugins`：

```bash
go run build.go --install
```

在项目 `manifest.json` 的 `Plugins` 中加入 `flash`：

```json
{
  "Language": "js",
  "Plugins": ["html-report", "flash"]
}
```

执行规格后，控制台会打印报告地址，例如：

```
Listening on port:12345
[Flash Plus] Starting progress reporting at http://127.0.0.1:56789
```

用浏览器打开该 HTTP 地址即可。HTTP 端口默认随机，可用环境变量固定：

```bash
FLASH_SERVER_PORT=8080
```

也可写在 `env/default/flash.properties` 中。

## 本地预览（无需 Gauge）

```bash
npm --prefix web install
go run build.go
./bin/linux_amd64/flash-server --demo
```

或：

```bash
FLASH_PLUS_DEMO=1 ./bin/linux_amd64/flash-server
```

## 从源码构建

依赖：Go 1.22+、Node.js 18+、npm。

```bash
go run build.go                 # 当前平台
go run build.go --all-platforms # 交叉编译
go run build.go --distro        # 生成 zip 分发包
go run build.go --install       # 安装到 ~/.gauge/plugins
```

报告树按 Gauge 执行结构展开：

| 层级 | 标签 | 显示名称 |
|------|------|----------|
| 规格书 | 规格书 | 规格文件名（如 `login.spec`），标题作为副标题 |
| 场景 | 场景 | 场景标题 |
| 步骤 / 概念 | 步骤、概念 | 步骤文本；概念可嵌套内部步骤；动态参数 `<name>` 替换为实际值 |

WebSocket 事件类型为 `suite` / `spec` / `scenario` / `concept` / `step` / `end`。步骤与概念名称会把 Gauge 动态参数 `<name>` 替换为运行时的实际值。gRPC 同时实现 `NotifyConceptExecutionStarting` / `NotifyConceptExecutionEnding`，兼容 Gauge 1.5.7+ 的 Reporter 调用。

本仓库作为 [uHIL](https://github.com/Linhanmic/uHIL) 子模块维护。Protobuf 消息来自 `github.com/getgauge/flash`。

## 配置

| 变量 | 说明 |
|------|------|
| `FLASH_SERVER_PORT` | HTTP 报告服务端口，未设置则使用随机端口 |
| `FLASH_PLUS_DEMO` | 设为 `1` 时播放内置演示时间线 |
| `GAUGE_PROJECT_ROOT` | Gauge 注入，用于报告标题中的项目名 |

## 许可

Apache License 2.0，见 [LICENSE](LICENSE)。
