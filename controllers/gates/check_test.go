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

package gates_test

import (
	"context"
	"testing"
	"time"

	deployerv1 "github.com/gitops-tools/kustomization-auto-deployer/api/v1alpha1"
	"github.com/gitops-tools/kustomization-auto-deployer/controllers/gates"
	"github.com/gitops-tools/kustomization-auto-deployer/controllers/gates/scheduled"
	"github.com/gitops-tools/kustomization-auto-deployer/test"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
)

func TestCheck(t *testing.T) {
	// 9am on the 14th May 2023
	now := time.Date(2023, time.May, 14, 9, 0, 0, 0, time.UTC)

	checkTests := []struct {
		name     string
		deployer *deployerv1.KustomizationAutoDeployer
		open     bool
		checks   deployerv1.GatesStatus
	}{
		{
			name:     "no gates should fail open",
			deployer: test.NewKustomizationAutoDeployer(),
			open:     true,
		},
		{
			name: "open gate is open",
			deployer: test.NewKustomizationAutoDeployer(func(d *deployerv1.KustomizationAutoDeployer) {
				d.Spec.Gates = []deployerv1.KustomizationGate{
					{
						Name: "within scheduled hours",
						Scheduled: &deployerv1.ScheduledCheck{
							Open:  "08:00",
							Close: "16:00",
						},
					},
				}
			}),
			open:   true,
			checks: deployerv1.GatesStatus{"within scheduled hours": {"ScheduledGate": true}},
		},
		{
			name: "closed gate is closed",
			deployer: test.NewKustomizationAutoDeployer(func(d *deployerv1.KustomizationAutoDeployer) {
				d.Spec.Gates = []deployerv1.KustomizationGate{
					{
						Name: "outwith scheduled hours",
						Scheduled: &deployerv1.ScheduledCheck{
							Open:  "10:00",
							Close: "18:00",
						},
					},
				}
			}),
			open:   false,
			checks: deployerv1.GatesStatus{"outwith scheduled hours": {"ScheduledGate": false}},
		},
	}

	gateValues := map[string]gates.Gate{
		"Scheduled": scheduled.New(logr.Discard(), func(s *scheduled.ScheduledGate) {
			s.Clock = func() time.Time { return now }
		}),
	}

	for _, tt := range checkTests {
		t.Run(tt.name, func(t *testing.T) {
			open, checks, err := gates.Check(context.TODO(), tt.deployer, gateValues)
			if err != nil {
				t.Fatal(err)
			}

			if open != tt.open {
				t.Errorf("got open %v, want %v", open, tt.open)
			}

			if diff := cmp.Diff(tt.checks, checks); diff != "" {
				t.Fatalf("failed to calculate checks:\n%s", diff)
			}
		})
	}
}
