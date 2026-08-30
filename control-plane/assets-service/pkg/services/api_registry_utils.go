package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/encryption"
	log "github.com/sirupsen/logrus"
)

const (
	ValidationPath = "/v2/"
	CatalogPath    = "/v2/_catalog"
	TagsListPath   = "/v2/%s/tags/list"
)

func encryptPassword(ctx context.Context, spec *api.RegistrySpec) (int32, *api.Error) {
	tenantKeys, keysErr := encryption.GetTenantEncryptionKeys(ctx)
	if keysErr != nil {
		return 0, keysErr
	}

	passwordPtr, encryptErr := encryption.Encrypt(ctx, tenantKeys, &spec.Password, string(api.ASSETKIND_REGISTRY))
	if encryptErr != nil {
		return 0, encryptErr
	}
	spec.Password = *passwordPtr
	return tenantKeys.ActiveKeyId, nil
}

func (s *RegistryApiService) validateRegistryRequest(ctx context.Context, registrySpec api.RegistrySpec) *api.Error {
	registrySpec.Url = strings.TrimSuffix(registrySpec.Url, "/")
	validationUrl, err := url.Parse(registrySpec.Url + ValidationPath)
	if err != nil {
		return api.NewValidationError("registry url is invalid")
	}

	integrationValidationReq, err := http.NewRequestWithContext(ctx, http.MethodGet, validationUrl.String(), nil)
	if err != nil {
		return api.NewValidationError("failed to initiate the registry API validation")
	}

	apiErr := setAuthorizationHeader(integrationValidationReq, registrySpec)
	if apiErr != nil {
		return apiErr
	}

	response, err := s.httpClient.Do(integrationValidationReq)
	if err != nil {
		log.Errorf("failed to validate the registry API %+v", err)
		return api.NewValidationError("failed to validate the registry API")
	}
	if response.StatusCode != http.StatusOK {
		log.Errorf("registry integration validation failed with status code %d", response.StatusCode)
		return api.NewValidationError("registry integration validation failed")
	}

	return nil
}

func (s *RegistryApiService) getRepositories(ctx context.Context, registrySpec api.RegistrySpec, searchName string) (*api.Repositories, *api.Error) {
	catalogUrl, err := url.Parse(registrySpec.Url + CatalogPath)
	if err != nil {
		return nil, api.NewError(http.StatusBadRequest, "registry url is invalid", err)
	}

	catalogReq, err := http.NewRequestWithContext(ctx, http.MethodGet, catalogUrl.String(), nil)
	if err != nil {
		return nil, api.NewError(http.StatusBadRequest, "failed to initiate the registry catalog request", err)
	}

	apiErr := setAuthorizationHeader(catalogReq, registrySpec)
	if apiErr != nil {
		return nil, apiErr
	}

	response, err := s.httpClient.Do(catalogReq)
	if err != nil {
		log.Errorf("catalog request to %s returned with an error: %+v", catalogUrl.String(), err)
		return nil, api.NewError(http.StatusServiceUnavailable, "failed to list the registry catalog", err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, api.NewError(http.StatusBadRequest, fmt.Sprintf("registry catalog request returned with unsupported status code %d", response.StatusCode), nil)
	}
	defer response.Body.Close()

	var repositories api.Repositories
	err = json.NewDecoder(response.Body).Decode(&repositories)
	if err != nil {
		return nil, api.NewError(http.StatusServiceUnavailable, "failed to read the registry catalog response", err)
	}

	repositories.Repositories = FilterByNameContained(repositories.Repositories, searchName)

	return &repositories, nil
}

func (s *RegistryApiService) getRepositoryTags(ctx context.Context, registrySpec api.RegistrySpec, repository string, searchName string) (*api.RepositoryTags, *api.Error) {
	tagsPath := fmt.Sprintf(TagsListPath, repository)
	tagsUrl, err := url.Parse(registrySpec.Url + tagsPath)
	if err != nil {
		return nil, api.NewError(http.StatusBadRequest, "registry url is invalid", err)
	}

	tagsReq, err := http.NewRequestWithContext(ctx, http.MethodGet, tagsUrl.String(), nil)
	if err != nil {
		return nil, api.NewError(http.StatusBadRequest, "failed to initiate the repository tags request", err)
	}

	apiErr := setAuthorizationHeader(tagsReq, registrySpec)
	if apiErr != nil {
		return nil, apiErr
	}

	response, err := s.httpClient.Do(tagsReq)
	if err != nil {
		return nil, api.NewError(http.StatusServiceUnavailable, "failed to list the repository tags", err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, api.NewError(http.StatusBadRequest, fmt.Sprintf("repository tags request returned with unsupported status code %d", response.StatusCode), nil)
	}
	defer response.Body.Close()

	var tags api.RepositoryTags
	err = json.NewDecoder(response.Body).Decode(&tags)
	if err != nil {
		return nil, api.NewError(http.StatusServiceUnavailable, "failed to read the registry tags response", err)
	}

	tags.Tags = FilterByNameContained(tags.Tags, searchName)

	return &tags, nil
}

func setAuthorizationHeader(req *http.Request, spec api.RegistrySpec) *api.Error {
	switch spec.CredentialKind {
	case api.CREDENTIALKIND_PASSWORD:
		credentials := fmt.Sprintf("%s:%s", spec.User, spec.Password)
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(credentials)))
	default:
		return api.NewError(http.StatusNotImplemented, "credential type not supported", nil)
	}
	return nil
}

func FilterByNameContained(names []string, searchName string) []string {
	i := 0
	for _, val := range names {
		if strings.Contains(val, searchName) {
			names[i] = val
			i++
		}
	}

	return names[:i]
}
