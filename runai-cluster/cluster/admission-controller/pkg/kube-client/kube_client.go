package kube_client

import (
	"context"
	"fmt"
	"log"

	runaitypes "github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/types"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DefaultLimitRangeConfigMap = "default-limit-range"
	RunaiQueueLabel            = "runai/queue"
	KaiSchemaGroup             = "kai.resources"
	KaiProjectVersion          = "v1alpha1"
	NodePoolVersion            = "v1alpha1"

	projectKind      = "Project"
	nodePoolKind     = "NodePool"
	nodePoolListKind = "NodePoolList"

	configMapNameFieldIndexerName = "metadata.name"
)

type K8sClient struct {
	cachedClient              cache.Cache
	runaiNsScopedCachedClient cache.Cache
}

func NewClient(kubeConfig *rest.Config, scheme *runtime.Scheme) (*K8sClient, error) {
	if kubeConfig == nil || scheme == nil {
		log.Printf("error: nil kubeconfig or scheme")
		return nil, fmt.Errorf("nil kubeconfig or scheme")
	}

	cachedClient, err := createCache(kubeConfig, scheme)
	if err != nil {
		return nil, err
	}

	runaiNsScopedCachedClient, err := createRunaiNsScopedCache(kubeConfig, scheme)
	if err != nil {
		return nil, err
	}

	return &K8sClient{
		cachedClient:              cachedClient,
		runaiNsScopedCachedClient: runaiNsScopedCachedClient,
	}, nil
}

func createCache(kubeConfig *rest.Config, scheme *runtime.Scheme) (cache.Cache, error) {
	nsLabelSelector, err := getRunaiNamespacesSelector()
	if err != nil {
		return nil, err
	}

	// for namespace objects - only save in cache the ones with the runai/queue label selector
	cacheOptions := cache.Options{
		Scheme: scheme,
		ByObject: map[client.Object]cache.ByObject{
			&v1.Namespace{}: {Label: nsLabelSelector},
		},
	}

	cachedClient, err := cache.New(kubeConfig, cacheOptions)
	if err != nil {
		log.Printf("Failed to create cached client, error: %s", err.Error())
		return nil, err
	}

	return cachedClient, nil
}

func createRunaiNsScopedCache(kubeConfig *rest.Config, scheme *runtime.Scheme) (cache.Cache, error) {
	// for configmaps - we only need 1 configmap from the runai namespace.
	// so first of all - use a cache object that only saves objects in runai namespace;
	// and for configmap objects - only save in cache the ones with the appropriate field selector.
	runaiNsScopedCachedClientOptions := cache.Options{
		Scheme: scheme,
		DefaultNamespaces: map[string]cache.Config{
			runaitypes.RunaiNamespace: {},
		},
		ByObject: map[client.Object]cache.ByObject{
			&v1.ConfigMap{}: {Field: getRelevantConfigMapFieldSelector()},
		},
	}

	runaiNsScopedCachedClient, err := cache.New(kubeConfig, runaiNsScopedCachedClientOptions)
	if err != nil {
		log.Printf("Failed to create runai-ns scoped cached client, error: %s", err.Error())
		return nil, err
	}

	return runaiNsScopedCachedClient, nil
}

// StartAndWaitForSync - in case of failure, this function panics, so be careful.
func (c *K8sClient) StartAndWaitForSync(ctx context.Context) error {
	go func() {
		err := c.cachedClient.Start(ctx)
		if err != nil {
			log.Printf("Failed to start cache, error: %s", err.Error())
			panic(fmt.Errorf("failed to start cache, error: %w", err))
		}
	}()
	go func() {
		err := c.runaiNsScopedCachedClient.Start(ctx)
		if err != nil {
			log.Printf("Failed to start runai-ns scoped cache, error: %s", err.Error())
			panic(fmt.Errorf("failed to start runai-ns scoped cache, error: %w", err))
		}
	}()

	if !c.cachedClient.WaitForCacheSync(ctx) {
		log.Printf("Failed waiting for cache to sync")
		return fmt.Errorf("failed waiting for cache to sync")
	}

	if !c.runaiNsScopedCachedClient.WaitForCacheSync(ctx) {
		log.Printf("Failed waiting for runai-ns scoped cache to sync")
		return fmt.Errorf("failed waiting for runai-ns scoped cache to sync")
	}

	return nil
}

func (c *K8sClient) GetNamespace(namespaceName string) (*v1.Namespace, error) {
	namespace := &v1.Namespace{}
	err := c.cachedClient.Get(context.Background(), types.NamespacedName{Name: namespaceName}, namespace)
	if err != nil {
		log.Printf("Failed getting namespace <%s>, error: %s", namespaceName, err.Error())
		return nil, err
	}
	return namespace, nil
}

func (c *K8sClient) GetProject(projectName string) (*unstructured.Unstructured, error) {
	project := &unstructured.Unstructured{}
	project.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   KaiSchemaGroup,
		Version: KaiProjectVersion,
		Kind:    projectKind,
	})
	err := c.cachedClient.Get(context.Background(), types.NamespacedName{Name: projectName}, project)
	if err != nil {
		log.Printf("Failed getting project <%s> gvk %s, error: %s", projectName, project.GroupVersionKind().String(), err.Error())
		return nil, err
	}
	return project, nil
}

func (c *K8sClient) GetNodePool(nodePoolName string) (*unstructured.Unstructured, error) {
	nodePool := &unstructured.Unstructured{}
	nodePool.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   KaiSchemaGroup,
		Version: NodePoolVersion,
		Kind:    nodePoolKind,
	})
	err := c.cachedClient.Get(context.Background(), types.NamespacedName{Name: nodePoolName}, nodePool)
	if err != nil {
		log.Printf("Failed getting nodePool <%s>, error: %s", nodePoolName, err.Error())
		return nil, err
	}
	return nodePool, nil
}

func (c *K8sClient) GetNodePoolsList() (*unstructured.UnstructuredList, error) {
	nodePoolList := &unstructured.UnstructuredList{}
	nodePoolList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   KaiSchemaGroup,
		Version: NodePoolVersion,
		Kind:    nodePoolListKind,
	})
	err := c.cachedClient.List(context.Background(), nodePoolList)
	if err != nil {
		log.Printf("Failed listing nodePools, error: %s", err.Error())
		return nil, err
	}
	return nodePoolList, nil
}

func (c *K8sClient) GetRunaiConfigMap() (*v1.ConfigMap, error) {
	configMap := &v1.ConfigMap{}
	err := c.runaiNsScopedCachedClient.Get(context.Background(),
		types.NamespacedName{Name: DefaultLimitRangeConfigMap, Namespace: runaitypes.RunaiNamespace},
		configMap)
	if err != nil {
		log.Printf("Failed getting configMap <%s/%s>, error: %s",
			runaitypes.RunaiNamespace, DefaultLimitRangeConfigMap, err.Error())
		return nil, err
	}
	return configMap, nil
}

func getRunaiNamespacesSelector() (labels.Selector, error) {
	requirement, err := labels.NewRequirement(RunaiQueueLabel, selection.Exists, []string{})
	if err != nil {
		log.Printf("Failed to create runai queue label requirement, error: %s", err.Error())
		return nil, err
	}

	labelSelector := labels.NewSelector()
	labelSelector = labelSelector.Add(*requirement)
	return labelSelector, nil
}

func (c *K8sClient) GetCachedReader() client.Reader {
	return c.cachedClient
}

func getRelevantConfigMapFieldSelector() fields.Selector {
	return fields.OneTermEqualSelector(configMapNameFieldIndexerName, DefaultLimitRangeConfigMap)
}
