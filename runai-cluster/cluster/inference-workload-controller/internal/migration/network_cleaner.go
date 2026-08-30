package migration

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"reflect"
	"strings"

	openShiftV1 "github.com/openshift/api/route/v1"
	commonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	prevName = "exposed-url"
)

type NetworkCleaner struct {
	k8sClient               client.Client
	getListObj              func() client.ObjectList
	ConvertObjectListToList func(client.ObjectList) ([]client.Object, error)
}

func NewNetworkCleaner(k8sClient client.Client, isOpenShift bool) *NetworkCleaner {
	listFactoryFn := getIngressListObjectList
	if isOpenShift {
		listFactoryFn = getRouteListObjectList
	}
	return &NetworkCleaner{
		k8sClient:               k8sClient,
		getListObj:              listFactoryFn,
		ConvertObjectListToList: convertListObjectToList,
	}

}

func getIngressListObjectList() client.ObjectList {
	return &networkingv1.IngressList{}
}
func getRouteListObjectList() client.ObjectList {
	return &openShiftV1.RouteList{}
}

func convertListObjectToList(listObj client.ObjectList) ([]client.Object, error) {
	var list []client.Object
	switch typedList := listObj.(type) {
	case *networkingv1.IngressList:
		for _, item := range typedList.Items {
			list = append(list, item.DeepCopy())
		}
	case *openShiftV1.RouteList:
		for _, item := range typedList.Items {
			list = append(list, item.DeepCopy())

		}
	case *corev1.ServiceList:
		for _, item := range typedList.Items {
			list = append(list, item.DeepCopy())
		}

	default:
		// This should never happen
		return nil, fmt.Errorf("unsupported list object type %s", reflect.TypeOf(listObj).String())

	}
	return list, nil
}

func (nc *NetworkCleaner) RequiresWebhooks() bool {
	return false
}

func (nc *NetworkCleaner) Migrate(ctx context.Context) error {
	namespaceList := &corev1.NamespaceList{}
	var continueToken string
	for {

		if ctx.Err() != nil {
			return fmt.Errorf("context is cancelled; %w", ctx.Err())
		}
		pagination := client.ListOptions{
			Limit:    paginationLimit,
			Continue: continueToken,
		}
		err := nc.k8sClient.List(ctx, namespaceList, &pagination, client.HasLabels{commonconstants.RunaiQueueLabel})
		if err != nil {
			return fmt.Errorf("unable to list namespaces with runai label; %w", err)
		}
		for _, ns := range namespaceList.Items {
			nsCP := ns
			errGroup, groupCtx := errgroup.WithContext(ctx)
			errGroup.Go(func() error {
				return nc.removeExternalAccessInNamespace(groupCtx, nsCP.Name)
			})
			errGroup.Go(func() error {
				return nc.removeServicesInNamespace(groupCtx, nsCP.Name)
			})
			err = errGroup.Wait()
			if err != nil {
				return err
			}

		}

		continueToken = namespaceList.Continue
		if continueToken == "" {
			break
		}
	}

	return nil
}

func (nc *NetworkCleaner) removeServicesInNamespace(ctx context.Context, namespace string) error {
	log.Info("check services in namespace", "namespace", namespace)
	serviceList := &corev1.ServiceList{}
	return nc.deleteResources(ctx, namespace, serviceList)
}

func (nc *NetworkCleaner) removeExternalAccessInNamespace(ctx context.Context, ns string) error {
	log.Info("check external access in namespace", "namespace", ns)
	listNetworkExternalAccess := nc.getListObj()

	return nc.deleteResources(ctx, ns, listNetworkExternalAccess)
}

func (nc *NetworkCleaner) deleteResources(ctx context.Context, ns string, networkResources client.ObjectList) error {
	var continueToken string
	var err error
	for {
		if ctx.Err() != nil {
			return fmt.Errorf("context is cancelled; %w", ctx.Err())
		}

		continueToken, err = nc.handleBatch(ctx, ns, continueToken, networkResources)
		if err != nil {
			return err
		}
		if continueToken == "" {
			break
		}

	}
	return nil
}

func (nc *NetworkCleaner) handleBatch(ctx context.Context, ns string, continueToken string, networkResources client.ObjectList) (string, error) {
	pagination := client.ListOptions{
		Limit:    paginationLimit,
		Continue: continueToken,
	}

	selector, err := LabelsSelector()
	if err != nil {
		return "", fmt.Errorf("unable to get resources to delete; %w", err)
	}

	listOptions := []client.ListOption{client.InNamespace(ns), &pagination,
		client.MatchingLabelsSelector{Selector: selector}}

	err = nc.k8sClient.List(ctx, networkResources, listOptions...)
	if err != nil {
		return "", fmt.Errorf("unable to list ingresses in namespace %s; %w", ns, err)
	}

	continueToken = networkResources.GetContinue()
	networkExternalAccessList, err := nc.ConvertObjectListToList(networkResources)
	if err != nil {
		return "", err
	}

	filteredList := filterList(networkExternalAccessList)
	for _, externalAccessObj := range filteredList {
		externalAccessObjCP := externalAccessObj
		log.Info("deleting external access", "name", externalAccessObjCP.GetName(),
			"kind", externalAccessObjCP.GetObjectKind().GroupVersionKind().Kind)
		err = nc.k8sClient.Delete(ctx, externalAccessObjCP)
		if err != nil {
			return "", fmt.Errorf("unable to delete external access %s; %w", externalAccessObjCP.GetName(), err)
		}

	}

	return continueToken, nil
}

func LabelsSelector() (labels.Selector, error) {
	selector := labels.NewSelector()
	hasAppLabel, err := labels.NewRequirement(commonconstants.LabelApp, selection.Exists, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create label requirement; %w", err)
	}
	selector = selector.Add(*hasAppLabel)
	return selector, nil
}

func filterList(ingressList []client.Object) []client.Object {
	var filteredList []client.Object
	for _, ingress := range ingressList {
		if isOwnedByWorkload(ingress) {
			filteredList = append(filteredList, ingress)
		}
	}
	return filteredList
}

func (nc *NetworkCleaner) String() string {
	return NetworkCleanerMigrationName
}

func isOwnedByWorkload(externalUrl client.Object) bool {
	if externalUrl.GetOwnerReferences() == nil {
		return isBackwardOwnership(externalUrl)
	}
	if len(externalUrl.GetOwnerReferences()) > 1 {
		return false
	}
	ownerKind := externalUrl.GetOwnerReferences()[0].Kind
	return ownerKind == commonconstants.KindInferenceWorkload
}

func isBackwardOwnership(ingress client.Object) bool {
	if strings.Contains(ingress.GetName(), prevName) {
		return true
	}
	if _, ok := ingress.GetLabels()[commonconstants.LabelWorkloadKind]; ok {
		return true
	}
	if _, ok := ingress.GetLabels()[commonconstants.LabelWorkloadName]; ok {
		return true
	}
	return false
}
