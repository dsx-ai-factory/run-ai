package describe

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	instancecfg "github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/ui/field"
	uitable "github.com/run-ai/runai/control-plane/runai-cli/pkg/ui/table"
	workloads_openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"k8s.io/utils/ptr"
)

var (
	sectionsSeparator = "\n\n"
	computeFields     = []string{fieldGPUCompute, fieldCPUCores, fieldGPUMemory, fieldCPUMemory}
)

const (
	// Field names
	fieldName          = "Name"
	fieldID            = "ID"
	fieldType          = "Type"
	fieldDescription   = "Description"
	fieldScope         = "Scope"
	fieldProject       = "Project"
	fieldDepartment    = "Department"
	fieldCreatedAt     = "Created At"
	fieldCreatedBy     = "Created By"
	fieldUpdatedAt     = "Updated At"
	fieldUpdatedBy     = "Updated By"
	fieldPreemptible   = "Preemptible"
	fieldPriorityClass = "Priority Class"
	fieldImage         = "Image"
	fieldNodePool      = "Requested Node Pool"
	fieldEnvVars       = "Environment Variables"
	fieldCommand       = "Command"
	fieldArgs          = "Args"
	fieldWorkingDir    = "Working Dir"

	// Section titles
	titleCompute        = "Compute Resources"
	titleNetworks       = "Networks"
	titleMasterSpec     = "Master Configuration"
	titleMasterCompute  = "Master Compute"
	titleMasterNetworks = "Master Networks"

	// Compute fields
	fieldGPUCompute = "GPU Compute"
	fieldCPUCores   = "CPU Cores"
	fieldGPUMemory  = "GPU Memory"
	fieldCPUMemory  = "CPU Memory"

	// Network fields
	fieldPorts        = "Ports"
	fieldExternalURLs = "External URLs"

	// Date format
	dateFormat = time.RFC3339

	// Column padding
	generalAndRuntimeColumnPadding = 35

	// Compute column widths
	ResourceColumnWidth  = 20
	RequestedColumnWidth = 15
	LimitColumnWidth     = 15
)

type DisplayField struct {
	Name  string
	Value interface{}
}

// DisplaySection wraps a table displayer
type DisplaySection struct {
	Displayer uitable.Constructor
}

// Displayer interface for rendering template descriptions
type Displayer interface {
	AsTable() error
	AsJSON() error
	AsYAML() error
}

// TemplateDisplayer implements Displayer for template descriptions
type TemplateDisplayer struct {
	template *Template
}

// NewTemplateDisplayer creates a new TemplateDisplayer
func NewTemplateDisplayer(template *Template) *TemplateDisplayer {
	return &TemplateDisplayer{
		template: template,
	}
}

// AsTable renders template in table format
func (td *TemplateDisplayer) AsTable() error {
	return td.renderSingleTable(td.template)
}

// AsJSON renders template in JSON format (uses clean structure with empty fields omitted)
func (td *TemplateDisplayer) AsJSON() error {
	return td.renderJSON(omitEmpty(td.buildCleanOutput()))
}

// AsYAML renders template in YAML format (uses clean structure with empty fields omitted)
func (td *TemplateDisplayer) AsYAML() error {
	return td.renderYAML(omitEmpty(td.buildCleanOutput()))
}

// buildCleanOutput creates a clean output structure from the raw API response
// Only includes: meta, spec (from combinedSpec), masterSpec (from combinedMasterSpec if present)
func (td *TemplateDisplayer) buildCleanOutput() interface{} {
	raw, _ := td.template.Raw.(*workloads_openapi.SupersetTemplate)

	output := map[string]interface{}{"meta": raw.Meta}

	if raw.HasCombinedSpec() {
		output["spec"] = raw.GetCombinedSpec()
	}

	if raw.HasCombinedMasterSpec() {
		output["masterSpec"] = raw.GetCombinedMasterSpec()
	}

	return output
}

func (td *TemplateDisplayer) renderJSON(data interface{}) error {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling to JSON: %w", err)
	}
	fmt.Println(string(jsonBytes)) //nolint:forbidigo // CLI output to user
	return nil
}

func (td *TemplateDisplayer) renderYAML(data interface{}) error {
	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshalling to YAML: %w", err)
	}
	fmt.Println(string(yamlBytes)) //nolint:forbidigo // CLI output to user
	return nil
}

func (td *TemplateDisplayer) renderSingleTable(template *Template) error {
	td.DisplayGeneralAndSpec(template)

	// Display sections
	if template.Spec != nil {
		if template.Spec.Compute != nil {
			if err := td.DisplaySection(NewComputeInfoDisplayer(template.Spec.Compute).Displayer); err != nil {
				return err
			}
		}
		if len(template.Spec.Ports) > 0 || len(template.Spec.ExposedUrls) > 0 {
			td.displayNetworkInfo(template.Spec)
		}
	}

	// Display master spec for distributed workloads (when master != worker)
	if template.MasterSpec != nil {
		td.displayMasterSpec(template.MasterSpec)
	}

	return nil
}

func (td *TemplateDisplayer) DisplaySection(displayer uitable.Constructor) error {
	if len(displayer.Data()) == 0 {
		return nil
	}

	fmt.Printf("%s%s%s", sectionsSeparator, displayer.Subject(), sectionsSeparator) //nolint:forbidigo // CLI output to user
	tableDisplay := uitable.NewTableDisplay(displayer)
	output, err := tableDisplay.Generate()
	if err != nil {
		return err
	}
	fmt.Println(output) //nolint:forbidigo // CLI output to user

	return nil
}

// DisplayGeneralAndSpec displays general metadata and spec fields
func (td *TemplateDisplayer) DisplayGeneralAndSpec(template *Template) {
	meta := &template.Meta
	spec := template.Spec

	fields := []DisplayField{
		{fieldName, meta.Name},
		{fieldID, meta.ID},
		{fieldType, meta.Type},
		{fieldDescription, field.StringOrEmpty(meta.Description)},
		{fieldScope, meta.Scope},
		{fieldProject, field.Int32OrEmpty(meta.ProjectID)},
		{fieldDepartment, field.StringOrEmpty(meta.DepartmentID)},
		{fieldCreatedAt, meta.CreatedAt.Format(dateFormat)},
		{fieldCreatedBy, meta.CreatedBy},
		{fieldUpdatedAt, field.TimeOrEmpty(meta.UpdatedAt)},
		{fieldUpdatedBy, field.StringOrEmpty(meta.UpdatedBy)},
	}

	// Add spec fields if present
	if spec != nil {
		fields = append(fields,
			DisplayField{fieldPreemptible, field.BoolOrEmpty(spec.Preemptible)},
			DisplayField{fieldPriorityClass, field.StringOrEmpty(spec.PriorityClass)},
			DisplayField{fieldImage, field.StringOrEmpty(spec.Image)},
		)

		if len(spec.NodePools) > 0 {
			fields = append(fields, DisplayField{fieldNodePool, strings.Join(spec.NodePools, ", ")})
		}

		if ptr.Deref(spec.Command, "") != "" {
			fields = append(fields, DisplayField{fieldCommand, *spec.Command})
		}

		if ptr.Deref(spec.Args, "") != "" {
			fields = append(fields, DisplayField{fieldArgs, *spec.Args})
		}

		if ptr.Deref(spec.WorkingDir, "") != "" {
			fields = append(fields, DisplayField{fieldWorkingDir, *spec.WorkingDir})
		}

		// Add environment variables field
		if len(spec.EnvironmentVariables) > 0 {
			fields = append(fields, DisplayField{fieldEnvVars, getEnvironmentVariablesDisplay(&spec.EnvironmentVariables, generalAndRuntimeColumnPadding)})
		}
	}

	displayFields(fields)
}

// displayFields displays a list of fields (package-level function like workloads)
func displayFields(fields []DisplayField) {
	for _, field := range fields {
		if value := fmt.Sprintf("%v", field.Value); value != "" && value != "<nil>" {
			fmt.Printf("%-*s %s\n", generalAndRuntimeColumnPadding, field.Name+":", value) //nolint:forbidigo // CLI output to user
		}
	}
}

// ComputeInfoDisplayer displays compute resources in table format
type ComputeInfoDisplayer struct {
	compute *Compute
	fields  []string
}

func NewComputeInfoDisplayer(compute *Compute) *DisplaySection {
	return &DisplaySection{
		Displayer: &ComputeInfoDisplayer{
			compute: compute,
			fields:  computeFields,
		},
	}
}

func (c *ComputeInfoDisplayer) Subject() string {
	return titleCompute
}

func (c *ComputeInfoDisplayer) Columns() []table.Column {
	return []table.Column{
		{Title: "Resource", Width: ResourceColumnWidth},
		{Title: "Requested", Width: RequestedColumnWidth},
		{Title: "Limit", Width: LimitColumnWidth},
	}
}

func (c *ComputeInfoDisplayer) Data() [][]string {
	var data [][]string
	for _, fieldType := range c.fields {
		var row []string
		switch fieldType {
		case fieldGPUCompute:
			row = []string{
				fieldGPUCompute,
				field.Int32OrEmpty(c.compute.GpuDevicesRequest),
				"",
			}
		case fieldCPUCores:
			row = []string{
				fieldCPUCores,
				field.Float64OrEmpty(c.compute.CpuCoreRequest),
				field.Float64OrEmpty(c.compute.CpuCoreLimit),
			}
		case fieldGPUMemory:
			row = []string{
				fieldGPUMemory,
				field.StringOrEmpty(c.compute.GpuMemoryRequest),
				field.StringOrEmpty(c.compute.GpuMemoryLimit),
			}
		case fieldCPUMemory:
			row = []string{
				fieldCPUMemory,
				field.StringOrEmpty(c.compute.CpuMemoryRequest),
				field.StringOrEmpty(c.compute.CpuMemoryLimit),
			}
		}
		data = append(data, row)
	}
	return data
}

// MasterComputeInfoDisplayer displays master compute resources in table format
type MasterComputeInfoDisplayer struct {
	compute *Compute
	fields  []string
}

func NewMasterComputeInfoDisplayer(compute *Compute) *DisplaySection {
	return &DisplaySection{
		Displayer: &MasterComputeInfoDisplayer{
			compute: compute,
			fields:  computeFields,
		},
	}
}

func (c *MasterComputeInfoDisplayer) Subject() string {
	return titleMasterCompute
}

func (c *MasterComputeInfoDisplayer) Columns() []table.Column {
	return []table.Column{
		{Title: "Resource", Width: ResourceColumnWidth},
		{Title: "Requested", Width: RequestedColumnWidth},
		{Title: "Limit", Width: LimitColumnWidth},
	}
}

func (c *MasterComputeInfoDisplayer) Data() [][]string {
	var data [][]string
	for _, fieldType := range c.fields {
		var row []string
		switch fieldType {
		case fieldGPUCompute:
			row = []string{
				fieldGPUCompute,
				field.Int32OrEmpty(c.compute.GpuDevicesRequest),
				"",
			}
		case fieldCPUCores:
			row = []string{
				fieldCPUCores,
				field.Float64OrEmpty(c.compute.CpuCoreRequest),
				field.Float64OrEmpty(c.compute.CpuCoreLimit),
			}
		case fieldGPUMemory:
			row = []string{
				fieldGPUMemory,
				field.StringOrEmpty(c.compute.GpuMemoryRequest),
				field.StringOrEmpty(c.compute.GpuMemoryLimit),
			}
		case fieldCPUMemory:
			row = []string{
				fieldCPUMemory,
				field.StringOrEmpty(c.compute.CpuMemoryRequest),
				field.StringOrEmpty(c.compute.CpuMemoryLimit),
			}
		}
		data = append(data, row)
	}
	return data
}

func (td *TemplateDisplayer) displayNetworkInfo(spec *TemplateSpec) {
	fmt.Printf("%s%s%s", sectionsSeparator, titleNetworks, sectionsSeparator) //nolint:forbidigo // CLI output to user

	var fields []DisplayField

	if len(spec.Ports) > 0 {
		fields = append(fields, DisplayField{fieldPorts, formatPorts(spec.Ports)})
	}

	if len(spec.ExposedUrls) > 0 {
		fields = append(fields, DisplayField{fieldExternalURLs, formatExposedUrls(spec.ExposedUrls)})
	}

	displayFields(fields)
}

// displayMasterSpec displays master spec information for distributed workloads
func (td *TemplateDisplayer) displayMasterSpec(masterSpec *MasterSpec) {
	fmt.Printf("%s%s%s", sectionsSeparator, titleMasterSpec, sectionsSeparator) //nolint:forbidigo // CLI output to user

	var fields []DisplayField

	if ptr.Deref(masterSpec.Image, "") != "" {
		fields = append(fields, DisplayField{fieldImage, *masterSpec.Image})
	}

	if len(masterSpec.NodePools) > 0 {
		fields = append(fields, DisplayField{fieldNodePool, strings.Join(masterSpec.NodePools, ", ")})
	}

	if ptr.Deref(masterSpec.Command, "") != "" {
		fields = append(fields, DisplayField{fieldCommand, *masterSpec.Command})
	}

	if ptr.Deref(masterSpec.Args, "") != "" {
		fields = append(fields, DisplayField{fieldArgs, *masterSpec.Args})
	}

	if ptr.Deref(masterSpec.WorkingDir, "") != "" {
		fields = append(fields, DisplayField{fieldWorkingDir, *masterSpec.WorkingDir})
	}

	if len(masterSpec.EnvironmentVariables) > 0 {
		fields = append(fields, DisplayField{fieldEnvVars, getEnvironmentVariablesDisplay(&masterSpec.EnvironmentVariables, generalAndRuntimeColumnPadding)})
	}

	displayFields(fields)

	// Display master compute resources
	if masterSpec.Compute != nil {
		if err := td.DisplaySection(NewMasterComputeInfoDisplayer(masterSpec.Compute).Displayer); err != nil {
			logger := instancecfg.GetInstance().GetLogger()
			logger.Debug("failed to display master compute info", zap.Error(err))
		}
	}

	// Display master network info
	if len(masterSpec.Ports) > 0 || len(masterSpec.ExposedUrls) > 0 {
		td.displayMasterNetworkInfo(masterSpec)
	}
}

// displayMasterNetworkInfo displays network info for master spec
func (td *TemplateDisplayer) displayMasterNetworkInfo(masterSpec *MasterSpec) {
	fmt.Printf("%s%s%s", sectionsSeparator, titleMasterNetworks, sectionsSeparator) //nolint:forbidigo // CLI output to user

	var fields []DisplayField

	if len(masterSpec.Ports) > 0 {
		fields = append(fields, DisplayField{fieldPorts, formatPorts(masterSpec.Ports)})
	}

	if len(masterSpec.ExposedUrls) > 0 {
		fields = append(fields, DisplayField{fieldExternalURLs, formatExposedUrls(masterSpec.ExposedUrls)})
	}

	displayFields(fields)
}

func formatPorts(ports []Port) string {
	portStrs := make([]string, 0, len(ports))
	for _, port := range ports {
		if port.External != nil {
			portStrs = append(portStrs, fmt.Sprintf("%d (External: %d)", port.Container, *port.External))
		} else {
			portStrs = append(portStrs, fmt.Sprintf("%d", port.Container))
		}
	}
	return fmt.Sprintf("[%s]", strings.Join(portStrs, ", "))
}

func formatExposedUrls(urls []ExposedUrl) string {
	urlStrs := make([]string, 0, len(urls))
	for _, exposedUrl := range urls {
		if exposedUrl.Name != "" {
			urlStrs = append(urlStrs, fmt.Sprintf("%s (%s)", exposedUrl.Url, exposedUrl.Name))
		} else {
			urlStrs = append(urlStrs, exposedUrl.Url)
		}
	}
	indent := strings.Repeat(" ", generalAndRuntimeColumnPadding+1)
	return strings.Join(urlStrs, "\n"+indent)
}

// TODO(Roee G.): This function is duplicated in workload/describe/displayer/environment.go
// Remove this duplicate and use the shared implementation from displayer package after refactoring
func getEnvironmentVariablesDisplay(environmentVariables *map[string]string, columnPadding int) string {
	if environmentVariables == nil || len(*environmentVariables) == 0 {
		return "[]"
	}

	var envVarsBuilder strings.Builder
	envVarsBuilder.WriteString("[")

	// Compute max key length for alignment
	maxKeyLength := 0
	for key := range *environmentVariables {
		if len(key) > maxKeyLength {
			maxKeyLength = len(key)
		}
	}

	keys := lo.Keys(*environmentVariables)
	slices.Sort(keys)

	valueColumnIndent := strings.Repeat(" ", columnPadding+1)

	// Iterate through environment variables
	for i, key := range keys {
		keyValuePart := fmt.Sprintf("%-*s: %s", maxKeyLength, key, (*environmentVariables)[key])
		if i == 0 {
			envVarsBuilder.WriteString(keyValuePart)
		} else {
			envVarsBuilder.WriteString(",\n")
			envVarsBuilder.WriteString(valueColumnIndent)
			envVarsBuilder.WriteString(keyValuePart)
		}
	}

	envVarsBuilder.WriteString("]")

	return envVarsBuilder.String()
}
