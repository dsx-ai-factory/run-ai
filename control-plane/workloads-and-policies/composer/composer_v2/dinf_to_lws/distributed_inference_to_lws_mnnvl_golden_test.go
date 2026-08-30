package dinf_to_lws

import (
	"context"
	"time"

	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	clusterApi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/testutils"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/stretchr/testify/assert"
)

// mnnvlGoldenBaseInput is a minimal valid distributed-inference spec reused by
// every MNNVL case so the goldens differ only in the resolved mnnvl plugin mode.
const mnnvlGoldenBaseInput = "dinfw_serving_port_basic"

func mnnvlTestClusterConfig(dinfEnabled bool) composer.ClusterStorageConfig {
	cluster := clusterApi.NewDisplayedCluster("cluster-a", 1, "cluster-a", time.Time{})
	cluster.SetDomain("https://test.run.ai")
	reported := clusterApi.ClusterReportedConfig{}
	reported.SetDinfGPUNetworkAcceleration(dinfEnabled)
	status := clusterApi.ClusterDisplayedStatus{}
	status.SetConfig(reported)
	cluster.Status.Set(&status)
	return composer.NewClusterStorageConfigFromCluster(cluster)
}

var _ = Describe("distributed inference MNNVL", func() {
	// The MNNVL plugin flag is resolved from the leader's run.ai/MNNVL label when
	// present, otherwise defaulted from the cluster's dinfGPUNetworkAcceleration flag.
	DescribeTable("resolves the MNNVL plugin mode from labels and cluster config",
		func(expectedFname string, labelValue string, dinfEnabled bool) {
			input, err := testutils.LoadInputFile(mnnvlGoldenBaseInput, ".")
			Expect(err).To(BeNil())

			spec := workloadsApi.NewDistributedInferenceSpecSpec()
			Expect(yaml.Unmarshal(input, spec)).To(Succeed())

			if labelValue != "" {
				leader := spec.Leader.Get()
				Expect(leader).NotTo(BeNil())
				leader.Labels = append(leader.Labels, workloadsApi.Label{
					Name:  *workloadsApi.NewNullableString(convutils.Ref("run.ai/MNNVL")),
					Value: *workloadsApi.NewNullableString(convutils.Ref(labelValue)),
				})
			}

			dinfComposer := NewDistributedInferenceToLWSComposer(
				"test-lws",
				"team-a",
				"runai-team-a",
				"tester",
				nil,
				nil,
				composer.FieldsComposerOptions{
					EmptyDirAsVolumeSupported:        true,
					FlexiblePriorityClassesSupported: true,
					NodeAffinitySupported:            true,
					PreemptibilitySupported:          true,
					UserCredentialsSupported:         true,
					VolumeDefaultModeSupported:       true,
					WorkloadV2NetworkingSupported:    true,
				},
				legacyTypedTestNodePools(),
			).WithWorkloadID("00000000-0000-0000-0000-000000001234").WithClusterStorageConfig(mnnvlTestClusterConfig(dinfEnabled))

			target := NewMinimalLeaderWorkerSetTarget()
			composeErr := dinfComposer.Compose(context.Background(), spec, target)()
			Expect(composeErr).To(BeNil())

			normalizeGeneratedPVCNames(target)
			actualYaml, marshalErr := yaml.Marshal(target)
			Expect(marshalErr).To(BeNil())

			expectedYaml := testutils.LoadExpectedFile(expectedFname, actualYaml, ".")
			assert.Equal(GinkgoT(), string(expectedYaml), string(actualYaml), expectedFname)
		},
		Entry("cluster acceleration on defaults to Required", "mnnvl_cluster_on_default_required", "", true),
		Entry("explicit None label overrides cluster on", "mnnvl_label_none_overrides_cluster_on", "None", true),
	)
})
