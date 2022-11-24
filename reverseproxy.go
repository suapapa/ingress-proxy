package main

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
)

// Serve a reverse proxy for a given url
func serveReverseProxy(trCtx context.Context, res http.ResponseWriter, req *http.Request, target, path string) {
	_, span := otel.Tracer(tracerName).Start(trCtx, "serve-reverse-proxy")
	defer span.End()

	rpc, err := getReverseProxy(target)
	if err != nil {
		log.Errorf("fail serve reverse proxy: %v", err)
	}
	url, proxy := rpc.URL, rpc.ReverseProxy

	// Update the headers to allow for SSL redirection
	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.URL.Path = path
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = url.Host

	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(res, req)
}
