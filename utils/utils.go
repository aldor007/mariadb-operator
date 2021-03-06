package utils

import (
	"github.com/aldor007/mariadb-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

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

func StringDiffIn(actual, desired []string) []string {
	diff := []string{}
	for _, str := range actual {
		// if is not in the desired list remove it
		if _, exists := StringIn(str, desired); !exists {
			diff = append(diff, str)
		}
	}

	return diff
}

func StringIn(str string, strs []string) (int, bool) {
	for i, s := range strs {
		if s == str {
			return i, true
		}
	}
	return 0, false
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
