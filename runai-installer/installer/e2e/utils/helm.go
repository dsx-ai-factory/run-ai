package utils

import (
	"os"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

func GetHelmChartVersion(namespace, releaseName string) (string, error) {
	actionCfg := &action.Configuration{}
	settings := cli.New()
	settings.SetNamespace(namespace)
	if err := actionCfg.Init(settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), nil); err != nil {
		return "", err
	}
	status := action.NewStatus(actionCfg)
	rel, err := status.Run(releaseName)
	if err != nil {
		return "", err
	}
	return rel.Chart.Metadata.Version, nil
}
