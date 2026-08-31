/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	"context"
	"errors"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/internal/autogenerate/mocks"
	dwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	iwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	twv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	workloadCommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	"github.com/stretchr/testify/mock"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

const (
	TestClusterVersion = "1.2.3.4"
	TwName             = "training-workload"
	IwName             = "training-workload"
	DwName             = "training-workload"
	Namespace          = "test-namespace"
	MaxWorkers         = 20
)

var (
	preVersionLabelValue = fmt.Sprintf(PreVersionLabelFormat, TestClusterVersion)
)

var _ = Describe("ClusterVersionLabelHandler", func() {
	var (
		handler    *ClusterVersionLabelHandler
		fakeClient client.Client
		tw         *twv2.TrainingWorkload
		iw         *iwv2.InteractiveWorkload
		dw         *dwv2.DistributedWorkload
		scheme     *runtime.Scheme
		ctx        context.Context
	)

	BeforeEach(func() {
		tw, iw, dw = getTestWorkloads("")

		ctx = context.Background()
		scheme = runtime.NewScheme()
		Expect(twv2.AddToScheme(scheme)).To(Succeed())
		Expect(iwv2.AddToScheme(scheme)).To(Succeed())
		Expect(dwv2.AddToScheme(scheme)).To(Succeed())

		fakeClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(tw, iw, dw).Build()
		handler = NewClusterVersionLabelHandler(fakeClient, TestClusterVersion)
	})

	DescribeTableSubtree("handle workloads",
		func(workloadObject client.Object, workloadName, kind string) {
			It("should add the cluster version label", func() {
				err := fakeClient.Get(ctx, client.ObjectKey{Name: workloadName, Namespace: Namespace}, workloadObject)
				Expect(err).NotTo(HaveOccurred())

				err = handler.Handle(ctx, workloadObject)
				Expect(err).NotTo(HaveOccurred())

				fetchAndValidateVersionLabel(workloadName, workloadObject, fakeClient)
			})

			It("patch should keep existing labels", func() {
				err := fakeClient.Get(ctx, client.ObjectKey{Name: workloadName, Namespace: Namespace}, workloadObject)
				Expect(err).NotTo(HaveOccurred())

				anotherLabelKey := "key"
				anotherLabelValue := "blablabla"
				setLabelAndUpdate(workloadObject, fakeClient, anotherLabelKey, anotherLabelValue)

				err = handler.Handle(ctx, workloadObject)
				Expect(err).NotTo(HaveOccurred())

				fetchAndValidateVersionLabel(workloadName, workloadObject, fakeClient)
				validateLabel(workloadObject, anotherLabelKey, anotherLabelValue)
			})

			//should not happen during migration
			Context("should not fail if label already exist", func() {
				It("same value", func() {
					err := fakeClient.Get(ctx, client.ObjectKey{Name: workloadName, Namespace: Namespace}, workloadObject)
					Expect(err).NotTo(HaveOccurred())

					setLabelAndUpdate(workloadObject, fakeClient, workloadCommon.CreatedAtClusterVersionLabel, preVersionLabelValue)

					err = handler.Handle(ctx, workloadObject)
					Expect(err).NotTo(HaveOccurred())

					fetchAndValidateVersionLabel(workloadName, workloadObject, fakeClient)
				})

				It("different value", func() {
					err := fakeClient.Get(ctx, client.ObjectKey{Name: workloadName, Namespace: Namespace}, workloadObject)
					Expect(err).NotTo(HaveOccurred())

					setLabelAndUpdate(workloadObject, fakeClient, workloadCommon.CreatedAtClusterVersionLabel, "blablabla")

					err = handler.Handle(ctx, workloadObject)
					Expect(err).NotTo(HaveOccurred())

					fetchAndValidateVersionLabel(workloadName, workloadObject, fakeClient)
				})
			})
		},
		Entry("training", &twv2.TrainingWorkload{}, TwName, constants.KindTrainingWorkload),
		Entry("interactive", &iwv2.InteractiveWorkload{}, IwName, constants.KindInteractiveWorkload),
		Entry("distributed", &dwv2.DistributedWorkload{}, DwName, constants.KindDistributedWorkload),
	)

	It("should not fail if not found", func() {
		w := &twv2.TrainingWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "i was not created",
				Namespace: Namespace,
			},
		}

		err := handler.Handle(ctx, w)
		Expect(err).NotTo(HaveOccurred())
	})
})

var _ = Describe("LabelWorkloads", func() {
	var (
		labelWorkloads *LabelWorkloads
		fakeClient     client.Client
		mockExecutor   *mocks.Executor
		scheme         *runtime.Scheme
		tw1, tw2       *twv2.TrainingWorkload
		iw1, iw2       *iwv2.InteractiveWorkload
		dw1, dw2       *dwv2.DistributedWorkload
		testObjects    []client.Object
		ctx            context.Context
	)
	scheme = runtime.NewScheme()
	ctx = context.Background()

	BeforeEach(func() {
		tw1, iw1, dw1 = getTestWorkloads("1")
		tw2, iw2, dw2 = getTestWorkloads("2")
		Expect(twv2.AddToScheme(scheme)).To(Succeed())
		Expect(iwv2.AddToScheme(scheme)).To(Succeed())
		Expect(dwv2.AddToScheme(scheme)).To(Succeed())
		testObjects = []client.Object{tw1, tw2, iw1, iw2, dw1, dw2}

		fakeClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(testObjects...).Build()
		mockExecutor = &mocks.Executor{}

		labelWorkloads = getTestLabelWorkloads(fakeClient, mockExecutor, nil, nil)
	})

	It("unable to create migrator if cluster version is empty", func() {
		labelWorkloads, err := NewLabelWorkloads(fakeClient, "", nil, nil, MaxWorkers)

		Expect(err).To(HaveOccurred())
		Expect(labelWorkloads).To(BeNil())
	})

	It("wl override", func() {
		labelWorkloads = getTestLabelWorkloads(fakeClient, mockExecutor, []WorkloadList{}, nil)
		mockExecutor.On("Execute", mock.Anything).Return(errors.New("should not return this"))

		err := labelWorkloads.Migrate(ctx)
		Expect(err).NotTo(HaveOccurred())
	})

	It("skips a workload kind whose API is not served", func() {
		missingAPIClient := fake.NewClientBuilder().WithScheme(scheme).WithInterceptorFuncs(interceptor.Funcs{
			List: func(context.Context, client.WithWatch, client.ObjectList, ...client.ListOption) error {
				return &meta.NoKindMatchError{GroupKind: iwv2.GroupVersion.WithKind(constants.KindInteractiveWorkload).GroupKind(), SearchedVersions: []string{iwv2.GroupVersion.Version}}
			},
		}).Build()
		labelWorkloads = getTestLabelWorkloads(missingAPIClient, mockExecutor, []WorkloadList{&iwv2.InteractiveWorkloadList{}}, nil)
		createCalled := false
		labelWorkloads.MigrateExecuteCreateFn = func(int, []client.Object, func(context.Context, client.Object) error) Executor {
			createCalled = true
			return mockExecutor
		}

		Expect(labelWorkloads.Migrate(ctx)).To(Succeed())
		Expect(createCalled).To(BeFalse())
	})

	It("returns authorization errors while listing a workload kind", func() {
		forbiddenErr := apierrors.NewForbidden(schema.GroupResource{Group: iwv2.GroupVersion.Group, Resource: "interactiveworkloads"}, "", errors.New("denied"))
		forbiddenClient := fake.NewClientBuilder().WithScheme(scheme).WithInterceptorFuncs(interceptor.Funcs{
			List: func(context.Context, client.WithWatch, client.ObjectList, ...client.ListOption) error {
				return forbiddenErr
			},
		}).Build()
		labelWorkloads = getTestLabelWorkloads(forbiddenClient, mockExecutor, []WorkloadList{&iwv2.InteractiveWorkloadList{}}, nil)

		Expect(labelWorkloads.Migrate(ctx)).To(MatchError(forbiddenErr))
	})

	It("handler override", func() {
		errMsg := "all i do is error"
		handler := mocks.Handler[client.Object]{}
		handler.On("Handle", mock.Anything, mock.Anything).Return(errors.New(errMsg))
		labelWorkloads = getTestLabelWorkloads(fakeClient, mockExecutor, nil, &handler)
		labelWorkloads.MigrateExecuteCreateFn = nil

		err := labelWorkloads.Migrate(ctx)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(errMsg))
	})

	It("migration should succeed", func() {
		mockExecutor.On("Execute", mock.Anything).Return(nil)
		err := labelWorkloads.Migrate(ctx)
		Expect(err).NotTo(HaveOccurred())
	})

	It("migration should use pagination", func() {
		labelWorkloads.PaginationLimit = 1
		mockExecutor.On("Execute", mock.Anything).Return(nil).Times(len(testObjects))
		err := labelWorkloads.Migrate(ctx)
		Expect(err).NotTo(HaveOccurred())
	})

	It("migration should fail on executor error", func() {
		execError := "exec error"
		mockExecutor.On("Execute", mock.Anything).Return(errors.New(execError))

		err := labelWorkloads.Migrate(ctx)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(execError))
	})

	Context("next batch function should filter out labeled", func() {
		BeforeEach(func() {
			workloadsToLabel := []client.Object{tw1, iw1, dw1}
			for _, w := range workloadsToLabel {
				setLabelAndUpdate(w, fakeClient, workloadCommon.CreatedAtClusterVersionLabel, preVersionLabelValue)
			}
		})

		DescribeTable("by workload type",
			func(wl WorkloadList) {
				validateBatchGetter(labelWorkloads.generateNextBatchGetterFn(wl), 1)
			},
			Entry("training", &twv2.TrainingWorkloadList{}),
			Entry("interactive", &iwv2.InteractiveWorkloadList{}),
			Entry("distributed", &dwv2.DistributedWorkloadList{}),
		)
	})
})

func getTestLabelWorkloads(fakeClient client.Client, executor Executor, wlListOverride []WorkloadList, handlerOverride Handler[client.Object]) *LabelWorkloads {
	labelWorkloads, err := NewLabelWorkloads(fakeClient, TestClusterVersion, wlListOverride, handlerOverride, MaxWorkers)
	Expect(err).NotTo(HaveOccurred())
	Expect(labelWorkloads).NotTo(BeNil())

	labelWorkloads.MigrateExecuteCreateFn = func(int, []client.Object, func(context.Context, client.Object) error) Executor {
		return executor
	}

	return labelWorkloads
}

func fetchAndValidateVersionLabel(workloadName string, workloadObject client.Object, fakeClient client.Client) {
	err := fakeClient.Get(context.Background(), client.ObjectKey{Name: workloadName, Namespace: Namespace}, workloadObject)
	Expect(err).NotTo(HaveOccurred())
	validateLabel(workloadObject, workloadCommon.CreatedAtClusterVersionLabel, preVersionLabelValue)
}

func validateLabel(workloadObject client.Object, key, value string) {
	labelValue, ok := workloadObject.GetLabels()[key]
	Expect(ok).To(BeTrue())
	Expect(labelValue).To(Equal(value))
}

func setLabelAndUpdate(workloadObject client.Object, fakeClient client.Client, key, value string) {
	labels := workloadObject.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}

	labels[key] = value
	workloadObject.SetLabels(labels)
	err := fakeClient.Update(context.Background(), workloadObject)
	Expect(err).NotTo(HaveOccurred())
}

func validateBatchGetter(nextBatchGetter func(continueToken string, ctx context.Context) ([]client.Object, string, error), expectedBatchSize int) {
	batch, cont, err := nextBatchGetter("", context.Background())

	Expect(err).NotTo(HaveOccurred())
	Expect(len(batch)).To(Equal(expectedBatchSize))
	Expect(cont).To(Equal(""))
}

func getTestWorkloads(nameSuffix string) (*twv2.TrainingWorkload, *iwv2.InteractiveWorkload, *dwv2.DistributedWorkload) {
	tw := &twv2.TrainingWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      TwName + nameSuffix,
			Namespace: Namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind: constants.KindTrainingWorkload,
		},
	}
	iw := &iwv2.InteractiveWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      IwName + nameSuffix,
			Namespace: Namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind: constants.KindInteractiveWorkload,
		},
	}
	dw := &dwv2.DistributedWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DwName + nameSuffix,
			Namespace: Namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind: constants.KindDistributedWorkload,
		},
	}

	return tw, iw, dw
}
