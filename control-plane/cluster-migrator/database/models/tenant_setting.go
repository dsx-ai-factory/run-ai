package models

type TenantSetting struct {
	TenantId int32
	Key      string
	Value    string
}

func (TenantSetting) TableName() string {
	return "tenant_settings"
}
