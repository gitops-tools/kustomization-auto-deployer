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

// KustomizationAutoDeployerSpec defines the desired state of KustomizationAutoDeployer
type KustomizationAutoDeployerSpec struct {
	// The Kustomization resource to track and wait for new commits to be
	// available.
	KustomizationRef meta.LocalObjectReference `json:"kustomizationRef"`
}

// KustomizationAutoDeployerStatus defines the observed state of KustomizationAutoDeployer
type KustomizationAutoDeployerStatus struct {
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
