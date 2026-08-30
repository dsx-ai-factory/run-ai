package handlers

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	runv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("KartaController (Karta-centric)", Ordered, func() {
	Context("status condition helpers", func() {
		It("writes only the EWI-owned conditions (RBACReady, SyncEnabled, RunaiReady)", func() {
			status := kartav1alpha1.KartaStatus{}
			SetConditions(&status, KartaConditionStatuses{
				RBACReady:   metav1.ConditionFalse,
				SyncEnabled: metav1.ConditionFalse,
			})
			Expect(status.Conditions).To(HaveLen(3))
			Expect(findCond(status.Conditions, runv1alpha1.RBACReady).Status).To(Equal(metav1.ConditionFalse))
			Expect(findCond(status.Conditions, runv1alpha1.SyncEnabled).Status).To(Equal(metav1.ConditionFalse))
			Expect(findCond(status.Conditions, runv1alpha1.RunaiReady).Status).To(Equal(metav1.ConditionFalse))

			// EWI must not fabricate the OSS-owned conditions.
			_, hasReady := tryFindCond(status.Conditions, string(runv1alpha1.Ready))
			_, hasValidated := tryFindCond(status.Conditions, string(runv1alpha1.KartaValidated))
			_, hasCRD := tryFindCond(status.Conditions, string(runv1alpha1.CRDExists))
			Expect(hasReady).To(BeFalse())
			Expect(hasValidated).To(BeFalse())
			Expect(hasCRD).To(BeFalse())
		})

		It("preserves OSS-owned conditions while writing EWI conditions", func() {
			status := kartav1alpha1.KartaStatus{}
			// OSS controller has reconciled: Karta is valid, CRD exists, Ready.
			seedOSSCondition(&status, runv1alpha1.KartaValidated, metav1.ConditionTrue)
			seedOSSCondition(&status, runv1alpha1.CRDExists, metav1.ConditionTrue)
			seedOSSCondition(&status, runv1alpha1.Ready, metav1.ConditionTrue)

			SetConditions(&status, KartaConditionStatuses{
				RBACReady:   metav1.ConditionTrue,
				SyncEnabled: metav1.ConditionTrue,
			})

			// 3 OSS-owned + 3 EWI-owned.
			Expect(status.Conditions).To(HaveLen(6))
			Expect(findCond(status.Conditions, runv1alpha1.Ready).Status).To(Equal(metav1.ConditionTrue))
			Expect(findCond(status.Conditions, runv1alpha1.KartaValidated).Status).To(Equal(metav1.ConditionTrue))
			Expect(findCond(status.Conditions, runv1alpha1.CRDExists).Status).To(Equal(metav1.ConditionTrue))
			// RunaiReady = OSS Ready AND RBACReady.
			Expect(findCond(status.Conditions, runv1alpha1.RunaiReady).Status).To(Equal(metav1.ConditionTrue))
		})

		It("sets RunaiReady=True only when both the OSS Ready and RBACReady are True", func() {
			status := kartav1alpha1.KartaStatus{}
			seedOSSCondition(&status, runv1alpha1.Ready, metav1.ConditionTrue)
			SetConditions(&status, KartaConditionStatuses{
				RBACReady:   metav1.ConditionTrue,
				SyncEnabled: metav1.ConditionTrue,
			})

			runaiReady := findCond(status.Conditions, runv1alpha1.RunaiReady)
			Expect(runaiReady.Status).To(Equal(metav1.ConditionTrue))
			Expect(runaiReady.Reason).To(BeEquivalentTo(runv1alpha1.RunaiReady))
			Expect(runaiReady.Message).To(BeEmpty())
		})

		It("sets RunaiReady=False with a Ready-specific reason when the OSS Ready is not met", func() {
			status := kartav1alpha1.KartaStatus{}
			// OSS Ready is False (or, as here, absent) even though RBACReady is True.
			SetConditions(&status, KartaConditionStatuses{
				RBACReady:   metav1.ConditionTrue,
				SyncEnabled: metav1.ConditionTrue,
			})

			runaiReady := findCond(status.Conditions, runv1alpha1.RunaiReady)
			Expect(runaiReady.Status).To(Equal(metav1.ConditionFalse))
			Expect(runaiReady.Reason).To(Equal("ReadyNotMet"))
			Expect(runaiReady.Message).To(ContainSubstring("Ready"))
		})

		It("sets RunaiReady=False with an RBAC-specific reason when only RBACReady is not met", func() {
			status := kartav1alpha1.KartaStatus{}
			// OSS Ready True but RBACReady False.
			seedOSSCondition(&status, runv1alpha1.Ready, metav1.ConditionTrue)
			SetConditions(&status, KartaConditionStatuses{
				RBACReady:   metav1.ConditionFalse,
				SyncEnabled: metav1.ConditionTrue,
			})

			runaiReady := findCond(status.Conditions, runv1alpha1.RunaiReady)
			Expect(runaiReady.Status).To(Equal(metav1.ConditionFalse))
			Expect(runaiReady.Reason).To(Equal("RBACNotReady"))
			Expect(runaiReady.Message).To(ContainSubstring("RBACReady"))
		})

		It("updates RunaiReady whenever the OSS Ready or RBACReady changes", func() {
			status := kartav1alpha1.KartaStatus{}
			seedOSSCondition(&status, runv1alpha1.Ready, metav1.ConditionTrue)

			// Start not runai-ready (RBACReady False).
			SetConditions(&status, KartaConditionStatuses{
				RBACReady:   metav1.ConditionFalse,
				SyncEnabled: metav1.ConditionTrue,
			})
			Expect(findCond(status.Conditions, runv1alpha1.RunaiReady).Status).To(Equal(metav1.ConditionFalse))

			// RBACReady flips True -> RunaiReady should flip True.
			afterRbacFalse := status.DeepCopy()
			SetConditions(&status, KartaConditionStatuses{
				RBACReady:   metav1.ConditionTrue,
				SyncEnabled: metav1.ConditionTrue,
			})
			Expect(findCond(status.Conditions, runv1alpha1.RunaiReady).Status).To(Equal(metav1.ConditionTrue))
			Expect(WasStatusUpdated(afterRbacFalse, &status)).To(BeTrue())

			// OSS Ready flips False -> RunaiReady should flip False.
			afterRunaiReady := status.DeepCopy()
			setConditionStatus(&status, runv1alpha1.Ready, metav1.ConditionFalse)
			SetConditions(&status, KartaConditionStatuses{
				RBACReady:   metav1.ConditionTrue,
				SyncEnabled: metav1.ConditionTrue,
			})
			Expect(findCond(status.Conditions, runv1alpha1.RunaiReady).Status).To(Equal(metav1.ConditionFalse))
			Expect(WasStatusUpdated(afterRunaiReady, &status)).To(BeTrue())
		})

		It("clears RunaiReady and sets deletion conditions", func() {
			status := kartav1alpha1.KartaStatus{}
			// Karta was runai-ready before deletion started.
			seedOSSCondition(&status, runv1alpha1.Ready, metav1.ConditionTrue)
			SetConditions(&status, KartaConditionStatuses{
				RBACReady:   metav1.ConditionTrue,
				SyncEnabled: metav1.ConditionTrue,
			})
			Expect(findCond(status.Conditions, runv1alpha1.RunaiReady).Status).To(Equal(metav1.ConditionTrue))

			// On error case -> all False
			setDeletionConditions(&status, KartaDeletionConditionStatuses{
				NoWorkloadsUseKarta: metav1.ConditionFalse,
				RBACCleanupSuccess:  metav1.ConditionFalse,
				SyncDisabled:        metav1.ConditionFalse,
			})
			Expect(findCond(status.Conditions, runv1alpha1.NoWorkloadsUseKarta).Status).To(Equal(metav1.ConditionFalse))
			Expect(findCond(status.Conditions, runv1alpha1.NoWorkloadsUseKarta).Reason).To(BeEquivalentTo(runv1alpha1.NoWorkloadsUseKarta))
			Expect(findCond(status.Conditions, runv1alpha1.NoWorkloadsUseKarta).Message).ToNot(BeEmpty())
			Expect(findCond(status.Conditions, runv1alpha1.RBACCleanupSuccess).Message).ToNot(BeEmpty())
			Expect(findCond(status.Conditions, runv1alpha1.SyncDisabled).Message).ToNot(BeEmpty())
			// RunaiReady is explicitly cleared during deletion.
			Expect(findCond(status.Conditions, runv1alpha1.RunaiReady).Status).To(Equal(metav1.ConditionFalse))

			old := status.DeepCopy()
			// Success case -> all True, messages should be empty
			setDeletionConditions(&status, KartaDeletionConditionStatuses{
				NoWorkloadsUseKarta: metav1.ConditionTrue,
				RBACCleanupSuccess:  metav1.ConditionTrue,
				SyncDisabled:        metav1.ConditionTrue,
			})
			Expect(findCond(status.Conditions, runv1alpha1.NoWorkloadsUseKarta).Status).To(Equal(metav1.ConditionTrue))
			Expect(findCond(status.Conditions, runv1alpha1.NoWorkloadsUseKarta).Message).To(BeEmpty())
			Expect(findCond(status.Conditions, runv1alpha1.RBACCleanupSuccess).Message).To(BeEmpty())
			Expect(findCond(status.Conditions, runv1alpha1.SyncDisabled).Message).To(BeEmpty())
			Expect(WasStatusUpdated(old, &status)).To(BeTrue())
		})

		It("upsert preserves LastTransitionTime when status unchanged, updates when changed", func() {
			status := kartav1alpha1.KartaStatus{}
			SetConditions(&status, KartaConditionStatuses{
				RBACReady:   metav1.ConditionFalse,
				SyncEnabled: metav1.ConditionFalse,
			})
			beforeSync := findCond(status.Conditions, runv1alpha1.SyncEnabled).LastTransitionTime
			beforeRbac := findCond(status.Conditions, runv1alpha1.RBACReady).LastTransitionTime

			// Change RBACReady to True, leave SyncEnabled unchanged.
			SetConditions(&status, KartaConditionStatuses{
				RBACReady:   metav1.ConditionTrue,
				SyncEnabled: metav1.ConditionFalse,
			})
			afterSync := findCond(status.Conditions, runv1alpha1.SyncEnabled).LastTransitionTime
			afterRbac := findCond(status.Conditions, runv1alpha1.RBACReady).LastTransitionTime

			// Unchanged status keeps same transition time
			Expect(afterSync).To(Equal(beforeSync))
			// Changed status gets a different transition time
			Expect(afterRbac.Time.Equal(beforeRbac.Time)).To(BeFalse())
		})

		It("upsert preserves conditions not present in the new map", func() {
			status := kartav1alpha1.KartaStatus{}
			// Seed OSS-owned conditions plus an EWI condition.
			seedOSSCondition(&status, runv1alpha1.Ready, metav1.ConditionTrue)
			seedOSSCondition(&status, runv1alpha1.CRDExists, metav1.ConditionTrue)
			SetConditions(&status, KartaConditionStatuses{
				RBACReady:   metav1.ConditionTrue,
				SyncEnabled: metav1.ConditionTrue,
			})

			// Upsert with only RBACReady - all other conditions must be preserved.
			before := len(status.Conditions)
			newMap := map[runv1alpha1.KartaConditionType]metav1.Condition{
				runv1alpha1.RBACReady: buildCondition(runv1alpha1.RBACReady, metav1.ConditionFalse, "changed"),
			}
			upsertKartaConditions(&status.Conditions, newMap)
			Expect(status.Conditions).To(HaveLen(before))
			Expect(findCond(status.Conditions, runv1alpha1.RBACReady).Status).To(Equal(metav1.ConditionFalse))
			Expect(findCond(status.Conditions, runv1alpha1.Ready).Status).To(Equal(metav1.ConditionTrue))
			Expect(findCond(status.Conditions, runv1alpha1.CRDExists).Status).To(Equal(metav1.ConditionTrue))
		})

		It("buildCondition sets reason to type and message only on non-True", func() {
			trueCond := buildCondition("SomeCond", metav1.ConditionTrue, "should-be-empty")
			falseCond := buildCondition("SomeCond", metav1.ConditionFalse, "error message")
			Expect(trueCond.Reason).To(Equal("SomeCond"))
			Expect(trueCond.Message).To(BeEmpty())
			Expect(falseCond.Reason).To(Equal("SomeCond"))
			Expect(falseCond.Message).To(Equal("error message"))
		})
	})
})

// seedOSSCondition appends a condition of the kind the OSS Karta controller owns
// (KartaValidated, CRDExists, Ready) so we can exercise how EWI reads/preserves them.
func seedOSSCondition(status *kartav1alpha1.KartaStatus, condType runv1alpha1.KartaConditionType, s metav1.ConditionStatus) {
	status.Conditions = append(status.Conditions, metav1.Condition{
		Type:   string(condType),
		Status: s,
		Reason: string(condType),
	})
}

// setConditionStatus flips the status of an existing condition in place.
func setConditionStatus(status *kartav1alpha1.KartaStatus, condType runv1alpha1.KartaConditionType, s metav1.ConditionStatus) {
	for i := range status.Conditions {
		if status.Conditions[i].Type == string(condType) {
			status.Conditions[i].Status = s
			return
		}
	}
}

func tryFindCond(conds []metav1.Condition, t string) (metav1.Condition, bool) {
	for _, c := range conds {
		if c.Type == t {
			return c, true
		}
	}
	return metav1.Condition{}, false
}

func findCond(conds []metav1.Condition, t runv1alpha1.KartaConditionType) metav1.Condition {
	c, ok := tryFindCond(conds, string(t))
	if !ok {
		Fail("condition not found: " + string(t))
	}
	return c
}
