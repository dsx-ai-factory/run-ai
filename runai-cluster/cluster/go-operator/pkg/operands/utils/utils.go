package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"slices"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	me "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/multi-error"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"
	runv1internal "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	KubernetesServiceHostEnvVarName = "KUBERNETES_SERVICE_HOST"
	fips140EnvVarName               = "fips140"
)

var allowedFips140Values = []string{"on", "only", "off"}

func ResourcesForRunaiConfig(ctx context.Context, k8sReader client.Reader,
	runaiConfig *runaiv1.RunaiConfig, resourceFunctions []operands.ResourceFunction,
	operandName string) ([]operands.ResourceOld, error) {

	type result struct {
		resource client.Object
		err      error
	}

	resultsCh := make(chan result, len(resourceFunctions))

	// Launch all resource functions concurrently
	for _, resourceFunction := range resourceFunctions {
		go func(fn operands.ResourceFunction) {
			resource, err := fn(ctx, k8sReader, runaiConfig)
			resultsCh <- result{resource: resource, err: err}
		}(resourceFunction)
	}

	operandObjects := make([]client.Object, 0, len(resourceFunctions))
	errs := me.New()

	// Collect all results
	for i := 0; i < len(resourceFunctions); i++ {
		res := <-resultsCh
		if res.err != nil {
			errs.Append(res.err)
		} else if res.resource != nil {
			operandObjects = append(operandObjects, res.resource)
		}
	}
	close(resultsCh)

	operandResources := make([]operands.ResourceOld, 0, len(operandObjects))
	for _, resource := range operandObjects {
		if resource != nil {
			operandResources = append(operandResources, operands.ResourceForOperand(resource, operandName))
		}
	}

	var retErr error
	if !errs.Empty() {
		retErr = errs
	}

	return operandResources, retErr
}

func GetEnabled(resources *[]operands.Resource) []operands.Resource {
	enabled := []operands.Resource{}
	for _, o := range *resources {
		if o.Enabled() {
			enabled = append(enabled, o)
		}
	}
	return enabled
}

func GetDisabled(resources *[]operands.Resource) []operands.Resource {
	disabled := []operands.Resource{}
	for _, o := range *resources {
		if !o.Enabled() {
			disabled = append(disabled, o)
		}
	}
	return disabled
}

func Status(ctx context.Context, resources []operands.Resource) (operands.Status, error) {
	multiErr := me.New()
	aggregatedStatus := operands.ReadyStatus()

	for _, resource := range resources {
		if !resource.Enabled() {
			continue
		}

		status := operands.ReadyStatus()
		err := resource.Status(ctx)
		if err != nil {
			multiErr.Append(err)
			status = operands.NotReadyStatus(err.Error())
			aggregatedStatus.Ready = false
			aggregatedStatus.Reasons = append(aggregatedStatus.Reasons, status.Reasons...)
		}
	}

	var retErr error
	if !multiErr.Empty() {
		retErr = multiErr
	}

	return aggregatedStatus, retErr
}

func StatusOld(ctx context.Context, k8sReader client.Reader,
	runaiConfig *runaiv1.RunaiConfig, resourceStatusFunctions []operands.ResourceStatusFunction) (operands.Status, error) {
	resourcesStatuses := []operands.Status{}
	multiErr := me.New()
	for _, resourceStatusFunction := range resourceStatusFunctions {
		status, err := resourceStatusFunction(ctx, k8sReader, runaiConfig)
		if err != nil {
			multiErr.Append(err)
			status = operands.NotReadyStatus(err.Error())
		}

		resourcesStatuses = append(resourcesStatuses, status)
	}

	aggregatedStatus := operands.ReadyStatus()
	for _, status := range resourcesStatuses {
		if !status.Ready {
			aggregatedStatus.Ready = false
			aggregatedStatus.Reasons = append(aggregatedStatus.Reasons, status.Reasons...)
		}
	}

	var retErr error
	if !multiErr.Empty() {
		retErr = multiErr
	}

	return aggregatedStatus, retErr
}

func MergeAffinities(localAffinity *v1.Affinity,
	globalAffinity *v1.Affinity,
	podAntiAffinityLabel map[string]string,
	requireDefaultPodAntiAffinity bool) *v1.Affinity {
	if localAffinity == nil {
		return globalAffinity
	}

	if globalAffinity == nil {
		return localAffinity
	}

	affinity := &v1.Affinity{}

	// If NodeAffinity is defined in localAffinity, use it; otherwise use from globalAffinity
	if localAffinity.NodeAffinity != nil {
		affinity.NodeAffinity = localAffinity.NodeAffinity
	} else if globalAffinity.NodeAffinity != nil {
		affinity.NodeAffinity = globalAffinity.NodeAffinity
	}

	// If PodAffinity is defined in localAffinity, use it; otherwise use from globalAffinity
	if localAffinity.PodAffinity != nil {
		affinity.PodAffinity = localAffinity.PodAffinity
	} else if globalAffinity.PodAffinity != nil {
		affinity.PodAffinity = globalAffinity.PodAffinity
	}

	podAntiAffinity := &v1.PodAntiAffinity{}

	if localAffinity.PodAntiAffinity != nil {
		podAntiAffinity = localAffinity.PodAntiAffinity
	} else if globalAffinity.PodAntiAffinity != nil {
		podAntiAffinity = globalAffinity.PodAntiAffinity
	} else if len(podAntiAffinityLabel) > 0 {
		podAffinityTerm := v1.PodAffinityTerm{
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: podAntiAffinityLabel,
			},
			TopologyKey: "kubernetes.io/hostname",
		}

		podAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(
			podAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
			v1.WeightedPodAffinityTerm{
				Weight:          100,
				PodAffinityTerm: podAffinityTerm,
			},
		)

		if requireDefaultPodAntiAffinity {
			podAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(
				podAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution,
				podAffinityTerm,
			)
		}
	}

	affinity.PodAntiAffinity = podAntiAffinity

	return affinity
}

func GetAffinity(runaiConfig *runaiv1.RunaiConfig) *v1.Affinity {
	global := runaiConfig.Spec.Global
	if global == nil || global.Affinity == nil {
		return nil
	}
	return global.Affinity
}

func GetTolerations(runaiConfig *runaiv1.RunaiConfig) []v1.Toleration {
	global := runaiConfig.Spec.Global
	if global == nil || len(global.Tolerations) == 0 {
		// Return nil rather than an empty slice: K8s drops the field via
		// omitempty on round-trip, so a non-nil empty slice causes a
		// false-positive DeepEqual diff every reconcile, triggering an
		// unnecessary Update that races concurrent writers on engine-config.
		return nil
	}
	return global.Tolerations
}

func GlobalSecurityContext(runaiConfig *runaiv1.RunaiConfig) *v1.SecurityContext {
	if !(runaiConfig.Spec.Internal.Global.Platform.KubernetesFlavor == runv1internal.KubernetesFlavorOpenShift) {
		ans := v1.SecurityContext(*runaiConfig.Spec.Global.SecurityContext)
		if ans.RunAsUser == nil {
			ans.RunAsUser = ptr.To[int64](10000)
		}
		return &ans
	}

	return nil
}

const (
	caVolumeName = "runai-ca"
	caMountDir   = "/etc/ssl/certs"
	caMountFile  = "runai-ca.pem"
)

func GlobalVolumes(runaiConfig *runaiv1.RunaiConfig) []v1.Volume {
	internal := runaiConfig.Spec.Internal
	volumes := []v1.Volume{}

	if internal.Global.CustomCA.Enabled != nil && *internal.Global.CustomCA.Enabled {
		secretKeySelector := internal.Global.CustomCA.Secret

		volumes = append(volumes, v1.Volume{
			Name: caVolumeName,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: secretKeySelector.Name,
					Items: []v1.KeyToPath{
						{
							Key:  secretKeySelector.Key,
							Path: caMountFile,
						},
					},
					Optional: secretKeySelector.Optional,
				},
			},
		})
	}

	return volumes
}

func GlobalVolumeMounts(runaiConfig *runaiv1.RunaiConfig) []v1.VolumeMount {
	internal := runaiConfig.Spec.Internal
	mounts := []v1.VolumeMount{}

	if internal.Global.CustomCA.Enabled != nil && *internal.Global.CustomCA.Enabled {
		mounts = append(mounts, v1.VolumeMount{
			Name:      caVolumeName,
			MountPath: path.Join(caMountDir, caMountFile),
			SubPath:   caMountFile,
		})
	}

	return mounts
}

// MergeAndOverrideLabelsAndAnnotations merges labels and annotations of 'to' and 'from'
// into 'to's respective fields
func MergeAndOverrideLabelsAndAnnotations(to, from client.Object) {
	if to.GetLabels() == nil {
		to.SetLabels(map[string]string{})
	}

	if to.GetAnnotations() == nil {
		to.SetAnnotations(map[string]string{})
	}

	if from.GetLabels() != nil {
		MergeStringStringMaps(to.GetLabels(), from.GetLabels())
	}

	if from.GetAnnotations() != nil {
		MergeStringStringMaps(to.GetAnnotations(), from.GetAnnotations())
	}
}

// mergeStringStringMaps merges map[string]string maps on top of 'to'
func MergeStringStringMaps(to, from map[string]string) {
	for k, v := range from {
		to[k] = v
	}
}

func GPUNodeAffinity(runaiConfig *runaiv1.RunaiConfig,
	localAffinity *v1.Affinity,
	additionalNodeAffinityMatchExpressions []v1.NodeSelectorRequirement,
	localRestrictScheduling bool) *v1.Affinity {
	global := runaiConfig.Spec.Global

	managedNodesNSTs := global.ManagedNodes.InclusionCriteria.NodeSelectorTerms

	if localAffinity != nil {
		return localAffinity
	} else if global.NodeAffinity.RestrictScheduling || localRestrictScheduling {
		return &v1.Affinity{
			NodeAffinity: &v1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
					NodeSelectorTerms: mergeNodeSelectorTerms(managedNodesNSTs, v1.NodeSelectorTerm{
						MatchExpressions: func() []v1.NodeSelectorRequirement {
							mex := []v1.NodeSelectorRequirement{
								{
									Key:      "node-role.kubernetes.io/runai-gpu-worker",
									Operator: v1.NodeSelectorOpExists,
								},
								{
									Key:      global.NodeAffinity.NFDGPULabel,
									Operator: v1.NodeSelectorOpIn,
									Values:   []string{"true"},
								},
							}

							if len(additionalNodeAffinityMatchExpressions) > 0 {
								mex = append(mex, additionalNodeAffinityMatchExpressions...)
							}

							return mex
						}(),
					}),
				},
			},
		}
	} else {
		return &v1.Affinity{
			NodeAffinity: &v1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
					NodeSelectorTerms: mergeNodeSelectorTerms(managedNodesNSTs, v1.NodeSelectorTerm{
						MatchExpressions: func() []v1.NodeSelectorRequirement {
							mex := []v1.NodeSelectorRequirement{
								{
									Key:      global.NodeAffinity.NFDGPULabel,
									Operator: v1.NodeSelectorOpIn,
									Values:   []string{"true"},
								},
							}

							if len(additionalNodeAffinityMatchExpressions) > 0 {
								mex = append(mex, additionalNodeAffinityMatchExpressions...)
							}

							return mex
						}(),
					}),
				},
			},
		}
	}
}

func InitMapIfNil(m map[string]string) map[string]string {
	if m == nil {
		return map[string]string{}
	} else {
		return m
	}
}

func GetGlobalImagePullSecrets(runaiConfig *runaiv1.RunaiConfig) []v1.LocalObjectReference {
	additionalImageSecrets := runaiConfig.Spec.Internal.Global.AdditionalImagePullSecrets
	secretObjArray := make([]v1.LocalObjectReference, 0, len(additionalImageSecrets))
	for _, secret := range additionalImageSecrets {
		deploymentObj := v1.LocalObjectReference{Name: secret}
		secretObjArray = append(secretObjArray, deploymentObj)
	}
	return secretObjArray
}

func FipsGodebugValue(runaiConfig *runaiv1.RunaiConfig, defaultFipsMode string) string {
	if runaiConfig == nil || runaiConfig.Spec.Internal == nil || runaiConfig.Spec.Internal.Global == nil {
		return fmt.Sprintf("%s=%s", fips140EnvVarName, defaultFipsMode)
	}

	fipsMode := runaiConfig.Spec.Internal.Global.FipsMode
	if fipsMode == "" {
		return fmt.Sprintf("%s=%s", fips140EnvVarName, defaultFipsMode)
	}

	if slices.Contains(allowedFips140Values, fipsMode) {
		return fmt.Sprintf("%s=%s", fips140EnvVarName, fipsMode)
	}

	log.Log.Error(fmt.Errorf("unsupported fips mode: %s", fipsMode), "unsupported fips mode")
	return fmt.Sprintf("%s=%s", fips140EnvVarName, defaultFipsMode)
}

// mergeNodeSelectorTerms merges the match expressions and fields of 'term' into each term in 'terms'.
// In case 'terms' is empty, a new term is created with the match expressions and fields of 'term'.
func mergeNodeSelectorTerms(terms []v1.NodeSelectorTerm, term v1.NodeSelectorTerm) []v1.NodeSelectorTerm {
	if len(terms) == 0 {
		return []v1.NodeSelectorTerm{term}
	}

	merged := make([]v1.NodeSelectorTerm, len(terms))

	for i, existingTerm := range terms {
		merged[i] = v1.NodeSelectorTerm{
			MatchExpressions: append(existingTerm.MatchExpressions, term.MatchExpressions...),
			MatchFields:      append(existingTerm.MatchFields, term.MatchFields...),
		}
	}

	return merged
}

func AddNMCRemoteClusterAPIServerURLEnvVarIfConfigured(runaiConfig *runaiv1.RunaiConfig, appendTo *[]v1.EnvVar) {
	if runaiConfig.Spec.Internal.Global.NMC.RemoteAPIServerURL != "" {
		*appendTo = append(*appendTo, v1.EnvVar{
			Name:  KubernetesServiceHostEnvVarName,
			Value: runaiConfig.Spec.Internal.Global.NMC.RemoteAPIServerURL,
		})
	}
}

func AddWorkloadInjectedContainerResourcesEnvVarsIfConfigured(globalConfig *global.Global, appendTo *[]v1.EnvVar) error {
	if globalConfig == nil || globalConfig.WorkloadInjectedContainers == nil {
		return nil
	}

	injected := globalConfig.WorkloadInjectedContainers
	configurations := []struct {
		name      string
		resources *global.WorkloadInjectedContainerResourceRequirements
	}{
		{name: constants.OAuth2ProxyResourcesEnvVar, resources: resourcesFromOAuth2Proxy(injected.OAuth2Proxy)},
		{name: constants.UsernameLoaderResourcesEnvVar, resources: resourcesFromUsernameLoader(injected.UsernameLoader)},
		{name: constants.GitSyncResourcesEnvVar, resources: resourcesFromGitSync(injected.GitSync)},
		{name: constants.S3FileSystemResourcesEnvVar, resources: resourcesFromS3FileSystem(injected.S3FileSystem)},
	}

	for _, configuration := range configurations {
		if configuration.resources == nil {
			continue
		}
		if err := configuration.resources.Validate(); err != nil {
			return fmt.Errorf("validate resources for %s: %w", configuration.name, err)
		}
		resourcesJSON, err := json.Marshal(configuration.resources)
		if err != nil {
			return fmt.Errorf("marshal resources for %s: %w", configuration.name, err)
		}
		*appendTo = append(*appendTo, v1.EnvVar{Name: configuration.name, Value: string(resourcesJSON)})
	}
	return nil
}

func resourcesFromOAuth2Proxy(config *global.OAuth2ProxyContainerConfig) *global.WorkloadInjectedContainerResourceRequirements {
	if config == nil {
		return nil
	}
	return config.Resources
}

func resourcesFromUsernameLoader(config *global.UsernameLoaderContainerConfig) *global.WorkloadInjectedContainerResourceRequirements {
	if config == nil {
		return nil
	}
	return config.Resources
}

func resourcesFromGitSync(config *global.GitSyncContainerConfig) *global.WorkloadInjectedContainerResourceRequirements {
	if config == nil {
		return nil
	}
	return config.Resources
}

func resourcesFromS3FileSystem(config *global.S3FileSystemContainerConfig) *global.WorkloadInjectedContainerResourceRequirements {
	if config == nil {
		return nil
	}
	return config.Resources
}

func GetNMCServiceAccountNameOrDefault(runaiConfig *runaiv1.RunaiConfig, def string) string {
	if runaiConfig.Spec.Internal.Global.NMC.ServiceAccountName != "" {
		return runaiConfig.Spec.Internal.Global.NMC.ServiceAccountName
	}

	return def
}

func Deref[T any](p *T) T {
	var empty T
	if p == nil {
		return empty
	}
	return *p
}
