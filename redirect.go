package main

import (
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
	subDomain, _ := getSubdomain(urlPath)

	// redirect for external sites
	link, ok := redirects[subDomain]
	if !ok {
		// 모르는 건 죄다 ingress 에 투척
		link = redirects["/"]
		r.URL.Path = "404"
		// link.Link = subDomain[1:] + ".default.svc.cluster.local:8080"
		// notFoundHandler(w, r)
		// return
	}

	// reverse proxy for apps from same k8s cluster
	if link.RPLink != "" {
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
