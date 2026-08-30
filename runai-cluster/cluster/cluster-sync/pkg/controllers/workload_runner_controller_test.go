package controllers

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestControllers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controllers Suite")
}

var _ = Describe("wlrSyncFieldsChanged", func() {
	It("returns true when networking plugin details change", func() {
		oldWlr := wlrWithPluginDetails(networkingPluginName, []byte(`{"schemaVersion":"v1","endpoints":[]}`))
		newWlr := wlrWithPluginDetails(networkingPluginName, []byte(`{"schemaVersion":"v1","endpoints":[{"phase":"Ready"}]}`))

		Expect(wlrSyncFieldsChanged(oldWlr, newWlr)).To(BeTrue())
	})

	It("returns false when non-networking plugin details change", func() {
		oldWlr := wlrWithPluginDetails("storage-plugin", []byte(`{"phase":"Pending"}`))
		newWlr := wlrWithPluginDetails("storage-plugin", []byte(`{"phase":"Ready"}`))

		Expect(wlrSyncFieldsChanged(oldWlr, newWlr)).To(BeFalse())
	})

	It("returns true when the workload template changes", func() {
		oldWlr := wlrWithTemplate([]byte(`{"spec":{"replicas":1}}`))
		newWlr := wlrWithTemplate([]byte(`{"spec":{"replicas":2}}`))

		Expect(wlrSyncFieldsChanged(oldWlr, newWlr)).To(BeTrue())
	})

	It("returns false when the workload template is unchanged", func() {
		oldWlr := wlrWithTemplate([]byte(`{"spec":{"replicas":1}}`))
		newWlr := wlrWithTemplate([]byte(`{"spec":{"replicas":1}}`))

		Expect(wlrSyncFieldsChanged(oldWlr, newWlr)).To(BeFalse())
	})
})

func wlrWithPluginDetails(pluginName string, details []byte) *dromosv1alpha1.WorkloadRunner {
	return &dromosv1alpha1.WorkloadRunner{
		Status: dromosv1alpha1.WorkloadRunnerStatus{
			Plugins: map[string]*dromosv1alpha1.PluginStatus{
				pluginName: {
					AdditionalInfo: &runtime.RawExtension{Raw: details},
				},
			},
		},
	}
}

func wlrWithTemplate(template []byte) *dromosv1alpha1.WorkloadRunner {
	return &dromosv1alpha1.WorkloadRunner{
		Spec: dromosv1alpha1.WorkloadRunnerSpec{
			Template: runtime.RawExtension{Raw: template},
		},
	}
}
