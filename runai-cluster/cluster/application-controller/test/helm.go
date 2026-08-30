package test

import (
	"context"
	"io"

	"github.com/run-ai/runai/runai-cluster/cluster/application-controller/pkg/deployers/helm_deployer"
	runv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	helmFake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewTestHelmStore() *storage.Storage {
	return storage.Init(driver.NewMemory())
}

func NewTestReleaseClient(id types.NamespacedName, store *storage.Storage, downloaderFactory helm_deployer.DownloaderFactory) *helm_deployer.ReleaseClient {
	cfg := &action.Configuration{
		Releases:     store,
		KubeClient:   &helmFake.FailingKubeClient{PrintingKubeClient: helmFake.PrintingKubeClient{Out: io.Discard}},
		Capabilities: chartutil.DefaultCapabilities,
		Log:          func(format string, v ...interface{}) {},
	}
	return &helm_deployer.ReleaseClient{
		ReleaseId:         id,
		Settings:          cli.New(),
		ActionConfig:      cfg,
		DownloaderFactory: downloaderFactory,
	}
}

// FailingGetClient is a client that fails any "get" for an application.
type FailingGetClient struct {
	client.Client
	getErr error
}

func NewFailingGetClient(c client.Client, err error) *FailingGetClient {
	return &FailingGetClient{Client: c, getErr: err}
}

func (f *FailingGetClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if _, ok := obj.(*runv1alpha1.Application); ok {
		return f.getErr
	}
	return f.Client.Get(ctx, key, obj, opts...)
}
