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
	"testing"
	"time"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	deployerv1 "github.com/gitops-tools/kustomization-auto-deployer/api/v1alpha1"
	"github.com/gitops-tools/kustomization-auto-deployer/controllers/gates"
	"github.com/gitops-tools/kustomization-auto-deployer/test"
)

var _ gates.Gate = (*ScheduledGate)(nil)

// func TestCheck_with_no_times(t *testing.T) {
// 	gen := GateFactory(logr.Discard(), nil)
// 	got, err := gen.Check(context.TODO(), &deployerv1.KustomizationGate{}, nil)

// 	test.AssertNoError(t, err)
// 	if got != true {
// 		t.Errorf("got %v, want %v with  generator", got, true)
// 	}
// }

func TestCheck(t *testing.T) {
	testCases := []struct {
		name   string
		open   time.Time
		closed time.Time
		want   bool
	}{
		{
			open:   time.Now(),
			closed: time.Now(),
			want:   true,
		},
		// {
		// 	name:     "nested key/values",
		// 	elements: []apiextensionsv1.JSON{{Raw: []byte(`{"cluster": "cluster","url": "url","values":{"foo":"bar"}}`)}},
		// 	want:     []map[string]any{{"cluster": "cluster", "url": "url", "values": map[string]any{"foo": "bar"}}},
		// },
	}

	for _, tt := range testCases {
		t.Run(fmt.Sprintf("open %s, close %s", tt.open.Format("15:04:05"), tt.closed.Format("15:04:05")), func(t *testing.T) {

			gen := GateFactory(logr.Discard(), nil)
			got, err := gen.Check(context.TODO(), &deployerv1.KustomizationGate{
				Scheduled: &deployerv1.ScheduledCheck{
					Open:  metav1.Time{Time: tt.open},
					Close: metav1.Time{Time: tt.closed},
				},
			}, nil)

			test.AssertNoError(t, err)
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// func TestCheck_errors(t *testing.T) {
// 	testCases := []struct {
// 		name      string
// 		generator *templatesv1.GitOpsSetGate
// 		wantErr   string
// 	}{
// 		{
// 			name: "bad json",
// 			generator: &templatesv1.GitOpsSetGate{
// 				List: &templatesv1.ListGate{
// 					Elements: []apiextensionsv1.JSON{{Raw: []byte(`{`)}},
// 				},
// 			},
// 			wantErr: "error unmarshaling list element: unexpected end of JSON input",
// 		},
// 		{
// 			name:      "no generator",
// 			generator: nil,
// 			wantErr:   "GitOpsSet is empty",
// 		},
// 	}

// 	for _, tt := range testCases {
// 		t.Run(tt.name, func(t *testing.T) {

// 			gen := GateFactory(logr.Discard(), nil)
// 			_, err := gen.Check(context.TODO(), tt.generator, nil)

// 			test.AssertErrorMatch(t, tt.wantErr, err)
// 		})
// 	}
// }

// func TestListGate_Interval(t *testing.T) {
// 	gen := NewGate(logr.Discard())
// 	sg := &templatesv1.GitOpsSetGate{
// 		List: &templatesv1.ListGate{
// 			Elements: []apiextensionsv1.JSON{{Raw: []byte(`{"cluster": "cluster","url": "url"}`)}},
// 		},
// 	}

// 	d := gen.Interval(sg)

// 	if d != generators.NoRequeueInterval {
// 		t.Fatalf("got %#v want %#v", d, generators.NoRequeueInterval)
// 	}
// }
