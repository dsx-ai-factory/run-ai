package builders

import (
	"context"
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	defaultNamespace = "default"
)

// MockClient is a mock implementation of client.Client for testing
type MockClient struct {
	client.Client
	MockGet func(ctx context.Context, key client.ObjectKey, obj client.Object) error
	scheme  *runtime.Scheme
}

func (m *MockClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if m.MockGet != nil {
		return m.MockGet(ctx, key, obj)
	}
	return nil
}

func (m *MockClient) Scheme() *runtime.Scheme {
	return m.scheme
}

// setupScheme creates a scheme for testing
func setupScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	Expect(corev1.AddToScheme(scheme)).To(Succeed())
	return scheme
}

var _ = Describe("PvcBuilder tests", func() {
	var (
		scheme     *runtime.Scheme
		fakeClient client.Client

		pb  *PvcBuilder
		ctx context.Context

		workloadBaseName string
		objectMeta       *metav1.ObjectMeta
		owner            *corev1.Pod
		nameSuffix       string
		jobType          string
	)

	BeforeEach(func() {
		// Setup common test variables
		scheme = setupScheme()
		fakeClient = fake.NewClientBuilder().WithScheme(scheme).Build()

		pb = NewPvcBuilder(fakeClient)
		ctx = context.Background()

		workloadBaseName = "test-workload"
		objectMeta = &metav1.ObjectMeta{
			Name:      workloadBaseName,
			Namespace: defaultNamespace,
		}
		owner = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: defaultNamespace,
				UID:       "test-uid",
			},
		}
		nameSuffix = ""
		jobType = "interactive"
	})

	Context("GetPreview and Build", func() {
		It("should return empty result for empty pvcs", func() {
			pvcs := &flds.PvcsField{}

			preview, err := pb.GetPreview(ctx, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).NotTo(HaveOccurred())
			Expect(preview).To(BeEmpty())

			result, err := pb.Build(ctx, jobType, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeEmpty())
		})

		It("should skip ephemeral volumes", func() {
			pvcsMap := map[string]flds.PvcsItemAndRules{
				"ephemeral-volume": {
					Value: &flds.PvcsItemValue{
						Ephemeral: ptr.To(true),
						Size:      ptr.To("10Gi"),
					},
				},
			}
			pvcs := &flds.PvcsField{
				Value: &pvcsMap,
			}

			preview, err := pb.GetPreview(ctx, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).NotTo(HaveOccurred())
			Expect(preview).To(BeEmpty())

			result, err := pb.Build(ctx, jobType, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeEmpty())
		})

		It("should skip existing pvcs", func() {
			pvcsMap := map[string]flds.PvcsItemAndRules{
				"existing-volume": {
					Value: &flds.PvcsItemValue{
						ExistingPvc: ptr.To(true),
						ClaimName:   ptr.To("existing-claim"),
					},
				},
			}
			pvcs := &flds.PvcsField{
				Value: &pvcsMap,
			}

			preview, err := pb.GetPreview(ctx, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).NotTo(HaveOccurred())
			Expect(preview).To(BeEmpty())

			result, err := pb.Build(ctx, jobType, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeEmpty())
		})

		It("should use correct claim name index when pvc types are mixed", func() {
			// important notes:
			// * pvc order is important - in order to keep the map sorted we use the prefixes pvc1, pvc2, ...
			// * we intentionally test a case where a new pvc comes AFTER an existing one, in order to check
			//   the indexing, as taken care in skipPvcBuilding function
			pvcsMap := map[string]flds.PvcsItemAndRules{
				"pvc1-new": {
					Value: &flds.PvcsItemValue{
						Size: ptr.To("10M"),
					},
				},
				"pvc2-new": {
					Value: &flds.PvcsItemValue{
						Size: ptr.To("20M"),
					},
				},
				"pvc3-ephemeral": {
					Value: &flds.PvcsItemValue{
						Size:      ptr.To("30M"),
						Ephemeral: ptr.To(true),
					},
				},
				"pvc4-existing": {
					Value: &flds.PvcsItemValue{
						ClaimName:   ptr.To("existing-pvc"),
						ExistingPvc: ptr.To(true),
					},
				},
				"pvc5-new": {
					Value: &flds.PvcsItemValue{
						Size: ptr.To("40M"),
					},
				},
			}
			pvcs := &flds.PvcsField{
				Value: &pvcsMap,
			}

			preview, err := pb.GetPreview(ctx, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).NotTo(HaveOccurred())
			Expect(preview).To(HaveLen(3)) // only new PVCs

			// check the claim names
			Expect(preview[0].Name).To(Equal("pvc-test-workload-0"))
			Expect(preview[1].Name).To(Equal("pvc-test-workload-1"))
			Expect(preview[2].Name).To(Equal("pvc-test-workload-3")) // skipping the index of the existing PVC (but not of the ephemeral one)

			result, err := pb.Build(ctx, jobType, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(3)) // only new PVCs

			// check the claim names
			Expect(result[0].Name).To(Equal("pvc-test-workload-0"))
			Expect(result[1].Name).To(Equal("pvc-test-workload-1"))
			Expect(result[2].Name).To(Equal("pvc-test-workload-3")) // skipping the index of the existing PVC (but not of the ephemeral one)
		})

		It("should build PVCs with default settings", func() {
			pvcsMap := map[string]flds.PvcsItemAndRules{
				"data-volume": {
					Value: &flds.PvcsItemValue{
						Size: ptr.To("10Gi"),
					},
				},
			}
			pvcs := &flds.PvcsField{
				Value: &pvcsMap,
			}

			preview, err := pb.GetPreview(ctx, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).NotTo(HaveOccurred())
			Expect(preview).To(HaveLen(1))

			// Verify the PVC preview properties
			pvcPreview := preview[0]
			Expect(pvcPreview.Name).To(Equal("pvc-test-workload-0"))
			Expect(pvcPreview.Namespace).To(Equal(defaultNamespace))
			Expect(pvcPreview.Kind.Kind).To(Equal(constants.KindPvc))

			result, err := pb.Build(ctx, jobType, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(1))

			// Verify the PVC properties
			pvc := result[0]
			Expect(pvc.Name).To(Equal("pvc-test-workload-0"))
			Expect(pvc.Namespace).To(Equal(defaultNamespace))
			Expect(pvc.Labels[constants.LabelCreatedBy]).To(Equal(jobType))

			// Check storage request
			storageRequest := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
			Expect(storageRequest.String()).To(Equal("10Gi"))

			// Check default access mode
			Expect(pvc.Spec.AccessModes).To(HaveLen(1))
			Expect(pvc.Spec.AccessModes).To(ContainElement(corev1.ReadWriteOnce))

			// Check controller reference
			Expect(metav1.IsControlledBy(&pvc, owner)).To(BeTrue())
		})

		It("should build PVCs with explicit storage class", func() {
			pvcsMap := map[string]flds.PvcsItemAndRules{
				"data-volume": {
					Value: &flds.PvcsItemValue{
						Size:         ptr.To("10Gi"),
						StorageClass: ptr.To("premium-storage"),
					},
				},
			}
			pvcs := &flds.PvcsField{
				Value: &pvcsMap,
			}

			preview, err := pb.GetPreview(ctx, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).NotTo(HaveOccurred())
			Expect(preview).To(HaveLen(1))
			Expect(preview[0].Name).To(Equal("pvc-test-workload-0"))
			Expect(preview[0].Namespace).To(Equal(defaultNamespace))

			result, err := pb.Build(ctx, jobType, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(1))
			Expect(*result[0].Spec.StorageClassName).To(Equal("premium-storage"))
		})

		It("should build PVCs with empty storage class", func() {
			pvcsMap := map[string]flds.PvcsItemAndRules{
				"data-volume": {
					Value: &flds.PvcsItemValue{
						Size:         ptr.To("10Gi"),
						StorageClass: ptr.To(constants.EmptyStorageClass),
					},
				},
			}
			pvcs := &flds.PvcsField{
				Value: &pvcsMap,
			}

			preview, err := pb.GetPreview(ctx, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).NotTo(HaveOccurred())
			Expect(preview).To(HaveLen(1))
			Expect(preview[0].Name).To(Equal("pvc-test-workload-0"))
			Expect(preview[0].Namespace).To(Equal(defaultNamespace))

			result, err := pb.Build(ctx, jobType, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(1))
			Expect(*result[0].Spec.StorageClassName).To(Equal(""))
		})

		It("should build PVCs with explicit claim name", func() {
			pvcsMap := map[string]flds.PvcsItemAndRules{
				"data-volume": {
					Value: &flds.PvcsItemValue{
						Size:      ptr.To("10Gi"),
						ClaimName: ptr.To("explicit-data-claim"),
					},
				},
			}
			pvcs := &flds.PvcsField{
				Value: &pvcsMap,
			}

			preview, err := pb.GetPreview(ctx, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).NotTo(HaveOccurred())
			Expect(preview).To(HaveLen(1))
			Expect(preview[0].Name).To(Equal("explicit-data-claim"))
			Expect(preview[0].Namespace).To(Equal(defaultNamespace))

			result, err := pb.Build(ctx, jobType, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(1))
			Expect(result[0].Name).To(Equal("explicit-data-claim"))
		})

		It("should build PVCs with custom access modes", func() {
			pvcsMap := map[string]flds.PvcsItemAndRules{
				"data-volume": {
					Value: &flds.PvcsItemValue{
						Size:          ptr.To("10Gi"),
						ReadOnlyMany:  ptr.To(true),
						ReadWriteMany: ptr.To(true),
						ReadWriteOnce: ptr.To(false),
					},
				},
			}
			pvcs := &flds.PvcsField{
				Value: &pvcsMap,
			}

			preview, err := pb.GetPreview(ctx, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).NotTo(HaveOccurred())
			Expect(preview).To(HaveLen(1))
			Expect(preview[0].Name).To(Equal("pvc-test-workload-0"))
			Expect(preview[0].Namespace).To(Equal(defaultNamespace))

			result, err := pb.Build(ctx, jobType, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(1))
			Expect(result[0].Spec.AccessModes).To(HaveLen(2))
			Expect(result[0].Spec.AccessModes).To(ContainElement(corev1.ReadOnlyMany))
			Expect(result[0].Spec.AccessModes).To(ContainElement(corev1.ReadWriteMany))
			Expect(result[0].Spec.AccessModes).NotTo(ContainElement(corev1.ReadWriteOnce))
		})

		It("should build PVCs with volume mode", func() {
			volumeMode := "Block"
			pvcsMap := map[string]flds.PvcsItemAndRules{
				"data-volume": {
					Value: &flds.PvcsItemValue{
						Size:       ptr.To("10Gi"),
						VolumeMode: &volumeMode,
					},
				},
			}
			pvcs := &flds.PvcsField{
				Value: &pvcsMap,
			}

			preview, err := pb.GetPreview(ctx, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).NotTo(HaveOccurred())
			Expect(preview).To(HaveLen(1))
			Expect(preview[0].Name).To(Equal("pvc-test-workload-0"))
			Expect(preview[0].Namespace).To(Equal(defaultNamespace))

			result, err := pb.Build(ctx, jobType, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(1))
			expectedVolumeMode := corev1.PersistentVolumeMode(volumeMode)
			Expect(result[0].Spec.VolumeMode).To(Equal(&expectedVolumeMode))
		})

		It("should build multiple PVCs with name suffix", func() {
			pvcsMap := map[string]flds.PvcsItemAndRules{
				"data-volume": {
					Value: &flds.PvcsItemValue{
						Size: ptr.To("10Gi"),
					},
				},
				"model-volume": {
					Value: &flds.PvcsItemValue{
						Size: ptr.To("50Gi"),
					},
				},
			}
			pvcs := &flds.PvcsField{
				Value: &pvcsMap,
			}
			nameSuffix2 := "job-123"

			preview, err := pb.GetPreview(ctx, objectMeta, workloadBaseName, owner, pvcs, nameSuffix2)
			Expect(err).NotTo(HaveOccurred())
			Expect(preview).To(HaveLen(2))

			// Sort preview by name to ensure consistent order for assertions
			previewNames := []string{preview[0].Name, preview[1].Name}
			Expect(previewNames).To(ContainElement("pvc-test-workload-job-123-0"))
			Expect(previewNames).To(ContainElement("pvc-test-workload-job-123-1"))
			Expect(preview[0].Namespace).To(Equal(defaultNamespace))
			Expect(preview[1].Namespace).To(Equal(defaultNamespace))

			result, err := pb.Build(ctx, jobType, objectMeta, workloadBaseName, owner, pvcs, nameSuffix2)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(2))

			// Sort result by name to ensure consistent order for assertions
			resultNames := []string{result[0].Name, result[1].Name}
			Expect(resultNames).To(ContainElement("pvc-test-workload-job-123-0"))
			Expect(resultNames).To(ContainElement("pvc-test-workload-job-123-1"))
		})

		It("should return error with invalid size", func() {
			pvcsMap := map[string]flds.PvcsItemAndRules{
				"data-volume": {
					Value: &flds.PvcsItemValue{
						Size: ptr.To("invalid-size"),
					},
				},
			}
			pvcs := &flds.PvcsField{
				Value: &pvcsMap,
			}

			preview, err := pb.GetPreview(ctx, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).NotTo(HaveOccurred())
			Expect(preview).To(HaveLen(1))

			result, err := pb.Build(ctx, jobType, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("invalid size for creating a PVC"))
		})

		It("should return error when PVC validation fails", func() {
			mockScheme := runtime.NewScheme()
			Expect(corev1.AddToScheme(mockScheme)).To(Succeed())

			mockClient := &MockClient{
				Client: fakeClient,
				MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
					// Simulate a non-NotFound error
					return errors.New("failed to get PVC")
				},
				scheme: mockScheme,
			}

			pb = NewPvcBuilder(mockClient)
			pvcsMap := map[string]flds.PvcsItemAndRules{
				"data-volume": {
					Value: &flds.PvcsItemValue{
						Size: ptr.To("10Gi"),
					},
				},
			}
			pvcs := &flds.PvcsField{
				Value: &pvcsMap,
			}

			preview, err := pb.GetPreview(ctx, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).To(HaveOccurred())
			Expect(preview).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("invalid PVC pvc-test-workload-0: failed to get PVC"))

			result, err := pb.Build(ctx, jobType, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("invalid PVC pvc-test-workload-0: failed to get PVC"))
		})

		It("should return error when setting controller reference fails", func() {
			// We need a scheme that doesn't have the owner type registered to trigger an error
			incompleteScheme := runtime.NewScheme()
			// Deliberately NOT adding Pod to the scheme

			mockClient := &MockClient{
				Client: fakeClient,
				MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
					return apierrors.NewNotFound(schema.GroupResource{Resource: "persistentvolumeclaims"}, key.Name)
				},
				scheme: incompleteScheme,
			}

			pb = NewPvcBuilder(mockClient)
			pvcsMap := map[string]flds.PvcsItemAndRules{
				"data-volume": {
					Value: &flds.PvcsItemValue{
						Size: ptr.To("10Gi"),
					},
				},
			}
			pvcs := &flds.PvcsField{
				Value: &pvcsMap,
			}

			// GetPreview should work fine since it doesn't set controller references
			preview, err := pb.GetPreview(ctx, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).NotTo(HaveOccurred())
			Expect(preview).To(HaveLen(1))

			result, err := pb.Build(ctx, jobType, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to set controller reference"))
		})

		It("should handle nil PVC value", func() {
			pvcsMap := map[string]flds.PvcsItemAndRules{
				"nil-volume": {
					Value: nil, // Nil value should be skipped
				},
				"valid-volume": {
					Value: &flds.PvcsItemValue{
						Size: ptr.To("10Gi"),
					},
				},
			}
			pvcs := &flds.PvcsField{
				Value: &pvcsMap,
			}

			preview, err := pb.GetPreview(ctx, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).NotTo(HaveOccurred())
			Expect(preview).To(HaveLen(1)) // Only the valid volume should be processed
			Expect(preview[0].Name).To(Equal("pvc-test-workload-0"))

			result, err := pb.Build(ctx, jobType, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(1)) // Only the valid volume should be processed
			Expect(result[0].Name).To(Equal("pvc-test-workload-0"))
		})

		It("should handle nil pvcs field", func() {
			var pvcs *flds.PvcsField = nil // Nil PvcsField

			preview, err := pb.GetPreview(ctx, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).NotTo(HaveOccurred())
			Expect(preview).To(BeEmpty())

			result, err := pb.Build(ctx, jobType, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeEmpty())
		})

		It("should handle expired context", func() {
			mockScheme := runtime.NewScheme()
			Expect(corev1.AddToScheme(mockScheme)).To(Succeed())

			// Setup mock client that will check if context is done and return error if it is
			mockClient := &MockClient{
				Client: fakeClient,
				MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
					select {
					case <-ctx.Done():
						return ctx.Err()
					default:
						return apierrors.NewNotFound(schema.GroupResource{Resource: "persistentvolumeclaims"}, key.Name)
					}
				},
				scheme: mockScheme,
			}

			pb = NewPvcBuilder(mockClient)
			// Create an already expired context
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Microsecond)
			defer cancel()
			time.Sleep(10 * time.Millisecond) // Ensure context is expired

			pvcsMap := map[string]flds.PvcsItemAndRules{
				"data-volume": {
					Value: &flds.PvcsItemValue{
						Size: ptr.To("10Gi"),
					},
				},
			}
			pvcs := &flds.PvcsField{
				Value: &pvcsMap,
			}

			preview, err := pb.GetPreview(ctx, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).To(HaveOccurred())
			Expect(preview).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("context deadline exceeded"))

			result, err := pb.Build(ctx, jobType, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("context deadline exceeded"))
		})

		It("should handle PVCs with backward compatibility ownership", func() {
			mockScheme := runtime.NewScheme()
			Expect(corev1.AddToScheme(mockScheme)).To(Succeed())

			// Create a PVC that already exists but has labels matching the owner
			// (backward compatibility ownership)
			existingPvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pvc-test-workload-0",
					Namespace: defaultNamespace,
					Labels: map[string]string{
						constants.LabelRelease:   workloadBaseName, // Backward compatibility ownership
						constants.LabelCreatedBy: jobType,
					},
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.VolumeResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("10Gi"),
						},
					},
				},
			}

			mockClient := &MockClient{
				Client: fake.NewClientBuilder().WithScheme(mockScheme).WithObjects(existingPvc).Build(),
				MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
					if key.Name == "pvc-test-workload-0" && key.Namespace == defaultNamespace {
						pvc := obj.(*corev1.PersistentVolumeClaim)
						*pvc = *existingPvc
						return nil
					}
					return apierrors.NewNotFound(schema.GroupResource{Resource: "persistentvolumeclaims"}, key.Name)
				},
				scheme: mockScheme,
			}

			pb = NewPvcBuilder(mockClient)
			pvcsMap := map[string]flds.PvcsItemAndRules{
				"data-volume": {
					Value: &flds.PvcsItemValue{
						Size: ptr.To("10Gi"),
					},
				},
			}
			pvcs := &flds.PvcsField{
				Value: &pvcsMap,
			}

			// GetPreview should work with backward compatibility
			preview, err := pb.GetPreview(ctx, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).NotTo(HaveOccurred())
			Expect(preview).To(HaveLen(1))
			Expect(preview[0].Name).To(Equal("pvc-test-workload-0"))

			// Build should accept the existing PVC due to backward compatibility
			result, err := pb.Build(ctx, jobType, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(1))
			Expect(result[0].Name).To(Equal("pvc-test-workload-0"))
			Expect(result[0].Namespace).To(Equal(defaultNamespace))
		})

		It("should handle PVCs with controller ownership", func() {
			mockScheme := runtime.NewScheme()
			Expect(corev1.AddToScheme(mockScheme)).To(Succeed())

			// Create a PVC with controller ownership
			existingPvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pvc-test-workload-0",
					Namespace: defaultNamespace,
					Labels: map[string]string{
						constants.LabelCreatedBy: jobType,
					},
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: "v1",
							Kind:       "Pod",
							Name:       owner.Name,
							UID:        owner.UID,
							Controller: ptr.To(true),
						},
					},
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.VolumeResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("10Gi"),
						},
					},
				},
			}

			mockClient := &MockClient{
				Client: fake.NewClientBuilder().WithScheme(mockScheme).WithObjects(existingPvc).Build(),
				MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
					if key.Name == "pvc-test-workload-0" && key.Namespace == defaultNamespace {
						pvc := obj.(*corev1.PersistentVolumeClaim)
						*pvc = *existingPvc
						return nil
					}
					return apierrors.NewNotFound(schema.GroupResource{Resource: "persistentvolumeclaims"}, key.Name)
				},
				scheme: mockScheme,
			}

			pb = NewPvcBuilder(mockClient)
			pvcsMap := map[string]flds.PvcsItemAndRules{
				"data-volume": {
					Value: &flds.PvcsItemValue{
						Size: ptr.To("10Gi"),
					},
				},
			}
			pvcs := &flds.PvcsField{
				Value: &pvcsMap,
			}

			preview, err := pb.GetPreview(ctx, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).NotTo(HaveOccurred())
			Expect(preview).To(HaveLen(1))
			Expect(preview[0].Name).To(Equal("pvc-test-workload-0"))

			result, err := pb.Build(ctx, jobType, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(1))
			Expect(result[0].Name).To(Equal("pvc-test-workload-0"))
			Expect(result[0].Namespace).To(Equal(defaultNamespace))
		})

		It("should return error when PVC is owned by another workload", func() {
			mockScheme := runtime.NewScheme()
			Expect(corev1.AddToScheme(mockScheme)).To(Succeed())

			// Create a different owner for the existing PVC
			differentOwner := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "different-pod",
					Namespace: defaultNamespace,
					UID:       "different-uid",
				},
			}

			// Create a PVC with a different controller owner
			existingPvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pvc-test-workload-0",
					Namespace: defaultNamespace,
					Labels: map[string]string{
						constants.LabelCreatedBy: jobType,
					},
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: "v1",
							Kind:       "Pod",
							Name:       differentOwner.Name,
							UID:        differentOwner.UID,
							Controller: ptr.To(true),
						},
					},
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.VolumeResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("10Gi"),
						},
					},
				},
			}

			mockClient := &MockClient{
				Client: fake.NewClientBuilder().WithScheme(mockScheme).WithObjects(existingPvc).Build(),
				MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
					if key.Name == "pvc-test-workload-0" && key.Namespace == defaultNamespace {
						pvc := obj.(*corev1.PersistentVolumeClaim)
						*pvc = *existingPvc
						return nil
					}
					return apierrors.NewNotFound(schema.GroupResource{Resource: "persistentvolumeclaims"}, key.Name)
				},
				scheme: mockScheme,
			}

			pb = NewPvcBuilder(mockClient)
			pvcsMap := map[string]flds.PvcsItemAndRules{
				"data-volume": {
					Value: &flds.PvcsItemValue{
						Size: ptr.To("10Gi"),
					},
				},
			}
			pvcs := &flds.PvcsField{
				Value: &pvcsMap,
			}

			preview, err := pb.GetPreview(ctx, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).To(HaveOccurred())
			Expect(preview).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("is owned by another workload"))

			// Build should return error because the existing PVC has a different owner
			result, err := pb.Build(ctx, jobType, objectMeta, workloadBaseName, owner, pvcs, nameSuffix)
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("is owned by another workload"))
		})

		It("should return error when owner is nil but pvc exists", func() {
			mockScheme := runtime.NewScheme()
			Expect(corev1.AddToScheme(mockScheme)).To(Succeed())

			// Create a PVC that already exists
			existingPvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pvc-test-workload-0",
					Namespace: defaultNamespace,
					Labels: map[string]string{
						constants.LabelCreatedBy: jobType,
					},
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.VolumeResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("10Gi"),
						},
					},
				},
			}

			mockClient := &MockClient{
				Client: fake.NewClientBuilder().WithScheme(mockScheme).WithObjects(existingPvc).Build(),
				MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
					if key.Name == "pvc-test-workload-0" && key.Namespace == defaultNamespace {
						pvc := obj.(*corev1.PersistentVolumeClaim)
						*pvc = *existingPvc
						return nil
					}
					return apierrors.NewNotFound(schema.GroupResource{Resource: "persistentvolumeclaims"}, key.Name)
				},
				scheme: mockScheme,
			}

			pb = NewPvcBuilder(mockClient)
			pvcsMap := map[string]flds.PvcsItemAndRules{
				"data-volume": {
					Value: &flds.PvcsItemValue{
						Size: ptr.To("10Gi"),
					},
				},
			}
			pvcs := &flds.PvcsField{
				Value: &pvcsMap,
			}
			var nilOwner client.Object = nil // Explicitly set owner to nil

			preview, err := pb.GetPreview(ctx, objectMeta, workloadBaseName, nilOwner, pvcs, nameSuffix)
			Expect(err).To(HaveOccurred())
			Expect(preview).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("pvc owner is nil"))

			// Build should return an error since we can't set owner reference when owner is nil
			result, err := pb.Build(ctx, jobType, objectMeta, workloadBaseName, nilOwner, pvcs, nameSuffix)
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("pvc owner is nil"))
		})
	})
})
