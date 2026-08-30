package dbmodel

import (
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
)

type TemplateAssetRec struct {
	AssetId     string                 `gorm:"primary_key"`
	SpecificEnv *api.SpecificRunParams `gorm:"serializer:json; type:jsonb;"`
}

func (TemplateAssetRec) TableName() string {
	return dbtables.TableNameAssetsTemplate
}

type TemplateContentRec struct {
	TemplateAssetId  string
	ContainedAssetId string
}

func (TemplateContentRec) TableName() string {
	return dbtables.TableNameAssetsTemplateContent
}

func (TemplateContentRec) AssetIdColumn() string {
	return dbtables.AssetsTemplateContentColumnAssetId
}

// TemplateCombinedRec holds result of joined information from all relevant tables
type TemplateCombinedRec struct {
	api.AssetMeta      `gorm:"embedded"`
	TemplateAssetRec   `gorm:"embedded"`
	TemplateContentRec `gorm:"embedded"`
}

type TemplateAssetsCombinedRec struct {
	TemplateAssetId  string
	ContainedAssetId string
	AssetMeta        AssetMetaRec `gorm:"embedded"`
}
