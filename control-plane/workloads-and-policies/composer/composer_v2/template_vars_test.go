package composer_v2

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("ExpandTemplateVarsInCommandAndArgs", func() {
	It("expands project and job name in command and args", func() {
		container := &corev1.Container{
			Command: []string{"start.sh", "--project=${RUNAI_PROJECT}"},
			Args:    []string{"--base-url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME}", "--job=${RUNAI_JOB_NAME}"},
		}

		Expect(ExpandTemplateVarsInCommandAndArgs(container, "team-a", "job-3")).To(BeNil())

		Expect(container.Command).To(Equal([]string{"start.sh", "--project=team-a"}))
		Expect(container.Args).To(Equal([]string{"--base-url=/team-a/job-3", "--job=job-3"}))
	})

	It("expands uid and gid from the container security context", func() {
		container := &corev1.Container{
			Args: []string{"--uid=${RUNAI_UID}", "--gid=${RUNAI_GID}"},
			SecurityContext: &corev1.SecurityContext{
				RunAsUser:  convutils.Ref(int64(1000)),
				RunAsGroup: convutils.Ref(int64(2000)),
			},
		}

		Expect(ExpandTemplateVarsInCommandAndArgs(container, "team-a", "job-3")).To(BeNil())

		Expect(container.Args).To(Equal([]string{"--uid=1000", "--gid=2000"}))
	})

	It("leaves uid and gid placeholders untouched when the security context is nil", func() {
		container := &corev1.Container{
			Args: []string{"--uid=${RUNAI_UID}", "--gid=${RUNAI_GID}"},
		}

		Expect(ExpandTemplateVarsInCommandAndArgs(container, "team-a", "job-3")).To(BeNil())

		Expect(container.Args).To(Equal([]string{"--uid=${RUNAI_UID}", "--gid=${RUNAI_GID}"}))
	})

	It("is a no-op when command and args are empty", func() {
		container := &corev1.Container{}

		Expect(ExpandTemplateVarsInCommandAndArgs(container, "team-a", "job-3")).To(BeNil())

		Expect(container.Command).To(BeEmpty())
		Expect(container.Args).To(BeEmpty())
	})
})
