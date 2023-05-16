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

// HealthCheck is a Gate that fetches a URL and is open if the requests are
// successful.
type HealthCheck struct {
	// URL is a  generic catch-all, query the configured URL and if returns
	// anything other than a 200 response, the check fails.
	URL string `json:"url,omitempty"`
}

// ScheduledCheck is a Gate that is open if the current time is between the open
// and close times.
type ScheduledCheck struct {
	// TODO: These need validation!
	// hh:mm for the time to "open" the gate at.
	Open string `json:"open"`
	// hh:mm for the time to "close" the gate at.
	Close string `json:"close"`
}

// KustomizationGate describes a gate to be checked before updating to the
// latest commit.
type KustomizationGate struct {
	// Name is a string used to identify the gate.
	Name string `json:"name"`

	// HealthCheck is a generic URL checker.
	HealthCheck *HealthCheck `json:"healthCheck"`

	// ScheduledCheck is a time-based gate.
	Scheduled *ScheduledCheck `json:"scheduled"`
}

// KustomizationAutoDeployerSpec defines the desired state of KustomizationAutoDeployer
type KustomizationAutoDeployerSpec struct {
	// The Kustomization resource to track and wait for new commits to be
	// available.
	//
	// This will access the GitRepository that is used by the Kustomization.
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
	// +kubebuilder:default=20
	// +kubebuilder:validation:Minimum:=5
	// +kubebuilder:validation:Maximum:=100
	CommitLimit int `json:"commitLimit,omitempty"`

	// Gates are the checks applied before advancing the commit in the
	// GitRepository for the referenced Kustomization.
	Gates []KustomizationGate `json:"gates,omitempty"`
}

// KustomizationAutoDeployerStatus defines the observed state of KustomizationAutoDeployer
type KustomizationAutoDeployerStatus struct {
	// LatestCommit is the latest commit processed by the Kustomization.
	LatestCommit string `json:"latestCommit"`

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
