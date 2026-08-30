package deletion_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/handlers/deletion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// testBlockerGroups mirrors the runai deletion-blocker set (defined for production by the
// go-operator's project-delete-blockers ConfigMap) so these tests exercise the generic
// ConfigurableBlocker against the same GVKs and label selectors the runai packaging uses.
func testBlockerGroups() []BlockerGroup {
	const (
		runaiGroup               = "run.ai"
		v2alpha1                 = "v2alpha1"
		v1alpha1                 = "v1alpha1"
		RunaiAssetLabel          = "run.ai/asset"
		RunaiResourceLabel       = "run.ai/resource"
		DepartmentScopeLabel     = "run.ai/department"
		ClusterWideResourceLabel = "run.ai/cluster-wide"
		TenantScopeLabel         = "run.ai/tenant-wide"
	)
	assetKinds := []string{"password", "access-key", "generic", "docker-registry"}

	projectScopeExpressions := func() []metav1.LabelSelectorRequirement {
		return []metav1.LabelSelectorRequirement{
			{Key: DepartmentScopeLabel, Operator: metav1.LabelSelectorOpDoesNotExist},
			{Key: ClusterWideResourceLabel, Operator: metav1.LabelSelectorOpDoesNotExist},
			{Key: TenantScopeLabel, Operator: metav1.LabelSelectorOpDoesNotExist},
		}
	}
	projectScopeWithType := func(typeLabel string) *metav1.LabelSelector {
		exprs := append(projectScopeExpressions(), metav1.LabelSelectorRequirement{
			Key: typeLabel, Operator: metav1.LabelSelectorOpIn, Values: assetKinds,
		})
		return &metav1.LabelSelector{MatchExpressions: exprs}
	}

	return []BlockerGroup{
		{DisplayName: "Workloads", Blockers: []Blocker{
			{Group: runaiGroup, Version: v2alpha1, Kind: "TrainingWorkload", DisplayName: "Workloads"},
			{Group: runaiGroup, Version: v2alpha1, Kind: "InferenceWorkload", DisplayName: "Workloads"},
			{Group: runaiGroup, Version: v2alpha1, Kind: "DistributedWorkload", DisplayName: "Workloads"},
			{Group: runaiGroup, Version: v2alpha1, Kind: "InteractiveWorkload", DisplayName: "Workloads"},
			{Group: runaiGroup, Version: v1alpha1, Kind: "ExternalWorkload", DisplayName: "Workloads"},
			{Group: runaiGroup, Version: v1alpha1, Kind: "DistributedInferenceWorkload", DisplayName: "Workloads"},
			{Group: runaiGroup, Version: "v1", Kind: "RunaiJob", DisplayName: "Workloads"},
		}},
		{DisplayName: "DataVolumes", Blockers: []Blocker{{Group: runaiGroup, Version: v1alpha1, Kind: "DataVolume", DisplayName: "DataVolumes"}}},
		{DisplayName: "Secrets", Blockers: []Blocker{
			{Group: "", Version: "v1", Kind: "Secret", LabelSelector: projectScopeWithType(RunaiResourceLabel), DisplayName: "Secrets"},
			{Group: "", Version: "v1", Kind: "Secret", LabelSelector: projectScopeWithType(RunaiAssetLabel), DisplayName: "Secrets"},
		}},
		{DisplayName: "Pvcs", Blockers: []Blocker{
			{Group: "", Version: "v1", Kind: "PersistentVolumeClaim", DisplayName: "Pvcs",
				LabelSelector: &metav1.LabelSelector{MatchExpressions: projectScopeExpressions()}},
		}},
	}
}

// blockerForDisplayName builds the generic ConfigurableBlocker from the test blocker group
// with the given display name (e.g. "Workloads", "Secrets").
func blockerForDisplayName(cli client.Client, displayName string) ConfigurableBlocker {
	for _, group := range testBlockerGroups() {
		if group.DisplayName == displayName {
			return NewConfigurableBlocker(cli, group)
		}
	}
	Fail("no test blocker group with displayName " + displayName)
	return ConfigurableBlocker{}
}

type listErrorClient struct {
	client.Client
	errors map[schema.GroupVersionKind]error
}

func (k8sClient listErrorClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	if err, ok := k8sClient.errors[list.GetObjectKind().GroupVersionKind()]; ok {
		return err
	}
	return k8sClient.Client.List(ctx, list, opts...)
}
