/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"context"
	"os"
	"time"

	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/spf13/pflag"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/init-ca/pkg/initca"

	_ "github.com/dsx-ai-factory/run-ai/runai-common-packages/fips"
)

var (
	setupLog = ctrl.Log.WithName("setup")
)

// 30 days
const certRotationInterval = 30 * 24 * time.Hour

type initCAOptions struct {
	kubeconfig           string
	webhooksPath         string
	enableLeaderElection bool
}

func (o *initCAOptions) addFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	fs.StringVar(&o.webhooksPath, "webhooks", "/webhooks/webhooks.yaml", "Path to the webhooks configuration file")
	fs.BoolVar(&o.enableLeaderElection, "leader-elect", false,
		"Enable leader election for init-ca. Enabling this will ensure there is only one active init-ca replica.")
}

func getKubeConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	return rest.InClusterConfig()
}

func main() {
	var opts initCAOptions
	opts.addFlags(pflag.CommandLine)
	pflag.Parse()
	// create kube rest configuration in cluster or from kubeconfig (according to flag)
	conf, err := getKubeConfig(opts.kubeconfig)
	if err != nil {
		panic(err)
	}
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zap.Options{
		Level: zapcore.InfoLevel,
	})))
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:           runtime.NewScheme(),
		LeaderElection:   opts.enableLeaderElection,
		LeaderElectionID: "init-ca.run.ai",
	})
	if err != nil {
		setupLog.Error(err, "Failed to configure manager")
		os.Exit(1)
	}

	managerContext := ctrl.SetupSignalHandler()
	go func() {
		<-mgr.Elected()
		setupLog.Info("Leader elected")
		if err := runInitCA(managerContext, opts, conf); err != nil {
			setupLog.Error(err, "Failed to run init-ca")
			os.Exit(1)
		}
	}()

	if err := mgr.Start(managerContext); err != nil {
		setupLog.Error(err, "Manager stopped")
		os.Exit(1)
	}
}

func runInitCA(ctx context.Context, opts initCAOptions, conf *rest.Config) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			err := initca.InitCA(ctx, conf, opts.webhooksPath)
			if err != nil {
				return err
			}

			// TODO make this duration configurable
			time.Sleep(certRotationInterval)
		}
	}
}
