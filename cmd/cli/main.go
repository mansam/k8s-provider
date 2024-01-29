package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	liblogr "github.com/jortel/go-utils/logr"
	"github.com/konveyor-ecosystem/k8s-provider/provider"
	libprovider "github.com/konveyor/analyzer-lsp/provider"
)

var (
	InitConfigPath    string
	ConditionInfoPath string
	Capability        string
)

func init() {
	flag.StringVar(&InitConfigPath, "initConfig", "settings.json", "path to initConfig json")
	flag.StringVar(&ConditionInfoPath, "conditionInfo", "condition.json", "path to condition info")
	flag.StringVar(&Capability, "capability", "rego.policy", "capability")
}

func main() {
	flag.Parse()
	err := runCLI()
	if err != nil {
		panic(err)
	}
}

func runCLI() (err error) {
	log := liblogr.WithName("k8s")
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
	resp, err := srv.Evaluate(context.TODO(), Capability, conditionInfo)
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
