package builders

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-application-packages/datastructs"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

var _ = Describe("Ports tests", func() {
	var (
		ctx               context.Context
		networkingFields  *field_groups.NetworkingFields
		allowServingPort  bool
		baseContainerPort int32
	)

	BeforeEach(func() {
		ctx = context.Background()
		networkingFields = &field_groups.NetworkingFields{}
		allowServingPort = true
		baseContainerPort = 8080
	})

	Context("Empty and nil cases", func() {
		It("should return empty result for empty networking fields", func() {
			result, err := GetPorts(ctx, networkingFields, allowServingPort)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeEmpty())
		})
	})

	Context("Port mapping", func() {
		It("should create containerPort for single port", func() {
			// Setup a single port
			portsMap := map[string]flds.PortItemAndRules{
				"http": {
					Value: &flds.PortItemValue{
						ContainerPort: ptr.To(int(baseContainerPort)),
					},
				},
			}
			networkingFields.Ports = flds.NewPortsField(portsMap)

			result, err := GetPorts(ctx, networkingFields, allowServingPort)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(1))
			Expect(result[0].ContainerPort).To(Equal(baseContainerPort))
			Expect(result[0].Name).To(Equal("port-8080"))
			Expect(result[0].Protocol).To(Equal(corev1.ProtocolTCP))
		})

		It("should skip nil port values", func() {
			// Setup ports with a nil value
			portsMap := map[string]flds.PortItemAndRules{
				"http": {
					Value: nil,
				},
				"api": {
					Value: &flds.PortItemValue{
						ContainerPort: ptr.To(int(baseContainerPort)),
					},
				},
			}
			networkingFields.Ports = flds.NewPortsField(portsMap)

			result, err := GetPorts(ctx, networkingFields, allowServingPort)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(1))
			Expect(result[0].ContainerPort).To(Equal(baseContainerPort))
		})

		It("should create containerPorts for multiple different ports", func() {
			// Setup multiple ports
			portsMap := map[string]flds.PortItemAndRules{
				"http": {
					Value: &flds.PortItemValue{
						ContainerPort: ptr.To(int(baseContainerPort)),
					},
				},
				"metrics": {
					Value: &flds.PortItemValue{
						ContainerPort: ptr.To(int(baseContainerPort + 1)),
					},
				},
			}
			networkingFields.Ports = flds.NewPortsField(portsMap)

			result, err := GetPorts(ctx, networkingFields, allowServingPort)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(2))

			// Verify containerPorts are added in order and with correct values
			portValues := []int32{result[0].ContainerPort, result[1].ContainerPort}
			Expect(portValues).To(ContainElements(baseContainerPort, baseContainerPort+1))
		})

		It("should deduplicate identical container ports", func() {
			// Setup duplicate port values
			portsMap := map[string]flds.PortItemAndRules{
				"http": {
					Value: &flds.PortItemValue{
						ContainerPort: ptr.To(int(baseContainerPort)),
					},
				},
				"http-alt": {
					Value: &flds.PortItemValue{
						ContainerPort: ptr.To(int(baseContainerPort)), // Same port value
					},
				},
			}
			networkingFields.Ports = flds.NewPortsField(portsMap)

			result, err := GetPorts(ctx, networkingFields, allowServingPort)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(1)) // Only one port should be added
			Expect(result[0].ContainerPort).To(Equal(baseContainerPort))
		})
	})

	Context("Service Types", func() {
		It("should return error when serving port is not allowed", func() {
			// Setup a serving port
			portsMap := map[string]flds.PortItemAndRules{
				"serve": {
					Value: &flds.PortItemValue{
						ContainerPort: ptr.To(int(baseContainerPort)),
						ServiceType:   ptr.To(constants.ServiceTypeServingPort),
					},
				},
			}
			networkingFields.Ports = flds.NewPortsField(portsMap)

			// Disallow serving port
			allowServingPort = false

			_, err := GetPorts(ctx, networkingFields, allowServingPort)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("serving port is not allowed"))
		})

		It("should allow serving port when allowed", func() {
			// Setup a serving port
			portsMap := map[string]flds.PortItemAndRules{
				"serve": {
					Value: &flds.PortItemValue{
						ContainerPort: ptr.To(int(baseContainerPort)),
						ServiceType:   ptr.To(constants.ServiceTypeServingPort),
					},
				},
			}
			networkingFields.Ports = flds.NewPortsField(portsMap)

			// Allow serving port
			allowServingPort = true

			result, err := GetPorts(ctx, networkingFields, allowServingPort)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(1))
			Expect(result[0].ContainerPort).To(Equal(baseContainerPort))
		})

		It("should use default service type if not specified", func() {
			// Setup a port without service type
			portsMap := map[string]flds.PortItemAndRules{
				"http": {
					Value: &flds.PortItemValue{
						ContainerPort: ptr.To(int(baseContainerPort)),
						// No service type specified
					},
				},
			}
			networkingFields.Ports = flds.NewPortsField(portsMap)

			// Set default service type
			defaultServiceType := constants.ServiceTypeClusterIP
			networkingFields.ServiceType = flds.NewStringField(defaultServiceType)

			result, err := GetPorts(ctx, networkingFields, allowServingPort)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(1))
			Expect(result[0].ContainerPort).To(Equal(baseContainerPort))
		})
	})

	Context("Node Port validation", func() {
		It("should return error for NodePort below valid range", func() {
			lowPort := minHostPort - 1 // Below minimum

			portsMap := map[string]flds.PortItemAndRules{
				"node": {
					Value: &flds.PortItemValue{
						ContainerPort: ptr.To(int(baseContainerPort)),
						ServiceType:   ptr.To(constants.ServiceTypeNodePort),
						HostPort:      ptr.To(int(lowPort)),
					},
				},
			}
			networkingFields.Ports = flds.NewPortsField(portsMap)

			_, err := GetPorts(ctx, networkingFields, allowServingPort)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("NodePort must be between"))
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("%d", minHostPort)))
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("%d", maxHostPort)))
		})

		It("should return error for NodePort above valid range", func() {
			highPort := maxHostPort + 1 // Above maximum

			portsMap := map[string]flds.PortItemAndRules{
				"node": {
					Value: &flds.PortItemValue{
						ContainerPort: ptr.To(int(baseContainerPort)),
						ServiceType:   ptr.To(constants.ServiceTypeNodePort),
						HostPort:      ptr.To(int(highPort)),
					},
				},
			}
			networkingFields.Ports = flds.NewPortsField(portsMap)

			_, err := GetPorts(ctx, networkingFields, allowServingPort)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("NodePort must be between"))
		})

		It("should accept NodePort within valid range", func() {
			validPort := (minHostPort + maxHostPort) / 2 // Middle of range

			portsMap := map[string]flds.PortItemAndRules{
				"node": {
					Value: &flds.PortItemValue{
						ContainerPort: ptr.To(int(baseContainerPort)),
						ServiceType:   ptr.To(constants.ServiceTypeNodePort),
						HostPort:      ptr.To(int(validPort)),
					},
				},
			}
			networkingFields.Ports = flds.NewPortsField(portsMap)

			result, err := GetPorts(ctx, networkingFields, allowServingPort)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(1))
			Expect(result[0].ContainerPort).To(Equal(baseContainerPort))
		})

		It("should skip NodePort validation if HostPort is nil", func() {
			portsMap := map[string]flds.PortItemAndRules{
				"node": {
					Value: &flds.PortItemValue{
						ContainerPort: ptr.To(int(baseContainerPort)),
						ServiceType:   ptr.To(constants.ServiceTypeNodePort),
						HostPort:      nil, // No host port specified
					},
				},
			}
			networkingFields.Ports = flds.NewPortsField(portsMap)

			result, err := GetPorts(ctx, networkingFields, allowServingPort)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(1))
			Expect(result[0].ContainerPort).To(Equal(baseContainerPort))
		})
	})

	Context("Exposed URLs", func() {
		It("should add ports from exposed URLs", func() {
			// Setup exposed URLs
			exposedUrlsMap := map[string]flds.ExposedUrlItemAndRules{
				"web": {
					Value: &flds.ExposedUrlItemValue{
						ContainerPort: ptr.To(int32(8888)),
					},
				},
			}
			networkingFields.ExposedUrls = flds.NewExposedUrlField(exposedUrlsMap)

			result, err := GetPorts(ctx, networkingFields, allowServingPort)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(1))
			Expect(result[0].ContainerPort).To(Equal(int32(8888)))
			Expect(result[0].Name).To(Equal("port-8888"))
		})

		It("should use DefaultIngressHttpPort when exposed URL container port is nil", func() {
			// Setup exposed URL without container port
			exposedUrlsMap := map[string]flds.ExposedUrlItemAndRules{
				"web": {
					Value: &flds.ExposedUrlItemValue{
						// No container port specified
					},
				},
			}
			networkingFields.ExposedUrls = flds.NewExposedUrlField(exposedUrlsMap)

			result, err := GetPorts(ctx, networkingFields, allowServingPort)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(1))
			Expect(result[0].ContainerPort).To(Equal(int32(DefaultIngressHttpPort)))
			Expect(result[0].Name).To(Equal(fmt.Sprintf("port-%d", DefaultIngressHttpPort)))
		})

		It("should skip nil exposed URL values", func() {
			// Setup exposed URLs with a nil value
			exposedUrlsMap := map[string]flds.ExposedUrlItemAndRules{
				"nil-url": {
					Value: nil,
				},
				"valid-url": {
					Value: &flds.ExposedUrlItemValue{
						ContainerPort: ptr.To(int32(8888)),
					},
				},
			}
			networkingFields.ExposedUrls = flds.NewExposedUrlField(exposedUrlsMap)

			result, err := GetPorts(ctx, networkingFields, allowServingPort)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(1)) // Only one valid URL port
			Expect(result[0].ContainerPort).To(Equal(int32(8888)))
		})

		It("should deduplicate ports from exposed URLs", func() {
			// Setup multiple exposed URLs with same port
			exposedUrlsMap := map[string]flds.ExposedUrlItemAndRules{
				"url1": {
					Value: &flds.ExposedUrlItemValue{
						ContainerPort: ptr.To(int32(8888)),
					},
				},
				"url2": {
					Value: &flds.ExposedUrlItemValue{
						ContainerPort: ptr.To(int32(8888)), // Same port
					},
				},
			}
			networkingFields.ExposedUrls = flds.NewExposedUrlField(exposedUrlsMap)

			result, err := GetPorts(ctx, networkingFields, allowServingPort)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(1)) // Only one port should be added
			Expect(result[0].ContainerPort).To(Equal(int32(8888)))
		})
	})

	Context("Combined port sources", func() {
		It("should combine ports from both maps and deduplicate", func() {
			// Setup both ports and exposed URLs
			portsMap := map[string]flds.PortItemAndRules{
				"http": {
					Value: &flds.PortItemValue{
						ContainerPort: ptr.To(int(baseContainerPort)),
					},
				},
				"metrics": {
					Value: &flds.PortItemValue{
						ContainerPort: ptr.To(int(baseContainerPort + 1)),
					},
				},
			}
			networkingFields.Ports = flds.NewPortsField(portsMap)

			// Add exposed URLs with one duplicate port
			exposedUrlsMap := map[string]flds.ExposedUrlItemAndRules{
				"web": {
					Value: &flds.ExposedUrlItemValue{
						ContainerPort: ptr.To(baseContainerPort), // Duplicate of "http" port
					},
				},
				"api": {
					Value: &flds.ExposedUrlItemValue{
						ContainerPort: ptr.To(baseContainerPort + 2), // Unique port
					},
				},
			}
			networkingFields.ExposedUrls = flds.NewExposedUrlField(exposedUrlsMap)

			result, err := GetPorts(ctx, networkingFields, allowServingPort)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(3)) // 3 unique ports

			// Verify all unique ports are present
			allContainerPorts := datastructs.NewSet[int32]()
			for _, cp := range result {
				allContainerPorts.Add(cp.ContainerPort)
			}
			Expect(allContainerPorts.Has(baseContainerPort)).To(BeTrue())
			Expect(allContainerPorts.Has(baseContainerPort + 1)).To(BeTrue())
			Expect(allContainerPorts.Has(baseContainerPort + 2)).To(BeTrue())
		})
	})
})
