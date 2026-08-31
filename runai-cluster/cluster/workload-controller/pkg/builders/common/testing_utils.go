/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders"
	"os"
	"testing"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/factory"

	workloadconfig "github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"

	"github.com/ghodss/yaml"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetBuilderFn returns a specific builder object (e.g. DistributedBuilder)
// initObjs is an optional set of objects to be applied by the (fake) client,
// before creating the builder itself. it can be used, for example, to apply a policy resource
type GetBuilderFn func(inputYaml []byte, t *testing.T, expectedErr error, curTest string,
	podLabels map[string]string, podAnnotations map[string]string, initObjs ...*runtime.Object) (factory.Builder, context.Context, error)

const (
	enabledSuffix = "_ENABLED"
	keySuffix     = "_KEY"
	secretSuffix  = "_SECRET"
)

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

func LoadPolicy(t *testing.T, testName string) *v2alpha1.DistributedPolicy {

	policyPath := fmt.Sprintf("testing/policy/%s.yaml", testName)

	//
	//   read the policy yaml from the policy directory
	//
	var err error
	var policyYaml []byte

	if _, err = os.Stat(policyPath); err != nil {
		// not found, that's okay as policy is optional
		return nil
	}

	if policyYaml, err = os.ReadFile(policyPath); err != nil {
		assert.NoError(t, err, "failed to read policy file %s", policyPath)
		return nil
	}

	var policyObj v2alpha1.DistributedPolicy
	if err := yaml.Unmarshal(policyYaml, &policyObj); err != nil {
		assert.NoError(t, err, "failed to unmarshal policy file %s", policyPath)
		return nil
	}

	return &policyObj
}

const ADSEncryptionKey = "905c18f0839affcddca117dcac633f79"

type TestEnvVars struct {
	SubdomainSupportEnvVarValue   string
	IngressAnnotationsEnvVarValue string
	IngressClassEnvVarValue       string
	JupyterUserIdentityEnvVar     string
	CustomCAEnvVarValue           string
	CustomCAGitEnvVarValue        string
	CustomCAS3EnvVarValue         string
	RedisFQDN                     string
	ADSEncryptionKey              string
}

type TestConfig struct {
	TestName          string
	TestEnvVars       *TestEnvVars
	TestExpectedError string
	Focus             bool
	AddCtxFn          func(context.Context, interface{}) context.Context
	ValueToCtx        interface{}
}

// FilterTests checks if there are tests that are marked as focus only them, ow returns all tests
func FilterTests(tests []TestConfig) []TestConfig {
	var filterTests []TestConfig
	for _, curTest := range tests {
		if curTest.Focus {
			filterTests = append(filterTests, curTest)
		}
	}
	if filterTests == nil {
		return tests
	}
	return filterTests

}

func Run(t *testing.T, curTest *TestConfig, expectedErr error,
	podLabels map[string]string, podAnnotations map[string]string, getBuilderFn GetBuilderFn) {
	// load input file
	inputYaml, err := LoadInputFile(t, curTest.TestName)
	if err != nil {
		assert.NoError(t, err, "loading input test %s", curTest.TestName)
		return
	}

	// set env vars
	os.Setenv(workloadconfig.EnvVarClusterDomain, "https://cluster-domain.mock")
	defer os.Unsetenv(workloadconfig.EnvVarClusterDomain)

	os.Setenv("RUNAI_OAUTH2_PROXY_SESSION_STORE_TYPE", "redis")
	defer os.Unsetenv("RUNAI_OAUTH2_PROXY_SESSION_STORE_TYPE")

	if curTest.TestEnvVars != nil {
		if curTest.TestEnvVars.SubdomainSupportEnvVarValue != "" {
			os.Setenv(workloadconfig.EnvVarSubdomainSupport, curTest.TestEnvVars.SubdomainSupportEnvVarValue)
			defer os.Unsetenv(workloadconfig.EnvVarSubdomainSupport)
		}
		if curTest.TestEnvVars.IngressAnnotationsEnvVarValue != "" {
			os.Setenv(workloadconfig.EnvVarIngressAnnotations, curTest.TestEnvVars.IngressAnnotationsEnvVarValue)
			defer os.Unsetenv(workloadconfig.EnvVarIngressAnnotations)
		}
		if curTest.TestEnvVars.IngressClassEnvVarValue != "" {
			os.Setenv(workloadconfig.EnvVarIngressClass, curTest.TestEnvVars.IngressClassEnvVarValue)
			defer os.Unsetenv(workloadconfig.EnvVarIngressClass)
		}
		if curTest.TestEnvVars.JupyterUserIdentityEnvVar != "" {
			os.Setenv(workloadconfig.EnvVarJupyterUserIdentity, curTest.TestEnvVars.JupyterUserIdentityEnvVar)
			defer os.Unsetenv(workloadconfig.EnvVarJupyterUserIdentity)
		}
		if curTest.TestEnvVars.CustomCAEnvVarValue != "" {
			envName := builders.EnvVarPrefixCARunAi + enabledSuffix
			envNameKey := builders.EnvVarPrefixCAGit + keySuffix
			envNameSecret := builders.EnvVarPrefixCAGit + secretSuffix
			os.Setenv(envName, curTest.TestEnvVars.CustomCAGitEnvVarValue)
			os.Setenv(envNameKey, "git-key.pem")
			os.Setenv(envNameSecret, "git-secret.pem")
			defer os.Unsetenv(envName)
			defer os.Unsetenv(envNameKey)
			defer os.Unsetenv(envNameSecret)
		}
		if curTest.TestEnvVars.CustomCAGitEnvVarValue != "" {
			envName := builders.EnvVarPrefixCAGit + enabledSuffix
			envNameKey := builders.EnvVarPrefixCAGit + keySuffix
			envNameSecret := builders.EnvVarPrefixCAGit + secretSuffix
			os.Setenv(envName, curTest.TestEnvVars.CustomCAGitEnvVarValue)
			os.Setenv(envNameKey, "git-key.pem")
			os.Setenv(envNameSecret, "git-secret.pem")
			defer os.Unsetenv(envName)
			defer os.Unsetenv(envNameKey)
			defer os.Unsetenv(envNameSecret)
		}
		if curTest.TestEnvVars.CustomCAS3EnvVarValue != "" {
			envName := builders.EnvVarPrefixCAS3 + enabledSuffix
			envNameKey := builders.EnvVarPrefixCAS3 + keySuffix
			envNameSecret := builders.EnvVarPrefixCAS3 + secretSuffix
			os.Setenv(envName, curTest.TestEnvVars.CustomCAS3EnvVarValue)
			os.Setenv(envNameKey, "s3-key.pem")
			os.Setenv(envNameSecret, "s3-secret.pem")
			defer os.Unsetenv(envName)
			defer os.Unsetenv(envNameKey)
			defer os.Unsetenv(envNameSecret)
		}
		if curTest.TestEnvVars.RedisFQDN != "" {
			os.Setenv("RUNAI_OAUTH2_PROXY_REDIS_FQDN", curTest.TestEnvVars.RedisFQDN)
			defer os.Unsetenv("RUNAI_OAUTH2_PROXY_REDIS_FQDN")
		}
		if curTest.TestEnvVars.ADSEncryptionKey != "" {
			os.Setenv(constants.RunaiAdsEncryptionKeyEnvVar, curTest.TestEnvVars.ADSEncryptionKey)
			defer os.Unsetenv(constants.RunaiAdsEncryptionKeyEnvVar)
		}
	}

	// load policy, if exists
	policyObj := LoadPolicy(t, curTest.TestName)
	var runtimeObjPtr *runtime.Object
	if policyObj != nil {
		runtimeObj := runtime.Object(policyObj)
		runtimeObjPtr = &runtimeObj
	}

	// get builder
	builder, ctx, err := getBuilderFn(inputYaml, t, expectedErr, curTest.TestName,
		podLabels, podAnnotations, runtimeObjPtr)
	if curTest.AddCtxFn != nil {
		ctx = curTest.AddCtxFn(ctx, curTest.ValueToCtx)
	}
	if err == nil && builder != nil {
		err = builder.Build(ctx)
	}
	if expectedErr != nil {
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), expectedErr.Error(), "expecting error for %s", curTest.TestName)
		}
	} else {
		assert.NoError(t, err, "expecting build to be ok for %s", curTest.TestName)
	}

	if err != nil || expectedErr != nil {
		// if we expect an error or got an error, we don't need to continue (no need for the output file)
		return
	}

	var actualYaml bytes.Buffer
	err = builder.ToYaml(ctx, &actualYaml)
	if err != nil {
		assert.NoError(t, err, "building test %s", curTest.TestName)
		return
	}
	assert.NoError(t, err)

	// load/write and compare output file
	expectedYaml := LoadExpectedFile(t, curTest.TestName, &actualYaml)

	isOk := assert.Equal(t, string(expectedYaml), actualYaml.String(), curTest.TestName)
	if !isOk {
		tmpPath := fmt.Sprintf("%s/%s.yaml", os.TempDir(), curTest.TestName)
		err := os.WriteFile(tmpPath, actualYaml.Bytes(), 0755)
		if err != nil {
			assert.Fail(t, fmt.Sprintf("failed to backup the result to %s", tmpPath))
		} else {
			assert.Fail(t, fmt.Sprintf("actual result backed up to %s", tmpPath))
		}

	}
}

func TestJobBuilder(ctx context.Context, t *testing.T, curTest *TestConfig, computeBuilder ComputeBuilder, commonBuilder *SharedResourcesBuilder, jobResourceYaml ResourceYaml, commonResourceYaml ResourceYaml) {
	// load input file
	inputYaml, err := LoadInputFile(t, curTest.TestName)
	if err != nil {
		assert.NoError(t, err, "loading input test %s", curTest.TestName)
		return
	}
	var jobWorkload jobs.JobWorkload
	err = yaml.Unmarshal(inputYaml, &jobWorkload)
	assert.NoError(t, err, "Unmarshal Workload test file %s", curTest.TestName)
	if jobWorkload.Annotations == nil {
		jobWorkload.Annotations = make(map[string]string)
	}
	if jobWorkload.Labels == nil {
		jobWorkload.Labels = make(map[string]string)
	}

	assert.NoError(t, err, "Unmarshal Workload test file %s", curTest.TestName)

	// set env vars
	os.Setenv(workloadconfig.EnvVarClusterDomain, "https://cluster-domain.mock")
	defer os.Unsetenv(workloadconfig.EnvVarClusterDomain)

	os.Setenv("RUNAI_OAUTH2_PROXY_SESSION_STORE_TYPE", "redis")
	defer os.Unsetenv("RUNAI_OAUTH2_PROXY_SESSION_STORE_TYPE")

	if curTest.TestEnvVars != nil {
		if curTest.TestEnvVars.SubdomainSupportEnvVarValue != "" {
			os.Setenv(workloadconfig.EnvVarSubdomainSupport, curTest.TestEnvVars.SubdomainSupportEnvVarValue)
			defer os.Unsetenv(workloadconfig.EnvVarSubdomainSupport)
		}
		if curTest.TestEnvVars.IngressAnnotationsEnvVarValue != "" {
			os.Setenv(workloadconfig.EnvVarIngressAnnotations, curTest.TestEnvVars.IngressAnnotationsEnvVarValue)
			defer os.Unsetenv(workloadconfig.EnvVarIngressAnnotations)
		}
		if curTest.TestEnvVars.IngressClassEnvVarValue != "" {
			os.Setenv(workloadconfig.EnvVarIngressClass, curTest.TestEnvVars.IngressClassEnvVarValue)
			defer os.Unsetenv(workloadconfig.EnvVarIngressClass)
		}
		if curTest.TestEnvVars.JupyterUserIdentityEnvVar != "" {
			os.Setenv(workloadconfig.EnvVarJupyterUserIdentity, curTest.TestEnvVars.JupyterUserIdentityEnvVar)
			defer os.Unsetenv(workloadconfig.EnvVarJupyterUserIdentity)
		}
		if curTest.TestEnvVars.CustomCAEnvVarValue != "" {
			envName := builders.EnvVarPrefixCARunAi + enabledSuffix
			envNameKey := builders.EnvVarPrefixCARunAi + keySuffix
			envNameSecret := builders.EnvVarPrefixCARunAi + secretSuffix
			os.Setenv(envName, curTest.TestEnvVars.CustomCAEnvVarValue)
			os.Setenv(envNameKey, "runai-ca.pem")
			os.Setenv(envNameSecret, "runai-ca-cert")
			defer os.Unsetenv(envName)
			defer os.Unsetenv(envNameKey)
			defer os.Unsetenv(envNameSecret)
		}
		if curTest.TestEnvVars.CustomCAGitEnvVarValue != "" {
			envName := builders.EnvVarPrefixCAGit + enabledSuffix
			envNameKey := builders.EnvVarPrefixCAGit + keySuffix
			envNameSecret := builders.EnvVarPrefixCAGit + secretSuffix
			os.Setenv(envName, curTest.TestEnvVars.CustomCAGitEnvVarValue)
			os.Setenv(envNameKey, "git-key.pem")
			os.Setenv(envNameSecret, "git-secret.pem")
			defer os.Unsetenv(envName)
			defer os.Unsetenv(envNameKey)
			defer os.Unsetenv(envNameSecret)
		}
		if curTest.TestEnvVars.CustomCAS3EnvVarValue != "" {
			envName := builders.EnvVarPrefixCAS3 + enabledSuffix
			envNameKey := builders.EnvVarPrefixCAS3 + keySuffix
			envNameSecret := builders.EnvVarPrefixCAS3 + secretSuffix
			os.Setenv(envName, curTest.TestEnvVars.CustomCAS3EnvVarValue)
			os.Setenv(envNameKey, "s3-key.pem")
			os.Setenv(envNameSecret, "s3-secret.pem")
			defer os.Unsetenv(envName)
			defer os.Unsetenv(envNameKey)
			defer os.Unsetenv(envNameSecret)
		}
		if curTest.TestEnvVars.RedisFQDN != "" {
			os.Setenv("RUNAI_OAUTH2_PROXY_REDIS_FQDN", curTest.TestEnvVars.RedisFQDN)
			defer os.Unsetenv("RUNAI_OAUTH2_PROXY_REDIS_FQDN")
		}
		if curTest.TestEnvVars.ADSEncryptionKey != "" {
			os.Setenv(constants.RunaiAdsEncryptionKeyEnvVar, curTest.TestEnvVars.ADSEncryptionKey)
			defer os.Unsetenv(constants.RunaiAdsEncryptionKeyEnvVar)
		}
	}

	var actualYaml bytes.Buffer
	var commonResources []client.Object

	// only build common resources if commonBuilder is provided
	if commonBuilder != nil {
		pvclist, err := commonBuilder.PVCBuilder.Build(ctx, &jobWorkload, ResourceBuilderOptions{})
		assert.NoError(t, err, "failed to build Pvc", curTest.TestName)
		commonResources = append(commonResources, pvclist...)
		services, err := commonBuilder.BuildServices(ctx, &jobWorkload, ResourceBuilderOptions{})
		assert.NoError(t, err, "failed to build services", curTest.TestName)
		for _, service := range services {
			commonResources = append(commonResources, service)
		}
		externalUrls, err := commonBuilder.BuildExternalUrls(ctx, &jobWorkload, &ExternalBuilderOptions{})
		assert.NoError(t, err, "failed to build external urls", curTest.TestName)
		commonResources = append(commonResources, externalUrls...)
		credentials, err := commonBuilder.BuildCredentials(ctx, &jobWorkload, ResourceBuilderOptions{})
		assert.NoError(t, err, "failed to build credentials", curTest.TestName)
		commonResources = append(commonResources, credentials...)
	}

	computeResources, err := computeBuilder.Build(ctx, &jobWorkload, &ComputeBuilderOptions{})
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

	jobObject := computeResources.Job
	assert.NoError(t, err, "building test %s", curTest.TestName)
	err = jobResourceYaml.ToYaml([]client.Object{jobObject}, &actualYaml)
	assert.NoError(t, err, "failed to convert to yaml", curTest.TestName)

	// only write common resources to yaml if commonBuilder was provided
	if commonBuilder != nil {
		err = commonResourceYaml.ToYaml(commonResources, &actualYaml)
		assert.NoError(t, err, "failed to convert to yaml", curTest.TestName)
	}

	// load/write and compare output file
	expectedYaml := LoadExpectedFile(t, curTest.TestName, &actualYaml)

	isOk := assert.Equal(t, string(expectedYaml), actualYaml.String(), curTest.TestName)
	if !isOk {
		tmpPath := fmt.Sprintf("%s/%s.yaml", os.TempDir(), curTest.TestName)
		err = os.WriteFile(tmpPath, actualYaml.Bytes(), 0755)
		if err != nil {
			assert.Fail(t, fmt.Sprintf("failed to backup the result to %s", tmpPath))
		} else {
			assert.Fail(t, fmt.Sprintf("actual result backed up to %s", tmpPath))
		}
	}
}
