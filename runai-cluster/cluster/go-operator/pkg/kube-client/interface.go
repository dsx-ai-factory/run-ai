package kube_client

import (
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// HybridClient represents a kubernetes client that can utilize cached and uncached clients
type HybridClient interface {
	client.Client
	Uncached() client.Client
	NewDiscoveryClient() (discovery.DiscoveryInterface, error)
}
