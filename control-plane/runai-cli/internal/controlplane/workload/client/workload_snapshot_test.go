package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Typed workload response recovery", func() {
	It("accepts the empty desired phase returned for V2 compatibility snapshots", func() {
		server := typedInferenceServer("")
		defer server.Close()

		workloadClient, err := NewWorkloadServiceClient(server.URL, "test-token")
		Expect(err).NotTo(HaveOccurred())

		inference, err := workloadClient.GetInference(context.Background(), "test-workload")

		Expect(err).NotTo(HaveOccurred())
		Expect(inference).NotTo(BeNil())
		Expect(inference.Spec).NotTo(BeNil())
		Expect(inference.Spec.GetImage()).To(Equal("snapshot-image:latest"))
		Expect(inference.Spec.GetWorkingDir()).To(Equal("/workspace"))
	})

	It("preserves generated-client errors for unexpected desired phase values", func() {
		server := typedInferenceServer("unexpected")
		defer server.Close()

		workloadClient, err := NewWorkloadServiceClient(server.URL, "test-token")
		Expect(err).NotTo(HaveOccurred())

		inference, err := workloadClient.GetInference(context.Background(), "test-workload")

		Expect(err).To(HaveOccurred())
		Expect(inference).To(BeNil())
	})
})

func typedInferenceServer(desiredPhase string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, `{
			"name": "test-inference",
			"requestedName": "test-inference",
			"workloadId": "1ca2089d-7c38-4295-bd2c-702d33cb434c",
			"projectId": "4500009",
			"clusterId": "4cc1a01d-64c9-49f4-ad1a-b8959a4e46ea",
			"createdBy": "test@run.ai",
			"createdAt": "2026-08-17T10:13:48Z",
			"desiredPhase": %q,
			"spec": {
				"image": "snapshot-image:latest",
				"workingDir": "/workspace"
			}
		}`, desiredPhase)
	}))
}
