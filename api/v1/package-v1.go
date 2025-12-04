package v1

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HelloSpec defines the desired state
type HelloSpec struct {
    Message string `json:"message,omitempty"`
}

// HelloStatus defines the observed state
type HelloStatus struct {
    Seen bool `json:"seen,omitempty"`
}

//+kubebuilder:object:root=true
type Hello struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   HelloSpec   `json:"spec,omitempty"`
    Status HelloStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
type HelloList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []Hello `json:"items"`
}