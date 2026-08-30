package async_events

type TenantDelete struct {
	TenantId string `json:"tenant_id"`
}

func NewTenantDelete(tenantId string) *TenantDelete {
	return &TenantDelete{
		TenantId: tenantId,
	}
}
