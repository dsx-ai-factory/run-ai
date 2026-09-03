/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"
	"fmt"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/project"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	flds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("ServiceBuilder tests", func() {
	var (
		scheme                      *runtime.Scheme
		fakeClient                  client.Client
		sb                          *ServiceBuilder
		ctx                         context.Context
		projectName                 string
		workloadKind                string
		jobType                     string
		workloadInitials            string
		workloadName                string
		objectMeta                  *metav1.ObjectMeta
		owner                       *corev1.Pod
		ports                       *flds.PortsField
		defaultServiceType          *flds.StringField
		calculatedPriorityClassName string

		priorityClassNames = []string{constants.InteractivePreemptiblePriorityClass, constants.InteractivePriorityClass, "train"}
	)

	BeforeEach(func() {
		// Setup common test variables
		scheme = runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(schedulingv1.AddToScheme(scheme)).To(Succeed())

		fakeClient = fake.NewClientBuilder().WithScheme(scheme).Build()
		sb = NewServiceBuilder(fakeClient)

		ctx = context.Background()
		projectName = "test-project"
		ctx = project.SetProjectInContext(ctx, projectName)

		workloadKind = constants.KindInteractiveWorkload
		jobType = "interactive"
		workloadInitials = "iw"
		workloadName = "test-workload"

		objectMeta = &metav1.ObjectMeta{
			Name:      workloadName,
			Namespace: "default",
			Labels:    map[string]string{},
		}

		owner = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "default",
				UID:       "test-uid",
			},
		}

		// Initialize empty ports
		ports = &flds.PortsField{
			Value: &map[string]flds.PortItemAndRules{},
		}

		// Default service type is ClusterIP
		defaultServiceType = &flds.StringField{
			Value: ptr.To(constants.ServiceTypeClusterIP),
		}

		calculatedPriorityClassName = ""

		createPriorityClasses(fakeClient, priorityClassNames)
	})

	AfterEach(func() {
		deletePriorityClasses(fakeClient, priorityClassNames)
	})

	// Helper function to add a port to the ports field
	addPort := func(nick string, containerPort int, hostPort int, serviceType string, autoGenHostPort bool) {
		(*ports.Value)[nick] = flds.PortItemAndRules{
			Value: &flds.PortItemValue{
				ContainerPort:   &containerPort,
				HostPort:        &hostPort,
				ServiceType:     &serviceType,
				AutoGenHostPort: &autoGenHostPort,
			},
		}
	}

	Context("Basic functionality", func() {
		It("should return empty when no ports are specified", func() {
			// GetPreview should return empty when no ports
			preview, err := sb.GetPreview(ctx, workloadKind, workloadInitials, objectMeta, ports, defaultServiceType)
			Expect(err).NotTo(HaveOccurred())
			Expect(preview).To(BeEmpty())

			// Build should return empty when no ports
			services, err := sb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadName, owner,
				ports, defaultServiceType, calculatedPriorityClassName)
			Expect(err).NotTo(HaveOccurred())
			Expect(services).To(BeEmpty())
		})

		It("should create service with ClusterIP (default) when single port is specified", func() {
			// Add a port with default ClusterIP service type
			addPort("main", 8080, 8080, constants.ServiceTypeClusterIP, false)

			// GetPreview should return a service resource
			preview, err := sb.GetPreview(ctx, workloadKind, workloadInitials, objectMeta, ports, defaultServiceType)
			Expect(err).NotTo(HaveOccurred())
			Expect(preview).To(HaveLen(1))
			Expect(preview[0].Kind).To(Equal(serviceGroupVersion))
			Expect(preview[0].Name).To(ContainSubstring("iw-test-workload-0-clusterip"))

			// Build should create a ClusterIP service
			services, err := sb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadName, owner,
				ports, defaultServiceType, calculatedPriorityClassName)
			Expect(err).NotTo(HaveOccurred())
			Expect(services).To(HaveLen(1))

			service := services[0]
			Expect(service.Name).To(ContainSubstring("iw-test-workload-0-clusterip"))
			Expect(service.Spec.Type).To(Equal(corev1.ServiceTypeClusterIP))
			Expect(service.Spec.Ports).To(HaveLen(1))
			Expect(service.Spec.Ports[0].Port).To(Equal(int32(8080)))
			Expect(service.Spec.Ports[0].TargetPort.IntVal).To(Equal(int32(8080)))
		})

		It("should create service with LoadBalancer when specified", func() {
			// Add a port with LoadBalancer service type
			addPort("main", 8080, 8080, constants.ServiceTypeLoadBalancer, false)

			// GetPreview should return a service resource
			preview, err := sb.GetPreview(ctx, workloadKind, workloadInitials, objectMeta, ports, defaultServiceType)
			Expect(err).NotTo(HaveOccurred())
			Expect(preview).To(HaveLen(1))
			Expect(preview[0].Name).To(ContainSubstring("iw-test-workload-0-loadbalancer"))

			// Build should create a LoadBalancer service
			services, err := sb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadName, owner,
				ports, defaultServiceType, calculatedPriorityClassName)
			Expect(err).NotTo(HaveOccurred())
			Expect(services).To(HaveLen(1))

			service := services[0]
			Expect(service.Name).To(ContainSubstring("iw-test-workload-0-loadbalancer"))
			Expect(service.Spec.Type).To(Equal(corev1.ServiceTypeLoadBalancer))
			Expect(service.Spec.Ports).To(HaveLen(1))
			Expect(service.Spec.Ports[0].Port).To(Equal(int32(8080)))
			Expect(service.Spec.Ports[0].TargetPort.IntVal).To(Equal(int32(8080)))
		})

		It("should create service with NodePort when specified", func() {
			// Add a port with NodePort service type
			addPort("main", 8080, 30080, constants.ServiceTypeNodePort, false)

			// GetPreview should return a service resource
			preview, err := sb.GetPreview(ctx, workloadKind, workloadInitials, objectMeta, ports, defaultServiceType)
			Expect(err).NotTo(HaveOccurred())
			Expect(preview).To(HaveLen(1))
			Expect(preview[0].Name).To(ContainSubstring("iw-test-workload-0-nodeport"))

			// Build should create a NodePort service
			services, err := sb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadName, owner,
				ports, defaultServiceType, calculatedPriorityClassName)
			Expect(err).NotTo(HaveOccurred())
			Expect(services).To(HaveLen(1))

			service := services[0]
			Expect(service.Name).To(ContainSubstring("iw-test-workload-0-nodeport"))
			Expect(service.Spec.Type).To(Equal(corev1.ServiceTypeNodePort))
			Expect(service.Spec.Ports).To(HaveLen(1))
			Expect(service.Spec.Ports[0].Port).To(Equal(int32(8080)))
			Expect(service.Spec.Ports[0].TargetPort.IntVal).To(Equal(int32(8080)))
			Expect(service.Spec.Ports[0].NodePort).To(Equal(int32(30080)))
		})
	})

	Context("Multiple ports/services scenarios", func() {
		It("should group multiple ports into a single service for the same service type", func() {
			// Add multiple ports with the same service type
			addPort("first", 8080, 8080, constants.ServiceTypeClusterIP, false)
			addPort("second", 9090, 9090, constants.ServiceTypeClusterIP, false)

			// GetPreview should return a single service resource
			preview, err := sb.GetPreview(ctx, workloadKind, workloadInitials, objectMeta, ports, defaultServiceType)
			Expect(err).NotTo(HaveOccurred())
			Expect(preview).To(HaveLen(1))
			Expect(preview[0].Name).To(ContainSubstring("iw-test-workload-0-clusterip"))

			// Build should create a single service with multiple ports
			services, err := sb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadName, owner,
				ports, defaultServiceType, calculatedPriorityClassName)
			Expect(err).NotTo(HaveOccurred())
			Expect(services).To(HaveLen(1))

			service := services[0]
			Expect(service.Name).To(ContainSubstring("iw-test-workload-0-clusterip"))
			Expect(service.Spec.Type).To(Equal(corev1.ServiceTypeClusterIP))
			Expect(service.Spec.Ports).To(HaveLen(2))

			// Verify both ports are included
			containerPorts := []int32{service.Spec.Ports[0].TargetPort.IntVal, service.Spec.Ports[1].TargetPort.IntVal}
			Expect(containerPorts).To(ConsistOf(int32(8080), int32(9090)))
		})

		It("should create multiple services for different service types", func() {
			// Add ports with different service types
			addPort("clusterip", 8080, 8080, constants.ServiceTypeClusterIP, false)
			addPort("nodeport", 9090, 30090, constants.ServiceTypeNodePort, false)

			// GetPreview should return two service resources
			preview, err := sb.GetPreview(ctx, workloadKind, workloadInitials, objectMeta, ports, defaultServiceType)
			Expect(err).NotTo(HaveOccurred())
			Expect(preview).To(HaveLen(2))

			// Check both service types are included in preview
			serviceTypes := []string{}
			for _, res := range preview {
				Expect(res.Kind).To(Equal(serviceGroupVersion))
				serviceName := res.Name
				if serviceName != "" {
					if strings.HasSuffix(serviceName, "clusterip") {
						serviceTypes = append(serviceTypes, "clusterip")
					} else if strings.HasSuffix(serviceName, "nodeport") {
						serviceTypes = append(serviceTypes, "nodeport")
					}
				}
			}
			Expect(serviceTypes).To(ConsistOf("clusterip", "nodeport"))

			// Build should create two services with different types
			services, err := sb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadName, owner,
				ports, defaultServiceType, calculatedPriorityClassName)
			Expect(err).NotTo(HaveOccurred())
			Expect(services).To(HaveLen(2))

			// Check both service types are created
			foundClusterIP := false
			foundNodePort := false
			for _, service := range services {
				if service.Spec.Type == corev1.ServiceTypeClusterIP {
					foundClusterIP = true
					Expect(service.Spec.Ports).To(HaveLen(1))
					Expect(service.Spec.Ports[0].TargetPort.IntVal).To(Equal(int32(8080)))
				} else if service.Spec.Type == corev1.ServiceTypeNodePort {
					foundNodePort = true
					Expect(service.Spec.Ports).To(HaveLen(1))
					Expect(service.Spec.Ports[0].TargetPort.IntVal).To(Equal(int32(9090)))
					Expect(service.Spec.Ports[0].NodePort).To(Equal(int32(30090)))
				}
			}
			Expect(foundClusterIP).To(BeTrue())
			Expect(foundNodePort).To(BeTrue())
		})

		It("should handle all three service types in the same call", func() {
			// Add ports with all three service types
			addPort("clusterip", 8080, 8080, constants.ServiceTypeClusterIP, false)
			addPort("nodeport", 9090, 30090, constants.ServiceTypeNodePort, false)
			addPort("loadbalancer", 7070, 7070, constants.ServiceTypeLoadBalancer, false)

			// GetPreview should return three service resources
			preview, err := sb.GetPreview(ctx, workloadKind, workloadInitials, objectMeta, ports, defaultServiceType)
			Expect(err).NotTo(HaveOccurred())
			Expect(preview).To(HaveLen(3))

			// Build should create three services with different types
			services, err := sb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadName, owner,
				ports, defaultServiceType, calculatedPriorityClassName)
			Expect(err).NotTo(HaveOccurred())
			Expect(services).To(HaveLen(3))

			// Check all three service types are created
			serviceTypes := []corev1.ServiceType{}
			for _, service := range services {
				serviceTypes = append(serviceTypes, service.Spec.Type)
			}
			Expect(serviceTypes).To(ConsistOf(
				corev1.ServiceTypeClusterIP,
				corev1.ServiceTypeNodePort,
				corev1.ServiceTypeLoadBalancer,
			))
		})
	})

	Context("Port configurations", func() {
		It("should use specified host port when provided", func() {
			// Add a port with specific host port
			addPort("main", 8080, 9000, constants.ServiceTypeClusterIP, false)

			// Build should create a service with the specified host port
			services, err := sb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadName, owner,
				ports, defaultServiceType, calculatedPriorityClassName)
			Expect(err).NotTo(HaveOccurred())
			Expect(services).To(HaveLen(1))

			// For ClusterIP service, the Port should be set to the host port
			service := services[0]
			Expect(service.Spec.Ports[0].Port).To(Equal(int32(9000)))
			Expect(service.Spec.Ports[0].TargetPort.IntVal).To(Equal(int32(8080)))
		})

		It("should handle auto-generated host port for NodePort services", func() {
			// Add a NodePort port with auto-generated host port
			addPort("main", 8080, 30080, constants.ServiceTypeNodePort, true)

			// Build should create a service with auto-generated host port
			services, err := sb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadName, owner,
				ports, defaultServiceType, calculatedPriorityClassName)
			Expect(err).NotTo(HaveOccurred())
			Expect(services).To(HaveLen(1))

			service := services[0]
			Expect(service.Spec.Type).To(Equal(corev1.ServiceTypeNodePort))
			// For auto-generated NodePort, the NodePort should be 0 (k8s will assign)
			Expect(service.Spec.Ports[0].NodePort).To(Equal(int32(0)))
		})

		It("should use service type override from port configuration", func() {
			// Add a port with service type override
			addPort("main", 8080, 8080, constants.ServiceTypeNodePort, false)

			// Set default service type to LoadBalancer
			defaultServiceType.Value = ptr.To(constants.ServiceTypeLoadBalancer)

			// Build should create a service with the overridden service type
			services, err := sb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadName, owner,
				ports, defaultServiceType, calculatedPriorityClassName)
			Expect(err).NotTo(HaveOccurred())
			Expect(services).To(HaveLen(1))

			service := services[0]
			Expect(service.Spec.Type).To(Equal(corev1.ServiceTypeNodePort))
		})

		It("should set protocol when not specified", func() {
			// Add a port without specifying protocol
			addPort("main", 8080, 8080, constants.ServiceTypeClusterIP, false)

			// Build should create a service with TCP protocol by default
			services, err := sb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadName, owner,
				ports, defaultServiceType, calculatedPriorityClassName)
			Expect(err).NotTo(HaveOccurred())
			Expect(services).To(HaveLen(1))

			service := services[0]
			Expect(service.Spec.Ports[0].Protocol).To(Equal(corev1.ProtocolTCP))
		})
	})

	Context("Service type handling", func() {
		It("should use default service type when not specified in port", func() {
			// Add a port without specifying service type
			(*ports.Value)["main"] = flds.PortItemAndRules{
				Value: &flds.PortItemValue{
					ContainerPort: ptr.To(int(8080)),
					HostPort:      ptr.To(int(8080)),
					// No service type specified
				},
			}

			// Set default service type to LoadBalancer
			defaultServiceType.Value = ptr.To(constants.ServiceTypeLoadBalancer)

			// Build should create a service with the default service type
			services, err := sb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadName, owner,
				ports, defaultServiceType, calculatedPriorityClassName)
			Expect(err).NotTo(HaveOccurred())
			Expect(services).To(HaveLen(1))

			service := services[0]
			Expect(service.Spec.Type).To(Equal(corev1.ServiceTypeLoadBalancer))
		})

		It("should skip ports with ServingPort service type", func() {
			// Add a port with ServingPort service type
			addPort("serving", 8080, 8080, constants.ServiceTypeServingPort, false)

			// Add another regular port for comparison
			addPort("regular", 9090, 9090, constants.ServiceTypeClusterIP, false)

			// GetPreview should only include the regular port
			preview, err := sb.GetPreview(ctx, workloadKind, workloadInitials, objectMeta, ports, defaultServiceType)
			Expect(err).NotTo(HaveOccurred())
			Expect(preview).To(HaveLen(1))
			Expect(preview[0].Name).To(ContainSubstring("iw-test-workload-0-clusterip"))

			// Build should only create a service for the regular port
			services, err := sb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadName, owner,
				ports, defaultServiceType, calculatedPriorityClassName)
			Expect(err).NotTo(HaveOccurred())
			Expect(services).To(HaveLen(1))

			service := services[0]
			Expect(service.Spec.Type).To(Equal(corev1.ServiceTypeClusterIP))
			Expect(service.Spec.Ports).To(HaveLen(1))
			Expect(service.Spec.Ports[0].TargetPort.IntVal).To(Equal(int32(9090)))
		})

		It("should handle nil port value", func() {
			// Add a nil port value
			(*ports.Value)["nil"] = flds.PortItemAndRules{
				Value: nil,
			}

			// Add a regular port for comparison
			addPort("regular", 9090, 9090, constants.ServiceTypeClusterIP, false)

			// Build should ignore the nil port and only create a service for the regular port
			services, err := sb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadName, owner,
				ports, defaultServiceType, calculatedPriorityClassName)
			Expect(err).NotTo(HaveOccurred())
			Expect(services).To(HaveLen(1))

			service := services[0]
			Expect(service.Spec.Ports).To(HaveLen(1))
			Expect(service.Spec.Ports[0].TargetPort.IntVal).To(Equal(int32(9090)))
		})

		It("should skip ports with invalid service type", func() {
			// Add a port with an invalid service type
			invalidServiceType := "invalid-type"
			(*ports.Value)["invalid"] = flds.PortItemAndRules{
				Value: &flds.PortItemValue{
					ContainerPort: ptr.To(int(8080)),
					HostPort:      ptr.To(int(8080)),
					ServiceType:   &invalidServiceType,
				},
			}

			// GetPreview should return an empty list
			preview, err := sb.GetPreview(ctx, workloadKind, workloadInitials, objectMeta, ports, defaultServiceType)
			Expect(err).NotTo(HaveOccurred())
			Expect(preview).To(HaveLen(0))

			// Build should return an empty list
			services, err := sb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadName, owner,
				ports, defaultServiceType, calculatedPriorityClassName)
			Expect(err).NotTo(HaveOccurred())
			Expect(services).To(HaveLen(0))
		})
	})

	Context("Metadata handling", func() {
		It("should set controller reference when owner is provided", func() {
			// Add a port
			addPort("main", 8080, 8080, constants.ServiceTypeClusterIP, false)

			// Build should set controller reference
			services, err := sb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadName, owner,
				ports, defaultServiceType, calculatedPriorityClassName)
			Expect(err).NotTo(HaveOccurred())
			Expect(services).To(HaveLen(1))

			service := services[0]
			Expect(service.OwnerReferences).To(HaveLen(1))
			Expect(service.OwnerReferences[0].Name).To(Equal(owner.Name))
			Expect(service.OwnerReferences[0].UID).To(Equal(owner.UID))
		})

		It("should set priority class label for preemptible workload", func() {
			// Add a port
			addPort("main", 8080, 8080, constants.ServiceTypeClusterIP, false)

			// Set preemptible
			calculatedPriorityClassName = constants.InteractivePreemptiblePriorityClass

			// Build should set preemptible priority class label
			services, err := sb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadName, owner,
				ports, defaultServiceType, calculatedPriorityClassName)
			Expect(err).NotTo(HaveOccurred())
			Expect(services).To(HaveLen(1))

			// Check that label was set on objectMeta
			Expect(objectMeta.Labels[constants.LabelPriorityClassName]).To(Equal(constants.InteractivePreemptiblePriorityClass))
		})

		It("should set priority class label", func() {
			// Add a port
			addPort("main", 8080, 8080, constants.ServiceTypeClusterIP, false)

			// Set calculatedPriorityClassName
			calculatedPriorityClassName = "train"

			// Build should set preemptible priority class label
			services, err := sb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadName, owner,
				ports, defaultServiceType, calculatedPriorityClassName)
			Expect(err).NotTo(HaveOccurred())
			Expect(services).To(HaveLen(1))

			// Check that label was set on objectMeta
			Expect(objectMeta.Labels[constants.LabelPriorityClassName]).To(Equal("train"))
		})

		It("should set priority class label for non-preemptible workload", func() {
			// Add a port
			addPort("main", 8080, 8080, constants.ServiceTypeClusterIP, false)

			// Not preemptible
			calculatedPriorityClassName = constants.InteractivePriorityClass

			// Build should set non-preemptible priority class label
			services, err := sb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadName, owner,
				ports, defaultServiceType, calculatedPriorityClassName)
			Expect(err).NotTo(HaveOccurred())
			Expect(services).To(HaveLen(1))

			// Check that label was set on objectMeta
			Expect(objectMeta.Labels[constants.LabelPriorityClassName]).To(Equal(constants.InteractivePriorityClass))
		})

		It("should set appropriate labels on services", func() {
			// Add a port
			addPort("main", 8080, 8080, constants.ServiceTypeClusterIP, false)

			// Build should set labels on services
			services, err := sb.Build(ctx, workloadKind, jobType, workloadInitials, objectMeta, workloadName, owner,
				ports, defaultServiceType, calculatedPriorityClassName)
			Expect(err).NotTo(HaveOccurred())
			Expect(services).To(HaveLen(1))

			service := services[0]
			// Labels from objectMeta should be copied to the service
			Expect(service.Labels).To(Equal(objectMeta.Labels))

			// Check that the selector is properly set
			Expect(service.Spec.Selector[constants.LabelRelease]).To(Equal(workloadName))
		})
	})

	Context("service and ports info", func() {
		It("information is collected correctly", func() {
			addPort("port-0", 80, 8080, constants.ServiceTypeClusterIP, false)
			addPort("port-1", 443, 4443, constants.ServiceTypeNodePort, false)
			addPort("port-2", 2, 2222, constants.ServiceTypeNodePort, false)
			addPort("port-3", 3, 3333, constants.ServiceTypeLoadBalancer, false)
			addPort("port-4", 4, 4444, constants.ServiceTypeNodePort, false)
			addPort("port-5", 5, 5555, constants.ServiceTypeClusterIP, false)

			portsPerSvcType := GetPortsPerServiceType("train", "my-workload", ports, nil)

			Expect(portsPerSvcType).To(HaveLen(3))

			Expect(portsPerSvcType[constants.ServiceTypeClusterIP].Name).To(Equal("train-my-workload-0-clusterip"))
			Expect(portsPerSvcType[constants.ServiceTypeLoadBalancer].Name).To(Equal("train-my-workload-1-loadbalancer"))
			Expect(portsPerSvcType[constants.ServiceTypeNodePort].Name).To(Equal("train-my-workload-2-nodeport"))

			Expect(portsPerSvcType[constants.ServiceTypeNodePort].Ports).To(HaveLen(3))
			Expect(portsPerSvcType[constants.ServiceTypeClusterIP].Ports).To(HaveLen(2))
			Expect(portsPerSvcType[constants.ServiceTypeLoadBalancer].Ports).To(HaveLen(1))

			for idx, port := range portsPerSvcType[constants.ServiceTypeNodePort].Ports {
				targetPort := port.TargetPort.IntValue()
				Expect(targetPort).To(Or(Equal(443), Equal(2), Equal(4)))
				Expect(port.Name).To(MatchRegexp(fmt.Sprintf("port-my-workload-%d-[0-5]", targetPort)))
				Expect(portsPerSvcType[constants.ServiceTypeNodePort].PortsNick[idx]).To(MatchRegexp("port-(1|2|4)"))
			}

			for idx, port := range portsPerSvcType[constants.ServiceTypeClusterIP].Ports {
				targetPort := port.TargetPort.IntValue()
				Expect(targetPort).To(Or(Equal(5), Equal(80)))
				Expect(port.Name).To(MatchRegexp(fmt.Sprintf("port-my-workload-%d-[0-5]", targetPort)))
				Expect(portsPerSvcType[constants.ServiceTypeClusterIP].PortsNick[idx]).To(MatchRegexp("port-(0|5)"))
			}

			for idx, port := range portsPerSvcType[constants.ServiceTypeLoadBalancer].Ports {
				targetPort := port.TargetPort.IntValue()
				Expect(targetPort).To(Equal(3))
				Expect(port.Name).To(MatchRegexp(fmt.Sprintf("port-my-workload-%d-[0-5]", targetPort)))
				Expect(portsPerSvcType[constants.ServiceTypeLoadBalancer].PortsNick[idx]).To(Equal("port-3"))
			}

		})
	})
})
