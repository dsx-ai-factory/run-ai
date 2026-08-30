package runai_scheduler

import (
	"reflect"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/operands/runai-scheduler/resources"
)

func TestRunaiSchedulerOperand(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Runai Scheduler Operand Suite")
}

// containsFunc reports whether funcs includes target. Function values are not comparable
// with ==, so we compare their code addresses; two references to the same package-level
// function share an address.
func containsFunc[T any](funcs []T, target T) bool {
	targetPtr := reflect.ValueOf(target).Pointer()
	for _, f := range funcs {
		if reflect.ValueOf(f).Pointer() == targetPtr {
			return true
		}
	}
	return false
}

var _ = Describe("operand ServiceMonitor gating", func() {
	enabled := &operand{name: "np", includeServiceMonitor: true}
	disabled := &operand{name: "np", includeServiceMonitor: false}

	Context("when the ServiceMonitor CRD is present (enabled)", func() {
		It("builds both the SchedulingShard and the ServiceMonitor", func() {
			funcs := enabled.resourceFunctions()
			Expect(containsFunc(funcs, resources.SchedulingShardForNodePool)).To(BeTrue(), "SchedulingShard should always be built")
			Expect(containsFunc(funcs, resources.ServiceMonitorForNodePool)).To(BeTrue(), "ServiceMonitor should be built when enabled")
		})

		It("reports status for both the SchedulingShard and the ServiceMonitor", func() {
			funcs := enabled.resourceStatusFunctions()
			Expect(containsFunc(funcs, resources.SchedulingShardStatus)).To(BeTrue(), "SchedulingShard status should always be reported")
			Expect(containsFunc(funcs, resources.ServiceMonitorStatus)).To(BeTrue(), "ServiceMonitor status should be reported when enabled")
		})
	})

	Context("when the ServiceMonitor CRD is absent (disabled)", func() {
		It("still builds the SchedulingShard but omits the ServiceMonitor", func() {
			funcs := disabled.resourceFunctions()
			Expect(containsFunc(funcs, resources.SchedulingShardForNodePool)).To(BeTrue(), "SchedulingShard is unaffected by the ServiceMonitor gate")
			Expect(containsFunc(funcs, resources.ServiceMonitorForNodePool)).To(BeFalse(), "ServiceMonitor must not be built without its CRD")
		})

		It("still reports SchedulingShard status but omits the ServiceMonitor status", func() {
			funcs := disabled.resourceStatusFunctions()
			Expect(containsFunc(funcs, resources.SchedulingShardStatus)).To(BeTrue(), "SchedulingShard status is unaffected by the ServiceMonitor gate")
			Expect(containsFunc(funcs, resources.ServiceMonitorStatus)).To(BeFalse(), "ServiceMonitor status must not be referenced without its CRD")
		})
	})
})
