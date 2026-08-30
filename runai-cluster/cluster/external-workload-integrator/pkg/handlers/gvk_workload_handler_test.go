package handlers

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("GvkWorkloadHandler network status", func() {
	It("enables network status updates only for LeaderWorkerSet", func() {
		lwsHandler := &GvkWorkloadHandler{gvk: schema.GroupVersionKind{
			Group: "leaderworkerset.x-k8s.io", Version: "v1", Kind: "LeaderWorkerSet",
		}}
		otherHandler := &GvkWorkloadHandler{gvk: schema.GroupVersionKind{
			Group: "example.com", Version: "v1", Kind: "OtherWorkload",
		}}

		Expect(lwsHandler.ShouldUpdateNetworkStatus()).To(BeTrue())
		Expect(otherHandler.ShouldUpdateNetworkStatus()).To(BeFalse())
	})
})
