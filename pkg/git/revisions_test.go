package git

import (
	"context"
	"testing"

	"github.com/gitops-tools/kustomization-auto-deployer/test"
	"github.com/go-git/go-git/v5"
	"github.com/google/go-cmp/cmp"
)

func TestListRevisions(t *testing.T) {
	tr := test.NewRepository(t)

	head := tr.Head()
	commit1 := tr.WriteFileAndCommit("namespace1.yaml", []byte("kind: Namespace\nmetadata:\n  name: namespace-1\n"))
	commit2 := tr.WriteFileAndCommit("namespace2.yaml", []byte("kind: Namespace\nmetadata:\n  name: namespace-2\n"))

	revisions, err := ListRevisionsInRepository(context.TODO(), tr.Dir, ListOptions{})
	test.AssertNoError(t, err)

	want := []string{commit2, commit1, head}
	if diff := cmp.Diff(want, revisions); diff != "" {
		t.Fatalf("failed to generate revisions:\n%s", diff)
	}
}

func TestListRevisions_bad_repo(t *testing.T) {
	dir := t.TempDir()

	_, err := ListRevisionsInRepository(context.TODO(), dir, ListOptions{})

	test.AssertErrorMatch(t, "failed to clone repository URL", err)
}

func TestListRevisions_empty_repo(t *testing.T) {
	dir := t.TempDir()
	_, err := git.PlainInit(dir, false)
	test.AssertNoError(t, err)

	_, err = ListRevisionsInRepository(context.TODO(), dir, ListOptions{})

	test.AssertErrorMatch(t, "failed to clone repository URL", err)
}
