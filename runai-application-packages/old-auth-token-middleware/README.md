# old auth token middleware
`old-auth-token-middleware` contains middleware that add the jwt token to cantext

The reason for sending the token to the backend is that it doesn't support Traefik labels. As a result, all services that need to communicate through the backend must pass the authentication token in the context. Once the system no longer makes direct calls to backend APIs, this package should be removed.
