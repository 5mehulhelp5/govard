package dockerassets

import (
	"embed"
	"io/fs"
)

//go:embed all:*
var files embed.FS

// FS exposes embedded Docker build contexts for local fallback image builds.
var FS fs.FS

func init() {
	var err error
	FS, err = fs.Sub(files, ".")
	if err != nil {
		panic(err)
	}
}
