package cluster_relations

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/run-ai/runai/runai-cluster/cluster/status-updater/pkg/controllers/metadata"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var serverResourcesByGroup = make(map[string]*metav1.APIResourceList)
var serverResourcesMutex sync.Mutex

func getOwnerFromTopOwnerReference(ctx context.Context, ownerRef *metadata.TopOwnerReference, namespace string,
	dynamicClient dynamic.Interface, discoveryClient discovery.DiscoveryInterface) (
	*unstructured.Unstructured, *schema.GroupVersionResource, error) {
	resource, err := getGVRForOwnerRef(ctx, ownerRef, discoveryClient)
	if err != nil {
		return nil, nil, err
	}

	owner, err := dynamicClient.Resource(*resource).Namespace(namespace).Get(ctx, ownerRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, resource, err
	}
	// In some edge cases the owner is already deleted and an identical resource with the same name has been created
	if ownerRef.UID != string(owner.GetUID()) {
		return nil, resource, kerrors.NewResourceExpired(
			fmt.Sprintf("Owner resource with uid %s not found. Found resource with different UID %v",
				ownerRef.UID, owner.GetUID()))
	}

	return owner, resource, nil
}

func getGVRForOwnerRef(ctx context.Context, ownerRef *metadata.TopOwnerReference,
	discoveryClient discovery.DiscoveryInterface) (*schema.GroupVersionResource, error) {
	res, err := getKindResourceName(ctx, discoveryClient, ownerRef.Kind, calcApiVersion(ownerRef))
	if err != nil {
		return nil, err
	}

	resource := schema.GroupVersionResource{
		Group:    ownerRef.Group,
		Version:  ownerRef.Version,
		Resource: res,
	}

	return &resource, nil
}

func calcApiVersion(ownerRef *metadata.TopOwnerReference) string {
	apiVersion := ownerRef.Version
	if len(ownerRef.Group) > 0 {
		apiVersion = ownerRef.Group + "/" + ownerRef.Version
	}
	return apiVersion
}

// getKindResourceName returns the plural name of the resource to allow querying it using the dynamic client.
// This function uses a cache from the api server which will not be updated if new CRDs are added to an existing
// groupVersion.
//
//nolint:nestif // copied from the pod-grouper
func getKindResourceName(ctx context.Context, discoveryClient discovery.DiscoveryInterface,
	kind string, groupVersion string) (string, error) {
	serverResourcesMutex.Lock()
	defer serverResourcesMutex.Unlock()

	logger := log.FromContext(ctx)

	groupResourceList, found := serverResourcesByGroup[groupVersion]
	if !found {
		_, _ = discoveryClient.ServerPreferredResources() // Update the discoveryClient with new CRDs
		_, apiResourceList, err := discoveryClient.ServerGroupsAndResources()
		if err != nil {
			if apiResourceList == nil {
				logger.V(1).Error(err, "Failed to fetch api resource list from server")
				return "", err
			}

			groupDiscoveryError := &discovery.ErrGroupDiscoveryFailed{}
			if errors.As(err, &groupDiscoveryError) {
				logger.V(1).Error(groupDiscoveryError, "error discovering some groups")
			}
		}
		for _, apiResource := range apiResourceList {
			serverResourcesByGroup[apiResource.GroupVersion] = apiResource
		}

		groupResourceList, found = serverResourcesByGroup[groupVersion]
		if !found {
			return "", fmt.Errorf("GroupVersion not supported by server: %s", groupVersion)
		}
	}
	for i := 0; i < len(groupResourceList.APIResources); i++ {
		apiResource := groupResourceList.APIResources[i]
		if apiResource.Kind == kind {
			return apiResource.Name, nil
		}
	}

	err := fmt.Errorf("kind not found in server %s %s", groupVersion, kind)
	return "", err
}
