/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/SemanticallyNull/golandreporter"
	. "github.com/onsi/ginkgo" //nolint:staticcheck // dot import for test framework is intentional
	"github.com/onsi/ginkgo/reporters"
	//nolint:staticcheck // dot import for test framework is intentional
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/handlers"
	kaiv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
)

// externalWorkloadAPIVersion is the group/version of the run.ai ExternalWorkload
// CRD. project-controller only ever counts these objects when deciding whether a
// project still has live work in its namespace, so the fixture is unstructured
// and matches how the deletion blockers actually read them on a cluster.
const externalWorkloadAPIVersion = "run.ai/v1alpha1"
const ProjectScopeLabel = "run.ai/project"
const ClusterWideResourceLabel = "run.ai/cluster-wide"
const RunaiResourceManualOverrideLabel = "run.ai/resource-override"
const RunaiProjectLabel = "project"
const RunaiQueueLabel = "runai/queue"
const QueueDepartmentNameLabel = "runai/department-name"
const RunaiLimitRangeName = "runai-limit-range"
const EnforceSchedulerAnnotationName = "runai/enforce-scheduler-name"

// externalWorkloadFixture builds a single-item ExternalWorkload list in the test
// namespace, shaped exactly as the deletion blockers see it.
func externalWorkloadFixture(name string) unstructured.UnstructuredList {
	item := unstructured.Unstructured{}
	item.SetAPIVersion(externalWorkloadAPIVersion)
	item.SetKind("ExternalWorkload")
	item.SetName(name)
	item.SetNamespace(TestNamespace.Name)

	list := unstructured.UnstructuredList{}
	list.SetAPIVersion(externalWorkloadAPIVersion)
	list.SetKind("ExternalWorkloadList")
	list.Items = []unstructured.Unstructured{item}
	return list
}

// Test data uses the runai-flavored vocabulary. We keep these as named
// consts (rather than inline string literals) so package-init time has
// stable values for the test-object factories below.
const (
	testRunaiNamespace             = "runai"
	testRunaiProjectIdLabel        = "run.ai/project-id"
	testRunaiNamespaceVersionLabel = "runai/namespace-version"
	DepartmentIdLabel              = "runai/department-id"
	DepartmentNameLabel            = "runai/department-name"
)

const (
	ConfigMapKind  = "ConfigMap"
	CircleCiEnvVar = "CIRCLECI"
	// NodePoolLabelKey is deliberately a distinct sentinel (not any production node-pool
	// label const) so that write and read paths must both go through the configured
	// NodePoolLabelKey — a hardcoded key on either side would fail these tests.
	NodePoolLabelKey = "example.com/node-pool"
	// DefaultNodePoolName is a distinct sentinel (not "default") so queue naming/label logic must
	// read it from config.Get().DefaultNodepoolName — a hardcoded "default" on either side fails.
	DefaultNodePoolName = "example-default-nodepool"
	TestDepartment1Id   = "1"
	TestDepartment2Id   = "2"
	TestDepartment3Id   = "3"
	TestDepartment1Name = "some-dep"
	TestDepartment2Name = "some-dep2"
	TestDepartment3Name = "some-dep3"
	SomeNodePoolName    = "nodepoola"

	JobControllerRoleBindingName        = "runai-job-controller-project"
	ProjectControllerServiceAccountName = "runai-project-controller"

	pvcKind = "PersistentVolumeClaim"
)

//goland:noinspection GoNameStartsWithPackageName
var (
	// Populate the package-level config before the fixture vars below are initialized.
	// Several fixtures compute queue names via config-dependent helpers (e.g.
	// GetQueueNameOfDepartment / GetQueueName read config.Get().DefaultNodepoolName), which
	// would nil-deref an unset config. This runs first (earliest declaration, no dependency on
	// the fixtures); individual suites re-set config in their BeforeSuite/BeforeEach.
	_ = config.SetForTest(RunaiConfigForTests())

	TestProject = kaiv1alpha1.Project{
		TypeMeta: metav1.TypeMeta{
			Kind:       common.ProjectKind,
			APIVersion: kaiv1alpha1.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "proj-1",
			UID:  "123-ani-ahash-ve-rosh",
			Labels: map[string]string{
				DepartmentIdLabel: TestDepartment1Id,
			},
		},
		Spec: kaiv1alpha1.ProjectSpec{
			Parent:              TestDepartment1Name,
			EnforceKaiScheduler: true,
			Queues: []kaiv1alpha1.QueueConfig{
				{
					Name:     "proj-1",
					Nodepool: DefaultNodePoolName,
					Resources: &kaiv1alpha1.QueueResourcesConfig{
						GPU: kaiv1alpha1.SystemResource{
							Deserved: 5,
						},
						CPU: kaiv1alpha1.SystemResource{
							Deserved: 10000,
						},
					},
				},
				{
					Name:     "proj-1-nodepoola",
					Nodepool: SomeNodePoolName,
					Resources: &kaiv1alpha1.QueueResourcesConfig{
						GPU: kaiv1alpha1.SystemResource{
							Deserved: 10,
						},
						CPU: kaiv1alpha1.SystemResource{
							Deserved: 10000,
						},
					},
				},
			},
		},
		Status: kaiv1alpha1.ProjectStatus{},
	}

	KaiModifiedTestProject = kaiv1alpha1.Project{
		TypeMeta: metav1.TypeMeta{
			Kind:       common.ProjectKind,
			APIVersion: kaiv1alpha1.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: TestProject.Name,
			UID:  TestProject.UID,
			Labels: map[string]string{
				DepartmentIdLabel: TestDepartment1Id,
			},
		},
		Spec: kaiv1alpha1.ProjectSpec{
			Parent:              TestDepartment1Name,
			EnforceKaiScheduler: false,
			Queues: []kaiv1alpha1.QueueConfig{
				{
					Name:     TestProject.Spec.Queues[0].Name,
					Nodepool: DefaultNodePoolName,
					Resources: &kaiv1alpha1.QueueResourcesConfig{
						GPU: kaiv1alpha1.SystemResource{
							Deserved: 10,
						},
						CPU: kaiv1alpha1.SystemResource{
							Deserved: 10000,
						},
					},
				},
				{
					Name:     TestProject.Spec.Queues[1].Name,
					Nodepool: SomeNodePoolName,
					Resources: &kaiv1alpha1.QueueResourcesConfig{
						GPU: kaiv1alpha1.SystemResource{
							Deserved: 5,
						},
						CPU: kaiv1alpha1.SystemResource{
							Deserved: 10000,
						},
					},
				},
			},
		},
		Status: kaiv1alpha1.ProjectStatus{},
	}

	KaiAnotherTestProject = kaiv1alpha1.Project{
		TypeMeta: metav1.TypeMeta{
			Kind:       common.ProjectKind,
			APIVersion: kaiv1alpha1.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "proj-555",
			UID:  "another-uid",
			Labels: map[string]string{
				DepartmentIdLabel: TestDepartment2Id,
			},
		},
		Spec: kaiv1alpha1.ProjectSpec{
			Parent:              "new-new-test-dep",
			EnforceKaiScheduler: false,
			Queues: []kaiv1alpha1.QueueConfig{
				{
					Name:     "proj-555",
					Nodepool: DefaultNodePoolName,
					Resources: &kaiv1alpha1.QueueResourcesConfig{
						GPU: kaiv1alpha1.SystemResource{
							Deserved: 2,
						},
						CPU: kaiv1alpha1.SystemResource{
							Deserved: 5000,
						},
						Memory: kaiv1alpha1.SystemResource{
							Deserved: 2048,
						},
					},
				},
			},
		},
		Status: kaiv1alpha1.ProjectStatus{},
	}

	TestQueue = kaiv2.Queue{
		TypeMeta: metav1.TypeMeta{
			APIVersion: kaiv2.GroupVersion.Identifier(),
			Kind:       common.QueueKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            TestProject.Spec.Queues[0].Name,
			OwnerReferences: ProjectOwnerRef,
			Labels: map[string]string{
				RunaiProjectLabel:       TestProject.Name,
				testRunaiProjectIdLabel: string(TestProject.UID),
				DepartmentNameLabel:     TestProject.Spec.Parent,
			},
		},
		Spec: kaiv2.QueueSpec{
			ParentQueue: TestProject.Spec.Parent,
			Resources: &kaiv2.QueueResources{
				GPU: kaiv2.QueueResource{
					Quota: 5,
				},
				CPU: kaiv2.QueueResource{
					Quota: 10000,
				},
			},
		},
	}

	TestNamespace = corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: common.CoreV1ApiVersion,
			Kind:       common.NamespaceKind,
		}, ObjectMeta: metav1.ObjectMeta{
			Name: "ns-proj-1",
			Labels: map[string]string{
				RunaiQueueLabel:                TestProject.Name,
				testRunaiNamespaceVersionLabel: common.AgentNamespaceVersion,
			},
			Annotations: map[string]string{
				EnforceSchedulerAnnotationName: "true",
			},
			OwnerReferences: ProjectOwnerRef,
		},
	}

	TestExternalWorkloads = externalWorkloadFixture("ew-1")

	TestPVCsProjectScope = corev1.PersistentVolumeClaimList{
		TypeMeta: metav1.TypeMeta{
			Kind:       pvcKind,
			APIVersion: common.CoreV1ApiVersion,
		},
		ListMeta: metav1.ListMeta{},
		Items: []corev1.PersistentVolumeClaim{
			{
				TypeMeta: metav1.TypeMeta{
					Kind:       pvcKind,
					APIVersion: common.CoreV1ApiVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "runai-pvc-project-scope",
					Namespace: TestNamespace.Name,
					Labels: map[string]string{
						ProjectScopeLabel: TestProject.Name,
					},
				},
			},
			{
				TypeMeta: metav1.TypeMeta{
					Kind:       pvcKind,
					APIVersion: common.CoreV1ApiVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "non-runai-pvc-project-scope",
					Namespace: TestNamespace.Name,
				},
			},
			{
				TypeMeta: metav1.TypeMeta{
					Kind:       pvcKind,
					APIVersion: common.CoreV1ApiVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "runai-pvc-cluster-scope",
					Namespace: TestNamespace.Name,
					Labels: map[string]string{
						ClusterWideResourceLabel: TestProject.Name,
					},
				},
			},
		},
	}

	DefaultLimitRangeConfigMap = corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       ConfigMapKind,
			APIVersion: common.CoreV1ApiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      common.DefaultLimitRangeConfigMapName,
			Namespace: testRunaiNamespace,
		},
		Data: map[string]string{
			"memoryMaxLimit":       "100M",
			"cpuMaxLimit":          "150m",
			"memoryDefaultLimit":   "20Mi",
			"cpuDefaultLimit":      "2Gi",
			"memoryDefaultRequest": "300Ki",
			"cpuDefaultRequest":    "350",
		},
	}

	NamespacedLimitRange = corev1.LimitRange{
		TypeMeta: metav1.TypeMeta{
			Kind:       common.LimitRangeKind,
			APIVersion: common.CoreV1ApiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            RunaiLimitRangeName,
			Namespace:       TestNamespace.Name,
			OwnerReferences: ProjectOwnerRef,
		},
		Spec: corev1.LimitRangeSpec{
			Limits: []corev1.LimitRangeItem{
				{
					Type:           corev1.LimitTypeContainer,
					Max:            map[corev1.ResourceName]resource.Quantity{common.MemoryResourceName: resource.MustParse("100M"), common.CpuResourceName: resource.MustParse("150m")},
					Default:        map[corev1.ResourceName]resource.Quantity{common.MemoryResourceName: resource.MustParse("20Mi"), common.CpuResourceName: resource.MustParse("2Gi")},
					DefaultRequest: map[corev1.ResourceName]resource.Quantity{common.MemoryResourceName: resource.MustParse("300Ki"), common.CpuResourceName: resource.MustParse("350")},
				},
			},
		},
	}

	AnotherDefaultLimitRangeConfigMap = corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       ConfigMapKind,
			APIVersion: common.CoreV1ApiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      common.DefaultLimitRangeConfigMapName,
			Namespace: testRunaiNamespace,
		},
		Data: map[string]string{
			"memoryMaxLimit":       "400M",
			"cpuMaxLimit":          "450m",
			"memoryDefaultLimit":   "500Mi",
			"cpuDefaultLimit":      "5Gi",
			"memoryDefaultRequest": "600Ki",
			"cpuDefaultRequest":    "650",
		},
	}

	AnotherNamespacedLimitRange = corev1.LimitRange{
		TypeMeta: metav1.TypeMeta{
			Kind:       common.LimitRangeKind,
			APIVersion: common.CoreV1ApiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            RunaiLimitRangeName,
			Namespace:       TestNamespace.Name,
			OwnerReferences: ProjectOwnerRef,
		},
		Spec: corev1.LimitRangeSpec{
			Limits: []corev1.LimitRangeItem{
				{
					Type:           corev1.LimitTypeContainer,
					Max:            map[corev1.ResourceName]resource.Quantity{common.MemoryResourceName: resource.MustParse("400M"), common.CpuResourceName: resource.MustParse("450m")},
					Default:        map[corev1.ResourceName]resource.Quantity{common.MemoryResourceName: resource.MustParse("500Mi"), common.CpuResourceName: resource.MustParse("5Gi")},
					DefaultRequest: map[corev1.ResourceName]resource.Quantity{common.MemoryResourceName: resource.MustParse("600Ki"), common.CpuResourceName: resource.MustParse("650")},
				},
			},
		},
	}

	ProjectOwnerRef = []metav1.OwnerReference{{
		APIVersion:         kaiv1alpha1.GroupVersion.Identifier(),
		Kind:               common.ProjectKind,
		Name:               TestProject.Name,
		UID:                TestProject.UID,
		Controller:         &common.TrueRef,
		BlockOwnerDeletion: &common.TrueRef,
	}}

	UnrelatedProjectOwnerRef = []metav1.OwnerReference{{
		APIVersion:         kaiv1alpha1.GroupVersion.Identifier(),
		Kind:               common.ProjectKind,
		Name:               "unrelated-proj",
		UID:                "unrelated-uid",
		Controller:         &common.TrueRef,
		BlockOwnerDeletion: &common.TrueRef,
	}}

	JobControllerRoleBinding = rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       common.LogRoleBindingTag,
			APIVersion: common.RbacGroup + "/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            JobControllerRoleBindingName,
			Namespace:       TestNamespace.Name,
			OwnerReferences: ProjectOwnerRef,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: common.RbacGroup,
			Kind:     common.ClusterRoleKind,
			Name:     JobControllerRoleBindingName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      common.ServiceAccountKind,
				APIGroup:  common.RbacGroup,
				Name:      ProjectControllerServiceAccountName,
				Namespace: TestNamespace.Name,
			},
		},
	}

	// KaiTestDepartment1/2/3 mirror TestDepartment1/2/3 but use the KAI Department CRD
	// (kai/v1alpha1) consumed by the non-legacy department reconciler and handler. The
	// legacy scheduling/v2alpha2 fixtures above are retained for the legacy code paths.
	KaiTestDepartment1 = &kaiv1alpha1.Department{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Department",
			APIVersion: kaiv1alpha1.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: TestDepartment1Name,
			UID:  "23232",
			Labels: map[string]string{
				DepartmentIdLabel: TestDepartment1Id,
			},
		},
		Spec: kaiv1alpha1.DepartmentSpec{
			Queues: []kaiv1alpha1.QueueConfig{
				{
					Name:     "unimportant-name",
					Nodepool: DefaultNodePoolName,
					Resources: &kaiv1alpha1.QueueResourcesConfig{
						GPU: kaiv1alpha1.SystemResource{
							Deserved: 10,
						},
						CPU: kaiv1alpha1.SystemResource{
							Deserved: 5000,
						},
					},
				},
				{
					Name:     "unimportant-name2",
					Nodepool: SomeNodePoolName,
					Resources: &kaiv1alpha1.QueueResourcesConfig{
						GPU: kaiv1alpha1.SystemResource{
							Deserved: 25,
						},
						CPU: kaiv1alpha1.SystemResource{
							Deserved: 20000,
						},
						Memory: kaiv1alpha1.SystemResource{
							Deserved:        1000,
							Limit:           2000,
							OverQuotaWeight: 4,
						},
					},
				},
			},
		},
	}

	KaiTestDepartment2 = &kaiv1alpha1.Department{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Department",
			APIVersion: kaiv1alpha1.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: TestDepartment2Name,
			UID:  "12345",
			Labels: map[string]string{
				DepartmentIdLabel: TestDepartment2Id,
			},
		},
		Spec: kaiv1alpha1.DepartmentSpec{
			Queues: []kaiv1alpha1.QueueConfig{
				{
					Name:     "unimportant-name3",
					Nodepool: DefaultNodePoolName,
					Resources: &kaiv1alpha1.QueueResourcesConfig{
						GPU: kaiv1alpha1.SystemResource{
							Deserved: 0,
							Limit:    0,
						},
						CPU: kaiv1alpha1.SystemResource{
							Deserved: 4000,
						},
					},
					Priority: ptr.To(int32(200)),
				},
				{
					Name:     "unimportant-name4",
					Nodepool: SomeNodePoolName,
					Resources: &kaiv1alpha1.QueueResourcesConfig{
						GPU: kaiv1alpha1.SystemResource{
							Deserved: 20,
						},
						CPU: kaiv1alpha1.SystemResource{
							Deserved: 25000,
						},
						Memory: kaiv1alpha1.SystemResource{
							Deserved:        1500,
							Limit:           2500,
							OverQuotaWeight: 0,
						},
					},
					Priority: ptr.To(int32(240)),
				},
			},
		},
	}

	KaiTestDepartment3 = &kaiv1alpha1.Department{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Department",
			APIVersion: kaiv1alpha1.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: TestDepartment3Name,
			UID:  "54321",
			Labels: map[string]string{
				DepartmentIdLabel: TestDepartment3Id,
			},
		},
		Spec: kaiv1alpha1.DepartmentSpec{
			Queues: []kaiv1alpha1.QueueConfig{
				{
					Name:     "unimportant-name33",
					Nodepool: DefaultNodePoolName,
					Resources: &kaiv1alpha1.QueueResourcesConfig{
						GPU: kaiv1alpha1.SystemResource{
							Deserved: 0,
							Limit:    0,
						},
						CPU: kaiv1alpha1.SystemResource{
							Deserved: 4000,
						},
					},
				},
				{
					Name:     "unimportant-name44",
					Nodepool: SomeNodePoolName,
					Resources: &kaiv1alpha1.QueueResourcesConfig{
						GPU: kaiv1alpha1.SystemResource{
							Deserved: 20,
						},
						CPU: kaiv1alpha1.SystemResource{
							Deserved: 25000,
						},
						Memory: kaiv1alpha1.SystemResource{
							Deserved:        1500,
							Limit:           2500,
							OverQuotaWeight: 0,
						},
					},
				},
			},
		},
	}

	// KaiDep1OwnerRef is the owner reference the non-legacy department handler stamps on
	// queues it creates for KaiTestDepartment1 (KAI Department GVK).
	KaiDep1OwnerRef = []metav1.OwnerReference{{
		APIVersion:         kaiv1alpha1.GroupVersion.Identifier(),
		Kind:               "Department",
		Name:               KaiTestDepartment1.Name,
		UID:                KaiTestDepartment1.UID,
		Controller:         &common.TrueRef,
		BlockOwnerDeletion: &common.TrueRef,
	}}

	KaiTestQueueDep1NpD = kaiv2.Queue{
		TypeMeta: metav1.TypeMeta{
			APIVersion: kaiv2.GroupVersion.Identifier(),
			Kind:       common.QueueKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            handlers.GetQueueName(KaiTestDepartment1.Spec.Queues[0], TestDepartment1Name),
			OwnerReferences: KaiDep1OwnerRef,
			Labels:          map[string]string{},
		},
	}

	KaiTestQueueDep1Np1 = kaiv2.Queue{
		TypeMeta: metav1.TypeMeta{
			APIVersion: kaiv2.GroupVersion.Identifier(),
			Kind:       common.QueueKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            handlers.GetQueueName(KaiTestDepartment1.Spec.Queues[1], TestDepartment1Name),
			OwnerReferences: KaiDep1OwnerRef,
			Labels: map[string]string{
				DepartmentNameLabel: TestDepartment1Name,
				NodePoolLabelKey:    SomeNodePoolName,
			},
		},
		Spec: kaiv2.QueueSpec{
			DisplayName: TestDepartment1Name,
			ParentQueue: "",
			Resources: &kaiv2.QueueResources{
				GPU: kaiv2.QueueResource{
					Quota: 25,
				},
				CPU: kaiv2.QueueResource{
					Quota: 20000,
				},
				Memory: kaiv2.QueueResource{
					Quota:           1000,
					Limit:           2000,
					OverQuotaWeight: 4,
				},
			},
			Priority: ptr.To(100),
		},
	}

	QueueStatusForTests = kaiv2.QueueStatus{
		Conditions: []kaiv2.QueueCondition{
			{
				Type:    kaiv2.OverQuota,
				Status:  corev1.ConditionTrue,
				Reason:  "a train job is over quota",
				Message: "the project is over quota",
			},
			{
				Type:   kaiv2.Orphan,
				Status: corev1.ConditionFalse,
			},
		},
		Allocated: corev1.ResourceList{
			"nvidia.com/gpu": resource.MustParse("1000m"),
			"cpu":            resource.MustParse("600m"),
			"memory":         resource.MustParse("300m"),
		},
		AllocatedNonPreemptible: corev1.ResourceList{
			"nvidia.com/gpu": resource.MustParse("500m"),
			"cpu":            resource.MustParse("300m"),
		},
		Requested: corev1.ResourceList{
			"nvidia.com/gpu": resource.MustParse("2000m"),
			"cpu":            resource.MustParse("1000m"),
			"memory":         resource.MustParse("600m"),
		},
	}
)

// RunaiConfigForTests returns the runai-flavored ProjectReconcilerConfig used by the test
// suites (runai/queue, runai/node-pool, runai namespace, run.ai finalizer domain, ...).
func RunaiConfigForTests() *config.ProjectReconcilerConfig {
	return &config.ProjectReconcilerConfig{
		NodePoolLabelKey:               NodePoolLabelKey,
		DefaultNodepoolName:            DefaultNodePoolName,
		QueueLabelKey:                  RunaiQueueLabel,
		NamespaceProjectLabelKey:       RunaiQueueLabel,
		ProjectLabelKey:                RunaiProjectLabel,
		ProjectIdLabelKey:              "run.ai/project-id",
		QueueDepartmentNameLabelKey:    QueueDepartmentNameLabel,
		NamespaceVersionLabelKey:       "runai/namespace-version",
		EnforceSchedulerAnnotationKey:  EnforceSchedulerAnnotationName,
		FinalizerDomain:                "run.ai",
		InstallNamespace:               "runai",
		ProjectNamePrefix:              "runai",
		LimitRangeName:                 RunaiLimitRangeName,
		ResourceManualOverrideLabelKey: RunaiResourceManualOverrideLabel,
	}
}

func RunTest(t *testing.T, testDescription string, reportFilePath string) {
	if _, isCi := os.LookupEnv(CircleCiEnvVar); isCi {
		RunSpecsWithDefaultAndCustomReporters(t, testDescription, []Reporter{reporters.NewJUnitReporter(reportFilePath)})
	} else {
		RunSpecsWithCustomReporters(t, testDescription, []Reporter{golandreporter.NewGolandReporter()})
	}
}

func SetUnexportedField(obj interface{}, fieldName string, value interface{}) {
	field := reflect.ValueOf(obj).
		Elem().
		FieldByName(fieldName)

	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).
		Elem().
		Set(reflect.ValueOf(value))
}
