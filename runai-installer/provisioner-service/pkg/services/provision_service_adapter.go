package services

import (
	"context"
	"net/http"

	"github.com/run-ai/runai/runai-installer/provisioner-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/runai-installer/provisioner-service/pkg/types"
)

// ProvisionServiceAdapter adapts ProvisionerService to implement ProvisionAPIServicer
type ProvisionServiceAdapter struct {
	service *ProvisionerService
}

// NewProvisionServiceAdapter creates a new adapter
func NewProvisionServiceAdapter(service *ProvisionerService) *ProvisionServiceAdapter {
	return &ProvisionServiceAdapter{
		service: service,
	}
}

// CreateEnvironment implements ProvisionAPIServicer.CreateEnvironment
func (a *ProvisionServiceAdapter) CreateEnvironment(ctx context.Context, req api.CreateEnvironmentRequest) (api.ImplResponse, error) {
	if err := validateCreateEnvironmentRequest(req); err != nil {
		return api.Response(http.StatusBadRequest, nil), err
	}

	var k8sDist *string
	if req.KubernetesDistribution != "" {
		k8sDist = &req.KubernetesDistribution
	}

	var userEmail *string
	if req.UserEmail != "" {
		userEmail = &req.UserEmail
	}

	var projectName *string
	if req.ProjectName != "" {
		projectName = &req.ProjectName
	}

	id, createdAt, err := a.service.CreateEnvironment(types.ProvisioningMode(req.ProvisioningMode), req.Domain, req.Wildcarddomain, k8sDist, userEmail, projectName)
	if err != nil {
		return api.Response(http.StatusBadRequest, nil), err
	}

	response := api.EnvironmentAccepted{
		Id:        string(id),
		Phase:     string(types.EnvironmentPhasePending),
		Mode:      req.ProvisioningMode,
		CreatedAt: createdAt,
	}

	return api.Response(http.StatusAccepted, response), nil
}

// GetEnvironment implements ProvisionAPIServicer.GetEnvironment
func (a *ProvisionServiceAdapter) GetEnvironment(ctx context.Context, id string) (api.ImplResponse, error) {
	status, exists := a.service.GetEnvironment(id)
	if !exists {
		return api.Response(http.StatusNotFound, nil), nil
	}

	response := api.EnvironmentStatus{
		Id:        string(status.ID),
		Phase:     string(status.Phase),
		Message:   status.Message,
		UpdatedAt: status.UpdatedAt,
		Mode:      string(status.Mode),
	}
	return api.Response(http.StatusOK, response), nil
}
