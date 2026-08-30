package run

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"

	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"

	operandsmocks "github.com/run-ai/runai/runai-cluster/cluster/go-operator/mocks/pkg/operands"
)

var _ = Describe("StatusCalculation", func() {
	Context("calculateOperandStatusMap", func() {
		type Mocks struct {
			managedOperands *MockManagedOperands
		}

		var (
			mocks       Mocks
			runaiConfig *runv1.RunaiConfig
			operandMock *operandsmocks.Operand
		)

		const operandName = "test-operand"

		BeforeEach(func() {
			mocks = Mocks{
				managedOperands: NewMockManagedOperands(GinkgoT()),
			}

			runaiConfig = &runv1.RunaiConfig{}
			runaiConfig.SetDefaultsWhereNeeded()
		})

		When("an operand was previously ready", func() {
			BeforeEach(func() {
				runaiConfig.Status.OperandStatuses = map[string]runv1.TimedOperandStatus{
					operandName: {
						OperandStatus: runv1.OperandStatus{
							Ready: true,
						},
						LastTransitionTime: metav1.Unix(0, 0),
					},
				}

				operandMock = operandsmocks.NewOperand(GinkgoT())
				operandMock.On("Name").Return(operandName)
				operandMock.On("Enabled").Return(true)
			})

			When("the operand is still ready", func() {
				BeforeEach(func() {
					operandMock.On("Status", mock.Anything).Return(runv1.OperandStatus{
						Ready: true,
					}, nil)

					operandMock.On("Enabled").Return(true)

					mocks.managedOperands.On("GetRemoteClusterEnabled").Return([]operands.Operand{operandMock})
				})

				It("should return a map with a ready operand and the last transition time that is equal to the previous last transition time", func() {
					// Act
					operandStatusMap, err := calculateOperandStatusMap(context.Background(), runaiConfig, mocks.managedOperands.GetRemoteClusterEnabled())

					// Assert
					Expect(err).To(BeNil())
					Expect(operandStatusMap).To(HaveLen(1))
					Expect(operandStatusMap[operandName].Ready).To(BeTrue())
					Expect(operandStatusMap[operandName].LastTransitionTime).To(Equal(runaiConfig.Status.OperandStatuses[operandName].LastTransitionTime))
				})
			})
			When("the operand is not ready", func() {
				BeforeEach(func() {
					operandMock.On("Status", mock.Anything).Return(runv1.OperandStatus{
						Ready:   false,
						Reasons: []string{"reason1", "reason2"},
					}, nil)

					mocks.managedOperands.On("GetRemoteClusterEnabled").Return([]operands.Operand{operandMock})
				})

				It("should return a map with a not ready operand and the last transition time that is after the previous last transition time", func() {
					// Act
					operandStatusMap, err := calculateOperandStatusMap(context.Background(), runaiConfig, mocks.managedOperands.GetRemoteClusterEnabled())

					// Assert
					Expect(err).To(BeNil())
					Expect(operandStatusMap).To(HaveLen(1))
					Expect(operandStatusMap[operandName].Ready).To(BeFalse())
					Expect(operandStatusMap[operandName].Reasons).To(ConsistOf("reason1", "reason2"))
					Expect(operandStatusMap[operandName].LastTransitionTime.Time).To(BeTemporally(">", runaiConfig.Status.OperandStatuses[operandName].LastTransitionTime.Time))
				})
			})
		})

		// Write a mirror case when the operand was not ready before
		When("an operand was previously not ready", func() {
			BeforeEach(func() {
				runaiConfig.Status.OperandStatuses = map[string]runv1.TimedOperandStatus{
					operandName: {
						OperandStatus: runv1.OperandStatus{
							Ready:   false,
							Reasons: []string{"reason1", "reason2"},
						},
						LastTransitionTime: metav1.Unix(0, 0),
					},
				}

				operandMock = operandsmocks.NewOperand(GinkgoT())
				operandMock.On("Name").Return(operandName)
				operandMock.On("Enabled").Return(true)
			})

			When("the operand is still not ready", func() {
				BeforeEach(func() {
					operandMock.On("Status", mock.Anything).Return(runv1.OperandStatus{
						Ready:   false,
						Reasons: []string{"reason1", "reason2"},
					}, nil)

					mocks.managedOperands.On("GetRemoteClusterEnabled").Return([]operands.Operand{operandMock})
				})

				It("should return a map with a not ready operand and the last transition time that is equal to the previous last transition time", func() {
					// Act
					operandStatusMap, err := calculateOperandStatusMap(context.Background(), runaiConfig, mocks.managedOperands.GetRemoteClusterEnabled())

					// Assert
					Expect(err).To(BeNil())
					Expect(operandStatusMap).To(HaveLen(1))
					Expect(operandStatusMap[operandName].Ready).To(BeFalse())
					Expect(operandStatusMap[operandName].Reasons).To(ConsistOf("reason1", "reason2"))
					Expect(operandStatusMap[operandName].LastTransitionTime).To(Equal(runaiConfig.Status.OperandStatuses[operandName].LastTransitionTime))
				})
			})
			When("the operand is ready", func() {
				BeforeEach(func() {
					operandMock.On("Status", mock.Anything).Return(runv1.OperandStatus{
						Ready: true,
					}, nil)

					mocks.managedOperands.On("GetRemoteClusterEnabled").Return([]operands.Operand{operandMock})
				})

				It("should return a map with a ready operand and the last transition time that is after the previous last transition time", func() {
					// Act
					operandStatusMap, err := calculateOperandStatusMap(context.Background(), runaiConfig, mocks.managedOperands.GetRemoteClusterEnabled())

					// Assert
					Expect(err).To(BeNil())
					Expect(operandStatusMap).To(HaveLen(1))
					Expect(operandStatusMap[operandName].Ready).To(BeTrue())
					Expect(operandStatusMap[operandName].LastTransitionTime.Time).To(BeTemporally(">", runaiConfig.Status.OperandStatuses[operandName].LastTransitionTime.Time))
				})
			})
		})
	})
})
