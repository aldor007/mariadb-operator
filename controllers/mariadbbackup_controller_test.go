package controllers_test

import (
	"context"
	"fmt"
	"github.com/aldor007/mariadb-operator/api/v1alpha1"
	"github.com/aldor007/mariadb-operator/controllers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("MariadbBackup Controller", func() {
	const (
		BackupName  = "backup"
		Namespace   = "default"
		ClusterName = "example"
	)

	var (
		s = scheme.Scheme
		r *controllers.MariaDBBackupReconciler
	)

	Context("Reconcile", func() {
		var (
			res     reconcile.Result
			req     reconcile.Request
			backup  *v1alpha1.MariaDBBackup
			cluster *v1alpha1.MariaDBCluster
		)

		BeforeEach(func() {
			req = reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      BackupName,
					Namespace: Namespace,
				},
			}
			s.AddKnownTypes(v1alpha1.GroupVersion, backup)
		})

		Context("create backup job", func() {
			var (
				cl  client.Client
				err error
			)

			BeforeEach(func() {
				backup = &v1alpha1.MariaDBBackup{
					ObjectMeta: metav1.ObjectMeta{
						Name:      BackupName,
						Namespace: Namespace,
					},
					Spec: v1alpha1.MariaDBBackupSpec{
						ClusterRef: v1alpha1.ClusterReference{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: ClusterName,
							},
							Namespace: Namespace,
						},
						BackupSecretName: "secret",
					},
				}
				cluster = &v1alpha1.MariaDBCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      ClusterName,
						Namespace: Namespace,
					},
					Spec: v1alpha1.MariaDBClusterSpec{
						Image: "image",
					},
				}
				err = v1alpha1.AddToScheme(s)
				if err != nil {
					panic(err)
				}
				var fakeObjects []runtime.Object
				fakeObjects = append(fakeObjects, cluster, backup)
				cl = fake.NewClientBuilder().WithScheme(s).WithRuntimeObjects(fakeObjects...).Build()

				r = &controllers.MariaDBBackupReconciler{
					Client: cl,
					Scheme: s,
					Log:    logf.Log,
				}
				res, err = r.Reconcile(context.Background(), req)
			})

			It("shouldn't error", func() {
				Ω(err).To(BeNil())
			})

			It("shouldn't requeue the request", func() {
				Ω(res.Requeue).To(BeFalse())
			})

			It("should create job with proper image", func() {
				var job batchv1.Job
				err = cl.Get(context.TODO(), types.NamespacedName{
					Name:      fmt.Sprintf("%s-%s", "backup", ClusterName),
					Namespace: Namespace,
				}, &job)
				Ω(err).To(BeNil())
				Expect(job.Spec.Template.Spec.Containers[0].Image).To(Equal(cluster.Spec.Image))
			})
		})
		Context("create backup cronjob", func() {
			var (
				cl  client.Client
				err error
			)

			BeforeEach(func() {
				backup = &v1alpha1.MariaDBBackup{
					ObjectMeta: metav1.ObjectMeta{
						Name:      BackupName,
						Namespace: Namespace,
					},
					Spec: v1alpha1.MariaDBBackupSpec{
						ClusterRef: v1alpha1.ClusterReference{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: ClusterName,
							},
							Namespace: Namespace,
						},
						BackupSecretName: "secret",
						CronExpression:   "22 * * * *",
					},
				}
				cluster = &v1alpha1.MariaDBCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      ClusterName,
						Namespace: Namespace,
					},
					Spec: v1alpha1.MariaDBClusterSpec{
						Image: "image",
					},
				}
				err = v1alpha1.AddToScheme(s)
				Expect(err).To(BeNil())
				var fakeObjects []runtime.Object
				fakeObjects = append(fakeObjects, cluster, backup)
				cl = fake.NewClientBuilder().WithScheme(s).WithRuntimeObjects(fakeObjects...).Build()

				r = &controllers.MariaDBBackupReconciler{
					Client: cl,
					Scheme: s,
					Log:    logf.Log,
				}
				res, err = r.Reconcile(context.Background(), req)
			})

			It("shouldn't error", func() {
				Ω(err).To(BeNil())
			})

			It("shouldn't requeue the request", func() {
				Ω(res.Requeue).To(BeFalse())
			})

			It("should create cronjob with proper image", func() {
				var job batchv1beta.CronJob
				err = cl.Get(context.TODO(), types.NamespacedName{
					Name:      fmt.Sprintf("%s-%s", "backup", ClusterName),
					Namespace: Namespace,
				}, &job)
				Ω(err).To(BeNil())
				Expect(job.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Image).To(Equal(cluster.Spec.Image))
			})
		})
		Context("update backup cronjob", func() {
			var (
				cl  client.Client
				err error
			)

			BeforeEach(func() {
				backup = &v1alpha1.MariaDBBackup{
					ObjectMeta: metav1.ObjectMeta{
						Name:      BackupName,
						Namespace: Namespace,
					},
					Spec: v1alpha1.MariaDBBackupSpec{
						ClusterRef: v1alpha1.ClusterReference{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: ClusterName,
							},
							Namespace: Namespace,
						},
						BackupSecretName: "secret",
						CronExpression:   "22 * * * *",
					},
				}
				cronJob := &batchv1beta.CronJob{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("%s-%s", "backup", ClusterName),
						Namespace: Namespace,
					},
					Spec:   batchv1beta.CronJobSpec{},
					Status: batchv1beta.CronJobStatus{},
				}
				cluster = &v1alpha1.MariaDBCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      ClusterName,
						Namespace: Namespace,
					},
					Spec: v1alpha1.MariaDBClusterSpec{
						Image: "image",
					},
				}
				err = v1alpha1.AddToScheme(s)
				Expect(err).To(BeNil())
				var fakeObjects []runtime.Object
				fakeObjects = append(fakeObjects, cluster, backup, cronJob)
				cl = fake.NewClientBuilder().WithScheme(s).WithRuntimeObjects(fakeObjects...).Build()

				r = &controllers.MariaDBBackupReconciler{
					Client: cl,
					Scheme: s,
					Log:    logf.Log,
				}
				res, err = r.Reconcile(context.Background(), req)
			})

			It("shouldn't error", func() {
				Ω(err).To(BeNil())
			})

			It("shouldn't requeue the request", func() {
				Ω(res.Requeue).To(BeFalse())
			})

			It("should create cronjob with proper image", func() {
				var job batchv1beta.CronJob
				err = cl.Get(context.TODO(), types.NamespacedName{
					Name:      fmt.Sprintf("%s-%s", "backup", ClusterName),
					Namespace: Namespace,
				}, &job)
				Ω(err).To(BeNil())
				Expect(job.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Image).To(Equal(cluster.Spec.Image))
				Expect(job.Annotations["mariadb/config"]).To(Equal(backup.GetConfigHash()))
			})
		})
	})
})
