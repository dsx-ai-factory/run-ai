package models

type DisabledRole struct {
	RoleId    int64     `gorm:""`
	ScopeType RoleScope `gorm:""`
	ScopeId   string    `gorm:""`
}
