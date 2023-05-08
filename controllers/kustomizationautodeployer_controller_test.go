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

package controllers

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	"github.com/go-logr/logr/testr"
	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log"

	deployerv1 "github.com/gitops-tools/kustomization-auto-deployer/api/v1alpha1"
	"github.com/gitops-tools/kustomization-auto-deployer/pkg/git"
	"github.com/gitops-tools/kustomization-auto-deployer/test"
)

func TestReconciliation(t *testing.T) {
	testEnv := &envtest.Environment{
		ErrorIfCRDPathMissing: true,
		CRDDirectoryPaths: []string{
			filepath.Join("..", "config", "crd", "bases"),
			"testdata/crds",
		},
	}

	cfg, err := testEnv.Start()
	test.AssertNoError(t, err)
	defer func() {
		if err := testEnv.Stop(); err != nil {
			t.Errorf("failed to stop the test environment: %s", err)
		}
	}()

	scheme := runtime.NewScheme()
	test.AssertNoError(t, clientgoscheme.AddToScheme(scheme))
	test.AssertNoError(t, deployerv1.AddToScheme(scheme))
	test.AssertNoError(t, kustomizev1.AddToScheme(scheme))
	test.AssertNoError(t, sourcev1.AddToScheme(scheme))

	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme})
	test.AssertNoError(t, err)

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{Scheme: scheme})
	test.AssertNoError(t, err)

	reconciler := &KustomizationAutoDeployerReconciler{
		Client:         k8sClient,
		Scheme:         scheme,
		RevisionLister: testRevisionLister(test.CommitIDs),
	}

	test.AssertNoError(t, reconciler.SetupWithManager(mgr))

	t.Run("reconciling GitRepository with missing Kustomization", func(t *testing.T) {
		ctx := log.IntoContext(context.TODO(), testr.New(t))
		deployer := test.NewKustomizationAutoDeployer()
		test.AssertNoError(t, k8sClient.Create(ctx, deployer))
		defer cleanupResource(t, k8sClient, deployer)

		repo := test.NewGitRepository()
		test.AssertNoError(t, k8sClient.Create(ctx, repo))
		defer cleanupResource(t, k8sClient, repo)

		test.UpdateRepoStatus(t, k8sClient, repo, func(r *sourcev1.GitRepository) {
			repo.Status.Artifact = &sourcev1.Artifact{
				Revision: "main@sha1:" + test.CommitIDs[0],
			}
		})

		_, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(deployer)})
		test.AssertErrorMatch(t, "failed to load kustomizationRef test-kustomization", err)
	})

	t.Run("reconciling GitRepository with unpopulated GitRepository artifact", func(t *testing.T) {
		ctx := log.IntoContext(context.TODO(), testr.New(t))
		deployer := test.NewKustomizationAutoDeployer()
		test.AssertNoError(t, k8sClient.Create(ctx, deployer))
		defer cleanupResource(t, k8sClient, deployer)

		repo := test.NewGitRepository()
		test.AssertNoError(t, k8sClient.Create(ctx, repo))
		defer cleanupResource(t, k8sClient, repo)

		kustomization := test.NewKustomization(repo)
		test.AssertNoError(t, k8sClient.Create(ctx, kustomization))
		defer cleanupResource(t, k8sClient, kustomization)

		_, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(deployer)})
		test.AssertNoError(t, err)
	})

	t.Run("reconciling error listing commits", func(t *testing.T) {
		ctx := log.IntoContext(context.TODO(), testr.New(t))
		deployer := test.NewKustomizationAutoDeployer(func(tr *deployerv1.KustomizationAutoDeployer) {
			tr.Spec.CommitLimit = 40
		})
		test.AssertNoError(t, k8sClient.Create(ctx, deployer))
		defer cleanupResource(t, k8sClient, deployer)

		// both use the same commit IDs test.CommitIDs[4]
		repo := test.NewGitRepository()
		test.AssertNoError(t, k8sClient.Create(ctx, repo))
		defer cleanupResource(t, k8sClient, repo)
		test.UpdateRepoStatus(t, k8sClient, repo, func(r *sourcev1.GitRepository) {
			r.Status.Artifact = &sourcev1.Artifact{
				Revision: "main@sha1:" + test.CommitIDs[4],
			}
		})

		kustomization := test.NewKustomization(repo)
		test.AssertNoError(t, k8sClient.Create(ctx, kustomization))
		defer cleanupResource(t, k8sClient, kustomization)
		kustomization.Status.LastAppliedRevision = "main@sha1:" + test.CommitIDs[4]
		test.AssertNoError(t, k8sClient.Status().Update(ctx, kustomization))

		_, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(deployer)})
		test.AssertErrorMatch(t, "not enough commit IDs to fulfill request", err)
	})

	t.Run("reconciling GitRepository with non-head commit", func(t *testing.T) {
		ctx := log.IntoContext(context.TODO(), testr.New(t))
		deployer := test.NewKustomizationAutoDeployer()
		test.AssertNoError(t, k8sClient.Create(ctx, deployer))
		defer cleanupResource(t, k8sClient, deployer)

		// both use the same commit IDs test.CommitIDs[4]
		repo := test.NewGitRepository()
		test.AssertNoError(t, k8sClient.Create(ctx, repo))
		defer cleanupResource(t, k8sClient, repo)

		test.UpdateRepoStatus(t, k8sClient, repo, func(r *sourcev1.GitRepository) {
			r.Status.Artifact = &sourcev1.Artifact{
				Revision: "main@sha1:" + test.CommitIDs[4],
			}
		})

		kustomization := test.NewKustomization(repo)
		test.AssertNoError(t, k8sClient.Create(ctx, kustomization))
		defer cleanupResource(t, k8sClient, kustomization)
		kustomization.Status.LastAppliedRevision = "main@sha1:" + test.CommitIDs[4]
		test.AssertNoError(t, k8sClient.Status().Update(ctx, kustomization))

		_, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(deployer)})
		test.AssertNoError(t, err)

		updated := &deployerv1.KustomizationAutoDeployer{}
		test.AssertNoError(t, k8sClient.Get(ctx, client.ObjectKeyFromObject(deployer), updated))

		// one closer to HEAD
		want := "main@sha1:" + test.CommitIDs[3]
		if updated.Status.LatestCommit != want {
			t.Errorf("failed to update with latest commit, got %q, want %q", updated.Status.LatestCommit, want)
		}

		updatedRepo := &sourcev1.GitRepository{}
		test.AssertNoError(t, k8sClient.Get(ctx, client.ObjectKeyFromObject(repo), updatedRepo))
		if updatedRepo.Spec.Reference.Commit != test.CommitIDs[3] {
			t.Errorf("failed to configure the GitRepository with the correct commit got %q, want %q", updatedRepo.Spec.Reference.Commit, test.CommitIDs[3])
		}
	})

	t.Run("reconciling GitRepository with head commit", func(t *testing.T) {
		ctx := log.IntoContext(context.TODO(), testr.New(t))
		deployer := test.NewKustomizationAutoDeployer()
		test.AssertNoError(t, k8sClient.Create(ctx, deployer))
		defer cleanupResource(t, k8sClient, deployer)

		// both use the same commit IDs test.CommitIDs[0]
		repo := test.NewGitRepository()
		test.AssertNoError(t, k8sClient.Create(ctx, repo))
		defer cleanupResource(t, k8sClient, repo)

		test.UpdateRepoStatus(t, k8sClient, repo, func(r *sourcev1.GitRepository) {
			r.Status.Artifact = &sourcev1.Artifact{
				Revision: "main@sha1:" + test.CommitIDs[0],
			}
		})

		kustomization := test.NewKustomization(repo)
		test.AssertNoError(t, k8sClient.Create(ctx, kustomization))
		defer cleanupResource(t, k8sClient, kustomization)
		kustomization.Status.LastAppliedRevision = "main@sha1:" + test.CommitIDs[0]
		test.AssertNoError(t, k8sClient.Status().Update(ctx, kustomization))

		result, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(deployer)})
		test.AssertNoError(t, err)

		updated := &deployerv1.KustomizationAutoDeployer{}
		test.AssertNoError(t, k8sClient.Get(ctx, client.ObjectKeyFromObject(deployer), updated))

		want := ctrl.Result{RequeueAfter: time.Minute * 3}
		if diff := cmp.Diff(want, result); diff != "" {
			t.Errorf("failed to requeue manually:\n%s", diff)
		}

		// the latest commit be the HEAD
		wantCommit := "main@sha1:" + test.CommitIDs[0]
		if updated.Status.LatestCommit != wantCommit {
			t.Errorf("failed to update with latest commit, got %q, want %q", updated.Status.LatestCommit, wantCommit)
		}

		// the latest commit be the HEAD
		updatedRepo := &sourcev1.GitRepository{}
		test.AssertNoError(t, k8sClient.Get(ctx, client.ObjectKeyFromObject(repo), updatedRepo))
		if updatedRepo.Spec.Reference.Commit != test.CommitIDs[0] {
			t.Errorf("failed to configure the GitRepository with the correct commit got %q, want %q", updatedRepo.Spec.Reference.Commit, test.CommitIDs[0])
		}
	})

	t.Run("reconciling with deployed HEAD commit", func(t *testing.T) {
		ctx := log.IntoContext(context.TODO(), testr.New(t))
		deployer := test.NewKustomizationAutoDeployer()
		test.AssertNoError(t, k8sClient.Create(ctx, deployer))
		defer cleanupResource(t, k8sClient, deployer)

		// both use the same commit IDs test.CommitIDs[0]
		repo := test.NewGitRepository()
		test.AssertNoError(t, k8sClient.Create(ctx, repo))
		defer cleanupResource(t, k8sClient, repo)

		test.UpdateRepoStatus(t, k8sClient, repo, func(r *sourcev1.GitRepository) {
			r.Status.Artifact = &sourcev1.Artifact{
				Revision: "main@sha1:" + test.CommitIDs[0],
			}
		})

		kustomization := test.NewKustomization(repo)
		test.AssertNoError(t, k8sClient.Create(ctx, kustomization))
		defer cleanupResource(t, k8sClient, kustomization)
		kustomization.Status.LastAppliedRevision = "main@sha1:" + test.CommitIDs[0]
		test.AssertNoError(t, k8sClient.Status().Update(ctx, kustomization))

		_, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(deployer)})
		test.AssertNoError(t, err)

		updated := &deployerv1.KustomizationAutoDeployer{}
		test.AssertNoError(t, k8sClient.Get(ctx, client.ObjectKeyFromObject(deployer), updated))
		// the latest commit be the HEAD
		wantCommit := "main@sha1:" + test.CommitIDs[0]
		if updated.Status.LatestCommit != wantCommit {
			t.Errorf("failed to update with latest commit, got %q, want %q", updated.Status.LatestCommit, wantCommit)
		}

		// the latest commit be the HEAD
		updatedRepo := &sourcev1.GitRepository{}
		test.AssertNoError(t, k8sClient.Get(ctx, client.ObjectKeyFromObject(repo), updatedRepo))
		if updatedRepo.Spec.Reference.Commit != test.CommitIDs[0] {
			t.Errorf("failed to configure the GitRepository with the correct commit got %q, want %q", updatedRepo.Spec.Reference.Commit, test.CommitIDs[0])
		}
	})

	t.Run("reconciling GitRepository with desired commit configured", func(t *testing.T) {
		ctx := log.IntoContext(context.TODO(), testr.New(t))
		deployer := test.NewKustomizationAutoDeployer()
		test.AssertNoError(t, k8sClient.Create(ctx, deployer))
		defer cleanupResource(t, k8sClient, deployer)

		// both use the same commit IDs test.CommitIDs[0]
		repo := test.NewGitRepository(func(gr *sourcev1.GitRepository) {
			gr.Spec.Reference.Commit = test.CommitIDs[0]
		})
		test.AssertNoError(t, k8sClient.Create(ctx, repo))
		defer cleanupResource(t, k8sClient, repo)
		// But the GitRepository hasn't updated from test.CommitIDs[1]

		test.UpdateRepoStatus(t, k8sClient, repo, func(r *sourcev1.GitRepository) {
			r.Status.Artifact = &sourcev1.Artifact{
				Revision: "main@sha1:" + test.CommitIDs[1],
			}
		})

		kustomization := test.NewKustomization(repo)
		test.AssertNoError(t, k8sClient.Create(ctx, kustomization))
		defer cleanupResource(t, k8sClient, kustomization)
		kustomization.Status.LastAppliedRevision = "main@sha1:" + test.CommitIDs[1]
		test.AssertNoError(t, k8sClient.Status().Update(ctx, kustomization))

		_, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(deployer)})
		test.AssertNoError(t, err)

		test.AssertNoError(t, k8sClient.Get(ctx, client.ObjectKeyFromObject(deployer), deployer))

		// the latest commit be the HEAD
		wantCommit := "main@sha1:" + test.CommitIDs[0]
		if deployer.Status.LatestCommit != wantCommit {
			t.Errorf("failed to update with latest commit, got %q, want %q", deployer.Status.LatestCommit, wantCommit)
		}

		// the latest commit be the HEAD
		updatedRepo := &sourcev1.GitRepository{}
		test.AssertNoError(t, k8sClient.Get(ctx, client.ObjectKeyFromObject(repo), updatedRepo))
		if updatedRepo.Spec.Reference.Commit != test.CommitIDs[0] {
			t.Errorf("failed to configure the GitRepository with the correct commit got %q, want %q", updatedRepo.Spec.Reference.Commit, test.CommitIDs[0])
		}
	})

}

func cleanupResource(t *testing.T, cl client.Client, obj client.Object) {
	t.Helper()
	if err := cl.Delete(context.TODO(), obj); err != nil {
		t.Fatal(err)
	}
}

// Make this an interface!
func testRevisionLister(commitIDs []string) RevisionLister {
	return func(ctx context.Context, url string, options git.ListOptions) ([]string, error) {
		if options.MaxCommits > len(commitIDs) {
			return nil, errors.New("not enough commit IDs to fulfill request")
		}
		return commitIDs, nil
	}
}
