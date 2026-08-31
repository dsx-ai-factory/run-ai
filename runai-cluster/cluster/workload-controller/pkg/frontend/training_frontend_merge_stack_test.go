/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package frontend

import (
	"context"
	"fmt"
	"testing"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/enforcer"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/enforcer/factory"

	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	openshiftV1 "github.com/openshift/api/route/v1"
	rjv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	schedv2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	distp "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	distw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	inferencepv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferencepolicy/v2alpha1"
	inferencewv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	ipv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	iwv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	iwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	mpiv1alpha2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/mpijob/v1alpha2"
	mpiv2beta1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/mpijob/v2beta1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	networkingv1 "k8s.io/api/networking/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/utils/ptr"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	tpv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingpolicy/v2alpha1"
)

func TestCreateMergeStackTraining(t *testing.T) {

	//
	//   cannot base workload on self
	//
	selfBased := v2alpha1.TrainingWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mytemp",
			Namespace: "runai-team-a",
		},
		Spec: tpv1.TrainingSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityFields: field_groups.IdentityFields{
					BaseWorkload: ptr.To("mytemp"),
				},
			},
		},
	}
	fakeClient := NewFakeClient(t)

	trainingMergeStackCreator := factory.NewTrainingWorkloadMergeCreator(fakeClient, false)
	_, err := trainingMergeStackCreator.CreateMergeStack(context.TODO(), &selfBased)
	assert.Error(t, err, "build merge stack - self based")
	if err != nil {
		fmt.Println(err.Error())
		assert.Contains(t, err.Error(), "base workload cannot be the same workload")
	}

	//
	//    base not found
	//
	notFound := v2alpha1.TrainingWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mytemp",
			Namespace: "runai-team-a",
		},
		Spec: tpv1.TrainingSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityFields: field_groups.IdentityFields{
					BaseWorkload: ptr.To("notfound"),
				},
			},
		},
	}

	fakeClient = NewFakeClient(t, &notFound)

	trainingMergeStackCreator = factory.NewTrainingWorkloadMergeCreator(fakeClient, false)
	_, err = trainingMergeStackCreator.CreateMergeStack(context.TODO(), &notFound)

	assert.Error(t, err, "build merge stack - not found")
	if err != nil {
		assert.Contains(t, err.Error(), "failed to find base workload notfound")
	}
}

func TestLoadPreviousJobTraining(t *testing.T) {

	workingDirFromBase := "working-dir-from-base"
	baseName := "base"
	basedOnSpecName := "based-on-spec"

	baseTemplate := v2alpha1.TrainingWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      baseName,
			Namespace: "runai-team-a",
		},
		Spec: tpv1.TrainingSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityDeprecatedFields: field_groups.IdentityDeprecatedFields{
					Name: &flds.StringField{
						Value: ptr.To("name-from-base"),
					},
				},
				ContainerFields: field_groups.ContainerFields{
					WorkingDir: flds.NewStringField(workingDirFromBase),
				},
			},
			ResourceAllocationFields: field_groups.ResourceAllocationFields{
				Gpu: flds.NewOptionalNumberField(ptr.To[float64](4.0)),
			},
		},
	}

	basedOnSpec := v2alpha1.TrainingWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      basedOnSpecName,
			Namespace: "runai-team-a",
		},
		Spec: tpv1.TrainingSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityFields: field_groups.IdentityFields{
					BaseWorkload: ptr.To(baseName),
				},
			},
		},
	}

	fakeClient := NewFakeClient(t,
		&baseTemplate, &basedOnSpec)

	/////////////////////////////////////////////////////////////////////////////////////////
	//    Part 1 --> no value set in status so take from the base spec itself
	/////////////////////////////////////////////////////////////////////////////////////////
	trainingMergeStackCreator := factory.NewTrainingWorkloadMergeCreator(fakeClient, false)
	mergeStackMap, err := trainingMergeStackCreator.CreateMergeStack(context.TODO(), &basedOnSpec)
	assert.NoError(t, err, "part1 -> build the stack")

	mergeStack1 := mergeStackMap[api.SpecSubResource]
	err = mergeStack1.MergeAndValidate(context.TODO(), enforcer.MergeAndValidateOptions{
		DisableEnforce: true,
	})
	assert.NoError(t, err, "part1 -> merge")

	//
	//   the expected results are:
	//   name - not given (we do not take name from base)
	//   gpu - taken from the workload (2.0)
	//   workingDir - taken from the base
	//
	merged1, err := jobs.UnmarshalWorkload(mergeStack1.MergedValues(), &basedOnSpec.TypeMeta, &basedOnSpec.ObjectMeta)
	assert.NoError(t, err, "part1 - unmarshal")

	assert.Nil(t, err, merged1.Name, "part1 - name - nil")
	assert.Equal(t, "4", ptr.Deref(merged1.Spec.Gpu.Value, ""), "part1 - gpu - from base")
	assert.Equal(t, workingDirFromBase, merged1.Spec.WorkingDir.Deref(), "part1 - workingDir - from workload")
}

// create a fake k8s client which holds policies and workloads
func NewFakeClient(t *testing.T, initObjs ...runtime.Object) client.WithWatch {

	err := v2alpha1.AddToScheme(scheme.Scheme)
	require.Nil(t, err)

	err = tpv1.AddToScheme(scheme.Scheme)
	require.Nil(t, err)
	utilruntime.Must(ipv2.AddToScheme(scheme.Scheme)) // interactive policy
	utilruntime.Must(iwv2.AddToScheme(scheme.Scheme)) // interactive workload
	utilruntime.Must(iwv1.AddToScheme(scheme.Scheme)) // interactive workload
	utilruntime.Must(servingv1.AddToScheme(scheme.Scheme))
	utilruntime.Must(networkingv1.AddToScheme(scheme.Scheme))
	utilruntime.Must(inferencepv1.AddToScheme(scheme.Scheme)) // inference policy
	utilruntime.Must(inferencewv1.AddToScheme(scheme.Scheme)) // inference workload
	utilruntime.Must(distp.AddToScheme(scheme.Scheme))        // distributed policy
	utilruntime.Must(distw.AddToScheme(scheme.Scheme))        // distrbuted workload
	utilruntime.Must(rjv1.AddToScheme(scheme.Scheme))         // runai job
	utilruntime.Must(schedv2.AddToScheme(scheme.Scheme))      // runai department
	utilruntime.Must(mpiv1alpha2.AddKnownTypes(scheme.Scheme))
	utilruntime.Must(mpiv1alpha2.AddToScheme(scheme.Scheme)) // mpi job
	utilruntime.Must(mpiv2beta1.AddToScheme(scheme.Scheme))  // mpi job
	utilruntime.Must(kubeflowv1.AddToScheme(scheme.Scheme))  // pytorch job, tfjob
	utilruntime.Must(v1alpha1.AddToScheme(scheme.Scheme))    // nodepools
	utilruntime.Must(servingv1.AddToScheme(scheme.Scheme))   // knativeservice
	utilruntime.Must(openshiftV1.AddToScheme(scheme.Scheme))

	return fake.NewClientBuilder().WithRuntimeObjects(initObjs...).Build()
}
