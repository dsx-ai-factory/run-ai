package models

import (
	"time"
)

type Common struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy string
}
