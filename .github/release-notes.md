Flash Plus **v1.1.1** 在 v1.1.0 基础上增加离线报告保存。

## 安装

从下方 Assets 下载对应平台 zip，在 Gauge 项目目录执行：

```bash
gauge install flash --file flash-1.1.1-linux.x86_64.zip
```

请按操作系统和架构替换文件名（`flash-1.1.1-<os>.<arch>.zip`）。然后在项目 `manifest.json` 的 `Plugins` 中加入 `flash`。

## 本版本

- 运行结束、暂停（SIGINT）或停止（Kill / SIGTERM）时，把报告保存到 `gauge_reports_dir`（默认 `reports/flash-plus/`）
- 尊重 `overwrite_reports`；可用 `FLASH_REPORTS_DIR` 覆盖根目录
- 保存可离线打开的 `index.html` 与 `snapshot.json`
- 保留 v1.1.0：Gauge 握手、树表报告、概念嵌套、动态参数替换
