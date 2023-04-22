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
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	"github.com/go-logr/logr/testr"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

var testCommitIDs = []string{
	"6f935147b28e38a99a700843e4893a801c3c8148",
	"6b36589c98beb53b564aa2b838d8d045f0d0794d",
	"d2fb242401eb5146b5935f2b8483ffe9ffcef3fd",
	"ece9b0c9300f65546ca8f57d80bcd6378f051e59",
	"dd587268153ad335545a53f15efee4ecfabcd1c8",
	"5ed746679fd79a91517ab32b83ed68c4a315040d",
	"4553f10114fd9c7e82b65d65ffa8e124911b08e6",
	"e5d6eeed2f5574737ae73a33bfd208a37b50ccd8",
	"848e164167cc4db24c21fc47a3da190e5bb801c1",
	"d756705b4a8462d68838add62892904bcaa1a279",
	"1d27b733842bb5a29a118c950f521e0ef6ee31f8",
	"9574855e863316ae161c2f9512cc1203d79d59fa",
	"edd180938ddfd80161f20c22fa58a2b16dfd66f3",
	"039dc250bc8bd17c1b3bf5abe6bce68e338e966e",
	"fa079ad89794b3ebf5932f55ced9e0a714841fac",
	"d7890a3d6b262a48c126a1ebdc0e71239d33cc8c",
	"6e430dd9f8fb6f1830f6d2d765e13cf9f3ddfe4a",
}

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
		RevisionLister: testRevisionLister(testCommitIDs),
	}

	test.AssertNoError(t, reconciler.SetupWithManager(mgr))

	// t.Run("reconciling with missing GitRepository", func(t *testing.T) {
	// 	ctx := log.IntoContext(context.TODO(), testr.New(t))
	// 	deployer := makeTestKustomizationAutoDeployer()
	// 	test.AssertNoError(t, k8sClient.Create(ctx, deployer))
	// 	defer cleanupResource(t, k8sClient, deployer)

	// 	_, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(deployer)})
	// 	// TODO: This should be asserting a condition exists!
	// 	test.AssertErrorMatch(t, "failed to load gitRepositoryRef test-gitrepository", err)
	// })

	t.Run("reconciling GitRepository with missing Kustomization", func(t *testing.T) {
		ctx := log.IntoContext(context.TODO(), testr.New(t))
		deployer := makeTestKustomizationAutoDeployer()
		test.AssertNoError(t, k8sClient.Create(ctx, deployer))
		defer cleanupResource(t, k8sClient, deployer)

		repo := makeTestGitRepository()
		test.AssertNoError(t, k8sClient.Create(ctx, repo))
		defer cleanupResource(t, k8sClient, repo)

		updateRepoStatus(t, k8sClient, repo, func(r *sourcev1.GitRepository) {
			repo.Status.Artifact = &sourcev1.Artifact{
				Revision: "main@sha1:" + testCommitIDs[0],
			}
		})

		_, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(deployer)})
		test.AssertErrorMatch(t, "failed to load kustomizationRef test-kustomization", err)
	})

	t.Run("reconciling GitRepository with unpopulated GitRepository artifact", func(t *testing.T) {
		ctx := log.IntoContext(context.TODO(), testr.New(t))
		deployer := makeTestKustomizationAutoDeployer()
		test.AssertNoError(t, k8sClient.Create(ctx, deployer))
		defer cleanupResource(t, k8sClient, deployer)

		kustomization := makeTestKustomization()
		test.AssertNoError(t, k8sClient.Create(ctx, kustomization))
		defer cleanupResource(t, k8sClient, kustomization)

		repo := makeTestGitRepository()
		test.AssertNoError(t, k8sClient.Create(ctx, repo))
		defer cleanupResource(t, k8sClient, repo)

		_, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(deployer)})
		test.AssertNoError(t, err)
	})

	t.Run("reconciling error listing commits", func(t *testing.T) {
		ctx := log.IntoContext(context.TODO(), testr.New(t))
		deployer := makeTestKustomizationAutoDeployer(func(tr *deployerv1.KustomizationAutoDeployer) {
			tr.Spec.CommitLimit = 40
		})
		test.AssertNoError(t, k8sClient.Create(ctx, deployer))
		defer cleanupResource(t, k8sClient, deployer)

		// both use the same commit IDs testCommitIDs[4]
		repo := makeTestGitRepository()
		test.AssertNoError(t, k8sClient.Create(ctx, repo))
		defer cleanupResource(t, k8sClient, repo)
		updateRepoStatus(t, k8sClient, repo, func(r *sourcev1.GitRepository) {
			r.Status.Artifact = &sourcev1.Artifact{
				Revision: "main@sha1:" + testCommitIDs[4],
			}
		})

		kustomization := makeTestKustomization()
		test.AssertNoError(t, k8sClient.Create(ctx, kustomization))
		defer cleanupResource(t, k8sClient, kustomization)
		kustomization.Status.LastAppliedRevision = "main@sha1:" + testCommitIDs[4]
		test.AssertNoError(t, k8sClient.Status().Update(ctx, kustomization))

		_, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(deployer)})
		test.AssertErrorMatch(t, "not enough commit IDs to fulfill request", err)
	})

	t.Run("reconciling GitRepository with non-head commit", func(t *testing.T) {
		ctx := log.IntoContext(context.TODO(), testr.New(t))
		deployer := makeTestKustomizationAutoDeployer()
		test.AssertNoError(t, k8sClient.Create(ctx, deployer))
		defer cleanupResource(t, k8sClient, deployer)

		// both use the same commit IDs testCommitIDs[4]
		repo := makeTestGitRepository()
		test.AssertNoError(t, k8sClient.Create(ctx, repo))
		defer cleanupResource(t, k8sClient, repo)

		updateRepoStatus(t, k8sClient, repo, func(r *sourcev1.GitRepository) {
			r.Status.Artifact = &sourcev1.Artifact{
				Revision: "main@sha1:" + testCommitIDs[4],
			}
		})

		kustomization := makeTestKustomization()
		test.AssertNoError(t, k8sClient.Create(ctx, kustomization))
		defer cleanupResource(t, k8sClient, kustomization)
		kustomization.Status.LastAppliedRevision = "main@sha1:" + testCommitIDs[4]
		test.AssertNoError(t, k8sClient.Status().Update(ctx, kustomization))

		_, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(deployer)})
		test.AssertNoError(t, err)

		updated := &deployerv1.KustomizationAutoDeployer{}
		test.AssertNoError(t, k8sClient.Get(ctx, client.ObjectKeyFromObject(deployer), updated))

		// one closer to HEAD
		want := "main@sha1:" + testCommitIDs[3]
		if updated.Status.LatestCommit != want {
			t.Errorf("failed to update with latest commit, got %q, want %q", updated.Status.LatestCommit, want)
		}

		updatedRepo := &sourcev1.GitRepository{}
		test.AssertNoError(t, k8sClient.Get(ctx, client.ObjectKeyFromObject(repo), updatedRepo))
		if updatedRepo.Spec.Reference.Commit != testCommitIDs[3] {
			t.Errorf("failed to configure the GitRepository with the correct commit got %q, want %q", updatedRepo.Spec.Reference.Commit, testCommitIDs[3])
		}
	})

	t.Run("reconciling GitRepository with head commit", func(t *testing.T) {
		ctx := log.IntoContext(context.TODO(), testr.New(t))
		deployer := makeTestKustomizationAutoDeployer()
		test.AssertNoError(t, k8sClient.Create(ctx, deployer))
		defer cleanupResource(t, k8sClient, deployer)

		// both use the same commit IDs testCommitIDs[0]
		repo := makeTestGitRepository()
		test.AssertNoError(t, k8sClient.Create(ctx, repo))
		defer cleanupResource(t, k8sClient, repo)

		updateRepoStatus(t, k8sClient, repo, func(r *sourcev1.GitRepository) {
			r.Status.Artifact = &sourcev1.Artifact{
				Revision: "main@sha1:" + testCommitIDs[0],
			}
		})

		kustomization := makeTestKustomization()
		test.AssertNoError(t, k8sClient.Create(ctx, kustomization))
		defer cleanupResource(t, k8sClient, kustomization)
		kustomization.Status.LastAppliedRevision = "main@sha1:" + testCommitIDs[0]
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
		wantCommit := "main@sha1:" + testCommitIDs[0]
		if updated.Status.LatestCommit != wantCommit {
			t.Errorf("failed to update with latest commit, got %q, want %q", updated.Status.LatestCommit, wantCommit)
		}

		// the latest commit be the HEAD
		updatedRepo := &sourcev1.GitRepository{}
		test.AssertNoError(t, k8sClient.Get(ctx, client.ObjectKeyFromObject(repo), updatedRepo))
		if updatedRepo.Spec.Reference.Commit != testCommitIDs[0] {
			t.Errorf("failed to configure the GitRepository with the correct commit got %q, want %q", updatedRepo.Spec.Reference.Commit, testCommitIDs[0])
		}
	})

	t.Run("reconciling with deployed HEAD commit", func(t *testing.T) {
		ctx := log.IntoContext(context.TODO(), testr.New(t))
		deployer := makeTestKustomizationAutoDeployer()
		test.AssertNoError(t, k8sClient.Create(ctx, deployer))
		defer cleanupResource(t, k8sClient, deployer)

		// both use the same commit IDs testCommitIDs[0]
		repo := makeTestGitRepository()
		test.AssertNoError(t, k8sClient.Create(ctx, repo))
		defer cleanupResource(t, k8sClient, repo)

		updateRepoStatus(t, k8sClient, repo, func(r *sourcev1.GitRepository) {
			r.Status.Artifact = &sourcev1.Artifact{
				Revision: "main@sha1:" + testCommitIDs[0],
			}
		})

		kustomization := makeTestKustomization()
		test.AssertNoError(t, k8sClient.Create(ctx, kustomization))
		defer cleanupResource(t, k8sClient, kustomization)
		kustomization.Status.LastAppliedRevision = "main@sha1:" + testCommitIDs[0]
		test.AssertNoError(t, k8sClient.Status().Update(ctx, kustomization))

		_, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(deployer)})
		test.AssertNoError(t, err)

		updated := &deployerv1.KustomizationAutoDeployer{}
		test.AssertNoError(t, k8sClient.Get(ctx, client.ObjectKeyFromObject(deployer), updated))
		// the latest commit be the HEAD
		wantCommit := "main@sha1:" + testCommitIDs[0]
		if updated.Status.LatestCommit != wantCommit {
			t.Errorf("failed to update with latest commit, got %q, want %q", updated.Status.LatestCommit, wantCommit)
		}

		// the latest commit be the HEAD
		updatedRepo := &sourcev1.GitRepository{}
		test.AssertNoError(t, k8sClient.Get(ctx, client.ObjectKeyFromObject(repo), updatedRepo))
		if updatedRepo.Spec.Reference.Commit != testCommitIDs[0] {
			t.Errorf("failed to configure the GitRepository with the correct commit got %q, want %q", updatedRepo.Spec.Reference.Commit, testCommitIDs[0])
		}
	})

	t.Run("reconciling GitRepository with desired commit configured", func(t *testing.T) {
		ctx := log.IntoContext(context.TODO(), testr.New(t))
		deployer := makeTestKustomizationAutoDeployer()
		test.AssertNoError(t, k8sClient.Create(ctx, deployer))
		defer cleanupResource(t, k8sClient, deployer)

		// both use the same commit IDs testCommitIDs[0]
		repo := makeTestGitRepository(func(gr *sourcev1.GitRepository) {
			gr.Spec.Reference.Commit = testCommitIDs[0]
		})
		test.AssertNoError(t, k8sClient.Create(ctx, repo))
		defer cleanupResource(t, k8sClient, repo)
		// But the GitRepository hasn't updated from testCommitIDs[1]

		updateRepoStatus(t, k8sClient, repo, func(r *sourcev1.GitRepository) {
			r.Status.Artifact = &sourcev1.Artifact{
				Revision: "main@sha1:" + testCommitIDs[1],
			}
		})

		kustomization := makeTestKustomization()
		test.AssertNoError(t, k8sClient.Create(ctx, kustomization))
		defer cleanupResource(t, k8sClient, kustomization)
		kustomization.Status.LastAppliedRevision = "main@sha1:" + testCommitIDs[1]
		test.AssertNoError(t, k8sClient.Status().Update(ctx, kustomization))

		_, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(deployer)})
		test.AssertNoError(t, err)

		test.AssertNoError(t, k8sClient.Get(ctx, client.ObjectKeyFromObject(deployer), deployer))

		// the latest commit be the HEAD
		wantCommit := "main@sha1:" + testCommitIDs[0]
		if deployer.Status.LatestCommit != wantCommit {
			t.Errorf("failed to update with latest commit, got %q, want %q", deployer.Status.LatestCommit, wantCommit)
		}

		// the latest commit be the HEAD
		updatedRepo := &sourcev1.GitRepository{}
		test.AssertNoError(t, k8sClient.Get(ctx, client.ObjectKeyFromObject(repo), updatedRepo))
		if updatedRepo.Spec.Reference.Commit != testCommitIDs[0] {
			t.Errorf("failed to configure the GitRepository with the correct commit got %q, want %q", updatedRepo.Spec.Reference.Commit, testCommitIDs[0])
		}
	})

}

func cleanupResource(t *testing.T, cl client.Client, obj client.Object) {
	t.Helper()
	if err := cl.Delete(context.TODO(), obj); err != nil {
		t.Fatal(err)
	}
}

func makeTestKustomizationAutoDeployer(opts ...func(*deployerv1.KustomizationAutoDeployer)) *deployerv1.KustomizationAutoDeployer {
	gt := &deployerv1.KustomizationAutoDeployer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-deployer",
			Namespace: "default",
		},
		Spec: deployerv1.KustomizationAutoDeployerSpec{
			CommitLimit: 10,
			Interval:    metav1.Duration{Duration: time.Minute * 3},
			KustomizationRef: meta.LocalObjectReference{
				Name: "test-kustomization",
			},
		},
	}

	for _, opt := range opts {
		opt(gt)
	}

	return gt
}

func makeTestGitRepository(opts ...func(*sourcev1.GitRepository)) *sourcev1.GitRepository {
	gr := &sourcev1.GitRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-gitrepository",
			Namespace: "default",
		},
		Spec: sourcev1.GitRepositorySpec{
			Interval: metav1.Duration{Duration: time.Minute * 5},
			URL:      "https://github.com/gitops-tools/gitrepository-deployer",
			Reference: &sourcev1.GitRepositoryRef{
				Commit: testCommitIDs[0],
			},
		},
	}

	for _, opt := range opts {
		opt(gr)
	}

	return gr
}

func makeTestKustomization(opts ...func(*kustomizev1.Kustomization)) *kustomizev1.Kustomization {
	k := &kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-kustomization",
			Namespace: "default",
		},
		Spec: kustomizev1.KustomizationSpec{
			Interval: metav1.Duration{Duration: time.Minute * 5},
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Kind: "GitRepository",
				Name: "test-gitrepository",
			},
			Path:  "./kustomize",
			Prune: true,
		},
	}

	for _, opt := range opts {
		opt(k)
	}

	return k
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

func updateRepoStatus(t *testing.T, k8sClient client.Client, repo *sourcev1.GitRepository, update func(*sourcev1.GitRepository)) {
	t.Helper()
	ctx := context.TODO()
	test.AssertNoError(t, k8sClient.Get(ctx, client.ObjectKeyFromObject(repo), repo))

	update(repo)
	repo.Status.Artifact.LastUpdateTime = metav1.Now()

	test.AssertNoError(t, k8sClient.Status().Update(ctx, repo))
}
