package application

import (
	"fmt"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/filter"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/list"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/application/client"
	applist "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/application/list"
	credclient "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/credential/client"
	credlist "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/credential/list"
	projectservice "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/project"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
	assetsopenapi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

const (
	errCredentialNotFound  = "credential %q not found"
	errCredentialAmbiguous = "multiple credentials named %q found"

	// the only credential kind the product supports for authenticating
	// private Helm chart URLs
	chartAuthCredentialKind = string(assetsopenapi.ASSETKIND_NGC_API_KEY)
)

var (
	lister           applist.AIApplicationLister
	deleter          applist.AIApplicationDeleter
	creator          applist.AIApplicationCreator
	credentialLister credlist.Lister
)

func SetLister(l applist.AIApplicationLister) {
	lister = l
}

func SetDeleter(d applist.AIApplicationDeleter) {
	deleter = d
}

func SetCreator(c applist.AIApplicationCreator) {
	creator = c
}

func SetCredentialLister(cl credlist.Lister) {
	credentialLister = cl
}

func getAIApplicationListService(domain, token string) (*applist.AIApplicationService, error) {
	// Use mock if injected (for testing)
	if lister != nil {
		return applist.NewAIApplicationService(lister), nil
	}

	// Create real client
	aiApplicationClient, err := client.NewAIApplicationClient(domain, token)
	if err != nil {
		return nil, err
	}

	return applist.NewAIApplicationService(aiApplicationClient), nil
}

func GetAIApplicationList(txn *transaction.Txn, domain, token string, pageOptions *list.PageOptions, fb *filter.Builder) (*applist.Output, error) {
	svc, err := getAIApplicationListService(domain, token)
	if err != nil {
		return nil, err
	}
	return svc.List(txn, pageOptions, fb)
}

func getAIApplicationDeleter(domain, token string) (applist.AIApplicationDeleter, error) {
	// Use mock if injected (for testing)
	if deleter != nil {
		return deleter, nil
	}

	// Create real client
	return client.NewAIApplicationClient(domain, token)
}

// GetAIApplicationByName resolves an application by exact name within a
// project on the given cluster. Returns nil (without an error) when no
// application matches.
func GetAIApplicationByName(txn *transaction.Txn, domain, token, clusterId, projectName, appName string) (*openapi.AIApplication, error) {
	svc, err := getAIApplicationListService(domain, token)
	if err != nil {
		return nil, err
	}

	fb := filter.NewBuilder().WithClusterID(clusterId).WithProject(projectName).AppendName(appName)
	return svc.GetByName(txn, fb, appName)
}

func DeleteAIApplication(txn *transaction.Txn, domain, token, aiApplicationId string) error {
	d, err := getAIApplicationDeleter(domain, token)
	if err != nil {
		return err
	}
	return d.Delete(txn.GetCtx(), aiApplicationId)
}

func getAIApplicationCreator(domain, token string) (applist.AIApplicationCreator, error) {
	// Use mock if injected (for testing)
	if creator != nil {
		return creator, nil
	}

	// Create real client
	return client.NewAIApplicationClient(domain, token)
}

func CreateAIApplication(txn *transaction.Txn, domain, token string, request openapi.AIApplicationCreationRequest) (*openapi.AIApplication, error) {
	c, err := getAIApplicationCreator(domain, token)
	if err != nil {
		return nil, err
	}
	return c.Create(txn.GetCtx(), request)
}

func getCredentialLister(domain, token string) (credlist.Lister, error) {
	// Use mock if injected (for testing)
	if credentialLister != nil {
		return credentialLister, nil
	}

	// Create real client
	return credclient.NewClient(domain, token)
}

// ResolveCredentialAssetId resolves an NGC API key credential asset name
// (scoped to the project and cluster) to its asset id, for authenticating
// private chart URLs. An exact-name match in more than one credential is
// rejected as ambiguous.
func ResolveCredentialAssetId(txn *transaction.Txn, domain, token, credentialName, projectId, clusterId string) (string, error) {
	cl, err := getCredentialLister(domain, token)
	if err != nil {
		return "", err
	}

	projectIdInt, err := projectservice.ParseProjectIdInt32(projectId)
	if err != nil {
		return "", err
	}

	credType := chartAuthCredentialKind
	response, err := cl.List(txn.GetCtx(), &credentialName, projectIdInt, &clusterId, &credType)
	if err != nil {
		return "", err
	}

	var matches []string
	if response != nil {
		for i := range response.Entries {
			entry := &response.Entries[i]
			if entry.Meta.Name != credentialName || string(entry.Meta.Kind) != chartAuthCredentialKind {
				continue
			}
			matches = append(matches, entry.Meta.Id)
		}
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf(errCredentialNotFound, credentialName)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf(errCredentialAmbiguous, credentialName)
	}
}
