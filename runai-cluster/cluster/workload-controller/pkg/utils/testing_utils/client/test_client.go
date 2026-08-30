package client

import (
	"context"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type FakeClientWithFieldSelector struct {
	client.WithWatch
}

func (f *FakeClientWithFieldSelector) List(
	ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	listOpts := client.ListOptions{}
	listOpts.ApplyOptions(opts)
	if listOpts.FieldSelector == nil || listOpts.FieldSelector.Empty() ||
		!strings.Contains(listOpts.FieldSelector.String(), common.PrivilegedWorkloadCrudPermissionsField) {
		return f.WithWatch.List(ctx, list, opts...)
	}
	fieldSelector := listOpts.FieldSelector.Requirements()[0]

	allAccessRulesList := &v1alpha1.AccessRuleList{}
	err := f.WithWatch.List(ctx, allAccessRulesList)
	if err != nil {
		return err
	}

	result := []runtime.Object{}
	for index := range allAccessRulesList.Items {
		accessRule := &allAccessRulesList.Items[index]
		resultKeys := common.PrivilegedWorkloadCrudPermissionsIndexer(accessRule)
		if len(resultKeys) == 0 {
			continue
		}

		for _, key := range resultKeys {
			if fieldSelector.Value == key {
				result = append(result, accessRule)
			}
		}
	}
	return apimeta.SetList(list, result)
}
