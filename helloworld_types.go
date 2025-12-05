/*
Copyright 2025.

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

// HelloWorldSpec defines the desired state of HelloWorld.
type HelloWorldSpec struct {
    // Message is the message the operator will log/use
    // +kubebuilder:validation:Optional
    Message string `json:"message,omitempty"`

    // Replicas is an optional replica count
    // +kubebuilder:validation:Optional
    Replicas *int32 `json:"replicas,omitempty"`
}

// HelloWorldStatus defines the observed state of HelloWorld.
type HelloWorldStatus struct {
    // Message reflects the last message reconciled
    // +kubebuilder:validation:Optional
    Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// HelloWorld is the Schema for the helloworlds API.
type HelloWorld struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   HelloWorldSpec   `json:"spec,omitempty"`
    Status HelloWorldStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HelloWorldList contains a list of HelloWorld.
type HelloWorldList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []HelloWorld `json:"items"`
}

func init() {
    SchemeBuilder.Register(&HelloWorld{}, &HelloWorldList{})
}