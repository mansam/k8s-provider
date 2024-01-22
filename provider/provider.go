package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/go-logr/logr"
	"github.com/konveyor-ecosystem/k8s-provider/k8s"
	libprovider "github.com/konveyor/analyzer-lsp/provider"
	"github.com/open-policy-agent/opa/rego"
	"go.lsp.dev/uri"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	CapabilityRego   = "rego"
	CapabilitySkopeo = "skopeo"
)

type RegoConditionInfo struct {
	Policy string `json:"policy"`
}

type PolicyIncident struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	ApiVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
}

func (r PolicyIncident) GroupVersion() (group string, version string, ok bool) {
	group, _, found := strings.Cut(r.ApiVersion, "/")
	if !found {
		version = group
		group = ""
	}
	return
}

type K8sInitConfig struct {
	libprovider.InitConfig
	ProviderSpecificConfig struct {
		KubeConfigPath    string                    `json:"kubeConfigPath"`
		BasePoliciesPath  string                    `json:"basePoliciesPath"`
		GroupVersionKinds []schema.GroupVersionKind `json:"groupVersionKinds"`
		Namespaces        []string                  `json:"namespaces"`
	}
}

func NewK8sInitConfig(initCfg libprovider.InitConfig) (k *K8sInitConfig, err error) {
	k = &K8sInitConfig{InitConfig: initCfg}
	psc, err := json.Marshal(initCfg.ProviderSpecificConfig)
	if err != nil {
		return
	}
	err = json.Unmarshal(psc, &k.ProviderSpecificConfig)
	if err != nil {
		return
	}

	return
}

type K8s struct {
	ctx          context.Context
	k8sClient    *k8s.Client
	basePolicies func(r *rego.Rego)
	resources    *k8s.UnstructuredResources
	log          logr.Logger
}

func (r *K8s) Evaluate(ctx context.Context, cap string, conditionInfo []byte) (resp libprovider.ProviderEvaluateResponse, err error) {
	switch cap {
	case CapabilityRego:
		resp, err = r.evaluateRegoPolicy(ctx, conditionInfo)
	case CapabilitySkopeo:
		err = errors.New("not yet implemented")
		return
	}

	return
}

func (r *K8s) evaluateRegoPolicy(ctx context.Context, conditionInfo []byte) (resp libprovider.ProviderEvaluateResponse, err error) {
	regoConditionInfo := &RegoConditionInfo{}
	err = json.Unmarshal(conditionInfo, regoConditionInfo)
	if err != nil {
		return
	}
	policy := rego.Module("policy.rego", regoConditionInfo.Policy)
	prepared, err := rego.New(
		rego.Query("incidents = data.policy.incidents"),
		r.basePolicies,
		policy,
	).PrepareForEval(ctx)
	if err != nil {
		return
	}
	resultSet, err := prepared.Eval(ctx, rego.EvalInput(r.resources))
	if err != nil {
		return
	}
	if len(resultSet) == 0 {
		return
	}
	incidents, ok := resultSet[0].Bindings["incidents"].([]interface{})
	if !ok {
		err = errors.New("unknown result")
		return
	}
	resp.Matched = true
	for _, i := range incidents {
		var u uri.URI
		var incident PolicyIncident
		bytes, _ := json.Marshal(i)
		_ = json.Unmarshal(bytes, &incident)
		u, err = r.GetResourceURI(incident)
		if err != nil {
			return
		}
		ic := libprovider.IncidentContext{
			FileURI:   u,
			Variables: i.(map[string]interface{}),
		}
		resp.Incidents = append(resp.Incidents, ic)
	}
	return
}

func (r *K8s) Stop() {
	return
}

func (r *K8s) GetDependencies(ctx context.Context) (deps map[uri.URI][]*libprovider.Dep, err error) {
	return
}

func (r *K8s) GetDependenciesDAG(ctx context.Context) (dag map[uri.URI][]libprovider.DepDAGItem, err error) {
	return
}

func (r *K8s) Capabilities() (caps []libprovider.Capability) {
	caps = []libprovider.Capability{
		{
			Name: CapabilityRego,
		},
		{
			Name: CapabilitySkopeo,
		},
	}
	return
}

func (r *K8s) Init(ctx context.Context, log logr.Logger, initCfg libprovider.InitConfig) (svc libprovider.ServiceClient, err error) {
	cfg, err := NewK8sInitConfig(initCfg)
	if err != nil {
		return
	}
	r.ctx = ctx
	r.log = log
	r.basePolicies = rego.Load([]string{cfg.ProviderSpecificConfig.BasePoliciesPath}, nil)

	//log.Info("Reading kubeConfig.")
	bytes, err := os.ReadFile(cfg.ProviderSpecificConfig.KubeConfigPath)
	if err != nil {
		return
	}

	//log.Info("Constructing k8s client.")
	r.k8sClient, err = k8s.NewClient(bytes)
	if err != nil {
		return
	}

	//log.Info("Gathering resources.", "namespaces", cfg.ProviderSpecificConfig.Namespaces)
	r.resources = k8s.NewUnstructuredResources(r.k8sClient)
	for _, ns := range cfg.ProviderSpecificConfig.Namespaces {
		err = r.resources.Gather(ns, cfg.ProviderSpecificConfig.GroupVersionKinds)
		if err != nil {
			return
		}
	}
	svc = r
	return
}

func (r *K8s) GetResourceURI(i PolicyIncident) (u uri.URI, err error) {
	group, version, _ := i.GroupVersion()
	gk := schema.GroupKind{Group: group, Kind: i.Kind}
	mapping, err := r.k8sClient.RESTMapper().RESTMapping(gk, version)
	if err != nil {
		return
	}
	var path string
	if mapping.Resource.Group == "" {
		path = fmt.Sprintf("%s/api/%s", r.k8sClient.Host, mapping.Resource.Version)
	} else {
		path = fmt.Sprintf("%s/apis/%s/%s", r.k8sClient.Host, mapping.Resource.Group, mapping.Resource.Version)
	}
	if i.Namespace == "" {
		path = fmt.Sprintf("%s/%s/%s", path, mapping.Resource, i.Name)
	} else {
		path = fmt.Sprintf("%s/namespaces/%s/%s/%s", path, i.Namespace, mapping.Resource.Resource, i.Name)
	}

	u, _ = uri.Parse(path)
	return
}

func New() (k *K8s) {
	k = &K8s{}
	return
}
