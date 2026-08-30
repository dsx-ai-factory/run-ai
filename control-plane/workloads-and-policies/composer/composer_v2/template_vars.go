package composer_v2

import (
	"strconv"
	"strings"

	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	corev1 "k8s.io/api/core/v1"
)

const (
	templateVarProject = "${RUNAI_PROJECT}"
	templateVarJobName = "${RUNAI_JOB_NAME}"
	templateVarUid     = "${RUNAI_UID}"
	templateVarGid     = "${RUNAI_GID}"
)

// ExpandTemplateVarsInCommandAndArgs is the native-path counterpart of the cluster
// builders' PatchTemplateVariables: the legacy CRD path expands these placeholders,
// but native manifests are final so the composer must. uid/gid come from the composed
// container security context, so this must run after the security step.
func ExpandTemplateVarsInCommandAndArgs(container *corev1.Container, projectName string, workloadName string) *workloadsApi.Error {
	var uid, gid *int64
	if container.SecurityContext != nil {
		uid = container.SecurityContext.RunAsUser
		gid = container.SecurityContext.RunAsGroup
	}
	expandRunaiTemplateVars(container.Command, projectName, workloadName, uid, gid)
	expandRunaiTemplateVars(container.Args, projectName, workloadName, uid, gid)
	return nil
}

func expandRunaiTemplateVars(values []string, projectName, jobName string, uid, gid *int64) {
	for i, value := range values {
		value = strings.ReplaceAll(value, templateVarProject, projectName)
		value = strings.ReplaceAll(value, templateVarJobName, jobName)
		if uid != nil {
			value = strings.ReplaceAll(value, templateVarUid, strconv.FormatInt(*uid, 10))
		}
		if gid != nil {
			value = strings.ReplaceAll(value, templateVarGid, strconv.FormatInt(*gid, 10))
		}
		values[i] = value
	}
}
