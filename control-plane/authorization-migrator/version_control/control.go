package version_control

import (
	"fmt"

	"github.com/run-ai/runai/control-plane/authorization-migrator/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
)

type MigrationControl struct {
	Version string `gorm:"primaryKey"`
}

var mc *MigrationControl

func Init(conf *config.ConfigData, client orm.DB) {
	if err := client.Actual().AutoMigrate(&MigrationControl{}); err != nil {
		logging.Errorf("Failed to auto-migrate migration control table: %v", err)
	}
	mc = &MigrationControl{
		Version: conf.MigrationVersion,
	}
}

func IsMigrationDone(client orm.DB) bool {
	if err := client.First(mc); err != nil {
		logging.Infof(fmt.Sprintf("Failed to get version %s", mc.Version), err)
		return false
	}
	return true
}

func MarkMigrationDone(client orm.DB) {
	if _, err := client.Create(mc); err != nil {
		logging.Errorf(fmt.Sprintf("Failed to save version %s", mc.Version), err)
	}
}
