# Command Package README

This README provides guidelines on how to add new commands with framework and workload type to each domain package and how to modify the factory accordingly.

## Adding a New Command

To add a new command, follow these steps:

1. Create a new package for your command (e.g., `attach`).
2. Implement the command-specific `Texts` struct.
3. Implement the `Factory` struct.
4. Implement the command creation function.
5. Add tests for your new implementations.

### 1. Create Command-Specific Texts

Create a struct that implements the `cmd.CommandTexts` interface:

```go
package attach

import (
    "github.com/run-ai/runai/control-plane/runai-cli/internal/api"
    "github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
    "strings"
)

// Texts struct
type Texts struct {
    WorkloadType api.WorkloadType
    Framework    api.Framework
}

const (
    shortMsg   = "Attach to a running container in a {{.WorkloadType}} {{.Framework}} job"
    exampleMsg = `# Attaching to the main worker of a {{.WorkloadType}} {{.Framework}}
runai {{.WorkloadType}} {{.Framework}} attach {{.Framework}}-01 --tty --stdin

# Attaching to a specific worker of a {{.Framework}}
runai {{.WorkloadType}} {{.Framework}} attach {{.Framework}}-01 --pod {{.Framework}}-01-worker-1 --tty --stdin`
    baseErrorMsg = "failed to attach to {{.WorkloadType}} {{.Framework}} job"
)

func (t *Texts) GetShort() string {
    return strings.NewReplacer(cmd.WorkloadTypeReplacer, t.WorkloadType.String(),
        cmd.FrameworkReplacer, t.Framework.String()).Replace(shortMsg)
}

func (t *Texts) GetExample() string {
    return strings.NewReplacer(cmd.WorkloadTypeReplacer, t.WorkloadType.String(),
        cmd.FrameworkReplacer, t.Framework.String()).Replace(exampleMsg)
}

func (t *Texts) GetLong() string {
    return ""
}

func (t *Texts) GetBaseErrorMsg() string {
    return strings.NewReplacer(cmd.WorkloadTypeReplacer, t.WorkloadType.String(),
        cmd.FrameworkReplacer, t.Framework.String()).Replace(baseErrorMsg)
}

func (t *Texts) GetAliases() []string {
    return nil
}
```

### 2. Implement the Factory

Create a `Factory` struct that implements the `cmd.Factory` interface:

```go
type Factory struct {
    instance  *config.Instance
    cmdHelper helper.Cmd
}

func NewCommandFactory(instance *config.Instance, cmdHelper helper.Cmd) *Factory {
    return &Factory{
        instance:  instance,
        cmdHelper: cmdHelper,
    }
}

func (f *Factory) IsSupported(framework api.Framework, workloadType api.WorkloadType) bool {
    switch framework {
    case api.MPI:
        return true
    case api.Runai:
        return workloadType == api.Training
    default:
        return false
    }
}

func (f *Factory) GetCommand(framework api.Framework, workloadType api.WorkloadType) (cmd.ActionCommand, error) {
    if !f.IsSupported(framework, workloadType) {
        return nil, cmd.NewUnsupportedError(framework, workloadType)
    }

    commandText := &Texts{
        WorkloadType: workloadType,
        Framework:    framework,
    }

    return NewAttachCommand(framework, nil, workloadType, commandText, f.instance, f.cmdHelper), nil
}
```

### 3. Implement Command Creation Function

Implement the `NewAttachCommand` function (or equivalent for your command):

```go
func NewAttachCommand(framework api.Framework, initBaseCommandFn func() *cobra.Command, workloadType api.WorkloadType, commandText cmd.CommandTexts, instance *config.Instance, cmdHelper helper.Cmd) cmd.ActionCommand {
    // Implement command creation logic here
}
```

### 4. Add Tests

Add tests for your new implementations:

```go
func TestAttachTexts_GetShort(t *testing.T) {
    texts := &Texts{
        WorkloadType: api.Training,
        Framework:    api.MPI,
    }
    expected := "Attach to a running container in a Training MPI job"
    if got := texts.GetShort(); got != expected {
        t.Errorf("GetShort() = %v, want %v", got, expected)
    }
}

// Add more tests for other methods and structs
```

## Modifying the Factory

To modify the factory for new frameworks or workload types:

1. Update the `IsSupported` method to include the new framework or workload type.
2. Update the `GetCommand` method if necessary to handle the new framework or workload type.

Example:

```go
func (f *Factory) IsSupported(framework api.Framework, workloadType api.WorkloadType) bool {
    switch framework {
    case api.MPI, api.NewFramework:
        return true
    case api.Runai:
        return workloadType == api.Training || workloadType == api.NewWorkloadType
    default:
        return false
    }
}

func (f *Factory) GetCommand(framework api.Framework, workloadType api.WorkloadType) (cmd.ActionCommand, error) {
    if !f.IsSupported(framework, workloadType) {
        return nil, cmd.NewUnsupportedError(framework, workloadType)
    }

    commandText := &Texts{
        WorkloadType: workloadType,
        Framework:    framework,
    }

    return NewAttachCommand(framework, nil, workloadType, commandText, f.instance, f.cmdHelper), nil
}
```

Remember to update tests accordingly when modifying the factory.

## Best Practices

1. Keep the `Factory` interface consistent across all command packages.
2. Implement thorough error handling and unsupported configuration checks.
3. Maintain consistent naming conventions across packages.
4. Write comprehensive tests for all new implementations and modifications.
5. Update relevant documentation when adding new commands or modifying existing ones.
6. Use the template-based approach with `{{.WorkloadType}}` and `{{.Framework}}` placeholders in your constant strings.
7. Always use `strings.NewReplacer` to replace the placeholders in your `Texts` struct methods.

By following these guidelines, you can easily extend the command system with new frameworks and workload types while maintaining consistency and testability across the project.