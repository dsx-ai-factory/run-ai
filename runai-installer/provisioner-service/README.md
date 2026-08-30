Provisioner Service (Standalone)
================================

Endpoints
---------
- POST /environments
  - Body: {"provisioningMode": "control-plane-only" | "full"}
  - Returns 202 with id
- GET /environments/{id}
  - Returns status

Run
---
- make build
- PORT=8080 ./bin/provisioner-service-$(go env GOOS)-$(go env GOARCH)

Docker
------
- make docker-build
- docker run -p 8080:8080 provisioner-service:local


