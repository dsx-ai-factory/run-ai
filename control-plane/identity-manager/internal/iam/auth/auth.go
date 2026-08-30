package auth

import (
	"context"
)

type LoginType string

const (
	LoginTypeUser           LoginType = "user"
	LoginTypeServiceAccount LoginType = "app"
)

type Configuration struct {
	Host     string
	Username string
	Password string
}

//go:generate mockery --name Auth --output ./mocks --filename auth.go
type Auth interface {
	GetAdminToken(context.Context) (*string, error)
	GetHost() string
	SetLoginType(context.Context, LoginType) error
}
