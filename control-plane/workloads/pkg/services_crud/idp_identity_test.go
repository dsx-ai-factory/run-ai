/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	policy "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/policy_compliance"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

func TestInferenceOmittedIdpIdentity(t *testing.T) {
	str := func(value string) *string { return &value }
	number := func(value int64) *int64 { return &value }
	claims := []struct {
		name                           string
		uidClaim, gidClaim, groupClaim *string
		uid, gid                       *int64
		groups                         *string
	}{
		{"valid claims", str("2001"), str("2002"), str("2003,2004"), number(2001), number(2002), str("2003,2004")},
		{"missing claims", nil, nil, nil, nil, nil, nil},
		{"partial claims", str("2001"), nil, nil, number(2001), nil, nil},
		{"invalid claims", str("invalid"), str("invalid"), str("invalid"), nil, nil, nil},
		{"empty claims", str(""), str(""), str(""), nil, nil, nil},
		{"zero claims", str("0"), str("0"), str("0"), number(0), number(0), str("0")},
		{"mixed groups", nil, nil, str(" 2003,invalid,2004 "), nil, nil, str("2003,2004")},
	}
	for _, workloadType := range []string{Inference, DistributedInference} {
		for _, withDefaults := range []bool{false, true} {
			prefix := workloadType + "/without policy defaults/"
			if withDefaults {
				prefix = workloadType + "/with locked policy defaults/"
			}
			for _, tc := range claims {
				t.Run(prefix+tc.name, func(t *testing.T) {
					var defaults, rules map[string]interface{}
					uid, gid, groups := tc.uid, tc.gid, tc.groups
					if withDefaults {
						defaults = readIdentitySpec(t, `{"security":{"uidGidSource":"fromIdpToken","runAsUid":10001,"runAsGid":10002,"supplementalGroups":"10003"}}`)
						rules = readIdentitySpec(t, `{"security":{"runAsUid":{"canEdit":false},"runAsGid":{"canEdit":false},"supplementalGroups":{"canEdit":false}}}`)
						if uid == nil {
							uid = number(10001)
						}
						if gid == nil {
							gid = number(10002)
						}
						if groups == nil {
							groups = str("10003")
						}
					}
					input := readIdentitySpec(t, `{"security":{"uidGidSource":"fromIdpToken","hostIpc":true}}`)
					if workloadType == DistributedInference {
						input = map[string]interface{}{"leader": input, "worker": readIdentitySpec(t, `{"security":{"uidGidSource":"fromIdpToken","hostIpc":true}}`)}
						if withDefaults {
							defaults = map[string]interface{}{"leader": defaults, "worker": defaults}
							rules = map[string]interface{}{"leader": rules, "worker": rules}
						}
					}
					mutated, err := MutateWorkload(input, defaults, false, nil, nil)
					if err != nil {
						t.Fatal(err)
					}
					var locked []string
					var complianceErrors policy.ComplianceErrors
					policy.ValidateCompliance("", rules, defaults, mutated, nil, &locked, nil, &complianceErrors)
					if len(complianceErrors) != 0 {
						t.Fatalf("omitted identity violates policy: %+v", complianceErrors)
					}
					spec, _, buildErr := buildResponseSpecFromMap[workloadsApi.SupersetSpec, workloadsApi.MasterSpec](workloadType, mutated)
					if buildErr != nil {
						t.Fatal(buildErr)
					}
					ctx := context.Background()
					if tc.uidClaim != nil {
						ctx = context.WithValue(ctx, runaictx.CfUId, *tc.uidClaim)
					}
					if tc.gidClaim != nil {
						ctx = context.WithValue(ctx, runaictx.CfGId, *tc.gidClaim)
					}
					if tc.groupClaim != nil {
						ctx = context.WithValue(ctx, runaictx.CfSupplementalGroups, *tc.groupClaim)
					}
					var securitySections []map[string]interface{}
					if workloadType == DistributedInference {
						if err := setTokenUIDGIDForDistributedInference(ctx, spec); err != nil {
							t.Fatal(err)
						}
						securitySections = []map[string]interface{}{identityAsMap(t, spec.Leader.Get().GetSecurity()), identityAsMap(t, spec.Worker.Get().GetSecurity())}
					} else {
						security, err := setTokenUIDGIDForSecurity(ctx, spec.GetSecurity())
						if err != nil {
							t.Fatal(err)
						}
						securitySections = []map[string]interface{}{identityAsMap(t, security)}
					}
					for _, security := range securitySections {
						expected := map[string]interface{}{"uidGidSource": "fromIdpToken", "hostIpc": true}
						if uid != nil {
							expected["runAsUid"] = *uid
						}
						if gid != nil {
							expected["runAsGid"] = *gid
						}
						if groups != nil {
							expected["supplementalGroups"] = *groups
						}
						if !reflect.DeepEqual(security, identityAsMap(t, expected)) {
							t.Fatalf("unexpected security: got %v, want %v", security, expected)
						}
					}
				})
			}
		}
	}
}

func TestDistributedInferenceIdentitySources(t *testing.T) {
	for _, sources := range [][2]string{{"fromIdpToken", "custom"}, {"custom", "fromIdpToken"}, {"fromTheImage", "custom"}, {"", "fromIdpToken"}} {
		t.Run("leader="+sources[0]+"/worker="+sources[1], func(t *testing.T) {
			spec := &workloadsApi.SupersetSpec{}
			for i, setter := range []func(workloadsApi.DistributedInferenceLeaderWorkerSpec){spec.SetLeader, spec.SetWorker} {
				var replica workloadsApi.DistributedInferenceLeaderWorkerSpec
				data := `{"security":{"uidGidSource":"` + sources[i] + `","runAsUid":1001,"runAsGid":1002,"supplementalGroups":"1003","hostIpc":true,"capabilities":["IPC_LOCK"]}}`
				if sources[i] == "" {
					data = strings.Replace(data, `"uidGidSource":"",`, "", 1)
				}
				if err := json.Unmarshal([]byte(data), &replica); err != nil {
					t.Fatal(err)
				}
				setter(replica)
			}
			ctx := context.WithValue(context.Background(), runaictx.CfUId, "2001")
			ctx = context.WithValue(ctx, runaictx.CfGId, "2002")
			ctx = context.WithValue(ctx, runaictx.CfSupplementalGroups, "2003")
			if err := setTokenUIDGIDForDistributedInference(ctx, spec); err != nil {
				t.Fatal(err)
			}
			for i, replica := range []*workloadsApi.DistributedInferenceLeaderWorkerSpec{spec.Leader.Get(), spec.Worker.Get()} {
				security := replica.GetSecurity()
				uid, gid, groups := int64(1001), int64(1002), "1003"
				if sources[i] == "fromIdpToken" {
					uid, gid, groups = 2001, 2002, "2003"
				}
				if security.GetRunAsUid() != uid || security.GetRunAsGid() != gid || security.GetSupplementalGroups() != groups {
					t.Fatalf("incorrect identity for source %s: %+v", sources[i], security)
				}
				if !security.GetHostIpc() || len(security.Capabilities) != 1 || security.Capabilities[0] != workloadsApi.CAPABILITY_IPC_LOCK {
					t.Fatalf("unrelated security fields changed: %+v", security)
				}
			}
		})
	}
}

func TestDistributedInferenceAbsentSecurity(t *testing.T) {
	for _, input := range []string{`{}`, `{"leader":null,"worker":null}`, `{"leader":{},"worker":{"security":null}}`} {
		t.Run(input, func(t *testing.T) {
			var spec workloadsApi.SupersetSpec
			if err := json.Unmarshal([]byte(input), &spec); err != nil {
				t.Fatal(err)
			}
			before := identityAsMap(t, spec)
			ctx := context.WithValue(context.Background(), runaictx.CfUId, "2001")
			if err := setTokenUIDGIDForDistributedInference(ctx, &spec); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(identityAsMap(t, spec), before) {
				t.Fatal("absent replica or security section was changed")
			}
		})
	}
}

func readIdentitySpec(t *testing.T, input string) map[string]interface{} {
	t.Helper()
	var value map[string]interface{}
	decoder := json.NewDecoder(strings.NewReader(input))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		t.Fatal(err)
	}
	return value
}

func identityAsMap(t *testing.T, input interface{}) map[string]interface{} {
	t.Helper()
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	return readIdentitySpec(t, string(data))
}
