package config

import "k8s.io/apimachinery/pkg/runtime/schema"

type Cluster struct {
	// Kind
	Kind string `json:"kind"` // kube, eks
	// Cluster name for hosted clusters
	Name string
	// Cluster region for hosted clusters
	Region string
	//
	CredentialsPath string `json:"credentialsPath"`
	// cluster's kube config
	KubeConfigPath string `json:"kubeConfigPath"`
	// list of namespaces to collect resources from
	Namespaces []string `json:"namespaces"`
	// list of GVKs to collect in addition to defaults
	GroupVersionKinds []schema.GroupVersionKind `json:"groupVersionKinds"`
}
