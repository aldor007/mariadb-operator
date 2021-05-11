package backup

import (
	"context"
	mariadbv1alpha1 "github.com/aldor007/mariadb-operator/api/v1alpha1"
	"github.com/aldor007/mariadb-operator/resources"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	componentName = "backup"
)

// Reconciler implements the Component Reconciler
type Reconciler struct {
	resources.Reconciler
}

func NewBackup(client client.Client, directClient client.Reader, scheme *runtime.Scheme, cluster *mariadbv1alpha1.MariaDBCluster) *Reconciler {
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
	return nil
}

//func (r *Reconciler) createCronJobs(cron *mariadbv1alpha1.MariaDBBackup) {
//	filename := "/var/lib/mysql/backup_`date +%F_%T`.sql"
//
//	backupCommand := "echo 'Starting DB Backup'  &&  " +
//		"mariadump -P 3306 -h '" + r.MariaDBCluster.GetPrimaryHeadlessSvc() +
//		"' --compress > " + filename +
//		"&& echo 'Completed DB Backup'"
//
//	job := batchv1.CronJob{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      fmt.Sprintf("%s-%s", "backup-", r.MariaDBCluster.Name),
//			Namespace: r.MariaDBCluster.Namespace,
//		},
//		Spec: batchv1.CronJobSpec{
//			Schedule: cron.Spec.CronExpression,
//			JobTemplate: batchv1.JobTemplateSpec{
//				Spec: v1.JobSpec{
//					Parallelism:           nil,
//					Completions:           nil,
//					ActiveDeadlineSeconds: nil,
//					BackoffLimit:          nil,
//					Selector:              nil,
//					ManualSelector:        nil,
//					Template: core.PodTemplateSpec{
//						ObjectMeta: metav1.ObjectMeta{},
//						Spec: core.PodSpec{
//							Volumes: nil,
//							Containers: []core.Container{
//								{
//									Name:    "bakcup",
//									Image:   r.MariaDBCluster.Spec.Image,
//									Command: []string{"/bin/sh", "-c"},
//									Args:    []string{backupCommand},
//									EnvFrom: []core.EnvFromSource{
//										{
//											SecretRef: &core.SecretEnvSource{
//												LocalObjectReference: core.LocalObjectReference{
//													Name: cron.Spec.BackupSecretName,
//												},
//											},
//										},
//									},
//									Env:           nil,
//									VolumeMounts:  nil,
//									VolumeDevices: nil,
//								},
//							},
//							RestartPolicy:      "",
//							ServiceAccountName: "",
//							ImagePullSecrets:   nil,
//						},
//					},
//				},
//			},
//		},
//	}
//}
