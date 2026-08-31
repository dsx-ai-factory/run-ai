/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"

	openshiftV1 "github.com/openshift/api/route/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"github.com/ghodss/yaml"
	workloadconfig "github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	dinfv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	lwsv1 "sigs.k8s.io/lws/api/leaderworkerset/v1"
)

const namespace = "runai-team-a"

type TestConfig struct {
	TestName          string
	TestExpectedError string
	TestEnvVars       *TestEnvVars
}

type TestEnvVars struct {
	IngressClassEnvVarValue string
}

func LoadInputFile(t *testing.T, testName string) ([]byte, error) {

	inputPath := fmt.Sprintf("testing/inputs/%s.yaml", testName)

	//
	//   read the Workload yaml from the inputs directory
	//
	var err error
	var inputYaml []byte

	if _, err = os.Stat(inputPath); err != nil {
		assert.NoError(t, err, "expecting to find input file %s", inputPath)
		return nil, err
	}

	if inputYaml, err = os.ReadFile(inputPath); err != nil {
		assert.NoError(t, err, "failed to read input file %s", inputPath)
		return nil, err
	}

	return inputYaml, nil
}

func LoadExpectedFile(t *testing.T, testName string, actualYaml *bytes.Buffer) []byte {

	expectedPath := fmt.Sprintf("testing/expected/%s.yaml", testName)

	if _, err := os.Stat(expectedPath); errors.Is(err, os.ErrNotExist) {
		err := os.WriteFile(expectedPath, actualYaml.Bytes(), 0755)
		if err != nil {
			assert.Fail(t, "expected file not found and failed to create it")
		} else {
			assert.Fail(t, "expected file not found, created it. make sure to review it!")
		}
	}

	expectedYaml, err := os.ReadFile(expectedPath)
	assert.NoError(t, err, "could not load expected file %s", expectedPath)

	return expectedYaml
}

type ResourceYaml interface {
	ToYaml(objects []client.Object, w *bytes.Buffer) error
}

var testFixedDate = metav1.NewTime(time.Date(2022, 2, 22, 17, 0, 0, 0, time.UTC))

type TestKnativeBuilder struct {
}

func (self *TestKnativeBuilder) ToYaml(objects []client.Object, w *bytes.Buffer) error {
	if len(objects) != 1 {
		return fmt.Errorf("expected 1 object, got %d", len(objects))
	}
	knativeService, ok := objects[0].(*servingv1.Service)
	if !ok {
		return fmt.Errorf("expected *servingv1.Service, got %T", objects[0])
	}
	knativeService.SetCreationTimestamp(testFixedDate)
	knativeService.SetDeletionTimestamp(&testFixedDate)
	knativeService.Kind = constants.KindService
	knativeService.APIVersion = constants.KnativeServiceApi
	marshalled, err := yaml.Marshal(objects[0])

	if err != nil {
		return err
	}

	_, err = w.Write(marshalled)
	if err != nil {
		return err
	}
	return nil
}

func DistributedInferenceResourcesToYaml(objects []client.Object, result *bytes.Buffer) error {
	var fixedDate = metav1.NewTime(time.Date(2022, 2, 22, 17, 0, 0, 0, time.UTC))
	var yamlSeparator = []byte("---\n")

	resources := make(map[string][]client.Object)
	for _, obj := range objects {
		switch obj.(type) {
		case *lwsv1.LeaderWorkerSet:
			resources[constants.LwsKind] = append(resources[constants.LwsKind], obj)
		case *v1.PersistentVolumeClaim:
			resources[constants.KindPvc] = append(resources[constants.KindPvc], obj)
		case *v1.Service:
			resources[constants.KindService] = append(resources[constants.KindService], obj)
		case *networkingv1.Ingress:
			resources[constants.KindIngress] = append(resources[constants.KindIngress], obj)
		case *openshiftV1.Route:
			resources[constants.KindRoute] = append(resources[constants.KindRoute], obj)
		case *nvapi.ComputeDomain:
			resources[constants.KindComputeDomain] = append(resources[constants.KindComputeDomain], obj)
		default:
			return fmt.Errorf("unknown resource type %v", obj)
		}
	}

	for _, lwsObj := range resources[constants.LwsKind] {
		lws := lwsObj.(*lwsv1.LeaderWorkerSet)
		lws.Kind = constants.LwsKind
		lws.APIVersion = fmt.Sprintf("%s/%s", constants.LwsGroup, constants.LwsVersion)
		lws.CreationTimestamp = fixedDate
		lws.DeletionTimestamp = &fixedDate
		marshalled, err := yaml.Marshal(lws)
		if err != nil {
			return err
		}
		result.Write(marshalled)
	}

	for _, pvcObj := range resources[constants.KindPvc] {
		pvc := pvcObj.(*v1.PersistentVolumeClaim)
		pvc.Kind = constants.KindPvc
		pvc.CreationTimestamp = fixedDate
		marshalled, err := yaml.Marshal(pvc)
		if err != nil {
			return err
		}
		result.Write(yamlSeparator)
		result.Write(marshalled)
	}

	for _, serviceObj := range resources[constants.KindService] {
		service := serviceObj.(*v1.Service)
		service.Kind = constants.KindService
		service.CreationTimestamp = fixedDate
		marshalled, err := yaml.Marshal(service)
		if err != nil {
			return err
		}
		result.Write(yamlSeparator)
		result.Write(marshalled)
	}

	for _, ingressObj := range resources[constants.KindIngress] {
		ingress := ingressObj.(*networkingv1.Ingress)
		ingress.Kind = constants.KindIngress
		ingress.CreationTimestamp = fixedDate
		marshalled, err := yaml.Marshal(ingress)
		if err != nil {
			return err
		}
		result.Write(yamlSeparator)
		result.Write(marshalled)
	}

	for _, routeObj := range resources[constants.KindRoute] {
		route := routeObj.(*openshiftV1.Route)
		route.Kind = constants.KindRoute
		route.CreationTimestamp = fixedDate
		marshalled, err := yaml.Marshal(route)
		if err != nil {
			return err
		}
		result.Write(yamlSeparator)
		result.Write(marshalled)
	}

	for _, cdObj := range resources[constants.KindComputeDomain] {
		cd := cdObj.(*nvapi.ComputeDomain)
		cd.Kind = constants.KindComputeDomain
		cd.CreationTimestamp = fixedDate
		marshalled, err := yaml.Marshal(cd)
		if err != nil {
			return err
		}
		result.Write(yamlSeparator)
		result.Write(marshalled)
	}

	return nil
}

func TestDistributedInferenceWorkloadBuilder(ctx context.Context, t *testing.T, curTest *TestConfig, builder *DistributedInferenceBuilder) {
	// load input file
	inputYaml, err := LoadInputFile(t, curTest.TestName)
	if err != nil {
		assert.NoError(t, err, "loading input test %s", curTest.TestName)
		return
	}
	var workload dinfv1.DistributedInferenceWorkload
	err = yaml.Unmarshal(inputYaml, &workload)
	assert.NoError(t, err, "Unmarshal Workload test file %s", curTest.TestName)
	if workload.Annotations == nil {
		workload.Annotations = make(map[string]string)
	}
	if workload.Labels == nil {
		workload.Labels = make(map[string]string)
	}

	assert.NoError(t, err, "Unmarshal Workload test file %s", curTest.TestName)

	// set env vars
	_ = os.Setenv(workloadconfig.EnvVarClusterDomain, "https://cluster-domain.mock")
	defer func() { _ = os.Unsetenv(workloadconfig.EnvVarClusterDomain) }()

	_ = os.Setenv("RUNAI_OAUTH2_PROXY_SESSION_STORE_TYPE", "redis")
	defer func() { _ = os.Unsetenv("RUNAI_OAUTH2_PROXY_SESSION_STORE_TYPE") }()

	if curTest.TestEnvVars != nil {
		if curTest.TestEnvVars.IngressClassEnvVarValue != "" {
			_ = os.Setenv(workloadconfig.EnvVarIngressClass, curTest.TestEnvVars.IngressClassEnvVarValue)
			defer func() { _ = os.Unsetenv(workloadconfig.EnvVarIngressClass) }()
		}
	}

	var actualYaml bytes.Buffer

	// build lws
	objects, err := builder.Build(ctx, &workload, nil)
	if curTest.TestExpectedError != "" {
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), curTest.TestExpectedError, "expecting error for %s", curTest.TestName)
		}
	} else {
		assert.NoError(t, err, "expecting build to be ok for %s", curTest.TestName)
	}

	if err != nil || curTest.TestExpectedError != "" {
		// if we expect an error or got an error, we don't need to continue (no need for the output file)
		return
	}

	assert.NoError(t, err, "building test %s", curTest.TestName)

	err = DistributedInferenceResourcesToYaml(objects, &actualYaml)
	assert.NoError(t, err, "failed to convert resources to yaml", curTest.TestName)

	// load/write and compare output file
	expectedYaml := LoadExpectedFile(t, curTest.TestName, &actualYaml)

	isOk := assert.Equal(t, string(expectedYaml), actualYaml.String(), curTest.TestName)
	if !isOk {
		tmpPath := fmt.Sprintf("%s/%s.yaml", os.TempDir(), curTest.TestName)
		// if tmpPath folder doesn't exist, create it
		tmpFolder := filepath.Dir(tmpPath)
		if _, err := os.Stat(tmpFolder); os.IsNotExist(err) {
			err = os.MkdirAll(tmpFolder, 0755)
			if err != nil {
				assert.Fail(t, fmt.Sprintf("failed to create tmpPath folder %s", tmpPath))
			}
		}
		err = os.WriteFile(tmpPath, actualYaml.Bytes(), 0755)
		if err != nil {
			assert.Fail(t, fmt.Sprintf("failed to backup the result to %s", tmpPath))
		} else {
			assert.Fail(t, fmt.Sprintf("actual result backed up to %s", tmpPath))
		}
	}
}
