package distributed

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/common/extensions"
)

type mockExtensionsClient struct {
	listFunc func(ctx context.Context, extensions []extensions.Type) (extensions.Group, error)
}

func (m *mockExtensionsClient) Get(ctx context.Context, extension extensions.Type) (*extensions.Extension, error) {
	return nil, errors.New("not implemented")
}

func (m *mockExtensionsClient) List(ctx context.Context, extensions []extensions.Type) (extensions.Group, error) {
	return m.listFunc(ctx, extensions)
}

var _ = Describe("StatusChangedPoll with multiple constants", func() {
	Context("Poll Method", func() {
		When("the status of Jax changes", func() {
			It("should return true", func() {
				initialStatus := extensions.Group{
					Jax: {Status: extensions.Status{Enabled: true, Capabilities: map[extensions.CapabilityType]bool{Suspension: true}}},
				}
				mockCurrent := extensions.Group{
					Jax: {Status: extensions.Status{Enabled: false, Capabilities: map[extensions.CapabilityType]bool{Suspension: true}}},
				}

				mockClient := &mockExtensionsClient{
					listFunc: func(ctx context.Context, extensions []extensions.Type) (extensions.Group, error) {
						return mockCurrent, nil
					},
				}

				poll := &StatusChangedPoll{
					jobExtensionsStatus: initialStatus,
					extensionsClient:    mockClient,
				}

				result, err := poll.Poll(context.Background())
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeTrue())
			})
			Context("when the capabilities of Jax changed", func() {
				It("should return true", func() {
					initialStatus := extensions.Group{
						Jax: {Status: extensions.Status{Enabled: true, Capabilities: map[extensions.CapabilityType]bool{Suspension: true}}},
					}
					mockCurrent := extensions.Group{
						Jax: {Status: extensions.Status{Enabled: true, Capabilities: map[extensions.CapabilityType]bool{Suspension: false}}},
					}
					mockClient := &mockExtensionsClient{
						listFunc: func(ctx context.Context, extensions []extensions.Type) (extensions.Group, error) {
							return mockCurrent, nil
						},
					}
					poll := &StatusChangedPoll{
						jobExtensionsStatus: initialStatus,
						extensionsClient:    mockClient,
					}
					result, err := poll.Poll(context.Background())
					Expect(err).NotTo(HaveOccurred())
					Expect(result).To(BeTrue())
				})
			})
		})

		When("a new extension (Pytorch) is added", func() {
			It("should return true", func() {
				initialStatus := extensions.Group{
					Jax: {Status: extensions.Status{Enabled: true, Capabilities: map[extensions.CapabilityType]bool{Suspension: true}}},
				}
				mockCurrent := extensions.Group{
					Jax:     {Status: extensions.Status{Enabled: true, Capabilities: map[extensions.CapabilityType]bool{Suspension: true}}},
					Pytorch: {Status: extensions.Status{Enabled: true, Capabilities: map[extensions.CapabilityType]bool{Suspension: true}}},
				}

				mockClient := &mockExtensionsClient{
					listFunc: func(ctx context.Context, extensions []extensions.Type) (extensions.Group, error) {
						return mockCurrent, nil
					},
				}

				poll := &StatusChangedPoll{
					jobExtensionsStatus: initialStatus,
					extensionsClient:    mockClient,
				}

				result, err := poll.Poll(context.Background())
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeTrue())
			})
		})

		When("an extension (Tensorflow) is removed", func() {
			It("should return true", func() {
				initialStatus := extensions.Group{
					Jax:        {Status: extensions.Status{Enabled: true, Capabilities: map[extensions.CapabilityType]bool{Suspension: true}}},
					Tensorflow: {Status: extensions.Status{Enabled: true, Capabilities: map[extensions.CapabilityType]bool{Suspension: true}}},
				}

				mockCurrent := extensions.Group{
					Jax: {Status: extensions.Status{Enabled: true, Capabilities: map[extensions.CapabilityType]bool{Suspension: true}}},
				}

				mockClient := &mockExtensionsClient{
					listFunc: func(ctx context.Context, extensions []extensions.Type) (extensions.Group, error) {
						return mockCurrent, nil
					},
				}

				poll := &StatusChangedPoll{
					jobExtensionsStatus: initialStatus,
					extensionsClient:    mockClient,
				}

				result, err := poll.Poll(context.Background())
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeTrue())
			})
		})

		When("the status of MPI does not change", func() {
			It("should return false", func() {
				initialStatus := extensions.Group{
					MPI: {Status: extensions.Status{Enabled: true, Capabilities: map[extensions.CapabilityType]bool{Suspension: true}}},
				}

				mockCurrent := extensions.Group{
					MPI: {Status: extensions.Status{Enabled: true, Capabilities: map[extensions.CapabilityType]bool{Suspension: true}}},
				}

				mockClient := &mockExtensionsClient{
					listFunc: func(ctx context.Context, extensions []extensions.Type) (extensions.Group, error) {
						return mockCurrent, nil
					},
				}

				poll := &StatusChangedPoll{
					jobExtensionsStatus: initialStatus,
					extensionsClient:    mockClient,
				}

				result, err := poll.Poll(context.Background())
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeFalse())
			})
		})

		When("multiple extensions, including XGBoost, are added", func() {
			It("should return true", func() {
				initialStatus := extensions.Group{
					Jax: {Status: extensions.Status{Enabled: true, Capabilities: map[extensions.CapabilityType]bool{Suspension: true}}},
				}

				mockCurrent := extensions.Group{
					Jax:     {Status: extensions.Status{Enabled: true, Capabilities: map[extensions.CapabilityType]bool{Suspension: true}}},
					XGBoost: {Status: extensions.Status{Enabled: true, Capabilities: map[extensions.CapabilityType]bool{Suspension: true}}},
				}

				mockClient := &mockExtensionsClient{
					listFunc: func(ctx context.Context, extensions []extensions.Type) (extensions.Group, error) {
						return mockCurrent, nil
					},
				}

				poll := &StatusChangedPoll{
					jobExtensionsStatus: initialStatus,
					extensionsClient:    mockClient,
				}

				result, err := poll.Poll(context.Background())
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeTrue())
			})
		})
		When("there are more than three extensions and no changes occur", func() {
			It("should return false", func() {
				initialStatus := extensions.Group{
					Jax:        {Status: extensions.Status{Enabled: true, Capabilities: map[extensions.CapabilityType]bool{Suspension: true}}},
					Pytorch:    {Status: extensions.Status{Enabled: true, Capabilities: map[extensions.CapabilityType]bool{Suspension: true}}},
					Tensorflow: {Status: extensions.Status{Enabled: false, Capabilities: map[extensions.CapabilityType]bool{Suspension: false}}},
					MPI:        {Status: extensions.Status{Enabled: true, Capabilities: map[extensions.CapabilityType]bool{Suspension: false}}},
					XGBoost:    {Status: extensions.Status{Enabled: true, Capabilities: map[extensions.CapabilityType]bool{Suspension: true}}},
				}

				mockCurrent := extensions.Group{
					Jax:        {Status: extensions.Status{Enabled: true, Capabilities: map[extensions.CapabilityType]bool{Suspension: true}}},
					Pytorch:    {Status: extensions.Status{Enabled: true, Capabilities: map[extensions.CapabilityType]bool{Suspension: true}}},
					Tensorflow: {Status: extensions.Status{Enabled: false, Capabilities: map[extensions.CapabilityType]bool{Suspension: false}}},
					MPI:        {Status: extensions.Status{Enabled: true, Capabilities: map[extensions.CapabilityType]bool{Suspension: false}}},
					XGBoost:    {Status: extensions.Status{Enabled: true, Capabilities: map[extensions.CapabilityType]bool{Suspension: true}}},
				}

				mockClient := &mockExtensionsClient{
					listFunc: func(ctx context.Context, extensions []extensions.Type) (extensions.Group, error) {
						return mockCurrent, nil
					},
				}

				poll := &StatusChangedPoll{
					jobExtensionsStatus: initialStatus,
					extensionsClient:    mockClient,
				}

				result, err := poll.Poll(context.Background())
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeFalse())
			})
		})
	})
})
