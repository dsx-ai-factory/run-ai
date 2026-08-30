package async_events

type AssetDeleted struct {
	TenantId  string `json:"tenant_id"`
	AssetId   string `json:"asset_id"`
	Kind      string `json:"kind"`
	ScopeType string `json:"scope_type"`
	ScopeId   string `json:"scope_id"`
}

func NewAssetDeleted(
	tenantId string,
	assetId string,
	kind string,
	scopeType string,
	scopeId string,
) *AssetDeleted {
	return &AssetDeleted{
		TenantId:  tenantId,
		AssetId:   assetId,
		Kind:      kind,
		ScopeType: scopeType,
		ScopeId:   scopeId,
	}
}
