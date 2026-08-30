# Plugin Authoring Guide

This guide explains how to write a plugin for the dromos controller. The same
pattern is used by the in-tree plugins (`plugins/mnnvl/`, `plugins/storage/`,
`plugins/nodeaffinity/`, and `../anyworkload-controller/plugins/networking/`) and by third-party
plugins shipped from external repos.

The reference plugins live at [`plugins/mnnvl/`](plugins/mnnvl),
[`plugins/storage/`](plugins/storage), [`plugins/nodeaffinity/`](plugins/nodeaffinity),
and [`../anyworkload-controller/plugins/networking/`](../anyworkload-controller/plugins/networking).
In-tree plugins are sub-packages of their host module - third-party plugins
live in their own repo with their own `go.mod`. Copy the shape of any in-tree
plugin to start your own.

## Mental model

dromos exposes 6 hook points where a plugin can participate. A plugin
implements as many or as few of these as it needs:

| Interface | When invoked | Typical use |
|---|---|---|
| `WatcherPlugin` | controller startup | register watches on related K8s resources (Services, ConfigMaps, etc.) |
| `ValidatorPlugin` | admission webhook | reject invalid `WorkloadRunner` specs |
| `MutatorPlugin` | admission webhook | mutate `WorkloadRunner` before persistence |
| `ResourceMutatePlugin` | reconcile, multi-round | mutate the rendered child manifest until convergence |
| `AdditionalResourcePlugin` | reconcile | create supporting K8s resources (Services, auth proxies, ComputeDomains, etc.) |
| `StatusPlugin` | reconcile | report plugin-specific conditions and phase onto `WorkloadRunner.status` |

The interface definitions are in
[`pkg/plugin/framework/interface.go`](pkg/plugin/framework/interface.go).

## Anatomy of a plugin

A plugin is a Go type with:

1. A `Name() string` method - the unique identifier used in registration and status.
2. One or more of the optional interface methods listed above.
3. A constructor matching the `framework.PluginBuilder` signature:

   ```go
   func New(ctx context.Context, deps framework.PluginDeps) (framework.Plugin, error)
   ```

   - `deps` carries framework-supplied dependencies (`Client`, `Logger`).
   - Return `(nil, framework.ErrPluginDisabled)` to opt out at start time
     (typically because the plugin's own config says disabled).

Per-plugin configuration is loaded by the plugin itself - typically a typed
struct registered with `runai-common-packages/config` via
`rcpconfig.RegisterRoot[*MyConfig]()` and read back inside `New(...)`. See
`plugins/mnnvl/config.go` for a reference.

The framework discovers which optional interfaces a plugin satisfies via Go type
assertions, so a single struct can implement several interfaces and the framework will
route each hook to the correct method automatically.

## Registering and running

In your custom `main.go`:

```go
package main

import (
    "os"

    "github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/app"
    "github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/plugin/framework"
    "github.com/run-ai/runai/runai-cluster/cluster/dromos/plugins/mnnvl"
    myplugin "github.com/yourorg/yourrepo/myplugin"
    ctrl "sigs.k8s.io/controller-runtime"
)

func main() {
    builders := []framework.PluginBuilder{
        mnnvl.NewBuilder(mnnvl.Options{ /* host conventions, injected by the consumer */ }),
        myplugin.New,
    }
    if err := app.RunController(ctrl.SetupSignalHandler(), builders); err != nil {
        os.Exit(1)
    }
}
```

Build a binary, ship it as a Docker image, point your Helm chart at it. That's the
whole loop.

## Worked example: a simple MutatorPlugin

```go
package hello

import (
    "context"

    "github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/plugin/framework"
    "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
    "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type Plugin struct{ greeting string }

func New(_ context.Context, _ framework.PluginDeps) (framework.Plugin, error) {
    return &Plugin{greeting: "world"}, nil
}

func (p *Plugin) Name() string { return "hello-plugin" }

func (p *Plugin) WebhookMutate(_ context.Context, runner *v1alpha1.WorkloadRunner) (admission.Warnings, error) {
    if runner.Labels == nil { runner.Labels = map[string]string{} }
    runner.Labels["anyworkload.run.ai/hello"] = p.greeting
    return nil, nil
}

var _ framework.MutatorPlugin = &Plugin{}
```

## Status reporting

`StatusPlugin.Status(...)` receives a `PluginStatusSetter`. Use it to set conditions
and the per-plugin phase on the `WorkloadRunner` status:

```go
func (p *Plugin) Status(ctx context.Context, runner *v1alpha1.WorkloadRunner,
    k8sClient client.Client, status framework.PluginStatusSetter, state *reconcile.State) error {

    if everythingOK {
        status.SetCondition(metav1.Condition{
            Type:   "Ready",
            Status: metav1.ConditionTrue,
            Reason: "Healthy",
        })
        status.SetPhase(v1alpha1.PluginPhaseReady)
        return nil
    }
    // ...
}
```

For non-fatal issues during reconciliation, plugins can also report into `*reconcile.State`
via `state.AddPluginIssue(...)` so they surface in `WorkloadRunner.status.plugins[name].issues`.

## API stability

`pkg/plugin/framework` is currently **v0**. We may break it across minor versions until
the disclaimer in `pkg/plugin/framework/doc.go` is removed. Pin to a specific version
of dromos in your `go.mod` and re-vet your plugin on each upgrade.

## Reference

- [`pkg/plugin/framework/interface.go`](pkg/plugin/framework/interface.go) - the 6 plugin interfaces and `PluginBuilder`.
- [`pkg/app/app.go`](pkg/app/app.go) - `RunController()`, the boot sequence.
- [`plugins/mnnvl/`](plugins/mnnvl), [`plugins/storage/`](plugins/storage),
  [`../anyworkload-controller/plugins/networking/`](../anyworkload-controller/plugins/networking) - full working in-tree plugins.
