package parsers

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	issuepkg "github.com/run-ai/runai/runai-cluster/cluster/workload-common/issue"
	common "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	commonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("appendWorkloadIssueConditions", func() {
	It("adds nothing when there are no WorkloadIssue conditions", func() {
		issues := &common.Issues{}
		ew := ewWithConditions(v1alpha1.ObservedCondition{Type: "Failed", Message: "x"})
		appendWorkloadIssueConditions(issues, ew)
		Expect(issues).To(Equal(&common.Issues{}))
	})

	It("adds a Compute issue from a WorkloadIssue condition", func() {
		issues := &common.Issues{}
		ew := ewWithConditions(workloadIssueCond(issuepkg.Compute.String(), "container 'main' was OOMKilled"))
		appendWorkloadIssueConditions(issues, ew)
		Expect(issues.Compute).To(HaveLen(1))
		Expect(issues.Compute[0].Description).To(Equal("container 'main' was OOMKilled"))
		Expect(issues.Compute[0].Name).To(Equal("my-workload"))
	})

	It("appends to existing issues without dropping them", func() {
		issues := &common.Issues{Network: []common.Issue{{Kind: "Service", Name: "svc", Description: "net"}}}
		ew := ewWithConditions(workloadIssueCond(issuepkg.Compute.String(), "boom"))
		appendWorkloadIssueConditions(issues, ew)
		Expect(issues.Network).To(HaveLen(1))
		Expect(issues.Compute).To(HaveLen(1))
	})

	It("maps an unknown reason to the Internal group", func() {
		issues := &common.Issues{}
		ew := ewWithConditions(workloadIssueCond("Weird", "huh"))
		appendWorkloadIssueConditions(issues, ew)
		Expect(issues.Internal).To(HaveLen(1))
	})
})

var _ = Describe("parseExternalWorkloadConditions filtering", func() {
	It("drops WorkloadIssue conditions but keeps native ones", func() {
		conds := []v1alpha1.ObservedCondition{
			{Type: "Failed", Status: corev1.ConditionTrue, Message: "native"},
			workloadIssueCond(issuepkg.Compute.String(), "issue detail"),
		}
		got := parseExternalWorkloadConditions(conds)
		Expect(got).To(HaveLen(1))
		Expect(got[0].Type).To(Equal("Failed"))
	})
})

func ewWithConditions(conds ...v1alpha1.ObservedCondition) *v1alpha1.ExternalWorkload {
	return &v1alpha1.ExternalWorkload{
		Spec: v1alpha1.ExternalWorkloadSpec{OriginalWorkloadName: "my-workload"},
		Status: v1alpha1.ExternalWorkloadStatus{
			ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{Conditions: conds},
		},
	}
}

func workloadIssueCond(reason, message string) v1alpha1.ObservedCondition {
	return v1alpha1.ObservedCondition{
		Type:    commonconstants.WorkloadIssueConditionType,
		Status:  corev1.ConditionTrue,
		Reason:  reason,
		Message: message,
	}
}
