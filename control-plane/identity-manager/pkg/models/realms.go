package models

type Realm struct {
	BaseModel
	TenantId string `gorm:"not null;unique_index"`
	Realm    string `gorm:"not null;unique_index"`
}

type DeleteRealm struct {
	DeleteBaseModel
	TenantId string `gorm:"not null;unique_index"`
	Realm    string `gorm:"not null;unique_index"`
}
