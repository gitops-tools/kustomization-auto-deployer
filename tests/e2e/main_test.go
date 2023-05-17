package tests

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fluxcd/pkg/runtime/testenv"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	deployerv1 "github.com/gitops-tools/kustomization-auto-deployer/api/v1alpha1"
	"github.com/gitops-tools/kustomization-auto-deployer/controllers"
	"github.com/gitops-tools/kustomization-auto-deployer/controllers/gates"
	"github.com/gitops-tools/kustomization-auto-deployer/controllers/gates/healthcheck"
	"github.com/gitops-tools/kustomization-auto-deployer/controllers/gates/scheduled"
	"github.com/gitops-tools/kustomization-auto-deployer/pkg/git"
	"github.com/gitops-tools/kustomization-auto-deployer/test"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	timeout = 5 * time.Second
)

var (
	testEnv *testenv.Environment
	ctx     = ctrl.SetupSignalHandler()
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestMain(m *testing.M) {
	utilruntime.Must(kustomizev1.AddToScheme(scheme.Scheme))
	utilruntime.Must(deployerv1.AddToScheme(scheme.Scheme))
	utilruntime.Must(sourcev1.AddToScheme(scheme.Scheme))
	testEnv = testenv.New(testenv.WithCRDPath(filepath.Join("..", "..", "config", "crd", "bases"),
		filepath.Join("..", "..", "controllers", "testdata", "crds"),
	))

	if err := (&controllers.KustomizationAutoDeployerReconciler{
		Client:         testEnv,
		Scheme:         testEnv.GetScheme(),
		RevisionLister: testRevisionLister(test.CommitIDs),
		GateFactories: map[string]gates.GateFactory{
			"HealthCheck": healthcheck.Factory(http.DefaultClient),
			"Scheduled":   scheduled.Factory,
		},
	}).SetupWithManager(testEnv); err != nil {
		panic(fmt.Sprintf("Failed to start KustomizationAutoDeployerReconciler: %v", err))
	}

	go func() {
		fmt.Println("Starting the test environment")
		if err := testEnv.Start(ctx); err != nil {
			panic(fmt.Sprintf("Failed to start the test environment manager: %v", err))
		}
	}()
	<-testEnv.Manager.Elected()

	code := m.Run()

	fmt.Println("Stopping the test environment")
	if err := testEnv.Stop(); err != nil {
		panic(fmt.Sprintf("Failed to stop the test environment: %v", err))
	}

	os.Exit(code)
}

// Make this an interface!
func testRevisionLister(commitIDs []string) controllers.RevisionLister {
	return func(ctx context.Context, url string, options git.ListOptions) ([]string, error) {
		if options.MaxCommits > len(commitIDs) {
			return nil, errors.New("not enough commit IDs to fulfill request")
		}
		return commitIDs, nil
	}
}
