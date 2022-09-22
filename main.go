package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

const (
	httpPort    = 80
	programName = "ingress-proxy"
	programVer  = "dev"
)

var (
	linksConf string
	acPath    string
	debug     bool
)

func main() {
	flag.StringVar(&linksConf, "c", "conf/links.yaml", "yaml file which has links")
	flag.StringVar(&acPath, "ac", "/tmp/letsencrypt/", "acme-challenge file path")
	flag.BoolVar(&debug, "ad", false, "enagle debug")
	flag.Parse()

	if debug {
		log.Logger.SetLevel(logrus.DebugLevel)
	}

	http.Handle("/.well-known/acme-challenge/", NewAcmeChallenge("/tmp/letsencrypt/"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		urlPath := r.URL.Path
		log.Printf("urlPath=%s", urlPath)
		if isPathForAsset(urlPath) {
			log.Printf("isAsset")
			assetHandler(w, r)
		} else /* if isPathForRedirect(urlPath) */ {
			log.Printf("isRedirect")
			redirectHadler(w, r)
			// } else {
			// 	notFoundHandler(w, r)
		}
	})

	go func() {
		log.Infof("listening http on :%d", httpPort)
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
