package describe

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/describe/displayer"
	nimclient "github.com/run-ai/runai/runai-clients/nim-client/generated"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

// runtimeSpecExtractor defines an interface for extracting runtime fields from different spec types
// All spec types (Training, Workspace, Distributed, Inference) have similar structures
type runtimeSpecExtractor interface {
	extractImagePullPolicy(runtimeBuilder *displayer.RuntimeBuilder)
	extractImagePullSecrets(runtimeBuilder *displayer.RuntimeBuilder)
	extractWorkingDir(runtimeBuilder *displayer.RuntimeBuilder)
	extractNodeType(runtimeBuilder *displayer.RuntimeBuilder)
	extractSecurity(runtimeBuilder *displayer.RuntimeBuilder)
	extractTolerations(runtimeBuilder *displayer.RuntimeBuilder)
	extractReadinessProbe(runtimeBuilder *displayer.RuntimeBuilder)
}

type imageProvider interface {
	GetImageOk() (*string, bool)
}

// trainingSpecAdapter adapts TrainingSpecSpec to runtimeSpecExtractor
type trainingSpecAdapter struct {
	spec *openapi.TrainingSpecSpec
}

func (a *trainingSpecAdapter) extractImagePullPolicy(runtimeBuilder *displayer.RuntimeBuilder) {
	if a.spec.ImagePullPolicy.IsSet() {
		runtimeBuilder.ImagePullPolicy(string(a.spec.GetImagePullPolicy()))
	}
}

func (a *trainingSpecAdapter) extractImagePullSecrets(runtimeBuilder *displayer.RuntimeBuilder) {
	if len(a.spec.ImagePullSecrets) > 0 {
		secrets := make([]string, 0, len(a.spec.ImagePullSecrets))
		for _, secret := range a.spec.ImagePullSecrets {
			if secret.Name != nil && *secret.Name != "" {
				secrets = append(secrets, *secret.Name)
			}
		}
		if len(secrets) > 0 {
			runtimeBuilder.ImagePullSecrets(secrets)
		}
	}
}

func (a *trainingSpecAdapter) extractWorkingDir(runtimeBuilder *displayer.RuntimeBuilder) {
	if a.spec.WorkingDir.IsSet() && a.spec.WorkingDir.Get() != nil {
		runtimeBuilder.WorkingDir(*a.spec.WorkingDir.Get())
	}
}

func (a *trainingSpecAdapter) extractNodeType(runtimeBuilder *displayer.RuntimeBuilder) {
	//nolint:staticcheck // NodeType is deprecated but still supported for backward compatibility
	if a.spec.NodeType.IsSet() && a.spec.NodeType.Get() != nil {
		runtimeBuilder.NodeType(*a.spec.NodeType.Get())
	}
}

func (a *trainingSpecAdapter) extractSecurity(runtimeBuilder *displayer.RuntimeBuilder) {
	if a.spec.Security.IsSet() {
		security := a.spec.GetSecurity()
		if security.RunAsUid.IsSet() {
			runtimeBuilder.RunAsUser(security.GetRunAsUid())
		}
		if security.RunAsGid.IsSet() {
			runtimeBuilder.RunAsGroup(security.GetRunAsGid())
		}
		if len(security.Capabilities) > 0 {
			caps := make([]string, 0, len(security.Capabilities))
			for _, cap := range security.Capabilities {
				caps = append(caps, string(cap))
			}
			runtimeBuilder.Capabilities(caps)
		}
	}
}

func (a *trainingSpecAdapter) extractTolerations(runtimeBuilder *displayer.RuntimeBuilder) {
	if len(a.spec.Tolerations) > 0 {
		tolerationStrs := make([]string, 0, len(a.spec.Tolerations))
		for _, t := range a.spec.Tolerations {
			tolStr := displayer.FormatToleration(&t)
			if tolStr != "" {
				tolerationStrs = append(tolerationStrs, tolStr)
			}
		}
		if len(tolerationStrs) > 0 {
			runtimeBuilder.Tolerations(tolerationStrs)
		}
	}
}

func (a *trainingSpecAdapter) extractReadinessProbe(runtimeBuilder *displayer.RuntimeBuilder) {
	if a.spec.Probes.IsSet() {
		probes := a.spec.GetProbes()
		if probes.Readiness.IsSet() {
			readiness := probes.GetReadiness()
			probeStr := displayer.FormatProbe(&readiness)
			if probeStr != "" {
				runtimeBuilder.ReadinessProbe(probeStr)
			}
		}
	}
}

// workspaceSpecAdapter adapts WorkspaceSpecSpec to runtimeSpecExtractor
type workspaceSpecAdapter struct {
	spec *openapi.WorkspaceSpecSpec
}

func (a *workspaceSpecAdapter) extractImagePullPolicy(runtimeBuilder *displayer.RuntimeBuilder) {
	if a.spec.ImagePullPolicy.IsSet() {
		runtimeBuilder.ImagePullPolicy(string(a.spec.GetImagePullPolicy()))
	}
}

func (a *workspaceSpecAdapter) extractImagePullSecrets(runtimeBuilder *displayer.RuntimeBuilder) {
	if len(a.spec.ImagePullSecrets) > 0 {
		secrets := make([]string, 0, len(a.spec.ImagePullSecrets))
		for _, secret := range a.spec.ImagePullSecrets {
			if secret.Name != nil && *secret.Name != "" {
				secrets = append(secrets, *secret.Name)
			}
		}
		if len(secrets) > 0 {
			runtimeBuilder.ImagePullSecrets(secrets)
		}
	}
}

func (a *workspaceSpecAdapter) extractWorkingDir(runtimeBuilder *displayer.RuntimeBuilder) {
	if a.spec.WorkingDir.IsSet() && a.spec.WorkingDir.Get() != nil {
		runtimeBuilder.WorkingDir(*a.spec.WorkingDir.Get())
	}
}

func (a *workspaceSpecAdapter) extractNodeType(runtimeBuilder *displayer.RuntimeBuilder) {
	//nolint:staticcheck // NodeType is deprecated but still supported for backward compatibility
	if a.spec.NodeType.IsSet() && a.spec.NodeType.Get() != nil {
		runtimeBuilder.NodeType(*a.spec.NodeType.Get())
	}
}

func (a *workspaceSpecAdapter) extractSecurity(runtimeBuilder *displayer.RuntimeBuilder) {
	if a.spec.Security.IsSet() {
		security := a.spec.GetSecurity()
		if security.RunAsUid.IsSet() {
			runtimeBuilder.RunAsUser(security.GetRunAsUid())
		}
		if security.RunAsGid.IsSet() {
			runtimeBuilder.RunAsGroup(security.GetRunAsGid())
		}
		if len(security.Capabilities) > 0 {
			caps := make([]string, 0, len(security.Capabilities))
			for _, cap := range security.Capabilities {
				caps = append(caps, string(cap))
			}
			runtimeBuilder.Capabilities(caps)
		}
	}
}

func (a *workspaceSpecAdapter) extractTolerations(runtimeBuilder *displayer.RuntimeBuilder) {
	if len(a.spec.Tolerations) > 0 {
		tolerationStrs := make([]string, 0, len(a.spec.Tolerations))
		for _, t := range a.spec.Tolerations {
			tolStr := displayer.FormatToleration(&t)
			if tolStr != "" {
				tolerationStrs = append(tolerationStrs, tolStr)
			}
		}
		if len(tolerationStrs) > 0 {
			runtimeBuilder.Tolerations(tolerationStrs)
		}
	}
}

func (a *workspaceSpecAdapter) extractReadinessProbe(runtimeBuilder *displayer.RuntimeBuilder) {
	if a.spec.Probes.IsSet() {
		probes := a.spec.GetProbes()
		if probes.Readiness.IsSet() {
			readiness := probes.GetReadiness()
			probeStr := displayer.FormatProbe(&readiness)
			if probeStr != "" {
				runtimeBuilder.ReadinessProbe(probeStr)
			}
		}
	}
}

// distributedSpecAdapter adapts DistributedSpecSpec to runtimeSpecExtractor
type distributedSpecAdapter struct {
	spec *openapi.DistributedSpecSpec
}

func (a *distributedSpecAdapter) extractImagePullPolicy(runtimeBuilder *displayer.RuntimeBuilder) {
	if a.spec.ImagePullPolicy.IsSet() {
		runtimeBuilder.ImagePullPolicy(string(a.spec.GetImagePullPolicy()))
	}
}

func (a *distributedSpecAdapter) extractImagePullSecrets(runtimeBuilder *displayer.RuntimeBuilder) {
	if len(a.spec.ImagePullSecrets) > 0 {
		secrets := make([]string, 0, len(a.spec.ImagePullSecrets))
		for _, secret := range a.spec.ImagePullSecrets {
			if secret.Name != nil && *secret.Name != "" {
				secrets = append(secrets, *secret.Name)
			}
		}
		if len(secrets) > 0 {
			runtimeBuilder.ImagePullSecrets(secrets)
		}
	}
}

func (a *distributedSpecAdapter) extractWorkingDir(runtimeBuilder *displayer.RuntimeBuilder) {
	if a.spec.WorkingDir.IsSet() && a.spec.WorkingDir.Get() != nil {
		runtimeBuilder.WorkingDir(*a.spec.WorkingDir.Get())
	}
}

func (a *distributedSpecAdapter) extractNodeType(runtimeBuilder *displayer.RuntimeBuilder) {
	//nolint:staticcheck // NodeType is deprecated but still supported for backward compatibility
	if a.spec.NodeType.IsSet() && a.spec.NodeType.Get() != nil {
		runtimeBuilder.NodeType(*a.spec.NodeType.Get())
	}
}

func (a *distributedSpecAdapter) extractSecurity(runtimeBuilder *displayer.RuntimeBuilder) {
	if a.spec.Security.IsSet() {
		security := a.spec.GetSecurity()
		if security.RunAsUid.IsSet() {
			runtimeBuilder.RunAsUser(security.GetRunAsUid())
		}
		if security.RunAsGid.IsSet() {
			runtimeBuilder.RunAsGroup(security.GetRunAsGid())
		}
		if len(security.Capabilities) > 0 {
			caps := make([]string, 0, len(security.Capabilities))
			for _, cap := range security.Capabilities {
				caps = append(caps, string(cap))
			}
			runtimeBuilder.Capabilities(caps)
		}
	}
}

func (a *distributedSpecAdapter) extractTolerations(runtimeBuilder *displayer.RuntimeBuilder) {
	if len(a.spec.Tolerations) > 0 {
		tolerationStrs := make([]string, 0, len(a.spec.Tolerations))
		for _, t := range a.spec.Tolerations {
			tolStr := displayer.FormatToleration(&t)
			if tolStr != "" {
				tolerationStrs = append(tolerationStrs, tolStr)
			}
		}
		if len(tolerationStrs) > 0 {
			runtimeBuilder.Tolerations(tolerationStrs)
		}
	}
}

func (a *distributedSpecAdapter) extractReadinessProbe(runtimeBuilder *displayer.RuntimeBuilder) {
	if a.spec.Probes.IsSet() {
		probes := a.spec.GetProbes()
		if probes.Readiness.IsSet() {
			readiness := probes.GetReadiness()
			probeStr := displayer.FormatProbe(&readiness)
			if probeStr != "" {
				runtimeBuilder.ReadinessProbe(probeStr)
			}
		}
	}
}

// inferenceSpecAdapter adapts InferenceSpecSpec to runtimeSpecExtractor
type inferenceSpecAdapter struct {
	spec *openapi.InferenceSpecSpec
}

func (a *inferenceSpecAdapter) extractImagePullPolicy(runtimeBuilder *displayer.RuntimeBuilder) {
	if a.spec.ImagePullPolicy.IsSet() {
		runtimeBuilder.ImagePullPolicy(string(a.spec.GetImagePullPolicy()))
	}
}

func (a *inferenceSpecAdapter) extractImagePullSecrets(runtimeBuilder *displayer.RuntimeBuilder) {
	if len(a.spec.ImagePullSecrets) > 0 {
		secrets := make([]string, 0, len(a.spec.ImagePullSecrets))
		for _, secret := range a.spec.ImagePullSecrets {
			if secret.Name != nil && *secret.Name != "" {
				secrets = append(secrets, *secret.Name)
			}
		}
		if len(secrets) > 0 {
			runtimeBuilder.ImagePullSecrets(secrets)
		}
	}
}

func (a *inferenceSpecAdapter) extractWorkingDir(runtimeBuilder *displayer.RuntimeBuilder) {
	if a.spec.WorkingDir.IsSet() && a.spec.WorkingDir.Get() != nil {
		runtimeBuilder.WorkingDir(*a.spec.WorkingDir.Get())
	}
}

func (a *inferenceSpecAdapter) extractNodeType(runtimeBuilder *displayer.RuntimeBuilder) {
	//nolint:staticcheck // NodeType is deprecated but still supported for backward compatibility
	if a.spec.NodeType.IsSet() && a.spec.NodeType.Get() != nil {
		runtimeBuilder.NodeType(*a.spec.NodeType.Get())
	}
}

func (a *inferenceSpecAdapter) extractSecurity(runtimeBuilder *displayer.RuntimeBuilder) {
	if a.spec.Security.IsSet() {
		security := a.spec.GetSecurity()
		if security.RunAsUid.IsSet() {
			runtimeBuilder.RunAsUser(security.GetRunAsUid())
		}
		if security.RunAsGid.IsSet() {
			runtimeBuilder.RunAsGroup(security.GetRunAsGid())
		}
		if len(security.Capabilities) > 0 {
			caps := make([]string, 0, len(security.Capabilities))
			for _, cap := range security.Capabilities {
				caps = append(caps, string(cap))
			}
			runtimeBuilder.Capabilities(caps)
		}
	}
}

func (a *inferenceSpecAdapter) extractTolerations(runtimeBuilder *displayer.RuntimeBuilder) {
	if len(a.spec.Tolerations) > 0 {
		tolerationStrs := make([]string, 0, len(a.spec.Tolerations))
		for _, t := range a.spec.Tolerations {
			tolStr := displayer.FormatToleration(&t)
			if tolStr != "" {
				tolerationStrs = append(tolerationStrs, tolStr)
			}
		}
		if len(tolerationStrs) > 0 {
			runtimeBuilder.Tolerations(tolerationStrs)
		}
	}
}

func (a *inferenceSpecAdapter) extractReadinessProbe(runtimeBuilder *displayer.RuntimeBuilder) {
	if a.spec.Probes.IsSet() {
		probes := a.spec.GetProbes()
		if probes.Readiness.IsSet() {
			readiness := probes.GetReadiness()
			probeStr := displayer.FormatProbe(&readiness)
			if probeStr != "" {
				runtimeBuilder.ReadinessProbe(probeStr)
			}
		}
	}
}

type nimSpecAdapter struct {
	spec *nimclient.NimServiceSpec
}

func (a *nimSpecAdapter) extractImagePullPolicy(runtimeBuilder *displayer.RuntimeBuilder) {
	if a.spec.ImagePullPolicy.IsSet() {
		runtimeBuilder.ImagePullPolicy(string(a.spec.GetImagePullPolicy()))
	}
}

func (a *nimSpecAdapter) extractImagePullSecrets(runtimeBuilder *displayer.RuntimeBuilder) {
	if len(a.spec.ImagePullSecrets) > 0 {
		secrets := make([]string, 0, len(a.spec.ImagePullSecrets))
		for _, secret := range a.spec.ImagePullSecrets {
			if secret.Name != nil && *secret.Name != "" {
				secrets = append(secrets, *secret.Name)
			}
		}
		if len(secrets) > 0 {
			runtimeBuilder.ImagePullSecrets(secrets)
		}
	}
}

func (a *nimSpecAdapter) extractWorkingDir(runtimeBuilder *displayer.RuntimeBuilder) {}

func (a *nimSpecAdapter) extractNodeType(runtimeBuilder *displayer.RuntimeBuilder) {}

func (a *nimSpecAdapter) extractSecurity(runtimeBuilder *displayer.RuntimeBuilder) {
	if a.spec.Security.IsSet() {
		security := a.spec.GetSecurity()
		if security.RunAsUid.IsSet() {
			runtimeBuilder.RunAsUser(security.GetRunAsUid())
		}
		if security.RunAsGid.IsSet() {
			runtimeBuilder.RunAsGroup(security.GetRunAsGid())
		}
	}
}

func (a *nimSpecAdapter) extractTolerations(runtimeBuilder *displayer.RuntimeBuilder) {
	if len(a.spec.Tolerations) > 0 {
		tolerationStrs := make([]string, 0, len(a.spec.Tolerations))
		for _, t := range a.spec.Tolerations {
			key := t.GetKey()
			value := t.GetValue()
			var operator *openapi.TolerationOperator
			var effect *openapi.TolerationEffect
			var seconds *openapi.NullableInt32
			if t.Operator.IsSet() {
				op := openapi.TolerationOperator(t.GetOperator())
				operator = &op
			}
			if t.Effect.IsSet() {
				ef := openapi.TolerationEffect(t.GetEffect())
				effect = &ef
			}
			if t.Seconds.IsSet() {
				sec := t.GetSeconds()
				seconds = openapi.NewNullableInt32(&sec)
			}
			workloadsTol := openapi.Toleration{
				Key:   *openapi.NewNullableString(&key),
				Value: *openapi.NewNullableString(&value),
			}
			if seconds != nil {
				workloadsTol.Seconds = *seconds
			}
			if operator != nil {
				workloadsTol.Operator = *openapi.NewNullableTolerationOperator(operator)
			}
			if effect != nil {
				workloadsTol.Effect = *openapi.NewNullableTolerationEffect(effect)
			}
			tolStr := displayer.FormatToleration(&workloadsTol)
			if tolStr != "" {
				tolerationStrs = append(tolerationStrs, tolStr)
			}
		}
		if len(tolerationStrs) > 0 {
			runtimeBuilder.Tolerations(tolerationStrs)
		}
	}
}

func (a *nimSpecAdapter) extractReadinessProbe(runtimeBuilder *displayer.RuntimeBuilder) {
	if a.spec.Probes.IsSet() {
		probes := a.spec.GetProbes()
		if probes.Readiness.IsSet() {
			readiness := probes.GetReadiness()
			workloadsProbe := openapi.Probe{}
			if readiness.InitialDelaySeconds.IsSet() {
				val := readiness.GetInitialDelaySeconds()
				workloadsProbe.InitialDelaySeconds = *openapi.NewNullableInt32(&val)
			}
			if readiness.PeriodSeconds.IsSet() {
				val := readiness.GetPeriodSeconds()
				workloadsProbe.PeriodSeconds = *openapi.NewNullableInt32(&val)
			}
			if readiness.TimeoutSeconds.IsSet() {
				val := readiness.GetTimeoutSeconds()
				workloadsProbe.TimeoutSeconds = *openapi.NewNullableInt32(&val)
			}
			if readiness.SuccessThreshold.IsSet() {
				val := readiness.GetSuccessThreshold()
				workloadsProbe.SuccessThreshold = *openapi.NewNullableInt32(&val)
			}
			if readiness.FailureThreshold.IsSet() {
				val := readiness.GetFailureThreshold()
				workloadsProbe.FailureThreshold = *openapi.NewNullableInt32(&val)
			}
			if readiness.Handler.IsSet() {
				nimHandler := readiness.GetHandler()
				if nimHandler.HasHttpGet() {
					nimHttpGet := nimHandler.GetHttpGet()
					workloadsHandler := openapi.NewProbeHandler()
					workloadsHttpGet := openapi.NewProbeHandlerHttpGet()
					if nimHttpGet.Path.IsSet() {
						path := nimHttpGet.GetPath()
						workloadsHttpGet.Path = *openapi.NewNullableString(&path)
					}
					if nimHttpGet.Port.IsSet() {
						port := nimHttpGet.GetPort()
						workloadsHttpGet.Port = *openapi.NewNullableInt32(&port)
					}
					workloadsHandler.SetHttpGet(*workloadsHttpGet)
					workloadsProbe.Handler = *openapi.NewNullableProbeHandler(workloadsHandler)
				}
			}
			probeStr := displayer.FormatProbe(&workloadsProbe)
			if probeStr != "" {
				runtimeBuilder.ReadinessProbe(probeStr)
			}
		}
	}
}

// extractRuntimeFromSpec is the single implementation that works for all spec types
func extractRuntimeFromSpec(
	runtimeBuilder *displayer.RuntimeBuilder,
	imageSource imageProvider,
	extractor runtimeSpecExtractor,
) *displayer.RuntimeBuilder {
	if image, found := imageSource.GetImageOk(); found && image != nil && *image != "" {
		runtimeBuilder.AddImage(*image)
	}
	extractor.extractImagePullPolicy(runtimeBuilder)
	extractor.extractImagePullSecrets(runtimeBuilder)
	extractor.extractWorkingDir(runtimeBuilder)
	extractor.extractNodeType(runtimeBuilder)
	extractor.extractSecurity(runtimeBuilder)
	extractor.extractTolerations(runtimeBuilder)
	extractor.extractReadinessProbe(runtimeBuilder)
	return runtimeBuilder
}
