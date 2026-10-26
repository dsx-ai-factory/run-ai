# About the REST API

The NVIDIA Run:ai Multi-Tenant REST APIs provide a programmatic interface for managing and operating multi-tenant NVIDIA Run:ai deployments.

These APIs are used by platform administrators to manage shared control plane resources across tenants.

You can use these APIs to:

* Manage tenants and tenant-level configuration
* Register and manage clusters
* Configure platform-level settings
* Monitor and audit multi-tenant operational activity

All API requests are authenticated using access tokens. For details on authentication and access setup, see [API Access Setup](api-access-setup.md).

## API Versions

The NVIDIA Run:ai Multi-Tenant REST APIs are release-based and are available only for self-hosted deployments.

Each API version aligns with a specific NVIDIA Run:ai product release (for example, 2.22, 2.23, 2.24).

* API behavior is fixed for the lifetime of the release
* No new features are added after the release is published
* Changes are introduced only in newer released versions

Using a versioned API ensures stability and predictability for platform administrators operating multi-tenant environments.

## API Conventions

The NVIDIA Run:ai Multi-Tenant REST APIs follow RESTful design principles and use JSON for request and response payloads. See [Making REST API Requests](making-rest-api-requests.md) for more details:

* Authentication is provided via bearer tokens in the `Authorization` header
* Requests use standard HTTP methods (`GET`, `POST`, `PUT`, `PATCH`, `DELETE`)
* Request and response bodies are formatted as JSON
* Standard HTTP status codes are used to indicate request success or failure (`2xx`, `4xx`, `5xx`)

## API Lifecycle and Deprecation

NVIDIA Run:ai is committed to providing stable and predictable Multi-Tenant APIs. When an API or field is deprecated, the deprecated version remains supported for two additional released versions. After the deprecation period ends, deprecated APIs or fields are removed from both the codebase and the documentation. For full details, see [NVIDIA Run:ai API Policy](../api-policy.md).
