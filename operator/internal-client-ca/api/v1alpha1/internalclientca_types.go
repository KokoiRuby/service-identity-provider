/*
Copyright 2024.

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

// Certificate defines properties related to the content of the CA certificate.
type Certificate struct {
	Subject `json:"subject"`
}

// Subject defines properties related to the content of the CA certificate.
type Subject struct {
	// The Subject Common Name (CN) of the CA certificate.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=63
	CN string `json:"cn"`
}

// Secret defines properties related to the storage of the certification
type Secret struct {
	// The secret where the CA certificate is stored. The same secret should not be used for multiple purposes.
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`
	Name string `json:"name"`
	// The YAML key name of the CA certificate in the secret.
	// If not given, 'ca.pem' is used.
	// +kubebuilder:validation:Pattern=`^[^\s]+$`
	CertName string `json:"certName,omitempty"`
}

// InternalClientCASpec defines the desired state of InternalClientCA
type InternalClientCASpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Certificate Certificate `json:"certificate"`
	Secret      Secret      `json:"secret"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=intclica;intclicas,singular=internalclientca
// +kubebuilder:printcolumn:name="CN",type=string,JSONPath=`.spec.certificate.subject.cn`
// +kubebuilder:printcolumn:name="Secret",type=string,JSONPath=`.spec.secret.name`

// InternalClientCA is used to request a client CA certificate from sip.
type InternalClientCA struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec InternalClientCASpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// InternalClientCAList contains a list of InternalClientCA
type InternalClientCAList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InternalClientCA `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InternalClientCA{}, &InternalClientCAList{})
}
