package test

import (
	"time"

	"github.com/fluxcd/pkg/apis/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	deployerv1 "github.com/gitops-tools/kustomization-auto-deployer/api/v1alpha1"
)

// NewKustomizationAutoDeployer creates and returns a new KustomizationDeployer.
func NewKustomizationAutoDeployer(opts ...func(*deployerv1.KustomizationAutoDeployer)) *deployerv1.KustomizationAutoDeployer {
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
