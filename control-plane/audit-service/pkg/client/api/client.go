package api

import (
	"context"

	openapi "github.com/run-ai/runai/control-plane/audit-service/pkg/client/generated"
)

type AuditClient struct {
	ApiClient *openapi.APIClient
}

func (a *AuditClient) GetAuditLogs(ctx context.Context) (*openapi.GetAuditLogs200Response, error) {
	auditLogsResponse, _, err := a.ApiClient.AuditLogsAPI.GetAuditLogs(ctx).Execute()
	if err != nil {
		return nil, err
	}
	return auditLogsResponse, nil
}

func (a *AuditClient) DownloadAuditLogs(ctx context.Context) error {
	_, err := a.ApiClient.AuditLogsAPI.DownloadAuditLogs(ctx).Execute()
	return err
}

func (a *AuditClient) SaveAuditLog(ctx context.Context, auditLog openapi.AuditLog) error {
	_, err := a.ApiClient.InternalAuditLogsAPI.SaveAuditLog(ctx).AuditLog(auditLog).Execute()
	return err
}
