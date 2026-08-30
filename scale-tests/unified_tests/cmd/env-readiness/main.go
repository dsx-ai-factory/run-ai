// env-readiness gates whether a provisioned environment has settled enough to start a
// measurement run against. Exits non-zero if any check fails, and always writes
// env-readiness-report.json to OUTPUT_DIR.
package main

import (
	"context"
	"log"
	"os"

	"github.com/run-ai/runai/scale-tests/unified_tests/internal/resource-rec/readiness"
	"github.com/run-ai/runai/scale-tests/unified_tests/internal/sampler_actions"
	"github.com/run-ai/runai/scale-tests/unified_tests/internal/scale"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

func main() {
	ctx := context.Background()

	if err := vpav1.AddToScheme(scheme.Scheme); err != nil {
		log.Fatalf("failed to register VPA scheme: %v", err)
	}

	clusterConfig := scale.ConstructConfig()
	cfg := readiness.ConstructConfig()

	kubeConfig := sampler_actions.InitKubeConfig(clusterConfig)
	clients, err := sampler_actions.InitClients(ctx, clusterConfig, kubeConfig)
	if err != nil {
		log.Fatalf("failed to initialize clients: %v", err)
	}

	if err := sampler_actions.SetClusterId(ctx, clients, clusterConfig); err != nil {
		log.Fatalf("failed to resolve cluster ID: %v", err)
	}
	log.Printf("cluster ID: %s", clusterConfig.ClusterID)

	report := readiness.NewChecker(cfg, clusterConfig, clients, kubeConfig).Run(ctx)
	report.Print()

	if err := report.Write(cfg.OutputDir); err != nil {
		log.Fatalf("failed to write readiness report: %v", err)
	}

	if !report.Ready {
		os.Exit(1)
	}
}
