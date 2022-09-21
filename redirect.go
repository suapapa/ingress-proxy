package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

func isPathForRedirect(path string) bool {
	err := updateLinks()
	if err != nil {
		log.Printf("ERR: %v", errors.Wrap(err, "fail to check path for redirect"))
		return false
	}

	_, ok := redirects[path]
	return ok
}

func redirectHadler(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path
	err := updateLinks()
	if err != nil {
		log.Printf("ERR: %v", err)
		return
	}

	// use first depth of path to sub-domain
	subDomain, _ := getSubdomain(urlPath)

	// redirect for external sites
	link, ok := redirects[subDomain]
	if !ok {
		return
	}

	// reverse proxy for apps from same k8s cluster
	if link.RP {
		// homin.dev/blog/page => blog.default.svc.cluster.local:8080 + /page
		serveReverseProxy(link.RPLink, w, r)
		return
	}

	http.Redirect(w, r, link.Link, http.StatusMovedPermanently)
}

func getSubdomain(urlPath string) (string, string) {
	if len(urlPath) == 0 {
		return "", ""
	}

	if urlPath[0] == '/' {
		urlPath = urlPath[1:]
	}

	i := strings.Index(urlPath, "/")
	if i < 0 {
		return urlPath, "/"
	}

	return urlPath[:i], urlPath[i:]
}
