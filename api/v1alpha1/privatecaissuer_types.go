/*


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
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type PrivateCAIssuerConditionStatus string
type PrivateCAIssuerConditionType string

const (
	PrivateCAIssuerConditionReady PrivateCAIssuerConditionType = "Ready"
	// ConditionTrue represents the fact that a given condition is true
	PrivateCAIssuerConditionTrue PrivateCAIssuerConditionStatus = "True"

	// ConditionFalse represents the fact that a given condition is false
	PrivateCAIssuerConditionFalse PrivateCAIssuerConditionStatus = "False"

	// ConditionUnknown represents the fact that a given condition is unknown
	PrivateCAIssuerConditionUnknown PrivateCAIssuerConditionStatus = "Unknown"
)

type PrivateCAIssuerCondition struct {
	// Type of the condition, currently ('Ready').
	Type PrivateCAIssuerConditionType `json:"type"`

	// Status of the condition, one of ('True', 'False', 'Unknown').
	// +kubebuilder:validation:Enum=True;False;Unknown
	Status PrivateCAIssuerConditionStatus `json:"status"`

	// LastTransitionTime is the timestamp corresponding to the last status
	// change of this condition.
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`

	// Reason is a brief machine readable explanation for the condition's last
	// transition.
	// +optional
	Reason string `json:"reason,omitempty"`

	// Message is a human readable description of the details of the last
	// transition, complementing reason.
	// +optional
	Message string `json:"message,omitempty"`
}

// PrivateCARequestSpec defines the desired state of PrivateCARequest
type PrivateCAIssuerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of PrivateCARequest. Edit PrivateCARequest_types.go to remove/update
	Issuer   string `json:"issuer,omitempty"`
	Location string `json:"location,omitempty"`
	Project  string `json:"project,omitempty"`
}

// PrivateCARequestStatus defines the observed state of PrivateCARequest
type PrivateCAIssuerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Conditions []PrivateCAIssuerCondition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PrivateCARequest is the Schema for the privatecarequests API
type PrivateCAIssuer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PrivateCAIssuerSpec   `json:"spec,omitempty"`
	Status PrivateCAIssuerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PrivateCARequestList contains a list of PrivateCARequest
type PrivateCAIssuerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PrivateCAIssuer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PrivateCAIssuer{}, &PrivateCAIssuerList{})
}
