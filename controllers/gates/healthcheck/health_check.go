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
	"time"

	deployerv1 "github.com/gitops-tools/kustomization-auto-deployer/api/v1alpha1"
	"github.com/gitops-tools/kustomization-auto-deployer/controllers/gates"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Factory is a function for creating per-reconciliation gates for
// the HealthCheckGate.
func Factory(httpClient *http.Client) gates.GateFactory {
	return func(l logr.Logger, c client.Client) gates.Gate {
		return New(l, httpClient)
	}
}

// New creates and returns a new HealthCheck gate.
func New(l logr.Logger, httpClient *http.Client) *HealthCheckGate {
	return &HealthCheckGate{
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
func (g HealthCheckGate) Check(ctx context.Context, gate *deployerv1.KustomizationGate, _ *deployerv1.KustomizationAutoDeployer) (bool, error) {
	// TODO: logging

	req, err := http.NewRequest(http.MethodGet, gate.HealthCheck.URL, nil)
	if err != nil {
		// TODO: improve this error!
		return false, err
	}

	g.Logger.Info("getting healthcheck", "gate", gate.Name, "url", gate.HealthCheck.URL)

	resp, err := g.HTTPClient.Do(req)
	if err != nil {
		// TODO: improve this error!
		return false, err
	}

	g.Logger.Info("healthcheck complete", "gate", gate.Name, "statusCode", resp.StatusCode)

	return resp.StatusCode == http.StatusOK, nil
}

// Interval returns the time after which to requeue this check.
func (g HealthCheckGate) Interval(gate *deployerv1.KustomizationGate) (time.Duration, error) {
	return gate.HealthCheck.Interval.Duration, nil

}
