package mnnvl

import (
	"context"
	"errors"
	"fmt"
	"strings"

	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	skipCreationMessage = "skipping ComputeDomain creation"

	// Labels set on the generated ComputeDomain. Kept stable so observability
	// dashboards that match on these keys keep working.
	labelWorkloadKind = "workloadKind"
	labelWorkloadName = "workloadName"
	labelRelease      = "release"
	labelApp          = "app"
	labelCreatedBy    = "createdBy"
)

var (
	errComputeDomainCpuOnly      = errors.New("ComputeDomain is not supported for CPU-only workloads")
	errSetControllerReferenceFmt = fmt.Sprintf("failed to set controller reference for %s", nvapi.ComputeDomainKind)
)

type computeDomainWorkloadParams struct {
	WorkloadKind      string
	JobType           string
	WorkloadInitials  string
	WorkloadName      string
	WorkloadNamespace string
	NumNodes          int
	ElasticEnabled    bool
}

// buildComputeDomain constructs an NVIDIA DRA ComputeDomain CR for the given
// workload. Only NVIDIA GPUs are supported (ComputeDomain is an NVIDIA-specific
// resource - the upstream vendor check was always a no-op for the supported
// case).
func buildComputeDomain(ctx context.Context, params computeDomainWorkloadParams, scheme *runtime.Scheme, owner client.Object) (*nvapi.ComputeDomain, error) {
	logger := log.FromContext(ctx)

	objectMeta := buildComputeDomainObjectMeta(params)

	spec, err := buildComputeDomainSpec(params)
	if err != nil {
		return nil, err
	}

	computeDomain := &nvapi.ComputeDomain{
		TypeMeta: metav1.TypeMeta{
			Kind:       nvapi.ComputeDomainKind,
			APIVersion: nvapi.Version,
		},
		ObjectMeta: *objectMeta,
		Spec:       *spec,
	}

	if owner != nil {
		if scheme == nil {
			return nil, fmt.Errorf("%s, scheme is nil", errSetControllerReferenceFmt)
		}

		if err = controllerutil.SetControllerReference(owner, computeDomain, scheme); err != nil {
			return nil, fmt.Errorf("%s, %v", errSetControllerReferenceFmt, err)
		}
	}

	logger.V(1).Info("built ComputeDomain", "name", computeDomain.Name, "numNodes", spec.NumNodes)
	return computeDomain, nil
}

func buildComputeDomainObjectMeta(params computeDomainWorkloadParams) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		Name:      composeComputeDomainName(params.WorkloadInitials, params.WorkloadName),
		Namespace: params.WorkloadNamespace,
		Labels: map[string]string{
			labelWorkloadKind: params.WorkloadKind,
			labelWorkloadName: params.WorkloadName,
			labelRelease:      params.WorkloadName,
			labelApp:          strings.ToLower(params.JobType),
			labelCreatedBy:    params.JobType,
		},
	}
}

func buildComputeDomainSpec(params computeDomainWorkloadParams) (*nvapi.ComputeDomainSpec, error) {
	if params.NumNodes == 0 {
		return nil, errComputeDomainCpuOnly
	}

	rct := createResourceClaimTemplate(params)

	// NumNodes=0 enables elastic compute domain.
	cdNodes := 0
	if !params.ElasticEnabled {
		cdNodes = params.NumNodes
	}

	return &nvapi.ComputeDomainSpec{
		NumNodes: cdNodes,
		Channel: &nvapi.ComputeDomainChannelSpec{
			ResourceClaimTemplate: *rct,
		},
	}, nil
}

func composeComputeDomainName(workloadInitials, workloadName string) string {
	return fmt.Sprintf("%s-%s", workloadInitials, workloadName)
}
