# Clients module
This module is meant to contain the common client APIs of our services.
Each client has its own directory, which contains 2 directories - `api` and `generated`.
The `api` directory contains the OpenAPI spec, while the `generated` directory contains the generated golang client code.

## How to update API?
Update the OpenAPI spec under the `api` directory of the client, and run:
```bash
make all
```

## How to add a new client?
1. Create a new directory under `clients` with the name of the client. 
2. Create a `api` directory under the client directory, and put the OpenAPI spec there.
3. Add a make target in the `Makefile` of the client directory to generate the client code, and add the target to the `all` target.
4. Run `make all` to generate the client code.