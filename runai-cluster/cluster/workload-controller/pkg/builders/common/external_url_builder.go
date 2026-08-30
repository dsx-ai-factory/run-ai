package common

import (
	"context"

	workloadbuilders "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders"

	runaijob "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	JupyterServiceSuffix     = "jupyter"
	TensorboardServiceSuffix = "tensorboard"
)

type ExternalsUrlsBuilder struct {
	ObjectNaming         *metav1.ObjectMeta
	SharedFields         *field_groups.WorkloadFields
	workloadTypeInitials WorkloadInitials
	Namespace            string
	resourceBaseName     string
	Labels               map[string]string
	k8sClient            client.Client
	IsOpenshift          bool
	IsJupyter            bool
	IsTensorboard        bool
}

func NewExternalsUrlsBuilder(k8sClient client.Client, isOpenshift bool) *ExternalsUrlsBuilder {
	return &ExternalsUrlsBuilder{
		k8sClient:   k8sClient,
		IsOpenshift: isOpenshift,
	}
}

// GetPreview returns a list of resource name and kind that will be created on build
func (externalUrlBuilder *ExternalsUrlsBuilder) GetPreview(ctx context.Context, workload *jobs.JobWorkload) ([]common.Resource, error) {
	err := externalUrlBuilder.From(workload)
	if err != nil {
		return nil, err
	}

	eb := workloadbuilders.NewExternalUrlBuilder(externalUrlBuilder.k8sClient, externalUrlBuilder.IsOpenshift)
	resources, err := eb.GetPreview(ctx, externalUrlBuilder.workloadTypeInitials.String(), externalUrlBuilder.ObjectNaming, externalUrlBuilder.SharedFields.ExposedUrls,
		externalUrlBuilder.IsJupyter, externalUrlBuilder.IsTensorboard)
	if err != nil {
		return nil, err
	}

	resourcesCommon := common.CommonResourceToResource(resources)

	return resourcesCommon, nil
}

func (externalUrlBuilder *ExternalsUrlsBuilder) From(workload *jobs.JobWorkload) error {
	externalUrlBuilder.ObjectNaming = &workload.ObjectMeta
	// Clear old job url annotation if it exists
	delete(externalUrlBuilder.ObjectNaming.Annotations, constants.AnnotationJobUrl)

	externalUrlBuilder.SharedFields = &workload.Spec.WorkloadFields
	externalUrlBuilder.Namespace = workload.ObjectMeta.Namespace
	externalUrlBuilder.resourceBaseName = workload.GetChildName()
	externalUrlBuilder.IsJupyter = workload.Spec.IsJupyter.Deref()
	externalUrlBuilder.IsTensorboard = workload.Spec.IsTensorboard.Deref()
	initials, err := GetWorkloadInitials(workload.Kind)
	if err != nil {
		return err
	}
	externalUrlBuilder.workloadTypeInitials = initials
	return nil

}

func (externalUrlBuilder *ExternalsUrlsBuilder) Build(ctx context.Context, workload *jobs.JobWorkload, options *ExternalBuilderOptions) ([]client.Object, error) {
	if options == nil {
		options = &ExternalBuilderOptions{}
	}

	err := externalUrlBuilder.From(workload)
	if err != nil {
		return nil, err
	}

	jobType := GetJobType(workload)

	calculatedPriorityClassNameLabel := CalculatePriorityClassNameLabel(workload.Spec.PriorityClassName.Deref(), workload.Spec.Preemptibility, workload.Kind)

	eb := workloadbuilders.NewExternalUrlBuilder(externalUrlBuilder.k8sClient, externalUrlBuilder.IsOpenshift).
		WithAuthPortOffset(options.AuthPortOffset)
	services, ingresses, routes, httpRoutes, err := eb.Build(ctx, workload.Kind, jobType, externalUrlBuilder.workloadTypeInitials.String(),
		externalUrlBuilder.ObjectNaming, externalUrlBuilder.resourceBaseName,
		options.Owner, externalUrlBuilder.SharedFields.ExposedUrls, externalUrlBuilder.SharedFields.IngressUrl,
		calculatedPriorityClassNameLabel,
		externalUrlBuilder.IsJupyter, externalUrlBuilder.IsTensorboard, ptr.Deref(options.DisableAuth, false))
	if err != nil {
		return nil, err
	}

	objects := make([]client.Object, 0)
	for _, service := range services {
		objects = append(objects, &service)
	}
	for _, ingress := range ingresses {
		objects = append(objects, &ingress)
	}
	for _, route := range routes {
		objects = append(objects, &route)
	}
	for _, httpRoute := range httpRoutes {
		objects = append(objects, &httpRoute)
	}

	return objects, nil
}

func IsJobSupportsAuth(job client.Object) bool {
	runaiJob, ok := job.(*runaijob.RunaiJob)
	if ok && runaiJob != nil {
		return workloadbuilders.IsWorkloadSupportsAuth(runaiJob.Spec.Template.Spec.Containers, runaiJob.Spec.Template.Spec.InitContainers)
	}

	return false
}
