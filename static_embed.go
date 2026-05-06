package main

import (
	"embed"
	"io/fs"
)

//go:embed all:static
var staticEmbedFS embed.FS

// StaticFS 返回嵌入的静态文件系统
func StaticFS() fs.FS {
	fsys, _ := fs.Sub(staticEmbedFS, "static")
	return fsys
}
