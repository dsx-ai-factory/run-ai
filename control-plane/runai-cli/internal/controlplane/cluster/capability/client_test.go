package capability

import (
	"errors"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/cluster/list/manipulate"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Checker Capabilities", func() {
	var (
		txn         *transaction.Txn
		cpURL       string
		token       string
		clusterUUID string
		config      *Checker

		mockCluster *manipulate.BriefCluster
		mockErr     error
	)

	BeforeEach(func() {
		txn = &transaction.Txn{}
		cpURL = "https://fake-cp-url"
		token = "fake-token"
		clusterUUID = "fake-cluster-id"

		config = &Checker{
			ControlPlaneURL: cpURL,
			Token:           token,
			ClusterUUID:     clusterUUID,
			getClusterByUuidFunc: func(t *transaction.Txn, url, tk, uuid string) (*manipulate.BriefCluster, error) {
				Expect(t).To(Equal(txn))
				Expect(url).To(Equal(cpURL))
				Expect(tk).To(Equal(token))
				Expect(uuid).To(Equal(clusterUUID))
				return mockCluster, mockErr
			},
		}
	})

	Context("when workload type is not inference", func() {
		It("should allow any action", func() {
			supported, err := config.IsSupported(txn, api.Training, Logs)
			Expect(err).ToNot(HaveOccurred())
			Expect(supported).To(BeTrue())
		})
	})

	Context("when workload type is inference", func() {
		It("should allow unsupported actions", func() {
			supported, err := config.IsSupported(txn, api.Inference, Actions(99))
			Expect(err).ToNot(HaveOccurred())
			Expect(supported).To(BeTrue())
		})

		DescribeTable("should validate cluster version for Logs/Exec/PortForward",
			func(action Actions, clusterVersionOrError interface{}, expectedSupported bool, expectedErrorMatcher OmegaMatcher) {
				switch val := clusterVersionOrError.(type) {
				case error:
					mockErr = val
					mockCluster = nil
				case nil:
					mockCluster = nil
					mockErr = nil
				case string:
					mockErr = nil
					mockCluster = &manipulate.BriefCluster{Version: val}
				default:
					Fail("Invalid test input")
				}

				supported, err := config.IsSupported(txn, api.Inference, action)

				if expectedErrorMatcher != nil {
					Expect(err).To(expectedErrorMatcher)
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(supported).To(Equal(expectedSupported))
				}
			},

			Entry("Logs: supported on minimal version", Logs, "2.22.0", true, nil),
			Entry("Logs: supported on higher version", Logs, "2.30.0", true, nil),
			Entry("Logs: not supported on lower version", Logs, "2.21.9", false, nil),

			Entry("Exec: supported on higher version", Exec, "2.23.0", true, nil),
			Entry("Exec: not supported on lower version", Exec, "2.0.0", false, nil),
			Entry("PortForward: supported on higher version", PortForward, "3.1.1", true, nil),
			Entry("PortForward: not supported on lower version", PortForward, "2.0.0", false, nil),

			Entry("Invalid version format returns error", Exec, "bad.version", false, MatchError(ContainSubstring("invalid cluster version"))),
			Entry("Nil cluster returns error", Exec, nil, false, MatchError(ContainSubstring("cluster data is nil"))),
			Entry("Fetch failure returns error", PortForward, errors.New("fetch failed"), false, MatchError(ContainSubstring("fetch failed"))),
		)
	})

	Describe("Test environment versions", func() {

		It("should allow version with 'master'", func() {
			mockCluster = &manipulate.BriefCluster{Version: "0.0.0-10003314.master-4f45659"}
			mockErr = nil
			supported, err := config.IsSupported(txn, api.Inference, Logs)
			Expect(err).NotTo(HaveOccurred())
			Expect(supported).To(BeTrue())
		})
	})
})
