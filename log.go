package main

import (
	"os"

	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var (
	log *logrus.Entry
)

func init() {
	initLogger()
}

func initLogger() {
	logger := &logrus.Logger{
		Out:   os.Stderr,
		Level: logrus.WarnLevel,
		Hooks: make(logrus.LevelHooks),
		Formatter: &prefixed.TextFormatter{
			ForceColors:     true,
			TimestampFormat: "2006-01-02 15:04:05",
			FullTimestamp:   true,
			ForceFormatting: true,
		},
	}

	hostname, _ := os.Hostname()

	log = logger.WithFields(logrus.Fields{
		"hostname": hostname,
		"program":  programName,
		"ver":      programVer,
	})
}
