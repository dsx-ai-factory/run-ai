/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
)

// BuildSystemNimServiceSpecFg builds the "nimServiceSpec" field group, covering
// the non-container fields of apps.nvidia.com/v1alpha1/NIMService .spec:
// storage (three flavors — hostPathStorage, nimCacheStorage, sharedMemorySizeLimit),
// expose (service + ingress enable), multiNode, scale, metrics, proxy,
// probes, runtimeClass, userID/groupID, authSecret, replicas,
// inferencePlatform, schedulerName, plus nodeSelector and tolerations
// (which reuse the runai-podSpec custom formatter because the array/map
// shapes are identical to the K8s ones).
//
// Field paths are rooted at the workload .spec (the NIMService spec selector
// "workloadSpec" points at .spec).
func BuildSystemNimServiceSpecFg() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
	fields := BuildFieldGroupFieldMap(nimServiceSpecFgFields())
	selectors := BuildItemSelectorMap(nimServiceSpecItemSelectorDefs())
	return fields, selectors
}

// BuildSystemNimServiceContainerFg builds the "nimServiceContainer" field
// group, covering the container-like fields of NIMService .spec — image,
// command/args, env, resources.
//
// Field paths are rooted at the workload .spec (the NIMService spec selector
// "allContainers" points at .spec, same anchor as workloadSpec — differing
// only in the FGs they pull in).
func BuildSystemNimServiceContainerFg() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
	fields := BuildFieldGroupFieldMap(nimServiceContainerFgFields())
	selectors := BuildItemSelectorMap(nimServiceContainerItemSelectorDefs())
	return fields, selectors
}

////////////////////////////////////////////////////////
// spec FG (non-container)
////////////////////////////////////////////////////////

func nimServiceSpecItemSelectorDefs() []SystemItemSel {
	return []SystemItemSel{
		{
			Name:  "allNodeSelectors",
			Desc:  "all node selectors",
			Query: `(.nodeSelector // {}) | to_entries[]`,
		},
		{
			Name:     "nodeSelectorByKey",
			Desc:     "all node selectors with a given key",
			Query:    `((.nodeSelector // {}) | to_entries[]) | select(.key == $ARGS.named.nodeSelectorKey)`,
			ArgsInfo: map[string]api.MappingFieldType{"nodeSelectorKey": api.MAPPINGFIELDTYPE_STRING},
		},
		{
			Name:  "allTolerations",
			Desc:  "all tolerations",
			Query: `(.tolerations // [])[]`,
		},
		{
			Name:     "tolerationByEffect",
			Desc:     "all tolerations with a given effect",
			Query:    `(.tolerations // [])[] | select(.effect == $ARGS.named.effect)`,
			ArgsInfo: map[string]api.MappingFieldType{"effect": api.MAPPINGFIELDTYPE_STRING},
			ArgsEnum: map[string][]string{"effect": {"NoSchedule", "PreferNoSchedule", "NoExecute"}},
		},
		{
			Name:     "tolerationByOperator",
			Desc:     "all tolerations with a given operator",
			Query:    `(.tolerations // [])[] | select(.operator == $ARGS.named.operator)`,
			ArgsInfo: map[string]api.MappingFieldType{"operator": api.MAPPINGFIELDTYPE_STRING},
			ArgsEnum: map[string][]string{"operator": {"Equal", "Exists"}},
		},
	}
}

func nimServiceSpecFgFields() []SystemFgField {
	keyValue := dbmodel.BuiltinFormatterKeyValue
	podSpec := SystemPodSpecFormatterName
	nimServiceSpec := SystemNimServiceSpecFormatterName

	selectorsNodeSelectors := []string{"allNodeSelectors", "nodeSelectorByKey"}
	selectorsTolerations := []string{"allTolerations", "tolerationByEffect", "tolerationByOperator"}

	return []SystemFgField{
		////////////////////////////////////////////////////////
		// spec scalars
		////////////////////////////////////////////////////////
		{Name: "authSecret", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_STRING},
		{Name: "replicas", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_INTEGER},
		{Name: "runtimeClassName", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_STRING},
		{Name: "schedulerName", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_STRING},
		{Name: "userID", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_INTEGER},
		{Name: "groupID", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_INTEGER},
		// enum-enforced via runai-nimServiceSpec
		{Name: "inferencePlatform", Formatter: nimServiceSpec, Type: api.MAPPINGFIELDTYPE_STRING},
		// podAffinity is a complex corev1.PodAffinity; we expose it as a plain
		// OBJECT for now (policy writers can SET the whole blob). Granular
		// modeling can be added later.
		{Name: "podAffinity", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_OBJECT},
		// livenessProbe / readinessProbe / startupProbe each wrap
		// {enabled, probe: corev1.Probe}. Exposed as OBJECTs for now.
		{Name: "livenessProbe", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{Name: "readinessProbe", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{Name: "startupProbe", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_OBJECT},

		////////////////////////////////////////////////////////
		// spec.storage – object-level fields (set entire storage type at once)
		//
		// hostPathStorage and nimCacheStorage both include the readOnly flag,
		// which maps to the same storage.readOnly path. This is intentional:
		// setting readOnly via either field has the same effect. Policy writers
		// should be aware that if both storage types are (unusually) configured,
		// setting readOnly on one will affect the other.
		////////////////////////////////////////////////////////
		{Name: "hostPathStorage", Formatter: nimServiceSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{Name: "nimCacheStorage", Formatter: nimServiceSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},

		// sharedMemorySizeLimit — orthogonal to host-path / nimCache; tunes
		// the emptyDir shared-memory volume the NIM uses for I/O.
		{Name: "sharedMemorySizeLimit", Formatter: keyValue, SpecRef: convutils.Ref("storage.sharedMemorySizeLimit"), Type: api.MAPPINGFIELDTYPE_STRING},

		////////////////////////////////////////////////////////
		// spec.expose.service
		////////////////////////////////////////////////////////
		// exposeService-Type is enum-enforced via runai-nimServiceSpec
		{Name: "exposeService-Type", Formatter: nimServiceSpec, Type: api.MAPPINGFIELDTYPE_STRING},
		{Name: "exposeService-Name", Formatter: keyValue, SpecRef: convutils.Ref("expose.service.name"), Type: api.MAPPINGFIELDTYPE_STRING},
		{Name: "exposeService-Port", Formatter: keyValue, SpecRef: convutils.Ref("expose.service.port"), Type: api.MAPPINGFIELDTYPE_INTEGER},
		{Name: "exposeService-GrpcPort", Formatter: keyValue, SpecRef: convutils.Ref("expose.service.grpcPort"), Type: api.MAPPINGFIELDTYPE_INTEGER},
		{Name: "exposeService-MetricsPort", Formatter: keyValue, SpecRef: convutils.Ref("expose.service.metricsPort"), Type: api.MAPPINGFIELDTYPE_INTEGER},
		{Name: "exposeService-Annotations", Formatter: keyValue, SpecRef: convutils.Ref("expose.service.annotations"), Type: api.MAPPINGFIELDTYPE_OBJECT},

		////////////////////////////////////////////////////////
		// spec.expose.ingress
		////////////////////////////////////////////////////////
		{Name: "exposeIngress-Enabled", Formatter: keyValue, SpecRef: convutils.Ref("expose.ingress.enabled"), Type: api.MAPPINGFIELDTYPE_BOOLEAN},
		{Name: "exposeIngress-Annotations", Formatter: keyValue, SpecRef: convutils.Ref("expose.ingress.annotations"), Type: api.MAPPINGFIELDTYPE_OBJECT},

		////////////////////////////////////////////////////////
		// spec.multiNode
		////////////////////////////////////////////////////////
		// enum-enforced via runai-nimServiceSpec
		{Name: "multiNode-BackendType", Formatter: nimServiceSpec, Type: api.MAPPINGFIELDTYPE_STRING},
		{Name: "multiNode-PipelineParallelism", Formatter: keyValue, SpecRef: convutils.Ref("multiNode.parallelism.pipeline"), Type: api.MAPPINGFIELDTYPE_INTEGER},
		{Name: "multiNode-TensorParallelism", Formatter: keyValue, SpecRef: convutils.Ref("multiNode.parallelism.tensor"), Type: api.MAPPINGFIELDTYPE_INTEGER},
		{Name: "multiNode-MpiStartTimeout", Formatter: keyValue, SpecRef: convutils.Ref("multiNode.mpi.mpiStartTimeout"), Type: api.MAPPINGFIELDTYPE_INTEGER},

		////////////////////////////////////////////////////////
		// spec.scale (Autoscaling + HPA)
		////////////////////////////////////////////////////////
		{Name: "scale-Enabled", Formatter: keyValue, SpecRef: convutils.Ref("scale.enabled"), Type: api.MAPPINGFIELDTYPE_BOOLEAN},
		{Name: "scale-Annotations", Formatter: keyValue, SpecRef: convutils.Ref("scale.annotations"), Type: api.MAPPINGFIELDTYPE_OBJECT},
		{Name: "scale-HpaMinReplicas", Formatter: keyValue, SpecRef: convutils.Ref("scale.hpa.minReplicas"), Type: api.MAPPINGFIELDTYPE_INTEGER},
		{Name: "scale-HpaMaxReplicas", Formatter: keyValue, SpecRef: convutils.Ref("scale.hpa.maxReplicas"), Type: api.MAPPINGFIELDTYPE_INTEGER},

		////////////////////////////////////////////////////////
		// spec.metrics
		////////////////////////////////////////////////////////
		{Name: "metrics-Enabled", Formatter: keyValue, SpecRef: convutils.Ref("metrics.enabled"), Type: api.MAPPINGFIELDTYPE_BOOLEAN},
		{Name: "metricsServiceMonitor-Annotations", Formatter: keyValue, SpecRef: convutils.Ref("metrics.serviceMonitor.annotations"), Type: api.MAPPINGFIELDTYPE_OBJECT},
		{Name: "metricsServiceMonitor-AdditionalLabels", Formatter: keyValue, SpecRef: convutils.Ref("metrics.serviceMonitor.additionalLabels"), Type: api.MAPPINGFIELDTYPE_OBJECT},
		{Name: "metricsServiceMonitor-Interval", Formatter: keyValue, SpecRef: convutils.Ref("metrics.serviceMonitor.interval"), Type: api.MAPPINGFIELDTYPE_STRING},
		{Name: "metricsServiceMonitor-ScrapeTimeout", Formatter: keyValue, SpecRef: convutils.Ref("metrics.serviceMonitor.scrapeTimeout"), Type: api.MAPPINGFIELDTYPE_STRING},

		////////////////////////////////////////////////////////
		// spec.proxy
		////////////////////////////////////////////////////////
		{Name: "proxy-HttpProxy", Formatter: keyValue, SpecRef: convutils.Ref("proxy.httpProxy"), Type: api.MAPPINGFIELDTYPE_STRING},
		{Name: "proxy-HttpsProxy", Formatter: keyValue, SpecRef: convutils.Ref("proxy.httpsProxy"), Type: api.MAPPINGFIELDTYPE_STRING},
		{Name: "proxy-NoProxy", Formatter: keyValue, SpecRef: convutils.Ref("proxy.noProxy"), Type: api.MAPPINGFIELDTYPE_STRING},
		{Name: "proxy-CertConfigMap", Formatter: keyValue, SpecRef: convutils.Ref("proxy.certConfigMap"), Type: api.MAPPINGFIELDTYPE_STRING},

		////////////////////////////////////////////////////////
		// spec.nodeSelector  (reuse runai-podSpec)
		////////////////////////////////////////////////////////
		{Name: "nodeSelectors", Formatter: podSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{
			Name:          "nodeSelectors-Value",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("value"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsNodeSelectors,
		},

		////////////////////////////////////////////////////////
		// spec.tolerations (reuse runai-podSpec)
		////////////////////////////////////////////////////////
		{Name: "tolerations", Formatter: podSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{
			Name:          "tolerations-Effect",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("effect"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsTolerations,
		},
		{
			Name:          "tolerations-Operator",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("operator"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsTolerations,
		},
		{
			Name:          "tolerations-Key",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("key"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsTolerations,
		},
		{
			Name:          "tolerations-Value",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("value"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsTolerations,
		},
		{
			Name:          "tolerations-Seconds",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("tolerationSeconds"),
			Type:          api.MAPPINGFIELDTYPE_INTEGER,
			ItemSelectors: selectorsTolerations,
		},
	}
}

////////////////////////////////////////////////////////
// container FG (image + command/args + env + resources)
////////////////////////////////////////////////////////

func nimServiceContainerItemSelectorDefs() []SystemItemSel {
	return []SystemItemSel{
		{
			Name:  "allEnvVars",
			Desc:  "all environment variables",
			Query: `(.env // [])[]`,
		},
		{
			Name:  "envVarAsValue",
			Desc:  "all environment variables defined with a literal value",
			Query: `(.env // [])[] | select(has("value"))`,
		},
		{
			Name:  "envVarFromConfigMap",
			Desc:  "all environment variables sourced from a ConfigMap",
			Query: `(.env // [])[] | select(.valueFrom.configMapKeyRef != null)`,
		},
		{
			Name:  "envVarFromField",
			Desc:  "all environment variables sourced from a field reference",
			Query: `(.env // [])[] | select(.valueFrom.fieldRef != null)`,
		},
		{
			Name:  "envVarFromSecret",
			Desc:  "all environment variables sourced from a secret",
			Query: `(.env // [])[] | select(.valueFrom.secretKeyRef != null)`,
		},
		{
			Name:  "envVarFromResource",
			Desc:  "all environment variables sourced from a resource field",
			Query: `(.env // [])[] | select(.valueFrom.resourceFieldRef != null)`,
		},
		{
			Name:  "allResourceRequests",
			Desc:  "all resource requests",
			Query: `(.resources.requests // {}) | to_entries[]`,
		},
		{
			Name:     "resourceRequestByKey",
			Desc:     "resource request with a given resource name",
			Query:    `((.resources.requests // {}) | to_entries[]) | select(.key == $ARGS.named.resourceRequestKey)`,
			ArgsInfo: map[string]api.MappingFieldType{"resourceRequestKey": api.MAPPINGFIELDTYPE_STRING},
		},
		{
			Name:  "allResourceLimits",
			Desc:  "all resource limits",
			Query: `(.resources.limits // {}) | to_entries[]`,
		},
		{
			Name:     "resourceLimitByKey",
			Desc:     "resource limit with a given resource name",
			Query:    `((.resources.limits // {}) | to_entries[]) | select(.key == $ARGS.named.resourceLimitKey)`,
			ArgsInfo: map[string]api.MappingFieldType{"resourceLimitKey": api.MAPPINGFIELDTYPE_STRING},
		},
	}
}

func nimServiceContainerFgFields() []SystemFgField {
	keyValue := dbmodel.BuiltinFormatterKeyValue
	podSpec := SystemPodSpecFormatterName
	nimServiceSpec := SystemNimServiceSpecFormatterName

	selectorsEnvVars := []string{"allEnvVars", "envVarAsValue", "envVarFromConfigMap", "envVarFromField", "envVarFromSecret", "envVarFromResource"}
	selectorsEnvVarsAsValue := []string{"envVarAsValue"}
	selectorsEnvVarFromConfigMap := []string{"envVarFromConfigMap"}
	selectorsEnvVarFromField := []string{"envVarFromField"}
	selectorsEnvVarFromSecret := []string{"envVarFromSecret"}
	selectorsEnvVarFromResource := []string{"envVarFromResource"}
	selectorsResourceRequests := []string{"allResourceRequests", "resourceRequestByKey"}
	selectorsResourceLimits := []string{"allResourceLimits", "resourceLimitByKey"}

	return []SystemFgField{
		////////////////////////////////////////////////////////
		// spec.image (NIMService has a single Image; the "container"
		// doesn't wrap the image on its own — it's shaped like
		// spec.image.{repository,tag,pullPolicy,pullSecrets})
		////////////////////////////////////////////////////////
		{Name: "image-Repository", Formatter: keyValue, SpecRef: convutils.Ref("image.repository"), Type: api.MAPPINGFIELDTYPE_STRING},
		{Name: "image-Tag", Formatter: keyValue, SpecRef: convutils.Ref("image.tag"), Type: api.MAPPINGFIELDTYPE_STRING},
		// enum-enforced via runai-nimServiceSpec
		{Name: "image-PullPolicy", Formatter: nimServiceSpec, Type: api.MAPPINGFIELDTYPE_STRING},
		{Name: "image-PullSecrets", Formatter: keyValue, SpecRef: convutils.Ref("image.pullSecrets"), Type: api.MAPPINGFIELDTYPE_STRINGS},

		////////////////////////////////////////////////////////
		// spec.command / spec.args
		////////////////////////////////////////////////////////
		{Name: "args", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_STRINGS},
		{Name: "command", Formatter: keyValue, Type: api.MAPPINGFIELDTYPE_STRINGS},

		////////////////////////////////////////////////////////
		// spec.env (reuse runai-podSpec env formatters)
		////////////////////////////////////////////////////////
		{
			Name:          "environmentVariables-Name",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("name"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsEnvVars,
		},
		{Name: "environmentVariables-AsValue", Formatter: podSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{
			Name:          "environmentVariables-Value",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("value"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsEnvVarsAsValue,
		},
		{Name: "environmentVariables-FromConfigMap", Formatter: podSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{
			Name:          "environmentVariables-ConfigMapName",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("valueFrom.configMapKeyRef.name"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsEnvVarFromConfigMap,
		},
		{
			Name:          "environmentVariables-ConfigMapKey",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("valueFrom.configMapKeyRef.key"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsEnvVarFromConfigMap,
		},
		{
			Name:          "environmentVariables-ConfigMapOptional",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("valueFrom.configMapKeyRef.optional"),
			Type:          api.MAPPINGFIELDTYPE_BOOLEAN,
			ItemSelectors: selectorsEnvVarFromConfigMap,
		},
		{Name: "environmentVariables-FromField", Formatter: podSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{
			Name:          "environmentVariables-FieldApiVersion",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("valueFrom.fieldRef.apiVersion"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsEnvVarFromField,
		},
		{
			Name:          "environmentVariables-FieldPath",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("valueFrom.fieldRef.fieldPath"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsEnvVarFromField,
		},
		{Name: "environmentVariables-FromSecret", Formatter: podSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{
			Name:          "environmentVariables-SecretName",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("valueFrom.secretKeyRef.name"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsEnvVarFromSecret,
		},
		{
			Name:          "environmentVariables-SecretKey",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("valueFrom.secretKeyRef.key"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsEnvVarFromSecret,
		},
		{
			Name:          "environmentVariables-SecretOptional",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("valueFrom.secretKeyRef.optional"),
			Type:          api.MAPPINGFIELDTYPE_BOOLEAN,
			ItemSelectors: selectorsEnvVarFromSecret,
		},
		{Name: "environmentVariables-FromResource", Formatter: podSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{
			Name:          "environmentVariables-ResourceContainerName",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("valueFrom.resourceFieldRef.containerName"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsEnvVarFromResource,
		},
		{
			Name:          "environmentVariables-ResourceDivisor",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("valueFrom.resourceFieldRef.divisor"),
			Type:          api.MAPPINGFIELDTYPE_QUANTITY,
			ItemSelectors: selectorsEnvVarFromResource,
		},
		{
			Name:          "environmentVariables-ResourceKey",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("valueFrom.resourceFieldRef.resource"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsEnvVarFromResource,
		},

		////////////////////////////////////////////////////////
		// spec.resources  (corev1.ResourceRequirements — matches the
		// shape runai-podSpec's resource* formatters expect)
		////////////////////////////////////////////////////////
		{Name: "cpu-Limit", Formatter: keyValue, SpecRef: convutils.Ref("resources.limits.cpu"), Type: api.MAPPINGFIELDTYPE_QUANTITY},
		{Name: "memory-Limit", Formatter: keyValue, SpecRef: convutils.Ref("resources.limits.memory"), Type: api.MAPPINGFIELDTYPE_QUANTITY},
		{Name: "cpu-Request", Formatter: keyValue, SpecRef: convutils.Ref("resources.requests.cpu"), Type: api.MAPPINGFIELDTYPE_QUANTITY},
		{Name: "memory-Request", Formatter: keyValue, SpecRef: convutils.Ref("resources.requests.memory"), Type: api.MAPPINGFIELDTYPE_QUANTITY},
		{Name: "resourceRequests", Formatter: podSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{
			Name:          "resourceRequests-Quantity",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("value"),
			Type:          api.MAPPINGFIELDTYPE_QUANTITY,
			ItemSelectors: selectorsResourceRequests,
		},
		{Name: "resourceLimits", Formatter: podSpec, Type: api.MAPPINGFIELDTYPE_OBJECT},
		{
			Name:          "resourceLimits-Quantity",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("value"),
			Type:          api.MAPPINGFIELDTYPE_QUANTITY,
			ItemSelectors: selectorsResourceLimits,
		},
	}
}
