package main

import (
	"fmt"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel"
)

const (
	tracerName = "http-handler"
)

func redirectHadler(w http.ResponseWriter, r *http.Request) {
	_, span := otel.Tracer(tracerName).Start(r.Context(), "redirect-handler")
	defer span.End()

	err := updateLinks()
	if err != nil {
		log.Errorf("fail to handle redirect: %v", err)
		return
	}

	urlPath := r.URL.Path

	// use first depth of path to sub-domain
	pathPrefix := getPathPrefix(urlPath)

	link, ok := redirects[pathPrefix]
	if !ok {
		log.Warnf("404: %s from %s", urlPath, r.RemoteAddr)
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "404")
		return
	}

	// reverse proxy for apps from same k8s cluster
	if link.RPLink != "" {
		// homin.dev/blog/page => blog.default.svc.cluster.local:8080 + /blog/page
		log.Debugf("RP: %s -> %s", urlPath, link.RPLink)
		serveReverseProxy(link.RPLink, w, r)
		return
	}

	// redirect for external sites
	log.Debugf("RD: %s -> link.Link", urlPath)
	http.Redirect(w, r, link.Link, http.StatusMovedPermanently)
}

func getPathPrefix(urlPath string) string {
	if len(urlPath) == 0 {
		return "/"
	}

	if urlPath[0] == '/' {
		urlPath = urlPath[1:]
	}

	i := strings.Index(urlPath, "/")
	if i < 0 {
		return "/" + urlPath
	}

	return "/" + urlPath[:i]
}
