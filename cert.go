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

	log.Info("(re)start https server")
	go startHTTPSServer()
}

func checkSSLCertUpdated() error {
	if !filesExist(SSL_CERT_FILE) || !filesExist(SSL_KEY_FILE) {
		out, err := exec.Command("/bin/create_ssl_cert.sh").Output()
		if err != nil {
			return errors.Wrap(err, "check ssl fail")
		}
		log.Info(out)
		return errors.New("SSL just created")
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

func startHTTPSServer() {
	startHTTPSMutex.Lock()
	if startingHTTPS {
		return
	}
	startingHTTPS = true
	startHTTPSMutex.Unlock()

	ctx, cancelF := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelF()
	tick := time.NewTicker(1 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Error("fail to launch https server")
			return
		case <-tick.C:
			if err := checkSSLCertUpdated(); err != nil {
				err = errors.Wrap(err, "fail to start HTTPS")
				log.Errorf("fail to start https server %v", err)
				notifyToTelegram(err.Error())
			} else {
				go func() {
					log.Infof("listening https on :%d", httpsPort)
					if err := http.ListenAndServeTLS(
						fmt.Sprintf(":%d", httpsPort),
						SSL_CERT_FILE, SSL_KEY_FILE,
						nil,
					); err != nil {
						err = errors.Wrap(err, "fail to start HTTPS")
						log.Errorf("fail to start https server %v", err)
						notifyToTelegram(err.Error())
					}
				}()
				return
			}
		}
	}
}
