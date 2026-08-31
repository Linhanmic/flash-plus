# Flash Plus

[![Go](https://img.shields.io/badge/go-1.22-blue.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](LICENSE)

Gauge 执行进度实时报告插件，原版 [getgauge/Flash](https://github.com/getgauge/Flash) 的前后端分离实现。

- 通过 gRPC 接收 Gauge 执行事件
- HTTP + WebSocket 向前端推送进度
- Vue 3 实时报告页（规格书文件名 / 场景 / 步骤与概念、失败堆栈、Hide all）
- 运行结束、暂停（SIGINT）或停止（Kill / SIGTERM）时，把可离线打开的报告保存到 env 配置的目录

## 安装

在 Gauge 项目目录：

```bash
go run build.go --distro
gauge install flash --file deploy/flash-1.1.1-linux.x86_64.zip
```

或本地直接安装到 `~/.gauge/plugins`：

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

保存的静态报告目录由 Gauge env 决定，例如 `env/default/default.properties`：

```
gauge_reports_dir = reports
overwrite_reports = true
```

运行结束、Ctrl+C 暂停，或 Gauge Kill / SIGTERM 停止时，会写入：

```
reports/flash-plus/index.html
reports/flash-plus/snapshot.json
```

`overwrite_reports = false` 时则为 `reports/flash-plus/2006-01-02_15.04.05/`。直接用浏览器打开 `index.html` 即可查看当时的树状结果，无需再连实时服务。

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
| `GAUGE_PROJECT_ROOT` | Gauge 注入，用于报告标题中的项目名，以及解析相对报告目录 |
| `gauge_reports_dir` | 报告根目录，默认 `reports`（相对路径基于项目根）。写在 `env/default/default.properties` |
| `overwrite_reports` | `true` 时覆盖 `reports/flash-plus/`；`false` 时写入带时间戳的子目录 |
| `FLASH_REPORTS_DIR` | 可选，覆盖 `gauge_reports_dir` |

## 许可

Apache License 2.0，见 [LICENSE](LICENSE)。
