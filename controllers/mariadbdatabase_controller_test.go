package controllers_test

import (
	"context"
	"fmt"
	"github.com/aldor007/mariadb-operator/api/v1alpha1"
	"github.com/aldor007/mariadb-operator/controllers"
	mysqlMock "github.com/aldor007/mariadb-operator/mocks/mysql"
	"github.com/aldor007/mariadb-operator/mysql"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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



type queryMatcher struct {
	query mysql.Query
}

func (m queryMatcher) Matches(arg interface{}) bool {
	q := arg.(mysql.Query)
	if q.String() == m.query.String() {
		return true
	}
	return false
}

func (m queryMatcher) String() string {
	return m.query.String()
}


func EqQuery(q mysql.Query) gomock.Matcher {
	return queryMatcher{query: q}
}

var _ = Describe("MariadbDatabase Controller", func() {
	const (
		Namespace   = "default"
		ClusterName = "example"
		dbName      = "db-name"
	)

	var (
		s = scheme.Scheme
		r *controllers.MariaDBDatabaseReconciler
	)

	Context("Reconcile", func() {
		var (
			res      reconcile.Result
			req      reconcile.Request
			db       *v1alpha1.MariaDBDatabase
			cluster  *v1alpha1.MariaDBCluster
			mockCtrl *gomock.Controller
		)

		BeforeEach(func() {
			req = reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      dbName,
					Namespace: Namespace,
				},
			}
			s.AddKnownTypes(v1alpha1.GroupVersion, db)
		})

		When("create Mariadb database", func() {
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
								Name: "mariadb-secret-key",
							},
							Key: "root",
						},
						DataStorageSize: "1Gi",
					},
				}
				rootSecret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "mariadb-secret-key",
						Namespace: Namespace,
					},
					Immutable: nil,
					Data:      nil,
					StringData: map[string]string{
						"root": "root-password",
					},
					Type: "Opaque",
				}
				db = &v1alpha1.MariaDBDatabase{
					ObjectMeta: metav1.ObjectMeta{
						Name:      dbName,
						Namespace: Namespace,
					},
					Spec: v1alpha1.MariaDBDatabaseSpec{
						ClusterRef: v1alpha1.ClusterReference{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: ClusterName,
							},
							Namespace: Namespace,
						},
						Database: dbName,
					},
				}
				err = v1alpha1.AddToScheme(s)
				if err != nil {
					panic(err)
				}
				var fakeObjects []runtime.Object
				fakeObjects = append(fakeObjects, cluster, db, rootSecret)
				cl = fake.NewClientBuilder().WithScheme(s).WithRuntimeObjects(fakeObjects...).Build()
				mockCtrl = gomock.NewController(GinkgoT())
				sqlRunner := mysqlMock.NewMockSQLRunner(mockCtrl)
				sqlRunner.EXPECT().QueryExec(gomock.Any(), EqQuery(mysql.NewQuery(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", mysql.Escape(dbName))))).Return(nil)

				r = &controllers.MariaDBDatabaseReconciler{
					Client: cl,
					Scheme: s,
					Log:    logf.Log,
					SQLRunnerFactory: func(_ *mysql.Config, errs ...error) (mysql.SQLRunner, func(), error) {
						return sqlRunner, func() {

						}, nil
					},
				}
				res, err = r.Reconcile(context.Background(), req)
			})

			AfterEach(func() {
				mockCtrl.Finish()
			})

			It("shouldn't error", func() {
				Ω(err).To(BeNil())
			})

			It("shouldn't requeue the request", func() {
				Ω(res.Requeue).To(BeFalse())
			})
		})
	})
})
