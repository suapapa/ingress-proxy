package main

import (
	"embed"
	"log"
	"net/http"
)

type Asset struct {
	contentType string
	dataPath    string
}

var (
	//go:embed asset/favicon.ico
	//go:embed asset/ads.txt
	//go:embed asset/sitemap.xml
	efs    embed.FS
	assets = map[string]*Asset{
		"/favicon.ico": {contentType: "image/x-icon", dataPath: "asset/favicon.ico"},
		"/ads.txt":     {contentType: "asset/ads.txt", dataPath: "asset/ads.txt"},
		"/sitemap.xml": {contentType: "asset/sitemap.xml", dataPath: "asset/favicon.ico"},
	}
)

func isPathForAsset(path string) bool {
	_, ok := assets[path]
	return ok
}

func assetHandler(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path
	a, ok := assets[urlPath]
	if !ok {
		log.Printf("ERR: fail to handle: %s", urlPath)
		notFoundHandler(w, r)
		return
	}

	if b, err := efs.ReadFile(a.dataPath); err != nil {
		log.Printf("ERR: fail to read asset %s for %s", a.dataPath, urlPath)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", a.contentType)
		w.Write(b)
	}
}
