# About the REST API

The NVIDIA Run:ai control plane APIs provide a programmatic interface for managing and operating the NVIDIA Run:ai platform. You can use these APIs to:

* Manage organizations, departments, and projects
* Configure policies and permissions
* Submit and manage workloads
* Manage workload-related assets and metadata
* Audit system activity and operational events

All API requests are authenticated and authorized using access tokens. For details on authentication, see [How to authenticate to the API](how-to-authenticate-to-the-api.md).

## API Versions

The NVIDIA Run:ai control plane APIs are versioned to support both continuously updated environments and fixed, release-based deployments.

### Latest API Version

The **latest** API version represents the most up-to-date control plane API and is continuously updated.

* Updated approximately every two weeks
* Includes the newest features, enhancements, and fixes

The latest API is recommended for environments that track continuous updates.

### Versioned APIs

For self-hosted deployments, NVIDIA Run:ai provides versioned control plane APIs that align with specific product releases (for example, 2.22, 2.23, 2.24).

* API behavior is fixed for the lifetime of the release
* No new features are added after the release is published
* Supported according to the API lifecycle and deprecation policy

Versioned APIs ensure stability and predictability for self-hosted environments that follow a release-based upgrade cadence.

## API Conventions

The NVIDIA Run:ai control plane APIs follow RESTful design principles and use JSON for request and response payloads. See [Making REST API requests](making-rest-api-requests.md) for more details:

* Authentication is provided via bearer tokens in the Authorization header
* Requests use standard HTTP methods (GET, POST, PUT, PATCH, DELETE)
* Request and response bodies are formatted as JSON
* Standard HTTP status codes are used to indicate request success or failure (2xx, 4xx, 5xx)

## API Lifecycle and Deprecation

NVIDIA Run:ai is committed to providing stable and predictable APIs. Deprecated APIs remain supported for a defined transition period:

* **Self-hosted deployments**: two additional released versions

For full details, see [NVIDIA Run:ai API Policy](api-policy.md).
