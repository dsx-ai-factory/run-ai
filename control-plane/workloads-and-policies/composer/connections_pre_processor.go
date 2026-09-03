/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	assetsApi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	managerApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/builders"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/common"
)

const fakeDomain = "https://fake-domain.com"

type ConnectionsPreProcessor struct {
	WorkloadType workloadsApi.WorkloadType
	WorkloadName string
	ProjectName  string

	// ResourceNamePrefix overrides the workload-initials prefix (iw/tw/dw/infw).
	// WorkloadRunner IW/TW composers use "job" to correlate notification
	// connections with the networking resources they create. Kubeflow composers
	// use "kf"; multi-worker networking is rejected before resource names
	// require replica indexes. When empty, the V1 workloadTypeToInitials mapping
	// is used.
	ResourceNamePrefix            string
	WorkloadV2NetworkingSupported bool
}

var portsConnectionTypes = datastructs.NewSet[string](
	string(assetsApi.INTERNALCONNECTIONTYPE_LOAD_BALANCER),
	string(assetsApi.INTERNALCONNECTIONTYPE_NODE_PORT),
	string(assetsApi.INTERNALCONNECTIONTYPE_SERVING_PORT),
)

var workloadTypeToInitials = map[workloadsApi.WorkloadType]string{
	workloadsApi.WORKLOADTYPE_WORKSPACE:   common.KindInteractiveWorkloadInitial,
	workloadsApi.WORKLOADTYPE_TRAINING:    common.KindTrainingWorkloadInitial,
	workloadsApi.WORKLOADTYPE_DISTRIBUTED: common.KindDistributedWorkloadInitial,
	workloadsApi.WORKLOADTYPE_INFERENCE:   common.KindInferenceWorkloadInitial,
}

var (
	subdomainSupported = builders.ComposeUrlOptions{SubdomainSupported: true}
	noSubdomainSupport = builders.ComposeUrlOptions{SubdomainSupported: false}
)

type ServiceAndPortName struct {
	ServiceName string
	PortName    string
}

func (cp *ConnectionsPreProcessor) HandlePorts(connections []workloadsApi.Connection, ports *field_types.PortsField,
	replicaType *workloadsApi.ReplicaType, framework *workloadsApi.DistributedFramework,
) {
	if len(connections) == 0 || ports == nil {
		return
	}

	switch cp.ResourceNamePrefix {
	case NetworkingResourcePrefixJob:
		if !cp.WorkloadV2NetworkingSupported {
			return
		}
		cp.handleWorkloadRunnerJobPorts(connections, ports)
		return
	case NetworkingResourcePrefixKubeflow:
		if !cp.WorkloadV2NetworkingSupported {
			return
		}
		cp.handleWorkloadRunnerKubeflowPorts(connections, ports, replicaType, framework)
		return
	}

	// Resolve the prefix used by the native V1 port Service naming helper. A
	// missing mapping is not an error - it just means there is no need to
	// pre-process.
	workloadInitials, typeFound := workloadTypeToInitials[cp.WorkloadType]
	if !typeFound {
		return
	}

	// activate the workload-common function that is responsible for creating the ports and service names.
	nickPerPortType := builders.GetPortsPerServiceType(workloadInitials, cp.WorkloadName, ports, nil)

	// find the port and service name that were assigned to each nick
	nickToResourceNames := make(map[string]ServiceAndPortName)

	for _, serviceNameAndPort := range nickPerPortType {
		for idx := 0; idx < len(serviceNameAndPort.Ports); idx++ {
			portName := serviceNameAndPort.Ports[idx].Name
			portNick := serviceNameAndPort.PortsNick[idx]

			nickToResourceNames[portNick] = ServiceAndPortName{
				ServiceName: serviceNameAndPort.Name,
				PortName:    portName,
			}
		}
	}

	cp.applyPortResourceNames(connections, nickToResourceNames)
}

func (cp *ConnectionsPreProcessor) handleWorkloadRunnerJobPorts(connections []workloadsApi.Connection, ports *field_types.PortsField) {
	nickToResourceNames := make(map[string]ServiceAndPortName)

	for _, nick := range field_types.SortedPortsMapKeys(ports.Deref()) {
		portItem := ports.Deref()[nick].Value
		if portItem == nil || portItem.ContainerPort == nil {
			continue
		}

		serviceType, err := portItemServiceType(portItem.ServiceType)
		if err != nil {
			continue
		}

		serviceName := ComposeNetworkingResourceName(cp.ResourceNamePrefix, cp.WorkloadName, ServiceTypeShortName(serviceType))
		nickToResourceNames[nick] = ServiceAndPortName{
			ServiceName: serviceName,
			PortName:    fmt.Sprintf("port-%d", *portItem.ContainerPort),
		}
	}

	cp.applyPortResourceNames(connections, nickToResourceNames)
}

func (cp *ConnectionsPreProcessor) handleWorkloadRunnerKubeflowPorts(connections []workloadsApi.Connection,
	ports *field_types.PortsField, replicaType *workloadsApi.ReplicaType, framework *workloadsApi.DistributedFramework,
) {
	nickToResourceNames := make(map[string]ServiceAndPortName)
	role := kubeflowNetworkingRole(replicaType, framework)

	for _, nick := range field_types.SortedPortsMapKeys(ports.Deref()) {
		portItem := ports.Deref()[nick].Value
		if portItem == nil || portItem.ContainerPort == nil {
			continue
		}

		serviceType, err := portItemServiceType(portItem.ServiceType)
		if err != nil {
			continue
		}

		serviceName := ComposeNetworkingResourceName(cp.ResourceNamePrefix, cp.WorkloadName,
			fmt.Sprintf("%s-%s", ServiceTypeShortName(serviceType), role))
		nickToResourceNames[nick] = ServiceAndPortName{
			ServiceName: serviceName,
			PortName:    fmt.Sprintf("port-%d", *portItem.ContainerPort),
		}
	}

	cp.applyPortResourceNames(connections, nickToResourceNames)
}

func kubeflowNetworkingRole(replicaType *workloadsApi.ReplicaType, framework *workloadsApi.DistributedFramework) string {
	if replicaType != nil && *replicaType == workloadsApi.REPLICATYPE_MASTER {
		if framework != nil && *framework == workloadsApi.DISTRIBUTEDFRAMEWORK_MPI {
			return "launcher"
		}
		return "master"
	}
	return "worker"
}

func portItemServiceType(serviceType *string) (managerApi.ServiceType, *workloadsApi.Error) {
	if serviceType == nil {
		return PortServiceTypeToServiceType(nil)
	}

	portServiceType := workloadsApi.PortServiceType(*serviceType)
	return PortServiceTypeToServiceType(&portServiceType)
}

func (cp *ConnectionsPreProcessor) applyPortResourceNames(
	connections []workloadsApi.Connection,
	nickToResourceNames map[string]ServiceAndPortName,
) {
	for idx, connection := range connections {
		if !portsConnectionTypes.Has(connection.ConnectionType) {
			continue
		}

		if convutils.IsEmptyOrNull(connection.CorrelationName) {
			continue
		}

		serviceNameAndPort, found := nickToResourceNames[*connection.CorrelationName]
		if found {
			connections[idx].ServiceResourceName.Set(&serviceNameAndPort.ServiceName)
			connections[idx].PortResourceName.Set(&serviceNameAndPort.PortName)
		}
	}
}

func (cp *ConnectionsPreProcessor) HandleExposedUrls(connections []workloadsApi.Connection, exposedUrls *field_types.ExposedUrlsField,
	replicaType *workloadsApi.ReplicaType, framework *workloadsApi.DistributedFramework,
) {
	if len(connections) == 0 || exposedUrls == nil {
		return
	}

	noSuffixGenerated := false
	nickToUrlWithSubdomain, err := builders.GetUrlPerNick(exposedUrls, cp.ProjectName, cp.WorkloadName, fakeDomain, &noSuffixGenerated, subdomainSupported)
	if err != nil {
		return
	}

	noSuffixGenerated = false
	nickToUrlNoSubdomain, err := builders.GetUrlPerNick(exposedUrls, cp.ProjectName, cp.WorkloadName, fakeDomain, &noSuffixGenerated, noSubdomainSupport)
	if err != nil {
		return
	}

	for idx, connection := range connections {
		if connection.ConnectionType != string(assetsApi.INTERNALCONNECTIONTYPE_EXTERNAL_URL) {
			continue
		}

		if convutils.IsEmptyOrNull(connection.CorrelationName) {
			continue
		}

		// this is a custom url, the final connection url is known. we do not need to get it from the cluster
		if convutils.IsStringHasValue(connection.Url) {
			continue
		}

		switch cp.ResourceNamePrefix {
		case NetworkingResourcePrefixJob, NetworkingResourcePrefixInference:
			serviceName := ComposeNetworkingResourceName(cp.ResourceNamePrefix, cp.WorkloadName, *connection.CorrelationName)
			connections[idx].ServiceResourceName.Set(&serviceName)
		case NetworkingResourcePrefixKubeflow:
			if replicaType == nil {
				break
			}
			serviceName := ComposeNetworkingResourceName(cp.ResourceNamePrefix, cp.WorkloadName, *connection.CorrelationName)
			connections[idx].ServiceResourceName.Set(&serviceName)
		}

		//
		// for cluster that work with no subdomain, we only need the path /myproject/myworload/url-1
		//
		urlForPathOnly, found := nickToUrlNoSubdomain[*connection.CorrelationName]
		if found {
			connections[idx].ExposedUrlPath.Set(&urlForPathOnly.Path)
		}

		//
		// for clusters that work with subdomain, we need only the subdomain. this is the first
		// part of the url's host (myproject-myworkload-url-1.fake-domain.com)
		//
		urlForSubdomain, found := nickToUrlWithSubdomain[*connection.CorrelationName]
		if found {
			subdomain := strings.Split(urlForSubdomain.Host, ".")[0]
			connections[idx].ExposedUrlSubDomain.Set(convutils.Ref(subdomain))
		}
	}
}

// ComposeNetworkingResourceName returns the K8s resource name for an exposed
// URL's networking objects (Service / Ingress / Route / HTTPRoute). It mirrors
// the cluster builder's "<workloadInitials>-<workloadName>-<suffix>" pattern so
// V2-created resources stay namespaced per workload-kind and never collide with
// other kinds that happen to share the same base name. The prefix may be empty
// (legacy callers) in which case the bare suffix is returned.
func ComposeNetworkingResourceName(prefix, workloadName, suffix string) string {
	var name string
	switch {
	case prefix == "" && workloadName == "":
		name = suffix
	case prefix == "":
		name = fmt.Sprintf("%s-%s", workloadName, suffix)
	case workloadName == "":
		name = fmt.Sprintf("%s-%s", prefix, suffix)
	default:
		name = fmt.Sprintf("%s-%s-%s", prefix, workloadName, suffix)
	}

	const (
		maxResourceNameLength = 63
		hashLength            = 10
	)
	if len(name) <= maxResourceNameLength {
		return name
	}

	// Kubernetes Service names are limited to 63 characters. Preserve a useful
	// prefix for diagnostics and add a stable hash of the complete name so that
	// truncation cannot make two networking resources share the same name.
	digest := fmt.Sprintf("%x", sha256.Sum256([]byte(name)))[:hashLength]
	return fmt.Sprintf("%s-%s", name[:maxResourceNameLength-hashLength-1], digest)
}

// ServiceTypeShortName returns the short token used in port Service names
// (cip / lb / np). Keeps Service names short and easy to scan.
func ServiceTypeShortName(t managerApi.ServiceType) string {
	switch t {
	case managerApi.SERVICETYPE_CLUSTER_IP:
		return "cip"
	case managerApi.SERVICETYPE_LOAD_BALANCER:
		return "lb"
	case managerApi.SERVICETYPE_NODE_PORT:
		return "np"
	default:
		return strings.ToLower(string(t))
	}
}

// PortServiceTypeToServiceType converts the workloads-client PortServiceType to
// the workloads-manager ServiceType. A nil/empty service type defaults to
// ClusterIP, matching the V1 cluster builder's getServiceType fallback.
func PortServiceTypeToServiceType(portServiceType *workloadsApi.PortServiceType) (managerApi.ServiceType, *workloadsApi.Error) {
	if portServiceType == nil {
		return managerApi.SERVICETYPE_CLUSTER_IP, nil
	}

	switch *portServiceType {
	case workloadsApi.PORTSERVICETYPE_CLUSTER_IP:
		return managerApi.SERVICETYPE_CLUSTER_IP, nil
	case workloadsApi.PORTSERVICETYPE_NODE_PORT:
		return managerApi.SERVICETYPE_NODE_PORT, nil
	case workloadsApi.PORTSERVICETYPE_LOAD_BALANCER:
		return managerApi.SERVICETYPE_LOAD_BALANCER, nil
	default:
		return "", workloadsApi.NewError(http.StatusBadRequest,
			fmt.Sprintf("PortsToWorkloadRunner: unsupported service type %q", *portServiceType))
	}
}
