package runaijob

import (
	"bytes"
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"

	"github.com/ghodss/yaml"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	workloadbuilders "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders"
	nodepoolbuilders "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders/nodepool"
	workloadproject "github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	workloadCommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	logger "sigs.k8s.io/controller-runtime/pkg/log"

	rjv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"

	apicommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/common"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/factory"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var builderLog = logger.Log.WithName("runaijob.builder")

type Builder struct {
	JobBuilder
	//   the pod template spec builder provides the shared implementation for building podspec
	//
	PodTemplateSpecBuilder
	podLabels      map[string]string
	podAnnotations map[string]string
	k8sClient      client.Client
}

func NewRunaiJobBuilder(k8sClient client.Client, podLabels map[string]string,
	podAnnotations map[string]string, elasticEnabled bool) common.ComputeBuilder {
	return &Builder{
		k8sClient:      k8sClient,
		podLabels:      podLabels,
		podAnnotations: podAnnotations,
	}
}

func (b *Builder) GetPreview(_ context.Context, workload *jobs.JobWorkload) ([]apicommon.Resource, error) {
	return []apicommon.Resource{{
		Kind: metav1.GroupVersionKind{
			Group:   rjv1.Group,
			Version: rjv1.Version,
			Kind:    constants.KindRunaiJob,
		},
		Name:      workload.GetChildName(),
		Namespace: workload.Namespace,
	}}, nil
}

func (b *Builder) Build(ctx context.Context, workload *jobs.JobWorkload, options *common.ComputeBuilderOptions) (*common.ComputeResources, error) {
	job, err := b.buildRunaiJob(ctx, workload, options)
	if err != nil {
		return nil, err
	}

	return &common.ComputeResources{
		Job: job,
	}, nil
}

func (b *Builder) buildRunaiJob(ctx context.Context, workload *jobs.JobWorkload, options *common.ComputeBuilderOptions) (client.Object, error) {
	if options == nil {
		options = &common.ComputeBuilderOptions{}
	}
	builderLog.Info("Building Runai job from workload")
	job := &rjv1.RunaiJob{}
	job.Labels = make(map[string]string)
	job.Annotations = make(map[string]string)
	b.JobBuilder = JobBuilder{
		MergedWorkload: workload,
		IsJupyter:      workload.Spec.InteractiveFields.IsJupyter.Deref(),
		IsTensorboard:  workload.Spec.InteractiveFields.IsTensorboard.Deref(),
		OldSharedBuilder: common.OldSharedBuilder{
			ObjectNaming:    &workload.ObjectMeta,
			SharedFields:    &workload.Spec.WorkloadFields,
			SharedStatus:    &workload.Status.WorkloadStatus,
			SharedResources: &common.SharedResources{},
			Options:         factory.BuilderOptions{},
			K8sClient:       b.k8sClient,
		},
	}
	var workerSpec *corev1.PodSpec
	if options.ExitingPodsSpec != nil && options.ExitingPodsSpec[workloadCommon.Worker] != nil {
		workerSpec = options.ExitingPodsSpec[workloadCommon.Worker]
	}
	b.PodTemplateSpecBuilder = PodTemplateSpecBuilder{
		WorkloadKind:     workload.TypeMeta.Kind,
		MergedSpec:       &workload.Spec,
		IsJupyter:        workload.Spec.InteractiveFields.IsJupyter.Deref(),
		IsTensorboard:    workload.Spec.InteractiveFields.IsTensorboard.Deref(),
		ObjectType:       &job.TypeMeta,
		ObjectNaming:     &job.ObjectMeta,
		Pts:              &job.Spec.Template,
		ExitingPodSpec:   workerSpec,
		ResourceBaseName: workload.GetChildName(),
		PodLabels:        b.PodLabels,
		PodAnnotations:   b.PodAnnotations,
		K8sClient:        b.k8sClient,
	}

	b.buildRjTypeMeta(ctx, job)

	err := b.buildRjObjectMeta(ctx, workload, job)
	if err != nil {
		return nil, fmt.Errorf("failed to build job object meta: %v", err)
	}

	err = b.buildRjSpec(ctx, workload, job)
	if err != nil {
		return nil, fmt.Errorf("failed to build job spec: %v", err)
	}

	builderLog.Info("completed building Runai job")
	if options.Owner != nil {
		err = controllerutil.SetControllerReference(options.Owner, job, b.k8sClient.Scheme())
		if err != nil {
			return nil, fmt.Errorf("failed to set controller reference for runai job, %v", err)
		}
	}

	return job, nil
}

func (b *Builder) buildRjTypeMeta(_ context.Context, job *rjv1.RunaiJob) {
	job.TypeMeta = metav1.TypeMeta{
		APIVersion: rjv1.GroupVersion.String(),
		Kind:       constants.KindRunaiJob,
	}
}

func (b *Builder) buildRjObjectMeta(ctx context.Context, workload *jobs.JobWorkload, job *rjv1.RunaiJob) error {

	err := b.BuildObjectMeta(ctx, &job.ObjectMeta)
	if err != nil {
		return err
	}

	err = workloadbuilders.SetObjectMetaLabels(ctx, workload.Kind, constants.KindRunaiJob, workload.Name, &job.ObjectMeta)
	if err != nil {
		return err
	}
	b.setWorkloadAnnotations(workload, job)
	return nil
}

func (b *Builder) setWorkloadAnnotations(workload *jobs.JobWorkload, job *rjv1.RunaiJob) {
	if workload.Annotations != nil {
		for k, v := range workload.Annotations {
			job.Annotations[k] = v
		}
	}
}

func (b *Builder) buildRjSpec(ctx context.Context, workload *jobs.JobWorkload, job *rjv1.RunaiJob) error {
	log := logger.FromContext(ctx)

	var err error

	workloadName := workload.GetChildName()

	job.Spec.Completions = ptr.To(workload.Spec.Completions.Deref32default(constants.DefaultCompletions))
	job.Spec.Parallelism = ptr.To(workload.Spec.Parallelism.Deref32default(constants.DefaultParallelism))
	if !b.IsWorkloadActive(b.MergedSpec.StateFields) {
		// runai suspend option is false as default
		job.Spec.Suspend = ptr.To(true)
	}

	if workload.Spec.AutoDeletionTimeAfterCompletionSeconds != nil {
		job.Spec.AutoDeletionTimeAfterCompletionSeconds = ptr.To(workload.Spec.AutoDeletionTimeAfterCompletionSeconds.Deref32())
	}
	defaultBackoffLimit := int32(constants.DefaultBackoffLimit)
	if workload.TypeMeta.Kind != constants.KindTrainingWorkload {
		defaultBackoffLimit = 0
	}
	job.Spec.BackoffLimit = ptr.To(workload.Spec.BackoffLimit.Deref32default(defaultBackoffLimit))

	projectName, err := workloadproject.GetProjectFromContext(ctx)
	if err != nil {
		return fmt.Errorf("project name not found in context")
	}
	// The project is a kai.resources object whose limits live in annotations.
	project := &kaiorg.Project{}
	if err = b.k8sClient.Get(ctx, client.ObjectKey{Name: projectName}, project); err != nil {
		return fmt.Errorf("could not get project %s. error: %s", projectName, err)
	}
	limits := limitsFromProject(project)
	err = b.setJobMaxIdleDuration(ctx, b.k8sClient, limits, workload, job)
	if err != nil {
		return fmt.Errorf("failed to set max idle duration for runaijob '%s', error: %s", job.Name, err.Error())
	}
	b.setInteractiveJobTimeLimit(limits, workload, job)
	b.setTrainingJobTimeLimit(limits, workload, job)

	log.V(4).Info("Creating node affinity for job")
	err = nodepoolbuilders.CreateNodeAffinityForNodePools(
		ctx,
		b.k8sClient,
		workload.Spec.NodePools.Deref(),
		workload.Spec.NodePool.Deref(),
		&job.Spec.Template.Spec.Affinity,
	)
	if err != nil {
		builderLog.Error(err, fmt.Sprintf("failed to create node affinity for runaijob %s.", job.Name))
		return err
	}

	// if the user provided node affinity, we need to merge it with the one we created
	workloadbuilders.MergeNodeAffinity(&workload.Spec.NodeAffinityFields, &job.Spec.Template.Spec.Affinity)

	err = workloadbuilders.CreatePodAffinity(
		workload.Spec.PodAffinity.Deref(),
		workload.Spec.PodAffinityTopology.Deref(),
		workload.Spec.PodAffinitySchedulingRule.Deref(),
		&job.Spec.Template.Spec.Affinity,
		workloadName,
		false,
	)

	if err != nil {
		builderLog.Error(err, fmt.Sprintf("failed to create pod affinity for runaijob %s.", job.Name))
		return err
	}

	err = GetSpecObjectMeta(ctx, b.k8sClient, workload, &job.Spec.Template, b.podLabels, b.podAnnotations)
	if err != nil {
		return err
	}

	err = workloadbuilders.CreateEphemeralVolumes(b.SharedFields.Pvcs, &job.Spec.Template.Spec.Volumes, workloadName)
	if err != nil {
		return err
	}

	if b.MergedSpec.TerminateAfterPreemption.Deref() {
		job.Spec.TerminateAfterPreemption = ptr.To(true)
	}

	return b.BuildPodTemplateSpec(ctx, b.k8sClient)
}

func (b *Builder) setJobMaxIdleDuration(ctx context.Context, k8sClient client.Client,
	limits projectLimits, workload *jobs.JobWorkload, job *rjv1.RunaiJob) error {
	if workload.TypeMeta.Kind == constants.KindInteractiveWorkload {
		preemptibilityEnabled := workload.Spec.Preemptibility.Deref() == field_groups.Preemptible

		isPreemptible, err := common.IsPreemptible(ctx, k8sClient, workload.Spec.PriorityClassName.Deref(),
			preemptibilityEnabled, workload.Kind)
		if err != nil {
			return fmt.Errorf("failed to determine if the job is preemptible, error: %s", err.Error())
		}

		if isPreemptible {
			if limits.interactivePreemptibleMaxIdleSecs != 0 {
				job.Spec.MaxIdleDurationSeconds = ptr.To(limits.interactivePreemptibleMaxIdleSecs)
			}
		} else {
			if limits.interactiveMaxIdleDurationSecs != 0 {
				job.Spec.MaxIdleDurationSeconds = ptr.To(limits.interactiveMaxIdleDurationSecs)
			}
		}
	} else if workload.TypeMeta.Kind == constants.KindTrainingWorkload {
		if limits.trainingMaxIdleDurationSecs != 0 {
			job.Spec.MaxIdleDurationSeconds = ptr.To(limits.trainingMaxIdleDurationSecs)
		}
	}

	return nil
}

func (b *Builder) setInteractiveJobTimeLimit(limits projectLimits, workload *jobs.JobWorkload, job *rjv1.RunaiJob) {
	if workload.TypeMeta.Kind == constants.KindInteractiveWorkload {
		if limits.interactiveTimeLimitSecs != 0 {
			job.Spec.ActiveDeadlineSeconds = ptr.To(limits.interactiveTimeLimitSecs)
		}
	}
}

func (b *Builder) setTrainingJobTimeLimit(limits projectLimits, workload *jobs.JobWorkload, job *rjv1.RunaiJob) {
	if workload.TypeMeta.Kind == constants.KindTrainingWorkload {
		if limits.trainingTimeLimitSecs != 0 {
			job.Spec.ActiveDeadlineSeconds = ptr.To(limits.trainingTimeLimitSecs)
		}
	}
}

type TestRunaiJobBuilder struct {
}

func (self *TestRunaiJobBuilder) ToYaml(objects []client.Object, w *bytes.Buffer) error {
	if len(objects) != 1 {
		return fmt.Errorf("expected 1 object, got %d", len(objects))
	}
	job, ok := objects[0].(*rjv1.RunaiJob)
	if !ok {
		return fmt.Errorf("expected *runaiv1.RunaiJob, got %T", objects[0])
	}
	job.SetCreationTimestamp(common.TestFixedDate)
	job.SetDeletionTimestamp(&common.TestFixedDate)

	marshalled, err := yaml.Marshal(objects)

	if err != nil {
		return err
	}

	_, err = w.Write(marshalled)
	if err != nil {
		return err
	}
	return nil
}
