package unschedulable

import (
	"testing"

	"github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	"github.com/run-ai/runai/runai-cluster/cluster/common/resourceutils/resources"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	queueRequest                  = "queueRequest"
	queueDeserved                 = "queueDeserved"
	queueAllocated                = "queueAllocated"
	queueAllocatedNonPreemptible  = "queueAllocatedNonPreemptible"
	queueLimits                   = "queueLimits"
	podgroupRequest               = "podgroupRequest"
	podgroupRequestNonPreemptible = "podgroupRequestNonPreemptible"
)

func TestParseExceedingResource(t *testing.T) {
	type test struct {
		name              string
		explanation       v2alpha2.UnschedulableExplanation
		expectedExceeding v1.ResourceList
		expectedError     bool
	}

	tests := []test{
		{
			name: "NonPreemptibleOverQuota",
			explanation: v2alpha2.UnschedulableExplanation{
				Reason: v2alpha2.NonPreemptibleOverQuota,
				Details: &v2alpha2.UnschedulableExplanationDetails{
					QueueDetails: getPopulatedQuotaDetails(),
				},
			},
			expectedExceeding: v1.ResourceList{
				queueAllocatedNonPreemptible:  resource.MustParse("1"),
				podgroupRequestNonPreemptible: resource.MustParse("1"),
			},
			expectedError: false,
		},
		{
			name: "OverLimit",
			explanation: v2alpha2.UnschedulableExplanation{
				Reason: v2alpha2.OverLimit,
				Details: &v2alpha2.UnschedulableExplanationDetails{
					QueueDetails: getPopulatedQuotaDetails(),
				},
			},
			expectedExceeding: v1.ResourceList{
				queueAllocated:  resource.MustParse("1"),
				podgroupRequest: resource.MustParse("1"),
			},
			expectedError: false,
		},
		{
			name: "UnsupportedReason",
			explanation: v2alpha2.UnschedulableExplanation{
				Reason: "UnsupportedReason",
				Details: &v2alpha2.UnschedulableExplanationDetails{
					QueueDetails: getPopulatedQuotaDetails(),
				},
			},
			expectedError: true,
		},
	}

	for i, entry := range tests {
		t.Run(entry.name, func(t *testing.T) {
			exceeding, err := ParseExceedingResources(entry.explanation)

			if entry.expectedError {
				if err == nil {
					t.Errorf("Test %d: Expected error, got nil", i)
				}
			} else {
				if err != nil {
					t.Errorf("Test %d: Expected no error, got %v", i, err)
				}
			}

			if err := resources.EqualResourceLists(exceeding, entry.expectedExceeding); err != nil {
				t.Errorf("Test %d: Expected exceeding:\n%v\ngot:\n%v", i, entry.expectedExceeding, exceeding)
			}
		})
	}
}

func getPopulatedQuotaDetails() *v2alpha2.QuotaDetails {
	return &v2alpha2.QuotaDetails{
		Name:                                     "queue",
		QueueRequestedResources:                  v1.ResourceList{queueRequest: resource.MustParse("1")},
		QueueDeservedResources:                   v1.ResourceList{queueDeserved: resource.MustParse("1")},
		QueueAllocatedResources:                  v1.ResourceList{queueAllocated: resource.MustParse("1")},
		QueueAllocatedNonPreemptibleResources:    v1.ResourceList{queueAllocatedNonPreemptible: resource.MustParse("1")},
		QueueResourceLimits:                      v1.ResourceList{queueLimits: resource.MustParse("1")},
		PodGroupRequestedResources:               v1.ResourceList{podgroupRequest: resource.MustParse("1")},
		PodGroupRequestedNonPreemptibleResources: v1.ResourceList{podgroupRequestNonPreemptible: resource.MustParse("1")},
	}
}
