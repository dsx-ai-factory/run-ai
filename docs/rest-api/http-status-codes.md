# HTTP Status Codes

NVIDIA Run:ai APIs use standard HTTP status codes to indicate whether a request succeeded or failed.

## Success Responses

| Status code        | Meaning                                                 |
| ------------------ | ------------------------------------------------------- |
| **200 OK**         | The request succeeded.                                  |
| **202 Accepted**   | The request was accepted but is still being processed.  |
| **204 No Content** | The request succeeded and no response body is returned. |

## Client Error Responses (4xx)

| Status code                | Meaning                                                |
| -------------------------- | ------------------------------------------------------ |
| **400 Bad Request**        | The request is invalid.                                |
| **401 Unauthorized**       | Authentication failed.                                 |
| **403 Forbidden**          | The caller is authenticated but not authorized.        |
| **404 Not Found**          | The requested resource does not exist.                 |
| **405 Method Not Allowed** | The method is not allowed for this endpoint.           |
| **409 Conflict**           | The request conflicts with the current resource state. |

## Server Error Responses (5xx)

| Status code                   | Meaning                                                                              |
| ----------------------------- | ------------------------------------------------------------------------------------ |
| **500 Internal Server Error** | An unexpected server error occurred.                                                 |
| **501 Not Implemented**       | The server does not support the requested functionality.                             |
| **502 Bad Gateway**           | The gateway received an invalid response from an upstream service.                   |
| **503 Service Unavailable**   | The service is temporarily unavailable or an internal service dependency is failing. |
| **504 Gateway Timeout**       | The gateway timed out waiting for an upstream service.                               |
