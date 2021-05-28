package rbac

import (
	"context"
	"fmt"
	mariadbv1alpha1 "github.com/aldor007/mariadb-operator/api/v1alpha1"
	"github.com/aldor007/mariadb-operator/resources"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	componentName = "rbac"
)

// Reconciler implements the Component Reconciler
type Reconciler struct {
	resources.Reconciler
}

func NewRBAC(client client.Client, directClient client.Reader, scheme *runtime.Scheme, cluster *mariadbv1alpha1.MariaDBCluster) *Reconciler {
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
	foundSA := &corev1.ServiceAccount{}
	sa := r.CreateServiceAccount()
	err := r.Client.Get(ctx, types.NamespacedName{
		Name:      sa.Name,
		Namespace: r.MariaDBCluster.Namespace,
	}, foundSA)
	if err != nil && errors.IsNotFound(err) {
		// Create the deployment
		log.Info("Creating a new ServiceAccount", "name", sa.Name)
		err = r.Client.Create(ctx, &sa)

		if err != nil {
			// Deployment failed
			log.Error(err, "Failed to create new serviceaccount", "Name", sa.Name)
			return err
		}
	} else if err != nil {
		// Error that isn't due to the deployment not existing
		log.Error(err, "Failed to get sa")
		return err
	}

	foundRole := &rbacv1.Role{}
	role := r.CreateRole()
	err = r.Client.Get(ctx, types.NamespacedName{
		Name:      role.Name,
		Namespace: r.MariaDBCluster.Namespace,
	}, foundRole)
	if err != nil && errors.IsNotFound(err) {
		// Create the deployment
		log.Info("Creating a new Role", "name", role.Name)
		err = r.Client.Create(ctx, &sa)

		if err != nil {
			// Deployment failed
			log.Error(err, "Failed to create new role", "Name", role.Name)
			return err
		}
	} else if err != nil {
		// Error that isn't due to the deployment not existing
		log.Error(err, "Failed to get role")
		return err
	}

	foundRoleBinding := &rbacv1.RoleBinding{}
	roleBinding := r.CreateRoleBinding()
	err = r.Client.Get(ctx, types.NamespacedName{
		Name:      roleBinding.Name,
		Namespace: r.MariaDBCluster.Namespace,
	}, foundRoleBinding)
	if err != nil && errors.IsNotFound(err) {
		// Create the deployment
		log.Info("Creating a new RoleBinding", "name", role.Name)
		err = r.Client.Create(ctx, &sa)

		if err != nil {
			// Deployment failed
			log.Error(err, "Failed to create new roleBinding", "Name", role.Name)
			return err
		}
	} else if err != nil {
		// Error that isn't due to the deployment not existing
		log.Error(err, "Failed to get roleBinding")
		return err
	}

	return nil
}

func (r *Reconciler) CreateRole() rbacv1.Role {
	role := rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-mariadb-list-pods", r.MariaDBCluster.Name),
			Namespace: r.MariaDBCluster.Namespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:     []string{"get", "list"},
				APIGroups: []string{""},
				Resources: []string{"pods/status", "pods"},
			},
		},
	}
	controllerutil.SetControllerReference(r.MariaDBCluster, &role, r.Scheme)
	return role
}

func (r *Reconciler) CreateServiceAccount() corev1.ServiceAccount {
	sa := corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.MariaDBCluster.GetServiceAccountName(),
			Namespace: r.MariaDBCluster.Namespace,
		},
	}
	controllerutil.SetControllerReference(r.MariaDBCluster, &sa, r.Scheme)
	return sa
}

func (r *Reconciler) CreateRoleBinding() rbacv1.RoleBinding {
	roleBinding := rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-list-pods", r.MariaDBCluster.Name),
			Namespace: r.MariaDBCluster.Namespace,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     fmt.Sprintf("%s-mariadb-list-pods", r.MariaDBCluster.Name),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      r.MariaDBCluster.GetServiceAccountName(),
				Namespace: r.MariaDBCluster.Namespace,
			},
		},
	}
	controllerutil.SetControllerReference(r.MariaDBCluster, &roleBinding, r.Scheme)
	return roleBinding
}
