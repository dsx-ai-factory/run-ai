module github.com/run-ai/runai/control-plane/{{cookiecutter.app_name}}

go {{cookiecutter.go_version}}

require (
	github.com/go-chi/chi/v5 v5.0.11
	github.com/jackc/pgx/v5 v5.5.5
	github.com/onsi/ginkgo/v2 v2.13.1
	github.com/onsi/gomega v1.30.0

	github.com/run-ai/runai/runai-common-packages/cache/v2 v2.4.2

	github.com/run-ai/runai/runai-common-packages/config v1.2.8
	github.com/run-ai/runai/runai-common-packages/dao-utils v1.4.19
	github.com/run-ai/runai/runai-application-packages/old-auth-token-middleware v1.0.0
	github.com/run-ai/runai/runai-common-packages/logging v1.0.13
	github.com/run-ai/runai/runai-common-packages/runaictx v1.1.2
	github.com/run-ai/runai/runai-common-packages/runaihttp v1.1.5
	gorm.io/gorm v1.25.5
	k8s.io/utils v0.0.0-20260507154919-ff6756f316d2
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	golang.org/x/crypto v0.52.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)



replace github.com/run-ai/runai/control-plane/authorization v0.0.0 => ../authorization


