# Run:AI job-controller

![image](https://user-images.githubusercontent.com/10087174/126780848-6d708627-d0ee-444c-b940-d9434741c7a5.png)

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