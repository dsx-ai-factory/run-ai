package dbtesting

import (
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/config"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/database/models"
	"github.com/run-ai/runai/runai-common-packages/integdb"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"gorm.io/gorm"
)

func CreateTables() func() {
	config.MockConfig()
	config.Init()

	conf, cleanup, err := integdb.SetupPostgresDB(&integdb.IntegdbConfig{
		Schema:           "k8s_objects_tracker",
		MigrationDirPath: "file://../database/migrations",
	})
	if err != nil {
		logging.Fatalf("Failed to set up postgres database for testing: %v", err)
	}

	conf.UseServiceSession = new(bool)
	if err := runaidb.Init(conf); err != nil {
		logging.Fatalf("Failed to initialize database for testing: %v", err)
	}

	return cleanup
}

func CleanTables() {
	db := runaidb.DB().Actual()
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Pvc{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Secret{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.ConfigMap{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Object{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.ResourceIssue{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.StorageClassConfiguration{})
}
