package web

import "embed"

// Dist 是 Vite 的生产构建。构建 Go 之前要先在 web/ 执行 npm run build。
//
//go:embed all:dist
var Dist embed.FS
