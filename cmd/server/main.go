package main

import (
	"context"
	"flag"

	liblogr "github.com/jortel/go-utils/logr"
	"github.com/konveyor-ecosystem/k8s-provider/provider"
	libprovider "github.com/konveyor/analyzer-lsp/provider"
)

var (
	Port int
)

func init() {
	flag.IntVar(&Port, "port", 0, "Port to serve on.")
}

func main() {
	flag.Parse()
	log := liblogr.WithName("k8s")
	prv := provider.New()
	srv := libprovider.NewServer(prv, Port, log)
	err := srv.Start(context.TODO())
	if err != nil {
		panic(err)
	}
}
