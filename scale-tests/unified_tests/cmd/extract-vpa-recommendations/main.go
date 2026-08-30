package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	resource_rec "github.com/run-ai/runai/scale-tests/unified_tests/internal/resource-rec"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func main() {
	ctrl.SetLogger(zap.New())

	if err := vpav1.AddToScheme(scheme.Scheme); err != nil {
		fmt.Fprintf(os.Stderr, "failed to register VPA scheme: %v\n", err)
		os.Exit(1)
	}

	namespaces := resource_rec.DefaultNamespaces
	if v := os.Getenv("LIMIT_NAMESPACES"); v != "" {
		namespaces = strings.Split(v, ",")
	}

	outputDir := os.Getenv("OUTPUT_DIR")
	if outputDir == "" {
		outputDir = "output"
	}

	cfg := ctrl.GetConfigOrDie()
	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create k8s client: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("namespaces: %v\n", namespaces)
	fmt.Printf("output dir: %s\n", outputDir)

	if err := resource_rec.ExtractVPARecommendations(context.Background(), k8sClient, namespaces, outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
