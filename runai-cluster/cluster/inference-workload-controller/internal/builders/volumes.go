package builders

import (
	"context"
	"encoding/json"

	infv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	logger "sigs.k8s.io/controller-runtime/pkg/log"

	workloadbuilders "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders"
	corev1 "k8s.io/api/core/v1"
)

var builderLog = logger.Log.WithName("builder.inference")

type VolumesWithMounts struct {
	Volumes []corev1.Volume
	Mounts  []corev1.VolumeMount
}

func (kb *KnativeBuilder) getCategorizedVolumes(ctx context.Context, workload *infv2.InferenceWorkload) (podSpecVolumes VolumesWithMounts, bypassedVolumes VolumesWithMounts, err error) {
	// inference workload should not contain s3 volumes, as these are not supported by knative -
	// the runai implementation relies on privileged=true which is prohibited by knative
	isS3VolumeSupported := false

	// if serving port auth is required, we require the addition of oauth2-proxy volume
	isOauth2ProxyRequired := isServingPortAuthRequired(&workload.Spec.RoutingAndAuthorizationFields)

	// inference workload doesn't support ephemeral PVCs
	isEphemeralPvcSupported := false

	volumes, err := workloadbuilders.GetVolumes(ctx, &workload.Spec.WorkloadFields.StorageFields, &workload.Spec.WorkloadFields.ContainerFields, &workload.Spec.WorkloadFields.NetworkingFields,
		kb.k8sClient, workload.Namespace, workload.GetChildName(), "", false, isS3VolumeSupported, isOauth2ProxyRequired)
	if err != nil {
		builderLog.Error(err, "failed to get volumes")
		return VolumesWithMounts{}, VolumesWithMounts{}, err
	}

	volumeMounts, err := workloadbuilders.GetVolumeMounts(ctx, &workload.Spec.WorkloadFields.StorageFields, &workload.Spec.WorkloadFields.ContainerFields, false, isEphemeralPvcSupported)
	if err != nil {
		builderLog.Error(err, "failed to get volume mounts")
		return VolumesWithMounts{}, VolumesWithMounts{}, err
	}
	podSpecVolumes, bypassedVolumes = categorizeVolumes(volumes, volumeMounts)

	return podSpecVolumes, bypassedVolumes, nil
}

// categorizeVolumes returns two lists of volumes and volume mounts:
// - volumes that should be added to the PodSpec
// - volumes that should be bypassed (i.e. added as annotations instead)
func categorizeVolumes(volumes []corev1.Volume, volumeMounts []corev1.VolumeMount) (podSpec VolumesWithMounts, bypassed VolumesWithMounts) {
	var bypassedVolumeSet = make(map[string]bool)

	for _, v := range volumes {
		if isValidationBypassVolume(v) {
			bypassed.Volumes = append(bypassed.Volumes, v)
			bypassedVolumeSet[v.Name] = true
		} else {
			podSpec.Volumes = append(podSpec.Volumes, v)
		}
	}

	for _, vm := range volumeMounts {
		if bypassedVolumeSet[vm.Name] {
			bypassed.Mounts = append(bypassed.Mounts, vm)
		} else {
			podSpec.Mounts = append(podSpec.Mounts, vm)
		}
	}

	return
}

func isValidationBypassVolume(volume corev1.Volume) bool {
	return volume.NFS != nil || volume.HostPath != nil
}

func createPatchAnnotationsValues(vwm VolumesWithMounts) (string, string, error) {
	if len(vwm.Volumes) == 0 || len(vwm.Mounts) == 0 {
		return "", "", nil
	}

	volumesAnnotationValue, err := json.Marshal(vwm.Volumes)
	if err != nil {
		builderLog.Error(err, "failed to marshal volumes")
		return "", "", err
	}

	mountsAnnotationValue, err := json.Marshal(vwm.Mounts)
	if err != nil {
		builderLog.Error(err, "failed to marshal volume mounts")
		return "", "", err
	}

	return string(volumesAnnotationValue), string(mountsAnnotationValue), nil
}
