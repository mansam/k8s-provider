package main

import (
	"context"
	"flag"
	"os"

	"github.com/bombsimon/logrusr/v3"
	"github.com/go-logr/logr"
	"github.com/konveyor-ecosystem/k8s-provider/provider"
	libprovider "github.com/konveyor/analyzer-lsp/provider"
	"github.com/sirupsen/logrus"
)

var (
	Port int
)

func init() {
	flag.IntVar(&Port, "port", 0, "Port to serve on.")
}

func main() {
	flag.Parse()
	log := setupLogging()
	prv := provider.New()
	srv := libprovider.NewServer(prv, Port, log)
	err := srv.Start(context.TODO())
	if err != nil {
		panic(err)
	}
}

func setupLogging() (log logr.Logger) {
	l := logrus.New()
	l.SetOutput(os.Stdout)
	l.SetFormatter(&logrus.TextFormatter{})
	l.SetLevel(logrus.Level(5))
	log = logrusr.New(l)
	return
}
