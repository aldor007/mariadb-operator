package resources

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/aldor007/mariadb-operator/api/v1alpha1"
	"github.com/aldor007/mariadb-operator/utils"
)

// Reconciler holds:
// - cached client : split client reading cached/watched resources from informers and writing to api-server
// - direct client : to read non-watched resources
// - MariaDBCluster CR
type Reconciler struct {
	client.Client
	DirectClient   client.Reader
	MariaDBCluster *v1alpha1.MariaDBCluster
	Scheme         *runtime.Scheme
}

func (r *Reconciler) createPV() corev1.PersistentVolume {
	labels := utils.Labels(r.MariaDBCluster)
	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:   fmt.Sprintf("%s-%s-pv", r.MariaDBCluster.Name, r.MariaDBCluster.Namespace),
			Labels: labels,
		},
		Spec: corev1.PersistentVolumeSpec{
			StorageClassName: r.MariaDBCluster.Spec.StorageClass,
			Capacity: corev1.ResourceList{
				corev1.ResourceName(corev1.ResourceStorage): resource.MustParse(r.MariaDBCluster.Spec.DataStorageSize),
			},
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
		},
	}
	controllerutil.SetControllerReference(r.MariaDBCluster, pv, r.Scheme)
	return *pv
}

func (r *Reconciler) createPVC() corev1.PersistentVolumeClaim {
	labels := utils.Labels(r.MariaDBCluster)
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s-pvc", r.MariaDBCluster.Name, r.MariaDBCluster.Namespace),
			Namespace: r.MariaDBCluster.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &r.MariaDBCluster.Spec.StorageClass,
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceName(corev1.ResourceStorage): resource.MustParse(r.MariaDBCluster.Spec.DataStorageSize),
				},
			},
			VolumeName: fmt.Sprintf("%s-%s-pvc", r.MariaDBCluster.Name, r.MariaDBCluster.Namespace),
		},
	}

	controllerutil.SetControllerReference(r.MariaDBCluster, pvc, r.Scheme)
	return *pvc
}

func (r *Reconciler) GetConfigAnnotation() string {
	return "mariadb/config"
}

// ComponentReconciler describes the Reconcile method
type ComponentReconciler interface {
	Reconcile(ctx context.Context, log logr.Logger) error
}

// Resource simple function without parameter
type Resource func() runtime.Object
