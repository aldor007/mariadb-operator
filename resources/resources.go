package resources

import (
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aldor007/mariadb-operator/api/v1alpha1"
)

// Reconciler holds:
// - cached client : split client reading cached/watched resources from informers and writing to api-server
// - direct client : to read non-watched resources
// - MariaDBCluster CR
type Reconciler struct {
	client.Client
	DirectClient   client.Reader
	MariaDBCluster *v1alpha1.MariaDBCluster
}

// ComponentReconciler describes the Reconcile method
type ComponentReconciler interface {
	Reconcile(log logr.Logger) error
}

// Resource simple function without parameter
type Resource func() runtime.Object
