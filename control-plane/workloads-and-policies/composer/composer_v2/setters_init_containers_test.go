package composer_v2

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("MapS3ToWorkloadK8s", func() {
	var podSpec *corev1.PodSpec

	BeforeEach(func() {
		podSpec = &corev1.PodSpec{
			Containers: []corev1.Container{{Name: "main"}},
		}
	})

	It("should be a no-op for nil info", func() {
		err := MapS3ToWorkloadK8s(nil, podSpec, nil)
		Expect(err).To(BeNil())
		Expect(podSpec.InitContainers).To(BeEmpty())
		Expect(podSpec.Volumes).To(BeEmpty())
		Expect(podSpec.Containers[0].VolumeMounts).To(BeEmpty())
	})

	It("should be a no-op for info with no items", func() {
		err := MapS3ToWorkloadK8s(&composer.S3Info{Config: composer.DefaultClusterStorageConfig()}, podSpec, nil)
		Expect(err).To(BeNil())
		Expect(podSpec.InitContainers).To(BeEmpty())
	})

	It("should reject a target that is not a PodSpec pointer", func() {
		info := &composer.S3Info{
			Config: composer.DefaultClusterStorageConfig(),
			Items:  map[string]*composer.S3Item{"s3-0": {Bucket: "b", Path: "/p"}},
		}

		err := MapS3ToWorkloadK8s(info, &struct{}{}, nil)

		Expect(err).NotTo(BeNil())
		Expect(err.Code).To(Equal(int32(http.StatusInternalServerError)))
	})

	It("should reject a PodSpec that has no main container", func() {
		info := &composer.S3Info{
			Config: composer.DefaultClusterStorageConfig(),
			Items:  map[string]*composer.S3Item{"s3-0": {Bucket: "b", Path: "/p"}},
		}

		err := MapS3ToWorkloadK8s(info, &corev1.PodSpec{}, nil)

		Expect(err).NotTo(BeNil())
		Expect(err.Code).To(Equal(int32(http.StatusInternalServerError)))
	})

	It("should append a privileged goofys init container, the emptyDir volume, and the main-container mount", func() {
		info := &composer.S3Info{
			Config: composer.DefaultClusterStorageConfig(),
			Items:  map[string]*composer.S3Item{"s3-0": {Bucket: "my-bucket", Path: "/mnt/s3"}},
		}

		err := MapS3ToWorkloadK8s(info, podSpec, nil)

		Expect(err).To(BeNil())

		Expect(podSpec.InitContainers).To(HaveLen(1))
		Expect(podSpec.InitContainers[0].Name).To(Equal("s3-fs-0"))
		Expect(podSpec.InitContainers[0].Image).To(Equal("nvcr.io/nvidia/runai/goofys:2.26.30"))
		Expect(podSpec.InitContainers[0].ImagePullPolicy).To(Equal(corev1.PullIfNotPresent))
		Expect(podSpec.InitContainers[0].SecurityContext).NotTo(BeNil())
		Expect(*podSpec.InitContainers[0].SecurityContext.Privileged).To(BeTrue())
		Expect(*podSpec.InitContainers[0].RestartPolicy).To(Equal(corev1.ContainerRestartPolicyAlways))

		Expect(podSpec.Volumes).To(HaveLen(1))
		Expect(podSpec.Volumes[0].Name).To(Equal("s3-0"))
		Expect(podSpec.Volumes[0].EmptyDir).NotTo(BeNil())

		Expect(podSpec.Containers[0].VolumeMounts).To(HaveLen(1))
		Expect(podSpec.Containers[0].VolumeMounts[0].Name).To(Equal("s3-0"))
		Expect(podSpec.Containers[0].VolumeMounts[0].MountPath).To(Equal("/mnt/s3"))
		Expect(*podSpec.Containers[0].VolumeMounts[0].MountPropagation).To(Equal(corev1.MountPropagationHostToContainer))
	})
})
