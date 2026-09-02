/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resource

import (
	"fmt"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openshiftV1 "github.com/openshift/api/route/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	runaiJobV1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	mpiv2beta1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/mpijob/v2beta1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("GetObjectResource", func() {

	DescribeTable("GetObjectResource tests",
		func(obj client.Object, exp metav1.GroupVersionKind, expErr error) {
			object := obj
			expected := exp
			result, err := GetObjectResource(object)
			if expErr == nil {
				Expect(err).To(BeNil())
			} else {
				Expect(err).To(MatchError(expErr.Error()))
			}
			Expect(result).To(Equal(expected))
		},
		Entry("nil object", nil, metav1.GroupVersionKind{}, fmt.Errorf("object is nil %v", nil)),
		Entry("PyTorchJob object", &kubeflowv1.PyTorchJob{}, metav1.GroupVersionKind{
			Kind: constants.KindPyTorchJob,
		}, nil),
		Entry("Ingress object", &networkingv1.Ingress{}, metav1.GroupVersionKind{
			Kind: constants.KindIngress,
		}, nil),
		Entry("RunaiJob object", &runaiJobV1.RunaiJob{}, metav1.GroupVersionKind{
			Kind: constants.KindRunaiJob,
		}, nil),
		Entry("MPIJob object", &mpiv2beta1.MPIJob{}, metav1.GroupVersionKind{
			Kind: constants.KindMpiJob,
		}, nil),
		Entry("TFJob object", &kubeflowv1.TFJob{}, metav1.GroupVersionKind{
			Kind: constants.KindTFJob,
		}, nil),
		Entry("XGBoostJob object", &kubeflowv1.XGBoostJob{}, metav1.GroupVersionKind{
			Kind: constants.KindXGBoostJob,
		}, nil),
		Entry("JAXJob object", &kubeflowv1.JAXJob{}, metav1.GroupVersionKind{
			Kind: constants.KindJAXJob,
		}, nil),
		Entry("CoreV1 Service object", &v1.Service{}, metav1.GroupVersionKind{
			Kind: constants.KindService,
		}, nil),
		Entry("Knative Service object", &servingv1.Service{}, metav1.GroupVersionKind{
			Group:   constants.KnativeServiceGroup,
			Version: constants.KnativeServiceVersion,
			Kind:    constants.KnativeServiceKind,
		}, nil),
		Entry("Knative Configuration object", &servingv1.Configuration{}, metav1.GroupVersionKind{
			Group:   constants.KnativeServiceGroup,
			Version: constants.KnativeServiceVersion,
			Kind:    constants.KnativeConfigurationKind,
		}, nil),
		Entry("Knative Revision object", &servingv1.Revision{}, metav1.GroupVersionKind{
			Group:   constants.KnativeServiceGroup,
			Version: constants.KnativeServiceVersion,
			Kind:    constants.KnativeRevisionKind,
		}, nil),
		Entry("Openshift Route object", &openshiftV1.Route{}, metav1.GroupVersionKind{
			Kind: constants.KindRoute,
		}, nil),
		Entry("PersistentVolumeClaim object", &v1.PersistentVolumeClaim{}, metav1.GroupVersionKind{
			Kind: constants.KindPvc,
		}, nil),
	)
})
