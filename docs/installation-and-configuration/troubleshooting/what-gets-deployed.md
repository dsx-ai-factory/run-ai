---
---

# What Gets Deployed

<Info>
**TODO:** Reference page listing every Kubernetes resource deployed by NVIDIA Run:ai across both namespaces, with a description of what each service does. Addresses RUN-36450 - customers (especially airgapped deployments) need this for security reviews, internal troubleshooting guides, and understanding unexpected resources after installation.

**runai-backend namespace** - control plane services:
- runai-backend-audit-service
- runai-backend-catalog-service
- runai-backend-workloads-manager
- runai-backend-notifications-proxy
- runai-backend-org-unit-helper
- runai-backend-workloads-helper
- runai-backend-cli-exporter
- runai-backend-datavolumes
- runai-backend-metrics-service
- runai-backend-notification-service
- runai-backend-org-unit-service
- runai-backend-nats (StatefulSet)

**runai namespace** - cluster components:
- application-controller
- workload-exporter
- cluster-redis
- engine-admission
- assets-sync
- cluster-api
- inference-workload-controller
- shared-object-controller
- accessrule-controller
</Info>
