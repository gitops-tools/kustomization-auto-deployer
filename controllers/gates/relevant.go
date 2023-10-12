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
	"fmt"
	"reflect"

	deployerv1 "github.com/gitops-tools/kustomization-auto-deployer/api/v1alpha1"
)

// GateNotEnabledError is returned when a gate is not enabled
// in the controller but a gated deploy tries to use it.
// If you want to handle this error you can either use
// errors.As(err, &GateNotEnabledError{}) to check for any gate, or use
// errors.Is(err, GateNotEnabledError{Name: HealthCheckGate}) for a specific gate.
type GateNotEnabledError struct {
	Name string
}

func (g GateNotEnabledError) Error() string {
	return fmt.Sprintf("gate %s not enabled", g.Name)
}

// FindRelevantGates takes a struct with keys of the same type as
// Gates in the map and finds relevant gates.
func FindRelevantGates(setGate deployerv1.KustomizationGate, enabledGates map[string]Gate) ([]Gate, error) {
	res := []Gate{}
	v := reflect.Indirect(reflect.ValueOf(setGate))
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := v.Type().Field(i).Name
		if !field.CanInterface() || fieldName == "Name" {
			continue
		}

		if !reflect.ValueOf(field.Interface()).IsNil() {
			gen, ok := enabledGates[fieldName]
			if !ok {
				return nil, GateNotEnabledError{Name: fieldName}
			}
			res = append(res, gen)
			continue
		}
	}

	return res, nil
}
