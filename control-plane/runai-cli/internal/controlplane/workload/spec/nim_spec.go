/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package spec

import (
	"context"
	"errors"
	"fmt"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/specset"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/specset/parser"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/specset/schedulig/toleration"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/specset/storage/pvc"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/submit"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/spec/spectype"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
)

func NewNimSpec() *GeneralSpec {
	return &GeneralSpec{
		workloadType: spectype.NimService,
		framework:    api.NimService,
	}
}

func (gs *GeneralSpec) GenerateNimServiceSpec(ctx context.Context, clusterId string) (*openapi.NimServiceSpec, error) {
	spec := openapi.NewNimServiceSpec()
	flags := &gs.flags.Nim

	if gs.flags.Container.Image != "" {
		spec.Image.Set(&gs.flags.Container.Image)
	}
	if flags.NgcSecret != "" {
		spec.SetNgcAuthSecret(flags.NgcSecret)
	}

	isAutoscaling := flags.Autoscale || (flags.MinReplicas > 0 && flags.MaxReplicas > 0)
	if isAutoscaling {
		autoScaling := openapi.NewNimServiceSpecAllOfAutoscaling()
		if flags.MinReplicas > 0 {
			autoScaling.SetMinReplicas(flags.MinReplicas)
		}
		if flags.MaxReplicas > 0 {
			autoScaling.SetMaxReplicas(flags.MaxReplicas)
		}
		if flags.Metric != "" {
			autoScaling.SetMetric(flags.Metric)
		}
		if flags.MetricThreshold > 0 {
			autoScaling.SetMetricThreshold(flags.MetricThreshold)
		}
		if flags.ScaleWindowSeconds > 0 {
			autoScaling.SetScaleWindowSeconds(flags.ScaleWindowSeconds)
		}
		spec.SetAutoscaling(*autoScaling)
		spec.UnsetReplicas()
	} else {
		spec.SetReplicas(flags.Replicas)
	}

	if flags.ServingPort != "" {
		sp, err := submit.ParseNimServingPort(flags.ServingPort)
		if err != nil {
			return nil, err
		}
		if sp.Port.IsSet() {
			spec.SetServingPort(*sp)
		}
	}
	modelStore := openapi.NewNimServiceSpecAllOfModelStore()
	if flags.ModelPvc != "" {
		pvcFields, err := parseNimExistingPvc(flags.ModelPvc)
		if err != nil {
			return nil, err
		}
		modelStore.SetPvc(*pvcFields)
		spec.SetModelStore(*modelStore)
	} else if flags.ModelNewPvc != "" {
		pvcFields, err := parseNimNewPvc(flags.ModelNewPvc)
		if err != nil {
			return nil, err
		}
		modelStore.SetPvc(*pvcFields)
		spec.SetModelStore(*modelStore)
	} else if flags.ModelCache != "" {
		cache, err := parseNimCache(flags.ModelCache)
		if err != nil {
			return nil, err
		}
		modelStore.SetNimCache(*cache)
		spec.SetModelStore(*modelStore)
	}

	compute := openapi.NewNimServiceSpecAllOfCompute()
	err := specset.SetGenericType[openapi.GpuComputeMode, openapi.NullableGpuComputeMode](
		gs.flags.Compute.GpuComputeMode, openapi.NewGpuComputeModeFromValue,
		&compute.GpuComputeMode, openapi.NewNullableGpuComputeMode)
	if err != nil && !errors.Is(err, specset.ErrEmptyValue) {
		return nil, err
	}
	computeSet := compute.GpuComputeMode.IsSet()
	if gs.flags.Compute.GpuMemoryPortionRequest.IsSet() {
		compute.SetGpuMemoryPortionRequest(*gs.flags.Compute.GpuMemoryPortionRequest.Get())
		computeSet = true
	}
	if gs.flags.Compute.GpuMemoryPortionLimit.IsSet() {
		compute.SetGpuMemoryPortionLimit(*gs.flags.Compute.GpuMemoryPortionLimit.Get())
		computeSet = true
	}
	if gs.flags.Compute.GpuDeviceRequest.IsSet() {
		compute.SetGpuDevicesRequest(*gs.flags.Compute.GpuDeviceRequest.Get())
		computeSet = true
	}
	if gs.flags.Compute.CpuCoreRequest.IsSet() {
		compute.SetCpuCoreRequest(*gs.flags.Compute.CpuCoreRequest.Get())
		computeSet = true
	}
	if gs.flags.Compute.CpuMemoryRequest != "" {
		compute.SetCpuMemoryRequest(gs.flags.Compute.CpuMemoryRequest)
		computeSet = true
	}
	if gs.flags.Compute.CpuCoreLimit.IsSet() {
		compute.SetCpuCoreLimit(*gs.flags.Compute.CpuCoreLimit.Get())
		computeSet = true
	}
	if gs.flags.Compute.CpuMemoryLimit != "" {
		compute.SetCpuMemoryLimit(gs.flags.Compute.CpuMemoryLimit)
		computeSet = true
	}
	if gs.flags.Compute.GpuMemoryRequest != "" {
		compute.SetGpuMemoryRequest(gs.flags.Compute.GpuMemoryRequest)
		computeSet = true
	}
	if gs.flags.Compute.GpuMemoryLimit != "" {
		compute.SetGpuMemoryLimit(gs.flags.Compute.GpuMemoryLimit)
		computeSet = true
	}
	if gs.flags.Compute.GpuPortionRequest.IsSet() {
		compute.SetGpuPortionRequest(*gs.flags.Compute.GpuPortionRequest.Get())
		computeSet = true
	}
	if gs.flags.Compute.GpuPortionLimit.IsSet() {
		compute.SetGpuPortionLimit(*gs.flags.Compute.GpuPortionLimit.Get())
		computeSet = true
	}
	if computeSet {
		spec.SetCompute(*compute)
	}

	if flags.Workers > 0 {
		multiNode := openapi.NewNimServiceMultiNode()
		multiNode.SetWorkers(flags.Workers)
		spec.SetMultiNode(*multiNode)
	}

	security := openapi.NewNimServiceSpecAllOfSecurity()
	securitySet := false
	if gs.flags.Security.RunAsUid > 0 {
		security.SetRunAsUid(int64(gs.flags.Security.RunAsUid))
		securitySet = true
	}
	if gs.flags.Security.RunAsGid > 0 {
		security.SetRunAsGid(int64(gs.flags.Security.RunAsGid))
		securitySet = true
	}
	if securitySet {
		spec.SetSecurity(*security)
	}

	if flags.ReadinessProbe != "" {
		readinessProbe, err := submit.ParseReadinessProbe(flags.ReadinessProbe)
		if err != nil {
			return nil, err
		}
		probes := openapi.NewProbes()
		probes.SetReadiness(*readinessProbe)
		spec.SetProbes(*probes)
	}

	if len(gs.flags.Scheduling.Toleration) > 0 {
		tols, err := toleration.NewList(gs.flags.Scheduling.Toleration)
		if err != nil {
			return nil, err
		}
		nimTols := make([]openapi.Toleration, len(tols))
		for i, t := range tols {
			nimTol := openapi.NewToleration()
			if t.Key.IsSet() {
				nimTol.SetKey(t.GetKey())
			}
			if t.Value.IsSet() {
				nimTol.SetValue(t.GetValue())
			}
			if t.Seconds.IsSet() {
				nimTol.SetSeconds(t.GetSeconds())
			}
			if t.Operator.IsSet() {
				op := openapi.TolerationOperator(t.GetOperator())
				nimTol.SetOperator(op)
			}
			if t.Effect.IsSet() {
				ef := openapi.TolerationEffect(t.GetEffect())
				nimTol.SetEffect(ef)
			}
			nimTols[i] = *nimTol
		}
		spec.SetTolerations(nimTols)
	}

	if len(gs.flags.General.Labels) > 0 {
		labels := make([]openapi.Label, len(gs.flags.General.Labels))
		for i, l := range gs.flags.General.Labels {
			parts := strings.SplitN(l, "=", 2)
			if len(parts) == 2 {
				labels[i] = openapi.Label{Name: *openapi.NewNullableString(&parts[0]), Value: *openapi.NewNullableString(&parts[1])}
			} else {
				labels[i] = openapi.Label{Name: *openapi.NewNullableString(&parts[0])}
			}
		}
		spec.SetLabels(labels)
	}
	if len(gs.flags.General.Annotations) > 0 {
		anns := make([]openapi.Annotation, len(gs.flags.General.Annotations))
		for i, a := range gs.flags.General.Annotations {
			parts := strings.SplitN(a, "=", 2)
			if len(parts) == 2 {
				anns[i] = openapi.Annotation{Name: *openapi.NewNullableString(&parts[0]), Value: *openapi.NewNullableString(&parts[1])}
			} else {
				anns[i] = openapi.Annotation{Name: *openapi.NewNullableString(&parts[0])}
			}
		}
		spec.SetAnnotations(anns)
	}

	if len(gs.flags.Container.EnvironmentVariables) > 0 {
		envs := make([]openapi.EnvironmentVariable, len(gs.flags.Container.EnvironmentVariables))
		for i, e := range gs.flags.Container.EnvironmentVariables {
			parts := strings.SplitN(e, "=", 2)
			env := openapi.NewEnvironmentVariable()
			env.SetName(parts[0])
			if len(parts) == 2 {
				env.SetValue(parts[1])
			}
			envs[i] = *env
		}
		spec.SetEnvironmentVariables(envs)
	}

	if flags.Category != "" {
		spec.SetCategory(flags.Category)
	}
	if flags.PriorityClass != "" {
		spec.SetPriorityClass(flags.PriorityClass)
	}
	if flags.Preemptibility != "" {
		spec.SetPreemptibility(openapi.Preemptibility(flags.Preemptibility))
	}
	if len(flags.ImagePullSecrets) > 0 {
		ips := make([]openapi.ImagePullSecret, len(flags.ImagePullSecrets))
		for i, s := range flags.ImagePullSecrets {
			ips[i] = openapi.ImagePullSecret{Name: &s}
		}
		spec.SetImagePullSecrets(ips)
	}
	if gs.flags.Container.ImagePullPolicy != "" {
		spec.SetImagePullPolicy(openapi.ImagePullPolicy(gs.flags.Container.ImagePullPolicy))
	}

	return spec, nil
}

func parseNimCache(input string) (*openapi.NimCache, error) {
	mp := parser.NewLooseMapParser("model-nim-cache", mapset.NewSet[string]("name", "profile"))
	am := parser.NewAttributeMapper(input, mp)
	if err := am.Parse(); err != nil {
		return nil, err
	}

	if err := parser.CheckMandatoryFields(am, "name"); err != nil {
		return nil, err
	}

	cache := openapi.NewNimCache()

	if name := am.Get("name"); name != nil {
		cache.SetName(*name)
	}

	if profile := am.Get("profile"); profile != nil {
		cache.SetProfile(*profile)
	}

	return cache, nil
}

func parseNimExistingPvc(input string) (*openapi.NimServicePvcFields, error) {
	am := parser.NewAttributeMapper(input, parser.NewLooseMapParser(pvc.ParserNameExisting, mapset.NewSet[string](pvc.ExistingPvcMap...)))
	if err := am.Parse(); err != nil {
		return nil, err
	}

	fields := openapi.NewNimServicePvcFields()
	fields.SetExistingPvc(true)

	claimName := am.Get(pvc.ClaimName)
	if claimName == nil {
		return nil, fmt.Errorf("claimname is required for --model-existing-pvc")
	}
	fields.SetClaimName(*claimName)

	if am.Get(pvc.ReadOnly) != nil {
		fields.SetReadOnly(true)
	}

	return fields, nil
}

func parseNimNewPvc(input string) (*openapi.NimServicePvcFields, error) {
	am := pvc.NewParser(input)
	if err := am.Parse(); err != nil {
		return nil, err
	}
	if err := parser.CheckMandatoryFields(am, pvc.Size); err != nil {
		return nil, err
	}

	fields := openapi.NewNimServicePvcFields()
	fields.SetExistingPvc(false)

	if cn := am.Get(pvc.ClaimName); cn != nil {
		fields.SetClaimName(*cn)
	}

	claimInfo := openapi.NewClaimInfo()
	if size := am.Get(pvc.Size); size != nil {
		claimInfo.SetSize(*size)
	}
	if sc := am.Get(pvc.StorageClass); sc != nil {
		claimInfo.SetStorageClass(*sc)
	}

	setNimAccessMode(am, claimInfo)

	fields.SetClaimInfo(*claimInfo)

	if am.Get(pvc.ReadOnly) != nil {
		fields.SetReadOnly(true)
	}
	return fields, nil
}

func setNimAccessMode(am *parser.AttributeMapper, claimInfo *openapi.ClaimInfo) {
	accessModes := []struct {
		key    string
		setter func(*openapi.PvcAccessModes)
	}{
		{pvc.AccessModeRwo, func(am *openapi.PvcAccessModes) { am.SetReadWriteOnce(true) }},
		{pvc.AccessModeRwm, func(am *openapi.PvcAccessModes) { am.SetReadWriteMany(true) }},
		{pvc.AccessModeRom, func(am *openapi.PvcAccessModes) { am.SetReadOnlyMany(true) }},
	}

	accessMode := openapi.NewPvcAccessModes()
	hasAnyMode := false
	for _, mode := range accessModes {
		if am.Get(mode.key) != nil {
			mode.setter(accessMode)
			hasAnyMode = true
		}
	}

	if hasAnyMode {
		claimInfo.SetAccessModes(*accessMode)
	}
}
