package main

import (
	"net/http"
	"strings"
)

func isPathForRedirect(path string) bool {
	err := updateLinks()
	if err != nil {
		log.Errorf("fail to chek for redirect: %v", err)
		return false
	}

	// use first depth of path to sub-domain
	subDomain, _ := getSubdomain(path)
	log.Printf("subDomain1=%s", subDomain)

	_, ok := redirects[subDomain]
	return ok
}

func redirectHadler(w http.ResponseWriter, r *http.Request) {
	err := updateLinks()
	if err != nil {
		log.Errorf("fail to handle redirect: %v", err)
		return
	}

	urlPath := r.URL.Path

	// use first depth of path to sub-domain
	subDomain, _ := getSubdomain(urlPath)
	log.Printf("subDomain2=%s", subDomain)

	// redirect for external sites
	link, ok := redirects[subDomain]
	if !ok {
		link = redirects["/"]
		link.Link = subDomain[1:] + ".default.svc.cluster.local:8080"
	}

	log.Println("link=", link, ok)

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
		return "/", ""
	}

	if urlPath[0] == '/' {
		urlPath = urlPath[1:]
	}

	i := strings.Index(urlPath, "/")
	if i < 0 {
		return "/" + urlPath, "/"
	}

	return "/" + urlPath[:i], urlPath[i:]
}
