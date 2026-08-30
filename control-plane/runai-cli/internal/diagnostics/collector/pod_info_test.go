package collector

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("PodInfo", func() {
	ownerName := "runai-backend-workloads-b8f4ff97d"
	namespace, containerName, podName, nodeName :=
		"runai-backend",
		"nginx-container",
		"runai-backend-workloads-b8f4ff97d-49l72",
		"control-plane-1"

	var pod corev1.Pod

	Describe("PodLogFilePath", func() {
		BeforeEach(func() {
			pod = corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      podName,
					Namespace: namespace,
					OwnerReferences: []metav1.OwnerReference{{
						Name: ownerName, Kind: "DaemonSet",
					}},
				},
				Spec: corev1.PodSpec{
					NodeName: nodeName, Containers: []corev1.Container{{
						Name: containerName,
					}}},
			}
		})

		It("should return daemon set log file path", func() {
			podLogFile := fmt.Sprintf("%s/daemon-set_%s/%s/%s_%s.log", namespace, ownerName, pod.Name, nodeName, containerName)
			f := NewPodInfo(&pod).getLogFilePathInArchive(containerName, false)
			Expect(f).To(Equal(podLogFile))
			cn := "container-name-1"
			podLogFile = fmt.Sprintf("%s/daemon-set_%s/%s/previous_%s_%s.log", namespace, ownerName, pod.Name, nodeName, cn)
			f = NewPodInfo(&pod).getLogFilePathInArchive(cn, true)
			Expect(f).To(Equal(podLogFile))
		})

		It("should return deployment log file path", func() {
			pod.Name = "deployment-name-7c8bd58998-tkdzb"
			pod.ObjectMeta.OwnerReferences[0].Kind = string(ReplicaSet)
			pod.ObjectMeta.OwnerReferences[0].Name = "deployment-name-7c8bd58998"
			podLogFile := fmt.Sprintf("%s/deployment_deployment-name/%s/%s.log", namespace, pod.Name, containerName)
			f := NewPodInfo(&pod).getLogFilePathInArchive(containerName, false)
			Expect(f).To(Equal(podLogFile))
			cn := "container-name-2"
			podLogFile = fmt.Sprintf("%s/deployment_deployment-name/%s/previous_%s.log", namespace, pod.Name, cn)
			f = NewPodInfo(&pod).getLogFilePathInArchive(cn, true)
			Expect(f).To(Equal(podLogFile))
		})

		It("should return stateful-set log file path", func() {
			pod.Name = podName
			pod.ObjectMeta.OwnerReferences[0].Kind = "StatefulSet"
			pod.ObjectMeta.OwnerReferences[0].Name = "abcd"
			podLogFile := fmt.Sprintf("%s/stateful-set_%s/%s/%s.log", namespace, pod.ObjectMeta.OwnerReferences[0].Name, pod.Name, containerName)
			f := NewPodInfo(&pod).getLogFilePathInArchive(containerName, false)
			Expect(f).To(Equal(podLogFile))
			cn := "container-name-3"
			podLogFile = fmt.Sprintf("%s/stateful-set_%s/%s/previous_%s.log", namespace, pod.ObjectMeta.OwnerReferences[0].Name, pod.Name, cn)
			f = NewPodInfo(&pod).getLogFilePathInArchive(cn, true)
			Expect(f).To(Equal(podLogFile))
		})

		It("should return job log file path", func() {
			pod.ObjectMeta.OwnerReferences[0].Kind = string(Job)
			pod.ObjectMeta.OwnerReferences[0].Name = fmt.Sprintf("abc-job-%d", time.Now().Second())
			podLogFile := fmt.Sprintf("%s/job_%s/%s/%s.log", namespace, pod.ObjectMeta.OwnerReferences[0].Name, pod.Name, containerName)
			f := NewPodInfo(&pod).getLogFilePathInArchive(containerName, false)
			Expect(f).To(Equal(podLogFile))
			cn := "container-name-4"
			podLogFile = fmt.Sprintf("%s/job_%s/%s/previous_%s.log", namespace, pod.ObjectMeta.OwnerReferences[0].Name, pod.Name, cn)
			f = NewPodInfo(&pod).getLogFilePathInArchive(cn, true)
			Expect(f).To(Equal(podLogFile))
		})

		It("should return cron-job log file path", func() {
			pod.ObjectMeta.OwnerReferences[0].Kind = string(Job)
			pod.ObjectMeta.OwnerReferences[0].Name = fmt.Sprintf("scheduled-job-%d", int(time.Now().UnixMilli()/1000/60))
			podLogFile := fmt.Sprintf("%s/cron-job_scheduled-job/%s/%s.log", namespace, pod.Name, containerName)
			f := NewPodInfo(&pod).getLogFilePathInArchive(containerName, false)
			Expect(f).To(Equal(podLogFile))
			cn := "container-name-5"
			podLogFile = fmt.Sprintf("%s/cron-job_scheduled-job/%s/previous_%s.log", namespace, pod.Name, cn)
			f = NewPodInfo(&pod).getLogFilePathInArchive(cn, true)
			Expect(f).To(Equal(podLogFile))
		})

		It("unknown kind", func() {
			p := pod
			p.ObjectMeta.OwnerReferences = nil
			podLogFile := fmt.Sprintf("%s/%s/%s.log", namespace, p.Name, containerName)
			f := NewPodInfo(&p).getLogFilePathInArchive(containerName, false)
			Expect(f).To(Equal(podLogFile))
			cn := "container-name-6"
			podLogFile = fmt.Sprintf("%s/%s/previous_%s.log", namespace, p.Name, cn)
			f = NewPodInfo(&p).getLogFilePathInArchive(cn, true)
			Expect(f).To(Equal(podLogFile))
		})
	})

	Context("guessReplicaSetOrigin", func() {
		It("should return deployment kind and name", func() {
			owner := pod.OwnerReferences[0]
			owner.Name = "runai-backend-traefik-7fd87df648"
			owner.Kind = string(ReplicaSet)

			kind, name := guessReplicaSetOrigin(&owner)
			Expect(kind).To(Equal(Deployment))
			Expect(name).To(Equal("runai-backend-traefik"))
		})
		It("should return replica set kind and name", func() {
			owner := pod.OwnerReferences[0]
			owner.Name = "runai-backend-traefik"
			owner.Kind = string(ReplicaSet)

			kind, name := guessReplicaSetOrigin(&owner)
			Expect(kind).To(Equal(ReplicaSet))
			Expect(name).To(Equal("runai-backend-traefik"))
		})
	})

	Context("guessJobOrigin", func() {
		It("should return cron job kind and name", func() {
			owner := pod.OwnerReferences[0]
			owner.Name = "cron-job-29291040"
			owner.Kind = string(Job)

			kind, name := guessJobOrigin(&owner)
			Expect(kind).To(Equal(CronJob))
			Expect(name).To(Equal("cron-job"))
		})
		It("should return job kind and name", func() {
			owner := pod.OwnerReferences[0]
			owner.Name = "regular-job-12345"
			owner.Kind = string(Job)

			kind, name := guessJobOrigin(&owner)
			Expect(kind).To(Equal(Job))
			Expect(name).To(Equal("regular-job-12345"))
		})
	})
})
