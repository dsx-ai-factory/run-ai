package controller_utils

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("specField", func() {
	It("returns Spec for types that have one", func() {
		dep := &appsv1.Deployment{
			Spec: appsv1.DeploymentSpec{
				MinReadySeconds: 5,
			},
		}
		spec, ok := specField(dep)
		Expect(ok).To(BeTrue())
		Expect(spec).To(Equal(dep.Spec))
	})

	It("returns (nil, false) for types without a Spec field", func() {
		cm := &corev1.ConfigMap{Data: map[string]string{"k": "v"}}
		_, ok := specField(cm)
		Expect(ok).To(BeFalse())

		sec := &corev1.Secret{Data: map[string][]byte{"k": []byte("v")}}
		_, ok = specField(sec)
		Expect(ok).To(BeFalse())

		role := &rbacv1.Role{Rules: []rbacv1.PolicyRule{{Verbs: []string{"get"}}}}
		_, ok = specField(role)
		Expect(ok).To(BeFalse())
	})
})

var _ = Describe("specsDiffer", func() {
	Context("with types that have a Spec field", func() {
		It("returns false when Spec is identical even if metadata differs", func() {
			// This is the engine-config repro scenario: server-side churn on
			// resourceVersion/labels/annotations should not be treated as a
			// reason to issue an Update.
			existing := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "x",
					Namespace:       "runai",
					ResourceVersion: "1000",
					Labels:          map[string]string{"app": "x", "extra": "from-server"},
					Annotations:     map[string]string{"deployment.kubernetes.io/revision": "5"},
				},
				Spec:   appsv1.DeploymentSpec{MinReadySeconds: 5},
				Status: appsv1.DeploymentStatus{ReadyReplicas: 3},
			}
			desired := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "x",
					Namespace: "runai",
					Labels:    map[string]string{"app": "x"},
				},
				Spec: appsv1.DeploymentSpec{MinReadySeconds: 5},
			}
			Expect(specsDiffer(existing, desired)).To(BeFalse())
		})

		It("returns true when Spec differs", func() {
			existing := &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{MinReadySeconds: 5},
			}
			desired := &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{MinReadySeconds: 10},
			}
			Expect(specsDiffer(existing, desired)).To(BeTrue())
		})

		It("returns false when Quantity values use different equivalent representations", func() {
			// This is the engine-config root-cause repro: resource.Quantity has
			// multiple internal representations of the same value ("100Mi" via
			// BinarySI vs "104857600" via DecimalSI) that reflect.DeepEqual
			// considers different but Quantity.Cmp() considers equal.
			// equality.Semantic.DeepEqual honors Quantity.Equal(); a regression
			// to bare reflect.DeepEqual in specsDiffer would fail this test.
			existing := &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{
								Name: "x",
								Resources: corev1.ResourceRequirements{
									Limits: corev1.ResourceList{
										corev1.ResourceMemory: resource.MustParse("100Mi"),
									},
								},
							}},
						},
					},
				},
			}
			desired := &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{
								Name: "x",
								Resources: corev1.ResourceRequirements{
									Limits: corev1.ResourceList{
										corev1.ResourceMemory: resource.MustParse("104857600"),
									},
								},
							}},
						},
					},
				},
			}
			Expect(specsDiffer(existing, desired)).To(BeFalse())
		})

		It("returns false for two Services with identical Spec but different ClusterIP-derived state", func() {
			existing := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{ResourceVersion: "2"},
				Spec: corev1.ServiceSpec{
					Type:  corev1.ServiceTypeClusterIP,
					Ports: []corev1.ServicePort{{Port: 80}},
				},
			}
			desired := &corev1.Service{
				Spec: corev1.ServiceSpec{
					Type:  corev1.ServiceTypeClusterIP,
					Ports: []corev1.ServicePort{{Port: 80}},
				},
			}
			Expect(specsDiffer(existing, desired)).To(BeFalse())
		})
	})

	Context("with types that do NOT have a Spec field", func() {
		It("falls back to whole-object DeepEqual", func() {
			// ConfigMap has no Spec, so any metadata difference falls through
			// to whole-object comparison. This is the documented fallback for
			// resource kinds without a fast concurrent writer.
			existing := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "c", ResourceVersion: "1"},
				Data:       map[string]string{"k": "v"},
			}
			desired := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "c"},
				Data:       map[string]string{"k": "v"},
			}
			// resourceVersion differs → whole-object compare returns differ=true.
			Expect(specsDiffer(existing, desired)).To(BeTrue())
		})

		It("returns false for identical Data with identical metadata", func() {
			existing := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "c"},
				Data:       map[string]string{"k": "v"},
			}
			desired := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "c"},
				Data:       map[string]string{"k": "v"},
			}
			Expect(specsDiffer(existing, desired)).To(BeFalse())
		})

		It("returns true for ConfigMaps with different Data", func() {
			existing := &corev1.ConfigMap{Data: map[string]string{"k": "v1"}}
			desired := &corev1.ConfigMap{Data: map[string]string{"k": "v2"}}
			Expect(specsDiffer(existing, desired)).To(BeTrue())
		})
	})
})
