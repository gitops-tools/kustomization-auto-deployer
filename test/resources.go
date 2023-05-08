package test

import (
	"time"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// This is the default name of the GitRepository created by
	// NewGitRepository.
	DefaultGitRepositoryName = "test-gitrepository"

	// This is the default name of the Kustomization created by NewKustomization.
	DefaultKustomizationName = "test-kustomization"

	DefaultNamespace = "default"
)

// NewGitRepository creates a new GitRepository.
//
// This will use the default name
func NewGitRepository(opts ...func(*sourcev1.GitRepository)) *sourcev1.GitRepository {
	gr := &sourcev1.GitRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultGitRepositoryName,
			Namespace: DefaultNamespace,
		},
		Spec: sourcev1.GitRepositorySpec{
			Interval: metav1.Duration{Duration: time.Minute * 5},
			URL:      "https://github.com/gitops-tools/gitrepository-deployer",
			Reference: &sourcev1.GitRepositoryRef{
				Commit: CommitIDs[0],
			},
		},
	}

	for _, opt := range opts {
		opt(gr)
	}

	return gr
}

func NewKustomization(gr *sourcev1.GitRepository, opts ...func(*kustomizev1.Kustomization)) *kustomizev1.Kustomization {
	k := &kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultKustomizationName,
			Namespace: DefaultNamespace,
		},
		Spec: kustomizev1.KustomizationSpec{
			Interval: metav1.Duration{Duration: time.Minute * 5},
			Path:     "./kustomize",
			Prune:    true,
		},
	}

	if gr != nil {
		k.Spec.SourceRef = kustomizev1.CrossNamespaceSourceReference{
			Kind: "GitRepository",
			Name: gr.GetName(),
		}
	}

	for _, opt := range opts {
		opt(k)
	}

	return k
}
