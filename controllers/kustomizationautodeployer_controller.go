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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	fluxv1alpha1 "github.com/gitops-tools/kustomization-auto-deployer/api/v1alpha1"
	"github.com/gitops-tools/kustomization-auto-deployer/pkg/git"
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

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *KustomizationAutoDeployerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KustomizationAutoDeployerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&fluxv1alpha1.KustomizationAutoDeployer{}).
		Complete(r)
}
