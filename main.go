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
)

var (
	linksConf  string
	acPath     string
	debug      bool
	programVer = "dev"
)

func main() {
	notifyToTelegram("started")
	defer func() {
		notifyToTelegram("terminated")
	}()

	flag.StringVar(&linksConf, "c", "conf/links.yaml", "yaml file which has links")
	flag.StringVar(&acPath, "ac", "/tmp/letsencrypt/", "acme-challenge file path")
	flag.BoolVar(&debug, "d", false, "enable debug")
	flag.Parse()

	if debug {
		log.Logger.SetLevel(logrus.DebugLevel)
	}

	acmeChallenge := NewAcmeChallenge("/tmp/letsencrypt/")

	// HTTP server
	httpSrvMux := http.NewServeMux()
	httpSrvMux.Handle("/.well-known/acme-challenge/", acmeChallenge)
	httpSrvMux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello stranger?")
	})
	go func() {
		log.Infof("listening http on :%d", httpPort)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", httpPort), httpSrvMux); err != nil {
			log.Fatal(err)
		}
	}()

	// HTTPS server
	http.Handle("/.well-known/acme-challenge/", acmeChallenge)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		urlPath := r.URL.Path
		// log.Printf("urlPath=%s", urlPath)
		if isPathForAsset(urlPath) {
			assetHandler(w, r)
		} else /* if isPathForRedirect(urlPath) */ {
			redirectHadler(w, r)
		}
	})
	go startHTTPSServer(true)

	// PortFowarding
	go startPortFoward()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}
