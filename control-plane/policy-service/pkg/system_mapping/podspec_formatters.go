/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
)

func BuildSystemPodSpecFormatterFields() map[string]api.FormatterField {
	m := make(map[string]api.FormatterField)

	////////////////////////////////////////////////////////
	// dnsOptions
	////////////////////////////////////////////////////////
	addPodSpecField(m, "dnsOptions", `{
	  "dnsConfig": {
		"options": [
		  {
			"name": $ARGS.named.name,
			"value": $ARGS.named.value
		  }
		]
	  }
	}`, nil,
		&api.FormatterArrayKey{ArgName: "name"}, nil, nil)

	////////////////////////////////////////////////////////
	// environmentVariables-AsValue
	////////////////////////////////////////////////////////
	addPodSpecField(m, "environmentVariables-AsValue", `{
	  "env": [
		{
		  "name": $ARGS.named.name,
		  "value": $ARGS.named.value
		}
	  ]
	}`, nil,
		&api.FormatterArrayKey{ArgName: "name"}, nil, nil)

	////////////////////////////////////////////////////////
	// configMapVolumes
	////////////////////////////////////////////////////////
	addPodSpecField(m, "configMapVolumes", `{
	  "volumes": [
		{
		  "name": $ARGS.named.name,
		  "configMap": {
			"name": $ARGS.named.configMapName,
			"defaultMode": $ARGS.named.defaultMode,
			"items": $ARGS.named.items,
			"optional": $ARGS.named.optional
		  }
		}
	  ]
	}`, nil,
		&api.FormatterArrayKey{ArgName: "name"},
		map[string]api.MappingFieldType{
			"defaultMode": api.MAPPINGFIELDTYPE_INTEGER,
			"items":       api.MAPPINGFIELDTYPE_OBJECT,
			"optional":    api.MAPPINGFIELDTYPE_BOOLEAN,
		}, nil)

	////////////////////////////////////////////////////////
	// emptyDirVolumes
	////////////////////////////////////////////////////////
	addPodSpecField(m, "emptyDirVolumes", `{
	  "volumes": [
		{
		  "name": $ARGS.named.name,
		  "emptyDir": {
			"medium": $ARGS.named.medium,
			"sizeLimit": $ARGS.named.sizeLimit
		  }
		}
	  ]
	}`, nil,
		&api.FormatterArrayKey{ArgName: "name"},
		map[string]api.MappingFieldType{
			"sizeLimit": api.MAPPINGFIELDTYPE_QUANTITY,
		}, nil)

	////////////////////////////////////////////////////////
	// environmentVariables-FromConfigMap
	////////////////////////////////////////////////////////
	addPodSpecField(m, "environmentVariables-FromConfigMap", `{
	  "env": [
		{
		  "name": $ARGS.named.name,
		  "valueFrom": {
			"configMapKeyRef": {
			  "name": $ARGS.named.configMapName,
			  "key": $ARGS.named.key,
			  "optional": $ARGS.named.optional
			}
		  }
		}
	  ]
	}`, nil,
		&api.FormatterArrayKey{ArgName: "name"},
		map[string]api.MappingFieldType{
			"optional": api.MAPPINGFIELDTYPE_BOOLEAN,
		}, nil)

	////////////////////////////////////////////////////////
	// environmentVariables-FromField
	////////////////////////////////////////////////////////
	addPodSpecField(m, "environmentVariables-FromField", `{
	  "env": [
		{
		  "name": $ARGS.named.name,
		  "valueFrom": {
			"fieldRef": {
			  "apiVersion": $ARGS.named.apiVersion,
			  "fieldPath": $ARGS.named.fieldPath
			}
		  }
		}
	  ]
	}`, nil,
		&api.FormatterArrayKey{ArgName: "name"}, nil, nil)

	////////////////////////////////////////////////////////
	// environmentVariables-FromResource
	////////////////////////////////////////////////////////
	addPodSpecField(m, "environmentVariables-FromResource", `{
	  "env": [
		{
		  "name": $ARGS.named.name,
		  "valueFrom": {
			"resourceFieldRef": {
			  "containerName": $ARGS.named.containerName,
			  "divisor": $ARGS.named.divisor,
			  "resource": $ARGS.named.resource
			}
		  }
		}
	  ]
	}`, nil,
		&api.FormatterArrayKey{ArgName: "name"},
		map[string]api.MappingFieldType{
			"divisor": api.MAPPINGFIELDTYPE_QUANTITY,
		}, nil)

	////////////////////////////////////////////////////////
	// environmentVariables-FromSecret
	////////////////////////////////////////////////////////
	addPodSpecField(m, "environmentVariables-FromSecret", `{
	  "env": [
		{
		  "name": $ARGS.named.name,
		  "valueFrom": {
			"secretKeyRef": {
			  "name": $ARGS.named.secretName,
			  "key": $ARGS.named.key,
			  "optional": $ARGS.named.optional
			}
		  }
		}
	  ]
	}`, nil,
		&api.FormatterArrayKey{ArgName: "name"},
		map[string]api.MappingFieldType{
			"optional": api.MAPPINGFIELDTYPE_BOOLEAN,
		}, nil)

	////////////////////////////////////////////////////////
	// environmentVariableBulks-FromConfigMap
	////////////////////////////////////////////////////////
	addPodSpecField(m, "environmentVariableBulks-FromConfigMap", `{
	  "envFrom": [
		{
		  "configMapRef": {
			"name": $ARGS.named.configMapName,
			"optional": $ARGS.named.optional
		  },
		  "prefix": $ARGS.named.prefix
		}
	  ]
	}`, nil,
		&api.FormatterArrayKey{
			ArgName: "configMapName",
			SpecRef: convutils.Ref("configMapRef.name"),
		}, map[string]api.MappingFieldType{
			"optional": api.MAPPINGFIELDTYPE_BOOLEAN,
		}, nil)

	////////////////////////////////////////////////////////
	// environmentVariableBulks-FromSecret
	////////////////////////////////////////////////////////
	addPodSpecField(m, "environmentVariableBulks-FromSecret", `{
	  "envFrom": [
		{
		  "secretRef": {
			"name": $ARGS.named.secretName,
			"optional": $ARGS.named.optional
		  },
		  "prefix": $ARGS.named.prefix
		}
	  ]
	}`, nil,
		&api.FormatterArrayKey{
			ArgName: "secretName",
			SpecRef: convutils.Ref("secretRef.name"),
		}, map[string]api.MappingFieldType{
			"optional": api.MAPPINGFIELDTYPE_BOOLEAN,
		}, nil)

	////////////////////////////////////////////////////////
	// hostAliases
	////////////////////////////////////////////////////////
	addPodSpecField(m, "hostAliases", `{
	  "hostAliases": [
		{
		  "ip": $ARGS.named.ip,
		  "hostnames": $ARGS.named.hostnames
		}
	  ]
	}`, nil,
		&api.FormatterArrayKey{ArgName: "ip"},
		map[string]api.MappingFieldType{
			"hostnames": api.MAPPINGFIELDTYPE_STRINGS,
			"ip":        api.MAPPINGFIELDTYPE_IPADDRESS,
		}, nil)

	////////////////////////////////////////////////////////
	// hostPathVolumes
	////////////////////////////////////////////////////////
	addPodSpecField(m, "hostPathVolumes", `{
	  "volumes": [
		{
		  "name": $ARGS.named.name,
		  "hostPath": {
			"path": $ARGS.named.path,
			"type": $ARGS.named.type
		  }
		}
	  ]
	}`, nil,
		&api.FormatterArrayKey{ArgName: "name"}, nil,
		map[string][]string{
			"type": {"", "BlockDevice", "CharDevice", "Directory", "DirectoryOrCreate", "File", "FileOrCreate", "Socket"},
		})

	////////////////////////////////////////////////////////
	// nodeSelectors
	////////////////////////////////////////////////////////
	addPodSpecField(m, "nodeSelectors", `{
	  "nodeSelector": {
		($ARGS.named.key): $ARGS.named.value
	  }
	}`, nil, &api.FormatterArrayKey{ArgName: "key"}, nil, nil)

	////////////////////////////////////////////////////////
	// nfsVolumes
	////////////////////////////////////////////////////////
	addPodSpecField(m, "nfsVolumes", `{
	  "volumes": [
		{
		  "name": $ARGS.named.name,
		  "nfs": {
			"server": $ARGS.named.server,
			"path": $ARGS.named.path,
			"readOnly": $ARGS.named.readOnly
		  }
		}
	  ]
	}`, nil,
		&api.FormatterArrayKey{ArgName: "name"},
		map[string]api.MappingFieldType{
			"readOnly": api.MAPPINGFIELDTYPE_BOOLEAN,
		}, nil)

	////////////////////////////////////////////////////////
	// ports
	////////////////////////////////////////////////////////
	addPodSpecField(m, "ports", `{
	  "ports": [
		{
		  "containerPort": $ARGS.named.containerPort,
		  "hostIP": $ARGS.named.hostIp,
		  "hostPort": $ARGS.named.hostPort,
		  "name": $ARGS.named.name,
		  "protocol": $ARGS.named.protocol
		}
	  ]
	}`, nil,
		&api.FormatterArrayKey{ArgName: "containerPort"},
		map[string]api.MappingFieldType{
			"containerPort": api.MAPPINGFIELDTYPE_INTEGER,
			"hostIp":        api.MAPPINGFIELDTYPE_IPADDRESS,
			"hostPort":      api.MAPPINGFIELDTYPE_INTEGER,
		}, map[string][]string{
			"protocol": {"TCP", "UDP", "SCTP"},
		})

	////////////////////////////////////////////////////////
	// pvcVolumes
	////////////////////////////////////////////////////////
	addPodSpecField(m, "pvcVolumes", `{
	  "volumes": [
		{
		  "name": $ARGS.named.name,
		  "persistentVolumeClaim": {
			"claimName": $ARGS.named.claimName,
			"readOnly": $ARGS.named.readOnly
		  }
		}
	  ]
	}`, nil,
		&api.FormatterArrayKey{ArgName: "name"},
		map[string]api.MappingFieldType{
			"readOnly": api.MAPPINGFIELDTYPE_BOOLEAN,
		}, nil)

	////////////////////////////////////////////////////////
	// readinessGates
	////////////////////////////////////////////////////////
	addPodSpecField(m, "readinessGates", `{
	  "readinessGates": [
		{
		  "conditionType": $ARGS.named.conditionType
		}
	  ]
	}`, nil,
		&api.FormatterArrayKey{ArgName: "conditionType"}, nil, nil)

	////////////////////////////////////////////////////////
	// resourceClaims
	////////////////////////////////////////////////////////
	addPodSpecField(m, "resourceClaims", `{
	  "resourceClaims": [
		{
		  "name": $ARGS.named.name,
		  "resourceClaimName": $ARGS.named.resourceClaimName,
		  "resourceClaimTemplateName": $ARGS.named.resourceClaimTemplateName
		}
	  ]
	}`, nil,
		&api.FormatterArrayKey{ArgName: "name"}, nil, nil)

	////////////////////////////////////////////////////////
	// resourceLimits
	////////////////////////////////////////////////////////
	addPodSpecField(m, "resourceLimits", `{
	  "resources": {
		"limits": {
		  ($ARGS.named.key): $ARGS.named.value
		}
	  }
	}`, nil, &api.FormatterArrayKey{ArgName: "key"},
		map[string]api.MappingFieldType{
			"value": api.MAPPINGFIELDTYPE_QUANTITY,
		}, nil)

	////////////////////////////////////////////////////////
	// resourceRequests
	////////////////////////////////////////////////////////
	addPodSpecField(m, "resourceRequests", `{
	  "resources": {
		"requests": {
		  ($ARGS.named.key): $ARGS.named.value
		}
	  }
	}`, nil, &api.FormatterArrayKey{ArgName: "key"},
		map[string]api.MappingFieldType{
			"value": api.MAPPINGFIELDTYPE_QUANTITY,
		}, nil)

	////////////////////////////////////////////////////////
	// resourceOverheads
	////////////////////////////////////////////////////////
	addPodSpecField(m, "resourceOverheads", `{
	  "overhead": {
		($ARGS.named.key): $ARGS.named.value
	  }
	}`, nil, &api.FormatterArrayKey{ArgName: "key"},
		map[string]api.MappingFieldType{
			"value": api.MAPPINGFIELDTYPE_QUANTITY,
		}, nil)

	////////////////////////////////////////////////////////
	// secretVolumes
	////////////////////////////////////////////////////////
	addPodSpecField(m, "secretVolumes", `{
	  "volumes": [
		{
		  "name": $ARGS.named.name,
		  "secret": {
			"secretName": $ARGS.named.secretName,
			"defaultMode": $ARGS.named.defaultMode,
			"items": $ARGS.named.items,
			"optional": $ARGS.named.optional
		  }
		}
	  ]
	}`, nil,
		&api.FormatterArrayKey{ArgName: "name"},
		map[string]api.MappingFieldType{
			"defaultMode": api.MAPPINGFIELDTYPE_INTEGER,
			"items":       api.MAPPINGFIELDTYPE_OBJECT,
			"optional":    api.MAPPINGFIELDTYPE_BOOLEAN,
		}, nil)

	////////////////////////////////////////////////////////
	// tolerations
	////////////////////////////////////////////////////////
	addPodSpecField(m, "tolerations", `{
	  "tolerations": [
		{
		  "key": $ARGS.named.key,
		  "operator": $ARGS.named.operator,
		  "value": $ARGS.named.value,
		  "effect": $ARGS.named.effect,
		  "tolerationSeconds": $ARGS.named.seconds
		}
	  ]
	}`, nil, nil,
		map[string]api.MappingFieldType{
			"seconds": api.MAPPINGFIELDTYPE_INTEGER,
		},
		map[string][]string{
			"operator": {"Equal", "Exists"},
			"effect":   {"NoSchedule", "PreferNoSchedule", "NoExecute"},
		})

	////////////////////////////////////////////////////////
	// volumeMounts
	////////////////////////////////////////////////////////
	addPodSpecField(m, "volumeMounts", `{
	  "volumeMounts": [
		{
		  "name": $ARGS.named.name,
		  "mountPath": $ARGS.named.mountPath,
		  "mountPropagation": $ARGS.named.mountPropagation,
		  "readOnly": $ARGS.named.readOnly,
		  "recursiveReadOnly": $ARGS.named.recursiveReadOnly,
		  "subPath": $ARGS.named.subPath,
		  "subPathExpr": $ARGS.named.subPathExpr
		}
	  ]
	}`, nil,
		&api.FormatterArrayKey{ArgName: "name"},
		map[string]api.MappingFieldType{
			"readOnly": api.MAPPINGFIELDTYPE_BOOLEAN,
		}, map[string][]string{
			"mountPropagation":  {"None", "HostToContainer", "Bidirectional"},
			"recursiveReadOnly": {"Disabled", "Enabled", "IfPossible"},
		})

	return m
}

func addPodSpecField(
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
