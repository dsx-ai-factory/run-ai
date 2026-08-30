package metrics

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/rs/zerolog"

	orgunit "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/common"
)

// Per-label variants collapse with sum for allocation, average for utilization.
type seriesKind struct {
	apiType orgunit.OrgUnitMetricType
	divisor float64
	sum     bool
}

const (
	metricGPUAllocation        = "gpuAllocation"
	metricCPUAllocation        = "cpuAllocation"
	metricCPUMemoryAllocation  = "cpuMemoryAllocation"
	metricGPUUtilization       = "gpuUtilization"
	metricGPUMemoryUtilization = "gpuMemoryUtilization"
	metricCPUUtilization       = "cpuUtilization"
	metricCPUMemoryUtilization = "cpuMemoryUtilization"
)

var orgMetricCatalog = map[string]seriesKind{
	metricGPUAllocation:        {orgunit.ORGUNITMETRICTYPE_GPU_ALLOCATION, 1, true},
	metricCPUAllocation:        {orgunit.ORGUNITMETRICTYPE_CPU_ALLOCATION_MILLICORES, common.MillicoresPerCore, true},
	metricCPUMemoryAllocation:  {orgunit.ORGUNITMETRICTYPE_CPU_MEMORY_ALLOCATION_MB, common.MbPerGiB, true},
	metricGPUUtilization:       {orgunit.ORGUNITMETRICTYPE_GPU_UTILIZATION, 1, false},
	metricGPUMemoryUtilization: {orgunit.ORGUNITMETRICTYPE_GPU_MEMORY_UTILIZATION, 1, false},
	metricCPUUtilization:       {orgunit.ORGUNITMETRICTYPE_CPU_UTILIZATION, 1, false},
	metricCPUMemoryUtilization: {orgunit.ORGUNITMETRICTYPE_CPU_MEMORY_UTILIZATION, 1, false},
}

var defaultMetricTypes = []string{metricGPUAllocation, metricGPUUtilization, metricGPUMemoryUtilization}

const defaultWindow = "last_24h"

type getOrgUnitMetricsInput struct {
	OrgType         string   `json:"orgType"                   jsonschema:"Which kind of org unit to query. One of: project, department."`
	OrgID           string   `json:"orgId,omitempty"           jsonschema:"Project or department ID. Mutually exclusive with orgName; one of the two is required."`
	OrgName         string   `json:"orgName,omitempty"         jsonschema:"Project or department name. Mutually exclusive with orgId; resolved to an ID. Returns ambiguous_org_unit if the name matches multiple."`
	ClusterID       string   `json:"clusterId,omitempty"       jsonschema:"Cluster ID. Used to disambiguate orgName resolution; if omitted, name resolution spans all accessible clusters."`
	MetricTypes     []string `json:"metricTypes,omitempty"     jsonschema:"Metrics to return; allocation/utilization for gpu, cpu, cpuMemory, plus gpuMemoryUtilization. Default: gpu trio."`
	Window          string   `json:"window,omitempty"          jsonschema:"Convenience time range ending now. One of: last_1h, last_24h, last_7d, last_30d. Defaults to last_24h. Ignored when start and end are both set."`
	Start           string   `json:"start,omitempty"           jsonschema:"Explicit range start (RFC3339). Overrides window; must be paired with end."`
	End             string   `json:"end,omitempty"             jsonschema:"Explicit range end (RFC3339). Overrides window; must be paired with start."`
	NumberOfSamples int32    `json:"numberOfSamples,omitempty" jsonschema:"Number of evenly spaced points across the range. Defaults to 20, capped at 1000."`
	NodePoolName    string   `json:"nodePoolName,omitempty"    jsonschema:"Scope the series to a single node pool by name. If omitted, spans the whole org unit."`
}

func (in getOrgUnitMetricsInput) Validate() error {
	if in.OrgType != common.OrgTypeProject && in.OrgType != common.OrgTypeDepartment {
		return fmt.Errorf("orgType must be one of: %s, %s", common.OrgTypeProject, common.OrgTypeDepartment)
	}
	if (in.OrgID == "") == (in.OrgName == "") {
		return fmt.Errorf("exactly one of orgId or orgName is required")
	}
	for _, m := range in.MetricTypes {
		if _, ok := orgMetricCatalog[m]; !ok {
			return fmt.Errorf("unknown metric type %q", m)
		}
	}
	if in.Window != "" {
		if _, ok := windowDurations[in.Window]; !ok {
			return fmt.Errorf("window must be one of: last_1h, last_24h, last_7d, last_30d")
		}
	}
	if (in.Start == "") != (in.End == "") {
		return fmt.Errorf("start and end must be provided together")
	}
	if err := common.ValidateTimeWindow(in.Start, in.End); err != nil {
		return err
	}
	if in.NumberOfSamples < 0 || in.NumberOfSamples > maxSamples {
		return fmt.Errorf("numberOfSamples must be between 0 and %d", maxSamples)
	}
	if err := common.ValidateNoCommas("orgName", in.OrgName); err != nil {
		return err
	}
	return common.ValidateNoCommas("clusterId", in.ClusterID)
}

type orgUnitSeries struct {
	GPUAllocation        []point `json:"gpuAllocation,omitempty"        jsonschema:"GPU allocation over time, in GPU device count."`
	CPUAllocation        []point `json:"cpuAllocation,omitempty"        jsonschema:"CPU allocation over time, in cores (normalized from millicores)."`
	CPUMemoryAllocation  []point `json:"cpuMemoryAllocation,omitempty"  jsonschema:"System memory allocation over time, in GiB (normalized from MB)."`
	GPUUtilization       []point `json:"gpuUtilization,omitempty"       jsonschema:"GPU utilization over time, in percent (0-100)."`
	GPUMemoryUtilization []point `json:"gpuMemoryUtilization,omitempty" jsonschema:"GPU memory utilization over time, in percent (0-100)."`
	CPUUtilization       []point `json:"cpuUtilization,omitempty"       jsonschema:"CPU utilization over time, in percent (0-100)."`
	CPUMemoryUtilization []point `json:"cpuMemoryUtilization,omitempty" jsonschema:"System memory utilization over time, in percent (0-100)."`
}

type getOrgUnitMetricsOutput struct {
	OrgType         string         `json:"orgType"           jsonschema:"The org-unit kind queried: project or department."`
	OrgID           string         `json:"orgId"             jsonschema:"The resolved org-unit ID."`
	OrgName         string         `json:"orgName,omitempty" jsonschema:"The org-unit name, when known (echoed from orgName input or resolution)."`
	NodePool        *string        `json:"nodePool"          jsonschema:"Node pool the series is scoped to, or null when spanning the whole org unit."`
	TimeRange       timeRange      `json:"timeRange"         jsonschema:"The time range and sample count the series cover."`
	Series          *orgUnitSeries `json:"series,omitempty"  jsonschema:"Time-series per requested metric. Omitted entirely when seriesTruncated is true."`
	SeriesTruncated bool           `json:"seriesTruncated"   jsonschema:"True when the request was too large and series were dropped; retry with a smaller numberOfSamples or fewer metricTypes."`
	PartialData     bool           `json:"partialData"       jsonschema:"True when the upstream returned partial results (some metrics failed); failed metrics are absent from series."`
}

func registerGetOrgUnitMetrics(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[getOrgUnitMetricsInput, getOrgUnitMetricsOutput]{
		Name: "get_org_unit_metrics",
		Description: "Get a project's or department's resource consumption over time — allocation and utilization trends for GPU, CPU, and memory. " +
			"Use when the user asks how an org unit has been using resources over a period (e.g. \"how has my team's GPU utilization trended this week?\"). " +
			"Returns time-series points per requested metric. " +
			"Use get_workload_metrics for one workload and list_workloads for per-workload allocated or requested GPU-count filters. " +
			"Resolve an ambiguous org-unit name with cluster context. " +
			"This aggregate tool cannot identify individual workloads above a utilization threshold. " +
			"Do NOT call for configured quota (use list_project_resources / list_department_resources), for a single current snapshot, " +
			"for cluster-wide metrics (use get_cluster_metrics), or for per-workload metrics.",
		ReadOnly: true,
		Handler: func(ctx context.Context, in getOrgUnitMetricsInput) (getOrgUnitMetricsOutput, error) {
			return getOrgUnitMetricsRun(ctx, client, in)
		},
	})
}

func getOrgUnitMetricsRun(ctx context.Context, client *runaiclient.Client, in getOrgUnitMetricsInput) (getOrgUnitMetricsOutput, error) {
	orgID := in.OrgID
	orgName := in.OrgName
	if orgID == "" {
		id, err := common.ResolveOrgUnitID(ctx, client, in.OrgType, in.OrgName, in.ClusterID)
		if err != nil {
			return getOrgUnitMetricsOutput{}, err
		}
		orgID = id
	}

	window := in.Window
	if window == "" {
		window = defaultWindow
	}
	from, to := resolveTimeRange(window, in.Start, in.End)
	samples := in.NumberOfSamples
	if samples == 0 {
		samples = defaultSamples
	}

	metricNames := dedupe(in.MetricTypes)
	if len(metricNames) == 0 {
		metricNames = defaultMetricTypes
	}

	out := getOrgUnitMetricsOutput{
		OrgType:   in.OrgType,
		OrgID:     orgID,
		OrgName:   orgName,
		NodePool:  common.StringPtrIfNonEmpty(in.NodePoolName),
		TimeRange: timeRange{From: common.TimestampString(&from), To: common.TimestampString(&to), NumberOfSamples: samples},
	}

	// Token guard: refuse before the upstream call when the requested grid is
	// too large, rather than fetch then drop.
	if int(samples)*len(metricNames) > maxResultPoints {
		out.SeriesTruncated = true
		return out, nil
	}

	apiTypes := make([]orgunit.OrgUnitMetricType, 0, len(metricNames))
	for _, m := range metricNames {
		apiTypes = append(apiTypes, orgMetricCatalog[m].apiType)
	}

	resp, httpResp, err := fetchOrgUnitMetrics(ctx, client, in, orgID, apiTypes, from, to, samples)
	if err != nil {
		return getOrgUnitMetricsOutput{}, err
	}
	// 207 Multi-Status means some metrics failed; the generated client treats it
	// as success (resp + httpResp, no error), so we read StatusCode here. If a
	// future client regen maps 207 → error, this branch silently stops firing —
	// revisit partialData handling then.
	if httpResp != nil && httpResp.StatusCode == 207 {
		out.PartialData = true
	}

	out.Series = buildSeries(ctx, resp, metricNames)
	return out, nil
}

func fetchOrgUnitMetrics(
	ctx context.Context, client *runaiclient.Client, in getOrgUnitMetricsInput,
	orgID string, apiTypes []orgunit.OrgUnitMetricType, from, to time.Time, samples int32,
) (*orgunit.MetricsResponse, *http.Response, error) {
	if in.OrgType == common.OrgTypeProject {
		req := client.OrgUnit.ProjectsAPI.GetProjectMetrics(ctx, orgID).
			MetricType(apiTypes).Start(from).End(to).NumberOfSamples(samples)
		if in.NodePoolName != "" {
			req = req.NodepoolName(in.NodePoolName)
		}
		resp, httpResp, err := req.Execute()
		if err != nil {
			return nil, httpResp, fmt.Errorf("get project metrics %q: %w", orgID, runaimcp.MapHTTPError(httpResp, err))
		}
		return resp, httpResp, nil
	}
	req := client.OrgUnit.DepartmentsAPI.GetDepartmentMetrics(ctx, orgID).
		MetricType(apiTypes).Start(from).End(to).NumberOfSamples(samples)
	if in.NodePoolName != "" {
		req = req.NodepoolName(in.NodePoolName)
	}
	resp, httpResp, err := req.Execute()
	if err != nil {
		return nil, httpResp, fmt.Errorf("get department metrics %q: %w", orgID, runaimcp.MapHTTPError(httpResp, err))
	}
	return resp, httpResp, nil
}

func buildSeries(ctx context.Context, resp *orgunit.MetricsResponse, requested []string) *orgUnitSeries {
	if resp == nil {
		return &orgUnitSeries{}
	}
	byAPIType := map[string][]orgunit.MeasurementResponse{}
	for _, m := range resp.Measurements {
		byAPIType[m.Type] = append(byAPIType[m.Type], m)
	}

	s := &orgUnitSeries{}
	for _, name := range requested {
		kind := orgMetricCatalog[name]
		points := collapseMeasurements(ctx, byAPIType[string(kind.apiType)], kind)
		if len(points) == 0 {
			continue
		}
		assignSeries(s, name, points)
	}
	return s
}

func collapseMeasurements(ctx context.Context, measurements []orgunit.MeasurementResponse, kind seriesKind) []point {
	type agg struct {
		sum   float64
		count int
	}
	byTS := map[string]*agg{}
	dropped := 0
	for _, m := range measurements {
		for _, v := range m.Values {
			ts := v.Timestamp.Get()
			if ts == nil {
				dropped++
				continue
			}
			f, err := strconv.ParseFloat(v.Value, 64)
			if err != nil || !common.IsFinite(f) {
				dropped++
				continue
			}
			key := ts.UTC().Format(time.RFC3339)
			a := byTS[key]
			if a == nil {
				a = &agg{}
				byTS[key] = a
			}
			a.sum += f
			a.count++
		}
	}
	// Malformed points (nil timestamp / unparseable value) are skipped silently
	// in the result; surface a count at debug level so a fully-empty series isn't
	// invisible.
	if dropped > 0 {
		zerolog.Ctx(ctx).Debug().Int("dropped", dropped).Str("metric", string(kind.apiType)).
			Msg("get_org_unit_metrics: skipped unparseable measurement points")
	}
	points := make([]point, 0, len(byTS))
	for ts, a := range byTS {
		val := a.sum
		if !kind.sum && a.count > 0 { // utilization → average across label variants
			val = a.sum / float64(a.count)
		}
		if kind.divisor != 1 {
			val /= kind.divisor
		}
		points = append(points, point{T: ts, V: common.RoundTo(val, 3)})
	}
	sort.Slice(points, func(i, j int) bool { return points[i].T < points[j].T })
	return points
}

func assignSeries(s *orgUnitSeries, name string, points []point) {
	switch name {
	case metricGPUAllocation:
		s.GPUAllocation = points
	case metricCPUAllocation:
		s.CPUAllocation = points
	case metricCPUMemoryAllocation:
		s.CPUMemoryAllocation = points
	case metricGPUUtilization:
		s.GPUUtilization = points
	case metricGPUMemoryUtilization:
		s.GPUMemoryUtilization = points
	case metricCPUUtilization:
		s.CPUUtilization = points
	case metricCPUMemoryUtilization:
		s.CPUMemoryUtilization = points
	}
}
