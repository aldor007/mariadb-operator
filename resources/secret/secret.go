package secret

import (
	"context"
	mariadbv1alpha1 "github.com/aldor007/mariadb-operator/api/v1alpha1"
	"github.com/aldor007/mariadb-operator/resources"
	"github.com/aldor007/mariadb-operator/utils"
	"github.com/go-logr/logr"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	componentName = "primary-server"
)

// Reconciler implements the Component Reconciler
type Reconciler struct {
	resources.Reconciler
}

func NewOperatorSecret(client client.Client, directClient client.Reader, scheme *runtime.Scheme, cluster *mariadbv1alpha1.MariaDBCluster) *Reconciler {
	return &Reconciler{
		Reconciler: resources.Reconciler{
			Client:         client,
			Scheme:         scheme,
			DirectClient:   directClient,
			MariaDBCluster: cluster,
		},
	}
}

func (r *Reconciler) Reconcile(ctx context.Context, log logr.Logger) error {
	log = log.WithValues("component", componentName, "clusterName", r.MariaDBCluster.Name, "clusterNamespace", r.MariaDBCluster.Namespace)

	log.V(1).Info("Reconciling")

	secret := &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.MariaDBCluster.GetOperatorSecretName(),
			Namespace: r.MariaDBCluster.Namespace,
		},
	}
	secret.StringData = make(map[string]string)
	secret.StringData["BACKUP_USER"] = "backup"
	secret.StringData["BACKUP_PASSWORD"] = utils.RandString(10)

	err := r.Client.Get(ctx, types.NamespacedName{
		Name:      secret.Name,
		Namespace: secret.Namespace,
	}, secret)

	if err != nil && apierrors.IsNotFound(err) {
		log.Info("creating secret")
		err = r.Client.Create(ctx, secret)
		if err != nil {
			return err
		}
	}
	controllerutil.SetControllerReference(r.MariaDBCluster, secret, r.Scheme)
	return nil
}
