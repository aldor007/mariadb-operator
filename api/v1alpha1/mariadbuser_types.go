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

// MariaDBUserSpec defines the desired state of MariaDBUser
type MariaDBUserSpec struct {
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

// MariaDBUserStatus defines the observed state of MariaDBUser
type MariaDBUserStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MariaDBUser is the Schema for the mariadbusers API
type MariaDBUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MariaDBUserSpec   `json:"spec,omitempty"`
	Status MariaDBUserStatus `json:"status,omitempty"`
}

// GetClusterKey is a helper function that returns the mariadb cluster object key
func (u *MariaDBUser) GetClusterKey() client.ObjectKey {
	ns := u.Spec.ClusterRef.Namespace
	if ns == "" {
		ns = u.Namespace
	}

	return client.ObjectKey{
		Name:      u.Spec.ClusterRef.Name,
		Namespace: ns,
	}
}

//+kubebuilder:object:root=true

// MariaDBUserList contains a list of MariaDBUser
type MariaDBUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MariaDBUser `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MariaDBUser{}, &MariaDBUserList{})
}
