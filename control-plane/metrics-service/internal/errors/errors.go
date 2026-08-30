package errors

import (
	"errors"
	"net/http"
	"strings"
)

const (
	MetricStoreNotFound = "metric store not found for tenant"
	NoSuchHost          = `no such host`
)

var (
	ErrClusterVersionNotSupported = errors.New("cluster version unsupported for metric")
	ErrMetricStoreAlreadyExists   = errors.New("metric store already exists")
)

type GetTenantConnection struct {
	error
	StatusCode int
}

func NewGetTenantConnection(err error, status int) error {
	return &GetTenantConnection{error: err, StatusCode: status}
}

func (e GetTenantConnection) Error() string {
	if e.error != nil {
		return e.error.Error()
	}
	return "failed to get tenant connection"
}

func IsTenantNotFoundErr(err error) bool {
	var pe *GetTenantConnection
	if errors.As(err, &pe) {
		return pe.StatusCode == http.StatusNotFound
	}
	return false
}

func FailedToGetDatasource(err error) {
}

func IsHostNotFound(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), MetricStoreNotFound) || strings.Contains(strings.ToLower(err.Error()), NoSuchHost)
}

func IsClusterVersionNotSupportedErr(err error) bool {
	return errors.Is(err, ErrClusterVersionNotSupported)
}
