/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"flag"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/utils/options"
	"os"
	"reflect"
	"testing"
)

func TestSetOptions(t *testing.T) {
	// Save original args and restore after test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Test with empty args (defaults)
	os.Args = []string{"program"}
	opts, err := SetOptions()
	if err != nil {
		t.Fatalf("SetOptions() failed: %v", err)
	}

	expected := &Options{
		UseDebugLogLevel:                        false,
		CertDir:                                 "/etc/webhook/certs",
		WebhookPort:                             8443,
		EnableLeaderElection:                    false,
		LegacyProjectPermissions:                true,
		LegacyProjectPermissionsIntervalSeconds: 30,
		AllowedServiceAccounts:                  nil,
	}

	if !reflect.DeepEqual(opts, expected) {
		t.Errorf("SetOptions() = %+v, want %+v", opts, expected)
	}
}

func TestParseCommandLineArgs_Defaults(t *testing.T) {
	opts := &Options{}
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)

	err := opts.parseCommandLineArgs(flagSet, []string{})
	if err != nil {
		t.Fatalf("parseCommandLineArgs() failed: %v", err)
	}

	expected := &Options{
		UseDebugLogLevel:                        false,
		CertDir:                                 "/etc/webhook/certs",
		WebhookPort:                             8443,
		EnableLeaderElection:                    false,
		LegacyProjectPermissions:                true,
		LegacyProjectPermissionsIntervalSeconds: 30,
		AllowedServiceAccounts:                  nil,
	}

	if !reflect.DeepEqual(opts, expected) {
		t.Errorf("parseCommandLineArgs() with defaults = %+v, want %+v", opts, expected)
	}
}

func TestParseCommandLineArgs_AllFlags(t *testing.T) {
	opts := &Options{}
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)

	args := []string{
		"-debug",
		"-cert-dir", "/custom/cert/dir",
		"-webhook-port", "9000",
		"-leader-elect",
		"-legacy-project-permissions=false",
		"-legacy-project-permissions-interval-seconds", "60",
		"-allowed-service-accounts", "account1,account2, account3",
	}

	err := opts.parseCommandLineArgs(flagSet, args)
	if err != nil {
		t.Fatalf("parseCommandLineArgs() failed: %v", err)
	}

	expected := &Options{
		UseDebugLogLevel:                        true,
		CertDir:                                 "/custom/cert/dir",
		WebhookPort:                             9000,
		EnableLeaderElection:                    true,
		LegacyProjectPermissions:                false,
		LegacyProjectPermissionsIntervalSeconds: 60,
		AllowedServiceAccounts:                  options.ArrayFlags{"account1", "account2", "account3"},
	}

	if !reflect.DeepEqual(opts, expected) {
		t.Errorf("parseCommandLineArgs() with all flags = %+v, want %+v", opts, expected)
	}
}

func TestParseCommandLineArgs_IndividualFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected func() *Options
	}{
		{
			name: "debug flag",
			args: []string{"-debug"},
			expected: func() *Options {
				opts := getDefaultOptions()
				opts.UseDebugLogLevel = true
				return opts
			},
		},
		{
			name: "cert-dir flag",
			args: []string{"-cert-dir", "/custom/path"},
			expected: func() *Options {
				opts := getDefaultOptions()
				opts.CertDir = "/custom/path"
				return opts
			},
		},
		{
			name: "webhook-port flag",
			args: []string{"-webhook-port", "9999"},
			expected: func() *Options {
				opts := getDefaultOptions()
				opts.WebhookPort = 9999
				return opts
			},
		},
		{
			name: "leader-elect flag",
			args: []string{"-leader-elect"},
			expected: func() *Options {
				opts := getDefaultOptions()
				opts.EnableLeaderElection = true
				return opts
			},
		},
		{
			name: "legacy-project-permissions false",
			args: []string{"-legacy-project-permissions=false"},
			expected: func() *Options {
				opts := getDefaultOptions()
				opts.LegacyProjectPermissions = false
				return opts
			},
		},
		{
			name: "legacy-project-permissions-interval-seconds",
			args: []string{"-legacy-project-permissions-interval-seconds", "120"},
			expected: func() *Options {
				opts := getDefaultOptions()
				opts.LegacyProjectPermissionsIntervalSeconds = 120
				return opts
			},
		},
		{
			name: "allowed-service-accounts single",
			args: []string{"-allowed-service-accounts", "single-account"},
			expected: func() *Options {
				opts := getDefaultOptions()
				opts.AllowedServiceAccounts = options.ArrayFlags{"single-account"}
				return opts
			},
		},
		{
			name: "allowed-service-accounts multiple",
			args: []string{"-allowed-service-accounts", "account1,account2,account3"},
			expected: func() *Options {
				opts := getDefaultOptions()
				opts.AllowedServiceAccounts = options.ArrayFlags{"account1", "account2", "account3"}
				return opts
			},
		},
		{
			name: "allowed-service-accounts with spaces",
			args: []string{"-allowed-service-accounts", "account1, account2 , account3"},
			expected: func() *Options {
				opts := getDefaultOptions()
				opts.AllowedServiceAccounts = options.ArrayFlags{"account1", "account2", "account3"}
				return opts
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &Options{}
			flagSet := flag.NewFlagSet("test", flag.ContinueOnError)

			err := opts.parseCommandLineArgs(flagSet, tt.args)
			if err != nil {
				t.Fatalf("parseCommandLineArgs() failed: %v", err)
			}

			expected := tt.expected()
			if !reflect.DeepEqual(opts, expected) {
				t.Errorf("parseCommandLineArgs() with %s = %+v, want %+v", tt.name, opts, expected)
			}
		})
	}
}

func TestParseCommandLineArgs_InvalidFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "invalid webhook port",
			args: []string{"-webhook-port", "invalid"},
		},
		{
			name: "invalid interval seconds",
			args: []string{"-legacy-project-permissions-interval-seconds", "invalid"},
		},
		{
			name: "unknown flag",
			args: []string{"-unknown-flag", "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &Options{}
			flagSet := flag.NewFlagSet("test", flag.ContinueOnError)

			err := opts.parseCommandLineArgs(flagSet, tt.args)
			if err == nil {
				t.Errorf("parseCommandLineArgs() with %s should have failed but didn't", tt.name)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    *Options
		wantErr bool
	}{
		{
			name:    "successful parse with defaults",
			args:    []string{},
			want:    getDefaultOptions(),
			wantErr: false,
		},
		{
			name: "successful parse with some flags",
			args: []string{"-debug", "-webhook-port", "9000"},
			want: func() *Options {
				opts := getDefaultOptions()
				opts.UseDebugLogLevel = true
				opts.WebhookPort = 9000
				return opts
			}(),
			wantErr: false,
		},
		{
			name:    "parse error with invalid flag",
			args:    []string{"-webhook-port", "invalid"},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
			got, err := parse(flagSet, tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parse() = %+v, want %+v", got, tt.want)
			}

			if tt.wantErr && got != nil {
				t.Errorf("parse() should return nil on error, got %+v", got)
			}
		})
	}
}

func TestOptions_FieldTypes(t *testing.T) {
	opts := &Options{}

	// Verify field types are as expected
	val := reflect.ValueOf(opts).Elem()
	typ := reflect.TypeOf(opts).Elem()

	expectedTypes := map[string]string{
		"UseDebugLogLevel":                        "bool",
		"CertDir":                                 "string",
		"WebhookPort":                             "int",
		"EnableLeaderElection":                    "bool",
		"LegacyProjectPermissions":                "bool",
		"LegacyProjectPermissionsIntervalSeconds": "int",
		"AllowedServiceAccounts":                  "options.ArrayFlags",
	}

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldType := field.Type.String()
		expectedType, exists := expectedTypes[field.Name]

		if !exists {
			t.Errorf("Unexpected field %s found in Options struct", field.Name)
			continue
		}

		if fieldType != expectedType {
			t.Errorf("Field %s has type %s, expected %s", field.Name, fieldType, expectedType)
		}
	}

	// Verify all expected fields exist
	for expectedField := range expectedTypes {
		if _, found := typ.FieldByName(expectedField); !found {
			t.Errorf("Expected field %s not found in Options struct", expectedField)
		}
	}
}

// Helper function to get default options for testing
func getDefaultOptions() *Options {
	return &Options{
		UseDebugLogLevel:                        false,
		CertDir:                                 "/etc/webhook/certs",
		WebhookPort:                             8443,
		EnableLeaderElection:                    false,
		LegacyProjectPermissions:                true,
		LegacyProjectPermissionsIntervalSeconds: 30,
		AllowedServiceAccounts:                  nil,
	}
}
