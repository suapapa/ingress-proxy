package main

import (
	"fmt"
	"net/http"
	"strings"
)

func redirectHadler(w http.ResponseWriter, r *http.Request) {
	err := updateLinks()
	if err != nil {
		log.Errorf("fail to handle redirect: %v", err)
		return
	}

	urlPath := r.URL.Path

	// use first depth of path to sub-domain
	urlPrefix := getURLPrefix(urlPath)

	// redirect for external sites
	link, ok := redirects[urlPrefix]
	if !ok {
		log.Warnf("404: %s from %s", urlPath, r.RemoteAddr)
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "404")
		return
	}

	// reverse proxy for apps from same k8s cluster
	if link.RPLink != "" {
		// homin.dev/blog/page => blog.default.svc.cluster.local:8080 + /page
		log.Debugf("RP: %s -> %s", urlPath, link.RPLink)
		serveReverseProxy(link.RPLink, w, r)
		return
	}

	log.Debugf("RD: %s -> link.Link", urlPath)
	http.Redirect(w, r, link.Link, http.StatusMovedPermanently)
}

func getURLPrefix(urlPath string) string {
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
