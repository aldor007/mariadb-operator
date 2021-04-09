package resources

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
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

func (r *Reconciler) CreateStatefulSet(dbType string) appsv1.StatefulSet {
	labels := utils.Labels(r.MariaDBCluster)
	labels["type.mariadb.org"] = dbType
	size := r.MariaDBCluster.Spec.PrimaryCount
	image := r.MariaDBCluster.Spec.Image

	rootPasswordSecret := &corev1.EnvVarSource{
		SecretKeyRef: &r.MariaDBCluster.Spec.RootPassword,
	}

	dataVolume := "data"
	statefulset := appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprint("%s-%s", r.MariaDBCluster.Name, dbType),
			Namespace: r.MariaDBCluster.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &size,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: nil,
						Selector:    nil,
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceName(corev1.ResourceStorage): resource.MustParse(r.MariaDBCluster.Spec.DataStorageSize),
							},
						},
						VolumeName:       dataVolume,
						StorageClassName: &r.MariaDBCluster.Spec.StorageClass,
					},
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:           image,
						ImagePullPolicy: corev1.PullIfNotPresent,
						Name:            "mariadb-service",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 3306,
							Name:          "mariadb",
						}},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      dataVolume,
								MountPath: "/var/lib/mysql",
							},
						},
						Env: []corev1.EnvVar{
							{
								Name:      "MYSQL_ROOT_PASSWORD",
								ValueFrom: rootPasswordSecret,
							},
						},
					}},
				},
			},
		},
	}

	controllerutil.SetControllerReference(r.MariaDBCluster, &statefulset, r.Scheme)
	return statefulset
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
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
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
			AccessModes:      []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
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

func (r *Reconciler) CreateService(dbType string) corev1.Service {
	labels := utils.Labels(r.MariaDBCluster)
	labels["type.mariadb.org"] = dbType

	s := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("mariadb-%s-%s", r.MariaDBCluster.Name, dbType),
			Namespace: r.MariaDBCluster.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Protocol:   corev1.ProtocolTCP,
				Port:       3306,
				TargetPort: intstr.FromInt(3306),
			}},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	controllerutil.SetControllerReference(r.MariaDBCluster, &s, r.Scheme)
	return s
}

// ComponentReconciler describes the Reconcile method
type ComponentReconciler interface {
	Reconcile(ctx context.Context, log logr.Logger) error
}

// Resource simple function without parameter
type Resource func() runtime.Object
