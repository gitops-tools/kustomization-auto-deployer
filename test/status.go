package test

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	sourcev1 "github.com/fluxcd/source-controller/api/v1"
)

// UpdateRepoStatus applies changes from an update function to a GitRepository
// and sets the LastUpdateTime at the same time.
func UpdateRepoStatus(t *testing.T, k8sClient client.Client, repo *sourcev1.GitRepository, update func(*sourcev1.GitRepository)) {
	t.Helper()
	ctx := context.TODO()
	AssertNoError(t, k8sClient.Get(ctx, client.ObjectKeyFromObject(repo), repo))

	update(repo)
	repo.Status.Artifact.LastUpdateTime = metav1.Now()

	AssertNoError(t, k8sClient.Status().Update(ctx, repo))
}
