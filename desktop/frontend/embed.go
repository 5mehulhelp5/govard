package frontend

import "embed"

// Assets contains the embedded frontend files served by the Wails asset server.
// Each pattern below is listed explicitly so that node_modules and build
// tooling (tailwind.config.js, assets/styles-src.css, …) are excluded from the binary.
//
//go:embed index.html main.js assets
//go:embed modules services state ui utils wailsjs
var Assets embed.FS
