package test

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	scheme_init "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	v2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v2"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	dpv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	dwv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	ipv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	iwv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	mpi "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/mpijob/v1alpha2"
	tpv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingpolicy/v2alpha1"
	twv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	testProjectName   = "test-project"
	testNamespaceName = "runai-test-namespace"
	RunaiQueueLabel   = "runai/queue"
)

var workerSleepSeconds = 5
var eventList = []int{50}

// MockEnv Notice that after test is done env variables are cleaned up
func MockEnv(t *testing.T) {
	t.Setenv("NUMBER_OF_WORKERS_PER_QUEUE", "2")
	t.Setenv("WORKER_SLEEP_SECONDS", strconv.Itoa(workerSleepSeconds))
	t.Setenv("EVENT_LIST", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(eventList)), ","), "[]"))
	t.Setenv("CLUSTER_VERSION", "2.99.0")
}

func GenerateWorkloadName() string {
	return fmt.Sprintf("test-workload-%s", rand.String(5))
}

func BuildNamespace(name string, labels map[string]string) *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}
}

var (
	testSchemeOnce sync.Once
	testSchemeInst *apiruntime.Scheme
)

func initSchema() *apiruntime.Scheme {
	testSchemeOnce.Do(func() {
		scheme := scheme_init.Scheme()
		if err := mpi.AddToScheme(scheme); err != nil {
			zap.S().Error("failed to add mpijob to scheme", err)
		}
		if err := iwv2a1.AddKnownTypes(scheme); err != nil {
			zap.S().Error("failed to add interactive workload types", err)
		}
		if err := iwv2a1.AddToScheme(scheme); err != nil {
			zap.S().Error("failed to add interactive workload to scheme", err)
		}
		if err := twv2a1.AddKnownTypes(scheme); err != nil {
			zap.S().Error("failed to add training workload types", err)
		}
		if err := twv2a1.AddToScheme(scheme); err != nil {
			zap.S().Error("failed to add training workload to scheme", err)
		}
		if err := dwv2a1.AddKnownTypes(scheme); err != nil {
			zap.S().Error("failed to add distributed workload types", err)
		}
		if err := dwv2a1.AddToScheme(scheme); err != nil {
			zap.S().Error("failed to add distributed workload to scheme", err)
		}
		if err := dromosv1alpha1.AddToScheme(scheme); err != nil {
			zap.S().Error("failed to add workloadrunner to scheme", err)
		}
		testSchemeInst = scheme
	})
	return testSchemeInst
}

func BuildInteractiveWorkload(namespace, name string, jobStartTime *metav1.Time) *iwv2a1.InteractiveWorkload {
	if jobStartTime == nil {
		jobStartTime = &metav1.Time{Time: time.Now()}
	}

	return &iwv2a1.InteractiveWorkload{
		TypeMeta: metav1.TypeMeta{
			Kind:       constants.KindInteractiveWorkload,
			APIVersion: "run.ai/v2alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				constants.RunaiProjectLabel: testProjectName,
			},
		},
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityDeprecatedFields: field_groups.IdentityDeprecatedFields{
					Name: &flds.StringField{
						Value: ptr.To(name),
					},
				},
				IdentityFields: field_groups.IdentityFields{
					Usage: common.UsedForSubmit,
				},
				ContainerFields: field_groups.ContainerFields{
					Image: &flds.StringField{
						Value: ptr.To("ubuntu"),
					},
				},
			},
		},
		Status: iwv2a1.InteractiveWorkloadStatus{
			WorkloadCommonStatus: common.WorkloadCommonStatus{
				StartTime: jobStartTime,
			},
			WorkloadDebugStatus: common.WorkloadDebugStatus{

				OwnedResources: []common.ObservedResource{
					{
						Resource: common.Resource{
							Kind: metav1.GroupVersionKind{
								Version: "run.ai/v1",
								Kind:    constants.KindRunaiJob,
							},
							Name: "test",
						},
						Uid: uuid.NewUUID(),
					},
				},
			},
		},
	}
}

func BuildTrainingWorkload(namespace, name string, jobStartTime *metav1.Time) *twv2a1.TrainingWorkload {
	if jobStartTime == nil {
		jobStartTime = &metav1.Time{Time: time.Now()}
	}

	return &twv2a1.TrainingWorkload{

		TypeMeta: metav1.TypeMeta{
			Kind:       constants.KindTrainingWorkload,
			APIVersion: "run.ai/v2alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				constants.RunaiProjectLabel: testProjectName,
			},
		},
		Spec: tpv1.TrainingSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityDeprecatedFields: field_groups.IdentityDeprecatedFields{
					Name: &flds.StringField{
						Value: ptr.To(name),
					},
				},
				IdentityFields: field_groups.IdentityFields{
					Usage: common.UsedForSubmit,
				},
				ContainerFields: field_groups.ContainerFields{
					Image: &flds.StringField{
						Value: ptr.To("ubuntu"),
					},
				},
			},
		},
		Status: twv2a1.TrainingWorkloadStatus{
			WorkloadCommonStatus: common.WorkloadCommonStatus{
				StartTime: jobStartTime,
			},
			WorkloadDebugStatus: common.WorkloadDebugStatus{

				OwnedResources: []common.ObservedResource{
					{
						Resource: common.Resource{
							Kind: metav1.GroupVersionKind{
								Version: "run.ai/v1",
								Kind:    constants.KindRunaiJob,
							},
							Name: "test",
						},
						Uid: uuid.NewUUID(),
					},
				},
			},
		},
	}
}

func BuildDistributedWorkload(namespace, name string, jobStartTime *metav1.Time) *dwv2a1.DistributedWorkload {
	if jobStartTime == nil {
		jobStartTime = &metav1.Time{Time: time.Now()}
	}

	return &dwv2a1.DistributedWorkload{

		TypeMeta: metav1.TypeMeta{
			Kind:       constants.KindDistributedWorkload,
			APIVersion: "run.ai/v2alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				constants.RunaiProjectLabel: testProjectName,
			},
		},
		Spec: dpv1.DistributedSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityDeprecatedFields: field_groups.IdentityDeprecatedFields{
					Name: &flds.StringField{
						Value: ptr.To(name),
					},
				},
				IdentityFields: field_groups.IdentityFields{
					Usage: common.UsedForSubmit,
				},
				ContainerFields: field_groups.ContainerFields{
					Image: &flds.StringField{
						Value: ptr.To("ubuntu"),
					},
				},
			},
			DistributedFields: field_groups.DistributedFields{
				JobType: constants.KindPyTorchJob,
				Workers: &flds.IntField{
					Value: ptr.To(3),
				},
				NoMaster: &flds.BoolField{
					Value: ptr.To(true),
				},
			},
		},
		Status: dwv2a1.DistributedWorkloadStatus{
			WorkloadCommonStatus: common.WorkloadCommonStatus{
				StartTime: jobStartTime,
			},
			WorkloadDebugStatus: common.WorkloadDebugStatus{

				OwnedResources: []common.ObservedResource{
					{
						Resource: common.Resource{
							Kind: metav1.GroupVersionKind{
								Version: "kubeflow/v1",
								Kind:    constants.KindPyTorchJob,
							},
							Name: "test",
						},
						Uid: uuid.NewUUID(),
					},
				},
			},
		},
	}
}

func BuildProject(projectName, testNamespaceName string, trainingJobTimeLimitSecs, InteractiveJobTimeLimitSecs int) *v2.Project {
	return &v2.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: projectName,
		},
		Spec: v2.ProjectSpec{
			Queues: []v2.QueueProjectSpec{
				{
					Name: projectName,
				},
			},
			//TrainingJobMaxIdleDurationSecs: 10, // not in use for now
			//InteractiveJobMaxIdleDurationSecs: 10, // not in use for now
			TrainingJobTimeLimitSecs:    int64(trainingJobTimeLimitSecs),
			InteractiveJobTimeLimitSecs: int64(InteractiveJobTimeLimitSecs),
		},
		Status: v2.ProjectStatus{
			Namespace: testNamespaceName,
		},
	}
}

// BuildProjectFixtures returns the project objects to seed a fake client with.
func BuildProjectFixtures(projectName, testNamespaceName string, trainingJobTimeLimitSecs, InteractiveJobTimeLimitSecs int) []client.Object {
	return []client.Object{BuildKaiProject(projectName, testNamespaceName, trainingJobTimeLimitSecs, InteractiveJobTimeLimitSecs)}
}

// BuildKaiProject builds a kai.resources project. The CRD has no home for the job time limits,
// so they live on annotations - which is where the runtime-rules read them from.
func BuildKaiProject(projectName, testNamespaceName string, trainingJobTimeLimitSecs, InteractiveJobTimeLimitSecs int) *kaiorg.Project {
	return &kaiorg.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: projectName,
			Annotations: map[string]string{
				constants.TrainingJobTimeLimitSecondsAnnotation:    strconv.Itoa(trainingJobTimeLimitSecs),
				constants.InteractiveJobTimeLimitSecondsAnnotation: strconv.Itoa(InteractiveJobTimeLimitSecs),
			},
		},
		Spec: kaiorg.ProjectSpec{
			Queues: []kaiorg.QueueConfig{
				{
					Name: projectName,
				},
			},
			Namespace: testNamespaceName,
		},
		Status: kaiorg.ProjectStatus{
			Namespace: testNamespaceName,
		},
	}
}
