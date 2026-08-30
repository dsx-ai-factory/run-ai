package models

import (
	"gorm.io/gorm"
)

const (
	ValidationTable = "validation"
)

type Validation struct {
	gorm.Model
	Data string `gorm:"not null" json:"data"`
}

func (v *Validation) TableName() string {
	return ValidationTable
}
