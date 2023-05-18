/*
Copyright 2023.

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
	"github.com/fluxcd/pkg/apis/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// GatesClosedReason is set when further deployments can't continue because
	// the gates are currently closed.
	GatesClosedReason string = "GatesClosed"

	// RevisionsErrorReason is set when we couldn't list the revisions in the
	// upstream repository.
	RevisionsErrorReason string = "RevisionsError"
)

// HealthCheck is a Gate that fetches a URL and is open if the requests are
// successful.
type HealthCheck struct {
	// URL is a  generic catch-all, query the configured URL and if returns
	// anything other than a 200 response, the check fails.
	// +required
	URL string `json:"url"`

	// Interval at which to check the URL for updates.
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ms|s|m|h))+$"
	// +required
	Interval metav1.Duration `json:"interval"`
}

// ScheduledCheck is a Gate that is open if the current time is between the open
// and close times.
type ScheduledCheck struct {
	// TODO: These need validation!
	// hh:mm for the time to "open" the gate at.
	// +required
	Open string `json:"open"`
	// hh:mm for the time to "close" the gate at.
	// +required
	Close string `json:"close"`
}

// KustomizationGate describes a gate to be checked before updating to the
// latest commit.
type KustomizationGate struct {
	// Name is a string used to identify the gate.
	// +required
	Name string `json:"name"`

	// HealthCheck is a generic URL checker.
	// +optional
	HealthCheck *HealthCheck `json:"healthCheck,omitempty"`

	// ScheduledCheck is a time-based gate.
	// +optional
	Scheduled *ScheduledCheck `json:"scheduled,omitempty"`
}

// KustomizationAutoDeployerSpec defines the desired state of KustomizationAutoDeployer
type KustomizationAutoDeployerSpec struct {
	// The Kustomization resource to track and wait for new commits to be
	// available.
	//
	// This will access the GitRepository that is used by the Kustomization.
	// +required
	KustomizationRef meta.LocalObjectReference `json:"kustomizationRef"`

	// Interval at which to check the GitRepository for updates.
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ms|s|m|h))+$"
	// +required
	Interval metav1.Duration `json:"interval"`

	// CloneDepth limits the number of commits to get from the GitRepository.
	//
	// This is an optimisation for fetching commits.
	//
	// +kubebuilder:default=100
	// +kubebuilder:validation:Minimum:=5
	// +kubebuilder:validation:Maximum:=100
	CommitLimit int `json:"commitLimit,omitempty"`

	// Gates are the checks applied before advancing the commit in the
	// GitRepository for the referenced Kustomization.
	// +optional
	Gates []KustomizationGate `json:"gates,omitempty"`
}

// KustomizationAutoDeployerStatus defines the observed state of KustomizationAutoDeployer
type KustomizationAutoDeployerStatus struct {
	// LatestCommit is the latest commit processed by the Kustomization.
	// +optional
	LatestCommit string `json:"latestCommit,omitempty"`

	// ObservedGeneration reflects the generation of the most recently observed
	// KustomizationAutoDeployer.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions holds the conditions for the KustomizationAutoDeployer.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// KustomizationAutoDeployer is the Schema for the kustomizationautodeployers API
type KustomizationAutoDeployer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KustomizationAutoDeployerSpec   `json:"spec,omitempty"`
	Status KustomizationAutoDeployerStatus `json:"status,omitempty"`
}

// SetConditions sets the status conditions on the object.
func (in *KustomizationAutoDeployer) SetConditions(conditions []metav1.Condition) {
	in.Status.Conditions = conditions
}

//+kubebuilder:object:root=true

// KustomizationAutoDeployerList contains a list of KustomizationAutoDeployer
type KustomizationAutoDeployerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KustomizationAutoDeployer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KustomizationAutoDeployer{}, &KustomizationAutoDeployerList{})
}
