/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"knative.dev/pkg/logging"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	workloadproject "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/project"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

type ValueFrom int

const (
	ValueFromNotFound ValueFrom = iota
	ValueFromSecret
	ValueFromConfigMap
	ValueFromPodFieldRef
	ValueFromUserCredentials

	SecretReferenceInEnvVar    = "SECRET"
	ConfigMapReferenceInEnvVar = "CONFIGMAP"
	PodFieldRefInEnvVar        = "PODFIELDREF"
	UserCredentialEnvVar       = "USERCREDENTIAL"
)

var (
	ErrNameValueIsEmpty = errors.New("name value is empty")
	ErrMissingKey       = errors.New("key value is empty")
	ErrMissingPath      = errors.New("path value is empty")
	ErrUnFamiliarType   = errors.New("missing value from reference object")
)

func ValueFromString(value string) ValueFrom {
	switch value {
	case ConfigMapReferenceInEnvVar:
		return ValueFromConfigMap
	case SecretReferenceInEnvVar:
		return ValueFromSecret
	case PodFieldRefInEnvVar:
		return ValueFromPodFieldRef
	case UserCredentialEnvVar:
		return ValueFromUserCredentials
	default:
		return ValueFromNotFound
	}
}

func (rp *ValueFrom) EnvVarSource(name, attribute string) (*corev1.EnvVarSource, error) {
	switch *rp {
	case ValueFromConfigMap:
		if name == "" {
			return nil, ErrNameValueIsEmpty
		}
		if attribute == "" {
			return nil, ErrMissingKey
		}
		return &corev1.EnvVarSource{
			ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: name,
				},
				Key: attribute,
			},
		}, nil
	case ValueFromSecret:
		if name == "" {
			return nil, ErrNameValueIsEmpty
		}
		return &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: name,
				},
				Key: attribute,
			},
		}, nil
	case ValueFromPodFieldRef:
		if name == "" {
			return nil, ErrMissingPath
		}
		return &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: name,
			},
		}, nil
	case ValueFromUserCredentials:
		if name == "" {
			return nil, ErrNameValueIsEmpty
		}
		return &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: name,
				},
				Key: attribute,
			},
		}, nil
	default:
		// if we reach this point, this is our bug.
		return nil, ErrUnFamiliarType
	}
}

func (rp *ValueFrom) getFieldsOfRef(envVarValue string) (string, string) {
	slice := strings.Split(envVarValue, ",")
	resourceName := slice[0]
	if len(slice) > 1 {
		return resourceName, slice[1]
	} else {
		return resourceName, ""
	}
}

func (rp *ValueFrom) patchName(envVarValue string, workloadInitial string, workloadName string) string {
	if envVarValue == "" {
		return ""
	}
	if *rp == ValueFromUserCredentials {
		return fmt.Sprintf("%s-%s-%s", workloadInitial, workloadName, envVarValue)
	}
	return envVarValue
}

func ConstructEnvVar(ctx context.Context, varName string, value *string, containerFields *field_groups.ContainerFields, objNaming *metav1.ObjectMeta, workloadName string, workloadInitial string) corev1.EnvVar {
	varValue := ptr.Deref(value, "")

	envVar := corev1.EnvVar{
		Name: varName,
	}

	// Lookup if value has unique environment source and return object if possible, else nil
	// SECRET:<key>,<name>/CONFIGMAP:<key>,<name>/PODFIELDREF:<path>
	fromValue, err := getEnvVarSource(varValue, workloadInitial, workloadName)
	if err != nil {
		logger := logging.FromContext(ctx)
		// we just print an error we don't want to prevent the environment set
		logger.Warnf(fmt.Sprintf("prefix for FromValue found (%s), but failed to create source, err: %v", varValue, err))
	} else if fromValue != nil {
		envVar.ValueFrom = fromValue
		return envVar
	}

	//
	//   if value is not aligned with SECRET: or CONFIGMAP: formats, treat it as
	//   plain text.
	//
	envVar.Value = PatchTemplateVariables(varValue, containerFields, objNaming, ctx)

	return envVar
}

// DecodeValue returns the ValueFrom type and the value from the environment variable value.
func DecodeValue(value string) (ValueFrom, string, string) {
	vf, value := getValueFrom(value)
	filedName, fieldValue := vf.getFieldsOfRef(value)
	return vf, filedName, fieldValue
}

func getEnvVarSource(value string, initials string, workloadName string) (*corev1.EnvVarSource, error) {
	if refType, refValue := getValueFrom(value); refType != ValueFromNotFound {
		filedName, fieldValue := refType.getFieldsOfRef(refValue)
		filedName = refType.patchName(filedName, initials, workloadName)
		evs, err := refType.EnvVarSource(filedName, fieldValue)
		if err != nil {
			return nil, err
		}
		return evs, nil
	}
	return nil, nil
}

func getValueFrom(envVarValue string) (ValueFrom, string) {
	slice := strings.Split(envVarValue, ":")
	if len(slice) < 2 {
		return ValueFromNotFound, ""
	}
	ref := ValueFromString(slice[0])
	if ref == ValueFromNotFound {
		return ValueFromNotFound, ""
	}
	return ref, slice[1]
}

func PatchTemplateVariables(varValue string, containerFields *field_groups.ContainerFields, objNaming *metav1.ObjectMeta, ctx context.Context) string {
	varValue = PatchUidGid(varValue, containerFields)
	varValue = patchJobName(varValue, objNaming)
	varValue = patchProjectName(varValue, ctx)
	return varValue
}

func PatchUidGid(varValue string, containerFields *field_groups.ContainerFields) string {
	if containerFields != nil {
		if containerFields.RunAsUid != nil {
			varValue = strings.ReplaceAll(varValue, "${RUNAI_UID}", fmt.Sprintf("%d", containerFields.RunAsUid.Deref()))
		}
		if containerFields.RunAsGid != nil {
			varValue = strings.ReplaceAll(varValue, "${RUNAI_GID}", fmt.Sprintf("%d", containerFields.RunAsGid.Deref()))
		}
	}
	return varValue
}

func patchJobName(varValue string, objNaming *metav1.ObjectMeta) string {
	if objNaming != nil && varValue != "" {
		varValue = strings.ReplaceAll(varValue, "${RUNAI_JOB_NAME}", objNaming.Name)
	}
	return varValue
}

func patchProjectName(varValue string, ctx context.Context) string {
	if ctx != nil && varValue != "" {
		projectName, err := workloadproject.GetProjectFromContext(ctx)
		if err == nil {
			varValue = strings.ReplaceAll(varValue, "${RUNAI_PROJECT}", projectName)
		}
	}
	return varValue
}
