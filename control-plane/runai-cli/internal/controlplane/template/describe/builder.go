package describe

import (
	"time"
)

// TemplateDescriptionBuilder builds Template (display model)
// Used to construct Template objects for CLI output and display
type TemplateDescriptionBuilder struct {
	template Template
}

func NewTemplateDescriptionBuilder() *TemplateDescriptionBuilder {
	return &TemplateDescriptionBuilder{
		template: Template{
			Meta: TemplateMeta{},
			Spec: &TemplateSpec{},
		},
	}
}

// Meta Methods
func (b *TemplateDescriptionBuilder) SetID(id string) *TemplateDescriptionBuilder {
	b.template.Meta.ID = id
	return b
}

func (b *TemplateDescriptionBuilder) SetName(name string) *TemplateDescriptionBuilder {
	b.template.Meta.Name = name
	return b
}

func (b *TemplateDescriptionBuilder) SetDescription(description *string) *TemplateDescriptionBuilder {
	b.template.Meta.Description = description
	return b
}

func (b *TemplateDescriptionBuilder) SetType(templateType string) *TemplateDescriptionBuilder {
	b.template.Meta.Type = templateType
	return b
}

func (b *TemplateDescriptionBuilder) SetScope(scope string) *TemplateDescriptionBuilder {
	b.template.Meta.Scope = scope
	return b
}

func (b *TemplateDescriptionBuilder) SetScopeID(scopeId string) *TemplateDescriptionBuilder {
	b.template.Meta.ScopeID = scopeId
	return b
}

func (b *TemplateDescriptionBuilder) SetClusterID(clusterId *string) *TemplateDescriptionBuilder {
	b.template.Meta.ClusterID = clusterId
	return b
}

func (b *TemplateDescriptionBuilder) SetProjectID(projectId *int32) *TemplateDescriptionBuilder {
	b.template.Meta.ProjectID = projectId
	return b
}

func (b *TemplateDescriptionBuilder) SetDepartmentID(departmentId *string) *TemplateDescriptionBuilder {
	b.template.Meta.DepartmentID = departmentId
	return b
}

func (b *TemplateDescriptionBuilder) SetCreatedAt(createdAt time.Time) *TemplateDescriptionBuilder {
	b.template.Meta.CreatedAt = createdAt
	return b
}

func (b *TemplateDescriptionBuilder) SetCreatedBy(createdBy string) *TemplateDescriptionBuilder {
	b.template.Meta.CreatedBy = createdBy
	return b
}

func (b *TemplateDescriptionBuilder) SetUpdatedAt(updatedAt *time.Time) *TemplateDescriptionBuilder {
	b.template.Meta.UpdatedAt = updatedAt
	return b
}

func (b *TemplateDescriptionBuilder) SetUpdatedBy(updatedBy *string) *TemplateDescriptionBuilder {
	b.template.Meta.UpdatedBy = updatedBy
	return b
}

// Spec Methods
func (b *TemplateDescriptionBuilder) SetImage(image *string) *TemplateDescriptionBuilder {
	b.template.Spec.Image = image
	return b
}

func (b *TemplateDescriptionBuilder) SetCommand(command *string) *TemplateDescriptionBuilder {
	b.template.Spec.Command = command
	return b
}

func (b *TemplateDescriptionBuilder) SetArgs(args *string) *TemplateDescriptionBuilder {
	b.template.Spec.Args = args
	return b
}

func (b *TemplateDescriptionBuilder) SetWorkingDir(workingDir *string) *TemplateDescriptionBuilder {
	b.template.Spec.WorkingDir = workingDir
	return b
}

func (b *TemplateDescriptionBuilder) SetPreemptible(preemptible *bool) *TemplateDescriptionBuilder {
	b.template.Spec.Preemptible = preemptible
	return b
}

func (b *TemplateDescriptionBuilder) SetPriorityClass(priorityClass *string) *TemplateDescriptionBuilder {
	b.template.Spec.PriorityClass = priorityClass
	return b
}

func (b *TemplateDescriptionBuilder) SetNodePools(nodePools []string) *TemplateDescriptionBuilder {
	b.template.Spec.NodePools = nodePools
	return b
}

func (b *TemplateDescriptionBuilder) SetEnvironmentVariables(envVars map[string]string) *TemplateDescriptionBuilder {
	b.template.Spec.EnvironmentVariables = envVars
	return b
}

func (b *TemplateDescriptionBuilder) SetCompute(compute *Compute) *TemplateDescriptionBuilder {
	b.template.Spec.Compute = compute
	return b
}

func (b *TemplateDescriptionBuilder) AddPort(port Port) *TemplateDescriptionBuilder {
	b.template.Spec.Ports = append(b.template.Spec.Ports, port)
	return b
}

func (b *TemplateDescriptionBuilder) AddExposedUrl(url ExposedUrl) *TemplateDescriptionBuilder {
	b.template.Spec.ExposedUrls = append(b.template.Spec.ExposedUrls, url)
	return b
}

func (b *TemplateDescriptionBuilder) SetMasterSpec(masterSpec *MasterSpec) *TemplateDescriptionBuilder {
	b.template.MasterSpec = masterSpec
	return b
}

func (b *TemplateDescriptionBuilder) Build() *Template {
	return &b.template
}
