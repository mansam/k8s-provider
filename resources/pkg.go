package resources

import (
	_ "embed"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

//go:embed inventory.rego
var InventoryModule string

type Resources interface {
	Gather(gvks []schema.GroupVersionKind) (resources []any, err error)
}
