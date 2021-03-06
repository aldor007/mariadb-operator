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
	"github.com/aldor007/mariadb-operator/resources/backup"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mariadbv1alpha1 "github.com/aldor007/mariadb-operator/api/v1alpha1"
)

// MariaDBBackupReconciler reconciles a MariaDBBackup object
type MariaDBBackupReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=mariadb.mkaciuba.com,resources=mariadbbackups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mariadb.mkaciuba.com,resources=mariadbbackups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mariadb.mkaciuba.com,resources=mariadbbackups/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MariaDBBackup object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *MariaDBBackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("Request.Namespace", req.NamespacedName, "Request.Name", req.Name)

	// Fetch the MariaDB instance
	backupCr := &mariadbv1alpha1.MariaDBBackup{}
	err := r.Client.Get(ctx, req.NamespacedName, backupCr)
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
	cluster := &mariadbv1alpha1.MariaDBCluster{}
	err = r.Client.Get(ctx, backupCr.GetClusterKey(), cluster)
	if err != nil {
		log.Error(err, "Unable to get cluster")
		return ctrl.Result{}, err
	}

	reconcilers := []resources.ComponentReconciler{
		backup.NewBackupJobs(r.Client, nil, r.Scheme, cluster, backupCr),
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
func (r *MariaDBBackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mariadbv1alpha1.MariaDBBackup{}).
		Complete(r)
}
