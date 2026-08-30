package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	resource_rec "github.com/run-ai/runai/scale-tests/unified_tests/internal/resource-rec"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const defaultTimeout = 10 * time.Minute

func main() {
	ctrl.SetLogger(zap.New())

	namespaces := resource_rec.DefaultNamespaces
	if v := os.Getenv("LIMIT_NAMESPACES"); v != "" {
		namespaces = strings.Split(v, ",")
	}

	waitTimeout := defaultTimeout
	if v := os.Getenv("REMOVE_LIMITS_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid REMOVE_LIMITS_TIMEOUT %q: %v\n", v, err)
			os.Exit(1)
		}
		waitTimeout = d
	}

	cfg := ctrl.GetConfigOrDie()
	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create k8s client: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("namespaces: %v\n", namespaces)
	fmt.Printf("wait timeout: %s\n", waitTimeout)

	if err := resource_rec.RemoveResourceLimits(context.Background(), k8sClient, namespaces, waitTimeout); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
