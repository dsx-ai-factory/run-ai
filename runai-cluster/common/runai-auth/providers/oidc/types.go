package oidc

type Configuration struct {
	Issuer           string `json:"issuer"`
	UserInfoEndpoint string `json:"userinfo_endpoint"`
	JWKSURI          string `json:"jwks_uri"`
}

type UserInfo struct {
	Groups []string `json:"groups"`
}
