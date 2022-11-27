package main

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
)

// Serve a reverse proxy for a given url
func serveReverseProxy(ctx context.Context, res http.ResponseWriter, req *http.Request, from, to string) {
	_, span := tracer.Start(ctx, "serve-reverse-proxy")
	defer span.End()

	rpc, err := getReverseProxy(from)
	if err != nil {
		log.Errorf("fail serve reverse proxy: %v", err)
	}
	span.SetAttributes(
		attribute.String("from", from),
		attribute.String("to", to),
	)

	url, proxy := rpc.URL, rpc.ReverseProxy

	// Update the headers to allow for SSL redirection
	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.URL.Path = to
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = url.Host

	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(res, req)
}
