package upgrade

import (
	"encoding/json"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func writeCache(path string, cache Cache, perm os.FileMode) {
	cacheBytes, err := json.Marshal(cache)
	Expect(err).ToNot(HaveOccurred())
	Expect(os.WriteFile(path, cacheBytes, perm)).To(Succeed())
}

func corruptCache(path string) {
	Expect(os.WriteFile(path, []byte("not-json"), 0600)).To(Succeed())
}

var _ = Describe("FlowCondition", func() {
	var tmpFile string
	var logger *zap.Logger

	BeforeEach(func() {
		f, err := os.CreateTemp("", "throttle_cache_*.json")
		Expect(err).ToNot(HaveOccurred())
		tmpFile = f.Name()
		f.Close()
		logger = zaptest.NewLogger(GinkgoT())
	})

	AfterEach(func() {
		os.Remove(tmpFile)
	})

	DescribeTable("IsRunningUpdateFlowRequired",
		func(setup func(string), ttl time.Duration, cpURL string, outputFormat string, autoUpdateEnabled bool, expectCheck bool) {
			if setup != nil {
				setup(tmpFile)
			}
			throttler := NewFlowCondition(tmpFile, ttl, cpURL, logger)
			Expect(throttler.IsRunningUpdateFlowRequired(outputFormat, autoUpdateEnabled)).To(Equal(expectCheck))
		},
		Entry("no Cache file", nil, time.Hour, "https://cp.example.com", "table", true, true),
		Entry("valid Cache, interval not passed, CP URL unchanged",
			func(path string) {
				writeCache(path, Cache{Timestamp: time.Now(), ControlPlaneURL: "https://cp.example.com"}, 0600)
			},
			time.Hour, "https://cp.example.com", "table", true, false),
		Entry("valid Cache, interval passed",
			func(path string) {
				writeCache(path, Cache{Timestamp: time.Now().Add(-2 * time.Hour), ControlPlaneURL: "https://cp.example.com"}, 0600)
			},
			time.Hour, "https://cp.example.com", "table", true, true),
		Entry("valid Cache, CP URL changed",
			func(path string) {
				writeCache(path, Cache{Timestamp: time.Now(), ControlPlaneURL: "https://old-cp.example.com"}, 0600)
			},
			time.Hour, "https://cp.example.com", "table", true, true),
		Entry("corrupted Cache",
			func(path string) { corruptCache(path) },
			time.Hour, "https://cp.example.com", "table", true, true),
		Entry("permission denied (read)",
			func(path string) {
				writeCache(path, Cache{Timestamp: time.Now(), ControlPlaneURL: "https://cp.example.com"}, 0600)
				os.Chmod(path, 0200)
			},
			time.Hour, "https://cp.example.com", "table", true, true),
		Entry("empty cachePath", nil, time.Hour, "https://cp.example.com", "table", true, true),
		Entry("empty currentCPURL", nil, time.Hour, "", "table", true, false),
		Entry("json output, autoUpdate disabled", nil, time.Hour, "https://cp.example.com", "json", false, false),
		Entry("yaml output, autoUpdate disabled", nil, time.Hour, "https://cp.example.com", "yaml", false, false),
		Entry("json output, autoUpdate enabled", nil, time.Hour, "https://cp.example.com", "json", true, true),
	)

	Describe("UpdateSuccessAPITimestamp", func() {
		It("should update the Cache with the current time and CP URL", func() {
			throttler := NewFlowCondition(tmpFile, time.Hour, "https://cp.example.com", logger)
			err := throttler.UpdateSuccessAPITimestamp()
			Expect(err).ToNot(HaveOccurred())
			cacheBytes, err := os.ReadFile(tmpFile)
			Expect(err).ToNot(HaveOccurred())
			var newCache Cache
			Expect(json.Unmarshal(cacheBytes, &newCache)).To(Succeed())
			Expect(newCache.ControlPlaneURL).To(Equal("https://cp.example.com"))
			Expect(time.Since(newCache.Timestamp)).To(BeNumerically("<", time.Minute))
		})

		It("should return error if cachePath is empty", func() {
			throttler := NewFlowCondition("", time.Hour, "https://cp.example.com", logger)
			err := throttler.UpdateSuccessAPITimestamp()
			Expect(err).To(HaveOccurred())
		})

		It("should return error if file cannot be written (permission denied)", func() {
			writeCache(tmpFile, Cache{Timestamp: time.Now(), ControlPlaneURL: "https://cp.example.com"}, 0400)
			os.Chmod(tmpFile, 0400)
			throttler := NewFlowCondition(tmpFile, time.Hour, "https://cp.example.com", logger)
			err := throttler.UpdateSuccessAPITimestamp()
			Expect(err).To(HaveOccurred())
			os.Chmod(tmpFile, 0600)
		})
	})
})
