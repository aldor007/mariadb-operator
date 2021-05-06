package utils

import (
	"github.com/aldor007/mariadb-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Labels(cluster *v1alpha1.MariaDBCluster) map[string]string {
	return map[string]string{
		"app":             "MariaDB",
		"mariadb/cluster": cluster.Name,
		"MariaDB_cr":      cluster.Name,
	}
}

// AddFinalizer add a finalizer in ObjectMeta.
func AddFinalizer(meta *metav1.ObjectMeta, finalizer string) {
	if !HasFinalizer(meta, finalizer) {
		meta.Finalizers = append(meta.Finalizers, finalizer)
	}
}

// HasFinalizer returns true if ObjectMeta has the finalizer.
func HasFinalizer(meta *metav1.ObjectMeta, finalizer string) bool {
	return containsString(meta.Finalizers, finalizer)
}

// RemoveFinalizer removes the finalizer from ObjectMeta.
func RemoveFinalizer(meta *metav1.ObjectMeta, finalizer string) {
	meta.Finalizers = removeString(meta.Finalizers, finalizer)
}

// containsString is a helper functions to check string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}

	return false
}

// removeString is a helper functions to remove string from a slice of strings.
func removeString(slice []string, s string) []string {
	result := []string{}

	for _, item := range slice {
		if item == s {
			continue
		}

		result = append(result, item)
	}

	return result
}
