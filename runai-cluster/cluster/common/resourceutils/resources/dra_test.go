package resources

import (
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/resource/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"testing"
)

func TestUpsertReservedFor(t *testing.T) {
	type data struct {
		name  string
		pod   *v1.Pod
		claim v1beta1.ResourceClaim

		expected []v1beta1.ResourceClaimConsumerReference
	}

	tests := []data{
		{
			name: "empty claim",
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pod1",
					UID:  "uid1",
				},
			},
			claim: v1beta1.ResourceClaim{},

			expected: []v1beta1.ResourceClaimConsumerReference{
				{
					APIGroup: "",
					Resource: "pods",
					Name:     "pod1",
					UID:      "uid1",
				},
			},
		},
		{
			name: "claim with existing reference",
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pod1",
					UID:  "uid1",
				},
			},
			claim: v1beta1.ResourceClaim{
				Status: v1beta1.ResourceClaimStatus{
					ReservedFor: []v1beta1.ResourceClaimConsumerReference{
						{
							APIGroup: "",
							Resource: "pods",
							Name:     "pod2",
							UID:      "uid2",
						},
					},
				},
			},
			expected: []v1beta1.ResourceClaimConsumerReference{
				{
					APIGroup: "",
					Resource: "pods",
					Name:     "pod1",
					UID:      "uid1",
				},
				{
					APIGroup: "",
					Resource: "pods",
					Name:     "pod2",
					UID:      "uid2",
				},
			},
		},
		{
			name: "claim with existing identical reference",
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pod1",
					UID:  "uid1",
				},
			},
			claim: v1beta1.ResourceClaim{
				Status: v1beta1.ResourceClaimStatus{
					ReservedFor: []v1beta1.ResourceClaimConsumerReference{
						{
							APIGroup: "",
							Resource: "pods",
							Name:     "pod1",
							UID:      "uid1",
						},
					},
				},
			},
			expected: []v1beta1.ResourceClaimConsumerReference{
				{
					APIGroup: "",
					Resource: "pods",
					Name:     "pod1",
					UID:      "uid1",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			UpsertReservedFor(&test.claim, test.pod)
			assert.ElementsMatch(t, test.claim.Status.ReservedFor, test.expected, "UpsertReservedFor() failed")
		})
	}
}

func TestRemoveReservedFor(t *testing.T) {
	type data struct {
		name  string
		pod   *v1.Pod
		claim v1beta1.ResourceClaim

		expected []v1beta1.ResourceClaimConsumerReference
	}

	tests := []data{
		{
			name: "empty claim",
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pod1",
					UID:  "uid1",
				},
			},
			claim: v1beta1.ResourceClaim{},

			expected: []v1beta1.ResourceClaimConsumerReference{},
		},
		{
			name: "claim with existing reference",
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pod1",
					UID:  "uid1",
				},
			},
			claim: v1beta1.ResourceClaim{
				Status: v1beta1.ResourceClaimStatus{
					ReservedFor: []v1beta1.ResourceClaimConsumerReference{
						{
							APIGroup: "",
							Resource: "pods",
							Name:     "pod2",
							UID:      "uid2",
						},
					},
				},
			},
			expected: []v1beta1.ResourceClaimConsumerReference{
				{
					APIGroup: "",
					Resource: "pods",
					Name:     "pod2",
					UID:      "uid2",
				},
			},
		},
		{
			name: "claim with existing identical reference",
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pod1",
					UID:  "uid1",
				},
			},
			claim: v1beta1.ResourceClaim{
				Status: v1beta1.ResourceClaimStatus{
					ReservedFor: []v1beta1.ResourceClaimConsumerReference{
						{
							APIGroup: "",
							Resource: "pods",
							Name:     "pod1",
							UID:      "uid1",
						},
					},
				},
			},
			expected: []v1beta1.ResourceClaimConsumerReference{},
		},
		{
			name: "claim with existing identical reference",
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pod1",
					UID:  "uid1",
				},
			},
			claim: v1beta1.ResourceClaim{
				Status: v1beta1.ResourceClaimStatus{
					ReservedFor: []v1beta1.ResourceClaimConsumerReference{
						{
							APIGroup: "",
							Resource: "pods",
							Name:     "pod1",
							UID:      "uid1",
						},
						{
							APIGroup: "",
							Resource: "pods",
							Name:     "pod2",
							UID:      "uid2",
						},
					},
				},
			},
			expected: []v1beta1.ResourceClaimConsumerReference{
				{
					APIGroup: "",
					Resource: "pods",
					Name:     "pod2",
					UID:      "uid2",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			RemoveReservedFor(&test.claim, test.pod)
			assert.ElementsMatch(t, test.claim.Status.ReservedFor, test.expected, "UpsertReservedFor() failed")
		})
	}
}

func TestGetResourceClaimName(t *testing.T) {
	type data struct {
		name  string
		pod   v1.Pod
		claim v1.PodResourceClaim

		expected    string
		expectedErr bool
	}

	tests := []data{
		{
			name: "simple resource claim",
			pod:  v1.Pod{},
			claim: v1.PodResourceClaim{
				Name:              "claim1",
				ResourceClaimName: pointer.String("claim1"),
			},
			expected:    "claim1",
			expectedErr: false,
		},
		{
			name: "simple resource template claim",
			pod: v1.Pod{
				Status: v1.PodStatus{
					ResourceClaimStatuses: []v1.PodResourceClaimStatus{
						{
							Name:              "claim1",
							ResourceClaimName: pointer.String("claimFromTemplate"),
						},
					},
				},
			},
			claim: v1.PodResourceClaim{
				Name:                      "claim1",
				ResourceClaimTemplateName: pointer.String("claim1"),
			},
			expected:    "claimFromTemplate",
			expectedErr: false,
		},
		{
			name: "missing resource template claim",
			pod: v1.Pod{
				Status: v1.PodStatus{
					ResourceClaimStatuses: []v1.PodResourceClaimStatus{
						{
							Name:              "another-claim",
							ResourceClaimName: pointer.String("claimFromTemplate"),
						},
					},
				},
			},
			claim: v1.PodResourceClaim{
				Name:                      "claim1",
				ResourceClaimTemplateName: pointer.String("claim1"),
			},
			expected:    "",
			expectedErr: true,
		},
		{
			name: "missing everything",
			pod: v1.Pod{
				Status: v1.PodStatus{
					ResourceClaimStatuses: []v1.PodResourceClaimStatus{
						{
							Name:              "another-claim",
							ResourceClaimName: pointer.String("claimFromTemplate"),
						},
					},
				},
			},
			claim: v1.PodResourceClaim{
				Name: "claim1",
			},
			expected:    "",
			expectedErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := GetResourceClaimName(&test.pod, &test.claim)
			if test.expectedErr {
				assert.Error(t, err, "GetResourceClaimName() succeeded, but expected an error")
			} else {
				assert.NoError(t, err, "GetResourceClaimName() failed")
				assert.Equal(t, test.expected, result, "GetResourceClaimName() returned unexpected result")
			}
		})
	}
}
