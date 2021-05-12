package backup

import (
	"context"
	"fmt"
	mariadbv1alpha1 "github.com/aldor007/mariadb-operator/api/v1alpha1"
	"github.com/aldor007/mariadb-operator/resources"
	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/batch/v1"
	batchv1beta "k8s.io/api/batch/v1beta1"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	componentName = "backup"
)

// Reconciler implements the Component Reconciler
type Reconciler struct {
	resources.Reconciler
	backup *mariadbv1alpha1.MariaDBBackup
}

func NewBackupJobs(client client.Client, directClient client.Reader, scheme *runtime.Scheme, cluster *mariadbv1alpha1.MariaDBCluster, backup *mariadbv1alpha1.MariaDBBackup) *Reconciler {
	return &Reconciler{
		Reconciler: resources.Reconciler{
			Client:         client,
			Scheme:         scheme,
			DirectClient:   directClient,
			MariaDBCluster: cluster,
		},
		backup: backup,
	}
}

func (r *Reconciler) Reconcile(ctx context.Context, log logr.Logger) error {
	log = log.WithValues("component", componentName, "clusterName", r.MariaDBCluster.Name, "clusterNamespace", r.MariaDBCluster.Namespace)

	log.V(1).Info("Reconciling")
	if r.backup.Spec.CronExpression != "" {
		job := r.createCronJobs(r.backup)
		err := r.Client.Get(ctx, types.NamespacedName{
			Namespace: r.backup.Namespace,
			Name:      job.Name,
		}, &job)
		if err != nil && apierrors.IsNotFound(err) {
			// Create the deployment
			log.Info("Creating a new cronjob", "name", job.Name)
			err = r.Client.Create(ctx, &job)
			if err != nil {
				// Deployment failed
				log.Error(err, "Failed to create new cronjob", "Name", job.Name)
				return err
			}
		}
	} else {
		job := r.createJob(r.backup)
		err := r.Client.Get(ctx, types.NamespacedName{
			Namespace: r.backup.Namespace,
			Name:      job.Name,
		}, &job)
		if err != nil && apierrors.IsNotFound(err) {
			// Create the deployment
			log.Info("Creating a new job", "name", job.Name)
			err = r.Client.Create(ctx, &job)
			if err != nil {
				// Deployment failed
				log.Error(err, "Failed to create new cronjob", "Name", job.Name)
				return err
			}
		}
	}

	return nil
}

func (r *Reconciler) createCronJobs(cron *mariadbv1alpha1.MariaDBBackup) batchv1beta.CronJob {

	job := batchv1beta.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", "backup", r.MariaDBCluster.Name),
			Namespace: r.MariaDBCluster.Namespace,
		},
		Spec: batchv1beta.CronJobSpec{
			Schedule: cron.Spec.CronExpression,
			JobTemplate: batchv1beta.JobTemplateSpec{
				Spec: v1.JobSpec{
					Parallelism:           nil,
					Completions:           nil,
					ActiveDeadlineSeconds: nil,
					BackoffLimit:          nil,
					Selector:              nil,
					ManualSelector:        nil,
					Template:              r.createPodTemplate(cron),
				},
			},
		},
	}
	controllerutil.SetControllerReference(r.MariaDBCluster, &job, r.Scheme)
	return job
}
func (r *Reconciler) createJob(cron *mariadbv1alpha1.MariaDBBackup) batchv1.Job {

	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", "backup", r.MariaDBCluster.Name),
			Namespace: r.MariaDBCluster.Namespace,
		},
		Spec: batchv1.JobSpec{
			Parallelism:             nil,
			Completions:             nil,
			ActiveDeadlineSeconds:   nil,
			BackoffLimit:            nil,
			Selector:                nil,
			ManualSelector:          nil,
			Template:                r.createPodTemplate(cron),
			TTLSecondsAfterFinished: nil,
		},
	}
	controllerutil.SetControllerReference(r.MariaDBCluster, &job, r.Scheme)
	return job
}
func (r *Reconciler) createPodTemplate(cron *mariadbv1alpha1.MariaDBBackup) core.PodTemplateSpec {
	return core.PodTemplateSpec{
		Spec: core.PodSpec{
			Volumes: nil,
			Containers: []core.Container{
				{
					Name:    "backup",
					Image:   r.MariaDBCluster.Spec.Image,
					Command: []string{"/bin/sh", "-c"},
					Args:    []string{"/usr/bin/create-backup.sh"},
					EnvFrom: []core.EnvFromSource{
						{
							SecretRef: &core.SecretEnvSource{
								LocalObjectReference: core.LocalObjectReference{
									Name: cron.Spec.BackupSecretName,
								},
							},
						},
						{

							SecretRef: &core.SecretEnvSource{
								LocalObjectReference: core.LocalObjectReference{
									Name: r.MariaDBCluster.GetOperatorSecretName(),
								},
							},
						},
					},
					Env: []core.EnvVar{
						{
							Name:  "CLUSTER_NAME",
							Value: r.MariaDBCluster.Name,
						},
						{
							Name:  "BACKUP_URL",
							Value: cron.Spec.BackupURL,
						},
						{
							Name:  "HOST",
							Value: r.MariaDBCluster.GetPrimaryHeadlessSvc(),
						},
						{
							Name:  "PORT",
							Value: "3306",
						},
					},
					VolumeMounts:  nil,
					VolumeDevices: nil,
				},
			},
			RestartPolicy: core.RestartPolicyOnFailure,
		},
	}
}
