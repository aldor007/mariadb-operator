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
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
type ClusterReference struct {
	corev1.LocalObjectReference `json:",inline"`
	// Namespace the MySQL cluster namespace
	Namespace string `json:"namespace,omitempty"`
}

// MariaDBClusterSpec defines the desired state of MariaDBCluster
type MariaDBClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// PrimartCount number of master pods
	PrimaryCount int32 `json:"primaryCount,omitempty"`

	// number of replica pods
	ReplicaCount int32 `json:"replicaCount,omitempty"`

	// secret reference for password
	RootPassword corev1.SecretKeySelector `json:"rootPassword"`

	// Image used for mariadb server
	// +kubebuilder:default:="ghcr.io/aldor007/mariadb-galera:1.0.1"
	Image string `json:"image,omitempty"`

	StorageClass string `json:"storageClass"`

	// Database storage Size (Ex. 1Gi, 100Mi)
	DataStorageSize string `json:"dataStorageSize"`

	// A bucket URL that contains a xtrabackup to initialize the mysql database.
	// +optional
	InitBucketURL string `json:"initBucketURL,omitempty"`

	// A map[string]string that will be passed to my.cnf file.
	// +optional
	MariaDBConf MariaDBConf `json:"MariaDBConf,omitempty"`

	// ServiceConf represents config for k8s service
	// +optional
	ServiceConf ServiceConf `json:"service,omitempty"`
}

// MariaDBConf defines type for extra cluster configs. It's a simple map between
// string and string.
type MariaDBConf map[string]intstr.IntOrString

// ServiceConf defines kubernetes service config
type ServiceConf struct {
	// Enabled flag indicated if service is enabled
	Enabled bool `json:"enabled,omitempty"`

	// Annotation for service
	Annotation map[string]string `json:"annotation,omitempty"`

	// LoadbalancerIP is a address assigned to service
	LoadbalancerIP string `json:"loadbalancerIP,omitempty"`

	// +kubebuilder:default:="ClusterIP"
	Type corev1.ServiceType `json:"type,omitempty"`
}

// MariaDBClusterStatus defines the observed state of MariaDBCluster
type MariaDBClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MariaDBCluster is the Schema for the MariaDBClusters API
type MariaDBCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MariaDBClusterSpec   `json:"spec,omitempty"`
	Status MariaDBClusterStatus `json:"status,omitempty"`
}

func (c *MariaDBCluster) GetPrimaryAddress() string {
	return fmt.Sprintf("%s.%s", c.GetPrimarySvc(), c.Namespace)
}

func (c *MariaDBCluster) GetPrimarySvc() string {
	return fmt.Sprintf("mariadb-%s-%s", c.Name, "primary")
}

func (c *MariaDBCluster) GetPrimaryHeadlessAddress() string {
	return fmt.Sprintf("%s.%s", c.GetPrimaryHeadlessSvc(), c.Namespace)
}

func (c *MariaDBCluster) GetPrimaryHeadlessSvc() string {
	return fmt.Sprintf("mariadb-headless-%s-%s", c.Name, "primary")
}

func (c *MariaDBCluster) GetOperatorSecretName() string {
	return fmt.Sprintf("mariadb-%s-operated", c.Name)
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
