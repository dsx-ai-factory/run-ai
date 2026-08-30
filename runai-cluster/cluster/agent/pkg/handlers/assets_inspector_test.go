package handlers

// calling the import here to ensure that the helper init takes place before anything else
// see inside the helper package for more info
import (
	_ "github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/tests/helper"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/testutils"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("AssetsTest", func() {
	var assetsInspector AssetsObjectInspector

	BeforeEach(func() {
	})

	Describe("AssetsInspectorTest", func() {
		Context("Deep copy PVC", func() {
			It("should copy spec + annotations", func() {
				from := &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-pvc",
						Namespace: "my-namespace",
						Annotations: map[string]string{
							UpdateCountAnnotation: "123",
						},
						Labels: map[string]string{
							constants.RunaiAssetLabel: "pvc",
						},
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						StorageClassName: convutils.Ref("my-storage-class"),
					},
				}
				to := &corev1.PersistentVolumeClaim{
					TypeMeta: metav1.TypeMeta{
						Kind: constants.KindPvc,
					},
				}

				assetsInspector.DeepCopySpec(from, to)

				//
				//  the annotations and spec are copied
				//
				expected := &corev1.PersistentVolumeClaim{
					TypeMeta: metav1.TypeMeta{
						Kind: constants.KindPvc,
					},
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							UpdateCountAnnotation: "123",
						},
						Labels: map[string]string{
							"run.ai/asset": "pvc",
						},
					},
					Spec: from.Spec,
				}

				testutils.Expect(to).ToEqual(expected)
			})
		})

		Context("Deep copy Secret", func() {
			It("should copy spec + annotations", func() {
				from := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-secret",
						Namespace: "my-namespace",
						Annotations: map[string]string{
							UpdateCountAnnotation: "456",
						},
						Labels: map[string]string{
							constants.RunaiAssetLabel: "secret",
						},
					},
					Data: map[string][]byte{
						"aaa": []byte("bbb"),
						"ccc": []byte("ddd"),
					},
				}
				to := &corev1.Secret{
					TypeMeta: metav1.TypeMeta{
						Kind: constants.KindSecret,
					},
				}

				assetsInspector.DeepCopySpec(from, to)

				//
				//  the annotations and data are copied
				//
				expected := &corev1.Secret{
					TypeMeta: metav1.TypeMeta{
						Kind: constants.KindSecret,
					},
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							UpdateCountAnnotation: "456",
						},
						Labels: map[string]string{
							"run.ai/asset": "secret",
						},
					},
					Data: from.Data,
				}

				testutils.Expect(to).ToEqual(expected)
			})
		})

		Context("IsUptodate", func() {
			update123 := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-pvc",
					Namespace: "my-namespace",
					Annotations: map[string]string{
						UpdateCountAnnotation: "123",
					},
				},
			}

			update124 := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-pvc",
					Namespace: "my-namespace",
					Annotations: map[string]string{
						UpdateCountAnnotation: "124",
					},
				},
			}

			unknownUpdate := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-pvc",
					Namespace: "my-namespace",
				},
			}

			result := assetsInspector.IsUpToDate(client.Object(update124), client.Object(update123))
			Expect(result).To(Equal(false), "124 VS 123")

			result = assetsInspector.IsUpToDate(client.Object(update123), client.Object(update124))
			Expect(result).To(Equal(true), "123 VS 124")

			result = assetsInspector.IsUpToDate(client.Object(update123), client.Object(update123))
			Expect(result).To(Equal(true), "123 VS 123")

			result = assetsInspector.IsUpToDate(client.Object(unknownUpdate), client.Object(update123))
			Expect(result).To(Equal(false), "123 VS unknown")

			result = assetsInspector.IsUpToDate(client.Object(update123), client.Object(unknownUpdate))
			Expect(result).To(Equal(false), "unknown VS 123")

		})
	})
})
