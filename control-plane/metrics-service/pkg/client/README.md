## Metrics Service Client: Usage and Versioning Policy

### Importing the Client
Services requiring the metrics service client must import it as a versioned Go package. For example:

```
import "control-plane/metrics-service/pkg/client/v1.0.0"
```

### Development Version Tagging
Developers may create a development version of the client within their working branch by executing the following command from the service's main directory:

```
make client-bump-version-dev
```

This process will generate a development version tag in the following format:

```
control-plane/metrics-service/pkg/client/v1.0.0-dev.0
```

Development versions are intended solely for use in feature branches or any branch other than `main`. **It is strictly prohibited to merge a pull request into the `main` branch if it references a development package version.** The continuous integration (CI) system will automatically fail any such pull request to enforce this policy.

### Production Release Procedure
To publish a production release of the client package, proceed as follows:
1. Merge all relevant client changes into the `main` branch.
2. From the service's main directory, execute:

```
make client-bump-version
```

This will create a production package release with the following format:

```
control-plane/metrics-service/pkg/client/v1.0.0
```

Production versions may be referenced in other branches and pull requests as required.

You can then use this prod version for your other PRs.