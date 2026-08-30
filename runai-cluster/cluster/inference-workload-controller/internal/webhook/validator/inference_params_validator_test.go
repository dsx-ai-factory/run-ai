package validator

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	workloadUtils "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource/workload"
	"github.com/run-ai/runai/runai-cluster/common/utils"
	"k8s.io/utils/ptr"
)

const (
	WorkloadName        = "test-workload"
	runaiClusterVersion = "1.1.2"
	namespace           = "runai-default"
	Image               = "test-Image"
)

var _ = Describe("Inference Parameters Validator", func() {
	var (
		workload  *jobs.JobWorkload
		validator *InferenceParamsValidator
		ctx       context.Context
	)

	Context("validate basic fields", func() {
		BeforeEach(func() {
			workload = workloadUtils.NewJobWorkloadBuilder().WithName(WorkloadName).WithImage(Image).WithInference().WithNamespace(namespace).WithLabels(nil).Build()
			ctx = context.Background()
			validator = NewInferenceParamsValidator()

			workload.Spec.NetworkingFields.Ports = &field_types.PortsField{
				Value: &map[string]field_types.PortItemAndRules{
					"serving-port": {
						Value: &field_types.PortItemValue{
							ServiceType:   ptr.To(constants.ServiceTypeServingPort),
							ContainerPort: ptr.To(8080),
						},
					},
				},
			}

			workload.Spec.InferenceFields = field_groups.InferenceFields{
				AutoScalingFields: field_groups.AutoScalingFields{},
			}
		})

		Context("validate serving port", func() {
			It("should be valid if provided", func() {
				// (serving port is set in BeforeEach)
				err := validator.Validate(ctx, workload)
				Expect(err).To(BeNil())
			})

			It("should be valid if provided with service type in the root level", func() {
				workload.Spec.NetworkingFields.Ports = &field_types.PortsField{
					Value: &map[string]field_types.PortItemAndRules{
						"serving-port": {
							Value: &field_types.PortItemValue{
								ContainerPort: ptr.To(8080),
								// ServiceType not set here!
							},
						},
					},
				}

				workload.Spec.NetworkingFields.ServiceType = &field_types.StringField{
					Value: ptr.To(constants.ServiceTypeServingPort),
				}

				err := validator.Validate(ctx, workload)
				Expect(err).To(BeNil())
			})

			It("should be error if not provided", func() {
				workload.Spec.NetworkingFields.Ports = nil
				err := validator.Validate(ctx, workload)

				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("inference workload should have exactly one serving port"))
			})

			It("should be error if provided more than once", func() {
				workload.Spec.NetworkingFields.Ports = &field_types.PortsField{
					Value: &map[string]field_types.PortItemAndRules{
						"serving-port1": {
							Value: &field_types.PortItemValue{
								ServiceType:   ptr.To(constants.ServiceTypeServingPort),
								ContainerPort: ptr.To(8080),
							},
						},
						"serving-port2": {
							Value: &field_types.PortItemValue{
								ServiceType:   ptr.To(constants.ServiceTypeServingPort),
								ContainerPort: ptr.To(8081),
							},
						},
					},
				}

				err := validator.Validate(ctx, workload)

				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("inference workload should have exactly one serving port"))
			})
		})

		Context("validate serving configurationFields fields", func() {
			It("initializationTimeoutSeconds - should be an invalid workload with negative initializationTimeoutSeconds", func() {
				workload.Spec.InferenceFields.ServingConfigurationFields.InitializationTimeoutSeconds = &field_types.IntField{
					Value: ptr.To(-10),
				}
				err := validator.Validate(ctx, workload)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("initializationTimeoutSeconds must be greater than 0"))
			})

			It("initializationTimeoutSeconds - should be an invalid workload with initializationTimeoutSeconds of 0", func() {
				workload.Spec.InferenceFields.ServingConfigurationFields.InitializationTimeoutSeconds = &field_types.IntField{
					Value: ptr.To(0),
				}
				err := validator.Validate(ctx, workload)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("initializationTimeoutSeconds must be greater than 0"))
			})

			It("RequestTimeoutSeconds - should be an invalid workload with negative RequestTimeoutSeconds", func() {
				workload.Spec.InferenceFields.ServingConfigurationFields.RequestTimeoutSeconds = &field_types.IntField{
					Value: ptr.To(-10),
				}
				err := validator.Validate(ctx, workload)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("requestTimeoutSeconds must be greater than 0"))
			})

			It("RequestTimeoutSeconds - should be an invalid workload with RequestTimeoutSeconds of 0", func() {
				workload.Spec.InferenceFields.ServingConfigurationFields.RequestTimeoutSeconds = &field_types.IntField{
					Value: ptr.To(0),
				}
				err := validator.Validate(ctx, workload)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("requestTimeoutSeconds must be greater than 0"))
			})
		})

		Context("validate basic autoscaling fields", func() {
			It("various legal fields - should be an valid", func() {
				workload.Spec.InferenceFields.AutoScalingFields.MinScale = &field_types.IntField{
					Value: ptr.To(1),
				}
				workload.Spec.InferenceFields.AutoScalingFields.MaxScale = &field_types.IntField{
					Value: ptr.To(3),
				}
				workload.Spec.InferenceFields.AutoScalingFields.Metric = &field_types.StringField{
					Value: ptr.To("throughput"),
				}
				workload.Spec.InferenceFields.AutoScalingFields.TargetUtilizationPercentage = &field_types.NumberField{
					Value: ptr.To("88.88"),
				}
				workload.Spec.InferenceFields.AutoScalingFields.InitialScale = &field_types.IntField{
					Value: ptr.To(2),
				}
				workload.Spec.InferenceFields.AutoScalingFields.ActivationScale = &field_types.IntField{
					Value: ptr.To(2),
				}
				workload.Spec.InferenceFields.AutoScalingFields.ScaleDownDelaySeconds = &field_types.IntField{
					Value: ptr.To(120),
				}
				workload.Spec.InferenceFields.AutoScalingFields.ConcurrencyHardLimit = &field_types.IntField{
					Value: ptr.To(2),
				}
				err := validator.Validate(ctx, workload)
				Expect(err).ToNot(HaveOccurred())
			})

			It("various legal fields with scale to zero - should be an valid", func() {
				workload.Spec.InferenceFields.AutoScalingFields.MinScale = &field_types.IntField{
					Value: ptr.To(0),
				}
				workload.Spec.InferenceFields.AutoScalingFields.MaxScale = &field_types.IntField{
					Value: ptr.To(1),
				}
				workload.Spec.InferenceFields.AutoScalingFields.Metric = &field_types.StringField{
					Value: ptr.To("throughput"),
				}
				workload.Spec.InferenceFields.AutoScalingFields.ScaleToZeroRetentionSeconds = &field_types.IntField{
					Value: ptr.To(120),
				}
				workload.Spec.InferenceFields.AutoScalingFields.ScaleDownDelaySeconds = &field_types.IntField{
					Value: ptr.To(60),
				}
				workload.Spec.InferenceFields.AutoScalingFields.InitialScale = &field_types.IntField{
					Value: ptr.To(0),
				}
				workload.Spec.InferenceFields.AutoScalingFields.ActivationScale = &field_types.IntField{
					Value: ptr.To(1),
				}
				workload.Spec.InferenceFields.AutoScalingFields.ConcurrencyHardLimit = &field_types.IntField{
					Value: ptr.To(0),
				}
				err := validator.Validate(ctx, workload)
				Expect(err).ToNot(HaveOccurred())
			})

			It("minScale and maxScale - should be an invalid workload with minScale > maxScale", func() {
				workload.Spec.InferenceFields.AutoScalingFields.MinScale = &field_types.IntField{
					Value: ptr.To(5),
				}
				workload.Spec.InferenceFields.AutoScalingFields.MaxScale = &field_types.IntField{
					Value: ptr.To(3),
				}
				workload.Spec.InferenceFields.AutoScalingFields.Metric = &field_types.StringField{
					Value: ptr.To("latency"),
				}
				err := validator.Validate(ctx, workload)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("maxScale must be greater than or equal to minScale"))
			})

			It("minScale and maxScale - should be an invalid workload with minScale < 0", func() {
				workload.Spec.InferenceFields.AutoScalingFields.MinScale = &field_types.IntField{
					Value: ptr.To(-1),
				}
				workload.Spec.InferenceFields.AutoScalingFields.MaxScale = &field_types.IntField{
					Value: ptr.To(3),
				}
				workload.Spec.InferenceFields.AutoScalingFields.Metric = &field_types.StringField{
					Value: ptr.To("latency"),
				}
				err := validator.Validate(ctx, workload)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("minScale must be non-negative"))
			})

			It("minScale = 0 with latency metric - should be an invalid workload with", func() {
				workload.Spec.InferenceFields.AutoScalingFields.MinScale = &field_types.IntField{
					Value: ptr.To(0),
				}
				workload.Spec.InferenceFields.AutoScalingFields.MaxScale = &field_types.IntField{
					Value: ptr.To(1),
				}
				workload.Spec.InferenceFields.AutoScalingFields.Metric = &field_types.StringField{
					Value: ptr.To("latency"),
				}
				err := validator.Validate(ctx, workload)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("minScale of 0 is only allowed when using the concurrency or throughput metrics"))
			})

			It("Metric - should be an invalid workload with no such a metric", func() {
				workload.Spec.InferenceFields.AutoScalingFields.MinScale = &field_types.IntField{
					Value: ptr.To(1),
				}
				workload.Spec.InferenceFields.AutoScalingFields.MaxScale = &field_types.IntField{
					Value: ptr.To(5),
				}
				workload.Spec.InferenceFields.AutoScalingFields.Metric = &field_types.StringField{
					Value: ptr.To("no-such-metric"),
				}
				err := validator.Validate(ctx, workload)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("'no-such-metric' is not supported"))
			})

			It("scaleToZeroRetentionSeconds - should be an invalid workload with no MinScale", func() {
				workload.Spec.InferenceFields.AutoScalingFields.ScaleToZeroRetentionSeconds = &field_types.IntField{
					Value: ptr.To(60),
				}
				err := validator.Validate(ctx, workload)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("minScale must be set to 0 if scaleToZeroRetentionSeconds is set"))
			})

			It("scaleToZeroRetentionSeconds - should be an invalid workload with MinScale > 0", func() {
				workload.Spec.InferenceFields.AutoScalingFields.MinScale = &field_types.IntField{
					Value: ptr.To(1),
				}
				workload.Spec.InferenceFields.AutoScalingFields.ScaleToZeroRetentionSeconds = &field_types.IntField{
					Value: ptr.To(60),
				}
				err := validator.Validate(ctx, workload)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("minScale must be set to 0 if scaleToZeroRetentionSeconds is set"))
			})

			It("scaleToZeroRetentionSeconds - should be an invalid workload with negative ScaleToZeroRetentionSeconds", func() {
				workload.Spec.InferenceFields.AutoScalingFields.MinScale = &field_types.IntField{
					Value: ptr.To(0),
				}
				workload.Spec.InferenceFields.AutoScalingFields.ScaleToZeroRetentionSeconds = &field_types.IntField{
					Value: ptr.To(-60),
				}
				err := validator.Validate(ctx, workload)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("scaleToZeroRetentionSeconds must be non-negative"))
			})

			It("scaleToZeroRetentionSeconds - should be an invalid workload with a too big ScaleToZeroRetentionSeconds", func() {
				workload.Spec.InferenceFields.AutoScalingFields.MinScale = &field_types.IntField{
					Value: ptr.To(0),
				}
				workload.Spec.InferenceFields.AutoScalingFields.ScaleToZeroRetentionSeconds = &field_types.IntField{
					Value: ptr.To(3601),
				}
				err := validator.Validate(ctx, workload)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("scaleToZeroRetentionSeconds must be less than or equal to 3600 (1 hour)"))
			})

			It("scaleToZeroRetentionSeconds - should be a valid workload with valid MinScale and ScaleToZeroRetentionSeconds", func() {
				workload.Spec.InferenceFields.AutoScalingFields.MinScale = &field_types.IntField{
					Value: ptr.To(0),
				}
				workload.Spec.InferenceFields.AutoScalingFields.ScaleToZeroRetentionSeconds = &field_types.IntField{
					Value: ptr.To(60),
				}
				err := validator.Validate(ctx, workload)
				Expect(err).To(BeNil())
			})

			It("scaleToZeroRetentionSeconds - should be a valid workload with valid MinScale and ScaleToZeroRetentionSeconds(0)", func() {
				workload.Spec.InferenceFields.AutoScalingFields.MinScale = &field_types.IntField{
					Value: ptr.To(0),
				}
				workload.Spec.InferenceFields.AutoScalingFields.ScaleToZeroRetentionSeconds = &field_types.IntField{
					Value: ptr.To(0),
				}
				err := validator.Validate(ctx, workload)
				Expect(err).To(BeNil())
			})

			It("scaleDownDelaySeconds - should be an invalid workload with negative scaleDownDelaySeconds", func() {
				workload.Spec.InferenceFields.AutoScalingFields.ScaleDownDelaySeconds = &field_types.IntField{
					Value: ptr.To(-60),
				}
				err := validator.Validate(ctx, workload)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("scaleDownDelaySeconds must be non-negative"))
			})

			It("scaleDownDelaySeconds - should be an invalid workload with a too big scaleDownDelaySeconds", func() {
				workload.Spec.InferenceFields.AutoScalingFields.ScaleDownDelaySeconds = &field_types.IntField{
					Value: ptr.To(3601),
				}
				err := validator.Validate(ctx, workload)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("scaleDownDelaySeconds must be less than or equal to 3600 (1 hour)"))
			})

			It("targetUtilizationPercentage - should be an invalid workload if < 1", func() {
				workload.Spec.InferenceFields.AutoScalingFields.TargetUtilizationPercentage = &field_types.NumberField{
					Value: ptr.To("0.1"),
				}
				err := validator.Validate(ctx, workload)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("targetUtilizationPercentage must be between 1 and 100"))
			})

			It("targetUtilizationPercentage - should be an invalid workload if > 100", func() {
				workload.Spec.InferenceFields.AutoScalingFields.TargetUtilizationPercentage = &field_types.NumberField{
					Value: ptr.To("102"),
				}
				err := validator.Validate(ctx, workload)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("targetUtilizationPercentage must be between 1 and 100"))
			})

			It("targetUtilizationPercentage - should be an invalid workload if < 0", func() {
				workload.Spec.InferenceFields.AutoScalingFields.TargetUtilizationPercentage = &field_types.NumberField{
					Value: ptr.To("-12"),
				}
				err := validator.Validate(ctx, workload)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("targetUtilizationPercentage must be between 1 and 100"))
			})

			It("targetUtilizationPercentage - should be an invalid workload if used with a non-kpa metric", func() {
				workload.Spec.InferenceFields.AutoScalingFields.TargetUtilizationPercentage = &field_types.NumberField{
					Value: ptr.To("80"),
				}
				workload.Spec.InferenceFields.AutoScalingFields.Metric = &field_types.StringField{
					Value: ptr.To("latency"),
				}
				err := validator.Validate(ctx, workload)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("targetUtilizationPercentage is only allowed when using the concurrency or throughput metrics"))
			})

			It("initialScale - should be an invalid workload with initialScale < 0", func() {
				workload.Spec.InferenceFields.AutoScalingFields.InitialScale = &field_types.IntField{
					Value: ptr.To(-3),
				}
				err := validator.Validate(ctx, workload)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("initialScale must be non-negative"))
			})

			It("initialScale - should be an invalid workload with initialScale < minScale", func() {
				workload.Spec.InferenceFields.AutoScalingFields.MinScale = &field_types.IntField{
					Value: ptr.To(3),
				}
				workload.Spec.InferenceFields.AutoScalingFields.MaxScale = &field_types.IntField{
					Value: ptr.To(5),
				}
				workload.Spec.InferenceFields.AutoScalingFields.InitialScale = &field_types.IntField{
					Value: ptr.To(2),
				}
				err := validator.Validate(ctx, workload)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("initialScale must be greater than or equal to minScale"))
			})

			It("initialScale - should be an invalid workload with initialScale > maxScale", func() {
				workload.Spec.InferenceFields.AutoScalingFields.MinScale = &field_types.IntField{
					Value: ptr.To(3),
				}
				workload.Spec.InferenceFields.AutoScalingFields.MaxScale = &field_types.IntField{
					Value: ptr.To(5),
				}
				workload.Spec.InferenceFields.AutoScalingFields.InitialScale = &field_types.IntField{
					Value: ptr.To(6),
				}
				err := validator.Validate(ctx, workload)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("initialScale must be less than or equal to maxScale"))
			})

			It("activationScale - should be an invalid workload with activationScale of 0", func() {
				workload.Spec.InferenceFields.AutoScalingFields.MinScale = &field_types.IntField{
					Value: ptr.To(0),
				}
				workload.Spec.InferenceFields.AutoScalingFields.MaxScale = &field_types.IntField{
					Value: ptr.To(1),
				}
				workload.Spec.InferenceFields.AutoScalingFields.ActivationScale = &field_types.IntField{
					Value: ptr.To(0),
				}
				err := validator.Validate(ctx, workload)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("activationScale must be greater than 0"))
			})

			It("activationScale - should be an invalid workload with activationScale < minScale", func() {
				workload.Spec.InferenceFields.AutoScalingFields.MinScale = &field_types.IntField{
					Value: ptr.To(3),
				}
				workload.Spec.InferenceFields.AutoScalingFields.MaxScale = &field_types.IntField{
					Value: ptr.To(5),
				}
				workload.Spec.InferenceFields.AutoScalingFields.ActivationScale = &field_types.IntField{
					Value: ptr.To(2),
				}
				err := validator.Validate(ctx, workload)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("activationScale must be greater than or equal to minScale"))
			})

			It("activationScale - should be an invalid workload with activationScale > maxScale", func() {
				workload.Spec.InferenceFields.AutoScalingFields.MinScale = &field_types.IntField{
					Value: ptr.To(3),
				}
				workload.Spec.InferenceFields.AutoScalingFields.MaxScale = &field_types.IntField{
					Value: ptr.To(5),
				}
				workload.Spec.InferenceFields.AutoScalingFields.ActivationScale = &field_types.IntField{
					Value: ptr.To(6),
				}
				err := validator.Validate(ctx, workload)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("activationScale must be less than or equal to maxScale"))
			})

			It("concurrencyHardLimit - should be an invalid workload with concurrencyHardLimit < 0", func() {
				workload.Spec.InferenceFields.AutoScalingFields.ConcurrencyHardLimit = &field_types.IntField{
					Value: ptr.To(-1),
				}
				err := validator.Validate(ctx, workload)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("concurrencyHardLimit must be non-negative"))
			})

		})

		Context("validate pvcs", func() {
			It("should be valid if no ephemeral pvc is used", func() {
				workload.Spec.StorageFields.Pvcs = &field_types.PvcsField{
					Value: &map[string]field_types.PvcsItemAndRules{
						"pvc1": {
							Value: &field_types.PvcsItemValue{
								ClaimName:    utils.Ref("claim1"),
								StorageClass: utils.Ref("sc1"),
								Size:         utils.Ref("1000"),
								Path:         utils.Ref("/path"),
							},
						},
					},
				}

				err := validator.Validate(ctx, workload)
				Expect(err).To(BeNil())
			})

			It("should be error if ephemeral is used", func() {
				workload.Spec.StorageFields.Pvcs = &field_types.PvcsField{
					Value: &map[string]field_types.PvcsItemAndRules{
						"pvc1": {
							Value: &field_types.PvcsItemValue{
								ClaimName:    utils.Ref("claim1"),
								StorageClass: utils.Ref("sc1"),
								Size:         utils.Ref("1000"),
								Path:         utils.Ref("/path"),
								Ephemeral:    ptr.To(true),
							},
						},
					},
				}

				err := validator.Validate(ctx, workload)

				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("ephemeral pvc is not supported"))
			})
		})

	})
})
