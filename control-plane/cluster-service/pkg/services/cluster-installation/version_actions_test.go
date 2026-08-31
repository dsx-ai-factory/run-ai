/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cluster_installation

import (
	"testing"

	"github.com/Masterminds/semver/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestClusterSupportedVersionsHelper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ClusterSupportedVersions Suite")
}

var _ = Describe("ClusterSupportedVersions", func() {
	minVer, _ := semver.NewVersion("2.7")
	maxVer, _ := semver.NewVersion("2.15")
	It("getLatestMinorVersions - test prerelease max ver", func() {
		versionsList := VersionsList{{"2.18.0"}, {"2.18.0-stage.1"}, {"2.18.0-stage.4"}, {"2.18.0-stage.3"}}
		expectedOutput := map[string]string{"2.18": "2.18.0-stage.4"}
		localMaxVer, _ := semver.NewVersion("2.18.0-stage.4")
		ret := getLatestMinorVersions(versionsList, minVer, localMaxVer)

		Expect(len(ret)).To(Equal(len(expectedOutput)))
		for key, value := range expectedOutput {
			Expect(ret[key]).To(Equal(value))
		}
	})
	It("getLatestMinorVersions - test prerelease 	", func() {
		versionsList := VersionsList{{"2.18.0"}, {"2.18.0-stage.1"}, {"2.18.0-stage.4"}, {"2.18.0-stage.3"}}
		expectedOutput := map[string]string{"2.18": "2.18.0"}
		localMinVer, _ := semver.NewVersion("2.9")
		localMaxVer, _ := semver.NewVersion("2.18")
		ret := getLatestMinorVersions(versionsList, localMinVer, localMaxVer)

		Expect(len(ret)).To(Equal(len(expectedOutput)))
		for key, value := range expectedOutput {
			Expect(ret[key]).To(Equal(value))
		}
	})
	It("getLatestMinorVersions - identical min", func() {
		versionsList := VersionsList{{"2.8.4"}, {"2.7.4"}, {"2.8.0"}, {"2.8.9"}}
		expectedOutput := map[string]string{"2.8": "2.8.9", "2.7": "2.7.4"}
		localMinVer, _ := semver.NewVersion("2.7.4")
		ret := getLatestMinorVersions(versionsList, localMinVer, maxVer)
		Expect(len(ret)).To(Equal(len(expectedOutput)))
		for key, value := range expectedOutput {
			Expect(ret[key]).To(Equal(value))
		}
	})
	It("getLatestMinorVersions - identical max", func() {
		versionsList := VersionsList{{"2.8.4"}, {"2.7.4"}, {"2.8.0"}, {"2.8.9"}}
		expectedOutput := map[string]string{"2.8": "2.8.9", "2.7": "2.7.4"}
		localMaxVer, _ := semver.NewVersion("2.8.9")
		ret := getLatestMinorVersions(versionsList, minVer, localMaxVer)
		Expect(len(ret)).To(Equal(len(expectedOutput)))
		for key, value := range expectedOutput {
			Expect(ret[key]).To(Equal(value))
		}
	})
	It("getLatestMinorVersions - test unordered array", func() {
		versionsList := VersionsList{{"2.8.4"}, {"2.7.4"}, {"2.8.0"}, {"2.7.5"}, {"2.8.9"}}
		expectedOutput := map[string]string{"2.8": "2.8.9", "2.7": "2.7.5"}
		ret := getLatestMinorVersions(versionsList, minVer, maxVer)
		Expect(len(ret)).To(Equal(len(expectedOutput)))
		for key, value := range expectedOutput {
			Expect(ret[key]).To(Equal(value))
		}
	})
	It("getLatestMinorVersions - test duplicate versions", func() {
		versionsList := VersionsList{{"2.7.0"}, {"2.7.0"}, {"2.7.9"}, {"2.7.0"}, {"2.7.3"}, {"2.9.3"}}
		expectedOutput := map[string]string{"2.7": "2.7.9", "2.9": "2.9.3"}
		ret := getLatestMinorVersions(versionsList, minVer, maxVer)
		Expect(len(ret)).To(Equal(len(expectedOutput)))
		for key, value := range expectedOutput {
			Expect(ret[key]).To(Equal(value))
		}
	})
	It("getLatestMinorVersions - test incorrect version syntax (2.7.1.2)", func() {
		input := VersionsList{{"2.7.0"}, {"2.7.1.2"}, {"2.7.2"}, {"2.8"}, {"2.8.2"}}
		expectedOutput := map[string]string{"2.7": "2.7.2", "2.8": "2.8.2"}
		ret := getLatestMinorVersions(input, minVer, maxVer)
		Expect(len(ret)).To(Equal(len(expectedOutput)))
		for key, value := range expectedOutput {
			Expect(ret[key]).To(Equal(value))
		}
	})
	It("getLatestMinorVersions - test incorrect version syntax (2.7)", func() {
		versionsList := VersionsList{{"2.7"}, {"2.7.9"}, {"2.7.0"}, {"2.8"}, {"2.8.3"}}
		expectedOutput := map[string]string{"2.7": "2.7.9", "2.8": "2.8.3"}
		ret := getLatestMinorVersions(versionsList, minVer, maxVer)
		Expect(len(ret)).To(Equal(len(expectedOutput)))
		for key, value := range expectedOutput {
			Expect(ret[key]).To(Equal(value))
		}
	})
	It("getLatestMinorVersions - test patch", func() {
		versionsList := VersionsList{{"2.8.0"}, {"2.8.678"}, {"2.10.0"}}
		expectedOutput := map[string]string{"2.8": "2.8.678", "2.10": "2.10.0"}
		ret := getLatestMinorVersions(versionsList, minVer, maxVer)
		Expect(len(ret)).To(Equal(len(expectedOutput)))
		for key, value := range expectedOutput {
			Expect(ret[key]).To(Equal(value))
		}
	})
	It("getLatestMinorVersions - test all version lower than 2.7", func() {
		versionsList := VersionsList{{"2.1.0"}, {"2.1.678"}}
		expectedOutput := map[string]string{}
		ret := getLatestMinorVersions(versionsList, minVer, maxVer)
		Expect(len(ret)).To(Equal(len(expectedOutput)))
		for key, value := range expectedOutput {
			Expect(ret[key]).To(Equal(value))
		}
	})

	It("getLatestMinorVersions - test versions higher than max supported version", func() {
		versionsList := VersionsList{{"2.16.0"}, {"2.16.678"}}
		expectedOutput := map[string]string{}
		ret := getLatestMinorVersions(versionsList, minVer, maxVer)
		Expect(len(ret)).To(Equal(len(expectedOutput)))
		for key, value := range expectedOutput {
			Expect(ret[key]).To(Equal(value))
		}
	})
	It("getLatestMinorVersions - test versions equal to max supported version", func() {
		versionsList := VersionsList{{"2.999.678"}, {"0.2.1"}}
		expectedOutput := map[string]string{}
		ret := getLatestMinorVersions(versionsList, minVer, maxVer)
		Expect(len(ret)).To(Equal(len(expectedOutput)))
		for key, value := range expectedOutput {
			Expect(ret[key]).To(Equal(value))
		}
	})
	It("getLatestMinorVersions - no versions in range", func() {
		versionsList := VersionsList{{"2.15.678"}, {"2.15.999"}, {"2.15.2"}}
		expectedOutput := map[string]string{"2.15": "2.15.678"}
		localMaxVer, _ := semver.NewVersion("2.15.678")
		ret := getLatestMinorVersions(versionsList, minVer, localMaxVer)
		Expect(len(ret)).To(Equal(len(expectedOutput)))
		for key, value := range expectedOutput {
			Expect(ret[key]).To(Equal(value))
		}
	})
	It("getLatestMinorVersions - max supported version x.y.z", func() {
		versionsList := VersionsList{{"2.15.678"}, {"2.15.999"}, {"2.15.2"}}
		expectedOutput := map[string]string{"2.15": "2.15.678"}
		localMaxVer, _ := semver.NewVersion("2.15.678")
		ret := getLatestMinorVersions(versionsList, minVer, localMaxVer)
		Expect(len(ret)).To(Equal(len(expectedOutput)))
		for key, value := range expectedOutput {
			Expect(ret[key]).To(Equal(value))
		}
	})
	It("getLatestMinorVersions - max supported version x.y.0", func() {
		versionsList := VersionsList{{"2.15.678"}, {"2.15.999"}, {"2.14.2"}}
		expectedOutput := map[string]string{"2.14": "2.14.2"}
		localMaxVer, _ := semver.NewVersion("2.15.0")
		ret := getLatestMinorVersions(versionsList, minVer, localMaxVer)
		Expect(len(ret)).To(Equal(len(expectedOutput)))
		for key, value := range expectedOutput {
			Expect(ret[key]).To(Equal(value))
		}
	})
	It("getLatestMinorVersions - test duplicate versions", func() {
		versionsList := VersionsList{{"2.7.0"}, {"2.7.0"}, {"2.7.9"}, {"2.7.0"}, {"2.7.3"}}
		expectedOutput := map[string]string{"2.7": "2.7.9"}
		ret := getLatestMinorVersions(versionsList, minVer, maxVer)
		Expect(len(ret)).To(Equal(len(expectedOutput)))
		for key, value := range expectedOutput {
			Expect(ret[key]).To(Equal(value))
		}
	})
	It("getLatestMinorVersions - under min ver", func() {
		versionsList := VersionsList{{"2.2.0"}, {"2.2.9"}, {"2.2.1"}, {"2.7.3"}}
		expectedOutput := map[string]string{"2.7": "2.7.3"}
		ret := getLatestMinorVersions(versionsList, minVer, maxVer)
		Expect(len(ret)).To(Equal(len(expectedOutput)))
		for key, value := range expectedOutput {
			Expect(ret[key]).To(Equal(value))
		}
	})
	It("getLatestMinorVersions - exclude version", func() {
		versionsList := VersionsList{{"2.1.2"}, {"2.2.0"}, {"2.2.9"}, {"2.2.1"}, {"2.7.3"}}

		expectedOutput := map[string]string{"2.7": "2.7.3"}
		ret := getLatestMinorVersions(versionsList, minVer, maxVer)
		Expect(len(ret)).To(Equal(len(expectedOutput)))
		for key, value := range expectedOutput {
			Expect(ret[key]).To(Equal(value))
		}
	})
})

var _ = Describe("exclude versions validation", func() {
	It("ExcludeVeriosn - happy flow ", func() {
		versionsMap := map[string]string{"2.7": "2.7.1", "2.8": "2.8.2", "2.9": "2.9.2", "2.10": "2.10.3", "2.11": "2.11.9"}
		exVersions := []string{"2.8", "2.9"}
		expectedOutput := map[string]string{"2.7": "2.7.1", "2.10": "2.10.3", "2.11": "2.11.9"}
		ret := excludeVersions(versionsMap, exVersions)
		Expect(len(ret)).To(Equal(len(expectedOutput)))
		for key, value := range expectedOutput {
			Expect(ret[key]).To(Equal(value))
		}
	})
	It("ExcludeVeriosn - empty exclude list ", func() {
		versionsMap := map[string]string{"2.7": "2.7.1", "2.8": "2.8.2", "2.9": "2.9.2", "2.10": "2.10.3", "2.11": "2.11.9"}
		exVersions := []string{}
		expectedOutput := map[string]string{"2.7": "2.7.1", "2.8": "2.8.2", "2.9": "2.9.2", "2.10": "2.10.3", "2.11": "2.11.9"}
		ret := excludeVersions(versionsMap, exVersions)
		Expect(len(ret)).To(Equal(len(expectedOutput)))
		for key, value := range expectedOutput {
			Expect(ret[key]).To(Equal(value))
		}
	})
	It("ExcludeVeriosn - nil exclude list ", func() {
		versionsMap := map[string]string{"2.7": "2.7.1", "2.8": "2.8.2", "2.9": "2.9.2", "2.10": "2.10.3", "2.11": "2.11.9"}
		var exVersions []string = nil
		expectedOutput := map[string]string{"2.7": "2.7.1", "2.8": "2.8.2", "2.9": "2.9.2", "2.10": "2.10.3", "2.11": "2.11.9"}
		ret := excludeVersions(versionsMap, exVersions)
		Expect(len(ret)).To(Equal(len(expectedOutput)))
		for key, value := range expectedOutput {
			Expect(ret[key]).To(Equal(value))
		}
	})
	It("ExcludeVeriosn - strip exclude list keys ", func() {
		versionsMap := map[string]string{"2.7": "2.7.1", "2.8": "2.8.2", "2.9": "2.9.2", "2.10": "2.10.3", "2.11": "2.11.9"}
		exVersions := []string{"- \"2.8\"\n", "- \"2.9\"\n"}
		expectedOutput := map[string]string{"2.7": "2.7.1", "2.10": "2.10.3", "2.11": "2.11.9"}
		ret := excludeVersions(versionsMap, exVersions)
		Expect(len(ret)).To(Equal(len(expectedOutput)))
		for key, value := range expectedOutput {
			Expect(ret[key]).To(Equal(value))
		}
	})
	It("ExcludeVeriosn - extra excluded ", func() {
		versionsMap := map[string]string{"2.7": "2.7.1", "2.8": "2.8.2", "2.9": "2.9.2", "2.10": "2.10.3", "2.11": "2.11.9"}
		exVersions := []string{"2.8", "2.9", "2.30"}
		expectedOutput := map[string]string{"2.7": "2.7.1", "2.10": "2.10.3", "2.11": "2.11.9"}
		ret := excludeVersions(versionsMap, exVersions)
		Expect(len(ret)).To(Equal(len(expectedOutput)))
		for key, value := range expectedOutput {
			Expect(ret[key]).To(Equal(value))
		}
	})
})

var _ = Describe("version config validation", func() {
	It("version config validation - legal value", func() {
		minVer, maxVer, curVer := "1.2", "1.3", ""
		minVerOut, maxVerOut := adjustMinMaxVersion(nil, minVer, maxVer, curVer)
		Expect(minVerOut.Original()).To(Equal("1.2"))
		Expect(maxVerOut.Original()).To(Equal("1.3"))
	})
	It("version config validation - legal value patch version", func() {
		minVer, maxVer, curVer := "3.0", "3.1.1", "3.1.1"
		minVerOut, maxVerOut := adjustMinMaxVersion(nil, minVer, maxVer, curVer)
		Expect(minVerOut.Original()).To(Equal("3.0"))
		Expect(maxVerOut.Original()).To(Equal("3.1.1"))
	})
	It("version config validation - illegal value", func() {
		minVer, maxVer, curVer := "assd", "1.2", ""
		minVerOut, maxVerOut := adjustMinMaxVersion(nil, minVer, maxVer, curVer)
		Expect(minVerOut).To(Equal(absoluteMinVersion))
		Expect(maxVerOut.Original()).To(Equal("1.2"))
	})
	It("version config validation - illegal range", func() {
		minVer, maxVer, curVer := "2.0", "1.2", ""
		minVerOut, maxVerOut := adjustMinMaxVersion(nil, minVer, maxVer, curVer)
		Expect(minVerOut).To(Equal(absoluteMinVersion))
		Expect(maxVerOut).To(Equal(absoluteMaxVersion))
	})
	It("version config validation - illegal range patch version", func() {
		minVer, maxVer, curVer := "1.2.1", "1.2.0", ""
		minVerOut, maxVerOut := adjustMinMaxVersion(nil, minVer, maxVer, curVer)
		Expect(minVerOut).To(Equal(absoluteMinVersion))
		Expect(maxVerOut).To(Equal(absoluteMaxVersion))
	})
	It("version config validation - empty value - min", func() {
		minVer, maxVer, curVer := "", "1.2", ""
		minVerOut, maxVerOut := adjustMinMaxVersion(nil, minVer, maxVer, curVer)
		Expect(minVerOut).To(Equal(absoluteMinVersion))
		Expect(maxVerOut.Original()).To(Equal("1.2"))
	})
	It("version config validation - empty value - max", func() {
		minVer, maxVer, curVer := "1.2", "", ""
		minVerOut, maxVerOut := adjustMinMaxVersion(nil, minVer, maxVer, curVer)
		Expect(minVerOut.Original()).To(Equal("1.2"))
		Expect(maxVerOut).To(Equal(absoluteMaxVersion))
	})
	It("version config validation - legal value - added string", func() {
		minVer, maxVer, curVer := "1.0-whatever", "1.2-whatever", ""
		minVerOut, maxVerOut := adjustMinMaxVersion(nil, minVer, maxVer, curVer)
		Expect(minVerOut.Original()).To(Equal("1.0-whatever"))
		Expect(maxVerOut.Original()).To(Equal("1.2-whatever"))
	})
	It("version config validation - curVer limit", func() {
		minVer, maxVer, curVer := "1.5", "1.5.9", "1.5.8"
		minVerOut, maxVerOut := adjustMinMaxVersion(nil, minVer, maxVer, curVer)
		Expect(minVerOut.Original()).To(Equal("1.5"))
		Expect(maxVerOut.Original()).To(Equal("1.5.8"))
	})
})
