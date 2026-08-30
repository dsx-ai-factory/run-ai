# runtime-rules
A simple tool to enforce runtime rules on the workloads of a Kubernetes cluster.
for now we support only the following workloads:
- InteractiveWorkload
  - InteractiveJobMaxIdleDurationSecs
  - InteractiveJobTimeLimitSecs
- TrainingWokload
  - TrainingJobMaxIdleDurationSecs
  - TrainingJobTimeLimitSecs


## Makefiles
Run linter on the code using `.golangcli.yml` as config:
```
make lint
```

## Linter Commands
Arrange the import section for the project
```
goimports -local "github.com/GoogleContainerTools/skaffold" -w .
```