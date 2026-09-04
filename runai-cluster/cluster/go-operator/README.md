# RunAI Cluster Operator 
## Testing

### Unit tests without a mocked environment
```
make test
```

### Testing against a live cluster using kind
When writing tests that require mocking, write them under `tests` and they would be executed against a kind cluster
```
make test-kind
```

## Adding a new Operand
- Implement the `SetDefaultsWhereNeeded()` method for all API types within the `apis/run/v1/$OPERAND` package
- Update the `SetDefaultsWhereNeeded()` method of `RunaiConfig` in `apis/run/v1/runaiconfig_types.go` to include the operand type you set defaults for
- Copy the `internal/run/operands/agent` package to `internal/run/operands/$OPERAND`, rename it to the name of your newly added operand
- Update the `name` constant within `internal/run/operands/$OPERAND/operand.go` file
- Create `resourceFunctions` and `resourceStatusFunctions` within the `internal/run/operands/$OPERAND/resource` package (signatures can be found in the slice definitions in `operand.go`), each resource kind in a separate file
- Write any tests that do not require mocking in your package, use `ginkgo` to write your tests
- Update the `resourceFunctions` and `resourceStatusFunctions` slices with the functions you implemented
- Include your new operand in the `ManagedOperands()` function of the `RunaiConfig` controller at `controllers/run/managed_operands.go`
- Implement any tests that require mocking in a new file in the `tests` package, tests in that package are designed to be run against both mocked and live clusters, peek at the `tests/utils` package for some helpful functions

## Adding new CRD version
* Run `hack/new-crd-version.sh <GROUPNAME> <VERSION> <KIND> [NAMESPACED]` to generate boilerplate for your new CRD or version of one.
* Group names should be `run` or `run.ai` (which will both be used as `run.ai`) or `scheduling`/`scheduling.run.ai`
* Try to only add fields between CRD versions, so that the automatic conversion will work. This is more important for CRDs such as PodGroups which are only created onces than to Projects/Departments that are being synced by the agent.
* You will have to manually move the `storageversion` tag to the new version of the CRD and add the `genclient` tag to generate listers/informers/clients.
* Add the new version to the `hack/update-client.sh` script to generate listers/informers/clients.
* run `make generate manifests generate-client` to generate the CRD new version, go clients, listers, informers etc (CICD will make you run it)
* You can use `hack/copy-crds.sh` to copy the output CRDs to the chart where they will be later moved to the be used in public.
