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
	"fmt"
	"strings"

	"github.com/fluxcd/pkg/runtime/patch"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	deployerv1 "github.com/gitops-tools/kustomization-auto-deployer/api/v1alpha1"
	"github.com/gitops-tools/kustomization-auto-deployer/pkg/git"
)

const (
	kustomizationIndexKey string = ".metadata.kustomization"
)

// RevisionLister is a function type that queries revisions from a git URL.
type RevisionLister func(ctx context.Context, url string, options git.ListOptions) ([]string, error)

// KustomizationAutoDeployerReconciler reconciles a KustomizationAutoDeployer object
type KustomizationAutoDeployerReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	RevisionLister RevisionLister
}

//+kubebuilder:rbac:groups=flux.gitops.pro,resources=kustomizationautodeployers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=flux.gitops.pro,resources=kustomizationautodeployers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=flux.gitops.pro,resources=kustomizationautodeployers/finalizers,verbs=update
//+kubebuilder:rbac:groups=source.toolkit.fluxcd.io,resources=gitrepositories,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=kustomize.toolkit.fluxcd.io,resources=kustomizations,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *KustomizationAutoDeployerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var deployer deployerv1.KustomizationAutoDeployer
	if err := r.Client.Get(ctx, req.NamespacedName, &deployer); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !deployer.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	var kustomization kustomizev1.Kustomization
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: deployer.Spec.KustomizationRef.Name}, &kustomization); err != nil {
		logger.Error(err, "loading Kustomization for KustomizationAutoDeployer")
		return ctrl.Result{}, fmt.Errorf("failed to load kustomizationRef %s: %w", deployer.Spec.KustomizationRef.Name, err)
	}

	// TODO: What if the Kustomization hasn't applied!
	kustomizationBranch, kustomizationCommitID := parseRevision(kustomization.Status.LastAppliedRevision)
	logger.Info("kustomization loaded", "branch", kustomizationBranch, "commitID", kustomizationCommitID)

	var gitRepository sourcev1.GitRepository
	sourceNamespace := kustomization.Spec.SourceRef.Namespace
	if sourceNamespace == "" {
		sourceNamespace = req.Namespace
	}

	// TODO: Check that the SourceRef is to a GitRepository
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: sourceNamespace, Name: kustomization.Spec.SourceRef.Name}, &gitRepository); err != nil {
		logger.Error(err, "loading GitRepository for KustomizationAutoDeployerReconciler")
		return ctrl.Result{}, fmt.Errorf("failed to load sourceRef %s/%s: %w", sourceNamespace, kustomization.Spec.SourceRef.Name, err)
	}

	// TODO: if the GitRepository is using a branch and not a ref, this is an error
	// TODO: What if the Artifact is not available - this will panic!
	if gitRepository.Status.Artifact == nil {
		logger.Info("git repository status not yet populated")
		return ctrl.Result{}, nil
	}

	repoBranch, repoCommitID := parseRevision(gitRepository.Status.Artifact.Revision)
	logger.Info("gitRepository loaded", "branch", repoBranch, "commitID", repoCommitID, "desiredCommitID", gitRepository.Spec.Reference.Commit)

	if repoBranch != kustomizationBranch {
		logger.Info("kustomization branch does not match repo branch - no further processing", "gitRepositoryBranch", repoBranch, "kustomizationBranch", kustomizationBranch)
		return ctrl.Result{}, nil
	}

	// IF the last applied version in the Kustomization == the current version
	// of the GitRepository, we can look for a new version.
	// TODO: is this right?
	if kustomizationCommitID != repoCommitID {
		logger.Info("kustomization commit does not match git repository commit no further processing", "gitRepositoryCommitID", repoCommitID, "kustomizationCommitID", kustomizationCommitID)
		return ctrl.Result{}, nil
	}

	revisions, err := r.RevisionLister(ctx, gitRepository.Spec.URL, git.ListOptions{MaxCommits: deployer.Spec.CommitLimit})
	if err != nil {
		logger.Error(err, "listing revisions", "url", gitRepository.Spec.URL)
		return ctrl.Result{}, fmt.Errorf("failed to list revisions in repo %s: %w", gitRepository.Spec.URL, err)
	}

	currentCommitIndex := stringIndex(kustomizationCommitID, revisions)
	if currentCommitIndex < 1 {
		logger.Info("no changes to deploy")
		// TODO: Refactor this to avoid duplication!
		deployer.Status.LatestCommit = commitReference(repoBranch, repoCommitID)
		deployer.Status.ObservedGeneration = deployer.Generation
		if err := r.patchStatus(ctx, req, deployer.Status); err != nil {
			logger.Error(err, "failed to reconcile")
			return ctrl.Result{}, err
		}

		// Return the duration here because we may not see a rereconciliation
		// triggered by a watch, and we need to check for new commits at some
		// point.
		return ctrl.Result{RequeueAfter: deployer.Spec.Interval.Duration}, nil
	}

	nextCommitToDeploy := revisions[currentCommitIndex-1]
	if repoCommitID == nextCommitToDeploy {
		logger.Info("already deployed, nothing to do")

		return ctrl.Result{}, nil
	}

	if gitRepository.Spec.Reference.Commit == nextCommitToDeploy {
		logger.Info("already requested deploy, nothing to do")
		// TODO: Refactor this to avoid duplication!
		deployer.Status.LatestCommit = commitReference(repoBranch, nextCommitToDeploy)
		deployer.Status.ObservedGeneration = deployer.Generation
		if err := r.patchStatus(ctx, req, deployer.Status); err != nil {
			logger.Error(err, "failed to reconcile")
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	patchHelper, err := patch.NewHelper(&gitRepository, r.Client)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create patch helper for GitRepository: %w", err)
	}

	logger.Info("identified next commit - patching GitRepository", "nextCommitID", nextCommitToDeploy, "repositoryName", gitRepository.GetName(), "repositoryNamespace", gitRepository.GetNamespace())
	gitRepository.Spec.Reference.Commit = nextCommitToDeploy

	if err := patchHelper.Patch(ctx, &gitRepository); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to update GitRepository: %w", err)
	}

	// TODO: Setup condition
	// TODO: Refactor this to avoid duplication!
	deployer.Status.LatestCommit = commitReference(repoBranch, nextCommitToDeploy)
	deployer.Status.ObservedGeneration = deployer.Generation
	if err := r.patchStatus(ctx, req, deployer.Status); err != nil {
		logger.Error(err, "failed to reconcile")
	}

	return ctrl.Result{}, nil
}

func (r *KustomizationAutoDeployerReconciler) patchStatus(ctx context.Context, req ctrl.Request, newStatus deployerv1.KustomizationAutoDeployerStatus) error {
	var deployer deployerv1.KustomizationAutoDeployer
	if err := r.Get(ctx, req.NamespacedName, &deployer); err != nil {
		return err
	}

	patch := client.MergeFrom(deployer.DeepCopy())
	deployer.Status = newStatus

	return r.Status().Patch(ctx, &deployer, patch)
}

// SetupWithManager sets up the controller with the Manager.
func (r *KustomizationAutoDeployerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Index the KustomizationAutoDeployer by the Kustomization references they point at.
	if err := mgr.GetCache().IndexField(
		context.TODO(), &deployerv1.KustomizationAutoDeployer{}, kustomizationIndexKey, func(o client.Object) []string {
			gt, ok := o.(*deployerv1.KustomizationAutoDeployer)
			if !ok {
				panic(fmt.Sprintf("Expected a GitRepositoryTracker, got %T", o))
			}

			return []string{fmt.Sprintf("%s/%s", gt.GetNamespace(), gt.Spec.KustomizationRef.Name)}
		}); err != nil {
		return fmt.Errorf("failed setting index fields for Kustomizations: %w", err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&deployerv1.KustomizationAutoDeployer{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Watches(
			&source.Kind{Type: &kustomizev1.Kustomization{}},
			handler.EnqueueRequestsFromMapFunc(r.kustomizationToAutoDeployer),
		).
		Complete(r)
}

// TODO: Fold these two into a closure with the index key!
func (r *KustomizationAutoDeployerReconciler) kustomizationToAutoDeployer(obj client.Object) []reconcile.Request {
	ctx := context.Background()
	var list deployerv1.KustomizationAutoDeployerList

	if err := r.List(ctx, &list, client.MatchingFields{
		kustomizationIndexKey: client.ObjectKeyFromObject(obj).String(),
	}); err != nil {
		return nil
	}

	result := []reconcile.Request{}
	for _, v := range list.Items {
		result = append(result, reconcile.Request{NamespacedName: types.NamespacedName{Name: v.GetName(), Namespace: v.GetNamespace()}})
	}

	return result
}

// parse main@sha1:40d6b21b888db0ca794876cf7bdd399e3da2137e
func parseRevision(s string) (string, string) {
	elements := strings.SplitN(s, ":", 2)
	if len(elements) != 2 {
		return "", ""
	}

	revision := elements[1]

	elements = strings.SplitN(elements[0], "@", 2)

	if len(elements) > 1 {
		return elements[0], revision
	}

	return "", revision
}

func stringIndex(s string, ss []string) int {
	for i := range ss {
		if ss[i] == s {
			return i
		}
	}

	return -1
}

func commitReference(branch, commitID string) string {
	return branch + "@sha1:" + commitID
}
