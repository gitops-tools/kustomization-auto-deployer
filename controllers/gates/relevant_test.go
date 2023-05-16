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

package gates

import (
	"errors"
	"reflect"
	"testing"

	deployerv1 "github.com/gitops-tools/kustomization-auto-deployer/api/v1alpha1"
	"github.com/gitops-tools/kustomization-auto-deployer/controllers/gates/healthcheck"
	"github.com/gitops-tools/kustomization-auto-deployer/controllers/gates/scheduled"
	"github.com/gitops-tools/kustomization-auto-deployer/test"
)

func TestFindRelevantGates(t *testing.T) {
	allGates := map[string]Gate{
		"HealthCheck": &healthcheck.HealthCheckGate{},
		"Scheduled":   &scheduled.ScheduledGate{},
	}

	tests := []struct {
		name string
		gate deployerv1.KustomizationGate
		want []Gate
	}{
		{
			name: "no gates",
			gate: deployerv1.KustomizationGate{},
			want: []Gate{},
		},
		{
			name: "one gate",
			gate: deployerv1.KustomizationGate{
				HealthCheck: &deployerv1.HealthCheck{},
			},
			want: []Gate{
				&healthcheck.HealthCheckGate{},
			},
		},
		{
			name: "two gates",
			gate: deployerv1.KustomizationGate{
				HealthCheck: &deployerv1.HealthCheck{},
				Scheduled:   &deployerv1.ScheduledCheck{},
			},
			want: []Gate{
				&healthcheck.HealthCheckGate{},
				&scheduled.ScheduledGate{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := FindRelevantGates(tt.gate, allGates); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindRelevantGates() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindReleveantGatesErrors(t *testing.T) {
	tests := []struct {
		name     string
		allGates map[string]Gate
		gate     deployerv1.KustomizationGate
		err      string
	}{
		{
			name: "unknown gate",
			allGates: map[string]Gate{
				"HealthCheck": &healthcheck.HealthCheckGate{},
			},
			gate: deployerv1.KustomizationGate{
				HealthCheck: &deployerv1.HealthCheck{},
				Scheduled:   &deployerv1.ScheduledCheck{},
			},
			err: "Scheduled not enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := FindRelevantGates(tt.gate, tt.allGates)
			test.AssertErrorMatch(t, tt.err, err)
			if !errors.As(err, &GateNotEnabledError{}) {
				t.Errorf("FindRelevantGates() error should be a GateNotEnabledError")
			}
			if !errors.Is(err, GateNotEnabledError{Name: "Scheduled"}) {
				t.Errorf(`FindRelevantGates() error should be GateNotEnabledError{Name: "Scheduled"}`)
			}
		})
	}

}
