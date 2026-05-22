package webui

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed dist
var embeddedDist embed.FS

func FileSystem() http.FileSystem {
	dist, err := fs.Sub(embeddedDist, "dist")
	if err != nil {
		return http.FS(embeddedDist)
	}

	return http.FS(dist)
}

func Ready() bool {
	dist, err := fs.Sub(embeddedDist, "dist")
	if err != nil {
		return false
	}

	index, err := dist.Open("index.html")
	if err != nil {
		return false
	}
	_ = index.Close()

	info, err := fs.Stat(dist, "index.html")
	return err == nil && info.Size() > 0
}
