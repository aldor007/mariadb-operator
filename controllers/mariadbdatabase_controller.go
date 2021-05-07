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
	"github.com/aldor007/mariadb-operator/mysql"
	"github.com/aldor007/mariadb-operator/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mariadbv1alpha1 "github.com/aldor007/mariadb-operator/api/v1alpha1"
)

const (
	mariadbPreventDeletionFinalizer = "mariadb-operator.mkaciuba.com/database"
)

// MariaDBDatabaseReconciler reconciles a MariaDBDatabase object
type MariaDBDatabaseReconciler struct {
	client.Client
	Log              logr.Logger
	Scheme           *runtime.Scheme
	SQLRunnerFactory mysql.SQLRunnerFactory
}

//+kubebuilder:rbac:groups=mariadb.mkaciuba.com,resources=mariadbdatabases,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mariadb.mkaciuba.com,resources=mariadbdatabases/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mariadb.mkaciuba.com,resources=mariadbdatabases/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// the MariaDBDatabase object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *MariaDBDatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("mariadbdatabase", req.NamespacedName)

	instance := &mariadbv1alpha1.MariaDBDatabase{}
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
	// if the user has been deleted then remove it from mysql cluster
	if !instance.ObjectMeta.DeletionTimestamp.IsZero() {
		err = r.deleteDatabase(ctx, instance, log)
		if err != nil {
			return reconcile.Result{}, err
		}

		// remove finalizer
		utils.RemoveFinalizer(&instance.ObjectMeta, mariadbPreventDeletionFinalizer)

		// update resource to remove finalizer, no status
		return reconcile.Result{}, r.Update(ctx, instance)
	}
	// reconcile database in mysql
	err = r.createDatabase(ctx, instance, log)
	if err != nil {
		return reconcile.Result{}, nil //r.updateReadyCondition(ctx, oldDBStatus, db, err)
	}

	// Add finalier if needed
	if !utils.HasFinalizer(&instance.ObjectMeta, mariadbPreventDeletionFinalizer) {
		utils.AddFinalizer(&instance.ObjectMeta, mariadbPreventDeletionFinalizer)
		if uErr := r.Update(ctx, instance); uErr != nil {
			return reconcile.Result{}, uErr
		}
	}

	return reconcile.Result{}, nil // r.updateReadyCondition(ctx, oldDBStatus, db, err)

}
func (r *MariaDBDatabaseReconciler) deleteDatabase(ctx context.Context, db *mariadbv1alpha1.MariaDBDatabase, log logr.Logger) error {
	log.Info("deleting MySQL database", "name", db.Name, "database", db.Spec.Database)

	sql, closeConn, err := r.SQLRunnerFactory(mysql.NewConfigFromClusterKey(ctx, r.Client, db.GetClusterKey()))
	if errors.IsNotFound(err) {
		// if the mysql cluster does not exists then we can safely assume that
		// the db is deleted so exist successfully

		return err

	} else if err != nil {
		return err
	}
	defer closeConn()
	log.Info("removing database from mysql cluster", "key", db, "database", db.Spec.Database)

	// Remove database from MySQL then remove finalizer
	if err = mysql.DropDatabase(ctx, sql, db.Spec.Database); err != nil {
		return err
	}

	return nil
}

func (r *MariaDBDatabaseReconciler) createDatabase(ctx context.Context, db *mariadbv1alpha1.MariaDBDatabase, log logr.Logger) error {
	log.Info("creating MySQL database", "name", db.Name, "database", db.Spec.Database)

	sql, closeConn, err := r.SQLRunnerFactory(mysql.NewConfigFromClusterKey(ctx, r.Client, db.GetClusterKey()))
	if err != nil {
		return err
	}

	defer closeConn()

	// Create database if does not exists
	return mysql.CreateDatabaseIfNotExists(ctx, sql, db.Spec.Database, db.Spec.CharacterSet, db.Spec.Collation)
}

// SetupWithManager sets up the controller with the Manager.
func (r *MariaDBDatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mariadbv1alpha1.MariaDBDatabase{}).
		Complete(r)
}
