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
	"fmt"
	"strings"

	deployerv1 "github.com/gitops-tools/kustomization-auto-deployer/api/v1alpha1"
)

// Check checks the gates defined in the KustomizationAutoDeployer and returns
// true if all gates are open.
func Check(ctx context.Context, r *deployerv1.KustomizationAutoDeployer, configuredGates map[string]Gate) (bool, deployerv1.GatesStatus, error) {
	// Open if no Gates are defined.
	if len(r.Spec.Gates) == 0 {
		return true, nil, nil
	}

	result := deployerv1.GatesStatus{}
	for _, gate := range r.Spec.Gates {
		checks, err := check(ctx, gate, r, configuredGates)
		if err != nil {
			return false, nil, err
		}

		result[gate.Name] = checks
	}

	return summarise(result), result, nil
}

func summarise(res map[string]map[string]bool) bool {
	for _, gate := range res {
		for _, check := range gate {
			if check == false {
				return false
			}
		}
	}

	return true
}

func check(ctx context.Context, gate deployerv1.KustomizationGate, deployer *deployerv1.KustomizationAutoDeployer, configuredGates map[string]Gate) (map[string]bool, error) {
	gates, err := FindRelevantGates(gate, configuredGates)
	if err != nil {
		return nil, err
	}

	result := map[string]bool{}
	for _, g := range gates {
		open, err := g.Check(ctx, &gate, deployer)
		if err != nil {
			return nil, err
		}
		result[gateName(g)] = open
	}

	return result, nil
}

func gateName(gate Gate) string {
	name := fmt.Sprintf("%T", gate)
	elements := strings.Split(name, ".")
	return elements[1]
}
