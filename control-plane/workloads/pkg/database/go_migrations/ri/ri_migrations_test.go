/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package ri_migrations

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/go_migrations/ri/source"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/go_migrations/ri/source/kubernetes"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	testingutils "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/integdb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/utils/ptr"
)

func TestRIMigrations(t *testing.T) {
	RegisterFailHandler(Fail)
	config.Init()
	conf, cleanup, err := integdb.SetupPostgresDB(&integdb.IntegdbConfig{
		Schema:           "workloads",
		MigrationDirPath: "file://../../../database/migrations",
	})
	if err != nil {
		t.Fatalf("failed to set up test postgres: %v", err)
	}
	t.Cleanup(cleanup)
	conf.UseServiceSession = new(bool)
	conf.SkipDefaultTransaction = true
	if err := runaidb.Init(conf); err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	t.Cleanup(func() { _ = runaidb.Close() })

	fmt.Printf("RI Migrations DB Port: %d\n", conf.Connection.Port)

	RunSpecs(t, "RI migrations tests")

	t.Cleanup(func() {
		testingutils.CleanTables()
	})
}

var _ = Describe("RI Migrations testing", func() {
	Context("Internal RI YAML parsing", func() {
		It("parses KServe InferenceService v1beta1 RI and validates key fields", func() {
			cfgs := getKartasToMigrate()
			all := make([]models.Karta, 0)
			for _, cfg := range cfgs {
				all = append(all, cfg.Kartas...)
			}

			var riForKServe *models.Karta
			for i := range all {
				gvk := all[i].GroupVersionKind()
				if gvk != nil && gvk.Group == "serving.kserve.io" && gvk.Kind == "InferenceService" && gvk.Version == "v1beta1" {
					riForKServe = &all[i]
					break
				}
			}
			Expect(riForKServe).ToNot(BeNil(), "expected to find KServe InferenceService v1beta1 in embedded RIs")

			karta := *riForKServe

			// root component name and fragmented pod spec paths
			root := karta.Spec.StructureDefinition.RootComponent
			Expect(root.Name).To(Equal("inferenceservice"))
			Expect(root.SpecDefinition).To(BeNil())

			// status mapping presence
			Expect(root.StatusDefinition).ToNot(BeNil())
			status := root.StatusDefinition
			Expect(status.ConditionsDefinition).ToNot(BeNil())
			Expect(status.ConditionsDefinition.Path).To(Equal(".status.conditions"))
			Expect(status.ConditionsDefinition.TypeFieldName).To(Equal("type"))
			Expect(status.ConditionsDefinition.StatusFieldName).To(Equal("status"))
			Expect(status.StatusMappings).ToNot(BeNil())
			Expect(len(status.StatusMappings.Running)).To(BeNumerically(">", 0))
			Expect(len(status.StatusMappings.Failed)).To(BeNumerically(">", 0))

			// child components
			children := karta.Spec.StructureDefinition.ChildComponents
			Expect(children).ToNot(BeNil())
			Expect(len(children)).To(Equal(2))
			// predictor and transformer (order-safe)
			var predictor, transformer kartav1alpha1.ComponentDefinition
			if children[0].Name == "predictor" {
				predictor = children[0]
				transformer = children[1]
			} else {
				predictor = children[1]
				transformer = children[0]
			}
			Expect(predictor.Name).To(Equal("predictor"))
			Expect(predictor.Kind).ToNot(BeNil())
			Expect(predictor.Kind.Group).To(Equal("apps"))
			Expect(predictor.Kind.Kind).To(Equal("Deployment"))
			Expect(predictor.Kind.Version).To(Equal("v1"))
			Expect(predictor.SpecDefinition).ToNot(BeNil())
			Expect(predictor.ScaleDefinition).ToNot(BeNil())
			Expect(*predictor.ScaleDefinition.MinReplicasPath).To(Equal(".spec.predictor.minReplicas"))
			Expect(*predictor.ScaleDefinition.MaxReplicasPath).To(Equal(".spec.predictor.maxReplicas"))
			Expect(predictor.PodSelector).ToNot(BeNil())
			Expect(predictor.PodSelector.ComponentTypeSelector).ToNot(BeNil())
			Expect(predictor.PodSelector.ComponentTypeSelector.KeyPath).To(Equal(".metadata.labels[\"component\"]"))
			Expect(*predictor.PodSelector.ComponentTypeSelector.Value).To(Equal("predictor"))

			fpsd := predictor.SpecDefinition.FragmentedPodSpecDefinition
			Expect(fpsd).ToNot(BeNil())
			Expect(*fpsd.SchedulerNamePath).To(Equal(".spec.predictor.schedulerName"))
			Expect(*fpsd.LabelsPath).To(Equal(".spec.predictor.labels"))
			Expect(*fpsd.AnnotationsPath).To(Equal(".spec.predictor.annotations"))
			Expect(*fpsd.PodAffinityPath).To(Equal(".spec.predictor.affinity.podAffinity"))
			Expect(*fpsd.NodeAffinityPath).To(Equal(".spec.predictor.affinity.nodeAffinity"))
			Expect(*fpsd.ContainerPath).To(Equal(".spec.predictor | ( (.[]?  | select(type ==\"object\" and .storageUri )))"))

			// transformer
			Expect(transformer.Name).To(Equal("transformer"))
			Expect(transformer.Kind).ToNot(BeNil())
			Expect(transformer.Kind.Group).To(Equal("apps"))
			Expect(transformer.Kind.Kind).To(Equal("Deployment"))
			Expect(transformer.Kind.Version).To(Equal("v1"))
			Expect(transformer.SpecDefinition).ToNot(BeNil())
			Expect(*transformer.SpecDefinition.PodSpecPath).To(Equal(".spec.transformer"))
			Expect(*transformer.SpecDefinition.MetadataPath).To(Equal(".spec.transformer"))
			Expect(transformer.ScaleDefinition).ToNot(BeNil())
			Expect(*transformer.ScaleDefinition.MinReplicasPath).To(Equal(".spec.transformer.minReplicas"))
			Expect(*transformer.ScaleDefinition.MaxReplicasPath).To(Equal(".spec.transformer.maxReplicas"))
			Expect(transformer.PodSelector).ToNot(BeNil())
			Expect(transformer.PodSelector.ComponentTypeSelector).ToNot(BeNil())
			Expect(transformer.PodSelector.ComponentTypeSelector.KeyPath).To(Equal(".metadata.labels[\"component\"]"))
			Expect(*transformer.PodSelector.ComponentTypeSelector.Value).To(Equal("transformer"))
		})
	})

	Context("validateGvks", func() {
		It("returns error for empty list", func() {
			_, err := validateGvks([]models.Karta{})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("no RIs"))
		})

		It("updates external workload type config (priority/category) on re-upsert", func() {
			db := runaidb.DB()
			group := fmt.Sprintf("test.example.%s", randSuffix3())
			kind := fmt.Sprintf("DummyType%s", randSuffix3())
			workloadType, apiErr := dao.DbUpsertWorkloadType(context.Background(), db, workloadTypeCreateFields(kind, group))
			Expect(apiErr).To(BeNil())
			id := workloadType.ID.String()

			// initial insert
			err := upsertExternalWorkloadsTypeConfig(db, id, "Deploy", "very-high", string(openapi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE))
			Expect(err).To(BeNil())

			// validate initial
			cfgView, getErr := dao.DbGetWorkloadTypeConfigById(context.Background(), 0, id)
			Expect(getErr).To(BeNil())
			Expect(cfgView.CategoryName).To(Equal("Deploy"))
			Expect(cfgView.PriorityName).To(Equal("very-high"))
			Expect(cfgView.Preemptibility).To(Equal(string(openapi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE)))
			// created/updated metadata
			Expect(cfgView.CreatedAt.IsZero()).To(BeFalse())
			Expect(cfgView.UpdatedAt.IsZero()).To(BeFalse())
			Expect(cfgView.CreatedBy).ToNot(BeNil())
			Expect(*cfgView.CreatedBy).To(Equal("Run.ai"))
			Expect(cfgView.UpdatedBy).ToNot(BeNil())
			Expect(*cfgView.UpdatedBy).To(Equal("Run.ai"))

			// change priority (and keep same category)
			err = upsertExternalWorkloadsTypeConfig(db, id, "Deploy", "high", string(openapi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE))
			Expect(err).To(BeNil())

			// validate updated
			cfgView2, getErr2 := dao.DbGetWorkloadTypeConfigById(context.Background(), 0, id)
			Expect(getErr2).To(BeNil())
			Expect(cfgView2.CategoryName).To(Equal("Deploy"))
			Expect(cfgView2.PriorityName).To(Equal("high"))
			Expect(cfgView2.Preemptibility).To(Equal(string(openapi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE)))
			// created should remain, updated should advance
			Expect(cfgView2.CreatedAt.Equal(cfgView.CreatedAt)).To(BeTrue())
			Expect(cfgView2.UpdatedAt.After(cfgView.UpdatedAt)).To(BeTrue())
			Expect(cfgView2.UpdatedBy).ToNot(BeNil())
			Expect(*cfgView2.UpdatedBy).To(Equal("Run.ai"))
		})

		It("updates resource interface content on re-upsert for same version", func() {
			db := runaidb.DB()
			group := fmt.Sprintf("test.example.%s", randSuffix3())
			kind := fmt.Sprintf("DummyType%s", randSuffix3())
			version := fmt.Sprintf("v1alpha1-%s", randSuffix3())
			workloadType, apiErr := dao.DbUpsertWorkloadType(context.Background(), db, workloadTypeCreateFields(kind, group))
			Expect(apiErr).To(BeNil())
			id := workloadType.ID.String()

			karta := buildMinimalRI(group, version, kind)
			// set an initial value
			initialPath := ".spec.initial"
			karta.Spec.StructureDefinition.RootComponent.SpecDefinition.PodTemplateSpecPath = ptr.To(initialPath)

			// first upsert
			err := upsertKarta(db, karta, id)
			Expect(err).To(BeNil())
			// validate stored value is initial and timestamps/created_by
			var (
				b1         []byte
				createdAt1 time.Time
				updatedAt1 time.Time
				createdBy1 string
			)
			q1 := db.Raw(`SELECT karta, created_at, updated_at, created_by FROM workload_type_kartas WHERE workload_type_id = ? AND version = ?`, id, version).Row()
			err = q1.Scan(&b1, &createdAt1, &updatedAt1, &createdBy1)
			Expect(err).To(BeNil())
			var got1 models.Karta
			err = json.Unmarshal(b1, &got1)
			Expect(err).To(BeNil())
			Expect(*got1.Spec.StructureDefinition.RootComponent.SpecDefinition.PodTemplateSpecPath).To(Equal(initialPath))
			Expect(createdBy1).To(Equal("Run.ai"))
			Expect(updatedAt1.IsZero()).To(BeFalse())
			Expect(createdAt1.IsZero()).To(BeFalse())

			// change the path and upsert again (should perform update)
			changedPath := ".spec.changed"
			karta.Spec.StructureDefinition.RootComponent.SpecDefinition.PodTemplateSpecPath = ptr.To(changedPath)
			err = upsertKarta(db, karta, id)
			Expect(err).To(BeNil())

			// validate stored value is changed
			var (
				b2         []byte
				createdAt2 time.Time
				updatedAt2 time.Time
				createdBy2 string
			)
			q2 := db.Raw(`SELECT karta, created_at, updated_at, created_by FROM workload_type_kartas WHERE workload_type_id = ? AND version = ?`, id, version).Row()
			err = q2.Scan(&b2, &createdAt2, &updatedAt2, &createdBy2)
			Expect(err).To(BeNil())
			var got2 models.Karta
			err = json.Unmarshal(b2, &got2)
			Expect(err).To(BeNil())
			Expect(*got2.Spec.StructureDefinition.RootComponent.SpecDefinition.PodTemplateSpecPath).To(Equal(changedPath))
			Expect(createdBy2).To(Equal("Run.ai"))
			// created_at should remain the same; updated_at should advance
			Expect(createdAt2.Equal(createdAt1)).To(BeTrue())
			Expect(updatedAt2.After(updatedAt1)).To(BeTrue())
		})

		It("returns error when first RI has invalid GVK", func() {
			invalid := models.Karta{} // no Kind set
			_, err := validateGvks([]models.Karta{invalid})
			Expect(err).ToNot(BeNil())
		})

		It("returns error on mismatched group/kind", func() {
			a := source.BuildKServeInferenceServiceV1beta1()
			b := source.BuildKServeInferenceServiceV1beta1()
			// change kind
			b.Spec.StructureDefinition.RootComponent.Kind.Kind = "OtherKind"
			_, err := validateGvks([]models.Karta{a, b})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("mismatched group/kind"))
		})

		It("returns error on duplicate version", func() {
			a := source.BuildKServeInferenceServiceV1beta1()
			b := source.BuildKServeInferenceServiceV1beta1()
			_, err := validateGvks([]models.Karta{a, b})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("duplicate RI version"))
		})

		It("returns success on valid dynamo graph deployment karta", func() {
			dgd := source.BuildDynamoGraphDeploymentV1beta1()
			gk, err := validateGvks([]models.Karta{dgd})
			Expect(err).To(BeNil())
			Expect(gk.Group).To(Equal("nvidia.com"))
			Expect(gk.Kind).To(Equal("DynamoGraphDeployment"))
		})

		It("returns success on valid dynamo graph deployment karta", func() {
			dgd := source.BuildDynamoGraphDeploymentV1beta1()

			converted := kartav1alpha1.Karta(dgd)
			validator := kartav1alpha1.NewKartaValidator(&converted)
			err := validator.Validate()
			Expect(err).To(BeNil())
		})

		It("returns GroupKind when valid", func() {
			a := source.BuildKServeInferenceServiceV1beta1()
			// a single RI is valid
			gk, err := validateGvks([]models.Karta{a})
			Expect(err).To(BeNil())
			Expect(gk.Group).To(Equal("serving.kserve.io"))
			Expect(gk.Kind).To(Equal("InferenceService"))
		})

		It("validate LeaderWorkerSet RI", func() {
			converted := kartav1alpha1.Karta(source.BuildLeaderWorkerSetV1())
			validator := kartav1alpha1.NewKartaValidator(&converted)
			err := validator.Validate()
			Expect(err).To(BeNil())
		})
	})

	Context("upserts and end-to-end migration", Ordered, func() {
		It("upserts workload type and returns stable ID across calls", func() {
			db := runaidb.DB()
			group := fmt.Sprintf("test.example.%s", randSuffix3())
			kind := fmt.Sprintf("DummyType%s", randSuffix3())
			workloadType1, apiErr := dao.DbUpsertWorkloadType(context.Background(), db, workloadTypeCreateFields(kind, group))
			Expect(apiErr).To(BeNil())
			Expect(workloadType1.ID).ToNot(BeEmpty())
			workloadType2, apiErr := dao.DbUpsertWorkloadType(context.Background(), db, workloadTypeCreateFields(kind, group))
			Expect(apiErr).To(BeNil())
			Expect(workloadType2.ID.String()).To(Equal(workloadType1.ID.String()))
		})

		It("upserts external config once (idempotent)", func() {
			db := runaidb.DB()
			group := fmt.Sprintf("test.example.%s", randSuffix3())
			kind := fmt.Sprintf("DummyType%s", randSuffix3())
			workloadType, apiErr := dao.DbUpsertWorkloadType(context.Background(), db, workloadTypeCreateFields(kind, group))
			Expect(apiErr).To(BeNil())
			id := workloadType.ID.String()
			// first insert
			err := upsertExternalWorkloadsTypeConfig(db, id, "Deploy", "very-high", string(openapi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE))
			Expect(err).To(BeNil())
			// second insert should no-op
			err = upsertExternalWorkloadsTypeConfig(db, id, "Deploy", "very-high", string(openapi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE))
			Expect(err).To(BeNil())

			type Row struct{ Count int64 }
			row := Row{}
			err = db.Raw(`SELECT COUNT(1) AS count FROM external_workloads_type_config WHERE workload_type_id = ?`, id).Row().Scan(&row.Count)
			Expect(err).To(BeNil())
			Expect(row.Count).To(Equal(int64(1)))
		})

		It("upserts resource interface once per version", func() {
			db := runaidb.DB()
			group := fmt.Sprintf("test.example.%s", randSuffix3())
			kind := fmt.Sprintf("DummyType%s", randSuffix3())
			version := fmt.Sprintf("v1alpha1-%s", randSuffix3())
			workloadType, apiErr := dao.DbUpsertWorkloadType(context.Background(), db, workloadTypeCreateFields(kind, group))
			Expect(apiErr).To(BeNil())
			karta := buildMinimalRI(group, version, kind)
			id := workloadType.ID.String()

			err := upsertKarta(db, karta, id)
			Expect(err).To(BeNil())
			// second call - update same row (matching workload_type_id + version)
			err = upsertKarta(db, karta, id)
			Expect(err).To(BeNil())
			// count rows for this WT/version
			type Row struct{ Count int64 }
			row := Row{}
			err = db.Raw(`SELECT COUNT(1) AS count FROM workload_type_kartas WHERE workload_type_id = ? AND version = ?`, id, version).Row().Scan(&row.Count)
			Expect(err).To(BeNil())
			Expect(row.Count).To(Equal(int64(1)))
		})

		It("runs InitInternalKartas end-to-end and is idempotent", func() {
			// First run
			err := InitInternalKartas(context.Background())
			Expect(err).To(BeNil())
			// Second run
			err = InitInternalKartas(context.Background())
			Expect(err).To(BeNil())

			// Validate presence using configured RI values
			cfgs := getKartasToMigrate()
			Expect(cfgs).ToNot(BeEmpty())
			gvk := cfgs[0].Kartas[0].GroupVersionKind()
			Expect(gvk).ToNot(BeNil())

			db := runaidb.DB()
			type Row struct{ Count int64 }
			row := Row{}
			err = db.Raw(`SELECT COUNT(1) AS count FROM workload_types WHERE name = ? AND "group" = ?`, gvk.Kind, gvk.Group).Row().Scan(&row.Count)
			Expect(err).To(BeNil())
			Expect(row.Count).To(Equal(int64(1)))

			idRow := struct{ ID string }{}
			err = db.Raw(`SELECT id FROM workload_types WHERE name = ? AND "group" = ?`, gvk.Kind, gvk.Group).Row().Scan(&idRow.ID)
			Expect(err).To(BeNil())
			Expect(idRow.ID).ToNot(BeEmpty())

			row = Row{}
			err = db.Raw(`SELECT COUNT(1) AS count FROM external_workloads_type_config WHERE workload_type_id = ?`, idRow.ID).Row().Scan(&row.Count)
			Expect(err).To(BeNil())
			Expect(row.Count).To(Equal(int64(1)))

			row = Row{}
			err = db.Raw(`SELECT COUNT(1) AS count FROM workload_type_kartas WHERE workload_type_id = ? AND version = ?`, idRow.ID, gvk.Version).Row().Scan(&row.Count)
			Expect(err).To(BeNil())
			Expect(row.Count).To(Equal(int64(1)))
		})

		It("Validate InitInternalKartas doesn't update updated_at if not needed", func() {
			// meaning models.WorkloadTypeResourceInterface.UpdatedAt remains the same across idempotent calls
			// First ensure baseline migration exists
			err := InitInternalKartas(context.Background())
			Expect(err).To(BeNil())

			// Snapshot all system-scoped, migration-managed (CreatedBy=Run.ai) RI updated_at values by WT ID + version
			before := map[string]time.Time{}
			labels := map[string]string{} // for better assertion messages
			configs, apiErr := dao.DbListWorkloadTypeConfigs(context.Background(), 0, nil, []queryapi.Filter{}, true)
			Expect(apiErr).To(BeNil())
			Expect(configs).ToNot(BeEmpty())
			for _, c := range configs {
				for _, karta := range c.Kartas {
					if karta.ScopeType == "system" && karta.TenantId == nil && karta.CreatedBy != nil && *karta.CreatedBy == "Run.ai" {
						key := fmt.Sprintf("%s|%s", c.ID.String(), karta.Version)
						before[key] = karta.UpdatedAt
						labels[key] = fmt.Sprintf("%s/%s@%s", c.Group, c.Name, karta.Version)
					}
				}
			}
			Expect(len(before)).To(BeNumerically(">", 0), "expected at least one migration-managed RI")

			// Ensure the clock advances enough to detect unintended updates
			time.Sleep(1100 * time.Millisecond)

			// Re-run migrations; no-op expected for unchanged content
			err = InitInternalKartas(context.Background())
			Expect(err).To(BeNil())

			// Verify updated_at did not change for any tracked entry
			configs2, apiErr2 := dao.DbListWorkloadTypeConfigs(context.Background(), 0, nil, []queryapi.Filter{}, true)
			Expect(apiErr2).To(BeNil())
			after := map[string]time.Time{}
			for _, c := range configs2 {
				for _, karta := range c.Kartas {
					if karta.ScopeType == "system" && karta.TenantId == nil && karta.CreatedBy != nil && *karta.CreatedBy == "Run.ai" {
						key := fmt.Sprintf("%s|%s", c.ID.String(), karta.Version)
						after[key] = karta.UpdatedAt
					}
				}
			}
			Expect(len(after)).To(Equal(len(before)), "mismatch in tracked RIs before/after migration")

			for key, t1 := range before {
				t2, ok := after[key]
				Expect(ok).To(BeTrue(), fmt.Sprintf("missing RI after migration for %s", labels[key]))
				Expect(t2.Equal(t1)).To(BeTrue(), fmt.Sprintf("updated_at changed for %s: before=%v after=%v", labels[key], t1, t2))
			}
		})
	})

	Context("migrations skip user-updated records", func() {
		It("does not override external config updated via UpdateWorkloadType", func() {
			// Ensure baseline migration exists
			err := InitInternalKartas(context.Background())
			Expect(err).To(BeNil())

			// Locate WT for KServe
			cfgs := getKartasToMigrate()
			Expect(cfgs).ToNot(BeEmpty())
			var gvk *schema.GroupVersionKind
			for _, cfg := range cfgs {
				currentGvk := cfg.Kartas[0].GroupVersionKind()
				if currentGvk != nil && currentGvk.Group == "serving.kserve.io" && currentGvk.Kind == "InferenceService" {
					gvk = currentGvk
					break
				}
			}
			Expect(gvk).ToNot(BeNil())

			wt, apiErr := dao.DbGetWorkloadTypeByNameGroup(context.Background(), gvk.Kind, gvk.Group)
			Expect(apiErr).To(BeNil())

			// Choose different category/priority from defaults
			cat, catErr := dao.DbGetCategoryByName(context.Background(), 0, "Build")
			Expect(catErr).To(BeNil())
			prio, prioErr := dao.GetPriorityByName(context.Background(), "0", "high")
			Expect(prioErr).To(BeNil())
			catUUID := uuid.MustParse(cat.Id)
			prioUUID := prio.ID

			// Manually simulate user update to external config (set updated_by != Run.ai)
			db := runaidb.DB()
			userId := "user2"
			_, execErr := db.Exec(`UPDATE external_workloads_type_config
				SET category_id = ?, priority_id = ?, preemptibility = ?, updated_by = ?, updated_at = now()
				WHERE workload_type_id = ? AND scope_type = 'system' AND tenant_id IS NULL`,
				catUUID, prioUUID, string(openapi.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE), userId, wt.ID)
			Expect(execErr).To(BeNil())

			// Re-run migrations; should NOT override user-updated values
			err = InitInternalKartas(context.Background())
			Expect(err).To(BeNil())

			cfgView, getErr := dao.DbGetWorkloadTypeConfigById(context.Background(), 0, wt.ID.String())
			Expect(getErr).To(BeNil())
			Expect(cfgView.CategoryName).To(Equal("Build"))
			Expect(cfgView.PriorityName).To(Equal("high"))
			Expect(cfgView.Preemptibility).To(Equal(string(openapi.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE)))
		})

		It("does not override user-modified RI content for the same version", func() {
			// Ensure baseline migration exists
			err := InitInternalKartas(context.Background())
			Expect(err).To(BeNil())

			// Locate WT for KServe
			cfgs := getKartasToMigrate()
			Expect(cfgs).ToNot(BeEmpty())
			baseRI := cfgs[0].Kartas[0]
			gvk := baseRI.GroupVersionKind()
			Expect(gvk).ToNot(BeNil())

			wt, apiErr := dao.DbGetWorkloadTypeByNameGroup(context.Background(), gvk.Kind, gvk.Group)
			Expect(apiErr).To(BeNil())

			// Build a user-modified RI with the SAME version, but change internal content
			userRI := source.BuildKServeInferenceServiceV1beta1()
			originalVersion := gvk.Version
			changedPath := ".spec.userChanged"
			userRI.Spec.StructureDefinition.ChildComponents[0].SpecDefinition.PodTemplateSpecPath = ptr.To(changedPath)

			// Manually mark the existing RI row as user-owned and replace its content
			b, marshalErr := json.Marshal(userRI)
			Expect(marshalErr).To(BeNil())
			db := runaidb.DB()
			_, execErr := db.Exec(`
				UPDATE workload_type_kartas
				SET karta = ?::jsonb, created_by = ?, updated_at = now()
				WHERE workload_type_id = ? AND version = ? AND scope_type = 'system' AND tenant_id IS NULL
			`, string(b), "user2", wt.ID, originalVersion)
			Expect(execErr).To(BeNil())

			// Re-run migrations; should not overwrite user-modified content for the same version
			err = InitInternalKartas(context.Background())
			Expect(err).To(BeNil())

			// Validate: the row for the SAME version still exists (1) and content remains user-modified
			type Row struct {
				Count int64
			}
			row := Row{}
			err = runaidb.DB().Raw(`SELECT COUNT(1) AS count
					FROM workload_type_kartas
					WHERE workload_type_id = ? AND version = ? AND scope_type = 'system' AND tenant_id IS NULL`,
				wt.ID.String(), originalVersion).Row().Scan(&row.Count)
			Expect(err).To(BeNil())
			Expect(row.Count).To(Equal(int64(1)))

			var (
				content   []byte
				createdBy string
			)
			q := runaidb.DB().Raw(`SELECT karta, COALESCE(created_by, '') FROM workload_type_kartas
					WHERE workload_type_id = ? AND version = ? AND scope_type = 'system' AND tenant_id IS NULL`,
				wt.ID.String(), originalVersion).Row()
			scanErr := q.Scan(&content, &createdBy)
			Expect(scanErr).To(BeNil())
			var got models.Karta
			err = json.Unmarshal(content, &got)
			Expect(err).To(BeNil())
			Expect(*got.Spec.StructureDefinition.ChildComponents[0].SpecDefinition.PodTemplateSpecPath).To(Equal(changedPath))
			Expect(createdBy).To(Equal("user2"))
		})
	})

	Context("suspend definitions and status mappings", func() {
		It("PyTorchJob karta has SuspendDefinition with suspend/resume actions", func() {
			karta := source.BuildKubeflowPyTorchJobV1()
			sd := karta.Spec.StructureDefinition.RootComponent.SuspendDefinition
			Expect(sd).NotTo(BeNil(), "PyTorchJob must have a SuspendDefinition")
			Expect(sd.SuspendActions).NotTo(BeEmpty(), "SuspendActions must not be empty")
			Expect(sd.ResumeActions).NotTo(BeEmpty(), "ResumeActions must not be empty")

			// Verify the well-known kubeflow suspend path is used.
			Expect(sd.SuspendActions[0].Path).To(Equal(".spec.runPolicy.suspend"))
			Expect(sd.SuspendActions[0].Value).To(Equal("true"))
			Expect(sd.ResumeActions[0].Path).To(Equal(".spec.runPolicy.suspend"))
			Expect(sd.ResumeActions[0].Value).To(Equal("false"))
		})

		It("PyTorchJob karta has Suspended StatusMatcher", func() {
			karta := source.BuildKubeflowPyTorchJobV1()
			sd := karta.Spec.StructureDefinition.RootComponent.StatusDefinition
			Expect(sd).NotTo(BeNil())
			Expect(sd.StatusMappings.Suspended).NotTo(BeEmpty(), "PyTorchJob must have a Suspended status matcher")

			// The condition-based matcher must check for Suspended=True.
			matcher := sd.StatusMappings.Suspended[0]
			Expect(matcher.ByConditions).NotTo(BeEmpty())
			Expect(matcher.ByConditions[0].Type).To(Equal("Suspended"))
			Expect(*matcher.ByConditions[0].Status).To(Equal("True"))
		})

		It("PyTorchJob karta has Suspending StatusMatcher", func() {
			karta := source.BuildKubeflowPyTorchJobV1()
			sd := karta.Spec.StructureDefinition.RootComponent.StatusDefinition
			Expect(sd).NotTo(BeNil())
			Expect(sd.StatusMappings.Suspending).NotTo(BeEmpty(), "PyTorchJob must have a Suspending status matcher")
			Expect(sd.StatusMappings.Resuming).To(BeEmpty(), "PyTorchJob must not have a Resuming status matcher")
		})

		It("Kubernetes Job karta has SuspendDefinition using spec.suspend", func() {
			karta := kubernetes.BuildK8sJobV1()
			sd := karta.Spec.StructureDefinition.RootComponent.SuspendDefinition
			Expect(sd).NotTo(BeNil(), "K8s Job must have a SuspendDefinition")
			Expect(sd.SuspendActions).NotTo(BeEmpty())
			Expect(sd.ResumeActions).NotTo(BeEmpty())
			Expect(sd.SuspendActions[0].Path).To(Equal(".spec.suspend"))
			Expect(sd.SuspendActions[0].Value).To(Equal("true"))
		})

		It("Kubernetes Job karta has Suspended StatusMatcher", func() {
			karta := kubernetes.BuildK8sJobV1()
			sd := karta.Spec.StructureDefinition.RootComponent.StatusDefinition
			Expect(sd).NotTo(BeNil())
			Expect(sd.StatusMappings.Suspended).NotTo(BeEmpty(), "K8s Job must have a Suspended status matcher")
		})

		It("after InitInternalKartas PyTorchJob karta in DB contains SuspendDefinition", func() {
			err := InitInternalKartas(context.Background())
			Expect(err).To(BeNil())

			db := runaidb.DB()
			idRow := struct{ ID string }{}
			err = db.Raw(`SELECT id FROM workload_types WHERE name = ? AND "group" = ?`, "PyTorchJob", "kubeflow.org").Row().Scan(&idRow.ID)
			Expect(err).To(BeNil())
			Expect(idRow.ID).NotTo(BeEmpty())

			var content []byte
			version, scopeType := "v1", "system"
			err = db.Raw(`SELECT karta FROM workload_type_kartas WHERE workload_type_id = ? AND version = ? AND scope_type = ? AND tenant_id IS NULL`, idRow.ID, version, scopeType).Row().Scan(&content)
			Expect(err).To(BeNil())

			var got models.Karta
			Expect(json.Unmarshal(content, &got)).To(Succeed())

			sd := got.Spec.StructureDefinition.RootComponent.SuspendDefinition
			Expect(sd).NotTo(BeNil(), "persisted PyTorchJob karta must have a SuspendDefinition")
			Expect(sd.SuspendActions).NotTo(BeEmpty())
			Expect(got.Spec.StructureDefinition.RootComponent.StatusDefinition.StatusMappings.Suspended).NotTo(BeEmpty())
		})
	})

	Context("initInternalKartas validations", func() {
		It("returns error when category is empty", func() {
			invalidCfg := []source.KartaConfiguration{
				{
					Category:       "",
					Priority:       "low",
					Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE,
					Kartas:         []models.Karta{source.BuildKServeInferenceServiceV1beta1()},
				},
			}
			err := initInternalKartas(context.Background(), invalidCfg)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("category must not be empty"))
		})

		It("returns error when priority is empty", func() {
			invalidCfg := []source.KartaConfiguration{
				{
					Category:       "Deploy",
					Priority:       "",
					Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
					Kartas:         []models.Karta{source.BuildKServeInferenceServiceV1beta1()},
				},
			}
			err := initInternalKartas(context.Background(), invalidCfg)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("priority must not be empty"))
		})

		It("returns error when preemptibility is empty", func() {
			invalidCfg := []source.KartaConfiguration{
				{
					Category:       "Deploy",
					Priority:       "very-high",
					Preemptibility: "",
					Kartas:         []models.Karta{source.BuildKServeInferenceServiceV1beta1()},
				},
			}
			err := initInternalKartas(context.Background(), invalidCfg)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("preemptibility must be a valid value"))
		})

		It("returns error when GVK is invalid", func() {
			invalidCfg := []source.KartaConfiguration{
				{
					Category:       "Deploy",
					Priority:       "very-high",
					Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
					Kartas:         []models.Karta{{}}, // invalid GVK
				},
			}
			err := initInternalKartas(context.Background(), invalidCfg)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("invalid or missing GVK"))
		})

		It("returns error when RIs have mixed group/kind", func() {
			a := source.BuildKServeInferenceServiceV1beta1()
			b := source.BuildKServeInferenceServiceV1beta1()
			b.Spec.StructureDefinition.RootComponent.Kind.Kind = "OtherKind"
			invalidCfg := []source.KartaConfiguration{
				{
					Category:       "Deploy",
					Priority:       "very-high",
					Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
					Kartas:         []models.Karta{a, b},
				},
			}
			err := initInternalKartas(context.Background(), invalidCfg)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("mismatched group/kind"))
		})

		It("returns error when duplicate versions exist", func() {
			a := source.BuildKServeInferenceServiceV1beta1()
			b := source.BuildKServeInferenceServiceV1beta1()
			invalidCfg := []source.KartaConfiguration{
				{
					Category:       "Deploy",
					Priority:       "very-high",
					Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE,
					Kartas:         []models.Karta{a, b},
				},
			}
			err := initInternalKartas(context.Background(), invalidCfg)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("duplicate RI version"))
		})
	})
})

func randSuffix3() string {
	return rand.String(3)
	//const letters = "abcdefghijklmnopqrstuvwxyz"
	//b := []byte{letters[rand.Intn(len(letters))], letters[rand.Intn(len(letters))], letters[rand.Intn(len(letters))]}
	//return string(b)
}

func buildMinimalRI(group, version, kind string) models.Karta {
	return models.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   group,
						Version: version,
						Kind:    kind,
					},
					SpecDefinition: &kartav1alpha1.SpecDefinition{
						PodTemplateSpecPath: ptr.To(".spec.template"),
					},
				},
			},
		},
	}
}
