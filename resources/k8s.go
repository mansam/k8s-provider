package resources

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"sync"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor-ecosystem/k8s-provider/config"
	"go.lsp.dev/uri"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ClusterEKS = "eks"
)

func NewClusterResources(client *Client, namespaces []string) (cr *ClusterResources, err error) {
	cr = &ClusterResources{
		client:     client,
		namespaces: namespaces,
		resources:  make(map[schema.GroupVersionKind][]any),
	}

	return
}

type ClusterResources struct {
	client     *Client
	namespaces []string
	resources  map[schema.GroupVersionKind][]any
	lock       sync.Mutex
}

func (r *ClusterResources) Gather(gvks []schema.GroupVersionKind) (resources []any, err error) {
	if len(gvks) == 0 {
		for _, v := range r.resources {
			resources = append(resources, v...)
		}
		return
	}
	for _, gvk := range gvks {
		r.lock.Lock()
		list, found := r.resources[gvk]
		if !found {
			for _, ns := range r.namespaces {
				nsr, rErr := GatherNamespaceResource(r.client, ns, gvk)
				if rErr != nil {
					err = rErr
					return
				}
				r.resources[gvk] = append(r.resources[gvk], nsr...)
			}
			list = r.resources[gvk]
		}
		r.lock.Unlock()
		resources = append(resources, list...)
	}

	return
}

func GatherNamespaceResource(client *Client, namespace string, gvk schema.GroupVersionKind) (resources []any, err error) {
	ul := unstructured.UnstructuredList{}
	ul.SetGroupVersionKind(gvk)
	err = client.List(context.TODO(), &ul, &k8s.ListOptions{Namespace: namespace})
	if err != nil {
		return
	}
	for _, item := range ul.Items {
		resources = append(resources, item.Object)
	}
	return
}

// NewClient builds new k8s client.
func NewClient(cluster config.Cluster) (client *Client, err error) {
	if cluster.Kind == ClusterEKS {
		err = PrepareEKSCredentials(cluster)
		if err != nil {
			return
		}
	}
	bytes, err := os.ReadFile(cluster.KubeConfigPath)
	if err != nil {
		fmt.Printf("Reading kubeconfig failed: %s\n", err)
		return
	}
	config, err := clientcmd.NewClientConfigFromBytes(bytes)
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

func PrepareEKSCredentials(cluster config.Cluster) (err error) {
	// copy aws credentials into container home dir so that
	// they can be discovered by tools
	home, _ := os.UserHomeDir()
	p := path.Join(home, ".aws", "credentials")
	err = MkDir(path.Join(home, ".aws"), 0755)
	if err != nil {
		err = liberr.Wrap(
			err,
			"path",
			p)
		fmt.Printf("%s\n", err)
		return
	}

	dest, err := os.Create(p)
	if err != nil {
		err = liberr.Wrap(
			err,
			"path",
			p)
		fmt.Printf("%s\n", err)
		return
	}

	src, err := os.ReadFile(cluster.CredentialsPath)
	if err != nil {
		err = liberr.Wrap(
			err,
			"path",
			cluster.CredentialsPath)
		fmt.Printf("%s\n", err)
		return
	}

	_, err = dest.Write(src)
	if err != nil {
		err = liberr.Wrap(
			err,
			"path",
			p)
		fmt.Printf("%s\n", err)
		return
	}

	cmd := exec.Command(
		"/usr/local/bin/eksctl",
		"utils", "write-kubeconfig",
		"--cluster", cluster.Name,
		"--region", cluster.Region,
		"--kubeconfig", cluster.KubeConfigPath,
	)
	err = cmd.Run()

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

// MkDir ensures the directory exists.
func MkDir(path string, mode os.FileMode) (err error) {
	err = os.MkdirAll(path, mode)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			err = nil
		} else {
			err = liberr.Wrap(
				err,
				"path",
				path)
		}
	}
	return
}

// Exists return if the path exists.
func Exists(path string) (found bool, err error) {
	_, err = os.Stat(path)
	if err == nil {
		found = true
		return
	}
	if !os.IsNotExist(err) {
		err = liberr.Wrap(
			err,
			"path",
			path)
	} else {
		err = nil
	}
	return
}
