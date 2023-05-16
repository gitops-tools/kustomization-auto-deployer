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
	"testing"

	deployerv1 "github.com/gitops-tools/kustomization-auto-deployer/api/v1alpha1"
)

func TestCheck(t *testing.T) {
	checkTests := []struct {
		name     string
		deployer *deployerv1.KustomizationAutoDeployer
	}{
		{
			"no gates should fail open",
		},
	}

	for _, tt := range checkTests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}
