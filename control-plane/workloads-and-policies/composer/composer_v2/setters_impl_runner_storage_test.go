package composer_v2

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	openapi "github.com/run-ai/runai/runai-clients/workloads-manager-client/generated"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("MapPvcToWorkloadRunner", func() {
	It("should create manifest volume and mount only for existing PVCs", func() {
		target := &composer.VolumesAndMounts{Config: &openapi.WorkloadV2Configuration{}}
		pvc := workloadsApi.NewPvcInstance()
		pvc.SetExistingPvc(true)
		pvc.SetClaimName("existing-claim")
		pvc.SetPath("/mnt/existing")
		pvc.SetReadOnly(true)

		err := MapPvcToWorkloadRunner(map[string]*workloadsApi.PvcInstance{"pvc-0": pvc}, target, nil)

		Expect(err).To(BeNil())
		Expect(target.Volume).To(HaveLen(1))
		Expect(target.Volume[0].Name).To(Equal("pvc-0"))
		Expect(target.Volume[0].PersistentVolumeClaim).NotTo(BeNil())
		Expect(target.Volume[0].PersistentVolumeClaim.ClaimName).To(Equal("existing-claim"))
		Expect(target.Volume[0].PersistentVolumeClaim.ReadOnly).To(BeTrue())
		Expect(target.VolumeMount).To(Equal([]corev1.VolumeMount{{
			Name:      "pvc-0",
			MountPath: "/mnt/existing",
			ReadOnly:  true,
		}}))
		Expect(target.Config.Pvcs).To(BeEmpty())
		Expect(target.Config.PvcsConfig.IsSet()).To(BeFalse())
	})

	It("should create manifest volume, mount, and configuration entry for new persistent PVCs", func() {
		target := &composer.VolumesAndMounts{Config: &openapi.WorkloadV2Configuration{}}
		pvc := workloadsApi.NewPvcInstance()
		pvc.SetClaimName("new-claim")
		pvc.SetPath("/mnt/new")
		pvc.SetReadOnly(true)

		accessModes := workloadsApi.NewPvcAccessModes()
		accessModes.SetReadWriteOnce(true)
		accessModes.SetReadOnlyMany(true)
		volumeMode := workloadsApi.PVCVOLUMEMODE_BLOCK
		claimInfo := workloadsApi.NewClaimInfo()
		claimInfo.SetSize("10Gi")
		claimInfo.SetStorageClass("fast")
		claimInfo.SetAccessModes(*accessModes)
		claimInfo.SetVolumeMode(volumeMode)
		claimInfo.SetAddedAttrValues([]workloadsApi.PvcAddedAttrValue{{
			Key:   "csi.example/hint",
			Value: convutils.Ref("enabled"),
		}})
		pvc.SetClaimInfo(*claimInfo)

		err := MapPvcToWorkloadRunner(map[string]*workloadsApi.PvcInstance{"pvc-0": pvc}, target, nil)

		Expect(err).To(BeNil())
		Expect(target.Volume).To(HaveLen(1))
		Expect(target.Volume[0].PersistentVolumeClaim).NotTo(BeNil())
		Expect(target.Volume[0].PersistentVolumeClaim.ClaimName).To(Equal("new-claim"))
		Expect(target.Volume[0].PersistentVolumeClaim.ReadOnly).To(BeTrue())
		Expect(target.VolumeMount).To(Equal([]corev1.VolumeMount{{
			Name:      "pvc-0",
			MountPath: "/mnt/new",
			ReadOnly:  true,
		}}))

		Expect(target.Config.Pvcs).To(HaveLen(1))
		configPvc := target.Config.Pvcs[0]
		Expect(configPvc.Metadata.Name).To(Equal("new-claim"))
		Expect(configPvc.Metadata.Annotations).NotTo(BeNil())
		Expect(*configPvc.Metadata.Annotations).To(Equal(map[string]string{"csi.example/hint": "enabled"}))
		Expect(configPvc.Spec.AccessModes).To(Equal([]openapi.WorkloadV2UpdateConfigurationPvcAccessMode{
			openapi.WORKLOADV2UPDATECONFIGURATIONPVCACCESSMODE_READ_WRITE_ONCE,
			openapi.WORKLOADV2UPDATECONFIGURATIONPVCACCESSMODE_READ_ONLY_MANY,
		}))
		Expect(configPvc.Spec.Resources.Requests.Storage).To(Equal("10Gi"))
		Expect(configPvc.Spec.GetStorageClassName()).To(Equal("fast"))
		Expect(configPvc.Spec.GetVolumeMode()).To(Equal(openapi.WORKLOADV2UPDATECONFIGURATIONPVCVOLUMEMODE_BLOCK))

		// the cleanup default is applied by ComposePvc per composer options, never by the mapper
		Expect(target.Config.PvcsConfig.IsSet()).To(BeFalse())
	})

	It("should not touch an already-set pvcs config for new persistent PVCs", func() {
		cleanupOnCompletion := false
		presetConfig := &openapi.WorkloadV2Configuration{}
		presetConfig.PvcsConfig.Set(&openapi.WorkloadV2UpdateConfigurationPvcsConfig{
			CleanupOnCompletion: *openapi.NewNullableBool(&cleanupOnCompletion),
		})
		target := &composer.VolumesAndMounts{Config: presetConfig}

		pvc := workloadsApi.NewPvcInstance()
		pvc.SetClaimName("new-claim")
		pvc.SetPath("/mnt/new")

		accessModes := workloadsApi.NewPvcAccessModes()
		accessModes.SetReadWriteOnce(true)
		claimInfo := workloadsApi.NewClaimInfo()
		claimInfo.SetSize("10Gi")
		claimInfo.SetAccessModes(*accessModes)
		pvc.SetClaimInfo(*claimInfo)

		err := MapPvcToWorkloadRunner(map[string]*workloadsApi.PvcInstance{"pvc-0": pvc}, target, nil)

		Expect(err).To(BeNil())
		Expect(target.Config.Pvcs).To(HaveLen(1))
		Expect(target.Config.PvcsConfig.IsSet()).To(BeTrue())
		Expect(target.Config.PvcsConfig.Get().CleanupOnCompletion.Get()).To(HaveValue(BeFalse()))
	})

	It("should create inline ephemeral volume and no configuration entry for new ephemeral PVCs", func() {
		target := &composer.VolumesAndMounts{}
		pvc := workloadsApi.NewPvcInstance()
		pvc.SetClaimName("ignored-for-ephemeral")
		pvc.SetPath("/mnt/ephemeral")
		pvc.SetReadOnly(true)
		pvc.SetEphemeral(true)

		accessModes := &workloadsApi.PvcAccessModes{}
		accessModes.SetReadWriteMany(true)
		volumeMode := workloadsApi.PVCVOLUMEMODE_FILESYSTEM
		claimInfo := workloadsApi.NewClaimInfo()
		claimInfo.SetSize("5Gi")
		claimInfo.SetStorageClass("ephemeral-class")
		claimInfo.SetAccessModes(*accessModes)
		claimInfo.SetVolumeMode(volumeMode)
		pvc.SetClaimInfo(*claimInfo)

		err := MapPvcToWorkloadRunner(map[string]*workloadsApi.PvcInstance{"pvc-0": pvc}, target, nil)

		Expect(err).To(BeNil())
		Expect(target.Volume).To(HaveLen(1))
		Expect(target.Volume[0].PersistentVolumeClaim).To(BeNil())
		Expect(target.Volume[0].Ephemeral).NotTo(BeNil())
		template := target.Volume[0].Ephemeral.VolumeClaimTemplate
		Expect(template).NotTo(BeNil())
		Expect(template.Spec.AccessModes).To(Equal([]corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany}))
		storageQuantity := template.Spec.Resources.Requests[corev1.ResourceStorage]
		Expect(storageQuantity.String()).To(Equal("5Gi"))
		Expect(template.Spec.StorageClassName).NotTo(BeNil())
		Expect(*template.Spec.StorageClassName).To(Equal("ephemeral-class"))
		Expect(template.Spec.VolumeMode).NotTo(BeNil())
		Expect(*template.Spec.VolumeMode).To(Equal(corev1.PersistentVolumeFilesystem))
		Expect(target.VolumeMount).To(Equal([]corev1.VolumeMount{{
			Name:      "pvc-0",
			MountPath: "/mnt/ephemeral",
			ReadOnly:  true,
		}}))
		Expect(target.Config).To(BeNil())
	})

	It("should default omitted access modes to ReadWriteOnce for new persistent PVCs", func() {
		target := &composer.VolumesAndMounts{Config: &openapi.WorkloadV2Configuration{}}
		pvc := workloadsApi.NewPvcInstance()
		pvc.SetClaimName("no-access-modes-claim")
		pvc.SetPath("/mnt/no-access-modes")

		claimInfo := workloadsApi.NewClaimInfo()
		claimInfo.SetSize("2Gi")
		claimInfo.SetStorageClass("standard")
		pvc.SetClaimInfo(*claimInfo)

		err := MapPvcToWorkloadRunner(map[string]*workloadsApi.PvcInstance{"pvc-0": pvc}, target, nil)

		Expect(err).To(BeNil())
		Expect(target.Config.Pvcs).To(HaveLen(1))
		Expect(target.Config.Pvcs[0].Spec.AccessModes).To(Equal([]openapi.WorkloadV2UpdateConfigurationPvcAccessMode{
			openapi.WORKLOADV2UPDATECONFIGURATIONPVCACCESSMODE_READ_WRITE_ONCE,
		}))
	})

	It("should default omitted access modes to ReadWriteOnce for new ephemeral PVCs", func() {
		target := &composer.VolumesAndMounts{}
		pvc := workloadsApi.NewPvcInstance()
		pvc.SetClaimName("ephemeral-no-access-modes")
		pvc.SetPath("/mnt/ephemeral-no-access-modes")
		pvc.SetEphemeral(true)

		claimInfo := workloadsApi.NewClaimInfo()
		claimInfo.SetSize("3Gi")
		claimInfo.SetStorageClass("ephemeral-class")
		pvc.SetClaimInfo(*claimInfo)

		err := MapPvcToWorkloadRunner(map[string]*workloadsApi.PvcInstance{"pvc-0": pvc}, target, nil)

		Expect(err).To(BeNil())
		Expect(target.Volume).To(HaveLen(1))
		Expect(target.Volume[0].Ephemeral).NotTo(BeNil())
		template := target.Volume[0].Ephemeral.VolumeClaimTemplate
		Expect(template).NotTo(BeNil())
		Expect(template.Spec.AccessModes).To(Equal([]corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}))
	})

	It("should reject a target that is not VolumesAndMounts", func() {
		pvc := workloadsApi.NewPvcInstance()
		pvc.SetExistingPvc(true)
		pvc.SetClaimName("claim")

		err := MapPvcToWorkloadRunner(map[string]*workloadsApi.PvcInstance{"pvc-0": pvc}, &struct{}{}, nil)

		Expect(err).NotTo(BeNil())
		Expect(err.Code).To(Equal(int32(http.StatusInternalServerError)))
	})

	Describe("CopyBoolean with BoolToPvcsCleanupConfig", func() {
		configWithNewPvc := func() *openapi.WorkloadV2Configuration {
			return &openapi.WorkloadV2Configuration{
				Pvcs: []openapi.WorkloadV2UpdateConfigurationPvc{{Metadata: openapi.WorkloadV2UpdateConfigurationPvcMetadata{Name: "new-claim"}}},
			}
		}
		apply := func(cleanup *bool, config *openapi.WorkloadV2Configuration) {
			fc := &composer.FieldsComposer{Options: composer.FieldsComposerOptions{CleanupCreatedPvcsOnCompletion: cleanup}}
			Expect(fc.CopyBoolean(fc.Options.CleanupCreatedPvcsOnCompletion, config, BoolToPvcsCleanupConfig, nil)()).To(BeNil())
		}

		It("should set cleanup for workload-created PVCs when the option is true", func() {
			config := configWithNewPvc()
			apply(convutils.Ref(true), config)
			Expect(config.PvcsConfig.IsSet()).To(BeTrue())
			Expect(config.PvcsConfig.Get().CleanupOnCompletion.Get()).To(HaveValue(BeTrue()))
		})

		It("should set explicit no-cleanup when the option is false", func() {
			config := configWithNewPvc()
			apply(convutils.Ref(false), config)
			Expect(config.PvcsConfig.IsSet()).To(BeTrue())
			Expect(config.PvcsConfig.Get().CleanupOnCompletion.Get()).To(HaveValue(BeFalse()))
		})

		It("should write nothing when the option is nil, leaving the API default", func() {
			config := configWithNewPvc()
			apply(nil, config)
			Expect(config.PvcsConfig.IsSet()).To(BeFalse())
		})

		It("should write nothing when the workload creates no PVCs", func() {
			config := &openapi.WorkloadV2Configuration{}
			apply(convutils.Ref(true), config)
			Expect(config.PvcsConfig.IsSet()).To(BeFalse())
		})

		It("should merge into an existing pvcs config whose cleanup field is unset", func() {
			config := configWithNewPvc()
			config.PvcsConfig.Set(&openapi.WorkloadV2UpdateConfigurationPvcsConfig{})
			apply(convutils.Ref(true), config)
			Expect(config.PvcsConfig.IsSet()).To(BeTrue())
			Expect(config.PvcsConfig.Get().CleanupOnCompletion.Get()).To(HaveValue(BeTrue()))
		})

		It("should not override an already-set pvcs config", func() {
			config := configWithNewPvc()
			preset := false
			config.PvcsConfig.Set(&openapi.WorkloadV2UpdateConfigurationPvcsConfig{
				CleanupOnCompletion: *openapi.NewNullableBool(&preset),
			})
			apply(convutils.Ref(true), config)
			Expect(config.PvcsConfig.Get().CleanupOnCompletion.Get()).To(HaveValue(BeFalse()))
		})

		It("should tolerate a nil config", func() {
			apply(convutils.Ref(true), nil)
		})

		It("should reject a target that is not a workload configuration", func() {
			err := BoolToPvcsCleanupConfig(convutils.Ref(true), &struct{}{}, nil)
			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(Equal(int32(http.StatusInternalServerError)))
		})
	})

	It("should reject new persistent PVCs when Config is missing", func() {
		pvc := workloadsApi.NewPvcInstance()
		pvc.SetClaimName("new-claim")
		pvc.SetPath("/mnt/new")

		accessModes := workloadsApi.NewPvcAccessModes()
		accessModes.SetReadWriteOnce(true)
		claimInfo := workloadsApi.NewClaimInfo()
		claimInfo.SetSize("1Gi")
		claimInfo.SetAccessModes(*accessModes)
		pvc.SetClaimInfo(*claimInfo)

		err := MapPvcToWorkloadRunner(map[string]*workloadsApi.PvcInstance{"pvc-0": pvc}, &composer.VolumesAndMounts{}, nil)

		Expect(err).NotTo(BeNil())
		Expect(err.Code).To(Equal(int32(http.StatusInternalServerError)))
		Expect(err.Message).To(ContainSubstring("target Config must not be nil"))
	})
})
