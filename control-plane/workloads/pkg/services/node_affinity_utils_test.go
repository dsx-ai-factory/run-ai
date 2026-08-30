package services

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models"
)

var _ = Describe("NodeAffinity Conversion Utilities", func() {
	Describe("NodeAffinityModelToApi and NodeAffinityApiToModel", func() {
		It("should round-trip convert a full NodeAffinity struct", func() {
			model := models.NodeAffinity{
				Required: &models.NodeSelector{
					NodeSelectorTerms: []models.NodeSelectorTerm{{
						MatchExpressions: []models.NodeSelectorRequirement{
							{Key: "k1", Operator: "In", Values: []string{"v1", "v2"}},
							{Key: "k2", Operator: "Exists"},
						},
					}},
				},
				Preferred: &models.PreferredNodeSelector{
					NodeSelectorTerms: []models.PreferredNodeSelectorTerm{{
						Weight: 10,
						Preference: models.NodeSelectorTerm{
							MatchExpressions: []models.NodeSelectorRequirement{{Key: "k3", Operator: "NotIn", Values: []string{"v3"}}},
						},
					}},
				},
			}
			apiVal := NodeAffinityModelToApi(model)

			Expect(apiVal.Required).NotTo(BeNil())
			Expect(apiVal.Required.NodeSelectorTerms).To(HaveLen(1))
			Expect(apiVal.Required.NodeSelectorTerms[0].MatchExpressions).To(HaveLen(2))
			Expect(apiVal.Preferred).NotTo(BeNil())
			Expect(apiVal.Preferred.NodeSelectorTerms).To(HaveLen(1))
			Expect(apiVal.Preferred.NodeSelectorTerms[0].Weight).To(Equal(float32(10)))
			Expect(apiVal.Preferred.NodeSelectorTerms[0].Preference.MatchExpressions).To(HaveLen(1))
			Expect(apiVal.Preferred.NodeSelectorTerms[0].Preference.MatchExpressions[0].Operator).To(Equal(api.MATCHEXPRESSIONOPERATOR_NOT_IN))

			modelBack := NodeAffinityApiToModel(apiVal)
			Expect(modelBack).To(Equal(model))
		})

		It("should handle empty and nil NodeAffinity", func() {
			model := models.NodeAffinity{}
			apiVal := NodeAffinityModelToApi(model)
			Expect(apiVal.Required).To(BeNil())
			Expect(apiVal.Preferred).To(BeNil())
			modelBack := NodeAffinityApiToModel(apiVal)
			Expect(modelBack).To(Equal(model))

			model = models.NodeAffinity{Required: nil, Preferred: nil}
			apiVal = NodeAffinityModelToApi(model)
			Expect(apiVal.Required).To(BeNil())
			Expect(apiVal.Preferred).To(BeNil())
			modelBack = NodeAffinityApiToModel(apiVal)
			Expect(modelBack).To(Equal(model))
		})
	})

	Describe("PreferredNodeSelectorTermModelToApi and PreferredNodeSelectorTermApiToModel", func() {
		It("should round-trip convert PreferredNodeSelectorTerm", func() {
			model := models.PreferredNodeSelectorTerm{
				Weight: 42,
				Preference: models.NodeSelectorTerm{
					MatchExpressions: []models.NodeSelectorRequirement{{Key: "foo", Operator: "Gt", Values: []string{"1"}}},
				},
			}
			apiVal := PreferredNodeSelectorTermModelToApi(model)
			Expect(apiVal.Weight).To(Equal(float32(42)))
			Expect(apiVal.Preference).NotTo(BeNil())
			Expect(apiVal.Preference.MatchExpressions).To(HaveLen(1))
			modelBack := PreferredNodeSelectorTermApiToModel(apiVal)
			Expect(modelBack).To(Equal(model))
		})
	})

	Describe("matchExpressionModelToApi and matchExpressionApiToModel", func() {
		It("should round-trip convert NodeSelectorRequirement/MatchExpression", func() {
			model := models.NodeSelectorRequirement{Key: "bar", Operator: "Lt", Values: []string{"5"}}
			apiVal := matchExpressionModelToApi(model)
			Expect(apiVal.Key).To(Equal("bar"))
			Expect(apiVal.Operator).To(Equal(api.MATCHEXPRESSIONOPERATOR_LT))
			Expect(apiVal.Values).To(Equal([]string{"5"}))
			modelBack := matchExpressionApiToModel(apiVal)
			Expect(modelBack).To(Equal(model))
		})
	})

	Describe("workloadNodeAffinityApiToModel and workloadNodeAffinityModelToApi", func() {
		It("should round-trip convert a WorkloadNodeAffinity with full data", func() {
			apiVal := api.WorkloadNodeAffinity{
				SpecName: "worker",
				NodeAffinity: api.PodNodeAffinity{
					Required: &api.PodNodeAffinityRequired{
						NodeSelectorTerms: []api.NodeSelectorTerm{{
							MatchExpressions: []api.MatchExpression{
								{Key: "gpu-type", Operator: api.MATCHEXPRESSIONOPERATOR_IN, Values: []string{"a100", "h100"}},
							},
						}},
					},
					Preferred: &api.PodNodeAffinityPreferred{
						NodeSelectorTerms: []api.PodPreferredNodeSelectorTerm{{
							Weight:     50,
							Preference: &api.NodeSelectorTerm{MatchExpressions: []api.MatchExpression{{Key: "zone", Operator: api.MATCHEXPRESSIONOPERATOR_IN, Values: []string{"us-east-1"}}}},
						}},
					},
				},
			}

			model := workloadNodeAffinityApiToModel(apiVal)
			Expect(model.SpecName).To(Equal("worker"))
			Expect(model.NodeAffinity.Required).NotTo(BeNil())
			Expect(model.NodeAffinity.Required.NodeSelectorTerms).To(HaveLen(1))
			Expect(model.NodeAffinity.Required.NodeSelectorTerms[0].MatchExpressions[0].Key).To(Equal("gpu-type"))
			Expect(model.NodeAffinity.Preferred).NotTo(BeNil())

			back := workloadNodeAffinityModelToApi(model)
			Expect(back.SpecName).To(Equal(apiVal.SpecName))
			Expect(back.NodeAffinity.Required.NodeSelectorTerms).To(HaveLen(1))
			Expect(back.NodeAffinity.Required.NodeSelectorTerms[0].MatchExpressions[0].Key).To(Equal("gpu-type"))
			Expect(back.NodeAffinity.Preferred.NodeSelectorTerms[0].Weight).To(Equal(float32(50)))
		})

		It("should handle empty NodeAffinity", func() {
			apiVal := api.WorkloadNodeAffinity{
				SpecName:     "leader",
				NodeAffinity: api.PodNodeAffinity{},
			}

			model := workloadNodeAffinityApiToModel(apiVal)
			Expect(model.SpecName).To(Equal("leader"))
			Expect(model.NodeAffinity.Required).To(BeNil())
			Expect(model.NodeAffinity.Preferred).To(BeNil())

			back := workloadNodeAffinityModelToApi(model)
			Expect(back.SpecName).To(Equal("leader"))
			Expect(back.NodeAffinity.Required).To(BeNil())
			Expect(back.NodeAffinity.Preferred).To(BeNil())
		})
	})

	Describe("Edge cases", func() {
		It("should handle nil and empty match expressions", func() {
			model := models.NodeAffinity{
				Required: &models.NodeSelector{
					NodeSelectorTerms: []models.NodeSelectorTerm{{}},
				},
			}
			apiVal := NodeAffinityModelToApi(model)
			Expect(apiVal.Required).NotTo(BeNil())
			Expect(apiVal.Required.NodeSelectorTerms).To(HaveLen(1))
			Expect(apiVal.Required.NodeSelectorTerms[0].MatchExpressions).To(BeEmpty())
			modelBack := NodeAffinityApiToModel(apiVal)
			Expect(modelBack).To(Equal(model))
		})

		It("should handle zero weights and missing preference in preferred terms", func() {
			model := models.NodeAffinity{
				Preferred: &models.PreferredNodeSelector{
					NodeSelectorTerms: []models.PreferredNodeSelectorTerm{{Weight: 0}},
				},
			}

			apiVal := NodeAffinityModelToApi(model)
			Expect(apiVal.Preferred).NotTo(BeNil())
			Expect(apiVal.Preferred.NodeSelectorTerms).To(HaveLen(1))
			Expect(apiVal.Preferred.NodeSelectorTerms[0].Weight).To(Equal(float32(0)))
			Expect(apiVal.Preferred.NodeSelectorTerms[0].Preference).NotTo(BeNil())

			modelBack := NodeAffinityApiToModel(apiVal)
			Expect(modelBack).To(Equal(model))
		})

		It("should handle invalid operator", func() {
			model := models.NodeSelectorRequirement{Key: "k1", Operator: "NotARealOp"}

			apiVal := matchExpressionModelToApi(model)
			Expect(string(apiVal.Operator)).To(Equal("NotARealOp"))

			modelBack := matchExpressionApiToModel(apiVal)
			Expect(modelBack.Operator).To(Equal("NotARealOp"))
		})
	})
})
