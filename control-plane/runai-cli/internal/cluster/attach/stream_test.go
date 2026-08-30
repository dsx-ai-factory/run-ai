package attach

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/pseudo"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/network"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/autogenerate/mocks"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/apis"
	clusterErrors "github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/errors"
)

var _ = Describe("WorkloadAttachStreamer", func() {
	var (
		ctrl             *gomock.Controller
		mockDialer       *mocks.MockDialer
		workloadStreamer *WorkloadAttach
		ctx              context.Context
		workload         *apis.Workload
		options          *pseudo.Options
		maxRetryDuration time.Duration
		timeBetweenTries time.Duration

		expectedConn     network.Conn
		expectedResponse *http.Response
		errExpected      error
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockDialer = mocks.NewMockDialer(ctrl)
		maxRetryDuration = 500 * time.Millisecond
		timeBetweenTries = 100 * time.Millisecond
		workloadStreamer = NewWorkloadAttach(mockDialer, "https://example.com", maxRetryDuration, timeBetweenTries, nil)
		ctx = context.Background()
		workload = &apis.Workload{
			Project: "example-project",
			Type:    api.Training,
			Name:    "example-workload",
		}
		options, _ = pseudo.NewOptions("example-pod", "example-container", true, true)
		expectedConn = mocks.NewMockConn(ctrl)
		expectedResponse = &http.Response{}
		errExpected = nil
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Stream", func() {
		It("returns the attach when streaming is successful", func() {
			mockDialer.EXPECT().Connect(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedConn, expectedResponse, errExpected).AnyTimes()

			reader, err := workloadStreamer.Stream(ctx, workload, options)

			Expect(err).NotTo(HaveOccurred())
			Expect(reader).To(Equal(expectedConn))
		})

		It("returns an error when building the URL fails", func() {
			errExpected = clusterErrors.NewNotReadyError("")
			invalidOptions := &pseudo.Options{
				Pod:       "podder",
				Container: "container",
				Stdin:     true,
				TTY:       true,
			}

			mockDialer.EXPECT().Connect(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil, errExpected).AnyTimes()

			reader, err := workloadStreamer.Stream(ctx, workload, invalidOptions)

			Expect(err).To(HaveOccurred())
			Expect(reader).To(BeNil())
		})

		It("returns an error when streaming fails with a non-retryable error", func() {
			errExpected = errors.New("non-retryable error")
			mockDialer.EXPECT().Connect(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil, errExpected)

			reader, err := workloadStreamer.Stream(ctx, workload, options)

			Expect(err.Error()).To(ContainSubstring(errExpected.Error()))
			Expect(reader).To(BeNil())
		})

		It("retries and returns an error when streaming fails with a retryable error and exceeds the maximum retry duration", func() {
			errExpected = clusterErrors.NewNotReadyError("")
			res := &http.Response{
				StatusCode:    http.StatusPreconditionFailed,
				Body:          nil,
				ContentLength: 1,
			}
			mockDialer.EXPECT().Connect(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, res, errExpected).AnyTimes()

			startTime := time.Now()
			reader, err := workloadStreamer.Stream(ctx, workload, options)
			duration := time.Since(startTime)

			Expect(err.Error()).To(ContainSubstring(errExpected.Error()))
			Expect(reader).To(BeNil())
			Expect(duration).To(BeNumerically(">=", maxRetryDuration))
		})
	})
})
