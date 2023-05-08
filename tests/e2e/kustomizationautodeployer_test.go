package tests

import (
	"context"
	"regexp"
	"testing"

	"github.com/gitops-tools/kustomization-auto-deployer/test"
	"github.com/onsi/gomega"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	deployerv1 "github.com/gitops-tools/kustomization-auto-deployer/api/v1alpha1"
)

var kustomizationGVK = schema.GroupVersionKind{
	Group:   "kustomize.toolkit.fluxcd.io",
	Kind:    "Kustomization",
	Version: "v1beta2",
}

func TestReconciling(t *testing.T) {
	ctx := context.TODO()

	// both use the same commit IDs test.CommitIDs[0]
	repo := test.NewGitRepository()
	test.AssertNoError(t, testEnv.Create(ctx, repo))
	defer cleanupResource(t, testEnv, repo)

	test.UpdateRepoStatus(t, testEnv, repo, func(r *sourcev1.GitRepository) {
		r.Status.Artifact = &sourcev1.Artifact{
			Revision: "main@sha1:" + test.CommitIDs[0],
		}
	})

	kustomization := test.NewKustomization(repo)
	test.AssertNoError(t, testEnv.Create(ctx, kustomization))
	defer cleanupResource(t, testEnv, kustomization)
	kustomization.Status.LastAppliedRevision = "main@sha1:" + test.CommitIDs[0]
	test.AssertNoError(t, testEnv.Status().Update(ctx, kustomization))

	kd := test.NewKustomizationAutoDeployer()
	test.AssertNoError(t, testEnv.Create(ctx, kd))
	defer cleanupResource(t, testEnv, kd)

	// // the latest commit be the HEAD
	// wantCommit := "main@sha1:" + test.CommitIDs[0]
	// if updated.Status.LatestCommit != wantCommit {
	// 	t.Errorf("failed to update with latest commit, got %q, want %q", updated.Status.LatestCommit, wantCommit)
	// }

	// // the latest commit be the HEAD
	// updatedRepo := &sourcev1.GitRepository{}
	// test.AssertNoError(t, testEnv.Get(ctx, client.ObjectKeyFromObject(repo), updatedRepo))
	// if updatedRepo.Spec.Reference.Commit != test.CommitIDs[0] {
	// 	t.Errorf("failed to configure the GitRepository with the correct commit got %q, want %q", updatedRepo.Spec.Reference.Commit, test.CommitIDs[0])
	// }
}

func cleanupResource(t *testing.T, cl client.Client, obj client.Object) {
	t.Helper()
	if err := cl.Delete(context.TODO(), obj); err != nil {
		t.Fatal(err)
	}
}

func waitForDeployerCondition(t *testing.T, k8sClient client.Client, deployer *deployerv1.KustomizationAutoDeployer, message string) {
	t.Helper()
	g := gomega.NewWithT(t)
	g.Eventually(func() bool {
		updated := &deployerv1.KustomizationAutoDeployer{}
		if err := k8sClient.Get(context.TODO(), client.ObjectKeyFromObject(deployer), updated); err != nil {
			return false
		}
		cond := apimeta.FindStatusCondition(updated.Status.Conditions, meta.ReadyCondition)
		if cond == nil {
			return false
		}

		match, err := regexp.MatchString(message, cond.Message)
		if err != nil {
			t.Fatal(err)
		}

		if !match {
			t.Logf("failed to match %q to %q", message, cond.Message)
		}
		return match
	}, timeout).Should(gomega.BeTrue())
}
