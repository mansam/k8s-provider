package k8s

import (
	"context"

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

type Client struct {
	k8s.Client
	Host string
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
