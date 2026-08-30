package common

import (
	"context"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"os"

	"math/rand"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("ingress builder utils", func() {
	var (
		ctx        context.Context
		nodeList   corev1.NodeList
		k8sClient  client.Client
		serviceObj []client.Object
	)

	BeforeEach(func() {
		rand.Seed(time.Now().UnixNano())
		ctx = context.Background()
		nodeList = corev1.NodeList{
			Items: []corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "node4"},
					Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{
							{Type: corev1.NodeInternalIP, Address: "192.168.0.4"},
							{Type: corev1.NodeInternalIP, Address: "10.0.0.4"},
							{Type: corev1.NodeExternalIP, Address: "203.0.113.4"},
							{Type: corev1.NodeHostName, Address: "node4.example.com"},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "node2"},
					Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{
							{Type: corev1.NodeInternalIP, Address: "192.168.0.2"},
							{Type: corev1.NodeInternalIP, Address: "10.0.0.2"},
							{Type: corev1.NodeExternalIP, Address: "203.0.113.2"},
							{Type: corev1.NodeHostName, Address: "node2.example.com"},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "node3"},
					Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{
							{Type: corev1.NodeInternalIP, Address: "192.168.0.3"},
							{Type: corev1.NodeInternalIP, Address: "10.0.0.3"},
							{Type: corev1.NodeExternalIP, Address: "203.0.113.3"},
							{Type: corev1.NodeHostName, Address: "node3.example.com"},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "node1"},
					Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{
							{Type: corev1.NodeInternalIP, Address: "192.168.0.1"},
							{Type: corev1.NodeInternalIP, Address: "10.0.0.1"},
							{Type: corev1.NodeExternalIP, Address: "203.0.113.1"},
							{Type: corev1.NodeHostName, Address: "node1.example.com"},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "node5"},
					Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{
							{Type: corev1.NodeInternalIP, Address: "192.168.0.5"},
							{Type: corev1.NodeInternalIP, Address: "10.0.0.5"},
							{Type: corev1.NodeExternalIP, Address: "203.0.113.5"},
							{Type: corev1.NodeHostName, Address: "node5.example.com"},
						},
					},
				},
			},
		}
		serviceObj = []client.Object{
			&corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "service1",
					Namespace: "default",
				},
				Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeNodePort, Ports: []corev1.ServicePort{{NodePort: 30000}, {NodePort: 30001}}},
			},
			&corev1.Service{Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeClusterIP}},
		}
		k8sClient = fake.NewClientBuilder().WithObjects(serviceObj...).
			WithLists(&nodeList).WithStatusSubresource(&corev1.Node{}).Build()

	})
	Context("getNodePorts", func() {
		It("should return a sorted list of node ports", func() {
			result := getNodePorts(serviceObj)
			Expect(result).To(HaveLen(2))
			Expect(result).To(Equal([]int32{30000, 30001}))
		})
	})
	Context("getNodeInternalIP", func() {
		Context("when there is no internal IP", func() {
			It("should return an empty string", func() {
				result := getNodeInternalIP(&corev1.Node{})
				Expect(result).To(Equal(""))
			})
		})
		Context("when there is an internal IP", func() {
			It("should return the internal IP", func() {
				node := nodeList.Items[0]
				result := getNodeInternalIP(&node)
				Expect(result).ToNot(Equal(""))
				for i := 0; i < 5; i++ {
					rand.Shuffle(len(node.Status.Addresses), func(i, j int) {
						node.Status.Addresses[i], node.Status.Addresses[j] = node.Status.Addresses[j], node.Status.Addresses[i]
					})
					Expect(result).To(Equal(getNodeInternalIP(&node)))
				}
			})
		})
	})
	Context("UpdateJobUrlWithNodePorts", func() {
		var (
			objectMeta metav1.ObjectMeta
		)
		BeforeEach(func() {
			objectMeta = metav1.ObjectMeta{
				Annotations: map[string]string{},
			}
		})
		Context("when there is no node-port service", func() {
			It("should not update the URL", func() {
				err := UpdateJobUrlWithNodePorts(ctx, []client.Object{}, k8sClient, &objectMeta)
				Expect(err).NotTo(HaveOccurred())
				Expect(objectMeta.Annotations).To(BeEmpty())
			})
		})
		Context("when there is a node-port service", func() {

			Context("when workload controller env is set", func() {
				BeforeEach(func() {
					Expect(os.Setenv(constants.WorkloadControllerNodeNameEnvKey, "node1")).To(Succeed())
				})
				AfterEach(func() {
					Expect(os.Unsetenv(constants.WorkloadControllerNodeNameEnvKey)).To(Succeed())
				})
				It("should update the URL", func() {

					err := UpdateJobUrlWithNodePorts(ctx, serviceObj, k8sClient, &objectMeta)
					Expect(err).NotTo(HaveOccurred())
					jobURl := objectMeta.Annotations[constants.AnnotationJobUrl]
					Expect(jobURl).To(Equal("10.0.0.1:30000, 10.0.0.1:30001"))
				})
			})
			Context("when workload controller env is not set", func() {
				BeforeEach(func() {
					// unset the env
					_ = os.Unsetenv(constants.WorkloadControllerNodeNameEnvKey)
				})
				It("should not update the URL", func() {
					err := UpdateJobUrlWithNodePorts(ctx, serviceObj, k8sClient, &objectMeta)
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})
})
