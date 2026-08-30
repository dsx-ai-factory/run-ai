package cluster_api

import (
	. "github.com/onsi/ginkgo/v2"
	"k8s.io/utils/ptr"

	. "github.com/onsi/gomega"
)

var _ = Describe("cluster api", func() {
	var (
		clusterApi *ClusterApi
	)

	BeforeEach(func() {
		clusterApi = &ClusterApi{
			Enabled: ptr.To[bool](true),
		}
	})

	It("nothing is set Replicas should be 1", func() {
		clusterApi.SetDefaultsWhereNeeded("", ptr.To[int32](1))
		Expect(*clusterApi.Replicas).To(Equal(int32(1)))
	})

	It("global.replicaCount is nil Replicas should be 1", func() {
		clusterApi.SetDefaultsWhereNeeded("", nil)
		Expect(*clusterApi.Replicas).To(Equal(int32(1)))
	})

	It("global.replicaCount is set, cluster.replicas is not set be equal to global.replicaCount", func() {
		clusterApi.SetDefaultsWhereNeeded("", ptr.To[int32](2))
		Expect(*clusterApi.Replicas).To(Equal(int32(2)))
	})

	It("global.replicaCount is set, cluster.replicas is set should bot change", func() {
		clusterApi.Replicas = ptr.To[int32](2)
		clusterApi.SetDefaultsWhereNeeded("", ptr.To[int32](3))
		Expect(*clusterApi.Replicas).To(Equal(int32(2)))
	})

})
