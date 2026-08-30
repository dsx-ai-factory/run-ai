package nim_autoscale_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli/clients"
)

var _ = Describe("NIM Autoscale Describe Test", Ordered, func() {
	var (
		mockCtrl       *gomock.Controller
		describeMocker *clients.DescribeMocker
		ctx            context.Context
	)

	const (
		podLimit   int32 = 10
		eventLimit int32 = 50
	)

	BeforeAll(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		ctx = context.Background()

		var err error
		describeMocker, err = clients.NewDescribeMocker(mockCtrl, clients.NimAutoscaleFixture)
		Expect(err).NotTo(HaveOccurred())
		Expect(describeMocker).NotTo(BeNil())
	})

	AfterAll(func() {
		mockCtrl.Finish()
	})

	Context("Describe NIM Autoscale Workload - Table Output", func() {
		It("should match expected table output", func() {
			fixtures := describeMocker.GetFixtures()
			Expect(fixtures.WorkloadDetailed).NotTo(BeNil())

			reqWorkload := &api.Workload{
				ID: fixtures.WorkloadDetailed.Id,
			}

			service := describeMocker.CreateDescribeService()
			result, err := service.GetWorkload(ctx, reqWorkload, podLimit, eventLimit)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())

			actualOutput, err := clients.CaptureDescribeTableOutput(result)
			Expect(err).NotTo(HaveOccurred())

			err = clients.CompareWithExpectedOutput(clients.NimAutoscaleFixture, "table", actualOutput)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("Describe NIM Autoscale Workload - JSON Output", func() {
		It("should match expected JSON output", func() {
			fixtures := describeMocker.GetFixtures()
			Expect(fixtures.WorkloadDetailed).NotTo(BeNil())

			reqWorkload := &api.Workload{
				ID: fixtures.WorkloadDetailed.Id,
			}

			service := describeMocker.CreateDescribeService()
			result, err := service.GetWorkload(ctx, reqWorkload, podLimit, eventLimit)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())

			actualOutput, err := clients.CaptureDescribeJSONOutput(result)
			Expect(err).NotTo(HaveOccurred())

			err = clients.CompareWithExpectedOutput(clients.NimAutoscaleFixture, "json", actualOutput)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
