/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"flag"
	"github.com/run-ai/runai/runai-cluster/cluster/common/utils/options"
	"os"
)

type Options struct {
	UseDebugLogLevel     bool
	CertDir              string
	WebhookPort          int
	EnableLeaderElection bool

	LegacyProjectPermissions                bool
	LegacyProjectPermissionsIntervalSeconds int
	AllowedServiceAccounts                  options.ArrayFlags
}

func SetOptions() (*Options, error) {
	return parse(flag.CommandLine, os.Args[1:])
}

func parse(flagSet *flag.FlagSet, options []string) (*Options, error) {
	opts := &Options{}
	if err := opts.parseCommandLineArgs(flagSet, options); err != nil {
		return nil, err
	}
	return opts, nil
}

func (opts *Options) parseCommandLineArgs(flagSet *flag.FlagSet, options []string) error {
	flagSet.BoolVar(&opts.UseDebugLogLevel, "debug", false, "Should use debug log level")
	flagSet.StringVar(&opts.CertDir, "cert-dir", "/etc/webhook/certs", "Directory that contains the server key and certificate for the webhooks")
	flagSet.IntVar(&opts.WebhookPort, "webhook-port", 8443, "The port to serve the webhook from")
	flagSet.BoolVar(&opts.EnableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flagSet.BoolVar(&opts.LegacyProjectPermissions, "legacy-project-permissions", true, "Should handle legacy project permissions")
	flagSet.IntVar(&opts.LegacyProjectPermissionsIntervalSeconds, "legacy-project-permissions-interval-seconds", 30, "Interval in seconds to check for legacy project permissions")
	flagSet.Var(&opts.AllowedServiceAccounts, "allowed-service-accounts", "Comma-separated list of service account name patterns that should be allowed to perform access rule operations")
	return flagSet.Parse(options)
}
