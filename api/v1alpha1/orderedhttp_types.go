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

// OrderedHttpSpec defines the desired state of OrderedHttp
type OrderedHttpSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Replicas is the number of PODs that will be maintained by the operator
	Replicas int32 `json:"replicas"`
}

// OrderedHttpStatus defines the observed state of OrderedHttp
type OrderedHttpStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// PodNames is the list of running PODs maintained by the operator
	PodNames []string `json:"podnames"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// OrderedHttp is the Schema for the orderedhttps API
type OrderedHttp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OrderedHttpSpec   `json:"spec,omitempty"`
	Status OrderedHttpStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OrderedHttpList contains a list of OrderedHttp
type OrderedHttpList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OrderedHttp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OrderedHttp{}, &OrderedHttpList{})
}
