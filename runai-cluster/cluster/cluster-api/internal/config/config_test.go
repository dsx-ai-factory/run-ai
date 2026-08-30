package config

import (
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	testDefaultPort                 = 4180
	testDefaultWriteTimeout         = 15
	testDefaultReadTimeout          = 15
	testDefaultIdleTimeout          = 60
	testDefaultK8sClientConfigQPS   = 50
	testDefaultK8sClientConfigBurst = 300
	testDefaultUseDebugLogLevel     = false
	testDefaultGraceShutdownTimeout = 30
)

var _ = Describe("Config", func() {
	var originalEnv map[string]string

	BeforeEach(func() {
		originalEnv = make(map[string]string)
		for _, env := range os.Environ() {
			parts := strings.SplitN(env, "=", 2)
			originalEnv[parts[0]] = parts[1]
		}
	})

	AfterEach(func() {
		for key, value := range originalEnv {
			Expect(os.Setenv(key, value)).To(Succeed())
		}
	})

	Describe("newConfig", func() {
		It("should return the default configuration values when environment variables are not set", func() {
			cfg := Config()
			Expect(cfg.Port).To(Equal(testDefaultPort))
			Expect(cfg.WriteTimeout).To(Equal(time.Duration(testDefaultWriteTimeout) * time.Second))
			Expect(cfg.ReadTimeout).To(Equal(time.Duration(testDefaultReadTimeout) * time.Second))
			Expect(cfg.IdleTimeout).To(Equal(time.Duration(testDefaultIdleTimeout) * time.Second))
			Expect(cfg.GraceShutdownTimeout).To(Equal(testDefaultGraceShutdownTimeout * time.Second))
			Expect(cfg.K8sClientConfigQPS).To(Equal(testDefaultK8sClientConfigQPS))
			Expect(cfg.K8sClientConfigBurst).To(Equal(testDefaultK8sClientConfigBurst))
			Expect(cfg.UseDebugLogLevel).To(Equal(testDefaultUseDebugLogLevel))
			Expect(cfg.CacheSyncTimeout).To(Equal(3 * time.Minute))
		})
	})
})
