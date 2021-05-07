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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MariaDBUserSpec defines the desired state of MariaDBUser
type MariaDBUserSpec struct {
	// ClusterRef represents a reference to the MySQL cluster.
	// This field should be immutable.
	ClusterRef ClusterReference `json:"clusterRef"`

	// User is the name of the user that will be created with will access the specified database.
	// This field should be immutable.
	User string `json:"user"`

	// Password is the password for the user.
	Password corev1.SecretKeySelector `json:"password"`

	// AllowedHosts is the allowed host to connect from.
	AllowedHosts []string `json:"allowedHosts"`

	// Permissions is the list of roles that user has in the specified database.
	Permissions []MariaDBPermission `json:"permissions,omitempty"`

	// ResourceLimits allow settings limit per mysql user as defined here:
	// https://mariadb.com/kb/en/create-user/
	// +optional
	ResourceLimits MariaDBUserLimits `json:"limits,omitempty"`
}

type MariaDBUserLimits struct {
	MaxQueriesPerHour     int `json:"MAX_QUERIES_PER_HOUR"`
	MaxUpdatePerHour      int `json:"MAX_UPDATES_PER_HOUR"`
	MaxConnectionsPerHour int `json:"MAX_CONNECTIONS_PER_HOUR"`
	MaxUserConnections    int `json:"MAX_USER_CONNECTIONS"`
	MaxStatementTime      int `json:"MAX_STATEMENT_TIME"`
}

func (l MariaDBUserLimits) Get() map[string]int {
	result := make(map[string]int)
	if l.MaxConnectionsPerHour != 0 {
		result["MAX_CONNECTIONS_PER_HOUR"] = l.MaxConnectionsPerHour
	}

	if l.MaxQueriesPerHour != 0 {
		result["MAX_QUERIES_PER_HOUR"] = l.MaxQueriesPerHour
	}

	if l.MaxUpdatePerHour != 0 {
		result["MAX_UPDATES_PER_HOUR"] = l.MaxUpdatePerHour
	}

	if l.MaxUserConnections != 0 {
		result["MAX_USER_CONNECTIONS"] = l.MaxUserConnections
	}

	return result
}

// MariaDBPermission defines a MariaDB schema permission
type MariaDBPermission struct {
	// Schema represents the schema to which the permission applies
	Schema string `json:"schema"`
	// Tables represents the tables inside the schema to which the permission applies
	Tables []string `json:"tables"`
	// Permissions represents the permissions granted on the schema/tables
	Permissions []string `json:"permissions"`
}
type MariaDBUserCondition struct {
	// Status of the condition, one of True, False, Unknown.
	Status MariaDBStatusType `json:"status"`
	// The last time this condition was updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	// The reason for the condition's last transition.
	Reason string `json:"reason"`
	// A human readable message indicating details about the transition.
	Message string `json:"message"`
}

// MariaDBUserStatus defines the observed state of MariaDBUser
type MariaDBUserStatus struct {
	// Conditions represents the MysqlUser resource conditions list.
	// +optional
	Condition MariaDBUserCondition `json:"conditions,omitempty"`

	// AllowedHosts contains the list of hosts that the user is allowed to connect from.
	AllowedHosts []string `json:"allowedHosts,omitempty"`
}

// MariaDBUser is the Schema for the mariadbusers API
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type == 'Ready')].status",description="The user status"
// +kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".spec.clusterRef.name"
// +kubebuilder:printcolumn:name="UserName",type="string",JSONPath=".spec.user"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
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

func (u *MariaDBUser) UpdateStatusCondition(status MariaDBStatusType, reason string, message string) {
	u.Status.Condition.Status = status
	u.Status.Condition.LastUpdateTime = metav1.NewTime(time.Now())
	u.Status.Condition.Reason = reason
	u.Status.Condition.Message = message
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
