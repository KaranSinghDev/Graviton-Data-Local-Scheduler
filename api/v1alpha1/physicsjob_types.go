package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PhysicsJobSpec defines the desired state of PhysicsJob.
type PhysicsJobSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of PhysicsJob. Edit physicsjob_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// PhysicsJobStatus defines the observed state of PhysicsJob.
type PhysicsJobStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PhysicsJob is the Schema for the physicsjobs API.
type PhysicsJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PhysicsJobSpec   `json:"spec,omitempty"`
	Status PhysicsJobStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PhysicsJobList contains a list of PhysicsJob.
type PhysicsJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PhysicsJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PhysicsJob{}, &PhysicsJobList{})
}
