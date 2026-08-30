package configprovider

import (
	"errors"
	"fmt"

	"gorm.io/datatypes"

	"github.com/lib/pq"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/postgres/models"
	"gorm.io/gorm"
)

const sinkIdCol = "sink_id"

type sinksDao struct {
	db     *gorm.DB
	schema string
}

func newSinksDao(db *gorm.DB, schema string) *sinksDao {
	return &sinksDao{db: db, schema: schema}
}

func (psd *sinksDao) FetchSink(sinkId string) (*models.SinkConfig, error) {
	sinks, err := psd.FetchSinks(sinkId)
	if err != nil {
		return nil, err
	}

	if len(sinks) == 0 {
		return nil, nil
	}

	return sinks[0], nil
}

func (psd *sinksDao) FetchSinks(ids ...string) ([]*models.SinkConfig, error) {
	if len(ids) == 0 {
		return []*models.SinkConfig{}, nil
	}

	query := fmt.Sprintf("%s = ANY(?)", sinkIdCol)
	return fetch[models.SinkConfig](psd.db, psd.sinkTable(), query, pq.Array(ids))
}

func (psd *sinksDao) FetchAllSinks() ([]*models.SinkConfig, error) {
	return fetch[models.SinkConfig](psd.db, psd.sinkTable(), nil)
}

func (psd *sinksDao) UpsertSink(sinkId string, sinkType string, sink datatypes.JSON) error {
	return psd.db.Transaction(func(tx *gorm.DB) error {
		var existing models.SinkConfig
		err := tx.Table(psd.sinkTable()).
			Where(models.SinkConfig{SinkId: sinkId}).
			First(&existing).Error
		if err == nil {
			res := tx.Table(psd.sinkTable()).
				Where(models.SinkConfig{SinkId: sinkId}).
				Select("type", "config").
				Updates(&models.SinkConfig{Type: sinkType, Config: sink})
			return res.Error
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Table(psd.sinkTable()).
				Create(&models.SinkConfig{SinkId: sinkId, Type: sinkType, Config: sink}).Error
		}
		return err
	})
}

func (psd *sinksDao) CreateSink(sinkId string, sinkType string, sink datatypes.JSON) error {
	err := psd.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Table(psd.sinkTable()).
			Create(&models.SinkConfig{SinkId: sinkId, Type: sinkType, Config: sink})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return NewError(fmt.Sprintf("sink %s already exists", sinkId), ErrCodeDuplicate)
		}
		return nil
	})
	if err != nil {
		if isDuplicateKeyError(err) {
			return NewError(fmt.Sprintf("sink %s already exists", sinkId), ErrCodeDuplicate)
		}
		return fmt.Errorf("failed creating sink %q: %w", sinkId, err)
	}

	return nil
}

func (psd *sinksDao) UpdateSink(sinkId string, sinkType string, sink datatypes.JSON) error {
	return psd.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Table(psd.sinkTable()).
			Where(models.SinkConfig{SinkId: sinkId}).
			Select("type", "config").
			Updates(&models.SinkConfig{Type: sinkType, Config: sink})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return NewError(fmt.Sprintf("sink %s not found", sinkId), ErrCodeNotFound)
		}
		return nil
	})
}

func (psd *sinksDao) DeleteSink(sinkId string) error {
	return psd.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Table(psd.sinkTable()).
			Where(models.SinkConfig{SinkId: sinkId}).
			Unscoped().
			Delete(&models.SinkConfig{})
		if result.Error != nil {
			return NewError(result.Error.Error(), ErrCodeUnknown)
		}
		if result.RowsAffected == 0 {
			return NewError(fmt.Sprintf("sink %s not found", sinkId), ErrCodeNotFound)
		}
		return nil
	})
}

func (psd *sinksDao) IsHealthy() (bool, error) {
	res := psd.db.Exec(fmt.Sprintf("select count (*) from %s", psd.sinkTable()))
	return res.Error == nil, res.Error
}

func (psd *sinksDao) sinkTable() string {
	return fmt.Sprintf("%s.%s", psd.schema, models.SinkTable)
}

func fetch[T any](db *gorm.DB, table string, query interface{}, args ...interface{}) ([]*T, error) {
	items := make([]*T, 0)
	q := db.Table(table)
	if query != nil {
		q = q.Where(query, args...)
	}
	result := q.Find(&items)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch from %q: %w", table, result.Error)
	}

	return items, nil
}
