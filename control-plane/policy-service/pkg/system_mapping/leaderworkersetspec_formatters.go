package system_mapping

import (
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
)

// BuildSystemLeaderWorkerSetSpecFormatterFields returns the formatter fields for LeaderWorkerSet.
// Provides enum validation for startupPolicy, rolloutStrategy-Type, and restartPolicy.
func BuildSystemLeaderWorkerSetSpecFormatterFields() map[string]api.FormatterField {
	m := make(map[string]api.FormatterField)

	// ===== Workload-level fields =====

	addLWSFormatterField(m, "replicas", `{
	  "replicas": ($ARGS.named.value | tonumber)
	}`, nil, nil)

	addLWSFormatterField(m, "startupPolicy", `{
	  "startupPolicy": $ARGS.named.value
	}`, nil, map[string][]string{
		"value": {"LeaderCreated", "LeaderReady"},
	})

	addLWSFormatterField(m, "networkConfig-SubdomainPolicy", `{
	  "networkConfig": {
		"subdomainPolicy": $ARGS.named.value
	  }
	}`, nil, map[string][]string{
		"value": {"Default", "Disabled", "Indexed"},
	})

	addLWSFormatterField(m, "rolloutStrategy-Type", `{
	  "rolloutStrategy": {
		"type": $ARGS.named.value
	  }
	}`, nil, map[string][]string{
		"value": {"RollingUpdate"},
	})

	addLWSFormatterField(m, "rolloutStrategy-MaxSurge", `{
	  "rolloutStrategy": {
		"rollingUpdateConfiguration": {
		  "maxSurge": $ARGS.named.value
		}
	  }
	}`, nil, nil)

	addLWSFormatterField(m, "rolloutStrategy-MaxUnavailable", `{
	  "rolloutStrategy": {
		"rollingUpdateConfiguration": {
		  "maxUnavailable": $ARGS.named.value
		}
	  }
	}`, nil, nil)

	// ===== Template-level fields =====

	addLWSFormatterField(m, "size", `{
	  "size": ($ARGS.named.value | tonumber)
	}`, nil, nil)

	addLWSFormatterField(m, "restartPolicy", `{
	  "restartPolicy": $ARGS.named.value
	}`, nil, map[string][]string{
		"value": {"Default", "RecreateGroupOnPodRestart", "None"},
	})

	addLWSFormatterField(m, "subGroupPolicy-SubGroupSize", `{
	  "subGroupPolicy": {
		"subGroupSize": ($ARGS.named.value | tonumber)
	  }
	}`, nil, nil)

	return m
}

func addLWSFormatterField(
	dst map[string]api.FormatterField,
	name string,
	setExpr string,
	argsInfo map[string]api.MappingFieldType,
	argsEnum map[string][]string,
) {
	ff := api.FormatterField{
		SetExpr: &setExpr,
	}
	if len(argsInfo) > 0 {
		ff.ArgsInfo = argsInfo
	}
	if len(argsEnum) > 0 {
		ff.ArgsEnum = argsEnum
	}
	dst[name] = ff
}
