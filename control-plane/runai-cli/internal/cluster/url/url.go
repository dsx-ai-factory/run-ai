package url

import (
	"fmt"
	"net/url"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/apis"
)

const Template = "/api/v1/%s/workloads/%s/%s/%s/%s" // /api/v1/{projectName}/workloads/{workloadType}/{workloadFramework}/{workloadName}
const External = "external"

func BuildClusterUrl(baseUrl string, workload *apis.Workload, action string, query *url.Values) (*url.URL, error) {
	base, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	workloadType := workload.Type.ClusterApiType()
	workloadFramework := workload.Framework.ClusterApiFramework()

	// Special handling for external workloads (identified by UndefinedFramework)
	// External workloads use "external" for both type and framework
	if workload.Framework == api.NimService || workload.Framework == api.UndefinedFramework {
		workloadType = External
		workloadFramework = External
	}

	endpoint := fmt.Sprintf(Template, workload.Project, workloadType, workloadFramework, workload.Name, action)
	newUrl := base.JoinPath(endpoint)
	if query != nil {
		newUrl.RawQuery = query.Encode()
	}
	return newUrl, nil
}
