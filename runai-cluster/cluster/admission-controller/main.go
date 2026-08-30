package main

import (
	"context"
	"log"

	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/server"

	_ "github.com/run-ai/runai/runai-common-packages/fips"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
	}()

	server := server.NewServer()
	err := server.Start(ctx)
	if err != nil {
		log.Printf("Failed to start: %v", err)
	}
}
