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
	"errors"
	"github.com/aldor007/mariadb-operator/mysql"
	"github.com/aldor007/mariadb-operator/utils"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mariadbv1alpha1 "github.com/aldor007/mariadb-operator/api/v1alpha1"
)
const (
	userFinalizer =  "mariadb-operator.mkaciuba.com/user"
)
// MariaDBUserReconciler reconciles a MariaDBUser object
type MariaDBUserReconciler struct {
	client.Client
	Log              logr.Logger
	Scheme           *runtime.Scheme
	SQLRunnerFactory mysql.SQLRunnerFactory
}

//+kubebuilder:rbac:groups=mariadb.mkaciuba.com,resources=mariadbusers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mariadb.mkaciuba.com,resources=mariadbusers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mariadb.mkaciuba.com,resources=mariadbusers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MariaDBUser object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *MariaDBUserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("mariadbuser", req.NamespacedName)
	user := &mariadbv1alpha1.MariaDBUser{}
	err := r.Get(ctx, req.NamespacedName, user)
	if err != nil {
	if apiErrors.IsNotFound(err) {
	// Object not found, return. Created objects are automatically garbage collected.
	// For additional cleanup logic use finalizers.
	return reconcile.Result{}, nil
	}
	// Error reading the object - requeue the request.
	return reconcile.Result{}, err
	}

	// if the user has been deleted then remove it from mysql cluster
	if !user.ObjectMeta.DeletionTimestamp.IsZero() {
	return reconcile.Result{}, r.removeUser(ctx, user, log)
	}

	 r.createUser(ctx, user, log)
	// enqueue the resource again after to keep the resource up to date in mysql
	// in case is changed directly into mysql
	return reconcile.Result{
		Requeue:      true,
		RequeueAfter: 2 * time.Minute,
	}, nil
}
func (r *MariaDBUserReconciler) removeUser(ctx context.Context, user *mariadbv1alpha1.MariaDBUser, log logr.Logger) error {
	// The resource has been deleted
	if utils.HasFinalizer(&user.ObjectMeta, userFinalizer) {
		// Drop the user if the finalizer is still present
		if err := r.dropUserFromDB(ctx, user, log); err != nil {
			return err
		}

		utils.RemoveFinalizer(&user.ObjectMeta, userFinalizer)

		// update resource so it will remove the finalizer
		if err := r.Update(ctx, user); err != nil {
			return err
		}
	}
	return nil
}

func (r *MariaDBUserReconciler) dropUserFromDB(ctx context.Context, user *mariadbv1alpha1.MariaDBUser, log logr.Logger) error {
	sql, closeConn, err := r.SQLRunnerFactory(mysql.NewConfigFromClusterKey(ctx, r.Client, user.GetClusterKey()))
	defer closeConn()
	if apiErrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	for _, host := range user.Status.AllowedHosts {
		log.Info("removing user from mysql cluster",  "username", user.Spec.User, "cluster", user.GetClusterKey())
		if err := mysql.DropUser(ctx, sql, user.Spec.User, host); err != nil {
			return err
		}
	}
	return nil
}

func (r *MariaDBUserReconciler) reconcileUserInDB(ctx context.Context, user *mariadbv1alpha1.MariaDBUser, log logr.Logger) error {
	sql, closeConn, err := r.SQLRunnerFactory(mysql.NewConfigFromClusterKey(ctx, r.Client, user.GetClusterKey()))
	if err != nil {
		return err
	}
	defer closeConn()

	secret := &corev1.Secret{}
	secretKey := client.ObjectKey{Name: user.Spec.Password.Name, Namespace: user.Namespace}

	if err := r.Get(ctx, secretKey, secret); err != nil {
		return err
	}

	password := string(secret.Data[user.Spec.Password.Key])
	if password == "" {
		return errors.New("the MariaDB user's password must not be empty")
	}

	// create/ update user in database
	log.Info("creating mysql user",  "username", user.Spec.User, "cluster", user.GetClusterKey())
	if err := mysql.CreateUserIfNotExists(ctx, sql, user.Spec.User, password, user.Spec.AllowedHosts,
		user.Spec.Permissions, user.Spec.ResourceLimits); err != nil {
		return err
	}

	// remove allowed hosts for user
	toRemove := utils.StringDiffIn(user.Status.AllowedHosts, user.Spec.AllowedHosts)
	for _, host := range toRemove {
		if err := mysql.DropUser(ctx, sql, user.Spec.User, host); err != nil {
			return err
		}
	}

	return nil
}
func (r *MariaDBUserReconciler) creatUser(ctx context.Context, user *mariadbv1alpha1.MariaDBUser, log logr.Logger) (err error) {
	// catch the error and set the failed status
	defer setFailedStatus(&err, user)

	// Reconcile the user into mysql
	if err = r.reconcileUserInDB(ctx, user); err != nil {
		return
	}

	// add finalizer if is not added on the resource
	if !utils.HasFinalizer(&user.ObjectMeta, userFinalizer) {
		utils.AddFinalizer(&user.ObjectMeta, userFinalizer)
		if err = r.Update(ctx, user); err != nil {
			return
		}
	}

	// update status for allowedHosts if needed, mark that status need to be updated
	if !reflect.DeepEqual(user.Status.AllowedHosts, user.Spec.AllowedHosts) {
		user.Status.AllowedHosts = user.Spec.AllowedHosts
	}

	// Update the status according to the result
	user.UpdateStatusCondition(
		mysqlv1alpha1.MySQLUserReady, corev1.ConditionTrue,
		mysqluser.ProvisionSucceededReason, "The user provisioning has succeeded.",
	)

	return
}


// SetupWithManager sets up the controller with the Manager.
func (r *MariaDBUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mariadbv1alpha1.MariaDBUser{}).
		Complete(r)
}
