package backend_db

import (
	"github.com/onsi/gomega"
	"gorm.io/gorm"
)

func CreateTables(db *gorm.DB) {
	err := db.AutoMigrate(&JobsRec{})
	gomega.Expect(err).To(gomega.BeNil())
}

func DeleteAll(gormDb *gorm.DB) {
	result := gormDb.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&JobsRec{})
	gomega.Expect(result.Error).To(gomega.BeNil())
}
