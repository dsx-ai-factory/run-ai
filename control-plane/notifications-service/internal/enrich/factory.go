package enrich

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/enrich/enrichers"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type KubeReaderGetter interface {
	Get(ctx context.Context) ctrlclient.Reader
}

type EnricherFactory struct {
	kubeReaderGetter KubeReaderGetter
}

func NewEnricherFactory() *EnricherFactory {
	return &EnricherFactory{
		kubeReaderGetter: newKubeReaderGetter(),
	}
}

func (ef *EnricherFactory) CreateEnricher(ctx context.Context, name string, configProvider EnrichingConfigProvider) (Enricher, error) {
	switch name {
	case enrichers.KubeMetadataEnricherName:
		kubeClient := ef.kubeReaderGetter.Get(ctx)
		return enrichers.NewKubeMetadataEnricher(kubeClient, *configProvider.GetGlobalEnrichingConfig().KubeMetadataEnricherConfig), nil
	case enrichers.StaticEnricherName:
		return enrichers.NewStaticEnricher(configProvider.GetGlobalEnrichingConfig().StaticEnricherConfig)
	}

	return nil, fmt.Errorf("enricher %s not found in factory", name)
}

func newKubeReaderGetter() *kubeReaderGetter {
	return &kubeReaderGetter{}
}

type kubeReaderGetter struct{}

func (krg *kubeReaderGetter) Get(ctx context.Context) ctrlclient.Reader {
	kubeConfig := ctrl.GetConfigOrDie()
	cacheClient, err := ctrlcache.New(kubeConfig, ctrlcache.Options{})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create kubernetes cache client")
	}

	go func() {
		err := cacheClient.Start(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to start kubernetes cache client")
		}
	}()

	if !cacheClient.WaitForCacheSync(ctx) {
		log.Fatal().Msg("failed to sync kubernetes cache client")
	}

	return cacheClient
}
