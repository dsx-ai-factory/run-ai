# Making REST API Requests

This guide explains how to interact with the NVIDIA Run:ai Multi-Tenant REST API and how API requests and responses are structured.

## Request to the REST API

Each request to the NVIDIA Run:ai Multi-Tenant REST API is composed of the following elements:

* HTTP method
* Endpoint path
* Headers
* Authentication
* Parameters (path, query, and body)

Each REST API request includes an HTTP method and an endpoint path. Depending on the endpoint, requests may also require headers, authentication details, query parameters, or a request body. The API reference documentation specifies the method, path, and parameters for each endpoint, and includes example requests and responses.

### HTTP Method

The HTTP method of an endpoint defines the type of action it performs on a given resource. NVIDIA Run:ai Multi-Tenant APIs use standard HTTP methods:

* `GET` - Retrieve a resource or list of resources
* `POST` - Create a new resource
* `PATCH` - Update specific fields of a resource
* `PUT` - Replace a resource or collection
* `DELETE` - Delete a resource

### Endpoint Paths

Each API endpoint is identified by a path. Values enclosed in `{}` represent path parameters that must be provided in the request. Replace each placeholder with the actual resource identifier.

```
/api/v1/tenants/{tenantId}
```

### Headers

Headers provide additional information about the request and response. Most API requests require the following headers. Missing or invalid headers may result in authentication or validation errors:

* `Authorization: Bearer <access-token>`
* `Content-Type: application/json`
* `Accept: application/json`

### Authentication

NVIDIA Run:ai API endpoints require authentication. Authentication is performed using **bearer tokens**, which are generated from access keys (client ID and client secret). Unauthenticated requests, or requests with invalid or expired tokens, return a `401 Unauthorized` error. For details on creating access keys and requesting tokens, see [API Access Setup](../about-the-rest-api.md).

### Parameters

API endpoints may accept parameters to control request behavior or provide input data.

#### Path parameters

Path parameters are part of the endpoint path and are **required**. Example:

```
/api/v1/clusters/{clusterId}
```

#### Query parameters

Query parameters modify the behavior of a request and are typically optional. Examples include:

* Filtering results
* Limiting or paging returned data
* Controlling sorting or output format

Query parameters are appended to the URL:

```
v1/authorization/access-rules?limit=100&offset=0&sortBy=name&sortOrder=asc
```

#### Body parameters

Body parameters are sent in the request body as JSON and are typically used with `POST`, `PUT`, or `PATCH` requests. The API reference documentation specifies which body fields are required or optional for each endpoint.

## Making a Request

After obtaining an access token, include it in the `Authorization` header of your request.

Example (cURL):

```bash
curl -X GET \
  'https://console.<DOMAIN>/api/v1/tenants' \
  -H 'Authorization: Bearer <TOKEN>' \
  -H 'Accept: application/json'
```

## Using the Response

Each API response includes:

* An HTTP status code
* Response headers
* An optional response body

### Status Codes

Status codes indicate whether a request succeeded or failed.

* **2xx** – Request succeeded
* **4xx** – Client error (authentication, authorization, or request validation)
* **5xx** – Server or gateway error

For a full list of status codes, see [HTTP Status Codes](http-status-codes.md).

### Response Body

Unless otherwise specified:

* Responses are returned in JSON format
* Timestamps use UTC and ISO 8601 format
* Fields may be returned as `null` when no value is available

List endpoints may return a summary representation of resources, while individual resource endpoints typically return a full representation.
