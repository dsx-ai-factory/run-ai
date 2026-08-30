package models

type ClustersMigrationControl struct {
	Version string `gorm:"primaryKey"`
}

func (ClustersMigrationControl) TableName() string {
	return "migration_control"
}
