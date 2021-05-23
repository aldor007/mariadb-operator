package primary

import (
	"context"
	"fmt"
	mariadbv1alpha1 "github.com/aldor007/mariadb-operator/api/v1alpha1"
	"github.com/aldor007/mariadb-operator/resources"
	"github.com/aldor007/mariadb-operator/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	componentName = "primary-server"
)

// Reconciler implements the Component Reconciler
type Reconciler struct {
	resources.Reconciler
}

func NewPrimary(client client.Client, directClient client.Reader, scheme *runtime.Scheme, cluster *mariadbv1alpha1.MariaDBCluster) *Reconciler {
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

	statefulSet, err := r.CreateStatefulSet("primary")
	if err != nil {
		return err
	}

	found := &appsv1.StatefulSet{}
	err = r.Client.Get(ctx, types.NamespacedName{
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

	if statefulSet.Annotations == nil || statefulSet.Annotations[r.GetConfigAnnotation()] != r.MariaDBCluster.GetConfigHash() {
		err = r.Client.Update(ctx, &statefulSet)
		if err != nil {
			log.Error(err, "Failed to update Deployment.", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return err
		}
		log.Info("Updated Deployment image. ")
	}

	return nil
}

func (r *Reconciler) CreateStatefulSet(dbType string) (appsv1.StatefulSet, error) {
	labels := utils.Labels(r.MariaDBCluster)
	labels["mariadb/type"] = dbType
	labels["mariadb/pods"] = fmt.Sprintf("%s-%s", r.MariaDBCluster.Name, dbType)

	annotations := make(map[string]string)
	annotations[r.GetConfigAnnotation()] = r.MariaDBCluster.GetConfigHash()
	size := r.MariaDBCluster.Spec.PrimaryCount
	image := r.MariaDBCluster.Spec.Image

	rootPasswordSecret := &corev1.EnvVarSource{
		SecretKeyRef: &r.MariaDBCluster.Spec.RootPassword,
	}
	quantity, err := resource.ParseQuantity(r.MariaDBCluster.Spec.DataStorageSize)
	if err != nil {
		return appsv1.StatefulSet{}, err
	}

	dataVolume := fmt.Sprintf("data-%s", dbType)
	statefulset := appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", r.MariaDBCluster.Name, dbType),
			Namespace: r.MariaDBCluster.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:    &size,
			ServiceName: fmt.Sprintf("%s-%s", r.MariaDBCluster.Name, dbType),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      dataVolume,
						Namespace: r.MariaDBCluster.Namespace,
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
						Selector:    nil,
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceName(corev1.ResourceStorage): quantity,
							},
						},

						StorageClassName: &r.MariaDBCluster.Spec.StorageClass,
					},
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "mariadb",
					Containers: []corev1.Container{{
						Image:           image,
						ImagePullPolicy: corev1.PullIfNotPresent,
						Name:            "mariadb-service",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 3306,
							Name:          "mariadb",
						}},
						ReadinessProbe: &corev1.Probe{
							Handler: corev1.Handler{
								Exec: &corev1.ExecAction{
									Command: []string{
										"/bin/bash",
										"-c",
										"/usr/bin/readiness-probe.sh",
									},
								},
							},
							InitialDelaySeconds: 120,
							TimeoutSeconds:      20,
							PeriodSeconds:       10,
							SuccessThreshold:    5,
							FailureThreshold:    2,
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      dataVolume,
								MountPath: "/var/lib/mysql",
							},
						},
						EnvFrom: []corev1.EnvFromSource{
							{
								SecretRef: &corev1.SecretEnvSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: r.MariaDBCluster.GetOperatorSecretName(),
									},
								},
							},
						},
						Env: []corev1.EnvVar{
							{
								Name:      "MYSQL_ROOT_PASSWORD",
								ValueFrom: rootPasswordSecret,
							},
							{
								Name:  "LABEL_SELECTOR",
								Value: fmt.Sprintf("mariadb/pods=%s-%s", r.MariaDBCluster.Name, dbType),
							},
							{
								Name:  "GALLERA_MODE",
								Value: "yes",
							},
							{
								Name:  "CLUSTER_NAME",
								Value: r.MariaDBCluster.Name,
							},
							{
								Name: "MY_POD_IP",
								ValueFrom: &corev1.EnvVarSource{
									FieldRef: &corev1.ObjectFieldSelector{
										FieldPath: `status.podIP`,
									},
								},
							},
							{
								Name: "MY_POD_NAMESPACE",
								ValueFrom: &corev1.EnvVarSource{
									FieldRef: &corev1.ObjectFieldSelector{
										FieldPath: `metadata.namespace`,
									},
								},
							},
						},
					}},
				},
			},
		},
	}

	controllerutil.SetControllerReference(r.MariaDBCluster, &statefulset, r.Scheme)
	return statefulset, nil
}
