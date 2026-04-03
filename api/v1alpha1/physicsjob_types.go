package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SchedulingPolicy controls how the operator selects the target node.
// +kubebuilder:validation:Enum=DataLocal;ClosestSite;AnyAvailable
type SchedulingPolicy string

const (
	// DataLocal schedules compute on the node holding the primary dataset replica.
	DataLocal SchedulingPolicy = "DataLocal"
	// ClosestSite schedules on the geographically nearest site with a replica.
	ClosestSite SchedulingPolicy = "ClosestSite"
	// AnyAvailable schedules on any node with available resources, ignoring locality.
	AnyAvailable SchedulingPolicy = "AnyAvailable"
)

// Phase represents the lifecycle state of a PhysicsJob.
// +kubebuilder:validation:Enum=Pending;Resolving;Scheduled;Running;Succeeded;Failed
type Phase string

const (
	PhasePending   Phase = "Pending"
	PhaseResolving Phase = "Resolving"
	PhaseScheduled Phase = "Scheduled"
	PhaseRunning   Phase = "Running"
	PhaseSucceeded Phase = "Succeeded"
	PhaseFailed    Phase = "Failed"
)

// PhysicsJobSpec defines the desired state of PhysicsJob.
type PhysicsJobSpec struct {
	// Dataset is the Rucio DID (scope:name) identifying the input dataset.
	// Example: "data23_13p6TeV:DAOD_PHYS.123456"
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9_.\-]+:[a-zA-Z0-9_.\-/]+$`
	Dataset string `json:"dataset"`

	// Image is the container image to run for the compute job.
	// +kubebuilder:validation:Required
	Image string `json:"image"`

	// Command overrides the container entrypoint.
	// +optional
	Command []string `json:"command,omitempty"`

	// Resources specifies the compute resource requirements for the job container.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// SchedulingPolicy controls how the operator selects the target node.
	// Defaults to DataLocal.
	// +kubebuilder:default=DataLocal
	// +optional
	SchedulingPolicy SchedulingPolicy `json:"schedulingPolicy,omitempty"`
}

// PhysicsJobStatus defines the observed state of PhysicsJob.
type PhysicsJobStatus struct {
	// Phase is the current lifecycle phase of the PhysicsJob.
	// +optional
	Phase Phase `json:"phase,omitempty"`

	// ResolvedRSE is the Rucio Storage Element selected as the data source.
	// +optional
	ResolvedRSE string `json:"resolvedRSE,omitempty"`

	// ScheduledNode is the Kubernetes node where the compute Job landed.
	// +optional
	ScheduledNode string `json:"scheduledNode,omitempty"`

	// JobRef is the name of the owned batch/v1.Job created for this PhysicsJob.
	// +optional
	JobRef string `json:"jobRef,omitempty"`

	// BytesTransferAvoided is the estimated bytes not moved over the WAN
	// due to data-local scheduling.
	// +optional
	BytesTransferAvoided int64 `json:"bytesTransferAvoided,omitempty"`

	// Conditions holds standard metav1.Condition entries describing the current state.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=pj
// +kubebuilder:printcolumn:name="Dataset",type="string",JSONPath=".spec.dataset",priority=0
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",priority=0
// +kubebuilder:printcolumn:name="RSE",type="string",JSONPath=".status.resolvedRSE",priority=0
// +kubebuilder:printcolumn:name="Node",type="string",JSONPath=".status.scheduledNode",priority=1
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",priority=0

// PhysicsJob is the Schema for the physicsjobs API.
// It represents a data-locality-aware compute workload that is scheduled
// on the Kubernetes node co-located with the specified Rucio dataset.
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
