package db

import (
	"gorm.io/gorm"
)

func DBMiddleware(errorCallback func(e error)) gorm.Plugin {
	if errorCallback == nil {
		errorCallback = func(e error) {}
	}
	return newGormPrometheusMetricsMiddleware(errorCallback)
}
