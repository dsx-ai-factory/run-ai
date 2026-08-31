/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping

import (
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
)

// BuildSystemNimCacheSpecFormatterFields builds the runai-nimCacheSpec system
// formatter, holding fields whose argsEnum we want enforced at write time
// (string fields whose apps.nvidia.com/v1alpha1/NIMCache values are restricted
// to a closed set), or which need object-level setting (source types, storage).
//
// The setExprs are rooted at the spec-selector anchor (the NIMCache .spec).
func BuildSystemNimCacheSpecFormatterFields() map[string]api.FormatterField {
	m := make(map[string]api.FormatterField)

	////////////////////////////////////////////////////////
	// ngcSource – sets the whole source.ngc object in one go
	//
	// Required args: authSecret, modelPuller (both strings)
	// Optional args: pullSecret, model (object with profiles, precision, engine,
	//                tensorParallelism, qosProfile, lora, buildable)
	////////////////////////////////////////////////////////
	addNimCacheSpecField(m, "ngcSource", `{
	  "source": {
		"ngc": {
		  "authSecret": $ARGS.named.authSecret,
		  "modelPuller": $ARGS.named.modelPuller,
		  "pullSecret": $ARGS.named.pullSecret,
		  "model": $ARGS.named.model
		}
	  }
	}`, nil, nil,
		map[string]api.MappingFieldType{
			"authSecret":  api.MAPPINGFIELDTYPE_STRING,
			"modelPuller": api.MAPPINGFIELDTYPE_STRING,
			"pullSecret":  api.MAPPINGFIELDTYPE_STRING,
			"model":       api.MAPPINGFIELDTYPE_OBJECT,
		}, nil)

	////////////////////////////////////////////////////////
	// dataStoreSource – sets the whole source.dataStore object
	//
	// All fields are optional strings: endpoint, namespace, authSecret,
	// modelPuller, pullSecret, modelName, datasetName, revision.
	////////////////////////////////////////////////////////
	addNimCacheSpecField(m, "dataStoreSource", `{
	  "source": {
		"dataStore": {
		  "endpoint": $ARGS.named.endpoint,
		  "namespace": $ARGS.named.namespace,
		  "authSecret": $ARGS.named.authSecret,
		  "modelPuller": $ARGS.named.modelPuller,
		  "pullSecret": $ARGS.named.pullSecret,
		  "modelName": $ARGS.named.modelName,
		  "datasetName": $ARGS.named.datasetName,
		  "revision": $ARGS.named.revision
		}
	  }
	}`, nil, nil,
		map[string]api.MappingFieldType{
			"endpoint":    api.MAPPINGFIELDTYPE_STRING,
			"namespace":   api.MAPPINGFIELDTYPE_STRING,
			"authSecret":  api.MAPPINGFIELDTYPE_STRING,
			"modelPuller": api.MAPPINGFIELDTYPE_STRING,
			"pullSecret":  api.MAPPINGFIELDTYPE_STRING,
			"modelName":   api.MAPPINGFIELDTYPE_STRING,
			"datasetName": api.MAPPINGFIELDTYPE_STRING,
			"revision":    api.MAPPINGFIELDTYPE_STRING,
		}, nil)

	////////////////////////////////////////////////////////
	// hfSource – sets the whole source.hf object
	//
	// All fields are optional strings: endpoint, namespace, authSecret,
	// modelPuller, pullSecret, modelName, datasetName, revision.
	////////////////////////////////////////////////////////
	addNimCacheSpecField(m, "hfSource", `{
	  "source": {
		"hf": {
		  "endpoint": $ARGS.named.endpoint,
		  "namespace": $ARGS.named.namespace,
		  "authSecret": $ARGS.named.authSecret,
		  "modelPuller": $ARGS.named.modelPuller,
		  "pullSecret": $ARGS.named.pullSecret,
		  "modelName": $ARGS.named.modelName,
		  "datasetName": $ARGS.named.datasetName,
		  "revision": $ARGS.named.revision
		}
	  }
	}`, nil, nil,
		map[string]api.MappingFieldType{
			"endpoint":    api.MAPPINGFIELDTYPE_STRING,
			"namespace":   api.MAPPINGFIELDTYPE_STRING,
			"authSecret":  api.MAPPINGFIELDTYPE_STRING,
			"modelPuller": api.MAPPINGFIELDTYPE_STRING,
			"pullSecret":  api.MAPPINGFIELDTYPE_STRING,
			"modelName":   api.MAPPINGFIELDTYPE_STRING,
			"datasetName": api.MAPPINGFIELDTYPE_STRING,
			"revision":    api.MAPPINGFIELDTYPE_STRING,
		}, nil)

	////////////////////////////////////////////////////////
	// pvcStorage – sets the whole storage.pvc object plus storage.readOnly
	//
	// The readOnly arg maps to storage.readOnly (a sibling of storage.pvc).
	// This is synonymous with hostPathStorage's readOnly – both end up at
	// the same JSON path. Documented so policy writers know setting either
	// one's readOnly will affect both if both storage types are (unusually)
	// defined.
	//
	// Required: name (string)
	// Optional: create (bool), storageClass (string), size (string),
	//           subPath (string), volumeAccessMode (enum), readOnly (bool)
	////////////////////////////////////////////////////////
	addNimCacheSpecField(m, "pvcStorage", `{
	  "storage": {
		"readOnly": $ARGS.named.readOnly,
		"pvc": {
		  "name": $ARGS.named.name,
		  "create": $ARGS.named.create,
		  "storageClass": $ARGS.named.storageClass,
		  "size": $ARGS.named.size,
		  "subPath": $ARGS.named.subPath,
		  "volumeAccessMode": $ARGS.named.volumeAccessMode
		}
	  }
	}`, nil, nil,
		map[string]api.MappingFieldType{
			"name":             api.MAPPINGFIELDTYPE_STRING,
			"create":           api.MAPPINGFIELDTYPE_BOOLEAN,
			"storageClass":     api.MAPPINGFIELDTYPE_STRING,
			"size":             api.MAPPINGFIELDTYPE_STRING,
			"subPath":          api.MAPPINGFIELDTYPE_STRING,
			"volumeAccessMode": api.MAPPINGFIELDTYPE_STRING,
			"readOnly":         api.MAPPINGFIELDTYPE_BOOLEAN,
		},
		map[string][]string{
			"volumeAccessMode": {"ReadWriteOnce", "ReadOnlyMany", "ReadWriteMany", "ReadWriteOncePod"},
		})

	////////////////////////////////////////////////////////
	// hostPathStorage – sets storage.hostPath plus storage.readOnly
	//
	// The readOnly arg maps to storage.readOnly (a sibling of storage.hostPath).
	// Synonymous behaviour: see pvcStorage notes above.
	//
	// Required: path (string)
	// Optional: readOnly (bool)
	////////////////////////////////////////////////////////
	addNimCacheSpecField(m, "hostPathStorage", `{
	  "storage": {
		"readOnly": $ARGS.named.readOnly,
		"hostPath": $ARGS.named.path
	  }
	}`, nil, nil,
		map[string]api.MappingFieldType{
			"path":     api.MAPPINGFIELDTYPE_STRING,
			"readOnly": api.MAPPINGFIELDTYPE_BOOLEAN,
		}, nil)

	return m
}

func addNimCacheSpecField(
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
