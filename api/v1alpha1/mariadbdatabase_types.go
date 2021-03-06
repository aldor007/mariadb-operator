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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MariaDBDatabaseSpec defines the desired state of MariaDBDatabase
type MariaDBDatabaseSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ClusterRef represents a reference to the MySQL cluster.
	// This field should be immutable.
	ClusterRef ClusterReference `json:"clusterRef"`

	// Database represents the database name which will be created.
	// This field should be immutable.
	Database string `json:"database"`

	// CharacterSet represents the charset name used when database is created
	CharacterSet string `json:"characterSet,omitempty"`

	// Collation represents the collation name used as default database collation
	Collation string `json:"collation,omitempty"`
}

// MariaDBDatabaseStatus defines the observed state of MariaDBDatabase
type MariaDBDatabaseStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MariaDBDatabase is the Schema for the mariadbdatabases API
type MariaDBDatabase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MariaDBDatabaseSpec   `json:"spec,omitempty"`
	Status MariaDBDatabaseStatus `json:"status,omitempty"`
}

// GetClusterKey is a helper function that returns the mariadb cluster object key
func (db *MariaDBDatabase) GetClusterKey() client.ObjectKey {
	ns := db.Spec.ClusterRef.Namespace
	if ns == "" {
		ns = db.Namespace
	}

	return client.ObjectKey{
		Name:      db.Spec.ClusterRef.Name,
		Namespace: ns,
	}
}

//+kubebuilder:object:root=true

// MariaDBDatabaseList contains a list of MariaDBDatabase
type MariaDBDatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MariaDBDatabase `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MariaDBDatabase{}, &MariaDBDatabaseList{})
}
