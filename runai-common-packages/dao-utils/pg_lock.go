package dao_utils

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

const DefaultLockRetires = 3

type PGLock struct {
	schema string
}

func NewPGLock(schema string) *PGLock {
	return &PGLock{schema: schema}
}

func (l *PGLock) TryAcquireSchemaLock(db *gorm.DB, retryAttempts int) (bool, error) {
	err := l.createSchema(db)
	if err != nil {
		return false, err
	}

	if retryAttempts <= 0 {
		retryAttempts = DefaultLockRetires
	}

	var acquired bool
	for i := 0; i < retryAttempts; i++ {
		acquired, err = l.acquireSchemaLock(db)
		if err == nil {
			return acquired, err
		}
		time.Sleep(500 * time.Millisecond)
	}

	return false, err
}

func (l *PGLock) AcquireSchemaLock(db *gorm.DB) (bool, error) {
	err := l.createSchema(db)
	if err != nil {
		return false, err
	}

	return l.acquireSchemaLock(db)
}

func (l *PGLock) ReleaseSchemaLock(db *gorm.DB) (bool, error) {
	var released *bool
	query := fmt.Sprintf("SELECT pg_advisory_unlock(to_regnamespace('%s')::integer)", l.schema)
	err := db.Raw(query).Scan(&released)

	return released == nil || *released, err.Error
}

func (l *PGLock) createSchema(db *gorm.DB) error {
	validateSchemaExistsQuery := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS \"%s\"", l.schema)
	result := db.Exec(validateSchemaExistsQuery)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (l *PGLock) acquireSchemaLock(db *gorm.DB) (bool, error) {
	var acquired bool
	query := fmt.Sprintf("SELECT pg_try_advisory_lock(to_regnamespace('%s')::integer)", l.schema)
	result := db.Raw(query).Scan(&acquired)
	return acquired, result.Error
}
