package legacycleanup

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var customResourceDefinitionGVR = schema.GroupVersionResource{
	Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions",
}

type resourceDescriptor struct {
	GVR                     schema.GroupVersionResource
	Kind                    string
	CRDName                 string
	RemoveWorkloadFinalizer bool
}

var legacyWorkloadResources = []resourceDescriptor{
	{GVR: schema.GroupVersionResource{Group: "run.ai", Version: "v2alpha1", Resource: "trainingworkloads"}, Kind: "TrainingWorkload", CRDName: "trainingworkloads.run.ai", RemoveWorkloadFinalizer: true},
	{GVR: schema.GroupVersionResource{Group: "run.ai", Version: "v2alpha1", Resource: "interactiveworkloads"}, Kind: "InteractiveWorkload", CRDName: "interactiveworkloads.run.ai", RemoveWorkloadFinalizer: true},
	{GVR: schema.GroupVersionResource{Group: "run.ai", Version: "v2alpha1", Resource: "distributedworkloads"}, Kind: "DistributedWorkload", CRDName: "distributedworkloads.run.ai", RemoveWorkloadFinalizer: true},
	{GVR: schema.GroupVersionResource{Group: "run.ai", Version: "v2alpha1", Resource: "inferenceworkloads"}, Kind: "InferenceWorkload", CRDName: "inferenceworkloads.run.ai"},
	{GVR: schema.GroupVersionResource{Group: "run.ai", Version: "v1alpha1", Resource: "distributedinferenceworkloads"}, Kind: "DistributedInferenceWorkload", CRDName: "distributedinferenceworkloads.run.ai"},
	{GVR: schema.GroupVersionResource{Group: "run.ai", Version: "v1", Resource: "runaijobs"}, Kind: "RunaiJob", CRDName: "runaijobs.run.ai"},
}

func discoverServedWorkloads(ctx context.Context, dynamicClient dynamic.Interface) ([]resourceDescriptor, error) {
	served := make([]resourceDescriptor, 0, len(legacyWorkloadResources))
	crds := dynamicClient.Resource(customResourceDefinitionGVR)
	for _, descriptor := range legacyWorkloadResources {
		crd, err := crds.Get(ctx, descriptor.CRDName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("discover legacy API %s: %w", descriptor.GVR, err)
		}
		isServed, err := crdServesVersion(crd, descriptor.GVR.Version)
		if err != nil {
			return nil, fmt.Errorf("inspect CRD %s: %w", descriptor.CRDName, err)
		}
		if isServed {
			served = append(served, descriptor)
		}
	}
	return served, nil
}

func crdServesVersion(crd *unstructured.Unstructured, version string) (bool, error) {
	versions, found, err := unstructured.NestedSlice(crd.Object, "spec", "versions")
	if err != nil {
		return false, err
	}
	if !found {
		return false, fmt.Errorf("spec.versions is missing")
	}
	for _, raw := range versions {
		item, ok := raw.(map[string]any)
		if !ok {
			return false, fmt.Errorf("spec.versions contains %T", raw)
		}
		name, _, _ := unstructured.NestedString(item, "name")
		served, _, _ := unstructured.NestedBool(item, "served")
		if name == version {
			return served, nil
		}
	}
	return false, nil
}

func listObjects(ctx context.Context, dynamicClient dynamic.Interface, descriptor resourceDescriptor) (
	[]unstructured.Unstructured, error,
) {
	list, err := dynamicClient.Resource(descriptor.GVR).Namespace(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list %s: %w", descriptor.GVR.Resource, err)
	}
	return list.Items, nil
}
