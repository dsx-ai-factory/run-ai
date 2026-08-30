package submit

import (
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

type CreationOptions struct {
	ClusterId         string
	ProjectId         string
	Name              string
	ChartURL          string
	Values            map[string]interface{}
	CredentialAssetId string
}

func BuildCreationRequest(options CreationOptions) openapi.AIApplicationCreationRequest {
	spec := openapi.AIApplicationSpec{
		ChartRef: openapi.AIApplicationSpecChartRef{
			ChartURL: &options.ChartURL,
		},
		Values: options.Values,
	}

	if options.CredentialAssetId != "" {
		spec.Auth = &openapi.AIApplicationAuth{AssetId: &options.CredentialAssetId}
	}

	return openapi.AIApplicationCreationRequest{
		ClusterId: options.ClusterId,
		ProjectId: options.ProjectId,
		Name:      options.Name,
		Spec:      spec,
	}
}
