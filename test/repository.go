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

package test

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"sigs.k8s.io/yaml"
)

// TestRepository is a self-contained Git repository used in testing.
type TestRepository struct {
	t          *testing.T
	Base       string
	Dir        string
	Repository *git.Repository
}

// NewRepository creates and initialises a new git repository in a temporary
// directory, with an initial commit with a README file.
func NewRepository(t *testing.T) *TestRepository {
	t.Helper()
	base := t.TempDir()
	dir := filepath.Join(base, "repo")
	r, err := git.PlainInit(dir, false)
	AssertNoError(t, err)
	writeInitialCommit(t, r)

	return &TestRepository{t: t, Dir: dir, Base: base, Repository: r}
}

// Heads returns the HEAD commit ID.
func (n *TestRepository) Head() string {
	ref, err := n.Repository.Head()
	if err != nil {
		n.t.Fatalf("failed to get repo HEAD %s", err)
	}

	commit, err := n.Repository.CommitObject(ref.Hash())
	if err != nil {
		n.t.Fatalf("failed to get commit for HEAD %s", err)
	}

	return commit.Hash.String()
}

// WriteFileAndCommit writes a file to the filesystem in a subdirectory of the
// temporary directory, it also commits and returns the commit ID.
func (n *TestRepository) WriteFileAndCommit(filename string, body []byte) string {
	n.t.Helper()
	wt, err := n.Repository.Worktree()
	if err != nil {
		n.t.Fatalf("failed to create a worktree %s", err)
	}
	writeTestFile(n.t, wt.Filesystem, filename, body)
	if _, err := wt.Add(filename); err != nil {
		n.t.Fatalf("failed to add file %s: %s", filename, err)
	}

	c, err := wt.Commit(fmt.Sprintf("Test commit %s", time.Now()), &git.CommitOptions{})
	if err != nil {
		n.t.Fatalf("failed to commit: %s", err)
	}

	return c.String()
}

// WriteProfileAndCommit serialises the provided value to YAML and writes it to the
// file.
func (n *TestRepository) WriteValueAndTag(filename string, v any) string {
	n.t.Helper()
	b, err := yaml.Marshal(v)
	if err != nil {
		n.t.Fatalf("failed to marshal: %s", err)
	}
	return n.WriteFileAndCommit(filename, b)
}

func writeInitialCommit(t *testing.T, r *git.Repository) {
	t.Helper()
	wt, err := r.Worktree()
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, wt.Filesystem, "README.md", []byte("Test Project"))
	if _, err := wt.Add("README.md"); err != nil {
		t.Fatal(err)
	}
	_, err = wt.Commit("Initial Commit", &git.CommitOptions{})
	if err != nil {
		t.Fatal(err)
	}
}

func writeTestFile(t *testing.T, fs billy.Filesystem, name string, body []byte) {
	t.Helper()
	newFile, err := fs.Create(name)
	defer newFile.Close()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := newFile.Write(body); err != nil {
		t.Fatal(err)
	}
}
