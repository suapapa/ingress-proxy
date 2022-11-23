package main

import (
	"net/http"
)

// Serve a reverse proxy for a given url
func serveReverseProxy(target string, res http.ResponseWriter, req *http.Request) {
	/*
		// parse the url
		url, err := url.Parse(target)
		if err != nil {
			log.Errorf("fail serve reverse proxy: %v", err)
		}

		// create the reverse proxy
		proxy := httputil.NewSingleHostReverseProxy(url)
	*/
	rpc, err := getReverseProxy(target)
	if err != nil {
		log.Errorf("fail serve reverse proxy: %v", err)
	}
	url, proxy := rpc.URL, rpc.ReverseProxy

	// Update the headers to allow for SSL redirection
	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = url.Host

	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(res, req)
}
