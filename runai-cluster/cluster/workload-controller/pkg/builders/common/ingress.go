/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"context"
	"fmt"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
	"knative.dev/pkg/apis"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"

	"os"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	workloadbuilders "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/builders"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// both jupyter and tensorboard use the same default container port.
const (
	DefaultJupyterAndTensorContainerPort = int32(8888)
	pageSize                             = 1000
)

func UpdateJobUrlWithNodePorts(ctx context.Context, Objects []client.Object,
	K8sClient client.Client, annotations *metav1.ObjectMeta) error {
	logging := logger.FromContext(ctx)
	serviceNodePorts := getNodePorts(Objects)

	if len(serviceNodePorts) == 0 {
		logging.V(4).Info("no node ports found")
		return nil
	}
	logging.V(4).Info("found node ports", "node ports", serviceNodePorts)
	logging.V(4).Info("using workload controller node name to get the node internal IP address")
	nodeName := GetWorkloadControllerNodeName()

	if nodeName == "" {
		logging.Error(fmt.Errorf("node name not found"), "workload controller node name env is not set")
		return fmt.Errorf("could not find the workload controller node name (node is nil)" +
			", node ports will not be added to the job URL")
	}
	logging.V(4).Info("workload controller node name", "node name", nodeName)
	exposedNode, err := getWorkloadControllerNode(ctx, K8sClient, nodeName)
	if err != nil {
		logging.Error(err, "failed to get the workload controller node for updating job URL with node ports")
		return err
	}
	logging.V(4).Info("found the workload controller node", "node", exposedNode.Name,
		"node IPs", exposedNode.Status.Addresses)

	nodeInternalIP := getNodeInternalIP(exposedNode)
	if nodeInternalIP == "" {
		logging.Info("no internal IP found for the exposed node, skipping updating job URL with node ports")
		return nil
	}
	logging.V(4).Info("found the internal IP of the exposed node", "node IP", nodeInternalIP)

	for _, port := range serviceNodePorts {
		exposedAddress := fmt.Sprintf("%s:%d", nodeInternalIP, port)
		workloadbuilders.AddJobUrlToAnnotations(annotations, exposedAddress)
	}
	logging.V(4).Info("Completed updating job URL with node ports",
		"job URL", annotations.Annotations[constants.AnnotationJobUrl])
	return nil
}

func getNodePorts(Objects []client.Object) []int32 {
	serviceNodePorts := make([]int32, 0)
	for _, obj := range Objects {
		if svc, ok := obj.(*corev1.Service); ok {
			if svc.Spec.Type == corev1.ServiceTypeNodePort {
				for _, port := range svc.Spec.Ports {
					serviceNodePorts = append(serviceNodePorts, port.NodePort)
				}
			}
		}
	}
	// sort the node ports to make sure the result is deterministic
	// otherwise, this will create endless reconcile loop (Very bad)
	sort.Slice(serviceNodePorts, func(i, j int) bool {
		return serviceNodePorts[i] < serviceNodePorts[j]
	})
	return serviceNodePorts
}

// getNodeInternalIP returns the internal IP address of the node
func getNodeInternalIP(node *corev1.Node) string {
	nodeAddresses := node.Status.Addresses

	var internalNodeIPs []corev1.NodeAddress
	for _, address := range nodeAddresses {
		if address.Type == corev1.NodeInternalIP {
			internalNodeIPs = append(internalNodeIPs, address)
		}
	}
	if len(internalNodeIPs) == 0 {
		// if no internal IP found, return empty string
		return ""
	}
	// this is important because we want to make sure the result is deterministic
	// we sort the addresses by string
	// otherwise, this will create endless reconcile loop (Very bad)
	sort.Slice(internalNodeIPs, func(i, j int) bool {
		return strings.Compare(internalNodeIPs[i].Address, internalNodeIPs[j].Address) < 0
	})

	return internalNodeIPs[0].Address
}

func GetWorkloadControllerNodeName() string {
	return os.Getenv(constants.WorkloadControllerNodeNameEnvKey)
}

func getWorkloadControllerNode(ctx context.Context, K8sClient client.Client, name string) (*corev1.Node, error) {
	var node corev1.Node
	err := K8sClient.Get(ctx, client.ObjectKey{Name: name}, &node)
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// ComposeWorkspaceUrl composes URL for accessing workspace directly. Sample URL:
// http://domain.com/workspaces?columnFilter=[{"term":"work-100","name":"name"}]&clusterId=5ee7956e-9e76-4244-9f7c-7fb1cedd38a0
func ComposeWorkspaceUrl(_ context.Context, workloadKind string, jobName string, clusterId string) (*apis.URL, error) {
	clusterDomain := config.GetClusterDomain()
	parsedUrlPattern, err := apis.ParseURL(clusterDomain)
	if err != nil || parsedUrlPattern == nil {
		return nil, fmt.Errorf("invalid cluster domain '%s': %v", clusterDomain, err)
	}
	if workloadKind == constants.KindInteractiveWorkload {
		parsedUrlPattern.Path = "workspaces"
	} else {
		parsedUrlPattern.Path = "trainings"
	}
	parsedUrlPattern.RawQuery = fmt.Sprintf("columnFilter=[{\"term\":\"%s\",\"name\":\"name\"}]&clusterId=%s",
		jobName, clusterId)
	return parsedUrlPattern, nil
}
