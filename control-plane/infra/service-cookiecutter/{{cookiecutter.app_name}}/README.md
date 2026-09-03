# {{cookiecutter.app_name}} Service

## Overview

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Phasellus pharetra venenatis lorem, dictum elementum felis accumsan non. Pellentesque a odio in massa posuere venenatis. Fusce sit amet bibendum est. Interdum et malesuada fames ac ante ipsum primis in faucibus. Mauris placerat lorem et risus maximus, in porta enim aliquet. Vivamus interdum, dui ac ultrices mollis, quam risus molestie tortor, sit amet semper erat tortor eget sem. Etiam molestie at quam vel vestibulum. Vestibulum a eros urna. Duis ut lectus in lacus iaculis semper eget et lacus. Cras eu diam dui. In cursus ullamcorper mi at interdum. Duis porttitor leo eu neque finibus, a feugiat felis imperdiet. Vivamus orci velit, malesuada et massa a, finibus tincidunt eros.

## Building

### Generating the server

To generate the server run
```
cd {{cookiecutter.app_name}} ; make build
```

### Cross-repo build

Asset service needs code from workload controller.
To allow the cross-repo build:
```
git config --global url.git@github.com:.insteadOf https://github.com/
export GOPRIVATE=github.com/run-ai/runai-operator,   
go get github.com/run-ai/runai-operator 
```
## Running

### Running the server
To run the server:
```
go run cmd/main.go
```

### Authorization

The asset service assumes that all authorization related information is received
as request headers, created by the service/gateway which is responsible for doing
the authentication/authorization.

The following headers are expected:

| Header                     | Purpose                                                                                          |
|----------------------------|--------------------------------------------------------------------------------------------------|
| X-Runai-Subject-Id         | Description of the user or app which request the service.                                        |
| X-Runai-Tenant-Id          | The Tenant ID of the requester.                                                                  |

