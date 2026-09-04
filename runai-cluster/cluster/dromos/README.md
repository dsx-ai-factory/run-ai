# dromos

`dromos` is the library that powers Run:ai's `anyworkload-controller` binary.

It manages the `WorkloadRunner` CRD - it lets users submit any AI
workload to a Kubernetes cluster, encapsulate it with instructions, and
delegate the lifecycle (mutation, validation, child-resource creation, status
reporting) to a set of plugins. The controller is plugin-oriented: behavior is
composed by passing plugin builders to `app.RunController` at startup.

## Layout

```
runai-cluster/cluster/dromos/
├── pkg/plugin/framework/   public Plugin SDK (interfaces, PluginBuilder)
├── pkg/app/                RunController() - the boot sequence
├── pkg/controller/         WorkloadRunner reconciler
├── pkg/webhook/            validator/mutator handlers
├── pkg/config/             DromosConfig
├── pkg/reconcile/          internal reconcile state
└── plugins/
    ├── mnnvl/              in-tree plugin (generic, ships with dromos)
    ├── nodeaffinity/       in-tree plugin (generic, ships with dromos)
    └── storage/            in-tree plugin (generic, ships with dromos)
```

Runai-specific plugins (`networking`, `personalcreds`) live in the
[`anyworkload-controller`](../anyworkload-controller) consumer module, not
here.

## Library / consumer architecture

The library produces no binary. Consumers write their own `main.go` that
imports dromos plus the plugins they want and calls `app.RunController` with a
slice of `framework.PluginBuilder`. Run:ai's binary at
[`../anyworkload-controller`](../anyworkload-controller) is one such consumer;
third parties write the same shape in their own repo with their own `go.mod`.

```go
import (
    "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/app"
    "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
    "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/plugins/mnnvl"
    ctrl "sigs.k8s.io/controller-runtime"
)

func main() {
    builders := []framework.PluginBuilder{mnnvl.NewBuilder(mnnvl.Options{ /* host conventions */ })}
    if err := app.RunController(ctrl.SetupSignalHandler(), builders); err != nil {
        os.Exit(1)
    }
}
```

While dromos lives inside the runai monorepo, the local replace directive in
the consumer's `go.mod` stands in for the future OSS module path. Once dromos
is published that replace drops away and consumers resolve through the public
module path.

## Writing a plugin

The full guide is in [`PLUGIN_AUTHORING.md`](PLUGIN_AUTHORING.md). The
in-tree plugins ([`plugins/mnnvl/`](plugins/mnnvl),
[`plugins/storage/`](plugins/storage), [`plugins/nodeaffinity/`](plugins/nodeaffinity),
and [`networking`](../anyworkload-controller/plugins/networking) in the consumer)
are the canonical references.

### Plugin shape

```go
package mycool

import (
    "context"

    "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
    "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
    "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const PluginName = "my-cool-plugin"

type Plugin struct{}

// New is the framework.PluginBuilder.
func New(_ context.Context, _ framework.PluginDeps) (framework.Plugin, error) {
    return &Plugin{}, nil
}

func (p *Plugin) Name() string { return PluginName }

func (p *Plugin) WebhookMutate(_ context.Context, runner *v1alpha1.WorkloadRunner) (admission.Warnings, error) {
    if runner.Labels == nil { runner.Labels = map[string]string{} }
    runner.Labels["myorg.com/cool"] = "true"
    return nil, nil
}

var _ framework.MutatorPlugin = &Plugin{}
```

A single struct can implement any subset of the 6 interfaces - the framework
type-asserts each one at startup and routes hooks accordingly. A builder may
return `(nil, framework.ErrPluginDisabled)` to opt out at start time.

> **Security note.** Plugins run **in-process** with the same RBAC as the
> controller. A registered plugin can read and mutate any object the
> controller can reach, including the admission webhook path. **Only register
> plugins you trust.**

## The 6 hook interfaces

| Interface | When invoked | Use for |
|---|---|---|
| `WatcherPlugin` | controller startup | register watches on related K8s resources |
| `ValidatorPlugin` | admission webhook | reject invalid `WorkloadRunner` specs |
| `MutatorPlugin` | admission webhook | mutate `WorkloadRunner` before persistence |
| `ResourceMutatePlugin` | reconcile, multi-round | mutate the rendered child manifest until convergence |
| `AdditionalResourcePlugin` | reconcile | create supporting K8s resources (Services, ConfigMaps, etc.) |
| `StatusPlugin` | reconcile | report plugin-specific conditions and phase onto status |

## API stability

The plugin SDK at `pkg/plugin/framework` is currently **v0**. The interface
set, struct layouts, and function signatures may break across minor releases
until the disclaimer in `pkg/plugin/framework/doc.go` is removed. Pin a
specific version in your `go.mod` and re-vet on upgrade.

## References

- [`PLUGIN_AUTHORING.md`](PLUGIN_AUTHORING.md) - full guide.
- [`plugins/mnnvl/`](plugins/mnnvl), [`plugins/storage/`](plugins/storage),
  [`plugins/nodeaffinity/`](plugins/nodeaffinity),
  [`../anyworkload-controller/plugins/networking/`](../anyworkload-controller/plugins/networking) - working in-tree plugins.
- [`pkg/plugin/framework/interface.go`](pkg/plugin/framework/interface.go) - the 6 plugin interfaces + PluginBuilder.
- [`pkg/app/app.go`](pkg/app/app.go) - `RunController()`, the boot sequence.
