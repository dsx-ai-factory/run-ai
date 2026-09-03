/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"
	"fmt"
	"strings"
	"time"

	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	controllerutils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/controller-utils"
	flds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var serviceTypes = []corev1.ServiceType{
	corev1.ServiceTypeClusterIP,
	corev1.ServiceTypeLoadBalancer,
	corev1.ServiceTypeNodePort,
}

type ServiceBuilder struct {
	k8sClient client.Client
}

func NewServiceBuilder(k8sClient client.Client) *ServiceBuilder {
	return &ServiceBuilder{
		k8sClient: k8sClient,
	}
}

// GetPreview returns a list of resources (name/kind only) that will be created on build
func (sb *ServiceBuilder) GetPreview(ctx context.Context, workloadKind string, workloadInitials string, objectMeta *metav1.ObjectMeta,
	ports *flds.PortsField, defaultServiceType *flds.StringField) ([]Resource, error) {
	var resources []Resource

	workloadName := objectMeta.Name

	logger := log.FromContext(ctx)

	// some validations on the ports
	for _, nick := range flds.SortedPortsMapKeys(ports.Deref()) {

		p := ports.Deref()[nick].Value
		if p == nil {
			logger.Info("port is nil, skipping port")
			continue
		}

		serviceTypeStr, _, err := getServiceType(p.ServiceType, defaultServiceType)
		if err != nil {
			logger.Error(err, "failed to get service type, skipping port")
			continue
		}

		if serviceTypeStr == constants.ServiceTypeServingPort {
			logger.Info(fmt.Sprintf("service type is %s, service is not created for this type, skipping port", constants.ServiceTypeServingPort))
			continue
		}
	}

	portsPerServiceType := GetPortsPerServiceType(workloadInitials, workloadName, ports, defaultServiceType)

	for _, serviceType := range serviceTypes {
		if portsPerServiceType[serviceType] == nil {
			continue
		}

		serviceName := portsPerServiceType[serviceType].Name

		resources = append(resources, Resource{
			Kind:      serviceGroupVersion,
			Name:      serviceName,
			Namespace: objectMeta.Namespace,
		})
	}

	return resources, nil
}

func (sb *ServiceBuilder) Build(ctx context.Context, workloadKind string, jobType string, workloadInitials string, objectMeta *metav1.ObjectMeta,
	workloadName string, owner client.Object, ports *flds.PortsField, defaultServiceType *flds.StringField,
	calculatedPriorityClassNameLabel string) ([]corev1.Service, error) {
	logger := log.FromContext(ctx)

	err := SetObjectMetaLabels(ctx, workloadKind, jobType, objectMeta.Name, objectMeta)
	if err != nil {
		return nil, err
	}

	err = SetPriorityClassLabel(ctx, sb.k8sClient, calculatedPriorityClassNameLabel, objectMeta)
	if err != nil {
		return nil, err
	}

	portsPerServiceType := GetPortsPerServiceType(workloadInitials, workloadName, ports, defaultServiceType)

	// create a service for each service type, each service with the collected ports
	var services []corev1.Service

	for _, serviceType := range serviceTypes {
		if portsPerServiceType[serviceType] == nil {
			continue
		}

		serviceName := portsPerServiceType[serviceType].Name

		service := corev1.Service{
			TypeMeta: serviceTypeMeta,
			ObjectMeta: metav1.ObjectMeta{
				CreationTimestamp: metav1.Time{Time: time.Now()},
				Labels:            objectMeta.Labels,
				Name:              serviceName,
				Namespace:         objectMeta.Namespace,
			},
			Spec: corev1.ServiceSpec{
				Ports:    portsPerServiceType[serviceType].Ports,
				Selector: map[string]string{constants.LabelRelease: objectMeta.Name},
				Type:     serviceType,
			},
		}

		if owner != nil {
			err := controllerutils.SetControllerReference(owner, &service, sb.k8sClient.Scheme())
			if err != nil {
				logger.Error(err, "failed to set controller reference")
				return nil, err
			}
		}

		services = append(services, service)
	}

	return services, nil
}

// getServiceType returns the service type and the service type string, taking into account
// the service type provided in the ports field and the default service type
func getServiceType(serviceType *string, defaultServiceType *flds.StringField) (string, corev1.ServiceType, error) {
	serviceTypeStr := ""
	if serviceType != nil {
		serviceTypeStr = *serviceType
	} else {
		if defaultServiceType != nil && defaultServiceType.Value != nil {
			serviceTypeStr = *defaultServiceType.Value
		} else {
			serviceTypeStr = constants.ServiceTypeClusterIP
		}
	}

	switch serviceTypeStr {
	case constants.ServiceTypeNodePort:
		return serviceTypeStr, corev1.ServiceTypeNodePort, nil
	case constants.ServiceTypeLoadBalancer:
		return serviceTypeStr, corev1.ServiceTypeLoadBalancer, nil
	case constants.ServiceTypeClusterIP:
		return serviceTypeStr, corev1.ServiceTypeClusterIP, nil
	case constants.ServiceTypeServingPort:
		return serviceTypeStr, "", nil
	default:
		return "", corev1.ServiceTypeClusterIP, fmt.Errorf("service type '%s' is not supported", serviceTypeStr)
	}
}

type PortNameAndNick struct {
	Port corev1.ServicePort
	Nick string
}

type ServiceNameAndPorts struct {
	Name      string
	Ports     []corev1.ServicePort
	PortsNick []string
}

// ////////////////////////////////////////////////////////////////////////////////////
// GetPortsPerServiceType receives the ports defined in the workload, and returns
// a map from serviceType to:
//
//   - service assigned name (e.g. service-xxx-yyy-0)
//
//   - service ports, for each port we have:
//
//   - the port info as corev1.ServicePort
//
//   - the port nickname from the workload
//
//     **************
//     **** NOTE ****
//     **************
//
// This function is used by both workload controller (cluster) AND workloads service (control plane)
// The CP uses it for mapping between the assigned service and port name to the nickname from the workload
// It stores this information in the workloads DB for any workload it created
// Therefore, any change in the naming scheme for either the service or the port, must take into account
// existing workloads that may depend on this mapping.
// ////////////////////////////////////////////////////////////////////////////////////
func GetPortsPerServiceType(workloadInitials string, workloadBaseName string, ports *flds.PortsField,
	defaultServiceType *flds.StringField,
) map[corev1.ServiceType]*ServiceNameAndPorts {
	portsPerServiceType := make(map[corev1.ServiceType]*ServiceNameAndPorts)

	portIdx := 0

	for _, nick := range flds.SortedPortsMapKeys(ports.Deref()) {
		//
		//  skip ports that are not valid. no need to log or return error as we already
		//  did so in GetPreview above
		//
		p := ports.Deref()[nick].Value
		if p == nil {
			continue
		}

		serviceTypeStr, serviceType, err := getServiceType(p.ServiceType, defaultServiceType)
		if err != nil {
			continue
		}

		if serviceTypeStr == constants.ServiceTypeServingPort {
			continue
		}

		// make sure array of ports per service type is allocated
		if portsPerServiceType[serviceType] == nil {
			portsPerServiceType[serviceType] = &ServiceNameAndPorts{
				Ports:     make([]corev1.ServicePort, 0),
				PortsNick: make([]string, 0),
			}
		}

		containerPort := int32(ptr.Deref(p.ContainerPort, 0))

		var port int32
		var nodePort int32
		var protocol corev1.Protocol

		if serviceType == corev1.ServiceTypeNodePort {
			if !ptr.Deref(p.AutoGenHostPort, false) {
				nodePort = int32(ptr.Deref(p.HostPort, 0))
			}
			port = int32(ptr.Deref(p.ContainerPort, 0))
		} else {
			// if host port wasn't provided, use the container port
			port = int32(ptr.Deref(p.HostPort, int(containerPort)))
			protocol = corev1.ProtocolTCP
		}

		portsPerServiceType[serviceType].Ports = append(portsPerServiceType[serviceType].Ports, corev1.ServicePort{
			Name:       fmt.Sprintf("port-%s-%d-%d", workloadBaseName, containerPort, portIdx),
			Port:       port,
			TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: containerPort},
			NodePort:   nodePort,
			Protocol:   protocol,
		})
		portsPerServiceType[serviceType].PortsNick = append(portsPerServiceType[serviceType].PortsNick, nick)

		portIdx += 1
	}

	serviceIdx := 0

	for _, serviceType := range serviceTypes {
		if portsPerServiceType[serviceType] == nil {
			continue
		}

		suffix := getServiceNameSuffix(serviceType, serviceIdx)
		serviceName := getServiceName(workloadInitials, workloadBaseName, suffix)

		portsPerServiceType[serviceType].Name = serviceName

		serviceIdx++
	}

	return portsPerServiceType
}

func getServiceNameSuffix(serviceType corev1.ServiceType, idx int) string {
	return fmt.Sprintf("%d-%s", idx, strings.ToLower(string(serviceType)))
}
