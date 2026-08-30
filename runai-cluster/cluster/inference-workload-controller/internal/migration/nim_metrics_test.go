package migration_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/migration"
	infv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

var _ = Describe("NIM metrics migration missing API handling", func() {
	DescribeTable("classifies errors while listing inference workloads",
		func(listErr error, expectedMatcher OmegaMatcher) {
			k8sClient := fake.NewClientBuilder().WithInterceptorFuncs(interceptor.Funcs{
				List: func(context.Context, client.WithWatch, client.ObjectList, ...client.ListOption) error {
					return listErr
				},
			}).Build()

			nimMigration := migration.NewNIMMetricsServiceMigration(k8sClient)
			Expect(nimMigration.Migrate(context.Background())).To(expectedMatcher)
		},
		Entry("skips when the InferenceWorkload API is not served",
			&meta.NoKindMatchError{GroupKind: schema.GroupKind{Group: infv2.GroupVersion.Group, Kind: "InferenceWorkload"}, SearchedVersions: []string{infv2.GroupVersion.Version}},
			Succeed()),
		Entry("returns authorization errors",
			apierrors.NewForbidden(schema.GroupResource{Group: infv2.GroupVersion.Group, Resource: "inferenceworkloads"}, "", errors.New("denied")),
			MatchError(ContainSubstring("denied"))),
	)
})
