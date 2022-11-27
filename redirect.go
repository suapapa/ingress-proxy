package main

import (
	"fmt"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func redirectHadler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := tracer.Start(ctx, "redirect-handler")
	defer span.End()
	// span := trace.SpanFromContext(ctx)
	// span.SetAttributes(
	// 	attribute.String("app.user.id", req.UserId),
	// 	attribute.String("app.user.currency", req.UserCurrency),
	// )

	err := updateLinks()
	if err != nil {
		log.Errorf("fail to handle redirect: %v", err)
		return
	}

	urlPath := r.URL.Path

	// use first depth of path to sub-domain
	span.AddEvent("find redirect link")
	pathPrefix, pathSurfix := splitPath(urlPath)
	link, ok := redirects[pathPrefix]
	if !ok {
		log.Warnf("404: %s from %s", urlPath, r.RemoteAddr)
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "404")
		return
	}

	// reverse proxy for apps from same k8s cluster
	if link.RPLink != "" {
		span.AddEvent("http-reverse-proxy",
			trace.WithAttributes(attribute.String("reverse-proxy-link", link.RPLink)),
		)
		// homin.dev/blog/page => blog.default.svc.cluster.local:8080 + /page
		log.Debugf("RP: %s -> %s", urlPath, link.RPLink)
		if link.RPOmitPrefix {
			serveReverseProxy(ctx, w, r, link.RPLink, pathSurfix)
		} else {
			serveReverseProxy(ctx, w, r, link.RPLink, urlPath)
		}
		return
	}

	span.AddEvent("http-redirect",
		trace.WithAttributes(attribute.String("redirect-link", link.Link)),
	)
	// redirect for external sites
	log.Debugf("RD: %s -> %s", urlPath, link.Link)
	http.Redirect(w, r, link.Link, http.StatusMovedPermanently)
}

func splitPath(urlPath string) (prefix, surfix string) {
	if len(urlPath) == 0 {
		return "/", ""
	}

	if urlPath[0] == '/' {
		urlPath = urlPath[1:]
	}

	i := strings.Index(urlPath, "/")
	if i < 0 {
		return "/" + urlPath, ""
	}

	return "/" + urlPath[:i], urlPath[i:]
}
