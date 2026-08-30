package project

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/common/getters"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type NamespaceTranslator struct {
	k8sClient client.Client
}

func NewNamespaceTranslator(k8sClient client.Client) *NamespaceTranslator {
	translator := &NamespaceTranslator{
		k8sClient: k8sClient,
	}

	return translator
}

func (nt *NamespaceTranslator) Translate(ctx context.Context, projectName string) (string, error) {
	log.Debug().Str("projectName", projectName).Msg("Starting translation")

	namespace, err := getters.GetNamespaceByProject(ctx, projectName, nt.k8sClient)
	if err != nil {
		log.Error().Err(err).Str("projectName", projectName).Msg("Failed to get namespace by project")
		return "", err
	}

	log.Info().
		Str("projectName", projectName).
		Str("namespaceName", namespace.Name).
		Msg("Successfully translated project to namespace")

	return namespace.Name, nil
}
