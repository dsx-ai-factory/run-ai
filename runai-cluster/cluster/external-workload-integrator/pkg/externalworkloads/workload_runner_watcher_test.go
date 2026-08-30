package externalworkloads

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	commonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/util/workqueue"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func wlr(phase dromosv1alpha1.WorkloadRunnerPhase, groups ...string) *dromosv1alpha1.WorkloadRunner {
	issues := make([]dromosv1alpha1.Issue, 0, len(groups))
	for _, g := range groups {
		issues = append(issues, dromosv1alpha1.Issue{Group: g, Description: "x"})
	}
	return &dromosv1alpha1.WorkloadRunner{Status: dromosv1alpha1.WorkloadRunnerStatus{Phase: phase, Issues: issues}}
}

func wlrWithPluginPhase(phase dromosv1alpha1.PluginPhase) *dromosv1alpha1.WorkloadRunner {
	return &dromosv1alpha1.WorkloadRunner{Status: dromosv1alpha1.WorkloadRunnerStatus{
		Plugins: map[string]*dromosv1alpha1.PluginStatus{
			"mnnvl": {Phase: phase},
		},
	}}
}

func ewForRunner(name, runnerName, originalWorkloadName, logicalType string) *v1alpha1.ExternalWorkload {
	return &v1alpha1.ExternalWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "team-a",
			Labels:    map[string]string{commonconstants.WorkloadRunnerNameLabel: runnerName},
		},
		// Group/Version are real (batch/v1); Type is the logical kind for a converted v2 workload.
		Spec: v1alpha1.ExternalWorkloadSpec{
			OriginalWorkloadName: originalWorkloadName, Group: "batch", Version: "v1", Type: logicalType,
		},
	}
}

var _ = Describe("wlrStatusRelevantChanged", func() {
	DescribeTable("reports whether the wlr phase or issue groups changed",
		func(old, updated *dromosv1alpha1.WorkloadRunner, want bool) {
			Expect(wlrStatusRelevantChanged(old, updated)).To(Equal(want))
		},
		Entry("identical -> no change",
			wlr(dromosv1alpha1.WorkloadRunnerPhaseRunning, "Compute"),
			wlr(dromosv1alpha1.WorkloadRunnerPhaseRunning, "Compute"),
			false),
		Entry("phase changed -> change",
			wlr(dromosv1alpha1.WorkloadRunnerPhaseRunning),
			wlr(dromosv1alpha1.WorkloadRunnerPhaseDegraded),
			true),
		Entry("issue group added -> change",
			wlr(dromosv1alpha1.WorkloadRunnerPhaseDegraded, "Compute"),
			wlr(dromosv1alpha1.WorkloadRunnerPhaseDegraded, "Compute", "Network"),
			true),
		Entry("same issues different order -> no change (deterministic signature)",
			wlr(dromosv1alpha1.WorkloadRunnerPhaseFailed, "Network", "Compute", "Storage"),
			wlr(dromosv1alpha1.WorkloadRunnerPhaseFailed, "Storage", "Compute", "Network"),
			false),
		Entry("issue description changed -> change",
			&dromosv1alpha1.WorkloadRunner{Status: dromosv1alpha1.WorkloadRunnerStatus{Issues: []dromosv1alpha1.Issue{{Group: "Network", Description: "old"}}}},
			&dromosv1alpha1.WorkloadRunner{Status: dromosv1alpha1.WorkloadRunnerStatus{Issues: []dromosv1alpha1.Issue{{Group: "Network", Description: "new"}}}},
			true),
		Entry("MNNVL plugin phase changed -> change",
			wlrWithPluginPhase(dromosv1alpha1.PluginPhaseCreating),
			wlrWithPluginPhase(dromosv1alpha1.PluginPhaseReady),
			true),
	)
})

var _ = Describe("enqueueWorkloadsForWorkloadRunner", func() {
	It("enqueues each supported V2 workload using its concrete resource handler", func() {
		const ns, wlrName = "team-a", "wlr-1"

		scheme := runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())

		c := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(
			ewForRunner("ew-training", wlrName, "training", "Training"),
			ewForRunner("ew-workspace", wlrName, "workspace", "Workspace"),
			ewForRunner("ew-inference", wlrName, "inference", "Inference"),
			ewForRunner("ew-dinf", wlrName, "dinf", "DistributedInference"),
			&v1alpha1.ExternalWorkload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ew-dist",
					Namespace: ns,
					Labels:    map[string]string{commonconstants.WorkloadRunnerNameLabel: wlrName},
					OwnerReferences: []metav1.OwnerReference{
						{APIVersion: "batch/v1", Kind: "Job", Name: "dist-worker"},
						{APIVersion: "kubeflow.org/v1", Kind: "PyTorchJob", Name: "dist"},
					},
				},
				Spec: v1alpha1.ExternalWorkloadSpec{
					OriginalWorkloadName: "dist",
					Type:                 "Distributed",
					Group:                kubeflowGroup,
					Version:              "v1",
				},
			},
			ewForRunner("ew-empty", wlrName, "", "Workspace"),
			ewForRunner("ew-other", "wlr-2", "other-job", "Training"),
			ewForRunner("ew-plain", wlrName, "plain-job", "Job"),
		).Build()

		q := workqueue.NewTypedRateLimitingQueue[utils.ExternalWorkloadTask](
			workqueue.DefaultTypedControllerRateLimiter[utils.ExternalWorkloadTask]())
		handler := &ExternalWorkloadHandler{tasksQueue: q}

		enqueueWorkloadsForWorkloadRunner(context.Background(), c, handler,
			&dromosv1alpha1.WorkloadRunner{ObjectMeta: metav1.ObjectMeta{Name: wlrName, Namespace: ns}})

		Expect(q.Len()).To(Equal(5))

		queued := map[string]schema.GroupVersionKind{}
		for q.Len() > 0 {
			task, _ := q.Get()
			queued[task.Key.Name] = task.GVK
			Expect(task.Key.Namespace).To(Equal(ns))
			q.Done(task)
		}
		Expect(queued).To(Equal(map[string]schema.GroupVersionKind{
			"training":  batchv1.SchemeGroupVersion.WithKind("Job"),
			"workspace": batchv1.SchemeGroupVersion.WithKind("Job"),
			"inference": {Group: "serving.knative.dev", Version: "v1", Kind: "Service"},
			"dinf":      {Group: "leaderworkerset.x-k8s.io", Version: "v1", Kind: "LeaderWorkerSet"},
			"dist":      {Group: kubeflowGroup, Version: "v1", Kind: "PyTorchJob"},
		}))
	})
})
