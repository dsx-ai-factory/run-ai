Use tokens to facilitate authentication to the NVIDIA Run:ai API.
The API server must be configured to use the NVIDIA Run:ai identity
service to validate authentication tokens.

Use the Tokens endpoints to obtain an access token. The `/api/v2/token` endpoint is compliant with the standard OAuth2 protocol and supports the common OAuth2 grant types. The `/api/v1/token` endpoint is deprecated. Use `/api/v2/token` instead.
