package main

import (
	"embed"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Asset struct {
	contentType string
	dataPath    string
}

var (
	//go:embed asset/favicon.ico
	//go:embed asset/ads.txt
	//go:embed asset/robots.txt
	//go:embed asset/sitemap.xml
	efs    embed.FS
	assets = map[string]*Asset{
		"/favicon.ico": {contentType: "image/x-icon", dataPath: "asset/favicon.ico"},
		"/ads.txt":     {contentType: "text/plain", dataPath: "asset/ads.txt"},
		"/robots.txt":  {contentType: "text/plain", dataPath: "asset/robots.txt"},
		"/sitemap.xml": {contentType: "text/xml", dataPath: "asset/sitemap.xml"},
	}
)

func isPathForAsset(path string) bool {
	_, ok := assets[path]
	return ok
}

func assetHandler(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path
	_, span := otel.Tracer("").Start(r.Context(), "asset-handler", trace.WithAttributes(
		attribute.String("path", urlPath),
	))
	defer span.End()

	a, ok := assets[urlPath]
	if !ok {
		return
	}

	if b, err := efs.ReadFile(a.dataPath); err != nil {
		log.Errorf("fail to read asset %s for %s", a.dataPath, urlPath)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", a.contentType)
		w.Write(b)
	}
}
