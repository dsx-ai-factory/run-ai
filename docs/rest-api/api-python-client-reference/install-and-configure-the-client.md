# Install and Configure the Client

This section explains how to install `runapy`, select the right client type for your environment and configure authentication.

<Warning>

**Important**

The Python client is not covered under standard NVIDIA Run:ai product support.

</Warning>

## Prerequisites

* **NVIDIA Run:ai control plane version ≥ 2.18** - Required for compatibility
* **Python version ≥ 3.8** - Required for dependency compatibility
* **NVIDIA Run:ai access keys** - Required to generate the `client_id` and `client_secret` credentials:
  * To create organization-level service accounts, see [Service accounts](../../infrastructure-setup/authentication/service-accounts.md)
  * To create user-level access keys, see [User access keys](../../settings/user-settings/user-access-keys.md)

<Info>

**Note**

Ensure your token has the required [role and scope](../../infrastructure-setup/authentication/accessrules.md) permissions for your intended API operations. API calls will fail with a 403 error if the token lacks sufficient role or scope.

</Info>

## Installation

* For the latest NVIDIA Run:ai control plane:

```bash
pip install runapy
```

* For self-hosted environments, specify the minor version matching your control plane version. For example, with NVIDIA Run:ai 2.19 control plane:

```bash
pip install runapy~=1.219.0
```

To understand the versioning scheme, refer to the [Versioning](api-python-client-runapy.md#versioning) section.

## Client Types

The Python client provides multiple client types tailored to different environments:

1. `ApiClient` - Standard synchronous client for sequential operations
2. `ThreadedApiClient` - Thread-safe client for parallel operations in multithreaded environments
3. `AsyncApiClient` - Asynchronous client for `async/await` support

For additional configuration options, see [Client configuration options](install-and-configure-the-client.md#client-configuration-options).

### Standard API Client

The below shows a basic example using the standard client:

```python
from runai.configuration import Configuration
from runai.api_client import ApiClient
from runai.runai_client import RunaiClient

config = Configuration(
    client_id="your-client-id",
    client_secret="your-client-secret",
    runai_base_url="https://your-org.run.ai",
)

client = RunaiClient(ApiClient(config))

# Start making API calls
projects = client.organizations.projects.get_projects()
```

### Multithreaded Operations

For parallel operations in a multithreaded environment, use `ThreadedApiClient`:

```python
from runai.api_client import ThreadedApiClient

with RunaiClient(ThreadedApiClient(config)) as client:
    # These operations can run concurrently
    projects = client.organizations.projects.get_projects()
```

### Asynchronous Operations

For async/await support, use `AsyncApiClient`:

```python
import asyncio
from runai.api_client import AsyncApiClient

async def main():
    async with RunaiClient(AsyncApiClient(config)) as client:
        projects = await client.organizations.projects.get_projects()

asyncio.run(main())
```

## Authentication Methods

The client supports three authentication methods:

### Client Credentials

Use the `client_id` and `client_secret` from your NVIDIA Run:ai [service accounts](../../infrastructure-setup/authentication/service-accounts.md) or [user access keys](../../settings/user-settings/user-access-keys.md):

```python
config = Configuration(
    client_id="your-client-id",
    client_secret="your-client-secret",
    runai_base_url="https://your-org.run.ai"
)
```

### Bearer Token

Direct authentication using a bearer token:

```python
config = Configuration(
    bearer_token="your-bearer-token",
    runai_base_url="https://your-org.run.ai"
)
```

### CLI v2 Token (Bearer Token)

The CLI v2 token method is useful for:

* Local development and testing
* Scripts running on machines with existing CLI authentication
* Maintaining consistent authentication with CLI sessions
* End user operations

Requirements:

* NVIDIA Run:ai CLI v2 is [installed](../../reference/cli/install-cli.md)
* Successful `runai login` completed
* Valid authentication token in CLI config

```python
from runai.cliv2_config_loader import CLIv2Config

# Default config path is ~/.runai
config = CLIv2Config()
# Or specify a custom path
config = CLIv2Config(cliv2_config_path="/path/to/.runai")

token = config.token
runai_base_url = config.control_plane_url
cluster_id = config.cluster_uuid

client = RunaiClient(
    cluster_id=cluster_id,
    bearer_token=token,
    runai_base_url=runai_base_url
)
```

## Client Configuration Options

| Parameter              | Type     | Description                                                                                                     |
| ---------------------- | -------- | --------------------------------------------------------------------------------------------------------------- |
| `client_id`            | `string` | **Required**: The NVIDIA Run:ai client ID, usually representing the access key name                             |
| `client_secret`        | `string` | **Required**: The client secret associated with the NVIDIA Run:ai access key                                    |
| `runai_base_url`       | `string` | **Required**: The base URL for the NVIDIA Run:ai instance your organization uses (e.g., `https://myorg.run.ai`) |
| `bearer_token`         | `string` | **Optional**: Bearer token for CLIv2 compatibility. Cannot be used together with client credentials.            |
| `verify_ssl`           | `bool`   | **Optional**: Whether to verify SSL certificates. Default is `True`                                             |
| `ssl_ca_cert`          | `string` | **Optional**: Path to CA certificate file                                                                       |
| `cert_file`            | `string` | **Optional**: Path to client certificate file                                                                   |
| `key_file`             | `string` | **Optional**: Path to client key file                                                                           |
| `pool_maxsize`         | `int`    | **Optional**: Maximum number of connections to keep in pool. Default is `4`                                     |
| `pool_size`            | `int`    | **Optional**: Initial number of connections in pool. Defaults to `pool_maxsize`                                 |
| `retry_enabled`        | `bool`   | **Optional**: Whether to enable request retries. Default is `True`                                              |
| `retry_max_retries`    | `int`    | **Optional**: Maximum number of retry attempts. Default is `3`                                                  |
| `retry_backoff_factor` | `float`  | **Optional**: Exponential backoff factor between retries. Default is `0.5`                                      |
| `proxy_url`            | `string` | **Optional**: URL for proxy server                                                                              |
| `proxy_headers`        | `dict`   | **Optional**: Additional headers for proxy                                                                      |
| `proxy_server_name`    | `string` | **Optional**: SNI hostname for TLS connections                                                                  |
| `auto_refresh_token`   | `bool`   | **Optional**: Whether to auto refresh token before expiry. Default is `True`                                    |
| `token_refresh_margin` | `int`    | **Optional**: Seconds before expiry to refresh token. Default is `60`                                           |
| `debug`                | `bool`   | **Optional**: Enable debug logging. Default is `False`                                                          |
