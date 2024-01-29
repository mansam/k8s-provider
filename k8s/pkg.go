package k8s

import (
	"context"
	"fmt"

	"go.lsp.dev/uri"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
)

func NewUnstructuredResources(client k8s.Client) (ur *UnstructuredResources) {
	ur = &UnstructuredResources{
		client:          client,
		NamespacedLists: make(map[string][]unstructured.UnstructuredList),
	}
	return
}

// UnstructuredResources is a namespace-separated cache of unstructured k8s resources.
type UnstructuredResources struct {
	client          k8s.Client
	NamespacedLists map[string][]unstructured.UnstructuredList `json:"namespaces"`
}

// Gather unstructured resources that match the provided GVK and namespace.
func (r *UnstructuredResources) Gather(namespace string, gvks []schema.GroupVersionKind) (err error) {
	for _, gvk := range gvks {
		ul := unstructured.UnstructuredList{}
		ul.SetGroupVersionKind(gvk)
		err = r.client.List(context.TODO(), &ul, &k8s.ListOptions{Namespace: namespace})
		if err != nil {
			return
		}
		r.NamespacedLists[namespace] = append(r.NamespacedLists[namespace], ul)
	}
	return
}

// NewClient builds new k8s client.
func NewClient(kubeConfig []byte) (client *Client, err error) {
	config, err := clientcmd.NewClientConfigFromBytes(kubeConfig)
	if err != nil {
		return
	}
	restCfg, err := config.ClientConfig()
	if err != nil {
		return
	}
	k8sClient, err := k8s.New(
		restCfg,
		k8s.Options{
			Scheme: scheme.Scheme,
		})
	if err != nil {
		return
	}
	client = &Client{
		Client: k8sClient,
		Host:   restCfg.Host,
	}
	return
}

// Client is a k8s client.
type Client struct {
	k8s.Client
	Host string
}

// GetResourceURI for a given GVK, namespace, and name.
func (r *Client) GetResourceURI(group, version, kind, namespace, name string) (u uri.URI, err error) {
	gk := schema.GroupKind{Group: group, Kind: kind}
	mapping, err := r.RESTMapper().RESTMapping(gk, version)
	if err != nil {
		return
	}
	var path string
	if mapping.Resource.Group == "" {
		path = fmt.Sprintf("%s/api/%s", r.Host, mapping.Resource.Version)
	} else {
		path = fmt.Sprintf("%s/apis/%s/%s", r.Host, mapping.Resource.Group, mapping.Resource.Version)
	}
	if namespace == "" {
		path = fmt.Sprintf("%s/%s/%s", path, mapping.Resource, name)
	} else {
		path = fmt.Sprintf("%s/namespaces/%s/%s/%s", path, namespace, mapping.Resource.Resource, name)
	}

	u, _ = uri.Parse(path)
	return
}
