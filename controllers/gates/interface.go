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
	"context"
	"time"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	deployerv1 "github.com/gitops-tools/kustomization-auto-deployer/api/v1alpha1"
)

// GateFactory is a way to create a per-reconciliation gate.
type GateFactory func(logr.Logger, client.Client) Gate

// Gate defines the interface implemented by all gates.
type Gate interface {
	// Check returns true if the Gate is open.
	//
	// Errors are only for exceptional cases.
	Check(context.Context, *deployerv1.KustomizationGate, *deployerv1.KustomizationAutoDeployer) (bool, error)

	// Interval is the time after which a Gate should be checked.
	//
	// A Gate can return an empty time.Duration value if it should not be
	// rechecked after a period.
	Interval(*deployerv1.KustomizationGate) time.Duration
}

// NoRequeueInterval is a simple default value that can be used to indicate that
// a Gate should not requeue after a time duration.
var NoRequeueInterval time.Duration
