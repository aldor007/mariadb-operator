package utils

import "github.com/aldor007/mariadb-operator/api/v1alpha1"

func Labels(cluster *v1alpha1.MariaDBCluster) map[string]string {
	return map[string]string{
		"app":             "MariaDB",
		"mariadb/cluster": cluster.Name,
		"MariaDB_cr":      cluster.Name,
	}
}
