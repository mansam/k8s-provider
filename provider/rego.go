package provider

import "strings"

// ModuleConditionInfo is the input for the rego.module
// capability, which takes an entire rego module and evaluates it.
type ModuleConditionInfo struct {
	Module string `json:"module"`
}

// ExpressionConditionInfo is the input for the rego.expr
// capability, which takes a single rego expression and injects it
// into a module template which will evaluate it in the context
// of a single resource collection.
type ExpressionConditionInfo struct {
	// Collection is the resource collection from the
	// base module that the expression should be evaluated against.
	Collection string `json:"collection"`
	// Expression is a single rego expression.
	Expression string `json:"expression"`
}

// ExpressionTemplate is the template that the parameters
// from the rego.expr capability will be injected into to
// create a complete module.
const ExpressionTemplate = `package policy
      import data.lib.konveyor
      import future.keywords

      incidents[msg] {
      	some item in data.lib.konveyor.%s
        %s
      	msg := {
            "apiVersion": item.apiVersion,
      		"namespace": item.metadata.namespace,
      		"kind": item.kind,
      		"name": item.metadata.name,
      	}
      }`

// RegoIncident describes the format that the output from
// each Rego rule must take.
type RegoIncident struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	ApiVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
}

// GroupVersion splits the resource's ApiVersion into an API group and a version.
func (r RegoIncident) GroupVersion() (group string, version string) {
	group, _, found := strings.Cut(r.ApiVersion, "/")
	if !found {
		version = group
		group = ""
	}
	return
}
