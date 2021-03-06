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
	"crypto/sha256"
	"encoding/hex"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MariaDBBackupSpec defines the desired state of MariaDBBackup
type MariaDBBackupSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ClusterRef represents a reference to the MySQL cluster.
	// This field should be immutable.
	ClusterRef ClusterReference `json:"clusterRef"`

	// BackupURL represents the URL to the backup location
	BackupURL string `json:"backupURL"`

	// BackupSecretName the name of secrets that contains the credentials to
	BackupSecretName string `json:"backupSecretName"`

	// BackupDBName the name of db to backup
	// +optional
	BackupDBName string `json:"backupDBName,omitempty"`

	// CronExpression represents cron syntax for kubernetes CronJob
	// +optional
	CronExpression string `json:"cron,omitempty"`
}

// MariaDBBackupStatus defines the observed state of MariaDBBackup
type MariaDBBackupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MariaDBBackup is the Schema for the mariadbbackups API
type MariaDBBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MariaDBBackupSpec   `json:"spec,omitempty"`
	Status MariaDBBackupStatus `json:"status,omitempty"`
}

func (db *MariaDBBackup) GetClusterKey() client.ObjectKey {
	ns := db.Spec.ClusterRef.Namespace
	if ns == "" {
		ns = db.Namespace
	}

	return client.ObjectKey{
		Name:      db.Spec.ClusterRef.Name,
		Namespace: ns,
	}
}

func (db *MariaDBBackup) GetConfigHash() string {
	h := sha256.New()
	h.Write([]byte(db.Spec.CronExpression))
	h.Write([]byte(db.Spec.BackupURL))
	h.Write([]byte(db.Spec.BackupDBName))
	return hex.EncodeToString(h.Sum(nil))
}

//+kubebuilder:object:root=true

// MariaDBBackupList contains a list of MariaDBBackup
type MariaDBBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MariaDBBackup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MariaDBBackup{}, &MariaDBBackupList{})
}
