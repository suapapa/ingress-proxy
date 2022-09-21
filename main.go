package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const (
	httpPort = 80
)

var (
	linksConf string
	acPath    string
)

func main() {
	flag.StringVar(&linksConf, "c", "conf/links.yaml", "yaml file which has links")
	flag.StringVar(&acPath, "ac", "/tmp/letsencrypt/", "acme-challenge file path")
	flag.Parse()

	http.Handle("/.well-known/acme-challenge/", NewAcmeChallenge("/tmp/letsencrypt/"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		urlPath := r.URL.Path
		if isPathForAsset(urlPath) {
			assetHandler(w, r)
		} else if isPathForRedirect(urlPath) {
			redirectHadler(w, r)
		} else {
			notFoundHandler(w, r)
		}
	})

	go func() {
		log.Printf("listening http on :%d", httpPort)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", httpPort), nil); err != nil {
			log.Fatal(err)
		}
	}()

	go startHTTPSServer()
	go startPortFoward()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}
