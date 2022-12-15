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
	// "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
)

const (
	httpPort    = 80
	programName = "ingress-proxy"

	// otplEP = "simplest-collector.default.svc.cluster.local:4317"
)

var (
	linksConf string
	acPath    string
	debug     bool
	// enableTrace bool
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

	// if enableTrace {
	// 	ctx := context.Background()
	// 	tp := initTracerProvider(ctx, otplEP)
	// 	defer func() {
	// 		if err := tp.Shutdown(ctx); err != nil {
	// 			log.Errorf("Error shutting down tracer provider: %v", err)
	// 		}
	// 	}()
	// 	tracer = tp.Tracer(programName)
	// }

	// mp := initMeterProvider(ctx, otplEP)
	// defer func() {
	// 	if err := mp.Shutdown(ctx); err != nil {
	// 		log.Errorf("Error shutting down meter provider: %v", err)
	// 	}
	// }()

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

	// go func() {
	// 	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", "5050"))
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	var srv = grpc.NewServer(
	// 		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
	// 		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	// 	)
	// 	if err := srv.Serve(lis); err != nil {
	// 		log.Fatal(err)
	// 	}
	// }()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}
