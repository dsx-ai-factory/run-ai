package testing_utils

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/utils"
	controller_utils "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/controller-utils"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func hasTrueConditionType(runaiConfig *runaiv1.RunaiConfig,
	conditionType runaiv1.ConditionType) bool {
	conditions := runaiConfig.Status.Conditions
	for i := range conditions {
		if conditions[i].Type == string(conditionType) &&
			conditions[i].ObservedGeneration == runaiConfig.Generation &&
			conditions[i].Status == metav1.ConditionTrue {
			return true
		}
	}

	return false
}

func hasConditionType(runaiConfig *runaiv1.RunaiConfig,
	conditionType runaiv1.ConditionType) bool {
	conditions := runaiConfig.Status.Conditions
	for i := range conditions {
		if conditions[i].Type == string(conditionType) &&
			conditions[i].ObservedGeneration == runaiConfig.Generation {
			return true
		}
	}

	return false
}

func WaitForOperatorToObserve(k8s client.Client, runaiConfig *runaiv1.RunaiConfig) {
	generation := runaiConfig.Generation
	By("waiting for operator to observe RunaiConfig")
	EventuallyWithTimeout(func() bool {
		newRunaiConfig := GetTestingRunaiConfig(k8s)
		return newRunaiConfig.Status.ObservedGeneration == generation
	}).Should(BeTrue())
}

func PollForCondition(k8sClient client.Client,
	conditionType runaiv1.ConditionType, runaiConfig *runaiv1.RunaiConfig) {
	generation := runaiConfig.Generation
	EventuallyWithOffset(1, func() bool {
		updatedRunaiConfig := GetTestingRunaiConfig(k8sClient)
		*runaiConfig = *updatedRunaiConfig
		return hasConditionType(runaiConfig, conditionType) && runaiConfig.Generation == generation
	}, TimeoutUnit).WithPolling(PollingUnit).Should(BeTrue())

	By("runaiConfig reached true condition")
}

func PollForTrueCondition(k8sClient client.Client,
	conditionType runaiv1.ConditionType, runaiConfig *runaiv1.RunaiConfig) {
	generation := runaiConfig.Generation
	EventuallyWithOffset(1, func() bool {
		updatedRunaiConfig := GetTestingRunaiConfig(k8sClient)
		*runaiConfig = *updatedRunaiConfig
		return hasTrueConditionType(runaiConfig, conditionType) && runaiConfig.Generation == generation
	}, TimeoutUnit).WithPolling(PollingUnit).Should(BeTrue())

	By("runaiConfig reached true condition")
}

func PollForFalseCondition(k8sClient client.Client,
	conditionType runaiv1.ConditionType, runaiConfig *runaiv1.RunaiConfig) {
	EventuallyWithOffset(1, func() bool {
		runaiConfig = GetTestingRunaiConfig(k8sClient)
		return !hasTrueConditionType(runaiConfig, conditionType)
	}, TimeoutUnit).WithPolling(PollingUnit).Should(BeTrue())
}

func NamespacedNameFor(resource client.Object) types.NamespacedName {
	return types.NamespacedName{
		Name:      resource.GetName(),
		Namespace: resource.GetNamespace(),
	}
}

func WaitForDeployedCondition(k8sClient client.Client, runaiConfig *runaiv1.RunaiConfig) {
	EventuallyWithOffset(1, func() bool {
		runaiConfig = GetTestingRunaiConfig(k8sClient)
		return hasTrueConditionType(runaiConfig, runaiv1.ConditionTypeDeployed)
	}, TimeoutUnit).Should(BeTrue())
}

func TestingRunaiConfig() *runaiv1.RunaiConfig {
	runaiConfig := &runaiv1.RunaiConfig{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: TestingNamespace,
			Name:      TestingRunaiConfigName,
		},
	}

	runaiConfig.SetDefaultsWhereNeeded()

	return runaiConfig
}

func CreateTestingRunaiConfig(k8sClient client.Client) *runaiv1.RunaiConfig {
	runaiConfig := TestingRunaiConfig()
	ExpectCreateResource(k8sClient, runaiConfig)

	return runaiConfig
}

func GetTestingRunaiConfig(k8sClient client.Client) *runaiv1.RunaiConfig {
	runaiConfig := TestingRunaiConfig()
	ExpectGetResource(k8sClient, runaiConfig)

	return runaiConfig
}

func DeleteTestingRunaiConfig(k8sClient client.Client) {
	DeleteAndPollUntilDeleted(k8sClient, TestingRunaiConfig())
}

func PollUntilDeleted(k8sClient client.Client, resource client.Object) {
	By(fmt.Sprintf("waiting for %s to be deleted...", utils.ResourceUniqueID(resource)))

	EventuallyWithOffset(1, func() bool {
		err := k8sClient.Get(globalAPIContext, NamespacedNameFor(resource), resource)
		return errors.IsNotFound(err)
	}, TimeoutUnit).Should(BeTrue())

	By(fmt.Sprintf("resource %s was deleted", utils.ResourceUniqueID(resource)))
}

func DeleteImmediatelyAndPollUntilDeleted(k8sClient client.Client, resource client.Object, intervals ...time.Duration) {
	deleteAndPollUntilDeletedInner(k8sClient, resource, true, intervals...)
}

func DeleteAndPollUntilDeleted(k8sClient client.Client, resource client.Object, intervals ...time.Duration) {
	deleteAndPollUntilDeletedInner(k8sClient, resource, false, intervals...)
}

func deleteAndPollUntilDeletedInner(k8sClient client.Client, resource client.Object, immediately bool, intervals ...time.Duration) {
	By(fmt.Sprintf("waiting for %s to be deleted...", utils.ResourceUniqueID(resource)))

	// deletedResourceUID := resource.GetUID()

	deleteOptions := &client.DeleteOptions{}
	var gracePeriodSeconds int64 = 0
	if immediately {
		deleteOptions.GracePeriodSeconds = &gracePeriodSeconds
	}

	deleteFn := func() error {
		err := k8sClient.Delete(globalAPIContext, resource, deleteOptions)
		if err == nil || errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	ExpectWithOffset(1, deleteFn()).To(Succeed())

	actual := func() bool {
		err := k8sClient.Get(globalAPIContext, NamespacedNameFor(resource), resource)
		// return errors.IsNotFound(err) || resource.GetUID() != deletedResourceUID
		return errors.IsNotFound(err)
	}

	if len(intervals) == 0 {
		EventuallyWithOffset(1, actual, TimeoutUnit).Should(BeTrue())
	} else {
		EventuallyWithOffset(1, actual, intervals[0]).Should(BeTrue())
	}

	By(fmt.Sprintf("succesfully deleted resource %s", utils.ResourceUniqueID(resource)))
}

func ExpectGetResource(k8sClient client.Client, resource client.Object) {
	By(fmt.Sprintf("getting resource %s...", utils.ResourceUniqueID(resource)))

	ExpectWithOffset(1, k8sClient.Get(globalAPIContext, NamespacedNameFor(resource), resource)).To(Succeed())

	By(fmt.Sprintf("succesfully got resource %s", utils.ResourceUniqueID(resource)))
}

func ExpectCreateResource(k8sClient client.Client, resource client.Object) {
	By(fmt.Sprintf("creating resource %s...", utils.ResourceUniqueID(resource)))

	ExpectWithOffset(1, k8sClient.Create(globalAPIContext, resource)).To(Succeed())

	By(fmt.Sprintf("succesfully created resource %s", utils.ResourceUniqueID(resource)))
}

func ExpectDeleteResource(k8sClient client.Client, resource client.Object) {
	By(fmt.Sprintf("deleting resource %s...", utils.ResourceUniqueID(resource)))

	ExpectWithOffset(1, k8sClient.Delete(globalAPIContext, resource)).To(Succeed())

	By(fmt.Sprintf("succesfully deleted resource %s", utils.ResourceUniqueID(resource)))
}

func ExpectUpdateResource(k8sClient client.Client, resource client.Object) {
	By(fmt.Sprintf("updating resource %s...", utils.ResourceUniqueID(resource)))

	ExpectWithOffset(1, k8sClient.Update(globalAPIContext, resource)).To(Succeed())

	By(fmt.Sprintf("succesfully updated resource %s", utils.ResourceUniqueID(resource)))
}

func EventuallyUpdateResource(k8sClient client.Client, resource client.Object) {
	By(fmt.Sprintf("eventually updating resource %s...", utils.ResourceUniqueID(resource)))

	EventuallyWithOffset(1, func() error {
		By(fmt.Sprintf("attempting to update resource %s", resource.GetName()))
		tmpResource := controller_utils.NewResource(resource)
		tmpResource.SetName(resource.GetName())
		tmpResource.SetNamespace(resource.GetNamespace())

		err := k8sClient.Get(globalAPIContext, NamespacedNameFor(tmpResource), tmpResource)
		ExpectNoErr(err)

		resource.SetResourceVersion(tmpResource.GetResourceVersion())
		resource.SetGeneration(tmpResource.GetGeneration())

		return k8sClient.Update(globalAPIContext, resource)
	}, TimeoutUnit).Should(Succeed())

	By(fmt.Sprintf("succesfully updated resource %s", utils.ResourceUniqueID(resource)))
}

func ExpectCreateOrUpdateResource(k8sClient client.Client, resource client.Object) {
	By(fmt.Sprintf("updating resource %s...", utils.ResourceUniqueID(resource)))

	err := k8sClient.Update(globalAPIContext, resource)
	if err != nil {
		if errors.IsNotFound(err) {
			ExpectCreateResource(k8sClient, resource)
		} else {
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
		}
	}

	By(fmt.Sprintf("succesfully updated resource %s", utils.ResourceUniqueID(resource)))
}

func ExpectDoesNotExist(k8sClient client.Client, resource client.Object) {
	By(fmt.Sprintf("expecting resource %s to not exist...", utils.ResourceUniqueID(resource)))

	err := k8sClient.Get(globalAPIContext, NamespacedNameFor(resource), resource)

	ExpectWithOffset(1, err).To(HaveOccurred())
	ExpectWithOffset(1, errors.IsNotFound(err)).To(BeTrue())

	By(fmt.Sprintf("resource %s does not exist", utils.ResourceUniqueID(resource)))
}

func TriggerController(k8sClient client.Client, resource client.Object) {
	annotation := "run.ai/trigger"
	annotations := resource.GetAnnotations()

	if annotations == nil {
		annotations = map[string]string{}
	}

	annotations[annotation] = uuid.New().String()

	resource.SetAnnotations(annotations)

	EventuallyUpdateResource(k8sClient, resource)
}

func ExpectNoErr(err error) {
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
}

func ExpectErr(err error) {
	ExpectWithOffset(1, err).To(HaveOccurred())
}

func ExpectSuccess(err error) {
	ExpectWithOffset(1, err).To(Succeed())
}

func ResourceExists(k8sClient client.Client, resource client.Object) bool {
	err := k8sClient.Get(globalAPIContext, NamespacedNameFor(resource), resource)
	if err != nil {
		if errors.IsNotFound(err) {
			return false
		} else {
			ExpectNoErr(err)
		}
	}

	return true
}

func EventuallyWithTimeout(actual interface{}) AsyncAssertion {
	return EventuallyWithOffset(1, actual, TimeoutUnit)
}
