package clients

import (
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/config"
	"github.com/run-ai/runai/control-plane/notifications-service/pkg/client/generated/config_client"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"net/url"
)

func GetConfigClient() *config_client.APIClient {
	conf := runai_config.Get[*config.ServiceConfig]()
	host, scheme := getHostAndScheme(conf.NotifyConfigUrl)
	cfgServer := config_client.NewConfiguration()
	cfgServer.Host = host
	cfgServer.Scheme = scheme

	return config_client.NewAPIClient(cfgServer)
}

func getHostAndScheme(fullUrl string) (string, string) {
	pUrl, err := url.Parse(fullUrl)
	if err != nil {
		return "", ""
	}

	return pUrl.Host, pUrl.Scheme
}
