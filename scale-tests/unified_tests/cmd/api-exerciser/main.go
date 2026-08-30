package main

import (
	"context"
	"log"

	api_exerciser "github.com/run-ai/runai/scale-tests/unified_tests/internal/api-exerciser"
	"github.com/run-ai/runai/scale-tests/unified_tests/internal/sampler_actions"
	"github.com/run-ai/runai/scale-tests/unified_tests/internal/scale"
)

func main() {
	ctx := context.Background()

	clusterConfig := scale.ConstructConfig()
	exerciserConfig := api_exerciser.ConstructConfig()

	kubeConfig := sampler_actions.InitKubeConfig(clusterConfig)
	clients, err := sampler_actions.InitClients(ctx, clusterConfig, kubeConfig)
	if err != nil {
		log.Fatalf("failed to initialize clients: %v", err)
	}

	api_exerciser.NewExerciser(clients, exerciserConfig, clusterConfig).Run(ctx)
}
