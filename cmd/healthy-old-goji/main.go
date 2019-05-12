package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"
	"syscall"
	"time"

	"github.com/goji/glogrus"
	"github.com/sirupsen/logrus"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/graceful"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
)

var myName string
var logger *logrus.Entry

func init() {
	myName = path.Base(os.Args[0])

	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		// GKE向けのフィールド名に置換え
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
	})

	logger = logrus.WithFields(logrus.Fields{
		"app": myName,
	})
}

func main() {
	noGojiAccessLog := flag.Bool("no-access-log", false, "Disable access log of goji")
	waitSec := flag.Int("wait", 30, "Seconds to sleep before PreHook ended")
	flag.Parse()

	goji.Abandon(middleware.Logger)
	if !*noGojiAccessLog {
		goji.Use(glogrus.NewGlogrus(logger.Logger, "goji"))
	}

	goji.Use(func(c *web.C, h http.Handler) http.Handler {
		return middleware.NoCache(h)
	})

	ctxSignal, cancelSignal := context.WithCancel(context.Background())
	defer cancelSignal()
	goji.Get("/", func(c web.C, w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "!\n")
	})
	goji.Get("/healthz", func(c web.C, w http.ResponseWriter, r *http.Request) {
		select {
		case <-ctxSignal.Done():
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, "ng\n")
		default:
			fmt.Fprintf(w, "ok\n")
		}
	})
	graceful.AddSignal(syscall.SIGTERM, syscall.SIGINT)
	graceful.PreHook(func() {
		logger.Infof("PreHook: sleep %d secs before return the func", *waitSec)
		cancelSignal()
		time.Sleep(time.Duration(*waitSec) * time.Second)
	})

	goji.Serve()
}
