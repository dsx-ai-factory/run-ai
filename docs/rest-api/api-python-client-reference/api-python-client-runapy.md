# API Python Client (runapy)

The NVIDIA Run:ai Community Python Client (`runapy`) provides a comprehensive, production-ready interface for interacting with the NVIDIA Run:ai platform's REST API. This client simplifies the process of programmatically managing your AI workloads, resources, and infrastructure without requiring manual API token handling or custom API integrations.

Refer to the [runapy GitHub repository](https://github.com/runai-professional-services/runapy) for more details.

<Warning>

**Important**

This Python client was developed by the NVIDIA Run:ai Professional Services team and is provided as is. It is not an official component of the NVIDIA Run:ai product and is not maintained by the NVIDIA Run:ai engineering team.

</Warning>

## Key Features

* **Unified interface** - Access all NVIDIA Run:ai REST API operations through a single client.
* **Automatic authentication** - Automatically handles API token refresh for long-running sessions.
* **Robust networking** - Includes built-in retry mechanism with exponential backoff for transient network errors.
* **Type safety** - Uses Pydantic models for runtime data and validation and serialization
* **Multiple client modes** - Supports synchronous, threaded, and asynchronous operations.
* **Production-ready** - Offers comprehensive error handling and logging for enterprise environments

## Versioning

The NVIDIA Run:ai API is an integral part of the control plane microservices. As the control plane evolves, the API may introduce additions, modifications, or deprecations. This version binding ensures that API changes in the client match the control plane capabilities.

The client follows semantic versioning (X.Y.Z) with a special alignment to NVIDIA Run:ai's control plane versions:

* **X (Major)** - Major changes in the client
* **Y (Minor)** - NVIDIA Run:ai control-plane version (e.g., 219 for control-plane 2.19)
* **Z (Patch)** - Small bug fixes and non-breaking changes

For example:

```
1.218.0  # Initial release for control-plane 2.18
1.218.1  # Patch release for control-plane 2.18
1.219.0  # Initial release for control-plane 2.19
```

This versioning convention ensures that:

* Each control plane version has a corresponding client version
* Client versions clearly indicate their compatible control plane version
* Client breaking changes are managed in major version bumps

## API Structure

The Python client mirrors the structure of the NVIDIA Run:ai REST API documentation using three levels:

1. **API categories** - Top-level groupings (e.g., `organizations`, `authentication_and_authorization`)
2. **API groups** - Resource-specific groups within categories (e.g., `projects`, `departments` under `organizations`)
3. **API methods** - Individual operations on resources (e.g., `get_projects`, `create_project`)

Example of navigating the API structure:

```python
# List available API groups in the organizations category
client.organizations.options()
# ['projects', 'departments', 'clusters', 'nodes', ...]

# List available methods in the projects group
client.organizations.projects.options()
# ['get_projects', 'create_project', 'delete_project', ...]
```

## Deprecated Methods

Deprecated methods are hidden by default. To view both supported and deprecated methods, use the `include_deprecated` flag:

```python
client.organizations.projects.options(include_deprecated=True)
```

## Examples

The repository provides:

* [Hand-crafted examples](https://github.com/runai-professional-services/runapy/tree/main/examples) - Common use cases with clean syntax
* [Auto-generated examples](https://github.com/runai-professional-services/runapy/tree/main/examples/generated) - Located under `/examples/generated` (some may require minor edits)

The client exposes nearly all API functionality available in the NVIDIA Run:ai UI. If a workflow isn’t shown in the examples directory, check the relevant API category directly.
