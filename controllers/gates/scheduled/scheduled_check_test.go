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

	deployerv1 "github.com/gitops-tools/kustomization-auto-deployer/api/v1alpha1"
	"github.com/gitops-tools/kustomization-auto-deployer/controllers/gates"
	"github.com/gitops-tools/kustomization-auto-deployer/test"
)

var _ gates.Gate = (*ScheduledGate)(nil)

func TestScheduledGate_Check(t *testing.T) {
	// 9am on the 14th May 2023
	now := time.Date(2023, time.May, 14, 9, 0, 0, 0, time.UTC)

	testCases := []struct {
		open   string
		closed string
		want   bool
	}{
		{
			open:   "07:00",
			closed: "17:00",
			want:   true,
		},
		{
			open:   "10:00",
			closed: "17:00",
			want:   false,
		},
	}

	for _, tt := range testCases {
		t.Run(fmt.Sprintf("open %s, close %s", tt.open, tt.closed), func(t *testing.T) {
			gen := New(logr.Discard())
			gen.Clock = func() time.Time {
				return now
			}

			got, err := gen.Check(context.TODO(), &deployerv1.KustomizationGate{
				Name: "testing",
				Scheduled: &deployerv1.ScheduledCheck{
					Open:  tt.open,
					Close: tt.closed,
				},
			}, nil)

			test.AssertNoError(t, err)
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScheduledGate_Check_errors(t *testing.T) {
	testCases := []struct {
		name    string
		open    string
		closed  string
		wantErr string
	}{
		{
			name:    "bad open time",
			open:    "25:00",
			closed:  "17:00",
			wantErr: "testing",
		},
		{
			name:    "bad closed time",
			open:    "10:00",
			closed:  "17:71",
			wantErr: "testing",
		},
		{
			name:    "closed before open time",
			open:    "17:00",
			closed:  "10:00",
			wantErr: "testing",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			gen := Factory(logr.Discard(), nil)
			_, err := gen.Check(context.TODO(), &deployerv1.KustomizationGate{
				Name: "testing",
				Scheduled: &deployerv1.ScheduledCheck{
					Open:  tt.open,
					Close: tt.closed,
				},
			}, nil)

			test.AssertErrorMatch(t, tt.wantErr, err)
		})
	}
}

func TestScheduledGate_Interval(t *testing.T) {
	intervalTests := []struct {
		now  time.Time
		want time.Duration
	}{
		{
			// 09:00 on the 14th May 2023
			now:  time.Date(2023, time.May, 14, 9, 0, 0, 0, time.UTC),
			want: time.Hour * 8,
		},
		{
			// 18:00 on the 14th May 2023
			now:  time.Date(2023, time.May, 14, 18, 0, 0, 0, time.UTC),
			want: time.Hour * 1,
		},
		{
			// 20:00 on the 14th May 2023
			now:  time.Date(2023, time.May, 14, 20, 0, 0, 0, time.UTC),
			want: time.Hour * 21, // 20:00 -> 00:00 + 17:00 = 21h
		},
	}

	gate := &deployerv1.KustomizationGate{
		Name: "testing",
		Scheduled: &deployerv1.ScheduledCheck{
			Open:  "17:00",
			Close: "19:00",
		},
	}

	for _, tt := range intervalTests {
		t.Run(fmt.Sprintf("%v", tt.now), func(t *testing.T) {
			gen := New(logr.Discard(), func(s *ScheduledGate) {
				s.Clock = func() time.Time {
					return tt.now
				}
			})

			i, err := gen.Interval(gate)
			if err != nil {
				t.Fatal(err)
			}

			if i != tt.want {
				t.Fatalf("Interval() got %v, want %v", i, tt.want)
			}
		})
	}
}
