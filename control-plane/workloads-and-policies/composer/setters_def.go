package composer

import (
	clusterApi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	corev1 "k8s.io/api/core/v1"
)

type FnStringSetter func(*string, any, any) *workloadsApi.Error
type FnArraySetter func([]string, any, any) *workloadsApi.Error
type FnBooleanSetter func(*bool, any, any) *workloadsApi.Error

type IntType interface {
	int | int32 | int64
}
type FnIntSetter[T IntType] func(*T, any, any) *workloadsApi.Error

type MinMaxPair[T IntType] struct {
	Min *T
	Max *T
}

type FnMinMaxIntSetter[T IntType] func(*MinMaxPair[T], any, any) *workloadsApi.Error

type FloatType interface {
	float32 | float64
}

type FnFloatSetter[T FloatType] func(*T, any, any) *workloadsApi.Error

type FnMapToMapSetter[T mapValue] func(map[string]*T, any, any) *workloadsApi.Error

type FnMapToHostPathSetter func(map[string]*workloadsApi.HostPathInstance, any, any) *workloadsApi.Error
type FnMapToNfsSetter func(map[string]*workloadsApi.NfsInstance, any, any) *workloadsApi.Error
type FnGitSyncSetter func(*GitSyncInfo, any, any) *workloadsApi.Error
type FnMapToPvcSetter func(map[string]*workloadsApi.PvcInstance, any, any) *workloadsApi.Error
type FnS3Setter func(*S3Info, any, any) *workloadsApi.Error
type FnMapToConfigMapSetter func(map[string]*workloadsApi.ConfigMapInstance, any, any) *workloadsApi.Error
type FnMapToEmptyDirSetter func(map[string]*workloadsApi.EmptyDirInstance, any, any) *workloadsApi.Error
type FnMapToDataVolumeSetter func(map[string]*workloadsApi.DataVolumeInstance, any, any) *workloadsApi.Error
type FnMapToSecretVolumeSetter func(map[string]*workloadsApi.SecretInstance, any, any) *workloadsApi.Error
type FnMapToTolerationSetter func(map[string]*workloadsApi.Toleration, any, any) *workloadsApi.Error
type FnExposedUrlSetter func(*ExposedUrlInfo, any, any) *workloadsApi.Error
type FnMapToPortSetter func(*PortsInfo, any, any) *workloadsApi.Error

type PortAuthInfo struct {
	AuthorizationType *string
	AuthorizedUsers   []string
	AuthorizedGroups  []string
}

type ServingPortInfo struct {
	ItemId      string
	ServiceType *string
	AuthInfo    PortAuthInfo
	PortInfo    *workloadsApi.SupersetServingPort
	ReplicaType *workloadsApi.ReplicaType

	// for distributed inference
	ExposedExternally *bool
	ExposedUrl        *string
}

type FnServingPortSetter func(*ServingPortInfo, any, any) *workloadsApi.Error

type FnAutoScalingSetter func(*workloadsApi.AutoScaling, any, any) *workloadsApi.Error

type FnProbeSetter func(*workloadsApi.Probe, any, any) *workloadsApi.Error

const (
	EnvValueText            = "text"
	EnvValueConfigMap       = "configmap"
	EnvValueSecret          = "secret"
	EnvValueUserCredentials = "user-credentials"
	EnvValuePodFieldRef     = "pod-field-ref"
)

type EnvVarInfo struct {
	Name      string
	ValueType string
	ValueFrom string
	Key       string // for configmap, secrets, etc...
	Deleted   bool
}

type FnEnvVarSetter func([]EnvVarInfo, any, any) *workloadsApi.Error

type GpuSizingInfo struct {
	NumDevices     *int32
	MemoryRequest  *string
	MemoryLimit    *string
	PortionRequest *float64
	PortionLimit   *float64
}

func (gsi *GpuSizingInfo) IsGpuRequested() bool {
	return gsi.NumDevices != nil || gsi.PortionRequest != nil || gsi.PortionLimit != nil || gsi.MemoryRequest != nil || gsi.MemoryLimit != nil
}

// IsFractionalPortion reports whether the GPU portion request denotes a fractional
// (sub-1) GPU allocation. A portion of 1 (or greater) is a whole-GPU allocation and
// must not be treated as fractional, matching the native workload controller behavior.
func (gsi *GpuSizingInfo) IsFractionalPortion() bool {
	return gsi.PortionRequest != nil && *gsi.PortionRequest > 0 && *gsi.PortionRequest < 1
}

type FnGpuSizingInfo func(*GpuSizingInfo, any, any) *workloadsApi.Error

type FnRequestedPodsSetter func(*workloadsApi.RequestedPods, any, any) *workloadsApi.Error

////////////////////////////////////////////////////////////////
//  k8s setters
////////////////////////////////////////////////////////////////

type FnNodeAffinitySetter func([][]corev1.NodeSelectorRequirement, any, any) *workloadsApi.Error

type FnPodAffinitySetter func(*workloadsApi.PodAffinity, any, any) *workloadsApi.Error

// NodePoolsInfo contains the requested node pools and the actual node pools from the cluster
type NodePoolsInfo struct {
	RequestedNodePools []string
	ActualNodePools    []clusterApi.Nodepool
}

type FnNodePoolsSetter func(*NodePoolsInfo, any, any) *workloadsApi.Error
