package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/konveyor-ecosystem/k8s-provider/config"
	"github.com/konveyor-ecosystem/k8s-provider/resources"
	libprovider "github.com/konveyor/analyzer-lsp/provider"
	"github.com/open-policy-agent/opa/rego"
	"go.lsp.dev/uri"
	"gopkg.in/yaml.v3"
)

const (
	ProviderName = "k8s-resource"
)

// Capabilities
const (
	CapabilityRegoModule     = "rego_module"
	CapabilityRegoExpression = "rego_expr"
)

// K8sInitConfig is the provider init config with the k8s provider-specific fields unmarshalled.
type K8sInitConfig struct {
	libprovider.InitConfig
	ProviderSpecificConfig struct {
		Source config.Cluster `json:"source"`
		//Destination Cluster `json:"destination"`
	}
}

// NewK8sInitConfig creates a k8s specific provider configuration from the generic provider init.
func NewK8sInitConfig(initCfg libprovider.InitConfig) (k *K8sInitConfig, err error) {
	k = &K8sInitConfig{InitConfig: initCfg}
	psc, err := json.Marshal(initCfg.ProviderSpecificConfig)
	if err != nil {
		return
	}
	fmt.Println(k)
	err = json.Unmarshal(psc, &k.ProviderSpecificConfig)
	if err != nil {
		return
	}

	return
}

// New constructs a new K8s provider.
func New() (k *K8s) {
	k = &K8s{}
	return
}

// K8s provider
type K8s struct {
	ctx         context.Context
	cfg         *K8sInitConfig
	k8sClient   *resources.Client
	baseModules func(r *rego.Rego)
	resources   resources.Resources
	log         logr.Logger
}

// Init the provider. Reads in base Rego modules, kubeconfig, and pulls resources from the cluster.
func (r *K8s) Init(ctx context.Context, log logr.Logger, initCfg libprovider.InitConfig) (svc libprovider.ServiceClient, err error) {
	r.cfg, err = NewK8sInitConfig(initCfg)
	if err != nil {
		return
	}
	r.ctx = ctx
	r.log = log
	r.baseModules = rego.Module("inventory.rego", resources.InventoryModule)

	r.k8sClient, err = resources.NewClient(r.cfg.ProviderSpecificConfig.Source)
	if err != nil {
		fmt.Printf("Configuring client failed: %s.", err)
		return
	}
	r.resources, err = resources.NewClusterResources(r.k8sClient, r.cfg.ProviderSpecificConfig.Source.Namespaces)

	svc = r
	return
}

// Capabilities returns the supported capabilities of the provider.
func (r *K8s) Capabilities() (caps []libprovider.Capability) {
	caps = []libprovider.Capability{
		{
			Name: CapabilityRegoExpression,
		},
		{
			Name: CapabilityRegoModule,
		},
	}
	return
}

// Evaluate a capability and return a result.
func (r *K8s) Evaluate(ctx context.Context, cap string, conditionBytes []byte) (resp libprovider.ProviderEvaluateResponse, err error) {
	condition := ConditionInfo{}
	err = yaml.Unmarshal(conditionBytes, &condition)
	if err != nil {
		return
	}
	switch cap {
	case CapabilityRegoExpression:
		resp, err = r.evaluateRegoExpression(ctx, condition.Expression)
	case CapabilityRegoModule:
		resp, err = r.evaluateRegoModule(ctx, condition.Module)
	}

	return
}

func (r *K8s) Stop() {
	fmt.Println("Goodbye (remain running).")
	//os.Exit(0)
}
func (r *K8s) GetDependencies(ctx context.Context) (deps map[uri.URI][]*libprovider.Dep, err error) {
	return
}
func (r *K8s) GetDependenciesDAG(ctx context.Context) (dag map[uri.URI][]libprovider.DepDAGItem, err error) {
	return
}

// evaluate a rego_expr rule
func (r *K8s) evaluateRegoExpression(ctx context.Context, condition ExpressionCondition) (resp libprovider.ProviderEvaluateResponse, err error) {
	policy := rego.Module("policy.rego", fmt.Sprintf(ExpressionTemplate, condition.Collection, condition.Expression))
	prepared, err := rego.New(
		rego.Query("incidents = data.policy.incidents"),
		r.baseModules,
		policy,
	).PrepareForEval(ctx)
	if err != nil {
		return
	}

	rs, err := r.resources.Gather(r.cfg.ProviderSpecificConfig.Source.GroupVersionKinds)
	if err != nil {
		return
	}
	resultSet, err := prepared.Eval(ctx, rego.EvalInput(rs))
	if err != nil {
		return
	}
	resp, err = r.interpretResultSet(resultSet)
	return
}

// evaluate a rego_module rule
func (r *K8s) evaluateRegoModule(ctx context.Context, condition ModuleCondition) (resp libprovider.ProviderEvaluateResponse, err error) {
	if err != nil {
		return
	}
	policy := rego.Module("policy.rego", condition.Module)
	prepared, err := rego.New(
		rego.Query("incidents = data.policy.incidents"),
		r.baseModules,
		policy,
	).PrepareForEval(ctx)
	if err != nil {
		return
	}

	var res []any
	if !condition.Defaults {
		res, err = r.resources.Gather(r.cfg.ProviderSpecificConfig.Source.GroupVersionKinds)
		if err != nil {
			return
		}
	}
	rs, err := r.resources.Gather(condition.Resources)
	if err != nil {
		return
	}
	res = append(res, rs...)

	resultSet, err := prepared.Eval(ctx, rego.EvalInput(res))
	if err != nil {
		return
	}
	resp, err = r.interpretResultSet(resultSet)
	return
}

// interpret a rego result set as a ProviderEvaluteResponse
func (r *K8s) interpretResultSet(results rego.ResultSet) (resp libprovider.ProviderEvaluateResponse, err error) {
	if len(results) == 0 {
		return
	}
	incidents, ok := results[0].Bindings["incidents"].([]interface{})
	if !ok {
		err = errors.New("unexpected result format")
		return
	}
	if len(incidents) == 0 {
		return
	}
	resp.Matched = true
	for _, i := range incidents {
		var u uri.URI
		var incident RegoIncident
		bytes, _ := json.Marshal(i)
		_ = json.Unmarshal(bytes, &incident)
		u, err = r.getResourceURI(incident)
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

// delegate to the k8s client to resolve resource URIs
func (r *K8s) getResourceURI(i RegoIncident) (u uri.URI, err error) {
	group, version := i.GroupVersion()
	u, err = r.k8sClient.GetResourceURI(group, version, i.Kind, i.Namespace, i.Name)
	return
}
