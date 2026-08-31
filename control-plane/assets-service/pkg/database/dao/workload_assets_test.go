/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"sort"
	"testing"

	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/backend_db"
	"github.com/run-ai/runai/runai-common-packages/integdb"
	"github.com/run-ai/runai/runai-common-packages/runaidb"

	"github.com/run-ai/runai/runai-application-packages/convutils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/config"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
)

func TestServices(t *testing.T) {
	config.MockConfig()
	config.Init()
	RegisterFailHandler(Fail)

	conf, cleanup, setupErr := integdb.SetupPostgresDB(&integdb.IntegdbConfig{Schema: "assets", MigrationDirPath: "file://../migrations"})
	if setupErr != nil {
		t.Fatalf("failed to set up test postgres: %v", setupErr)
	}
	t.Cleanup(cleanup)
	conf.UseServiceSession = new(bool)
	if initErr := runaidb.Init(conf); initErr != nil {
		t.Fatalf("failed to init database: %v", initErr)
	}
	t.Cleanup(func() { _ = runaidb.Close() })

	backend_db.CreateTables(runaidb.DB().Actual())

	RunSpecs(t, "dao suite")
}

var _ = Describe("workload assets validation", func() {
	Context("[]assetMetaRec -> AssetsIds", func() {
		It("should populate the assets properly", func() {
			var result api.AssetsRef
			source := []dbmodel.AssetMetaRec{
				{AssetMeta: dbmodel.AssetMeta{Kind: api.ASSETKIND_ENVIRONMENT, Id: "env-1"}},
				{AssetMeta: dbmodel.AssetMeta{Kind: api.ASSETKIND_COMPUTE, Id: "comp-1"}},
				{AssetMeta: dbmodel.AssetMeta{Kind: api.ASSETKIND_HOST_PATH, Id: "host-1"}},
				{AssetMeta: dbmodel.AssetMeta{Kind: api.ASSETKIND_HOST_PATH, Id: "host-2"}},
				{AssetMeta: dbmodel.AssetMeta{Kind: api.ASSETKIND_NFS, Id: "nfs-1"}},
				{AssetMeta: dbmodel.AssetMeta{Kind: api.ASSETKIND_NFS, Id: "nfs-2"}},
				{AssetMeta: dbmodel.AssetMeta{Kind: api.ASSETKIND_GIT, Id: "git-1"}},
				{AssetMeta: dbmodel.AssetMeta{Kind: api.ASSETKIND_GIT, Id: "git-2"}},
				{AssetMeta: dbmodel.AssetMeta{Kind: api.ASSETKIND_S3, Id: "s3-1"}},
				{AssetMeta: dbmodel.AssetMeta{Kind: api.ASSETKIND_S3, Id: "s3-2"}},
				{AssetMeta: dbmodel.AssetMeta{Kind: api.ASSETKIND_PVC, Id: "pvc-1"}},
				{AssetMeta: dbmodel.AssetMeta{Kind: api.ASSETKIND_PVC, Id: "pvc-2"}},
			}
			err := assetRecToAssetRef(&result, source, nil)
			Expect(err).To(BeNil())

			expected := api.AssetsRef{
				Environment: api.EnvironmentAssetRef{Id: "env-1"},
				Compute:     &api.AssetsRefCompute{Id: "comp-1"},
				Datasources: []api.DatasourceRef{
					{Kind: api.ASSETKIND_HOST_PATH, Id: "host-1"},
					{Kind: api.ASSETKIND_HOST_PATH, Id: "host-2"},
					{Kind: api.ASSETKIND_NFS, Id: "nfs-1"},
					{Kind: api.ASSETKIND_NFS, Id: "nfs-2"},
					{Kind: api.ASSETKIND_GIT, Id: "git-1"},
					{Kind: api.ASSETKIND_GIT, Id: "git-2"},
					{Kind: api.ASSETKIND_S3, Id: "s3-1"},
					{Kind: api.ASSETKIND_S3, Id: "s3-2"},
					{Kind: api.ASSETKIND_PVC, Id: "pvc-1"},
					{Kind: api.ASSETKIND_PVC, Id: "pvc-2"},
				},
			}
			sort.Slice(expected.Datasources, func(i, j int) bool {
				return expected.Datasources[i].Id < expected.Datasources[j].Id
			})

			Expect(result).To(Equal(expected))
		})
	})

	Context("[]containmentRecords -> []AssetRef", func() {
		It("should populate the assets properly", func() {
			toolTypesEnv1 := []api.ToolType{
				"tool-1",
				"tool-2",
			}
			toolTypes := EnvironmentToolsMap{
				"env-1": toolTypesEnv1,
			}
			source := []dbmodel.TemplateAssetsCombinedRec{
				{TemplateAssetId: "1", AssetMeta: dbmodel.AssetMetaRec{AssetMeta: dbmodel.AssetMeta{Kind: api.ASSETKIND_ENVIRONMENT, Id: "env-1"}}},
				{TemplateAssetId: "1", AssetMeta: dbmodel.AssetMetaRec{AssetMeta: dbmodel.AssetMeta{Kind: api.ASSETKIND_COMPUTE, Id: "comp-1"}}},
				{TemplateAssetId: "1", AssetMeta: dbmodel.AssetMetaRec{AssetMeta: dbmodel.AssetMeta{Kind: api.ASSETKIND_HOST_PATH, Id: "host-1"}}},
				{TemplateAssetId: "1", AssetMeta: dbmodel.AssetMetaRec{AssetMeta: dbmodel.AssetMeta{Kind: api.ASSETKIND_NFS, Id: "nfs-1"}}},
				{TemplateAssetId: "1", AssetMeta: dbmodel.AssetMetaRec{AssetMeta: dbmodel.AssetMeta{Kind: api.ASSETKIND_NFS, Id: "nfs-3"}}},
				{TemplateAssetId: "1", AssetMeta: dbmodel.AssetMetaRec{AssetMeta: dbmodel.AssetMeta{Kind: api.ASSETKIND_GIT, Id: "git-1"}}},
				{TemplateAssetId: "1", AssetMeta: dbmodel.AssetMetaRec{AssetMeta: dbmodel.AssetMeta{Kind: api.ASSETKIND_S3, Id: "s3-1"}}},
				{TemplateAssetId: "1", AssetMeta: dbmodel.AssetMetaRec{AssetMeta: dbmodel.AssetMeta{Kind: api.ASSETKIND_S3, Id: "s3-3"}}},
				{TemplateAssetId: "1", AssetMeta: dbmodel.AssetMetaRec{AssetMeta: dbmodel.AssetMeta{Kind: api.ASSETKIND_PVC, Id: "pvc-1"}}},

				{TemplateAssetId: "2", AssetMeta: dbmodel.AssetMetaRec{AssetMeta: dbmodel.AssetMeta{Kind: api.ASSETKIND_ENVIRONMENT, Id: "env-2"}}},
				{TemplateAssetId: "2", AssetMeta: dbmodel.AssetMetaRec{AssetMeta: dbmodel.AssetMeta{Kind: api.ASSETKIND_HOST_PATH, Id: "host-2"}}},
				{TemplateAssetId: "2", AssetMeta: dbmodel.AssetMetaRec{AssetMeta: dbmodel.AssetMeta{Kind: api.ASSETKIND_HOST_PATH, Id: "host-3"}}},
				{TemplateAssetId: "2", AssetMeta: dbmodel.AssetMetaRec{AssetMeta: dbmodel.AssetMeta{Kind: api.ASSETKIND_NFS, Id: "nfs-2"}}},
				{TemplateAssetId: "2", AssetMeta: dbmodel.AssetMetaRec{AssetMeta: dbmodel.AssetMeta{Kind: api.ASSETKIND_GIT, Id: "git-2"}}},
				{TemplateAssetId: "2", AssetMeta: dbmodel.AssetMetaRec{AssetMeta: dbmodel.AssetMeta{Kind: api.ASSETKIND_GIT, Id: "git-3"}}},
				{TemplateAssetId: "2", AssetMeta: dbmodel.AssetMetaRec{AssetMeta: dbmodel.AssetMeta{Kind: api.ASSETKIND_S3, Id: "s3-2"}}},
				{TemplateAssetId: "2", AssetMeta: dbmodel.AssetMetaRec{AssetMeta: dbmodel.AssetMeta{Kind: api.ASSETKIND_PVC, Id: "pvc-2"}}},
				{TemplateAssetId: "2", AssetMeta: dbmodel.AssetMetaRec{AssetMeta: dbmodel.AssetMeta{Kind: api.ASSETKIND_PVC, Id: "pvc-3"}}},
			}

			for idx := range source {
				source[idx].AssetMeta.Name = source[idx].AssetMeta.Id + "-name"
			}
			result, err := templateAssetsRecToMap(source, toolTypes)
			Expect(err).To(BeNil())

			expected := TemplateToAssetsMap{
				"1": api.AssetsRef{
					Environment: api.EnvironmentAssetRef{Id: "env-1", Name: "env-1-name", ToolTypes: &toolTypesEnv1},
					Compute:     &api.AssetsRefCompute{Id: "comp-1", Name: "comp-1-name"},
					Datasources: []api.DatasourceRef{
						{Kind: api.ASSETKIND_HOST_PATH, Id: "host-1", Name: "host-1-name"},
						{Kind: api.ASSETKIND_NFS, Id: "nfs-1", Name: "nfs-1-name"},
						{Kind: api.ASSETKIND_NFS, Id: "nfs-3", Name: "nfs-3-name"},
						{Kind: api.ASSETKIND_GIT, Id: "git-1", Name: "git-1-name"},
						{Kind: api.ASSETKIND_S3, Id: "s3-1", Name: "s3-1-name"},
						{Kind: api.ASSETKIND_S3, Id: "s3-3", Name: "s3-3-name"},
						{Kind: api.ASSETKIND_PVC, Id: "pvc-1", Name: "pvc-1-name"},
					},
				},
				"2": api.AssetsRef{
					Environment: api.EnvironmentAssetRef{Id: "env-2", Name: "env-2-name"},
					Datasources: []api.DatasourceRef{
						{Kind: api.ASSETKIND_HOST_PATH, Id: "host-2", Name: "host-2-name"},
						{Kind: api.ASSETKIND_HOST_PATH, Id: "host-3", Name: "host-3-name"},
						{Kind: api.ASSETKIND_NFS, Id: "nfs-2", Name: "nfs-2-name"},
						{Kind: api.ASSETKIND_GIT, Id: "git-2", Name: "git-2-name"},
						{Kind: api.ASSETKIND_GIT, Id: "git-3", Name: "git-3-name"},
						{Kind: api.ASSETKIND_S3, Id: "s3-2", Name: "s3-2-name"},
						{Kind: api.ASSETKIND_PVC, Id: "pvc-2", Name: "pvc-2-name"},
						{Kind: api.ASSETKIND_PVC, Id: "pvc-3", Name: "pvc-3-name"},
					},
				},
			}
			Expect(result).To(Equal(expected))
		})
	})

	Context("validate contained assets", func() {
		It("should populate the assets properly", func() {
			ids := api.AssetsIds{
				Environment: "env-1",
				Compute:     convutils.Ref("comp-1"),
				Datasources: []api.AssetIdAndKind{
					{Kind: api.ASSETKIND_HOST_PATH, Id: "host-1"},
					{Kind: api.ASSETKIND_HOST_PATH, Id: "host-2"},
					{Kind: api.ASSETKIND_PVC, Id: "pvc-1"},
					{Kind: api.ASSETKIND_PVC, Id: "pvc-2"},
					{Kind: api.ASSETKIND_NFS, Id: "nfs-1"},
					{Kind: api.ASSETKIND_NFS, Id: "nfs-2"},
					{Kind: api.ASSETKIND_GIT, Id: "git-1"},
					{Kind: api.ASSETKIND_GIT, Id: "git-2"},
					{Kind: api.ASSETKIND_S3, Id: "s3-1"},
					{Kind: api.ASSETKIND_S3, Id: "s3-2"},
				},
			}
			assets, kindToIdMap, workloadVolume := assetIdsContainedInWorkload(&ids)
			Expect(len(workloadVolume)).To(Equal(0))

			expectedAssets := []string{
				"env-1",
				"comp-1",
				"host-1",
				"host-2",
				"nfs-1",
				"nfs-2",
				"pvc-1",
				"pvc-2",
				"git-1",
				"git-2",
				"s3-1",
				"s3-2",
			}
			sort.Strings(assets)
			sort.Strings(expectedAssets)
			Expect(assets).To(Equal(expectedAssets))

			expectedKindMap := dbmodel.AssetKindToIdsMap{
				api.ASSETKIND_ENVIRONMENT: []string{"env-1"},
				api.ASSETKIND_COMPUTE:     []string{"comp-1"},
				api.ASSETKIND_HOST_PATH:   []string{"host-1", "host-2"},
				api.ASSETKIND_NFS:         []string{"nfs-1", "nfs-2"},
				api.ASSETKIND_PVC:         []string{"pvc-1", "pvc-2"},
				api.ASSETKIND_S3:          []string{"s3-1", "s3-2"},
				api.ASSETKIND_GIT:         []string{"git-1", "git-2"},
			}
			Expect(kindToIdMap).To(Equal(expectedKindMap))
		})
	})

	Context("validate append asset/assets", func() {
		It("should append correctly", func() {
			ids := make([]string, 0)
			kindToIds := make(dbmodel.AssetKindToIdsMap)

			ids = appendContainedAsset(ids, &kindToIds, convutils.Ref("1"), "s3")
			ids = appendContainedAsset(ids, &kindToIds, nil, "s3")
			ids = appendContainedAsset(ids, &kindToIds, convutils.Ref("2"), "s3")
			ids = appendContainedAssets(ids, &kindToIds, []string{"3", "4"}, "s3")
			ids = appendContainedAssets(ids, &kindToIds, []string{}, "s3")
			ids = appendContainedAssets(ids, &kindToIds, []string{"5"}, "s3")

			expectedIds := []string{"1", "2", "3", "4", "5"}
			Expect(ids).To(Equal(expectedIds))
			Expect(kindToIds).To(Equal(dbmodel.AssetKindToIdsMap{"s3": expectedIds}))
		})
	})

	Context("validate unsupported assets", func() {
		It("should fail", func() {
			wrongKind := dbmodel.AssetMeta{Kind: "xx"}
			wrongMetaRec := dbmodel.AssetMetaRec{AssetMeta: wrongKind}

			assets := []dbmodel.AssetMetaRec{
				{
					AssetMeta: dbmodel.AssetMeta{
						Kind: wrongKind.Kind,
					},
				},
			}
			var assetsRef api.AssetsRef
			err := assetRecToAssetRef(&assetsRef, assets, nil)
			Expect(err).NotTo(BeNil())

			_, err = templateAssetsRecToMap([]dbmodel.TemplateAssetsCombinedRec{{AssetMeta: wrongMetaRec}}, make(EnvironmentToolsMap))
			Expect(err).NotTo(BeNil())
		})
	})
})
