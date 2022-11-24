package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
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

	tp, err := tracerProvider("http://simplest.default.svc.cluster.local:14268")
	if err != nil {
		log.Fatal(err)
	}

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)

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
