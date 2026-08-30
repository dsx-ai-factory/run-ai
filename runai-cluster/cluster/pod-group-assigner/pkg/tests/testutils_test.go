// Test-only resource helpers for the envtest suites in this package.
//
// These are local copies of the three helpers previously taken from the
// go-operator's tests/utils package; duplicating them keeps this service free of
// the go-operator module. They live in a _test.go file so the ginkgo/gomega
// dependency stays out of the module's production build graph.
package tests

import (
	"context"
	"fmt"
	"reflect"
	"time"

	. "github.com/onsi/ginkgo/v2" //nolint:staticcheck // dot import for test framework is intentional
	. "github.com/onsi/gomega"    //nolint:staticcheck // dot import for test framework is intentional
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// timeoutUnit bounds the polling helpers below.
const timeoutUnit = 240 * time.Second

var apiContext = context.TODO()

// resourceUniqueID renders a resource as "[namespace/Type/name]" for test output.
func resourceUniqueID(resource client.Object) string {
	return fmt.Sprintf("[%s/%s/%s]",
		resource.GetNamespace(),
		reflect.TypeOf(resource).Elem().Name(),
		resource.GetName())
}

func namespacedNameFor(resource client.Object) types.NamespacedName {
	return types.NamespacedName{
		Name:      resource.GetName(),
		Namespace: resource.GetNamespace(),
	}
}

// expectCreateResource creates the resource and fails the spec if creation fails.
func expectCreateResource(k8sClient client.Client, resource client.Object) {
	By(fmt.Sprintf("creating resource %s...", resourceUniqueID(resource)))

	ExpectWithOffset(1, k8sClient.Create(apiContext, resource)).To(Succeed())

	By(fmt.Sprintf("successfully created resource %s", resourceUniqueID(resource)))
}

// deleteImmediatelyAndPollUntilDeleted deletes with a zero grace period and waits
// until the resource is gone.
func deleteImmediatelyAndPollUntilDeleted(k8sClient client.Client, resource client.Object, intervals ...time.Duration) {
	deleteAndPollUntilDeletedInner(k8sClient, resource, true, intervals...)
}

// deleteAndPollUntilDeleted deletes the resource and waits until it is gone.
func deleteAndPollUntilDeleted(k8sClient client.Client, resource client.Object, intervals ...time.Duration) {
	deleteAndPollUntilDeletedInner(k8sClient, resource, false, intervals...)
}

func deleteAndPollUntilDeletedInner(
	k8sClient client.Client, resource client.Object, immediately bool, intervals ...time.Duration,
) {
	By(fmt.Sprintf("waiting for %s to be deleted...", resourceUniqueID(resource)))

	deleteOptions := &client.DeleteOptions{}
	if immediately {
		var gracePeriodSeconds int64
		deleteOptions.GracePeriodSeconds = &gracePeriodSeconds
	}

	deleteFn := func() error {
		err := k8sClient.Delete(apiContext, resource, deleteOptions)
		if err == nil || errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	ExpectWithOffset(1, deleteFn()).To(Succeed())

	gone := func() bool {
		return errors.IsNotFound(k8sClient.Get(apiContext, namespacedNameFor(resource), resource))
	}

	timeout := timeoutUnit
	if len(intervals) > 0 {
		timeout = intervals[0]
	}
	EventuallyWithOffset(1, gone, timeout).Should(BeTrue())

	By(fmt.Sprintf("successfully deleted resource %s", resourceUniqueID(resource)))
}
