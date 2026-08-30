package distributedinference

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/describe/displayer"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

func BuildWorkloadRuntimeConfigFromSpec(spec *openapi.DistributedInferenceLeaderWorkerSpec) *displayer.WorkloadRuntimeConfig {
	if spec == nil {
		return nil
	}

	builder := displayer.NewWorkloadRuntimeConfigBuilder()

	if spec.Image.IsSet() {
		builder.Image(spec.GetImage())
	}

	if spec.ImagePullPolicy.IsSet() {
		builder.ImagePullPolicy(string(spec.GetImagePullPolicy()))
	}

	if len(spec.ImagePullSecrets) > 0 {
		secrets := make([]string, 0, len(spec.ImagePullSecrets))
		for _, secret := range spec.ImagePullSecrets {
			if secret.Name != nil && *secret.Name != "" {
				secrets = append(secrets, *secret.Name)
			}
		}
		if len(secrets) > 0 {
			builder.ImagePullSecrets(secrets)
		}
	}

	if spec.Command.IsSet() {
		builder.Command(spec.GetCommand())
	}

	if spec.Args.IsSet() {
		builder.Arguments(spec.GetArgs())
	}

	if spec.WorkingDir.IsSet() {
		builder.WorkingDir(spec.GetWorkingDir())
	}

	if len(spec.EnvironmentVariables) > 0 {
		envVars := make(map[string]string)
		for _, env := range spec.EnvironmentVariables {
			if env.Name.IsSet() && env.Name.Get() != nil && *env.Name.Get() != "" {
				name := *env.Name.Get()
				value := ""
				if env.Value.IsSet() && env.Value.Get() != nil {
					value = *env.Value.Get()
				}
				envVars[name] = value
			}
		}
		if len(envVars) > 0 {
			builder.EnvironmentVariables(envVars)
		}
	}

	if len(spec.Labels) > 0 {
		labels := make(map[string]string)
		for _, label := range spec.Labels {
			if label.Name.IsSet() && label.Name.Get() != nil && *label.Name.Get() != "" {
				name := *label.Name.Get()
				value := ""
				if label.Value.IsSet() && label.Value.Get() != nil {
					value = *label.Value.Get()
				}
				labels[name] = value
			}
		}
		if len(labels) > 0 {
			builder.Labels(labels)
		}
	}

	//nolint:staticcheck // NodeType is deprecated but still supported for backward compatibility
	if spec.NodeType.IsSet() {
		builder.NodeType(spec.GetNodeType())
	}

	if len(spec.Tolerations) > 0 {
		tolerations := make([]string, 0, len(spec.Tolerations))
		for _, t := range spec.Tolerations {
			tolStr := displayer.FormatToleration(&t)
			if tolStr != "" {
				tolerations = append(tolerations, tolStr)
			}
		}
		if len(tolerations) > 0 {
			builder.Tolerations(tolerations)
		}
	}

	if spec.Security.IsSet() {
		security := spec.GetSecurity()

		if security.RunAsUid.IsSet() {
			builder.RunAsUser(security.GetRunAsUid())
		}

		if security.RunAsGid.IsSet() {
			builder.RunAsGroup(security.GetRunAsGid())
		}

		if len(security.Capabilities) > 0 {
			caps := make([]string, 0, len(security.Capabilities))
			for _, cap := range security.Capabilities {
				caps = append(caps, string(cap))
			}
			if len(caps) > 0 {
				builder.Capabilities(caps)
			}
		}
	}

	if spec.Probes.IsSet() {
		probes := spec.GetProbes()
		if probes.Readiness.IsSet() {
			readiness := probes.GetReadiness()
			probeStr := displayer.FormatProbe(&readiness)
			if probeStr != "" {
				builder.ReadinessProbe(probeStr)
			}
		}
	}

	if spec.Compute.IsSet() {
		compute := spec.GetCompute()
		builder.Compute(displayer.BuildComputeFromSpec(&compute))
	}

	if spec.Storage.IsSet() {
		storage := spec.GetStorage()
		builder.Storage(displayer.BuildStorageFromDistributedInferenceSpec(&storage))
	}

	return builder.BuildPtr()
}
