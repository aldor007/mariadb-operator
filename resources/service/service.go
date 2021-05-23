package service

import (
	"context"
	mariadbv1alpha1 "github.com/aldor007/mariadb-operator/api/v1alpha1"
	"github.com/aldor007/mariadb-operator/resources"
	"github.com/aldor007/mariadb-operator/utils"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	componentName = "service"
)

// Reconciler implements the Component Reconciler
type Reconciler struct {
	resources.Reconciler
}

func NewService(client client.Client, directClient client.Reader, scheme *runtime.Scheme, cluster *mariadbv1alpha1.MariaDBCluster) *Reconciler {
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

	if !r.MariaDBCluster.Spec.ServiceConf.Enabled {
		log.Info("Service not enabled")
		return nil
	}

	svc := r.CreatePrimaryService()
	foundSvc := &v1.Service{}
	err := r.Client.Get(ctx, types.NamespacedName{
		Name:      svc.Name,
		Namespace: r.MariaDBCluster.Namespace,
	}, foundSvc)
	if err != nil && apierrors.IsNotFound(err) {
		// Create the deployment
		log.Info("Creating a new svc", "name", svc.Name)
		err = r.Client.Create(ctx, &svc)

		if err != nil {
			// Deployment failed
			log.Error(err, "Failed to create new service", "service.Name", svc.Name)
			return err
		} else {
			// Deployment was successful
			return nil
		}
	} else if err != nil {
		// Error that isn't due to the deployment not existing
		log.Error(err, "Failed to get service")
		return err
	}

	return nil
}
func (r *Reconciler) CreatePrimaryService() corev1.Service {
	labels := utils.Labels(r.MariaDBCluster)

	s := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        r.MariaDBCluster.GetPrimarySvcName(),
			Namespace:   r.MariaDBCluster.Namespace,
			Annotations: r.MariaDBCluster.Spec.ServiceConf.Annotation,
			Labels:      labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Protocol:   corev1.ProtocolTCP,
				Port:       3306,
				TargetPort: intstr.FromInt(3306),
			}},
			Type:           r.MariaDBCluster.Spec.ServiceConf.Type,
			LoadBalancerIP: r.MariaDBCluster.Spec.ServiceConf.LoadbalancerIP,
		},
	}

	controllerutil.SetControllerReference(r.MariaDBCluster, &s, r.Scheme)
	return s
}
