module github.com/dsx-ai-factory/run-ai/runai-common-packages/auth

go 1.23.4

require (
	github.com/coreos/go-oidc/v3 v3.11.0
	github.com/dsx-ai-factory/run-ai/runai-common-packages/safemapcache v0.0.0-00010101000000-000000000000
	golang.org/x/oauth2 v0.22.0
)

require (
	github.com/go-jose/go-jose/v4 v4.0.2 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	golang.org/x/crypto v0.31.0 // indirect
)

replace github.com/dsx-ai-factory/run-ai/runai-common-packages/safemapcache => ../runai-safemapcache
