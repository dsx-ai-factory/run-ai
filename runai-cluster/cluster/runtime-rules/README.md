# runtime-rules
A simple tool to enforce runtime rules on the workloads of a Kubernetes cluster.
for now we support only WorkloadRunner-backed workloads:
- Interactive
  - InteractiveJobMaxIdleDurationSecs
  - InteractiveJobTimeLimitSecs
- Training and distributed training
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
