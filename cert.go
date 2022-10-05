package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/pkg/errors"
)

const (
	SSL_CERT_FILE = "/etc/letsencrypt/live/homin.dev/fullchain.pem"
	SSL_KEY_FILE  = "/etc/letsencrypt/live/homin.dev/privkey.pem"

	httpsPort = 443
)

var (
	md5SSLCert, md5SSLKey []byte

	startingHTTPS   bool
	startHTTPSMutex sync.Mutex
)

type AcmeChallenge struct {
	fileHandler http.Handler
}

func NewAcmeChallenge(acPath string) *AcmeChallenge {
	os.MkdirAll(acPath, 0700)
	return &AcmeChallenge{
		fileHandler: http.FileServer(http.FileSystem(http.Dir(acPath))),
	}
}

func (ac *AcmeChallenge) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ac.fileHandler.ServeHTTP(w, r)

	log.Debugf("(re)start https server")
	go startHTTPSServer(false)
}

func checkSSLCertUpdated() error {
	if !filesExist(SSL_CERT_FILE) || !filesExist(SSL_KEY_FILE) {
		out, err := exec.Command("/bin/create_ssl_cert.sh").Output()
		log.Info(string(out))
		if err != nil {
			return errors.Wrap(err, "check ssl fail")
		}
		log.Warn("SSL just created")
		time.Sleep(1 * time.Second)
	}

	currMD5SSLCert, err := md5sumFile(SSL_CERT_FILE)
	if err != nil {
		return errors.Wrap(err, "fail to check ssl cert update")
	}
	currMD5SSLKey, err := md5sumFile(SSL_KEY_FILE)
	if err != nil {
		return errors.Wrap(err, "fail to check ssl cert update")
	}

	// same cert as before
	if bytes.Equal(currMD5SSLCert, md5SSLCert) && bytes.Equal(currMD5SSLKey, md5SSLKey) {
		return fmt.Errorf("same as old ssl cert")
	}

	md5SSLCert = currMD5SSLCert
	md5SSLKey = currMD5SSLKey

	return nil
}

func startHTTPSServer(runCertbot bool) {
	startHTTPSMutex.Lock()
	if startingHTTPS {
		log.Debugf("already trying to start https server...")
		startHTTPSMutex.Unlock()
		return
	}
	defer startHTTPSMutex.Unlock()
	startingHTTPS = true

	// we will try 5 times
	ctx, cancelF := context.WithTimeout(context.Background(), 5*time.Minute+30*time.Second)
	defer cancelF()
	tick := time.NewTicker(1 * time.Minute)
	defer tick.Stop()

	time.Sleep(3 * time.Second) //wait for http server ready
	if err := startHTTPSServerInternal(runCertbot); err != nil {
		log.Errorf("failt to start https server: %v", err)
	} else {
		return
	}

	for {
		select {
		case <-ctx.Done():
			log.Error("fail to launch https server. trying to restart pod...")
			os.Exit(-1)
			return
		case <-tick.C:
			if err := startHTTPSServerInternal(runCertbot); err != nil {
				log.Errorf("failt to start https server: %v", err)
			} else {
				return
			}
		}
	}
}

func startHTTPSServerInternal(checkSSLCert bool) error {
	if checkSSLCert {
		if err := checkSSLCertUpdated(); err != nil {
			err = errors.Wrap(err, "fail to start HTTPS")
			log.Errorf("fail to create ssl cert %v", err)
			notifyToTelegram(err.Error())
			return err
		}
	}
	go func() {
		log.Infof("listening https on :%d", httpsPort)
		if err := http.ListenAndServeTLS(
			fmt.Sprintf(":%d", httpsPort),
			SSL_CERT_FILE, SSL_KEY_FILE,
			nil,
		); err != nil {
			err = errors.Wrap(err, "fail to start HTTPS")
			notifyToTelegram(err.Error())
			log.Fatalf("fail to listen and serve https %v", err)
		}
	}()
	// for restart https
	startingHTTPS = false
	return nil
}
