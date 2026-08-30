package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm/clause"
)

const typeNumber = "Number"

type TenantSetting struct {
	TenantID int    `gorm:"column:tenant_id;primaryKey"`
	Key      string `gorm:"column:key;primaryKey"`
	Value    string `gorm:"column:value"`
	Type     string `gorm:"column:type"`
}

func (TenantSetting) TableName() string {
	return "tenant_settings"
}

type TenantSettingsDAO struct{}

func NewTenantSettingsDAO() *TenantSettingsDAO {
	return &TenantSettingsDAO{}
}

func (d *TenantSettingsDAO) db(ctx context.Context) orm.DB {
	return runaidb.DB().WithContext(ctx)
}

func (d *TenantSettingsDAO) Set(ctx context.Context, tenantID int, settings map[string]interface{}) error {
	for key, value := range settings {
		valueStr, valueType := convertToString(value)
		row := TenantSetting{
			TenantID: tenantID,
			Key:      key,
			Value:    valueStr,
			Type:     valueType,
		}
		if _, err := d.db(ctx).
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "tenant_id"}, {Name: "key"}},
				DoUpdates: clause.AssignmentColumns([]string{"value", "type"}),
			}).
			Create(&row); err != nil {
			return fmt.Errorf("failed to upsert setting %q: %w", key, err)
		}
	}
	return nil
}

func convertToString(value interface{}) (string, string) {
	switch v := value.(type) {
	case bool:
		return strconv.FormatBool(v), "Boolean"
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), typeNumber
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32), typeNumber
	case int:
		return strconv.Itoa(v), typeNumber
	case int64:
		return strconv.FormatInt(v, 10), typeNumber
	case string:
		return v, "String"
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v), "String"
		}
		return string(b), "Object"
	}
}

func (d *TenantSettingsDAO) GetByKeyForAllTenants(ctx context.Context, key string) (map[int]interface{}, error) {
	var rows []TenantSetting
	if err := d.db(ctx).Where("key = ?", key).Find(&rows); err != nil {
		return nil, err
	}

	result := make(map[int]interface{}, len(rows))
	for _, row := range rows {
		result[row.TenantID] = convertFromString(row.Value, row.Type)
	}
	return result, nil
}

func (d *TenantSettingsDAO) GetAllByTenantID(ctx context.Context, tenantID int) (map[string]interface{}, error) {
	var rows []TenantSetting
	if err := d.db(ctx).Where("tenant_id = ?", tenantID).Find(&rows); err != nil {
		return nil, err
	}

	result := make(map[string]interface{}, len(rows))
	for _, row := range rows {
		result[row.Key] = convertFromString(row.Value, row.Type)
	}
	return result, nil
}

func convertFromString(value string, valueType string) interface{} {
	switch valueType {
	case "Boolean":
		return value != "false"
	case typeNumber:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			logging.Errorf("failed to parse number setting value %q: %v", value, err)
			return value
		}
		return f
	case "Object", "Array":
		var parsed interface{}
		if err := json.Unmarshal([]byte(value), &parsed); err != nil {
			logging.Errorf("failed to parse JSON setting value %q: %v", value, err)
			return value
		}
		return parsed
	default:
		return value
	}
}
