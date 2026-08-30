/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"os"

	"github.com/run-ai/runai/runai-cluster/cluster/runaijob-controller/pkg/config"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"

	"go.uber.org/zap"
	"k8s.io/client-go/tools/clientcmd"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	"github.com/run-ai/runai/runai-cluster/cluster/runaijob-controller/pkg/runaijobcontroller"

	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	// runaijobinformers "github.com/run-ai/runai/runai-cluster/cluster/sdk/client/informers/externalversions"
	runaijobinformers "github.com/run-ai/runai/runai-cluster/cluster/sdk/client/informers/externalversions"

	clientsetrunai "github.com/run-ai/runai/runai-cluster/cluster/sdk/client/clientset/versioned"

	"github.com/run-ai/runai/runai-cluster/cluster/runaijob-controller/pkg/signals"

	_ "github.com/run-ai/runai/runai-common-packages/fips"
)

const (
	defaultPromServiceUrl = "http://prometheus-operated.runai.svc:9090"
)

func main() {
	var startPromClient bool
	var promServiceUrl string

	klog.Infof("job controller started (1)")

	flag.BoolVar(&startPromClient, "start-prom-client", true,
		"Defines rather prom client should be started or not.")
	flag.StringVar(&promServiceUrl, "prom-service-url", defaultPromServiceUrl,
		"Prometheus service url to use")

	klog.InitFlags(flag.CommandLine)
	flag.Parse()

	// set up signals, so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	cfg, err := rest.InClusterConfig()
	if err != nil {
		kubeConfigPath := os.Getenv("KUBECONFIG")
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			klog.Fatalf("Error building kubeconfig: %s", err.Error())
		} else {
			zap.S().Debugf("Got kube config from  path %s", kubeConfigPath)
		}
	}

	cfg.QPS = config.GetK8sClientQPS()
	cfg.Burst = config.GetK8sClientBurst()

	klog.Info("client throttling applied", "QPS", cfg.QPS, "burst", cfg.Burst)

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	runaijobClient, err := clientsetrunai.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building example clientset: %s", err.Error())
	}
	runaiJobAvailable, err := checkCRD(runaijobClient.RunV1().RunaiJobs(""))
	if err != nil {
		klog.Fatalf("Failed checking RunaiJob API availability: %s", err.Error())
	}
	if !runaiJobAvailable {
		klog.Warning("RunaiJob API is not served; controller will remain idle until the pod is restarted")
		<-stopCh
		return
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, 0)
	runaiInformerFactory := runaijobinformers.NewSharedInformerFactory(runaijobClient, 0)

	klog.V(3).Infof("Initializing Runai Job Controller")
	controller := runaijobcontroller.NewController(
		cfg,
		kubeInformerFactory,
		runaiInformerFactory,
		kubeClient,
		runaijobClient,
	)

	// notice that there is no need to run Start methods in a separate goroutine. (i.e. go kubeInformerFactory.Start(stopCh)
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	kubeInformerFactory.Start(stopCh)
	runaiInformerFactory.Start(stopCh)

	controller.Run(5, stopCh)
}

type runaiJobLister interface {
	List(context.Context, metav1.ListOptions) (*runv1.RunaiJobList, error)
}

func checkCRD(lister runaiJobLister) (bool, error) {
	_, err := lister.List(context.TODO(), metav1.ListOptions{})
	if err == nil {
		return true, nil
	}
	if apierrors.IsNotFound(err) {
		return false, nil
	}
	return false, err
}
