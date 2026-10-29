/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package installconfig

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/artifacts"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
)

// RunOptions are the CLI-flag-derived knobs that map into the orchestrator spec.
// Cross-cutting execution knobs (timeouts, confirmation) live in
// specrun.ExecOptions, not here.
type RunOptions struct {
	DryRun        bool
	PreflightOnly bool // accepted; not yet modeled in the plan (TODO)
	SkipPreflight bool // accepted; not yet modeled in the plan (TODO)
	Yes           bool
	Output        spec.OutputFormat // "" => auto-resolved by specrun via TTY
	EventsFile    string
	StatusFile    string // accepted; status writer deferred (TODO)
	OnFailure     spec.FailurePolicy
	Env           spec.Env
	Registry      *spec.Registry
	// DevMode relaxes input validation (e.g. allows an out-of-matrix Run:ai
	// version). Escape hatch for development/testing.
	DevMode bool
}

// ToSpec maps the config file plus the run options into the orchestrator spec.
func (c *Config) ToSpec(opts RunOptions) *spec.Spec {
	s := &spec.Spec{
		Command:                spec.CommandInstall,
		InstallMode:            toSpecInstallMode(c.Install.Mode),
		DryRun:                 opts.DryRun,
		NonInteractive:         opts.Yes,
		DevMode:                opts.DevMode,
		OnStepFailure:          opts.OnFailure,
		Connectivity:           spec.Connectivity(c.Install.Connectivity),
		Env:                    opts.Env,
		KubernetesDistribution: toSpecKubernetesDistribution(c.Install.KubernetesDistribution),
		Output: spec.Output{
			Format:     opts.Output,
			EventsFile: opts.EventsFile,
		},
		Registry:      opts.Registry,
		TLS:           toSpecTLS(c.TLS),
		SelfSignedTLS: toSpecSelfSignedTLS(c),
	}

	if c.Airgap != nil {
		valuesDir := c.Airgap.ValuesDir
		if valuesDir == "" {
			valuesDir = c.Airgap.WorkingDir
		}
		s.Airgap = &spec.Airgap{
			BundlePath:  c.Airgap.BundlePath,
			WorkingDir:  c.Airgap.WorkingDir,
			ValuesDir:   valuesDir,
			SkipPrepare: c.Airgap.SkipPrepare,
		}
	}

	if c.ControlPlane != nil && c.ControlPlane.Enabled {
		s.ControlPlane = &spec.ControlPlane{
			Domain:           c.ControlPlane.Domain,
			Version:          c.ControlPlane.Version,
			TLS:              toSpecTLS(c.ControlPlane.TLS),
			CustomHelmValues: toStringMap(c.ControlPlane.CustomHelmValues),
		}
	}

	if c.Cluster != nil && c.Cluster.Enabled {
		s.Cluster = &spec.Cluster{
			Name:                    c.Cluster.Name,
			Domain:                  c.Cluster.Domain,
			Version:                 c.Cluster.Version,
			WorkloadRouting:         c.Cluster.WorkloadRouting,
			TLS:                     toSpecTLS(c.Cluster.TLS),
			Inference:               toSpecInference(c.Cluster.Inference),
			ControlPlaneURL:         c.Cluster.ControlPlaneURL,
			UID:                     c.Cluster.UID,
			BootstrapToken:          c.Cluster.bootstrapToken,
			ClientSecret:            c.Cluster.clientSecret,
			ControlPlaneCredentials: toSpecControlPlaneCredentials(c.Cluster.ControlPlaneCredentials),
			CustomHelmValues:        toStringMap(c.Cluster.CustomHelmValues),
		}
	}

	if s.UsesAirgapLoad() && s.Airgap != nil && s.Registry != nil {
		destination := artifacts.NormalizeTarget(s.Registry.Host)
		s.AirgapLoad = &spec.AirgapLoad{
			BundlePath:  s.Airgap.BundlePath,
			WorkingDir:  s.Airgap.WorkingDir,
			ValuesDir:   s.Airgap.ValuesDir,
			Destination: destination,
			Credential: spec.RegistryCredential{
				Prefix:             destination,
				Username:           s.Registry.Username,
				Password:           s.Registry.Password,
				CAChainPath:        s.Registry.CAChainPath,
				Insecure:           s.Registry.Insecure,
				InsecureSkipVerify: s.Registry.InsecureSkipVerify,
			},
		}
	}

	if c.Prerequisites != nil {
		gpu := c.Prerequisites.GPUOperator
		s.Prerequisites = &spec.Prerequisites{
			Skip:            c.Prerequisites.Skip,
			Prometheus:      toSpecDependency(c.Prerequisites.Prometheus, true),
			GPUOperator:     toSpecDependency(gpu, true),
			GPUType:         toSpecGPUType(gpu),
			NetworkOperator: toSpecDependency(c.Prerequisites.NetworkOperator, true),
		}
	}

	if c.Features != nil {
		s.Features = &spec.Features{
			Skip:        c.Features.Skip,
			Knative:     toSpecDependency(c.Features.Knative, false),
			Kubeflow:    toSpecDependency(c.Features.Kubeflow, false),
			MPIOperator: toSpecDependency(c.Features.MPIOperator, false),
			LWS:         toSpecDependency(c.Features.LWS, false),
			NIMOperator: toSpecDependency(c.Features.NIMOperator, false),
			Grove:       toSpecDependency(c.Features.Grove, false),
			Dynamo:      toSpecDependency(c.Features.Dynamo, false),
		}
	}

	s.ExternalConnection = toSpecExternalConnection(c.ExternalConnection)

	s.SystemNodes = toSpecSystemNodes(c.SystemNodes)

	s.OpenShift = toSpecOpenShift(c.OpenShift)

	return s
}

func toSpecOpenShift(o *OpenShift) *spec.OpenShift {
	if o == nil || o.OLMCatalog == nil {
		return nil
	}

	return &spec.OpenShift{
		OLMCatalogSource:          o.OLMCatalog.Source,
		OLMCatalogSourceNamespace: o.OLMCatalog.SourceNamespace,
	}
}

const defaultSystemNodeLabel = "node-role.kubernetes.io/runai-system"

// toSpecSystemNodes maps the system-node placement config to the spec, applying
// defaults (mode=preferred, manageLabels=true, label=runai-system). A nil or
// disabled block leaves the feature off (nil spec).
func toSpecSystemNodes(c *SystemNodes) *spec.SystemNodes {
	if c == nil || !c.Enabled {
		return nil
	}
	mode := spec.SystemNodeMode(c.Mode)
	if mode == "" {
		mode = spec.SystemNodeModePreferred
	}
	manage := true
	if c.ManageLabels != nil {
		manage = *c.ManageLabels
	}
	label := c.Label
	if label == "" {
		label = defaultSystemNodeLabel
	}
	return &spec.SystemNodes{
		Mode:         mode,
		Label:        label,
		ManageLabels: manage,
		Nodes:        c.Nodes,
		Selector:     c.Selector,
	}
}

// toSpecInstallMode maps the config's install.mode string to the spec enum,
// defaulting an empty value to full installation.
func toSpecInstallMode(mode string) spec.InstallMode {
	if mode == "" {
		return spec.InstallModeFull
	}
	return spec.InstallMode(mode)
}

func toSpecKubernetesDistribution(d string) string {
	if d == "vanilla" {
		return ""
	}
	return d
}

// toSpecDependency maps a config dependency to the spec. A nil block is left
// unset (no implicit enable). When the block is present but its mode is omitted,
// defaultOn decides: mandatory dependencies default to install, optional ones to
// skip.
func toSpecDependency(d *Dependency, defaultOn bool) *spec.Dependency {
	if d == nil {
		return nil
	}
	mode := spec.DependencyMode(d.Mode)
	if mode == "" {
		mode = spec.DependencyModeSkip
		if defaultOn {
			mode = spec.DependencyModeInstall
		}
	}
	out := &spec.Dependency{
		Mode:             mode,
		Version:          d.Version,
		CustomHelmValues: toStringMap(d.CustomHelmValues),
	}
	if d.OpenShift != nil {
		out.OpenShiftInstallPlanApproval = d.OpenShift.InstallPlanApproval
		out.OpenShiftCR = d.OpenShift.CustomResource
		if d.OpenShift.SubscriptionConfig != nil {
			out.OpenShiftSubscriptionEnv = d.OpenShift.SubscriptionConfig.Env
		}
	}

	return out
}

// toSpecGPUType reads the GPU operator flavor, defaulting to real when unset.
// An invalid value is passed through unchanged so input validation can reject it
// rather than silently treating it as real.
func toSpecGPUType(d *Dependency) spec.GPUType {
	if d == nil || d.Type == "" {
		return spec.GPUTypeReal
	}
	return spec.GPUType(d.Type)
}

func toSpecControlPlaneCredentials(cc *ControlPlaneCredentials) *spec.ControlPlaneCredentials {
	if cc == nil {
		return nil
	}
	password, token := cc.password, cc.token
	if password == "" {
		password = cc.Password
	}
	if token == "" {
		token = cc.Token
	}
	return &spec.ControlPlaneCredentials{
		Username: cc.Username,
		Password: password,
		Token:    token,
	}
}

func toSpecInference(i *Inference) *spec.Inference {
	if i == nil {
		return nil
	}
	return &spec.Inference{
		WildcardDomain: i.WildcardDomain,
		WildcardTLS:    toSpecTLS(i.WildcardTLS),
	}
}

func toSpecExternalConnection(e *ExternalConnection) *spec.ExternalConnection {
	if e == nil {
		return nil
	}
	out := &spec.ExternalConnection{Type: spec.ExternalConnectionType(e.Type)}
	if e.Ingress != nil {
		out.Ingress = &spec.IngressConnection{ClassName: e.Ingress.ClassName}
	}
	if e.GatewayAPI != nil {
		out.GatewayAPI = &spec.GatewayAPIConnection{Name: e.GatewayAPI.Name, Namespace: e.GatewayAPI.Namespace}
	}
	return out
}

func toSpecTLS(t *TLS) *spec.TLS {
	if t == nil {
		return nil
	}
	return &spec.TLS{CertPath: t.CertPath, KeyPath: t.KeyPath, CAChainPath: t.CAChainPath}
}

// toStringMap renders the config's customHelmValues into the Helm --set
// arguments the installer eventually passes to the chart. Helm addresses nested
// values with dotted paths, so nested maps are flattened
// (`telemetryService: {enabled: true}` becomes `telemetryService.enabled=true`)
// and lists are rendered in Helm's brace syntax (`{a,b}`). Keys that are
// already dotted are passed through unchanged, so both spellings keep working.
func toStringMap(m map[string]any) map[string]string {
	if len(m) == 0 {
		return nil
	}
	out := make(map[string]string, len(m))
	flattenHelmValues(out, "", m)

	if len(out) == 0 {
		return nil
	}
	return out
}

// flattenHelmValues walks a customHelmValues map, writing one entry per leaf
// value. A leaf is anything that is not a nested map or list.
func flattenHelmValues(out map[string]string, prefix string, m map[string]any) {
	// A key can be written in dotted form (`global.foo`) or as a nested map
	// (`global: {foo: ...}`), and the two can name the same path. Sorting makes
	// the result independent of Go's map iteration order, and since a prefix
	// sorts before the longer key it produces, the key the user spelled out in
	// full is the one that survives.
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		key := k
		if prefix != "" {
			// A key nested in the config is a literal key, so the separators it
			// contains are escaped: `podAnnotations: {"prometheus.io/scrape": x}`
			// must reach the chart as that annotation, not as nested keys. A key
			// written in dotted form at the top level stays a path on purpose.
			key = prefix + "." + escapeHelmKey(k)
		}

		flattenHelmValue(out, key, m[k])
	}
}

func flattenHelmValue(out map[string]string, key string, v any) {
	switch t := v.(type) {
	case map[string]any:
		// An empty map has no leaves. Emitting `key=` would clear the chart
		// default rather than leave it alone, so skip it.
		if len(t) > 0 {
			flattenHelmValues(out, key, t)
		}
	case map[string]string:
		if len(t) > 0 {
			nested := make(map[string]any, len(t))
			for k, s := range t {
				nested[k] = s
			}

			flattenHelmValues(out, key, nested)
		}
	case []any:
		flattenHelmList(out, key, t)
	case []string:
		items := make([]any, 0, len(t))
		for _, s := range t {
			items = append(items, s)
		}

		flattenHelmList(out, key, items)
	case nil:
		// Helm parses a bare `null` as an empty value; `%v` would send it the
		// literal "<nil>" instead.
		out[key] = "null"
	default:
		// Helm parses what it is given, so a value that reads as a number, a
		// boolean or null reaches the chart as that type. There is no way to
		// force a string inside a --set argument; that needs --set-string
		// (strvals' ParseIntoString) from the caller.
		out[key] = escapeHelmValue(fmt.Sprintf("%v", v))
	}
}

// flattenHelmList renders a list the way Helm's --set can address it. A list of
// scalars uses the brace form; as soon as an element is itself a map or a list,
// the brace form would hand the chart a Go map string, so the entries are
// addressed by index instead (`key[0].sub=...`, `key[0][1]=...`).
//
// An empty list emits nothing: Helm has no empty-list form (`key={}` is a list
// holding one empty string) and leaving the key out keeps the chart default,
// exactly as an empty map does.
func flattenHelmList(out map[string]string, key string, items []any) {
	if len(items) == 0 {
		return
	}

	if !helmListNeedsIndexes(items) {
		out[key] = helmListValue(items)
		return
	}

	for i, item := range items {
		flattenHelmValue(out, fmt.Sprintf("%s[%d]", key, i), item)
	}
}

// helmListNeedsIndexes reports whether an element is a map or a list, which
// Helm's brace syntax cannot express.
func helmListNeedsIndexes(items []any) bool {
	for _, item := range items {
		switch item.(type) {
		case map[string]any, map[string]string, []any, []string:
			return true
		}
	}

	return false
}

// helmListValue renders a list in the brace form Helm's --set parser expects.
// Elements are escaped so a comma or a closing brace inside one is not read as
// the end of the list.
func helmListValue(items []any) string {
	parts := make([]string, 0, len(items))

	for _, item := range items {
		if item == nil {
			parts = append(parts, "null")
			continue
		}

		parts = append(parts, escapeHelmValue(fmt.Sprintf("%v", item)))
	}

	return "{" + strings.Join(parts, ",") + "}"
}

// escapeHelmValue backslash-escapes the characters Helm's --set parser treats
// as separators, so a value containing them survives the round trip: a comma
// separates list elements (and is an error on its own, "a,b" does not parse)
// and a closing brace ends a list.
func escapeHelmValue(s string) string {
	return strings.NewReplacer(
		`\`, `\\`,
		`,`, `\,`,
		`{`, `\{`,
		`}`, `\}`,
	).Replace(s)
}

// escapeHelmKey backslash-escapes the characters Helm's --set parser treats as
// key separators, so a nested key that contains one stays a single key instead
// of splitting into path segments: `.`, `,`, `=` and `[` all end a key there.
func escapeHelmKey(s string) string {
	return strings.NewReplacer(
		`\`, `\\`,
		`.`, `\.`,
		`,`, `\,`,
		`=`, `\=`,
		`[`, `\[`,
	).Replace(s)
}
