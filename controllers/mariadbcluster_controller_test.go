package controllers_test

import (
	"context"
	"github.com/aldor007/mariadb-operator/api/v1alpha1"
	"github.com/aldor007/mariadb-operator/controllers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
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

var _ = Describe("MariadbCluster Controller", func() {
	const (
		Namespace   = "default"
		ClusterName = "example"
	)

	var (
		s = scheme.Scheme
		r *controllers.MariaDBClusterReconciler
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
					Name:      ClusterName,
					Namespace: Namespace,
				},
			}
			s.AddKnownTypes(v1alpha1.GroupVersion, backup)
		})

		When("create Mariadb cluster", func() {
			var (
				cl  client.Client
				err error
			)

			BeforeEach(func() {
				cluster = &v1alpha1.MariaDBCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      ClusterName,
						Namespace: Namespace,
					},
					Spec: v1alpha1.MariaDBClusterSpec{
						Image:        "image",
						PrimaryCount: 3,
						RootPassword: corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "secret-key",
							},
							Key: "root",
						},
						DataStorageSize: "1Gi",
					},
				}
				err = v1alpha1.AddToScheme(s)
				if err != nil {
					panic(err)
				}
				var fakeObjects []runtime.Object
				fakeObjects = append(fakeObjects, cluster)
				cl = fake.NewClientBuilder().WithScheme(s).WithRuntimeObjects(fakeObjects...).Build()

				r = &controllers.MariaDBClusterReconciler{
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

			It("should create statefulset", func() {
				var s appsv1.StatefulSet
				err = cl.Get(context.TODO(), types.NamespacedName{
					Name:      cluster.GetStatefulsetName("primary"),
					Namespace: Namespace,
				}, &s)
				Ω(err).To(BeNil())
				Expect(s.Spec.Template.Spec.Containers[0].Image).To(Equal(cluster.Spec.Image))
				Expect(*s.Spec.Replicas).To(Equal(cluster.Spec.PrimaryCount))
			})

			It("should create headless svc", func() {
				var svc corev1.Service
				err = cl.Get(context.TODO(), types.NamespacedName{
					Name:      cluster.GetPrimaryHeadlessSvcName(),
					Namespace: Namespace,
				}, &svc)
				Ω(err).To(BeNil())
				Expect(svc.Spec.ClusterIP).To(Equal("None"))
			})

			It("shouldn't create svc", func() {
				var svc corev1.Service
				err = cl.Get(context.TODO(), types.NamespacedName{
					Name:      cluster.GetPrimarySvcName(),
					Namespace: Namespace,
				}, &svc)
				Ω(err).NotTo(BeNil())
			})
		})
		When("create Mariadb with svc", func() {
			var (
				cl  client.Client
				err error
			)

			BeforeEach(func() {
				cluster = &v1alpha1.MariaDBCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      ClusterName,
						Namespace: Namespace,
					},
					Spec: v1alpha1.MariaDBClusterSpec{
						Image:        "image",
						PrimaryCount: 3,
						RootPassword: corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "secret-key",
							},
							Key: "root",
						},
						DataStorageSize: "1Gi",
						ServiceConf: v1alpha1.ServiceConf{
							Enabled:        true,
							LoadbalancerIP: "1.2.3.4",
							Type:           "LoadBalancer",
						},
					},
				}
				err = v1alpha1.AddToScheme(s)
				Expect(err).To(BeNil())
				var fakeObjects []runtime.Object
				fakeObjects = append(fakeObjects, cluster)
				cl = fake.NewClientBuilder().WithScheme(s).WithRuntimeObjects(fakeObjects...).Build()

				r = &controllers.MariaDBClusterReconciler{
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

			It("should create svc", func() {
				var svc corev1.Service
				err = cl.Get(context.TODO(), types.NamespacedName{
					Name:      cluster.GetPrimarySvcName(),
					Namespace: Namespace,
				}, &svc)
				Ω(err).To(BeNil())
				Expect(svc.Spec.Type).To(Equal(corev1.ServiceTypeLoadBalancer))
				Expect(svc.Spec.LoadBalancerIP).To(Equal("1.2.3.4"))
			})
		})
	})
})
