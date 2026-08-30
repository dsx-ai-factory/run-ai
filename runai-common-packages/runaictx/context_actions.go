package runaictx

func getRunaiCtxMapping() map[interface{}]string {
	return map[interface{}]string{
		CfTenantId:           "tenant_id",
		CfClientId:           "client_id",
		CfClientType:         "client_type",
		CfGroups:             "groups",
		CfSubjectID:          "subject_id",
		CfSubjectType:        "subject_type",
		CfSupplementalGroups: "supplemental_groups",
		CfTraceId:            "trace_id",
		CfSpanId:             "span_id",
		CfApplicationId:      "application_id",
	}
}

func toStringArray(arr []interface{}) []string {
	strArr := make([]string, len(arr))
	for i, e := range arr {
		strArr[i] = e.(string)
	}
	return strArr
}

func extractCtxFieldValue(ctxMap map[string]interface{}, fieldIdentifier string) any {
	val, ok := ctxMap[fieldIdentifier]
	if ok {
		switch val := val.(type) {
		case string:
			return val
		case []string:
			return val
		case []interface{}:
			return toStringArray(val)
		}
	}
	return nil
}
