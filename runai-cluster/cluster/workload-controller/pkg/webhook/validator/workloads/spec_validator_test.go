package workloads

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Spec Validator Tests", func() {
	var (
		validator *SpecValidator
		ctx       context.Context
	)

	ctx = context.Background()
	validator = NewSpecValidator()

	It("Test a valid workload", func() {
		workload := &jobs.JobWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadName,
				Namespace: namespace,
			},
			TypeMeta: metav1.TypeMeta{
				Kind: "InteractiveWorkload",
			},
		}

		err := validator.Validate(ctx, workload)
		Expect(err).NotTo(HaveOccurred())
	})

	// TODO: uncomment this
	XIt("Test a workload with spec name", func() {
		specName := "specName"
		workload := &jobs.JobWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadName,
				Namespace: namespace,
			},
			Spec: jobs.JobSpec{
				WorkloadFields: field_groups.WorkloadFields{
					IdentityFields: field_groups.IdentityFields{},
					IdentityDeprecatedFields: field_groups.IdentityDeprecatedFields{
						Name: &field_types.StringField{Value: &specName},
					},
				},
			},
		}

		err := validator.Validate(ctx, workload)
		Expect(err.Error()).To(ContainSubstring("spec.name is deprecated, please use metadata.name"))
	})

	It("Test a workload with spec name prefix", func() {
		specNamePrefix := "prefix"
		workload := &jobs.JobWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadName,
				Namespace: namespace,
			},
			Spec: jobs.JobSpec{
				WorkloadFields: field_groups.WorkloadFields{
					IdentityFields: field_groups.IdentityFields{},
					IdentityDeprecatedFields: field_groups.IdentityDeprecatedFields{
						NamePrefix: &field_types.StringField{Value: &specNamePrefix},
					},
				},
			},
		}

		err := validator.Validate(ctx, workload)
		Expect(err.Error()).To(ContainSubstring("spec.namePrefix is deprecated, please use metadata.generateName"))
	})
})
