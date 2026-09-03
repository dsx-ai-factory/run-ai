/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/scopes"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	. "github.com/onsi/gomega" //nolint:staticcheck // dot-import is the idiomatic gomega usage in test helpers
)

var systemScope = scopes.OrgUnit{
	Scope: api.SCOPETYPE_SYSTEM,
}

var SystemPolicy = api.FallbackPolicy{
	/*******************************************************************************
	 * Note!
	 * Any change of policy requires changing the injectSystemPolicyForTesting below
	 * AND of the version in the name
	 ******************************************************************************/
	Meta: api.PolicyMeta{
		Scope: api.SCOPETYPE_SYSTEM,
		Name:  "system-policy-v5",
	},
	Policy: &api.FallbackPolicyDefaultsAndRules{
		Rules: &api.SupersetRules{
			Security: &api.SupersetRulesAllOfSecurity{
				Privileged: &api.BooleanRules{
					SourceOfRule: &api.SourceOfRule{
						Scope: api.SCOPETYPE_SYSTEM,
					},
					CanEdit: convutils.Ref(false),
				},
			},
			TerminationGracePeriodSeconds: &api.IntegerRules{
				SourceOfRule: &api.SourceOfRule{
					Scope: api.SCOPETYPE_SYSTEM,
				},
				Max: convutils.Ref(int32(300)),
			},
		},
		Defaults: &api.SupersetDefaults{
			Security: &api.SupersetDefaultsAllOfSecurity{
				Privileged: convutils.Ref(false),
			},
		},
	},
}

type effectiveTypes interface {
	api.WorkspacePolicyDefaultsAndRulesV2 | api.TrainingPolicyDefaultsAndRulesV2 | api.InferencePolicyDefaultsAndRulesV2 | api.DistributedPolicyDefaultsAndRulesV2
}

var systemPolicyDB dbmodel.PolicyWithOptionalEffective

func HandleSystemPolicy() {
	logging.Infof("handle system policy")

	db := runaidb.DB()

	systemPolicyDB = dbmodel.PolicyWithOptionalEffective{
		Policy: dbmodel.Policy{
			BaseModel: dbmodels.BaseModel{
				CreatedAt: time.Now(),
				CreatedBy: "run:ai",
				UpdatedAt: time.Now(),
				UpdatedBy: "run:ai",
			},
			Name:  SystemPolicy.Meta.Name,
			Scope: SystemPolicy.Meta.Scope,
		},
	}

	defaults, _ := gojsonq.LoadJson(SystemPolicy.Policy.Defaults)
	rules, _ := gojsonq.LoadJson(SystemPolicy.Policy.Rules)

	systemPolicyDB.Policy.DefaultsV2 = defaults
	systemPolicyDB.Policy.RulesV2 = rules

	systemPolicyDB.EffectivePolicy.DefaultsV2 = defaults
	systemPolicyDB.EffectivePolicy.RulesV2 = rules

	newVersion := false

	//
	//   get the system policy from the table in order to check if there is a new version of it
	//
	storedSystemPolicy, err := FindGivenPolicyOfOrgUnit(context.Background(), db, api.SCOPETYPE_SYSTEM, nil, nil, nil,
		FallbackPolicyAsType, false)
	if err != nil {
		logging.Errorf("failed to load system policy from DB: %s: %v", err.Message, err.Details)
		return
	} else if storedSystemPolicy == nil {
		newVersion = true
		logging.Infof("new system policy %s, remove all effective policies from DB", SystemPolicy.Meta.Name)
	} else {
		newVersion = storedSystemPolicy.Name != SystemPolicy.Meta.Name

		systemPolicyDB.ID = storedSystemPolicy.ID
		systemPolicyDB.PolicyId = storedSystemPolicy.ID

		if newVersion {
			logging.Infof("new system policy: %s (was: %s), remove all effective policies from DB", SystemPolicy.Meta.Name, storedSystemPolicy.Name)
		}
	}

	//
	//   if there is a new version, update it and also remove all the effective policies, because they will have
	//   to be re-calculated
	//
	if !newVersion {
		logging.Infof("effective policies are uptodate with the sdystem policies")
		return
	}

	//
	//   delete all the effective policies. system-scope seeding is a
	//   legacy non-GVK path, so we discard the returned descendant ids —
	//   only the GVK flows feed them into the breakdown-cache cleanup.
	//
	_, err = SoftDeleteEffectivePoliciesOfSubtree(context.Background(), db, systemScope, scopes.FallbackPolicyAsType)
	if err != nil {
		logging.Errorf("failed to delete all effective policies: %s: %v", err.Message, err.Details)
	} else {
		logging.Infof("deleted all effective policies")
	}

	//
	//   update the system policy in policy table. we do not use the policy itself from
	//   the DB, we mainly need to keep the name in order to identify a new revision of
	//   the system policy
	//
	err = AddPolicy(context.Background(), db, &systemPolicyDB.Policy)
	if err != nil {
		logging.Errorf("failed to save system policy: %s: %v", err.Message, err.Details)
	}

	systemPolicyDB.PolicyId = systemPolicyDB.ID
}

func InjectSystemPolicyForTesting[T effectiveTypes](effectivePolicy *T) {
	var superEffective api.SupersetPolicy
	copyErr := convutils.CopySpecs(&superEffective, effectivePolicy)
	Expect(copyErr).To(BeNil())

	if superEffective.Rules == nil {
		superEffective.Rules = &api.SupersetRules{}
	}
	if superEffective.Rules.Security == nil {
		superEffective.Rules.Security = &api.SupersetRulesAllOfSecurity{}
	}

	if superEffective.Defaults == nil {
		superEffective.Defaults = &api.SupersetDefaults{}
	}
	if superEffective.Defaults.Security == nil {
		superEffective.Defaults.Security = &api.SupersetDefaultsAllOfSecurity{}
	}

	injectSystemPolicyOfSecurity(superEffective.Rules.Security, superEffective.Defaults.Security)
	superEffective.Rules.TerminationGracePeriodSeconds = SystemPolicy.Policy.Rules.TerminationGracePeriodSeconds

	copyErr = convutils.CopySpecs(&effectivePolicy, superEffective)
	Expect(copyErr).To(BeNil())
}

func InjectSystemPolicyForDistributedTesting(effective *api.DistributedPolicyDefaultsAndRulesV2) {
	if effective.Rules == nil {
		effective.Rules = &api.DistributedPolicyRulesV2{}
	}
	if effective.Rules.Worker == nil {
		effective.Rules.Worker = &api.DistributedPolicyRulesV2Worker{}
	}
	if effective.Rules.Master == nil {
		effective.Rules.Master = &api.ReplicaRulesV2{}
	}

	if effective.Defaults == nil {
		effective.Defaults = &api.DistributedPolicyDefaultsV2{}
	}
	if effective.Defaults.Worker == nil {
		effective.Defaults.Worker = &api.DistributedPolicyDefaultsV2Worker{}
	}
	if effective.Defaults.Master == nil {
		effective.Defaults.Master = &api.ReplicaDefaultsV2{}
	}

	injectSystemPolicyOfSecurity(effective.Rules.Master.Security, effective.Defaults.Master.Security)
	injectSystemPolicyOfSecurity(effective.Rules.Worker.Security, effective.Defaults.Worker.Security)

	effective.Rules.Master.TerminationGracePeriodSeconds = SystemPolicy.Policy.Rules.TerminationGracePeriodSeconds
	effective.Rules.Worker.TerminationGracePeriodSeconds = SystemPolicy.Policy.Rules.TerminationGracePeriodSeconds
}

func InjectSystemPolicyForDistributedInferenceTesting(effective *api.DistributedInferencePolicyDefaultsAndRulesV2) {
	if effective.Rules == nil {
		effective.Rules = &api.DistributedInferenceRulesV2{}
	}
	if effective.Rules.Leader == nil {
		effective.Rules.Leader = &api.DistributedInferenceLeaderWorkerRulesV2{}
	}
	if effective.Rules.Worker == nil {
		effective.Rules.Worker = &api.DistributedInferenceLeaderWorkerRulesV2{}
	}

	if effective.Defaults == nil {
		effective.Defaults = &api.DistributedInferenceDefaultsV2{}
	}
	if effective.Defaults.Worker == nil {
		effective.Defaults.Worker = &api.DistributedInferenceLeaderWorkerDefaultsV2{}
	}
	if effective.Defaults.Leader == nil {
		effective.Defaults.Leader = &api.DistributedInferenceLeaderWorkerDefaultsV2{}
	}

	injectSystemPolicyOfSecurityForDistributedInference(effective.Rules.Leader.Security, effective.Defaults.Leader.Security)
	injectSystemPolicyOfSecurityForDistributedInference(effective.Rules.Worker.Security, effective.Defaults.Worker.Security)

}

func injectSystemPolicyOfSecurity(rules *api.SupersetRulesAllOfSecurity, defaults *api.SupersetDefaultsAllOfSecurity) {
	rules.Privileged = SystemPolicy.Policy.Rules.Security.Privileged
	defaults.Privileged = SystemPolicy.Policy.Defaults.Security.Privileged
}

func injectSystemPolicyOfSecurityForDistributedInference(rules *api.DistributedInferenceLeaderWorkerRulesV2Security, defaults *api.DistributedInferenceLeaderWorkerDefaultsV2Security) {
	rules.Privileged = SystemPolicy.Policy.Rules.Security.Privileged
	defaults.Privileged = SystemPolicy.Policy.Defaults.Security.Privileged
}
