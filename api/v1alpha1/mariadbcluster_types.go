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
	"k8s.io/apimachinery/pkg/util/intstr"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MariaDBClusterSpec defines the desired state of MariaDBCluster
type MariaDBClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// PrimartCount number of master pods
	PrimaryCount int32 `json:"primaryCount,omitempty"`

	// number of replica pods
	ReplicaCount int32 `json:"replicaCount,omitempty"`

	// Image used for mariadb server
	Image string `json:"image,omitempty"`

	// A bucket URL that contains a xtrabackup to initialize the mysql database.
	// +optional
	InitBucketURL string `json:"initBucketURL,omitempty"`

	// A map[string]string that will be passed to my.cnf file.
	// +optional
	MariaDBConf MariaDBConf `json:"mariaDBConf,omitempty"`
}

// MariaDBConf defines type for extra cluster configs. It's a simple map between
// string and string.
type MariaDBConf map[string]intstr.IntOrString

// MariaDBClusterStatus defines the observed state of MariaDBCluster
type MariaDBClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MariaDBCluster is the Schema for the mariadbclusters API
type MariaDBCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MariaDBClusterSpec   `json:"spec,omitempty"`
	Status MariaDBClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MariaDBClusterList contains a list of MariaDBCluster
type MariaDBClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MariaDBCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MariaDBCluster{}, &MariaDBClusterList{})
}
