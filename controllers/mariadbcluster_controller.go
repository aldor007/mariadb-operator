/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"github.com/aldor007/mariadb-operator/resources"
	"github.com/aldor007/mariadb-operator/resources/headless"
	"github.com/aldor007/mariadb-operator/resources/primary"
	"github.com/aldor007/mariadb-operator/resources/rbac"
	"github.com/aldor007/mariadb-operator/resources/secret"
	"github.com/aldor007/mariadb-operator/resources/service"
	"github.com/go-logr/logr"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"

	mariadbv1alpha1 "github.com/aldor007/mariadb-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
)

// MariaDBClusterReconciler reconciles a MariaDBCluster object
type MariaDBClusterReconciler struct {
	client.Client
	DirectClient client.Reader
	Log          logr.Logger
	Scheme       *runtime.Scheme
}

//+kubebuilder:rbac:groups=mariadb.mkaciuba.com,resources=MariaDBClusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mariadb.mkaciuba.com,resources=MariaDBClusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mariadb.mkaciuba.com,resources=MariaDBClusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *MariaDBClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("Request.Namespace", req.NamespacedName, "Request.Name", req.Name)

	log.Info("Reconcile MariaDB cluster")
	// Fetch the MariaDB instance
	instance := &mariadbv1alpha1.MariaDBCluster{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}
	reconcilers := []resources.ComponentReconciler{
		secret.NewOperatorSecret(r.Client, r.DirectClient, r.Scheme, instance),
		rbac.NewRBAC(r.Client, r.DirectClient, r.Scheme, instance),
		primary.NewPrimary(r.Client, r.DirectClient, r.Scheme, instance),
		primary.NewPrimary(r.Client, r.DirectClient, r.Scheme, instance),
		headless.NewHeadlessService(r.Client, r.DirectClient, r.Scheme, instance, "primary"),
		service.NewService(r.Client, r.DirectClient, r.Scheme, instance),
	}

	for _, rec := range reconcilers {
		err = rec.Reconcile(ctx, log)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MariaDBClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mariadbv1alpha1.MariaDBCluster{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&policyv1beta1.PodDisruptionBudget{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
