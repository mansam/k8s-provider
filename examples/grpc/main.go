package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	liblogr "github.com/jortel/go-utils/logr"
	liboutput "github.com/konveyor/analyzer-lsp/output/v1/konveyor"
	"github.com/konveyor/analyzer-lsp/provider"
	"github.com/konveyor/analyzer-lsp/provider/grpc"
	"gopkg.in/yaml.v3"
)

var (
	KubeConfigPath string
	RulesPaths     string
	ServerAddr     string
	Namespaces     string
	Stop           bool
)

func init() {
	flag.StringVar(&KubeConfigPath, "kubeconfig", ".kube/config", "path to kubeconfig")
	flag.StringVar(&RulesPaths, "rules", "", "comma-delimited list of paths to rulesets")
	flag.StringVar(&ServerAddr, "server", "127.0.0.1:9000", "address of provider gRPC server")
	flag.StringVar(&Namespaces, "namespaces", "konveyor-tackle", "comma-delimited list of namespaces to analyze")
	flag.BoolVar(&Stop, "stop", false, "send stop signal to gRPC server after evaluating rules.")
}

func main() {
	flag.Parse()
	log := liblogr.WithName("k8s")
	config, err := ConstructProviderConfig()
	if err != nil {
		panic(err)
	}
	client := grpc.NewGRPCClient(config, log)
	err = client.Start(context.TODO())
	if err != nil {
		panic(err)
	}
	err = client.ProviderInit(context.TODO())
	if err != nil {
		panic(err)
	}
	ruleSets, err := ReadRulesets(strings.Split(RulesPaths, ","))
	if err != nil {
		return
	}
	results := []ResultRuleset{}
	for _, rs := range ruleSets {
		result, eErr := EvaluateRuleset(client, rs)
		if eErr != nil {
			return
		}
		results = append(results, result)
	}

	dump, _ := json.Marshal(results)
	fmt.Printf("%s\n", dump)
	if Stop {
		client.Stop()
	}
}

func ConstructProviderConfig() (config provider.Config, err error) {
	kubeConfigBytes, err := os.ReadFile(KubeConfigPath)
	if err != nil {
		panic(err)
	}

	var namespaces []interface{}
	for _, s := range strings.Split(Namespaces, ",") {
		namespaces = append(namespaces, s)
	}
	config.Name = "k8s"
	config.Address = ServerAddr
	config.Proxy = &provider.Proxy{}
	config.InitConfig = []provider.InitConfig{
		{
			Proxy: &provider.Proxy{},
			ProviderSpecificConfig: map[string]interface{}{
				"groupVersionKinds": []interface{}{
					map[string]interface{}{"group": "apps", "version": "v1", "kind": "Deployment"},
					map[string]interface{}{"group": "", "version": "v1", "kind": "Pod"},
					map[string]interface{}{"group": "route.openshift.io", "version": "v1", "kind": "Route"},
				},
				"kubeConfig": kubeConfigBytes,
				"namespaces": namespaces,
			},
		},
	}
	return
}

func EvaluateRuleset(svc provider.ServiceClient, rs RuleSet) (result ResultRuleset, err error) {
	result.Name = rs.Name
	result.Description = rs.Description
	result.Violations = make(map[string]Violation)
	result.Errors = make(map[string]string)
	for _, r := range rs.Rules {
		var resp provider.ProviderEvaluateResponse
		var bytes []byte
		if condInfo, ok := r.When["k8s.rego_expr"]; ok {
			bytes, err = json.Marshal(condInfo)
			if err != nil {
				return
			}
			resp, err = svc.Evaluate(context.TODO(), "rego_expr", bytes)
		} else if condInfo, ok = r.When["k8s.rego_module"]; ok {
			bytes, err = json.Marshal(condInfo)
			if err != nil {
				return
			}
			resp, err = svc.Evaluate(context.TODO(), "rego_module", bytes)
		}
		if err != nil {
			result.Errors[r.RuleID] = err.Error()
			continue
		}
		if !resp.Matched {
			result.Unmatched = append(result.Unmatched, r.RuleID)
			continue
		}
		category := liboutput.Category(r.Category)
		v := Violation{
			Description: r.Message,
			Category:    &category,
			Labels:      nil,
			Incidents:   nil,
			Links:       nil,
			Extras:      nil,
			Effort:      &r.Effort,
		}
		for _, incidentContext := range resp.Incidents {
			incident := liboutput.Incident{
				URI:       incidentContext.FileURI,
				Message:   r.Message,
				Variables: incidentContext.Variables,
			}
			v.Incidents = append(v.Incidents, incident)
		}
		result.Violations[r.RuleID] = v
	}

	return
}

type ResultRuleset struct {
	Name        string               `yaml:"name,omitempty" json:"name,omitempty"`
	Description string               `yaml:"description,omitempty" json:"description,omitempty"`
	Tags        []string             `yaml:"tags,omitempty" json:"tags,omitempty"`
	Violations  map[string]Violation `yaml:"violations,omitempty" json:"violations,omitempty"`
	Errors      map[string]string    `yaml:"errors,omitempty" json:"errors,omitempty"`
	Unmatched   []string             `yaml:"unmatched,omitempty" json:"unmatched,omitempty"`
	Skipped     []string             `yaml:"skipped,omitempty" json:"skipped,omitempty"`
	Labels      []string             `yaml:"labels,omitempty" json:"labels,omitempty"`
}

type Violation struct {
	Description string               `yaml:"description" json:"description"`
	Category    *liboutput.Category  `yaml:"category,omitempty" json:"category,omitempty"`
	Labels      []string             `yaml:"labels,omitempty" json:"labels,omitempty"`
	Incidents   []liboutput.Incident `yaml:"incidents" json:"incidents"`
	Links       []liboutput.Link     `yaml:"links,omitempty" json:"links,omitempty"`
	Extras      json.RawMessage      `yaml:"extras,omitempty" json:"extras,omitempty"`
	Effort      *int                 `yaml:"effort,omitempty" json:"effort,omitempty"`
}

type RulesRegistry struct {
	RuleSets []string `json:"ruleSets"`
}

type RuleSet struct {
	Name        string
	Description string
	Rules       []Rule `yaml:"-"`
}

type Rule struct {
	RuleID   string `json:"ruleID" yaml:"ruleID"`
	Effort   int
	Category string
	Message  string
	When     map[string]map[string]string
}

const (
	RulesetYaml = "ruleset.yaml"
	RuleSuffix  = ".yaml"
)

func ReadRulesets(paths []string) (ruleSets []RuleSet, err error) {
	for _, p := range paths {
		ruleSet, rErr := ReadRuleset(p)
		if rErr != nil {
			err = rErr
			return
		}
		ruleSets = append(ruleSets, ruleSet)
	}
	return
}

func ReadRuleset(ruleSetDir string) (rs RuleSet, err error) {
	rulesetYaml := path.Join(ruleSetDir, RulesetYaml)
	f, err := os.Open(rulesetYaml)
	if err != nil {
		return
	}
	defer f.Close()

	rs = RuleSet{}
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&rs)
	if err != nil {
		return
	}

	entries, err := os.ReadDir(ruleSetDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if entry.Name() == RulesetYaml {
			continue
		}
		if !strings.HasSuffix(entry.Name(), RuleSuffix) {
			continue
		}
		err = func() (err error) {
			filePath := path.Join(ruleSetDir, entry.Name())
			f, err := os.Open(filePath)
			if err != nil {
				return
			}
			defer f.Close()

			decoder := yaml.NewDecoder(f)

			for {
				rule := Rule{}
				err = decoder.Decode(&rule)
				if err != nil {
					if errors.Is(err, io.EOF) {
						err = nil
						break
					}
					return
				}
				rs.Rules = append(rs.Rules, rule)
			}
			return
		}()
	}
	return
}
