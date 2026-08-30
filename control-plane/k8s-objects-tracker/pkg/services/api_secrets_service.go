package services

import (
	"context"
	"net/http"

	authorizationClient "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/database/models"
)

const (
	NamespaceRunai = "runai"
)

type SecretsService struct{}

func NewSecretsService() *SecretsService {
	return &SecretsService{}
}

func (ss *SecretsService) CreateClusterWideSecret(ctx context.Context, clusterId string, secretName string, request api.SecretCreationRequest) (api.ImplResponse, error) {
	secret, err := ss.CreateSecret(ctx, clusterId, authorizationClient.SCOPETYPE_CLUSTER, clusterId, secretName, request)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, *secret), nil
}

func (ss *SecretsService) CreateDepartmentWideSecret(ctx context.Context, clusterId string, departmentId string, secretName string, request api.SecretCreationRequest) (api.ImplResponse, error) {
	return ss.CreateDepartmentsWideSecret(ctx, clusterId, departmentId, secretName, request)
}

func (ss *SecretsService) CreateDepartmentsWideSecret(ctx context.Context, clusterId string, departmentId string, secretName string, request api.SecretCreationRequest) (api.ImplResponse, error) {
	secret, err := ss.CreateSecret(ctx, clusterId, authorizationClient.SCOPETYPE_DEPARTMENT, departmentId, secretName, request)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, *secret), nil
}

func (ss *SecretsService) CreateProjectSecret(ctx context.Context, clusterId string, namespace string, secretName string, request api.SecretCreationRequest) (api.ImplResponse, error) {
	return ss.CreateProjectsSecret(ctx, clusterId, namespace, secretName, request)
}

func (ss *SecretsService) CreateProjectsSecret(ctx context.Context, clusterId string, namespace string, secretName string, request api.SecretCreationRequest) (api.ImplResponse, error) {
	secret, err := ss.CreateSecret(ctx, clusterId, authorizationClient.SCOPETYPE_PROJECT, namespace, secretName, request)
	if err != nil {
		return NewErrorResponse(err), nil
	}
	return api.Response(http.StatusOK, *secret), nil
}

func (ss *SecretsService) SyncClusterWideSecrets(ctx context.Context, clusterId string, request api.Secrets) (api.ImplResponse, error) {
	secrets, err := ss.SyncSecrets(ctx, clusterId, authorizationClient.SCOPETYPE_CLUSTER, clusterId, request)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, *secrets), nil
}

func (ss *SecretsService) SyncDepartmentWideSecrets(ctx context.Context, clusterId string, departmentId string, request api.Secrets) (api.ImplResponse, error) {
	return ss.SyncDepartmentsWideSecrets(ctx, clusterId, departmentId, request)
}

func (ss *SecretsService) SyncDepartmentsWideSecrets(ctx context.Context, clusterId string, departmentId string, request api.Secrets) (api.ImplResponse, error) {
	secrets, err := ss.SyncSecrets(ctx, clusterId, authorizationClient.SCOPETYPE_DEPARTMENT, departmentId, request)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, *secrets), nil
}

func (ss *SecretsService) SyncProjectSecrets(ctx context.Context, clusterId string, namespace string, request api.Secrets) (api.ImplResponse, error) {
	return ss.SyncProjectsSecrets(ctx, clusterId, namespace, request)
}

func (ss *SecretsService) SyncProjectsSecrets(ctx context.Context, clusterId string, namespace string, request api.Secrets) (api.ImplResponse, error) {
	secrets, err := ss.SyncSecrets(ctx, clusterId, authorizationClient.SCOPETYPE_PROJECT, namespace, request)
	if err != nil {
		return NewErrorResponse(err), nil
	}
	return api.Response(http.StatusOK, *secrets), nil
}

func (ss *SecretsService) DeleteClusterWideSecret(ctx context.Context, clusterId string, name string) (api.ImplResponse, error) {
	err := ss.DeleteSecret(ctx, clusterId, authorizationClient.SCOPETYPE_CLUSTER, clusterId, name)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return NoContentResponse(), nil
}

func (ss *SecretsService) DeleteDepartmentWideSecret(ctx context.Context, clusterId string, departmentId string, name string) (api.ImplResponse, error) {
	return ss.DeleteDepartmentsWideSecret(ctx, clusterId, departmentId, name)
}

func (ss *SecretsService) DeleteDepartmentsWideSecret(ctx context.Context, clusterId string, departmentId string, name string) (api.ImplResponse, error) {
	err := ss.DeleteSecret(ctx, clusterId, authorizationClient.SCOPETYPE_DEPARTMENT, departmentId, name)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return NoContentResponse(), nil
}

func (ss *SecretsService) DeleteProjectSecret(ctx context.Context, clusterId string, namespace string, name string) (api.ImplResponse, error) {
	return ss.DeleteProjectsSecret(ctx, clusterId, namespace, name)
}

func (ss *SecretsService) DeleteProjectsSecret(ctx context.Context, clusterId string, namespace string, name string) (api.ImplResponse, error) {
	err := ss.DeleteSecret(ctx, clusterId, authorizationClient.SCOPETYPE_PROJECT, namespace, name)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return NoContentResponse(), nil
}

func (ss *SecretsService) ObsoleteCreateSecret(ctx context.Context, clusterId string, namespace string, request api.SingleSecret) (api.ImplResponse, error) {
	newRequest := api.SecretCreationRequest{
		Type: models.GenericSecretType,
	}

	var err *api.Error
	if namespace == NamespaceRunai {
		_, err = ss.CreateSecret(ctx, clusterId, authorizationClient.SCOPETYPE_CLUSTER, clusterId, request.SecretName, newRequest)
	} else {
		_, err = ss.CreateSecret(ctx, clusterId, authorizationClient.SCOPETYPE_PROJECT, namespace, request.SecretName, newRequest)
	}
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, request), nil
}

func (ss *SecretsService) ObsoleteDeleteSecret(ctx context.Context, clusterId string, namespace string, secretName string) (api.ImplResponse, error) {
	var err *api.Error

	if namespace == NamespaceRunai {
		err = ss.DeleteSecret(ctx, clusterId, authorizationClient.SCOPETYPE_CLUSTER, clusterId, secretName)
	} else {
		err = ss.DeleteSecret(ctx, clusterId, authorizationClient.SCOPETYPE_PROJECT, namespace, secretName)
	}
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, api.SingleSecret{
		SecretName: secretName,
	}), nil
}

func (ss *SecretsService) ObsoleteSyncSecret(ctx context.Context, clusterId string, namespace string, secrets []api.SingleSecret) (api.ImplResponse, error) {
	newRequest := api.Secrets{
		Secrets: make([]api.Secret, 0, len(secrets)),
	}

	for _, secret := range secrets {
		newRequest.Secrets = append(newRequest.Secrets, api.Secret{
			Name: secret.SecretName,
			Type: models.GenericSecretType,
		})
	}

	var err *api.Error

	if namespace == NamespaceRunai {
		_, err = ss.SyncSecrets(ctx, clusterId, authorizationClient.SCOPETYPE_CLUSTER, clusterId, newRequest)
	} else {
		_, err = ss.SyncSecrets(ctx, clusterId, authorizationClient.SCOPETYPE_PROJECT, namespace, newRequest)
	}
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, secrets), nil
}
