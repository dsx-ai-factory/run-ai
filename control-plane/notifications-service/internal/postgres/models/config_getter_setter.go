package models

import "gorm.io/datatypes"

type ConfigGetterSetter interface {
	GetConfig() datatypes.JSON
	SetConfig(datatypes.JSON)
}
