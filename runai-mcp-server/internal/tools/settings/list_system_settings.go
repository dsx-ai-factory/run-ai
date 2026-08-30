package settings

import (
	"context"
	"fmt"
	"strings"
	"sync"

	authz "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/common"
)

type listSystemSettingsInput struct {
	Category string `json:"category,omitempty" jsonschema:"Optional filter to one settings group: workloads, authorization, cluster, org-unit, catalog, datavolumes, node-pool, workloads-manager. Omit to return all."`
}

func (in listSystemSettingsInput) Validate() error {
	if in.Category != "" && !validCategory(in.Category) {
		return fmt.Errorf("category must be one of: %s", strings.Join(categories(), ", "))
	}
	return nil
}

type systemSetting struct {
	Category    string  `json:"category" jsonschema:"Settings group: workloads, authorization, cluster, org-unit, catalog, datavolumes, node-pool, or workloads-manager."`
	Key         string  `json:"key"      jsonschema:"Setting key — pass to update_system_settings."`
	Enabled     *bool   `json:"enabled"  jsonschema:"Current value. Null when the platform has no value set for this key."`
	DisplayName *string `json:"display_name,omitempty" jsonschema:"Human-readable setting name, when the platform provides one."`
	Description *string `json:"description,omitempty"   jsonschema:"What the setting controls, when the platform provides it."`
	Stage       *string `json:"stage,omitempty"         jsonschema:"Lifecycle stage: GA, Beta, Alpha, or Legacy. Absent when the platform does not report it; treat a non-GA setting with care and confirm before toggling."`
}

type listSystemSettingsOutput struct {
	Settings []systemSetting `json:"settings"           jsonschema:"Current tenant-wide boolean settings. Always present; empty array when none match."`
	Warnings []string        `json:"warnings,omitempty" jsonschema:"Categories that could not be read (e.g. a service was unavailable or forbidden) and were skipped; the rest are still returned."`
}

func registerListSystemSettings(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[listSystemSettingsInput, listSystemSettingsOutput]{
		Name: "list_system_settings",
		Description: "List tenant-wide Run:ai platform settings and their current values — the boolean feature-flags " +
			"across workloads, authorization, cluster, org-unit, catalog, datavolumes, node-pool, and workloads-manager. " +
			"Call this to discover which settings exist and their state before changing one with update_system_settings, or " +
			"to confirm a change. Optionally filter by category. Does NOT cover security or UI settings.",
		ReadOnly: true,
		Handler: func(ctx context.Context, in listSystemSettingsInput) (listSystemSettingsOutput, error) {
			return listSystemSettingsRun(ctx, client, in)
		},
	})
}

func listSystemSettingsRun(ctx context.Context, client *runaiclient.Client, in listSystemSettingsInput) (listSystemSettingsOutput, error) {
	scope, err := common.PermittedScopeForAction(ctx, client, authz.RESOURCETYPE_SETTINGS, authz.ACTION_READ)
	if err != nil {
		return listSystemSettingsOutput{}, err
	}
	if !common.ScopeAllowsTenant(scope) {
		return listSystemSettingsOutput{}, fmt.Errorf("not permitted to read system settings in this tenant: %w", runaimcp.ErrForbidden)
	}

	cats := categories()
	if in.Category != "" {
		cats = []string{in.Category}
	}

	// Read categories concurrently — they hit independent services. A category that fails to
	// read (a service down or forbidden) degrades to a warning instead of aborting the whole
	// listing; each goroutine writes its own slot, so no shared-state synchronization is needed.
	type catResult struct {
		settings []systemSetting
		warning  string
	}
	results := make([]catResult, len(cats))
	var wg sync.WaitGroup
	for i, cat := range cats {
		wg.Add(1)
		go func(i int, cat string) {
			defer wg.Done()
			byKey, err := readCategorySettings(ctx, client, cat)
			if err != nil {
				results[i] = catResult{warning: fmt.Sprintf("category %q could not be read and was skipped: %v", cat, err)}
				return
			}
			var rows []systemSetting
			for _, key := range keyNames(cat) {
				v := byKey[responseKey(cat, key)]
				rows = append(rows, systemSetting{
					Category:    cat,
					Key:         key,
					Enabled:     v.enabled,
					DisplayName: v.displayName,
					Description: v.description,
					Stage:       v.stage,
				})
			}
			results[i] = catResult{settings: rows}
		}(i, cat)
	}
	wg.Wait()

	out := listSystemSettingsOutput{Settings: []systemSetting{}}
	for _, r := range results {
		if r.warning != "" {
			out.Warnings = append(out.Warnings, r.warning)
			continue
		}
		out.Settings = append(out.Settings, r.settings...)
	}
	return out, nil
}
