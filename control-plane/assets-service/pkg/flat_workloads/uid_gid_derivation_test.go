/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package flat_workloads

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/runai-application-packages/convutils"
)

var _ = Describe("uid/gid derivation tests", func() {
	Context("determine uid/gid", func() {
		//   Case   Specific Env               Environment Asset                    Expected uid/gid
		//   ----   ------------------------   ----------------------------------   -------------------
		//   (1)    Not Provided               Does not have default for uid/gid    nil
		//   (2)          "                    Have default for uid/gid             asset default used
		//   (3)    Provided with no uid/gid   Does not have default for uid/gid    nil
		//   (4)          "                    Have default for uid/gid             asset default used
		//   (5)    Provided with uid/gid      Does not have default for uid/gid    specific env used
		//   (6)          "                    Have default for uid/gid             Specific env used
		//
		var defaultUidSpecificEnv int64 = 0
		var defaultGidSpecificEnv int64 = 100
		defaultSupplementalSpecificEnv := "1,3,5,7"

		var defaultUidEnvAsset int64 = 1
		var defaultGidEnvAsset int64 = 300
		defaultSupplementalEnvAsset := "2,4,6,8"

		var specificEnvNotProvided *api.SpecificRunParams = nil
		specificEnvNoUidGid := api.SpecificRunParams{
			RunAsUid:           nil,
			RunAsGid:           nil,
			SupplementalGroups: nil,
		}
		specificEnvWithUidGid := api.SpecificRunParams{
			RunAsUid:           convutils.Ref(defaultUidSpecificEnv),
			RunAsGid:           convutils.Ref(defaultGidSpecificEnv),
			SupplementalGroups: convutils.Ref(defaultSupplementalSpecificEnv),
		}
		specificEnvEmptyUidGid := api.SpecificRunParams{
			RunAsUid:           convutils.Ref(int64(-1)),
			RunAsGid:           convutils.Ref(int64(-1)),
			SupplementalGroups: convutils.Ref(""),
		}

		envAssetNoDefault := api.EnvironmentAssetSpec{
			UidGidSource:       convutils.Ref(CustomUidGid),
			RunAsUid:           nil,
			RunAsGid:           nil,
			SupplementalGroups: nil,
		}

		envAssetWithDefault := api.EnvironmentAssetSpec{
			UidGidSource:       convutils.Ref(CustomUidGid),
			RunAsUid:           convutils.Ref(defaultUidEnvAsset),
			RunAsGid:           convutils.Ref(defaultGidEnvAsset),
			SupplementalGroups: convutils.Ref(defaultSupplementalEnvAsset),
		}

		uid, gid, supplemental, _ := DetermineUidAndGid(specificEnvNotProvided, envAssetNoDefault)
		Expect(uid).To(BeNil()) // specific env not provided, env asset no default uid -> nil
		Expect(gid).To(BeNil()) // specific env not provided, env asset no default gid -> nil
		Expect(supplemental).To(BeNil())

		uid, gid, supplemental, _ = DetermineUidAndGid(specificEnvNotProvided, envAssetWithDefault)
		Expect(convutils.Deref(uid, -1)).To(Equal(defaultUidEnvAsset)) // specific env not provided, env asset has default uid -> use asset default
		Expect(convutils.Deref(gid, -1)).To(Equal(defaultGidEnvAsset)) // specific env not provided, env asset has default gid -> use asset default
		Expect(convutils.Deref(supplemental, "")).To(Equal(defaultSupplementalEnvAsset))

		uid, gid, supplemental, _ = DetermineUidAndGid(&specificEnvNoUidGid, envAssetNoDefault)
		Expect(uid).To(BeNil()) // specific env no uid/gid, env asset no default uid -> nil
		Expect(gid).To(BeNil()) // specific env no uid/gid, env asset no default gid -> nil
		Expect(supplemental).To(BeNil())

		uid, gid, supplemental, _ = DetermineUidAndGid(&specificEnvNoUidGid, envAssetWithDefault)
		Expect(convutils.Deref(uid, -1)).To(Equal(defaultUidEnvAsset)) // specific env no uid/gid, env asset has default uid -> use asset default
		Expect(convutils.Deref(gid, -1)).To(Equal(defaultGidEnvAsset)) // specific env no uig/gid, env asset has default gid -> use asset default
		Expect(convutils.Deref(supplemental, "")).To(Equal(defaultSupplementalEnvAsset))

		uid, gid, supplemental, _ = DetermineUidAndGid(&specificEnvWithUidGid, envAssetNoDefault)
		Expect(convutils.Deref(uid, -1)).To(Equal(defaultUidSpecificEnv)) // specific env with uid/gid, env asset no default uid -> use specific env
		Expect(convutils.Deref(gid, -1)).To(Equal(defaultGidSpecificEnv)) // specific env with uig/gid, env asset no default gid -> use specific env
		Expect(convutils.Deref(supplemental, "")).To(Equal(defaultSupplementalSpecificEnv))

		uid, gid, supplemental, _ = DetermineUidAndGid(&specificEnvWithUidGid, envAssetWithDefault)
		Expect(convutils.Deref(uid, -1)).To(Equal(defaultUidSpecificEnv)) // specific env with uid/gid, env asset has default uid -> use specific env
		Expect(convutils.Deref(gid, -1)).To(Equal(defaultGidSpecificEnv)) // specific env with uig/gid, env asset has default gid -> use specific env
		Expect(convutils.Deref(supplemental, "")).To(Equal(defaultSupplementalSpecificEnv))

		uid, gid, supplemental, _ = DetermineUidAndGid(&specificEnvEmptyUidGid, envAssetWithDefault)
		Expect(uid).To(BeNil()) // specific env with empty uid/gid, env asset has default uid -> revert to image uid
		Expect(gid).To(BeNil()) // specific env with empty uig/gid, env asset has default gid -> revert to image gid
		Expect(supplemental).To(BeNil())
	})
})
