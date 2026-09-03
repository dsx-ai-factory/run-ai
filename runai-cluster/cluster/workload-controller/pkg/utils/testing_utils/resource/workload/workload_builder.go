/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	clustercommonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	fields "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	flds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	"k8s.io/utils/ptr"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	LowestHostPort       = 30000
	HighestHostPort      = 32767
	LowestContainerPort  = 0
	HighestContainerPort = 65536
)

// this is a builder for JobWorkload mainly used for tests
var logBuilder = logger.Log.WithName("workload.builder")

type JobWorkloadBuilder struct {
	workload *jobs.JobWorkload
}

// SetOldJupyter sets the workload to be an old jupyter workload
// this only for backward compatibility with v1 jobs
func (jobWorkloadBuilder *JobWorkloadBuilder) SetOldJupyter() *JobWorkloadBuilder {
	jobWorkloadBuilder.workload.Spec.IsJupyter = flds.NewOptionalBoolField(ptr.To(true))
	jobWorkloadBuilder.workload.Spec.IsTensorboard = flds.NewOptionalBoolField(ptr.To(false))
	jobWorkloadBuilder.WithImage("jupyter/scipy-notebook")
	jobWorkloadBuilder.workload.Labels[constants.LabelJupyter] = "true"
	// any fqdn will do
	jobWorkloadBuilder.workload.Spec.IngressUrl = flds.NewStringField("https://jupyter-notebook.runai.net")

	return jobWorkloadBuilder
}
func (jobWorkloadBuilder *JobWorkloadBuilder) SetOldTensorboard() *JobWorkloadBuilder {
	jobWorkloadBuilder.workload.Spec.IsJupyter = flds.NewOptionalBoolField(ptr.To(false))
	jobWorkloadBuilder.workload.Spec.IsTensorboard = flds.NewOptionalBoolField(ptr.To(true))
	jobWorkloadBuilder.WithImage("tensorboard/scipy-notebook")
	jobWorkloadBuilder.workload.Labels[constants.LabelTensorboard] = "true"
	// any fqdn will do
	jobWorkloadBuilder.workload.Spec.IngressUrl = flds.NewStringField("https://tesnsorboard.runai.net")

	return jobWorkloadBuilder
}

func (jobWorkloadBuilder *JobWorkloadBuilder) Build() *jobs.JobWorkload {
	return jobWorkloadBuilder.workload
}
func (jobWorkloadBuilder *JobWorkloadBuilder) WithNamespace(namespace string) *JobWorkloadBuilder {
	jobWorkloadBuilder.workload.Namespace = namespace
	return jobWorkloadBuilder
}

func (jobWorkloadBuilder *JobWorkloadBuilder) WithMasterSpec(spec *jobs.JobSpec) *JobWorkloadBuilder {
	jobWorkloadBuilder.workload.MasterSpec = spec
	return jobWorkloadBuilder
}
func (jobWorkloadBuilder *JobWorkloadBuilder) WithDefaultServiceType(serviceType string) *JobWorkloadBuilder {
	jobWorkloadBuilder.workload.Spec.ServiceType = flds.NewOptionalStringField(serviceType, "")
	return jobWorkloadBuilder
}

func (jobWorkloadBuilder *JobWorkloadBuilder) WithUid(uid types.UID) *JobWorkloadBuilder {
	jobWorkloadBuilder.workload.SetUID(uid)
	return jobWorkloadBuilder
}

func (jobWorkloadBuilder *JobWorkloadBuilder) WithName(name string) *JobWorkloadBuilder {
	jobWorkloadBuilder.workload.Name = name
	jobWorkloadBuilder.workload.Spec.Name = flds.NewOptionalStringField(name, "")
	return jobWorkloadBuilder
}

func (jobWorkloadBuilder *JobWorkloadBuilder) WithDistribution() *JobWorkloadBuilder {
	jobWorkloadBuilder.workload.TypeMeta = metav1.TypeMeta{
		Kind: constants.KindDistributedWorkload,
	}
	return jobWorkloadBuilder
}

func (jobWorkloadBuilder *JobWorkloadBuilder) WithJobType(jobType string) *JobWorkloadBuilder {
	jobWorkloadBuilder.workload.Spec.JobType = jobType
	return jobWorkloadBuilder
}

// with gpu count
func (jobWorkloadBuilder *JobWorkloadBuilder) WithGpuCount(count float64) *JobWorkloadBuilder {
	jobWorkloadBuilder.workload.Spec.Gpu = flds.NewOptionalNumberField(ptr.To(count))
	return jobWorkloadBuilder
}

// WithServiceTypes
func (jobWorkloadBuilder *JobWorkloadBuilder) WithPortServiceTypes(ports []flds.PortItemValue) *JobWorkloadBuilder {
	temp := make(map[string]flds.PortItemAndRules)
	for index, port := range ports {
		temp[fmt.Sprintf("port-%d", index)] = flds.PortItemAndRules{
			Value: port.DeepCopy(),
		}
	}
	jobWorkloadBuilder.workload.Spec.Ports = flds.NewPortsField(temp)
	return jobWorkloadBuilder
}
func (jobWorkloadBuilder *JobWorkloadBuilder) WithPorts(ports []string) *JobWorkloadBuilder {
	portsMap := make(map[string]flds.PortItemAndRules)
	for _, port := range ports {
		svcPort, err := CreateStructServicePort(port, true)

		if err != nil {
			logBuilder.Error(err, "error parsing port %s", port)
			continue
		}
		portsMap[fmt.Sprintf("port-%d", ptr.Deref(svcPort.Value.ContainerPort, 0))] = *svcPort
	}

	jobWorkloadBuilder.workload.Spec.Ports = flds.NewPortsField(portsMap)
	return jobWorkloadBuilder
}

func (jobWorkloadBuilder *JobWorkloadBuilder) WithInteractive() *JobWorkloadBuilder {
	jobWorkloadBuilder.workload.TypeMeta = metav1.TypeMeta{
		Kind: constants.KindInteractiveWorkload,
	}
	return jobWorkloadBuilder
}

func (jobWorkloadBuilder *JobWorkloadBuilder) WithTraining() *JobWorkloadBuilder {
	jobWorkloadBuilder.workload.TypeMeta = metav1.TypeMeta{
		Kind: constants.KindTrainingWorkload,
	}
	return jobWorkloadBuilder
}

func (jobWorkloadBuilder *JobWorkloadBuilder) WithInference() *JobWorkloadBuilder {
	jobWorkloadBuilder.workload.TypeMeta = metav1.TypeMeta{
		Kind: constants.KindInferenceWorkload,
	}
	return jobWorkloadBuilder
}

func (jobWorkloadBuilder *JobWorkloadBuilder) WithServingPort(port int) *JobWorkloadBuilder {
	jobWorkloadBuilder.workload.Spec.Ports = &fields.PortsField{
		Value: &map[string]fields.PortItemAndRules{
			"serving-port": {
				Value: &field_types.PortItemValue{
					ServiceType:   pointer.String(clustercommonconstants.ServiceTypeServingPort),
					ContainerPort: pointer.Int(port),
				},
			},
		},
	}
	return jobWorkloadBuilder
}

func (jobWorkloadBuilder *JobWorkloadBuilder) WithNodePools(nodepool ...string) *JobWorkloadBuilder {
	nodepoolsBuilder := strings.Builder{}
	for index, pool := range nodepool {
		nodepoolsBuilder.WriteString(fmt.Sprint(pool))
		if index < len(nodepool)-1 {
			nodepoolsBuilder.WriteString(",")
		}
	}
	jobWorkloadBuilder.workload.Spec.NodePools = flds.NewOptionalStringField(nodepoolsBuilder.String(), "")
	return jobWorkloadBuilder
}

func (jobWorkloadBuilder *JobWorkloadBuilder) WithImage(image string) *JobWorkloadBuilder {
	jobWorkloadBuilder.workload.Spec.Image = flds.NewOptionalStringField(image, "")
	return jobWorkloadBuilder
}
func (jobWorkloadBuilder *JobWorkloadBuilder) WithCommand(command string) *JobWorkloadBuilder {
	jobWorkloadBuilder.workload.Spec.Command = flds.NewOptionalStringField(command, "")
	return jobWorkloadBuilder
}
func (jobWorkloadBuilder *JobWorkloadBuilder) SetActive() *JobWorkloadBuilder {
	jobWorkloadBuilder.workload.Spec.Active = flds.NewOptionalBoolField(ptr.To(true))
	return jobWorkloadBuilder
}
func (jobWorkloadBuilder *JobWorkloadBuilder) SetSuspend() *JobWorkloadBuilder {
	jobWorkloadBuilder.workload.Spec.Active = flds.NewOptionalBoolField(ptr.To(false))
	return jobWorkloadBuilder
}
func (jobWorkloadBuilder *JobWorkloadBuilder) WithLabels(labels map[string]string) *JobWorkloadBuilder {

	for key, value := range labels {
		jobWorkloadBuilder.workload.Labels[key] = value
	}
	return jobWorkloadBuilder
}

func (JobWorkloadBuilder *JobWorkloadBuilder) WithAnnotations(annotations map[string]string) *JobWorkloadBuilder {
	if JobWorkloadBuilder.workload.Annotations == nil {
		JobWorkloadBuilder.workload.Annotations = make(map[string]string)
	}
	for key, value := range annotations {
		JobWorkloadBuilder.workload.Annotations[key] = value
	}
	return JobWorkloadBuilder
}

func (jobWorkloadBuilder *JobWorkloadBuilder) WithPVC(pvc *v1.PersistentVolumeClaim, isExists, isEphemeral bool) *JobWorkloadBuilder {

	pvcMap := make(map[string]flds.PvcsItemAndRules)

	quantity := pvc.Spec.Resources.Requests[v1.ResourceStorage]
	size := "0"
	if !quantity.IsZero() {
		size = quantity.String()
	}
	volumeMode := ""
	if pvc.Spec.VolumeMode != nil {
		volumeMode = string(*pvc.Spec.VolumeMode)
	}
	isReadWriteOnce := false
	isReadOnlyMany := false
	isReadWriteMany := false
	for _, accessMode := range pvc.Spec.AccessModes {
		switch accessMode {
		case v1.ReadWriteOnce:
			isReadWriteOnce = true
		case v1.ReadOnlyMany:
			isReadWriteMany = true
		case v1.ReadWriteMany:
			isReadOnlyMany = true
		}
	}

	pvcMap[pvc.Name] = flds.PvcsItemAndRules{
		Value: &flds.PvcsItemValue{
			ClaimName:     ptr.To(pvc.Name),
			StorageClass:  pvc.Spec.StorageClassName,
			Size:          ptr.To(size),
			ExistingPvc:   ptr.To(isExists),
			ReadWriteOnce: ptr.To(isReadWriteOnce),
			ReadOnlyMany:  ptr.To(isReadOnlyMany),
			ReadWriteMany: ptr.To(isReadWriteMany),
			Ephemeral:     ptr.To(isEphemeral),
			VolumeMode:    ptr.To(volumeMode),
		},
	}
	if jobWorkloadBuilder.workload.Spec.Pvcs == nil {
		jobWorkloadBuilder.workload.Spec.Pvcs = flds.NewPvcsField(pvcMap)
	} else {
		jobWorkloadBuilder.workload.Spec.Pvcs.Deref()[pvc.Name] = pvcMap[pvc.Name]
	}

	return jobWorkloadBuilder
}

func (JobWorkloadBuilder *JobWorkloadBuilder) WithNoMaster() *JobWorkloadBuilder {
	JobWorkloadBuilder.workload.Spec.NoMaster = flds.NewOptionalBoolField(ptr.To(true))
	return JobWorkloadBuilder
}

func (jobWorkloadBuilder *JobWorkloadBuilder) WithOldMPI() *JobWorkloadBuilder {
	jobWorkloadBuilder.workload.Spec.Mpi = flds.NewOptionalBoolField(ptr.To(true))
	return jobWorkloadBuilder
}

func (JobWorkloadBuilder *JobWorkloadBuilder) WithWorkers(replicas int) *JobWorkloadBuilder {
	JobWorkloadBuilder.workload.Spec.Workers = flds.NewOptionalIntField(ptr.To(replicas))
	return JobWorkloadBuilder
}

func (jwb *JobWorkloadBuilder) AsPyTorch() *JobWorkloadBuilder {
	jwb.workload.Spec.JobType = constants.KindPyTorchJob
	return jwb
}

func (jobWorkloadBuilder *JobWorkloadBuilder) WithExternalUrl(urls []string, ports []int) *JobWorkloadBuilder {
	urlsMap := map[string]flds.ExposedUrlItemAndRules{}
	for index, url := range urls {

		customUrl := false
		// Check if a custom URL is provided and validate it.
		if url != "" {
			customUrl = true

		}
		// Check if a port argument is provided.

		// Parse the port argument.
		port := ports[index]

		key := fmt.Sprintf("exposed-url-%d", index)
		// Construct the exposed URL item and add it to the URLs map.
		value := &flds.ExposedUrlItemValue{
			ContainerPort: ptr.To(int32(port)),
			CustomUrl:     ptr.To(customUrl),
		}
		if ptr.Deref(value.CustomUrl, false) {
			value.Url = ptr.To(url)
		}

		urlsMap[key] = flds.ExposedUrlItemAndRules{
			Rules:   nil,
			Value:   value,
			Deleted: nil,
		}
	}
	jobWorkloadBuilder.workload.Spec.ExposedUrls = flds.NewExposedUrlField(urlsMap)
	return jobWorkloadBuilder
}

func (jobWorkloadBuilder *JobWorkloadBuilder) WithEnvironmentVariable(key, value string) *JobWorkloadBuilder {
	if jobWorkloadBuilder.workload.Spec.Environment == nil {
		jobWorkloadBuilder.workload.Spec.Environment = flds.NewDictionaryField(make(map[string]string))
	}
	jobWorkloadBuilder.workload.Spec.Environment.Put(key, value)
	return jobWorkloadBuilder
}

func NewJobWorkloadBuilder() *JobWorkloadBuilder {
	workload := &jobs.JobWorkload{}
	workload.Labels = make(map[string]string)
	workload.Annotations = make(map[string]string)
	return &JobWorkloadBuilder{
		workload: workload,
	}
}

func CreateStructServicePort(portStr string, allowSpecifyHostPort bool) (*flds.PortItemAndRules, error) {
	hostPort, containerPort, err := parseServicePortLiterals(portStr)
	if err != nil {
		return nil, err
	}

	if !allowSpecifyHostPort && hostPort != 0 {
		return nil, fmt.Errorf("invalid port. Only container-port should be specified with the chosen service type")
	}

	var svcPort *flds.PortItemAndRules
	if !allowSpecifyHostPort { // only container port is given, and should be used as both host and container ports
		svcPort = &flds.PortItemAndRules{
			Value: &flds.PortItemValue{
				HostPort:        ptr.To(containerPort),
				ContainerPort:   ptr.To(containerPort),
				AutoGenHostPort: ptr.To(false),
			},
		}
	} else if hostPort != 0 { // both host and container ports are given
		svcPort = &flds.PortItemAndRules{
			Value: &flds.PortItemValue{
				HostPort:        ptr.To(hostPort),
				ContainerPort:   ptr.To(containerPort),
				AutoGenHostPort: ptr.To(false),
			},
		}
	} else { // only container port is given, let host port to be auto-generated
		svcPort = &flds.PortItemAndRules{
			Value: &flds.PortItemValue{
				ContainerPort:   ptr.To(containerPort),
				AutoGenHostPort: ptr.To(true),
			},
		}
	}

	return svcPort, nil
}

// parseServicePortLiterals returns host and container ports of a service
func parseServicePortLiterals(portStr string) (hostPort int, containerPort int, err error) {
	ports := strings.Split(portStr, ":")
	if (len(ports) != 1 && len(ports) != 2) || (len(ports) == 2 && (ports[0] == "" || ports[1] == "")) {
		return 0, 0, fmt.Errorf("invalid port. Use one of the following formats: 'host-port:container-port' or 'container-port'")
	}

	hostPort = 0
	if len(ports) == 1 {
		containerPort, err = ParsePort(ports[0], false)
		if err != nil {
			return 0, 0, err
		}
	} else {
		hostPort, err = ParsePort(ports[0], true)
		if err != nil {
			return 0, 0, err
		}
		containerPort, err = ParsePort(ports[1], false)
		if err != nil {
			return 0, 0, err
		}
	}

	return hostPort, containerPort, nil
}

func ParsePort(strPort string, isHost bool) (int, error) {
	num, err := strconv.Atoi(strPort)
	if err != nil {
		return 0, fmt.Errorf("invalid port number %s. Use only numbers", strPort)
	}

	var lowestPort int
	var highestPort int
	var portTypeStr string
	if isHost {
		lowestPort = LowestHostPort
		highestPort = HighestHostPort
		portTypeStr = "host"
	} else {
		lowestPort = LowestContainerPort
		highestPort = HighestContainerPort
		portTypeStr = "container"
	}

	if num < lowestPort || num > highestPort {
		return 0, fmt.Errorf("invalid %s port number %s. Valid %s port numbers are between %d and %d ", portTypeStr, strPort, portTypeStr, lowestPort, highestPort)
	}

	return num, nil
}
