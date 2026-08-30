package utils

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"reflect"
	"time"

	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	RateLimiterBaseDelay = time.Millisecond * 500
	RateLimiterMaxDelay  = time.Second * 60
)

type Reconciler interface {
	SetupWithManager(mgr ctrl.Manager) error
	Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error)
}

func GetObjectForRequest(ctx context.Context, req ctrl.Request, kind string,
	object client.Object, reader client.Reader) (isObjectFound bool, err error) {
	err = reader.Get(ctx, req.NamespacedName, object)
	if err == nil {
		return true, nil
	}

	if errors.IsNotFound(err) {
		log.Ctx(ctx).Warn().Msgf("Can't find %s <%s> in cluster, this might also mean it was deleted",
			kind, getNameStr(req))
		// Nothing really more to do.
		return false, nil
	}

	log.Ctx(ctx).Error().Msgf("Failed to get %s <%s> from incoming request, error: %s",
		kind, getNameStr(req), err.Error())
	return false, err
}

func getNameStr(req ctrl.Request) string {
	if req.Namespace == "" {
		return req.Name
	}

	return fmt.Sprintf("%s/%s", req.Namespace, req.Name)
}

func IsItemInList[T any](list []T, item T) bool {
	return indexOfFirstItemInList(list, item) > -1
}

func DeleteFirstFromList[T any](list []T, itemToRemove T) []T {
	indexToRemove := indexOfFirstItemInList(list, itemToRemove)

	if indexToRemove >= len(list) || indexToRemove == -1 {
		return list
	}
	if indexToRemove < len(list)-1 {
		copy(list[indexToRemove:], list[indexToRemove+1:]) // Shift a[i+1:] left one index.
	}

	list = list[:len(list)-1] // Truncate slice.
	return list
}

func indexOfFirstItemInList[T any](list []T, item T) int {
	for i, element := range list {
		if reflect.DeepEqual(element, item) {
			return i
		}
	}
	return -1
}

func ValidateNamespaceExists(ctx context.Context, k8sClient client.Client, namespaceName string, project string) error {
	var namespace corev1.Namespace
	err := k8sClient.Get(ctx, client.ObjectKey{Name: namespaceName}, &namespace)
	if err != nil {
		log.Ctx(ctx).Info().Msgf("Could not handle cluster-wide resources in project %s. Failed to get namespace %s. %s",
			project, namespaceName, err.Error())
	}
	return err
}
