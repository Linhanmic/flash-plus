/*----------------------------------------------------------------
 *  Flash - 前后端分离版本
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package main

import (
	"embed"
	"io/fs"

	"github.com/getgauge/flash-server/event"
	flashGrpc "github.com/getgauge/flash-server/grpc"
	flashHttp "github.com/getgauge/flash-server/http"
)

//go:embed dist/*
var distFS embed.FS

func main() {
	e := make(chan event.Event)

	// 启动 gRPC 服务（接收 Gauge 事件）
	go flashGrpc.Start(e)

	// 获取 dist 子目录
	staticFS, _ := fs.Sub(distFS, "dist")

	// 启动 HTTP 服务（API + WebSocket + 前端静态文件）
	flashHttp.Start(e, staticFS)
}
