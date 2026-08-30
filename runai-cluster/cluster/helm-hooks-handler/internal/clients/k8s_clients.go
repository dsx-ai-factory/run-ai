package clients

import (
	"fmt"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	scheme_init "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/installation/util"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Clients struct {
	LocalClusterClient        client.Client
	LocalClusterDynamicClient dynamic.Interface

	RemoteClusterClient        client.Client
	RemoteClusterDynamicClient dynamic.Interface
}

func InitClients(opts *util.Options) (*Clients, error) {
	// The kai CRD (kai.resources/v1alpha1) is not part of the shared
	// scheme but is needed by the resource-management migration, so register it
	// on the scheme before the clients are built.
	scheme := scheme_init.Scheme()
	if err := kaiv1alpha1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to register kai scheme: %v", err)
	}

	k8sConfig := ctrl.GetConfigOrDie()
	k8s, err := client.New(k8sConfig, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("failed building kubernetes client: %v", err)
	}
	dynamicClient := dynamic.NewForConfigOrDie(k8sConfig)
	var remoteClusterClient client.Client
	var remoteClusterDynamicClient dynamic.Interface

	if opts.NMCRemoteClusterURL != "" {
		remoteClusterConfig := ctrl.GetConfigOrDie()
		remoteClusterConfig.Host = opts.NMCRemoteClusterURL
		remoteClusterClient, err = client.New(remoteClusterConfig, client.Options{Scheme: scheme})
		if err != nil {
			return nil, fmt.Errorf("failed building kubernetes client: %v", err)
		}
		remoteClusterDynamicClient = dynamic.NewForConfigOrDie(remoteClusterConfig)
	} else {
		remoteClusterClient = k8s
		remoteClusterDynamicClient = dynamicClient
	}

	return &Clients{LocalClusterClient: k8s,
		LocalClusterDynamicClient:  dynamicClient,
		RemoteClusterClient:        remoteClusterClient,
		RemoteClusterDynamicClient: remoteClusterDynamicClient,
	}, nil
}
