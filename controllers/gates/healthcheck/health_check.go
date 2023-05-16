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

package healthcheck

import (
	"context"
	"net/http"

	deployerv1 "github.com/gitops-tools/kustomization-auto-deployer/api/v1alpha1"
	"github.com/gitops-tools/kustomization-auto-deployer/controllers/gates"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GateFactory is a function for creating per-reconciliation gates for
// the HealthCheckGate.
func GateFactory(l logr.Logger, _ client.Client) gates.Gate {
	return NewGate(l)
}

// NewGate creates and returns a new HealthCheck gate.
func NewGate(l logr.Logger, httpClient *http.Client) *HealthCheckGate {
	return &HealthCheckGateGate{
		Logger:     l,
		HTTPClient: httpClient,
	}
}

// HealthCheckGate checks an HTTP endpoint and is open if the gate returns a 200
// response.
//
// Any other response, is closed.
type HealthCheckGate struct {
	Logger     logr.Logger
	HTTPClient *http.Client
}

// Check returns true if the URL returns a 200 response.
func (g HealthCheckGate) Check(context.Context, deployerv1.KustomizationGate, deployerv1.GatedKustomizationDeployer) (bool, error) {
	return false, nil
}
