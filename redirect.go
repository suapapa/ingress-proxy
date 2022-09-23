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
		// 모르는 건 죄다 ingress, 404 에 투척
		log.Debugf("404: %s", urlPath)
		link = redirects["/ingress"]
		r.URL.Path = "/404"
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
