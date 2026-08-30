package services_crud

import (
	"context"
	"fmt"

	authorization_api "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/run-ai/runai/control-plane/authorization/pkg/client/helpers"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/dao_crud"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models_crud"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

const (
	createdByRunai          = "Run:ai"
	systemTenantId          = "0"
	systemSubjectID         = "system"
	mediumHighPriorityClass = "medium-high"
	PyTorchImage            = "nvcr.io/nvidia/pytorch"
	PyTorchTag              = "25.06-py3"
	BioNemoImage            = "nvcr.io/nvidia/clara/bionemo-framework"
	BioNemoTag              = "2.6.3"
	NemoImage               = "nvcr.io/nvidia/nemo"
	NemoTag                 = "25.07"
)

var (
	systemGroups          = []string{"System"}
	systemPermittedScopes = &authorization_api.PermittedScopes{
		System:  convutils.Ref(true),
		Tenants: []string{"0"},
	}
)

type presetEntry struct {
	template      api.SupersetTemplate
	routingMethod string
}

// Path-base presets

var nemoFourGpus = presetEntry{
	routingMethod: models_crud.TemplatesColumnRoutingMethodPathBase,
	template: api.SupersetTemplate{
		Meta: api.TemplateMetaFields{
			Name:         "nemo-four-gpus",
			Description:  convutils.Ref("Nemo / 4 GPUs (v4)"),
			WorkloadType: api.WORKLOADTYPE_TRAINING,
			ScopeType:    api.SCOPETYPE_SYSTEM,
			CreatedBy:    createdByRunai,
			UpdatedBy:    convutils.Ref(createdByRunai),
		},
		Spec: api.SupersetSpec{
			Image:           convutils.Ref(fmt.Sprintf("%s:%s", NemoImage, NemoTag)),
			ImagePullPolicy: convutils.Ref(api.IMAGEPULLPOLICY_IF_NOT_PRESENT),
			Command:         convutils.Ref("jupyter-lab"),
			Args:            convutils.Ref("--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token='' --ServerApp.allow_remote_access=true --allow-root --port=8888 --no-browser"),
			WorkingDir:      convutils.Ref("/opt/NeMo"),
			ExposedUrls: &[]api.ExposedUrl{
				{
					Container: convutils.Ref(int32(8888)),
					ToolName:  convutils.Ref("Jupyter"),
					ToolType:  convutils.Ref("jupyter-notebook"),
				},
			},
			Compute: &api.WorkspaceSpecSpecAllOfCompute{
				GpuDevicesRequest: convutils.Ref(int32(4)),
			},
			Security: &api.WorkspaceSpecSpecAllOfSecurity{
				UidGidSource: convutils.Ref(api.UIDGIDSOURCE_FROM_THE_IMAGE),
			},
		},
	},
}

var bioNemoFourGpus = presetEntry{
	routingMethod: models_crud.TemplatesColumnRoutingMethodPathBase,
	template: api.SupersetTemplate{
		Meta: api.TemplateMetaFields{
			Name:         "bio-nemo-four-gpus",
			Description:  convutils.Ref("Bio Nemo / 4 GPUs (v4)"),
			WorkloadType: api.WORKLOADTYPE_TRAINING,
			ScopeType:    api.SCOPETYPE_SYSTEM,
			CreatedBy:    createdByRunai,
			UpdatedBy:    convutils.Ref(createdByRunai),
		},
		Spec: api.SupersetSpec{
			Image:           convutils.Ref(fmt.Sprintf("%s:%s", BioNemoImage, BioNemoTag)),
			ImagePullPolicy: convutils.Ref(api.IMAGEPULLPOLICY_IF_NOT_PRESENT),
			Command:         convutils.Ref("jupyter-lab"),
			Args:            convutils.Ref("--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token='' --ServerApp.allow_remote_access=true --allow-root --port=8888 --no-browser"),
			WorkingDir:      convutils.Ref("/opt/NeMo"),
			ExposedUrls: &[]api.ExposedUrl{
				{
					Container: convutils.Ref(int32(8888)),
					ToolName:  convutils.Ref("Jupyter"),
					ToolType:  convutils.Ref("jupyter-notebook"),
				},
			},
			Compute: &api.WorkspaceSpecSpecAllOfCompute{
				GpuDevicesRequest: convutils.Ref(int32(4)),
			},
			Security: &api.WorkspaceSpecSpecAllOfSecurity{
				UidGidSource: convutils.Ref(api.UIDGIDSOURCE_FROM_THE_IMAGE),
			},
		},
	},
}

var jupyterLabOneGpu = presetEntry{
	routingMethod: models_crud.TemplatesColumnRoutingMethodPathBase,
	template: api.SupersetTemplate{
		Meta: api.TemplateMetaFields{
			Name:         "jupyter-lab-one-gpu",
			Description:  convutils.Ref("Jupyter Lab / 1 GPU (v4)"),
			WorkloadType: api.WORKLOADTYPE_WORKSPACE,
			ScopeType:    api.SCOPETYPE_SYSTEM,
			CreatedBy:    createdByRunai,
			UpdatedBy:    convutils.Ref(createdByRunai),
		},
		Spec: api.SupersetSpec{
			Image:           convutils.Ref("jupyter/scipy-notebook"),
			ImagePullPolicy: convutils.Ref(api.IMAGEPULLPOLICY_IF_NOT_PRESENT),
			Command:         convutils.Ref("start-notebook.sh"),
			Args:            convutils.Ref("--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=''"),
			ExposedUrls: &[]api.ExposedUrl{
				{
					Container: convutils.Ref(int32(8888)),
					ToolName:  convutils.Ref("Jupyter"),
					ToolType:  convutils.Ref("jupyter-notebook"),
				},
			},
			Compute: &api.WorkspaceSpecSpecAllOfCompute{
				GpuDevicesRequest: convutils.Ref(int32(1)),
			},
			Security: &api.WorkspaceSpecSpecAllOfSecurity{
				UidGidSource: convutils.Ref(api.UIDGIDSOURCE_FROM_THE_IMAGE),
			},
			PriorityClass: convutils.Ref(mediumHighPriorityClass),
		},
	},
}

var pytorchOneGpu = presetEntry{
	routingMethod: models_crud.TemplatesColumnRoutingMethodPathBase,
	template: api.SupersetTemplate{
		Meta: api.TemplateMetaFields{
			Name:         "pytorch-one-gpu",
			Description:  convutils.Ref("Pytorch / 1 GPU (v4)"),
			WorkloadType: api.WORKLOADTYPE_TRAINING,
			ScopeType:    api.SCOPETYPE_SYSTEM,
			CreatedBy:    createdByRunai,
			UpdatedBy:    convutils.Ref(createdByRunai),
		},
		Spec: api.SupersetSpec{
			Image:           convutils.Ref(fmt.Sprintf("%s:%s", PyTorchImage, PyTorchTag)),
			ImagePullPolicy: convutils.Ref(api.IMAGEPULLPOLICY_IF_NOT_PRESENT),
			Command:         convutils.Ref("jupyter-lab"),
			Args:            convutils.Ref("--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token='' --ServerApp.allow_remote_access=true --allow-root --port=8888 --no-browser"),
			WorkingDir:      convutils.Ref("/opt/NeMo"),
			ExposedUrls: &[]api.ExposedUrl{
				{
					Container: convutils.Ref(int32(8888)),
					ToolName:  convutils.Ref("Jupyter"),
					ToolType:  convutils.Ref("jupyter-notebook"),
				},
			},
			Compute: &api.WorkspaceSpecSpecAllOfCompute{
				GpuDevicesRequest: convutils.Ref(int32(1)),
			},
			Security: &api.WorkspaceSpecSpecAllOfSecurity{
				UidGidSource: convutils.Ref(api.UIDGIDSOURCE_FROM_THE_IMAGE),
			},
		},
	},
}

// Host-base presets

var nemoFourGpusHost = presetEntry{
	routingMethod: models_crud.TemplatesColumnRoutingMethodHostBase,
	template: api.SupersetTemplate{
		Meta: api.TemplateMetaFields{
			Name:         "nemo-4-gpus",
			Description:  convutils.Ref("Nemo / 4 GPUs (v1)"),
			WorkloadType: api.WORKLOADTYPE_TRAINING,
			ScopeType:    api.SCOPETYPE_SYSTEM,
			CreatedBy:    createdByRunai,
			UpdatedBy:    convutils.Ref(createdByRunai),
		},
		Spec: api.SupersetSpec{
			Image:           convutils.Ref(fmt.Sprintf("%s:%s", NemoImage, NemoTag)),
			ImagePullPolicy: convutils.Ref(api.IMAGEPULLPOLICY_IF_NOT_PRESENT),
			Command:         convutils.Ref("jupyter-lab"),
			Args:            convutils.Ref("--NotebookApp.token='' --ServerApp.allow_remote_access=true --allow-root --port=8888 --no-browser"),
			WorkingDir:      convutils.Ref("/opt/NeMo"),
			ExposedUrls: &[]api.ExposedUrl{
				{
					Container: convutils.Ref(int32(8888)),
					ToolName:  convutils.Ref("Jupyter"),
					ToolType:  convutils.Ref("jupyter-notebook"),
				},
			},
			Compute: &api.WorkspaceSpecSpecAllOfCompute{
				GpuDevicesRequest: convutils.Ref(int32(4)),
			},
			Security: &api.WorkspaceSpecSpecAllOfSecurity{
				UidGidSource: convutils.Ref(api.UIDGIDSOURCE_FROM_THE_IMAGE),
			},
		},
	},
}

var bioNemoFourGpusHost = presetEntry{
	routingMethod: models_crud.TemplatesColumnRoutingMethodHostBase,
	template: api.SupersetTemplate{
		Meta: api.TemplateMetaFields{
			Name:         "bio-nemo-4-gpus",
			Description:  convutils.Ref("Bio Nemo / 4 GPUs (v1)"),
			WorkloadType: api.WORKLOADTYPE_TRAINING,
			ScopeType:    api.SCOPETYPE_SYSTEM,
			CreatedBy:    createdByRunai,
			UpdatedBy:    convutils.Ref(createdByRunai),
		},
		Spec: api.SupersetSpec{
			Image:           convutils.Ref(fmt.Sprintf("%s:%s", BioNemoImage, BioNemoTag)),
			ImagePullPolicy: convutils.Ref(api.IMAGEPULLPOLICY_IF_NOT_PRESENT),
			Command:         convutils.Ref("jupyter-lab"),
			Args:            convutils.Ref("--NotebookApp.token='' --ServerApp.allow_remote_access=true --allow-root --port=8888 --no-browser"),
			WorkingDir:      convutils.Ref("/opt/NeMo"),
			ExposedUrls: &[]api.ExposedUrl{
				{
					Container: convutils.Ref(int32(8888)),
					ToolName:  convutils.Ref("Jupyter"),
					ToolType:  convutils.Ref("jupyter-notebook"),
				},
			},
			Compute: &api.WorkspaceSpecSpecAllOfCompute{
				GpuDevicesRequest: convutils.Ref(int32(4)),
			},
			Security: &api.WorkspaceSpecSpecAllOfSecurity{
				UidGidSource: convutils.Ref(api.UIDGIDSOURCE_FROM_THE_IMAGE),
			},
		},
	},
}

var jupyterLabOneGpuHost = presetEntry{
	routingMethod: models_crud.TemplatesColumnRoutingMethodHostBase,
	template: api.SupersetTemplate{
		Meta: api.TemplateMetaFields{
			Name:         "jupyter-lab-1-gpu",
			Description:  convutils.Ref("Jupyter Lab / 1 GPU (v2)"),
			WorkloadType: api.WORKLOADTYPE_WORKSPACE,
			ScopeType:    api.SCOPETYPE_SYSTEM,
			CreatedBy:    createdByRunai,
			UpdatedBy:    convutils.Ref(createdByRunai),
		},
		Spec: api.SupersetSpec{
			Image:           convutils.Ref("jupyter/scipy-notebook"),
			ImagePullPolicy: convutils.Ref(api.IMAGEPULLPOLICY_IF_NOT_PRESENT),
			Command:         convutils.Ref("start-notebook.sh"),
			Args:            convutils.Ref("--NotebookApp.token=''"),
			ExposedUrls: &[]api.ExposedUrl{
				{
					Container: convutils.Ref(int32(8888)),
					ToolName:  convutils.Ref("Jupyter"),
					ToolType:  convutils.Ref("jupyter-notebook"),
				},
			},
			Compute: &api.WorkspaceSpecSpecAllOfCompute{
				GpuDevicesRequest: convutils.Ref(int32(1)),
			},
			Security: &api.WorkspaceSpecSpecAllOfSecurity{
				UidGidSource: convutils.Ref(api.UIDGIDSOURCE_FROM_THE_IMAGE),
			},
			PriorityClass: convutils.Ref(mediumHighPriorityClass),
		},
	},
}

var pytorchOneGpuHost = presetEntry{
	routingMethod: models_crud.TemplatesColumnRoutingMethodHostBase,
	template: api.SupersetTemplate{
		Meta: api.TemplateMetaFields{
			Name:         "pytorch-1-gpu",
			Description:  convutils.Ref("Pytorch / 1 GPU (v1)"),
			WorkloadType: api.WORKLOADTYPE_TRAINING,
			ScopeType:    api.SCOPETYPE_SYSTEM,
			CreatedBy:    createdByRunai,
			UpdatedBy:    convutils.Ref(createdByRunai),
		},
		Spec: api.SupersetSpec{
			Image:           convutils.Ref(fmt.Sprintf("%s:%s", PyTorchImage, PyTorchTag)),
			ImagePullPolicy: convutils.Ref(api.IMAGEPULLPOLICY_IF_NOT_PRESENT),
			Command:         convutils.Ref("jupyter-lab"),
			Args:            convutils.Ref("--NotebookApp.token='' --ServerApp.allow_remote_access=true --allow-root --port=8888 --no-browser"),
			WorkingDir:      convutils.Ref("/opt/NeMo"),
			ExposedUrls: &[]api.ExposedUrl{
				{
					Container: convutils.Ref(int32(8888)),
					ToolName:  convutils.Ref("Jupyter"),
					ToolType:  convutils.Ref("jupyter-notebook"),
				},
			},
			Compute: &api.WorkspaceSpecSpecAllOfCompute{
				GpuDevicesRequest: convutils.Ref(int32(1)),
			},
			Security: &api.WorkspaceSpecSpecAllOfSecurity{
				UidGidSource: convutils.Ref(api.UIDGIDSOURCE_FROM_THE_IMAGE),
			},
		},
	},
}

func nameAndTypeFromPreset(preset presetEntry) dao_crud.TemplateNameAndType {
	return dao_crud.TemplateNameAndType{
		Name:         preset.template.Meta.Name,
		WorkloadType: string(preset.template.Meta.WorkloadType),
	}
}

var presets = map[dao_crud.TemplateNameAndType]presetEntry{
	nameAndTypeFromPreset(nemoFourGpus):     nemoFourGpus,
	nameAndTypeFromPreset(bioNemoFourGpus):  bioNemoFourGpus,
	nameAndTypeFromPreset(jupyterLabOneGpu): jupyterLabOneGpu,
	nameAndTypeFromPreset(pytorchOneGpu):    pytorchOneGpu,

	nameAndTypeFromPreset(nemoFourGpusHost):     nemoFourGpusHost,
	nameAndTypeFromPreset(bioNemoFourGpusHost):  bioNemoFourGpusHost,
	nameAndTypeFromPreset(jupyterLabOneGpuHost): jupyterLabOneGpuHost,
	nameAndTypeFromPreset(pytorchOneGpuHost):    pytorchOneGpuHost,
}

// InitTemplatePresets makes sure to add all necessary presets to the DB
func InitTemplatePresets() {
	// Create a context with all necessary values for system operations
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfTenantId, systemTenantId)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, systemSubjectID)
	ctx = context.WithValue(ctx, runaictx.CfGroups, systemGroups)
	ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, systemPermittedScopes)

	ts := TemplatesService{}

	systemWorkloads, dbErr := dao_crud.ListPresetTemplates(ctx)
	if dbErr != nil {
		logging.Errorf("error listing template presets: %v", dbErr)
		return
	}

	for nameAndType, preset := range presets {
		existingTemplate := systemWorkloads[nameAndType]

		// case 1: creating preset that not yet exist
		if existingTemplate == nil {
			ts.createPreset(ctx, nameAndType, preset)
			continue
		}

		// case 2: updating template that has changes. we know that it has changed because we
		// keep the version of the template in the description, and change it every time we want
		// to change its spec
		if convutils.Deref(existingTemplate.Description, "") != convutils.Deref(preset.template.Meta.Description, "") {
			ts.updatePreset(ctx, existingTemplate.ID.String(), nameAndType, preset)
		}
	}
}

func (ts *TemplatesService) createPreset(ctx context.Context, nameAndType dao_crud.TemplateNameAndType, preset presetEntry) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("creating preset template: %+v", nameAndType)

	err := ts.CreateTemplate(ctx, &UnifiedRequest{
		workloadType:  api.WorkloadType(nameAndType.WorkloadType),
		template:      &preset.template,
		migrating:     true,
		routingMethod: preset.routingMethod,
	})
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("error creating preset template: %v", err)
	}
}

func (ts *TemplatesService) updatePreset(ctx context.Context, id string, nameAndType dao_crud.TemplateNameAndType, preset presetEntry) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("updating preset template %s: %+v to %v", id, nameAndType, preset.template.Meta.Description)

	_, err := ts.UpdateTemplate(ctx, id, &UnifiedRequest{
		workloadType:  api.WorkloadType(nameAndType.WorkloadType),
		template:      &preset.template,
		migrating:     true,
		routingMethod: preset.routingMethod,
	}, updateOptions{})
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("error updating preset template: %v", err)
	}
}
