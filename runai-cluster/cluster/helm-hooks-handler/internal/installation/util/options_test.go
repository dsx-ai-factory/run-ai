/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package util

import (
	"flag"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicParsing(t *testing.T) {
	flags := []string{"BaseParse",
		"--pre-upgrade",
		"--control-plane-url=https://testing.example.com",
		"--control-plane-client-id=",
		"--control-plane-client-secret=C6RxwmHVLD4N3CINIPVdY4iHslzdm1ra",
		"--cluster-uid=551012a5-6759-4966-ac48-bdbd4491a31f",
		"--cluster-url=https://op-1306-iris-test.example.com",
		"--custom-ca-enabled=false",
		"--custom-ca-secret-name=runai-ca-cert",
		"--custom-ca-secret-key=runai-ca.pem",
		"--custom-ca-secret-optional=true",
		"--image-registry=",
		"--additional-image-pull-secrets=pull_secret_1,             pull_secret_2"}
	os.Args = flags
	flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	opts := parse(flagSet)
	assert.True(t, opts.PreUpgrade)
	assert.False(t, opts.PreCrdUpgrade)
	assert.False(t, opts.PreInstall)
	assert.False(t, opts.PostUpgrade)
	assert.Equal(t, "https://testing.example.com", opts.ControlPlaneURL)
	assert.False(t, opts.CustomCaEnabled)
	assert.True(t, opts.CustomCaSecretOptional)
	assert.Equal(t, "runai-ca.pem", opts.CustomCaSecretKey)
	assert.Equal(t, 2, len(opts.AdditionalImagePullSecrets))
}

func TestPreDelete(t *testing.T) {
	flags := []string{"BaseParse",
		"--pre-delete"}
	os.Args = flags
	flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	opts := parse(flagSet)
	assert.True(t, opts.PreDelete)
	assert.False(t, opts.PreInstall)
	assert.False(t, opts.PostUpgrade)
}

func TestPreCrdUpgrade(t *testing.T) {
	flags := []string{"BaseParse",
		"--pre-crd-upgrade"}
	os.Args = flags
	flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	opts := parse(flagSet)
	assert.False(t, opts.PreDelete)
	assert.True(t, opts.PreCrdUpgrade)
	assert.False(t, opts.PreInstall)
	assert.False(t, opts.PostUpgrade)
}

func TestCMDParsingAdditionalImagePullSecretsWithWhiteSpace(t *testing.T) {
	flags := []string{"1 test", "--additional-image-pull-secrets=pull_secret_1,      pull_secret_2"}
	os.Args = flags
	flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	opts := parse(flagSet)
	assert.Equal(t, 2, len(opts.AdditionalImagePullSecrets))
	assert.Equal(t, "pull_secret_1", opts.AdditionalImagePullSecrets[0])
	assert.Equal(t, "pull_secret_2", opts.AdditionalImagePullSecrets[1])
}

func TestCMDParsingAdditionalImagePullSecretsOneValue(t *testing.T) {
	flags := []string{"2 test", "--additional-image-pull-secrets=pull_secret_1"}
	os.Args = flags
	flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	opts := parse(flagSet)
	assert.Equal(t, 1, len(opts.AdditionalImagePullSecrets))
	assert.Equal(t, "pull_secret_1", opts.AdditionalImagePullSecrets[0])
	assert.True(t, opts.CMConfigParams.Affinity == nil)
	assert.Equal(t, 0, len(opts.CMConfigParams.Tolerations))
}

func TestCMDParsingAdditionalImagePullSecretsEmptyValue(t *testing.T) {
	flags := []string{"3 test", "--pre-install",
		"--control-plane-url=https://testing.example.com"}
	os.Args = flags
	flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	opts := parse(flagSet)
	assert.Equal(t, 0, len(opts.AdditionalImagePullSecrets))
}

func TestCMDParsingAdditionalImagePullSecretsExplicitEmptyValue(t *testing.T) {
	flags := []string{"3 test", "--additional-image-pull-secrets="}
	os.Args = flags
	flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	opts := parse(flagSet)
	assert.Equal(t, 0, len(opts.AdditionalImagePullSecrets))
}

func TestBasicParsingAndFile(t *testing.T) {
	flags := []string{"BaseParse",
		"--pre-upgrade",
		"--control-plane-url=https://testing.example.com",
		"--control-plane-client-id=",
		"--control-plane-client-secret=C6RxwmHVLD4N3CINIPVdY4iHslzdm1ra",
		"--cluster-uid=551012a5-6759-4966-ac48-bdbd4491a31f",
		"--cluster-url=https://op-1306-iris-test.example.com",
		"--custom-ca-enabled=false",
		"--custom-ca-secret-name=runai-ca-cert",
		"--custom-ca-secret-key=runai-ca.pem",
		"--custom-ca-secret-optional=true",
		"--image-registry=",
		"--cm-file-path=./../test/params_all_set.json",
		"--additional-image-pull-secrets=pull_secret_1,             pull_secret_2"}
	os.Args = flags
	flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	opts := parse(flagSet)
	assert.True(t, opts.PreUpgrade)
	assert.False(t, opts.PreInstall)
	assert.False(t, opts.PostUpgrade)
	assert.Equal(t, "https://testing.example.com", opts.ControlPlaneURL)
	assert.False(t, opts.CustomCaEnabled)
	assert.True(t, opts.CustomCaSecretOptional)
	assert.Equal(t, "runai-ca.pem", opts.CustomCaSecretKey)
	assert.Equal(t, 2, len(opts.AdditionalImagePullSecrets))
	assert.NotNil(t, opts.CMConfigParams.Affinity)
	assert.NotEqual(t, len(opts.CMConfigParams.Tolerations), 0)
}

func TestClientSecretFromFile(t *testing.T) {
	// Create a temporary file to hold the client secret
	tmpDir, err := os.MkdirTemp("", "client")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Write the client secret to the file
	clientSecretFilePath := path.Join(tmpDir, "client_secret")
	clientSecret := "C6RxwmHVLD4N3CINIPVdY4iHslzdm1ra"
	if err := os.WriteFile(clientSecretFilePath, []byte(clientSecret), 0644); err != nil {
		t.Fatal(err)
	}

	flags := []string{"BaseParse",
		"--pre-upgrade",
		"--control-plane-url=https://testing.example.com",
		"--control-plane-client-id=",
		"--control-plane-client-secret=@" + clientSecretFilePath,
		"--cluster-uid=551012a5-6759-4966-ac48-bdbd4491a31f",
		"--cluster-url=https://op-1306-iris-test.example.com",
		"--custom-ca-enabled=false",
		"--custom-ca-secret-name=runai-ca-cert",
		"--custom-ca-secret-key=runai-ca.pem",
		"--custom-ca-secret-optional=true",
		"--image-registry=",
		"--cm-file-path=./../test/params_all_set.json",
		"--additional-image-pull-secrets=pull_secret_1,             pull_secret_2"}
	os.Args = flags
	flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	opts := parse(flagSet)
	assert.Equal(t, clientSecret, opts.ControlPlaneClientSecret.String())
}

func TestDisableInfrastructureMetricsRemoteWrite(t *testing.T) {
	flags := []string{"TestDisable",
		"--pre-install",
		"--disable-infrastructure-metrics-remote-write"}
	os.Args = flags
	flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	opts := parse(flagSet)
	assert.True(t, opts.PreInstall)
	assert.True(t, opts.DisableInfrastructureMetricsRemoteWrite)
}

func TestDisableInfrastructureMetricsRemoteWriteDefault(t *testing.T) {
	flags := []string{"TestDefault",
		"--pre-install"}
	os.Args = flags
	flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	opts := parse(flagSet)
	assert.True(t, opts.PreInstall)
	assert.False(t, opts.DisableInfrastructureMetricsRemoteWrite)
}

func TestFipsModeDefault(t *testing.T) {
	flags := []string{"TestFipsModeDefault", "--post-install"}
	os.Args = flags
	flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	opts := parse(flagSet)
	assert.Equal(t, "off", opts.FipsMode)
}

func TestFipsModeParsing(t *testing.T) {
	flags := []string{"TestFipsModeParsing", "--post-upgrade", "--fips-mode=on"}
	os.Args = flags
	flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	opts := parse(flagSet)
	assert.Equal(t, "on", opts.FipsMode)
}
