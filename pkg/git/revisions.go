package git

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

const defaultDepth = 20

// ListOptions configures how the commits are fetched and processed.
type ListOptions struct {
	MaxCommits int
}

// ListRevisionsInRepository lists the revisions in the repository.
//
// It will clone the repository to a configurable depth and list all revisions
// in the clone.
func ListRevisionsInRepository(ctx context.Context, url string, options ListOptions) (result []string, listErr error) {
	dir, err := ioutil.TempDir(os.TempDir(), "tracker")
	if err != nil {
		return nil, fmt.Errorf("failed to create tempdir opening repository %s: %w", url, err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			listErr = err
		}
	}()

	maxCommits := options.MaxCommits
	if maxCommits == 0 {
		maxCommits = defaultDepth
	}

	r, err := git.PlainCloneContext(ctx, dir, false, &git.CloneOptions{Depth: maxCommits, NoCheckout: true, URL: url})
	if err != nil {
		return nil, fmt.Errorf("failed to clone repository URL %s: %w", url, err)
	}

	ref, err := r.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get repo HEAD for %s: %w", url, err)
	}

	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get commit for HEAD %s: %w", url, err)
	}

	commitIter, err := r.Log(&git.LogOptions{From: commit.Hash})
	if err != nil {
		return nil, fmt.Errorf("failed to list commits for HEAD %s: %w", url, err)
	}

	commits := []string{}
	err = commitIter.ForEach(func(c *object.Commit) error {
		commits = append(commits, c.Hash.String())

		return nil
	})

	return commits, nil
}
