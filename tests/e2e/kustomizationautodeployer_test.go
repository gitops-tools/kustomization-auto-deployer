package tests

import (
	"context"
	"testing"

	"github.com/gitops-tools/kustomization-auto-deployer/test"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var kustomizationGVK = schema.GroupVersionKind{
	Group:   "kustomize.toolkit.fluxcd.io",
	Kind:    "Kustomization",
	Version: "v1beta2",
}

func TestReconciling(t *testing.T) {
	ctx := context.TODO()
	kd := test.NewKustomizationAutoDeployer()

	test.AssertNoError(t, testEnv.Create(ctx, kd))
	defer cleanupResource(t, testEnv, kd)
}

func cleanupResource(t *testing.T, cl client.Client, obj client.Object) {
	t.Helper()
	if err := cl.Delete(context.TODO(), obj); err != nil {
		t.Fatal(err)
	}
}
