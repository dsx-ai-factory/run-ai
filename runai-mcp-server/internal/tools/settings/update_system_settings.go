package settings

import (
	"context"
	"fmt"
	"strings"

	authz "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/common"
)

type updateSystemSettingsInput struct {
	Category string `json:"category"          jsonschema:"Settings group. One of: workloads, authorization, cluster, org-unit, catalog, datavolumes, node-pool, workloads-manager."`
	Key      string `json:"key"               jsonschema:"The setting to change; must be a key of the chosen category (see list_system_settings)."`
	Enabled  *bool  `json:"enabled"           jsonschema:"The new value for the boolean setting. Required."`
	DryRun   bool   `json:"dryRun,omitempty"  jsonschema:"When true, preview the before/after without mutating anything upstream."`
	Confirm  bool   `json:"confirm,omitempty" jsonschema:"Required for a non-dry-run change when the client can't prompt. Without it the call returns confirmation_required with the current value."`
}

func (in updateSystemSettingsInput) Validate() error {
	if !validCategory(in.Category) {
		return fmt.Errorf("category must be one of: %s", strings.Join(categories(), ", "))
	}
	if !validKey(in.Category, in.Key) {
		return fmt.Errorf("key %q is not a %s setting; valid keys: %s", in.Key, in.Category, strings.Join(keyNames(in.Category), ", "))
	}
	if !keyUpdatable(in.Category, in.Key) {
		return fmt.Errorf("setting %q is read-only (a provisioning-owned attribute, not an operator-toggleable feature flag)", in.Key)
	}
	if in.Enabled == nil {
		return fmt.Errorf("enabled is required")
	}
	return nil
}

type updateSystemSettingsOutput struct {
	Category string         `json:"category"           jsonschema:"The settings group changed."`
	Key      string         `json:"key"                jsonschema:"The setting key changed."`
	DryRun   bool           `json:"dryRun"             jsonschema:"True when this was a preview; no upstream change was made."`
	Applied  bool           `json:"applied"            jsonschema:"True when the change was actually applied upstream. False on dry-run."`
	Before   map[string]any `json:"before"             jsonschema:"Value before the change ({enabled: <bool|null>}); null when the platform had no value set."`
	After    map[string]any `json:"after"              jsonschema:"Value after the change ({enabled: <bool>})."`
	Diff     map[string]any `json:"diff"               jsonschema:"Field-level change as path -> {before, after}."`
	Warnings []string       `json:"warnings,omitempty" jsonschema:"Non-fatal notes."`
}

func registerUpdateSystemSettings(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[updateSystemSettingsInput, updateSystemSettingsOutput]{
		Name: "update_system_settings",
		Description: "Turn a tenant-wide Run:ai platform setting on or off. The category field selects the settings group " +
			"(workloads | authorization | cluster | org-unit | catalog | datavolumes | node-pool | workloads-manager); key " +
			"selects the specific boolean setting; enabled is the new value. For high-impact changes, call once with " +
			"dryRun=true to preview the before/after, then again with dryRun=false and confirm=true to apply — these are " +
			"tenant-wide. Use list_system_settings to discover keys and current values. Covers boolean feature-flags only " +
			"(not security or UI settings).",
		ReadOnly:    false,
		Destructive: true,
		Handler: func(ctx context.Context, in updateSystemSettingsInput) (updateSystemSettingsOutput, error) {
			return updateSystemSettingsRun(ctx, client, in)
		},
	})
}

func updateSystemSettingsRun(ctx context.Context, client *runaiclient.Client, in updateSystemSettingsInput) (updateSystemSettingsOutput, error) {
	scope, err := common.PermittedScopeForAction(ctx, client, authz.RESOURCETYPE_SETTINGS, authz.ACTION_UPDATE)
	if err != nil {
		return updateSystemSettingsOutput{}, err
	}
	if !common.ScopeAllowsTenant(scope) {
		return updateSystemSettingsOutput{}, fmt.Errorf("not permitted to update system settings in this tenant: %w", runaimcp.ErrForbidden)
	}

	before, err := readSetting(ctx, client, in.Category, in.Key)
	if err != nil {
		return updateSystemSettingsOutput{}, err
	}

	out := updateSystemSettingsOutput{
		Category: in.Category,
		Key:      in.Key,
		DryRun:   in.DryRun,
		Before:   map[string]any{"enabled": boolOrNil(before)},
		After:    map[string]any{"enabled": *in.Enabled},
	}
	out.Diff = common.DiffMaps(out.Before, out.After)

	if in.DryRun {
		return out, nil
	}
	if err := common.ConfirmWrite(ctx, in.Confirm, "update", "system setting", in.Category+"."+in.Key, in.Key, out.Diff); err != nil {
		return updateSystemSettingsOutput{}, err
	}

	if err := updateSetting(ctx, client, in.Category, in.Key, *in.Enabled); err != nil {
		return updateSystemSettingsOutput{}, err
	}
	out.Applied = true

	// Report the value the platform actually holds after the write — services may normalize or
	// ignore a value, so don't trust the intended value as the result.
	after, err := readSetting(ctx, client, in.Category, in.Key)
	if err != nil {
		out.Warnings = append(out.Warnings, "change applied, but re-reading the new value failed: "+err.Error())
	} else {
		out.After = map[string]any{"enabled": boolOrNil(after)}
	}
	out.Diff = common.DiffMaps(out.Before, out.After)
	return out, nil
}

func boolOrNil(b *bool) any {
	if b == nil {
		return nil
	}
	return *b
}
