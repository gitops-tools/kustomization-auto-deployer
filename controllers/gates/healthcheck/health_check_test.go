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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	deployerv1 "github.com/gitops-tools/kustomization-auto-deployer/api/v1alpha1"
	"github.com/gitops-tools/kustomization-auto-deployer/controllers/gates"
	"github.com/gitops-tools/kustomization-auto-deployer/test"
	"github.com/go-logr/logr"
)

var _ gates.Gate = (*HealthCheckGate)(nil)

func TestCheck(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/closed":
			http.Error(w, "currently closed", http.StatusInternalServerError)
		case "/open":
			fmt.Fprintln(w, "currently open")
		default:
			http.Error(w, "unknown", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	testCases := []struct {
		path string
		want bool
	}{
		{
			"/open", true,
		},
		{
			"/closed", false,
		},
		{
			"/other", false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.path, func(t *testing.T) {
			gen := New(logr.Discard(), ts.Client())

			got, err := gen.Check(context.TODO(), &deployerv1.KustomizationGate{
				Name: "testing",
				HealthCheck: &deployerv1.HealthCheck{
					URL: ts.URL + tt.path,
				},
			}, nil)

			test.AssertNoError(t, err)
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheck_errors(t *testing.T) {
	// testCases := []struct {
	// 	name    string
	// 	open    string
	// 	closed  string
	// 	wantErr string
	// }{
	// 	{
	// 		name:    "bad open time",
	// 		open:    "25:00",
	// 		closed:  "17:00",
	// 		wantErr: "testing",
	// 	},
	// 	{
	// 		name:    "bad closed time",
	// 		open:    "10:00",
	// 		closed:  "17:71",
	// 		wantErr: "testing",
	// 	},
	// 	{
	// 		name:    "closed before open time",
	// 		open:    "17:00",
	// 		closed:  "10:00",
	// 		wantErr: "testing",
	// 	},
	// }

	// for _, tt := range testCases {
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		gen := Factory(logr.Discard(), nil)
	// 		_, err := gen.Check(context.TODO(), &deployerv1.KustomizationGate{
	// 			Name: "testing",
	// 			Scheduled: &deployerv1.ScheduledCheck{
	// 				Open:  tt.open,
	// 				Close: tt.closed,
	// 			},
	// 		}, nil)

	// 		test.AssertErrorMatch(t, tt.wantErr, err)
	// 	})
	// }
}
