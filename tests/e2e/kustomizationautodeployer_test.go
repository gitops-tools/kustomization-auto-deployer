package tests

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fluxcd/pkg/apis/meta"
	"github.com/gitops-tools/kustomization-auto-deployer/test"
	"github.com/google/go-cmp/cmp"
	"github.com/onsi/gomega"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

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
	repo := test.NewGitRepository()
	test.AssertNoError(t, testEnv.Create(ctx, repo))
	defer cleanupResource(t, testEnv, repo)

	test.UpdateRepoStatus(t, testEnv, repo, func(r *sourcev1.GitRepository) {
		r.Status.Artifact = &meta.Artifact{
			Revision: "main@sha1:" + test.CommitIDs[2],
		}
	})

	kustomization := test.NewKustomization(repo)
	test.AssertNoError(t, testEnv.Create(ctx, kustomization))
	defer cleanupResource(t, testEnv, kustomization)
	kustomization.Status.LastAppliedRevision = "main@sha1:" + test.CommitIDs[2]
	test.AssertNoError(t, testEnv.Status().Update(ctx, kustomization))

	kd := test.NewKustomizationAutoDeployer()
	test.AssertNoError(t, testEnv.Create(ctx, kd))
	defer cleanupResource(t, testEnv, kd)

	// The GitRepository is updated to reflect the next commit
	waitForGitRepository(t, testEnv, client.ObjectKeyFromObject(repo), &sourcev1.GitRepositoryRef{
		Branch: "main",
		Commit: test.CommitIDs[1],
	})

	waitForDeployerCheck(t, testEnv, client.ObjectKeyFromObject(kd), func(depl deployerv1.KustomizationAutoDeployer) bool {
		return depl.Status.LatestCommit == fmt.Sprintf("%s@sha1:%s", repo.Spec.Reference.Branch, test.CommitIDs[1])
	})

	test.UpdateRepoStatus(t, testEnv, repo, func(r *sourcev1.GitRepository) {
		r.Status.Artifact = &meta.Artifact{
			Revision: "main@sha1:" + test.CommitIDs[1],
		}
	})

	// Update the Kustomization to reflect applying the new version from the
	// GitRepository.
	kustomization.Status.LastAppliedRevision = "main@sha1:" + test.CommitIDs[1]
	test.AssertNoError(t, testEnv.Status().Update(ctx, kustomization))

	// The GitRepository is updated to reflect the next commit
	waitForGitRepository(t, testEnv, client.ObjectKeyFromObject(repo), &sourcev1.GitRepositoryRef{
		Branch: "main",
		Commit: test.CommitIDs[0],
	})

	test.UpdateRepoStatus(t, testEnv, repo, func(r *sourcev1.GitRepository) {
		r.Status.Artifact = &meta.Artifact{
			Revision: "main@sha1:" + test.CommitIDs[0],
		}
	})

	waitForDeployerCheck(t, testEnv, client.ObjectKeyFromObject(kd), func(depl deployerv1.KustomizationAutoDeployer) bool {
		return depl.Status.LatestCommit == fmt.Sprintf("%s@sha1:%s", repo.Spec.Reference.Branch, test.CommitIDs[0])
	})
}

func TestReconciling_with_gates(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "test error", http.StatusNotFound)
	}))
	t.Cleanup(ts.Close)

	ctx := context.TODO()
	repo := test.NewGitRepository()
	test.AssertNoError(t, testEnv.Create(ctx, repo))
	defer cleanupResource(t, testEnv, repo)

	kustomization := test.NewKustomization(repo)
	test.AssertNoError(t, testEnv.Create(ctx, kustomization))
	defer cleanupResource(t, testEnv, kustomization)
	kustomization.Status.LastAppliedRevision = "main@sha1:" + test.CommitIDs[2]
	test.AssertNoError(t, testEnv.Status().Update(ctx, kustomization))

	test.UpdateRepoStatus(t, testEnv, repo, func(r *sourcev1.GitRepository) {
		r.Status.Artifact = &meta.Artifact{
			Revision: "main@sha1:" + test.CommitIDs[2],
		}
	})

	kd := test.NewKustomizationAutoDeployer(func(kd *deployerv1.KustomizationAutoDeployer) {
		kd.Spec.Gates = []deployerv1.KustomizationGate{
			{
				Name: "accessing a test server",
				HealthCheck: &deployerv1.HealthCheck{
					URL: ts.URL,
				},
			},
		}
	})
	test.AssertNoError(t, testEnv.Create(ctx, kd))
	defer cleanupResource(t, testEnv, kd)

	waitForDeployerCheck(t, testEnv, client.ObjectKeyFromObject(kd), func(depl deployerv1.KustomizationAutoDeployer) bool {
		cond := apimeta.FindStatusCondition(depl.Status.Conditions, meta.ReadyCondition)
		if cond == nil {
			return false
		}

		return cond.Message == "gates are currently closed"
	})
}

func cleanupResource(t *testing.T, cl client.Client, obj client.Object) {
	t.Helper()
	if err := cl.Delete(context.TODO(), obj); err != nil {
		t.Fatal(err)
	}
}

func waitForGitRepository(t *testing.T, k8sClient client.Client, name client.ObjectKey, wantRef *sourcev1.GitRepositoryRef) {
	t.Helper()
	g := gomega.NewWithT(t)
	g.Eventually(func() bool {
		repo := &sourcev1.GitRepository{}
		if err := k8sClient.Get(context.TODO(), name, repo); err != nil {
			return false
		}
		if diff := cmp.Diff(repo.Spec.Reference, wantRef); diff != "" {
			return false
		}
		return true
	}, timeout).Should(gomega.BeTrue())
}

func waitForDeployerCheck(t *testing.T, k8sClient client.Client, name client.ObjectKey, check func(deployerv1.KustomizationAutoDeployer) bool) {
	t.Helper()
	g := gomega.NewWithT(t)
	g.Eventually(func() bool {
		var deployer deployerv1.KustomizationAutoDeployer
		if err := k8sClient.Get(context.TODO(), name, &deployer); err != nil {
			return false
		}

		return check(deployer)
	}, timeout).Should(gomega.BeTrue())
}
