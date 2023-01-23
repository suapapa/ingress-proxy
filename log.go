package main

import (
	"fmt"
	"os"
	"time"

	"github.com/evalphobia/logrus_fluent"
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
		Out:       os.Stderr,
		Level:     logrus.InfoLevel,
		Hooks:     make(logrus.LevelHooks),
		Formatter: newLogFormatter(),
	}

	hostname, _ := os.Hostname()

	log = logger.WithFields(logrus.Fields{
		"hostname": hostname,
		"program":  programName,
		"ver":      programVer,
	})

	tryCnt := 5
	var fluentHook *logrus_fluent.FluentHook
	var err error
	for tryCnt > 0 {
		// fluent hook
		fluentHook, err = logrus_fluent.NewWithConfig(logrus_fluent.Config{
			Host: "localhost",
			Port: 24224,
		})
		if err != nil {
			fmt.Printf("fail to connect fluentd (remain cnt: %d)", tryCnt)
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}

	fluentHook.SetMessageField("message")
	fluentHook.SetLevels([]logrus.Level{
		logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel, logrus.WarnLevel,
		logrus.InfoLevel,
	})
	fluentHook.SetTag("app." + programName)
	logger.AddHook(fluentHook)
}

// log formatter to print log in KST timezone
type logFommater struct {
	ptf *prefixed.TextFormatter
	loc *time.Location
}

func newLogFormatter() *logFommater {
	ptf := prefixed.TextFormatter{
		ForceColors:     true,
		TimestampFormat: time.RFC3339,
		FullTimestamp:   true,
		ForceFormatting: true,
	}

	return &logFommater{
		ptf: &ptf,
		loc: time.FixedZone("KST", +9*60*60),
	}
}

func (f *logFommater) Format(e *logrus.Entry) ([]byte, error) {
	e.Time = e.Time.In(f.loc)
	return f.ptf.Format(e)
}
