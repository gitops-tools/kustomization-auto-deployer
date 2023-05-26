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

package scheduled

import (
	"context"
	"fmt"
	"time"

	deployerv1 "github.com/gitops-tools/kustomization-auto-deployer/api/v1alpha1"
	"github.com/gitops-tools/kustomization-auto-deployer/controllers/gates"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Factory is a function for creating per-reconciliation gates for
// the ScheduledGate.
func Factory(l logr.Logger, _ client.Client) gates.Gate {
	return New(l)
}

// New creates and returns a new ScheduledGate.
func New(l logr.Logger, opts ...func(*ScheduledGate)) *ScheduledGate {
	sg := &ScheduledGate{
		Logger: l,
		Clock:  time.Now,
	}

	for _, opt := range opts {
		opt(sg)
	}

	return sg
}

// ScheduledGate is open based on the current time.
type ScheduledGate struct {
	Logger logr.Logger
	Clock  func() time.Time
}

// TODO: if closed is earlier than open, we could assume an "overnight" type
// scenario.

// Check returns true if now is within the the Scheduled gate time duration.
func (g ScheduledGate) Check(ctx context.Context, gate *deployerv1.KustomizationGate, _ *deployerv1.KustomizationAutoDeployer) (bool, error) {
	// TODO: Logging
	now := g.Clock()
	open, closed, err := parseScheduledTimes(now, gate.Name, gate.Scheduled)
	if err != nil {
		return false, err
	}

	if closed.Before(open) {
		return false, fmt.Errorf("parsing Scheduled %s %v is before %v", gate.Name, gate.Scheduled.Close, gate.Scheduled.Open)
	}

	return now.After(open) && now.Before(closed), nil
}

func parseAndMerge(now time.Time, name, phase, str string) (time.Time, error) {
	parsed, err := time.Parse("15:04", str)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse %s time for %s %q: %w", phase, name, str, err)
	}

	return time.Date(now.Year(), now.Month(), now.Day(), parsed.Hour(), parsed.Minute(), 0, 0, now.Location()), nil
}

// Interval returns the time after which to requeue this check.
func (g ScheduledGate) Interval(gate *deployerv1.KustomizationGate) (time.Duration, error) {
	now := g.Clock()
	open, closed, err := parseScheduledTimes(now, gate.Name, gate.Scheduled)
	if err != nil {
		return 0, err
	}

	if now.Before(open) {
		return open.Sub(now), nil
	}

	if now.Before(closed) {
		return closed.Sub(now), nil
	}

	return open.Add(time.Hour * 24).Sub(now), nil
}

func parseScheduledTimes(now time.Time, name string, check *deployerv1.ScheduledCheck) (open time.Time, closed time.Time, err error) {
	open, err = parseAndMerge(now, name, "open", check.Open)
	if err != nil {
		return
	}

	closed, err = parseAndMerge(now, name, "close", check.Close)
	if err != nil {
		return
	}

	return
}
