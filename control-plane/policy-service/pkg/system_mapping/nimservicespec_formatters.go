/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping

import (
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
)

// BuildSystemNimServiceSpecFormatterFields builds the runai-nimServiceSpec
// system formatter, holding fields whose argsEnum we want enforced at write
// time (string fields whose apps.nvidia.com/v1alpha1/NIMService values are
// restricted to a closed set).
//
// The setExprs are rooted at the spec-selector anchor (the NIMService .spec).
func BuildSystemNimServiceSpecFormatterFields() map[string]api.FormatterField {
	m := make(map[string]api.FormatterField)

	////////////////////////////////////////////////////////
	// inferencePlatform
	////////////////////////////////////////////////////////
	addNimServiceSpecField(m, "inferencePlatform", `{
	  "inferencePlatform": $ARGS.named.value
	}`, nil, nil, nil,
		map[string][]string{
			"value": {"standalone", "kserve"},
		})

	////////////////////////////////////////////////////////
	// image-PullPolicy
	////////////////////////////////////////////////////////
	addNimServiceSpecField(m, "image-PullPolicy", `{
	  "image": {
		"pullPolicy": $ARGS.named.value
	  }
	}`, nil, nil, nil,
		map[string][]string{
			"value": {"Always", "Never", "IfNotPresent"},
		})

	////////////////////////////////////////////////////////
	// exposeService-Type
	////////////////////////////////////////////////////////
	addNimServiceSpecField(m, "exposeService-Type", `{
	  "expose": {
		"service": {
		  "type": $ARGS.named.value
		}
	  }
	}`, nil, nil, nil,
		map[string][]string{
			"value": {"ClusterIP", "NodePort", "LoadBalancer", "ExternalName"},
		})

	////////////////////////////////////////////////////////
	// multiNode-BackendType (only "lws" is valid per CRD)
	////////////////////////////////////////////////////////
	addNimServiceSpecField(m, "multiNode-BackendType", `{
	  "multiNode": {
		"backendType": $ARGS.named.value
	  }
	}`, nil, nil, nil,
		map[string][]string{
			"value": {"lws"},
		})

	////////////////////////////////////////////////////////
	// hostPathStorage – sets storage.hostPath plus storage.readOnly
	//
	// The readOnly arg maps to storage.readOnly (a sibling of storage.hostPath).
	// Synonymous behaviour: nimCacheStorage's readOnly also maps to the same
	// path. Setting readOnly via either field has the same effect.
	//
	// Required: path (string)
	// Optional: readOnly (bool)
	////////////////////////////////////////////////////////
	addNimServiceSpecField(m, "hostPathStorage", `{
	  "storage": {
		"readOnly": $ARGS.named.readOnly,
		"hostPath": $ARGS.named.path
	  }
	}`, nil, nil,
		map[string]api.MappingFieldType{
			"path":     api.MAPPINGFIELDTYPE_STRING,
			"readOnly": api.MAPPINGFIELDTYPE_BOOLEAN,
		}, nil)

	////////////////////////////////////////////////////////
	// nimCacheStorage – sets the whole NIMCache reference plus PVC-based
	// fallback storage in a single object-level SET
	//
	// The readOnly arg maps to storage.readOnly (a sibling of storage.nimCache
	// and storage.pvc). Synonymous with hostPathStorage's readOnly.
	//
	// Optional args for NIMCache reference: name, profile (both strings)
	// Optional args for PVC-based storage: pvcName, pvcCreate (bool),
	//   pvcStorageClass, pvcSize, pvcSubPath, pvcVolumeAccessMode (strings)
	// Optional: readOnly (bool)
	////////////////////////////////////////////////////////
	addNimServiceSpecField(m, "nimCacheStorage", `{
	  "storage": {
		"readOnly": $ARGS.named.readOnly,
		"nimCache": {
		  "name": $ARGS.named.name,
		  "profile": $ARGS.named.profile
		},
		"pvc": {
		  "name": $ARGS.named.pvcName,
		  "create": $ARGS.named.pvcCreate,
		  "storageClass": $ARGS.named.pvcStorageClass,
		  "size": $ARGS.named.pvcSize,
		  "subPath": $ARGS.named.pvcSubPath,
		  "volumeAccessMode": $ARGS.named.pvcVolumeAccessMode
		}
	  }
	}`, nil, nil,
		map[string]api.MappingFieldType{
			"name":                api.MAPPINGFIELDTYPE_STRING,
			"profile":             api.MAPPINGFIELDTYPE_STRING,
			"pvcName":             api.MAPPINGFIELDTYPE_STRING,
			"pvcCreate":           api.MAPPINGFIELDTYPE_BOOLEAN,
			"pvcStorageClass":     api.MAPPINGFIELDTYPE_STRING,
			"pvcSize":             api.MAPPINGFIELDTYPE_STRING,
			"pvcSubPath":          api.MAPPINGFIELDTYPE_STRING,
			"pvcVolumeAccessMode": api.MAPPINGFIELDTYPE_STRING,
			"readOnly":            api.MAPPINGFIELDTYPE_BOOLEAN,
		},
		map[string][]string{
			"pvcVolumeAccessMode": {"ReadWriteOnce", "ReadOnlyMany", "ReadWriteMany", "ReadWriteOncePod"},
		})

	return m
}

func addNimServiceSpecField(
	dst map[string]api.FormatterField,
	name string,
	setExpr string,
	getExpr *string,
	arrayKey *api.FormatterArrayKey,
	argsInfo map[string]api.MappingFieldType,
	argsEnum map[string][]string,
) {
	ff := api.FormatterField{
		SetExpr:  &setExpr,
		GetExpr:  getExpr,
		ArrayKey: arrayKey,
	}
	if len(argsInfo) > 0 {
		ff.ArgsInfo = argsInfo
	}
	if len(argsEnum) > 0 {
		ff.ArgsEnum = argsEnum
	}
	dst[name] = ff
}
