package controllers

import "testing"

func Test_parseRevision(t *testing.T) {
	revisionTests := []struct {
		revision     string
		wantBranch   string
		wantRevision string
	}{
		{
			"main@sha1:40d6b21b888db0ca794876cf7bdd399e3da2137e",
			"main",
			"40d6b21b888db0ca794876cf7bdd399e3da2137e",
		},
		{
			"sha1:a128eb807dbe4556d444cb426601f4cb8309585b",
			"",
			"a128eb807dbe4556d444cb426601f4cb8309585b",
		},
	}

	for _, tt := range revisionTests {
		t.Run(tt.revision, func(t *testing.T) {
			branch, revision := parseRevision(tt.revision)

			if branch != tt.wantBranch {
				t.Errorf("got branch %v, want %v", branch, tt.wantBranch)
			}

			if revision != tt.wantRevision {
				t.Errorf("got revision %v, want %v", revision, tt.wantRevision)
			}
		})
	}
}
