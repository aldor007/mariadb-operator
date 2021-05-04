package replica

import (
	"context"
	mariadbv1alpha1 "github.com/aldor007/mariadb-operator/api/v1alpha1"
	"github.com/aldor007/mariadb-operator/resources"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	componentName = "replica-server"
)

// Reconciler implements the Component Reconciler
type Reconciler struct {
	resources.Reconciler
}

func NewReplica(client client.Client, directClient client.Reader, scheme *runtime.Scheme, cluster *mariadbv1alpha1.MariaDBCluster) *Reconciler {
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

	statefulSet := r.Reconciler.CreateStatefulSet("replica")

	found := &appsv1.StatefulSet{}
	err := r.Client.Get(ctx, types.NamespacedName{
		Name:      statefulSet.Name,
		Namespace: r.MariaDBCluster.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {
		// Create the deployment
		log.Info("Creating a new Statefulset", "name", statefulSet.Name)
		err = r.Client.Create(ctx, &statefulSet)

		if err != nil {
			// Deployment failed
			log.Error(err, "Failed to create new statefulset", "Deployment.Name", statefulSet.Name)
			return err
		} else {
			// Deployment was successful
			return nil
		}
	} else if err != nil {
		// Error that isn't due to the deployment not existing
		log.Error(err, "Failed to get Deployment")
		return err
	}

	// Check for any updates for redeployment
	applyChange := false

	// Ensure the deployment size is same as the spec
	size := r.MariaDBCluster.Spec.ReplicaCount
	if size == 0 {
		log.Info("Skipping processing")
		return nil
	}
	if *statefulSet.Spec.Replicas != size {
		statefulSet.Spec.Replicas = &size
		applyChange = true
	}

	// Ensure image name is correct, update image if required
	image := r.MariaDBCluster.Spec.Image
	var currentImage string = ""

	if found.Spec.Template.Spec.Containers != nil {
		currentImage = found.Spec.Template.Spec.Containers[0].Image
	}

	if image != currentImage {
		statefulSet.Spec.Template.Spec.Containers[0].Image = image
		applyChange = true
	}

	if applyChange {
		err = r.Client.Update(ctx, &statefulSet)
		if err != nil {
			log.Error(err, "Failed to update Deployment.", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return err
		}
		log.Info("Updated Deployment image. ")
	}

	svc := r.CreateService("replica")
	foundSvc := &v1.Service{}
	err = r.Client.Get(ctx, types.NamespacedName{
		Name:      svc.Name,
		Namespace: r.MariaDBCluster.Namespace,
	}, foundSvc)
	if err != nil && errors.IsNotFound(err) {
		// Create the deployment
		log.Info("Creating a new svc", "name", svc.Name)
		err = r.Client.Create(ctx, &svc)

		if err != nil {
			// Deployment failed
			log.Error(err, "Failed to create new statefulset", "Deployment.Name", svc.Name)
			return err
		} else {
			// Deployment was successful
			return nil
		}
	} else if err != nil {
		// Error that isn't due to the deployment not existing
		log.Error(err, "Failed to get Deployment")
		return err
	}

	return nil
}
