package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/bombsimon/logrusr/v3"
	"github.com/go-logr/logr"
	"github.com/konveyor-ecosystem/k8s-provider/provider"
	libprovider "github.com/konveyor/analyzer-lsp/provider"
	"github.com/sirupsen/logrus"
)

var (
	InitConfigPath    string
	ConditionInfoPath string
)

func init() {
	flag.StringVar(&InitConfigPath, "initConfig", "settings.json", "path to initConfig json")
	flag.StringVar(&ConditionInfoPath, "conditionInfo", "condition.json", "path to condition info")
}

func main() {
	flag.Parse()
	err := runCLI()
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

func runCLI() (err error) {
	log := setupLogging()
	prv := provider.New()
	bytes, err := os.ReadFile(InitConfigPath)
	if err != nil {
		return
	}
	initConfig := libprovider.InitConfig{}
	err = json.Unmarshal(bytes, &initConfig)
	if err != nil {
		return
	}
	conditionInfo, err := os.ReadFile(ConditionInfoPath)
	if err != nil {
		return
	}
	srv, err := prv.Init(context.TODO(), log, initConfig)
	if err != nil {
		return
	}
	resp, err := srv.Evaluate(context.TODO(), provider.CapabilityRego, conditionInfo)
	if err != nil {
		return
	}
	responseJson, err := json.Marshal(resp)
	if err != nil {
		return
	}
	fmt.Printf("%s\n", responseJson)
	return
}
