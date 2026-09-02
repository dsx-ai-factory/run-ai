## Authorization Service Client: Usage and Versioning Policy

In charge of calling the authorization service evaluation requests based on user context and requested resource permissions.

### Importing the Client
Services requiring the authorization service client must import it as a versioned Go package. For example:

```
import "control-plane/authorization/pkg/client/v1.0.0"
```

To wire the client in your service:
1. Add an `Authorization` section to your config (taken from the `authorization_client` package) and set the authorization url (typically `runai-backend-authorization.runai-backend.svc:8080`).
2. Add to the service init function:

   ```
   authorization_client.Init(conf.Authorization)
   ```
3. For each exposed API, define its requested resource or mark it as not-authorized via the OpenAPI mustache parameters:

   For a non-authorized endpoint (e.g. health checks):
   ```
   get:
     x-no-authorization: true
   ```

   For an authorized endpoint that does not require specific resourceId permissions:
   ```
   get:
     x-authorize:
       resource-type: <ResourceType>
       action: <Action>
   ```

   For an authorized endpoint that requires specific resourceId permissions:
   ```
   get:
     x-authorize:
       resource-type: <ResourceType>
       action: <Action>
       resource-id: <IdParamName>
   ```

### Development Version Tagging
Developers may create a development version of the client within their working branch by executing the following command from the service's main directory:

```
make client-bump-version-dev
```

This process will generate a development version tag in the following format:

```
control-plane/authorization/pkg/client/v1.0.0-dev.0
```

Development versions are intended solely for use in feature branches or any branch other than `master`. **It is strictly prohibited to merge a pull request into the `master` branch if it references a development package version.** The continuous integration (CI) system will automatically fail any such pull request to enforce this policy.

### Production Release Procedure
To publish a production release of the client package, proceed as follows:
1. Merge all relevant client changes into the `master` branch.
2. From the service's main directory, execute:

```
make client-bump-version
```

This will create a production package release with the following format:

```
control-plane/authorization/pkg/client/v1.0.0
```

Production versions may be referenced in other branches and pull requests as required.

You can then use this prod version for your other PRs.
