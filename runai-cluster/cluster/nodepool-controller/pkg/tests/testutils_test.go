package tests

import (
	"context"
	"fmt"
	"reflect"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Local replacements for go-operator/tests/utils, so this module's tests do not
// depend on another service's test helpers.

var (
	globalAPIContext = context.TODO()
	timeoutUnit      = 240 * time.Second
)

// clusterCRGVK mirrors the GVK the controller uses to read the run.ai Cluster CR
// as unstructured. Kept as strings so the tests carry no run.ai API dependency.
var clusterCRGVK = schema.GroupVersionKind{Group: "run.ai", Version: "v1", Kind: "Cluster"}

const (
	clusterCRName     = "cluster"
	runaiNamespace    = "runai"
	clusterCRListKind = "ClusterList"
)

// registerClusterCRType teaches the test scheme about the run.ai Cluster GVK so
// the fake client can serve it as unstructured.
func registerClusterCRType(s *runtime.Scheme) {
	s.AddKnownTypeWithName(clusterCRGVK, &unstructured.Unstructured{})
	s.AddKnownTypeWithName(clusterCRGVK.GroupVersion().WithKind(clusterCRListKind), &unstructured.UnstructuredList{})
}

// newClusterCR builds the run.ai Cluster fixture. Its presence (with no deletion
// timestamp) is what tells the controller run:ai is not being uninstalled.
func newClusterCR() *unstructured.Unstructured {
	clusterCR := &unstructured.Unstructured{}
	clusterCR.SetGroupVersionKind(clusterCRGVK)
	clusterCR.SetName(clusterCRName)
	return clusterCR
}

func resourceStructTypeName(resource client.Object) string {
	return reflect.TypeOf(resource).Elem().Name()
}

func resourceUniqueID(resource client.Object) string {
	return fmt.Sprintf("[%s/%s/%s]",
		resource.GetNamespace(), resourceStructTypeName(resource), resource.GetName())
}

func namespacedNameFor(resource client.Object) types.NamespacedName {
	return types.NamespacedName{Name: resource.GetName(), Namespace: resource.GetNamespace()}
}

func ExpectNoErr(err error) {
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
}

func ResourceExists(k8sClient client.Client, resource client.Object) bool {
	err := k8sClient.Get(globalAPIContext, namespacedNameFor(resource), resource)
	if err != nil {
		if errors.IsNotFound(err) {
			return false
		}
		ExpectNoErr(err)
	}

	return true
}

func ExpectGetResource(k8sClient client.Client, resource client.Object) {
	By(fmt.Sprintf("getting resource %s...", resourceUniqueID(resource)))

	ExpectWithOffset(1, k8sClient.Get(globalAPIContext, namespacedNameFor(resource), resource)).To(Succeed())
}

func ExpectCreateResource(k8sClient client.Client, resource client.Object) {
	By(fmt.Sprintf("creating resource %s...", resourceUniqueID(resource)))

	ExpectWithOffset(1, k8sClient.Create(globalAPIContext, resource)).To(Succeed())
}

func EventuallyUpdateResource(k8sClient client.Client, resource client.Object) {
	By(fmt.Sprintf("eventually updating resource %s...", resourceUniqueID(resource)))

	EventuallyWithOffset(1, func() error {
		tmpResource := reflect.New(reflect.TypeOf(resource).Elem()).Interface().(client.Object)
		tmpResource.SetName(resource.GetName())
		tmpResource.SetNamespace(resource.GetNamespace())

		ExpectNoErr(k8sClient.Get(globalAPIContext, namespacedNameFor(tmpResource), tmpResource))

		resource.SetResourceVersion(tmpResource.GetResourceVersion())
		resource.SetGeneration(tmpResource.GetGeneration())

		return k8sClient.Update(globalAPIContext, resource)
	}, timeoutUnit).Should(Succeed())
}

func DeleteImmediatelyAndPollUntilDeleted(k8sClient client.Client, resource client.Object, intervals ...time.Duration) {
	By(fmt.Sprintf("waiting for %s to be deleted...", resourceUniqueID(resource)))

	var gracePeriodSeconds int64 = 0
	deleteOptions := &client.DeleteOptions{GracePeriodSeconds: &gracePeriodSeconds}

	deleteFn := func() error {
		err := k8sClient.Delete(globalAPIContext, resource, deleteOptions)
		if err == nil || errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	ExpectWithOffset(1, deleteFn()).To(Succeed())

	actual := func() bool {
		err := k8sClient.Get(globalAPIContext, namespacedNameFor(resource), resource)
		return errors.IsNotFound(err)
	}

	timeout := timeoutUnit
	if len(intervals) > 0 {
		timeout = intervals[0]
	}
	EventuallyWithOffset(1, actual, timeout).Should(BeTrue())
}
