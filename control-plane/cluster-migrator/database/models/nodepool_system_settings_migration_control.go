package models

type NodepoolSystemSettingsMigrationControl struct {
	Version  string `gorm:"uniqueIndex:np_v_tenant"`
	TenantId *int32 `gorm:"uniqueIndex:np_v_tenant"`
	Status   MigrationStatus
}

func (NodepoolSystemSettingsMigrationControl) TableName() string {
	return "nodepool_system_settings_migration_control"
}
