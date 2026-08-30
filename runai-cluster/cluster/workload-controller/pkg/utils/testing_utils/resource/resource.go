package resource

import (
	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	kube "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	runaiJobV1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	runaiv2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v2"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	ssv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferencepolicy/v2alpha1"
	infv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	mpiv2beta1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/mpijob/v2beta1"
	v1 "k8s.io/api/core/v1"
	network "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func CreateDumbRunaiJob(name, namespace string, ownerUID types.UID, ownerKind string, spec runaiJobV1.JobSpec) *runaiJobV1.RunaiJob {
	runaiJob := &runaiJobV1.RunaiJob{
		TypeMeta: metav1.TypeMeta{
			Kind: constants.KindRunaiJob,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: spec,
	}
	if ownerUID != "" && ownerKind != "" {
		runaiJob.OwnerReferences = []metav1.OwnerReference{
			{
				UID:  ownerUID,
				Kind: ownerKind,
			},
		}
	}
	return runaiJob
}

func CreateDumbInferenceWorkload(name, namespace string, spec ssv1.InferenceSpec) *infv2.InferenceWorkload {
	return &infv2.InferenceWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: spec,
	}
}

// CreateNodePool builds a legacy run.ai NodePool fixture.
// TODO: delete once clusterconstants.UseKaiNodePools=true
func CreateNodePool(name, key, value string) *v1alpha1.NodePool {
	nodePools := v1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.NodePoolSpec{
			LabelKey:   key,
			LabelValue: value,
		},
	}
	return &nodePools
}

// CreateKaiNodePool is the kai.resources twin of CreateNodePool. Seed both so fixtures
// resolve the same node pool under either value of the compile-time UseKaiNodePools flag.
func CreateKaiNodePool(name, key, value string) *kaiorg.NodePool {
	nodePools := kaiorg.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: kaiorg.NodePoolSpec{
			LabelKey:   key,
			LabelValue: value,
		},
	}
	return &nodePools
}

func CreateDumbService(name, namespace string, ownerUID types.UID, ownerKind string, labels map[string]string) *v1.Service {
	service := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind: constants.KindService,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
	}
	if ownerUID != "" && ownerKind != "" {
		service.OwnerReferences = []metav1.OwnerReference{
			{
				UID:  ownerUID,
				Kind: ownerKind,
			},
		}
	}
	return service
}

func CreateDumbIngress(name, namespace string, ownerUID types.UID, ownerKind string, labels map[string]string) *network.Ingress {
	ingress := &network.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind: constants.KindIngress,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
	}
	if ownerUID != "" && ownerKind != "" {
		ingress.OwnerReferences = []metav1.OwnerReference{
			{
				UID:  ownerUID,
				Kind: ownerKind,
			},
		}
	}
	return ingress
}

func CreateDumbRoute(name, namespace string, ownerUID types.UID, ownerKind string, labels map[string]string) *routev1.Route {
	route := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
	}
	if ownerUID != "" && ownerKind != "" {
		route.OwnerReferences = []metav1.OwnerReference{
			{
				UID:  ownerUID,
				Kind: ownerKind,
			},
		}
	}
	return route
}

func CreateDumbHTTPRoute(name, namespace string, ownerUID types.UID, ownerKind string, labels map[string]string) *gatewayv1.HTTPRoute {
	httpRoute := &gatewayv1.HTTPRoute{
		TypeMeta: metav1.TypeMeta{
			Kind: constants.KindHTTPRoute,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
	}
	if ownerUID != "" && ownerKind != "" {
		httpRoute.OwnerReferences = []metav1.OwnerReference{
			{
				UID:  ownerUID,
				Kind: ownerKind,
			},
		}
	}
	return httpRoute
}

func CreateDumbPytorch(name, namespace string, ownerUID types.UID, ownerKind string, labels map[string]string) *kube.PyTorchJob {
	pytorch := &kube.PyTorchJob{
		TypeMeta: metav1.TypeMeta{
			Kind: constants.KindPyTorchJob,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
	}
	if ownerUID != "" && ownerKind != "" {
		pytorch.OwnerReferences = []metav1.OwnerReference{
			{
				UID:  ownerUID,
				Kind: ownerKind,
			},
		}
	}
	return pytorch
}
func CreateDumbTensorflow(name, namespace string, ownerUID types.UID, ownerKind string, labels map[string]string) *kube.TFJob {
	tfJob := &kube.TFJob{
		TypeMeta: metav1.TypeMeta{
			Kind: constants.KindTFJob,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
	}
	if ownerUID != "" && ownerKind != "" {
		tfJob.OwnerReferences = []metav1.OwnerReference{
			{
				UID:  ownerUID,
				Kind: ownerKind,
			},
		}
	}
	return tfJob
}

func CreateDumbXGBoost(name, namespace string, ownerUID types.UID, ownerKind string, labels map[string]string) *kube.XGBoostJob {
	xgboostJob := &kube.XGBoostJob{
		TypeMeta: metav1.TypeMeta{
			Kind: constants.KindXGBoostJob,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
	}
	if ownerUID != "" && ownerKind != "" {
		xgboostJob.OwnerReferences = []metav1.OwnerReference{
			{
				UID:  ownerUID,
				Kind: ownerKind,
			},
		}
	}
	return xgboostJob
}

func CreateDumbJAXJob(name, namespace string, ownerUID types.UID, ownerKind string, labels map[string]string) *kube.JAXJob {
	jaxJob := &kube.JAXJob{
		TypeMeta: metav1.TypeMeta{
			Kind: constants.KindJAXJob,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
	}
	if ownerUID != "" && ownerKind != "" {
		jaxJob.OwnerReferences = []metav1.OwnerReference{
			{
				UID:  ownerUID,
				Kind: ownerKind,
			},
		}
	}
	return jaxJob
}

func CreateDumbMPI(name, namespace string, ownerUID types.UID, ownerKind string, labels map[string]string) *mpiv2beta1.MPIJob {
	mpi := &mpiv2beta1.MPIJob{
		TypeMeta: metav1.TypeMeta{
			Kind: constants.KindMpiJob,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
	}
	if ownerUID != "" && ownerKind != "" {
		mpi.OwnerReferences = []metav1.OwnerReference{
			{
				UID:  ownerUID,
				Kind: ownerKind,
			},
		}
	}
	return mpi
}

func CreatePvc(name, namespace string, capacitySize resource.Quantity, ownerUID types.UID, ownerKind, StorageClassName string, accessModes []v1.PersistentVolumeAccessMode, mode *v1.PersistentVolumeMode, labels map[string]string) *v1.PersistentVolumeClaim {
	pvc := &v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind: constants.KindPvc,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: accessModes,
			Selector:    nil,
			Resources: v1.VolumeResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: capacitySize,
				},
			},
			VolumeName:       name,
			StorageClassName: ptr.To(StorageClassName),
			VolumeMode:       mode,
		},
	}
	if ownerUID != "" && ownerKind != "" {
		pvc.OwnerReferences = []metav1.OwnerReference{
			{
				UID:  ownerUID,
				Kind: ownerKind,
			},
		}
	}
	return pvc
}

func CreateSecret(name string, namespace string, ownerUID types.UID, ownerKind string, labels map[string]string) *v1.Secret {
	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind: constants.KindSecret,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
	}
	if ownerUID != "" && ownerKind != "" {
		secret.OwnerReferences = []metav1.OwnerReference{
			{
				UID:  ownerUID,
				Kind: ownerKind,
			},
		}
	}
	return secret
}

func CreateNamespace(name, project string) *v1.Namespace {

	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: map[string]string{constants.RunaiQueueLabel: project},
		},
	}
}

func CreateProject(name string) *runaiv2.Project {

	fakeProject := runaiv2.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: runaiv2.ProjectSpec{
			InteractiveJobTimeLimitSecs:                  120,
			InteractiveJobMaxIdleDurationSecs:            1200,
			InteractivePreemptibleJobMaxIdleDurationSecs: 1800,
			TrainingJobMaxIdleDurationSecs:               2400,
			TrainingJobTimeLimitSecs:                     120,
		},
	}
	return &fakeProject
}

// CreateKaiProject builds the kai.resources twin of CreateProject, carrying the same
// limits as the run.ai/*-seconds annotations consumers read.
func CreateKaiProject(name string) *kaiorg.Project {
	return &kaiorg.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				constants.InteractiveJobTimeLimitSecondsAnnotation:                  "120",
				constants.InteractiveJobMaxIdleDurationSecondsAnnotation:            "1200",
				constants.InteractivePreemptibleJobMaxIdleDurationSecondsAnnotation: "1800",
				constants.TrainingJobMaxIdleDurationSecondsAnnotation:               "2400",
				constants.TrainingJobTimeLimitSecondsAnnotation:                     "120",
			},
		},
	}
}

func NewRunaiObject(name string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      name,
		Namespace: constants.NamespaceRunAi,
	}
}

func CreateElasticPyTorchJob(minReplicas, maxReplicas *int) *jobs.JobWorkload {
	return &jobs.JobWorkload{
		TypeMeta: metav1.TypeMeta{
			Kind: constants.KindDistributedWorkload,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "w1",
		},
		Spec: jobs.JobSpec{
			DistributedFields: field_groups.DistributedFields{
				JobType: constants.KindPyTorchJob,
				PyTorchJob: &field_types.PyTorchField{
					Values: &field_types.PyTorchMembers{
						MinReplicas: field_types.NewOptionalIntField(minReplicas),
						MaxReplicas: field_types.NewOptionalIntField(maxReplicas),
					},
				},
			},
			ResourceAllocationFields: field_groups.ResourceAllocationFields{
				GpuDevices: field_types.NewOptionalIntField(ptr.To(1)),
			},
			RunaiSchedulerFields: field_groups.RunaiSchedulerFields{
				NodePools: field_types.NewStringField("default"),
			},
		},
	}
}

func CreateDumbComputeDomain(name, namespace string, ownerUID types.UID, ownerKind string, labels map[string]string) *nvapi.ComputeDomain {
	computeDomain := &nvapi.ComputeDomain{
		TypeMeta: metav1.TypeMeta{
			Kind: nvapi.ComputeDomainKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
	}
	if ownerUID != "" && ownerKind != "" {
		computeDomain.OwnerReferences = []metav1.OwnerReference{
			{
				UID:  ownerUID,
				Kind: ownerKind,
			},
		}
	}
	return computeDomain
}
