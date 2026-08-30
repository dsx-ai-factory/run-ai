package builders

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	"os"
	"strings"
)

var _ = Describe("Network tests", func() {
	var (
		ctx context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		// Set project in context for tests that need it
		ctx = project.SetProjectInContext(ctx, "test-project")
	})

	Describe("UrlToPort tests", func() {
		It("should return DefaultIngressHttpPort for nil URL", func() {
			port, err := UrlToPort(nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(port).To(Equal(int32(DefaultIngressHttpPort)))
		})

		It("should return DefaultIngressHttpPort for URL with no port", func() {
			url, err := apis.ParseURL("http://example.com")
			Expect(err).NotTo(HaveOccurred())
			port, err := UrlToPort(url)
			Expect(err).NotTo(HaveOccurred())
			Expect(port).To(Equal(int32(DefaultIngressHttpPort)))
		})

		It("should return the correct port for URL with valid port", func() {
			url, err := apis.ParseURL("http://example.com:8080")
			Expect(err).NotTo(HaveOccurred())
			port, err := UrlToPort(url)
			Expect(err).NotTo(HaveOccurred())
			Expect(port).To(Equal(int32(8080)))
		})

		It("should return DefaultIngressHttpPort for URL with invalid port", func() {
			url, err := apis.ParseURL("http://example.com:8080")
			Expect(err).NotTo(HaveOccurred())
			url.Host = "example.com:invalid"
			port, err := UrlToPort(url)
			Expect(err).NotTo(HaveOccurred())
			Expect(port).To(Equal(int32(DefaultIngressHttpPort)))
		})
	})

	Describe("AddJobUrlToAnnotations tests", func() {
		var (
			resourceMeta *metav1.ObjectMeta
			jobUrlToAdd  string
		)

		BeforeEach(func() {
			resourceMeta = &metav1.ObjectMeta{}
			jobUrlToAdd = "http://example.com/job"
		})

		AfterEach(func() {
			_ = os.Unsetenv(config.EnvVarOpenshift)
			_ = os.Unsetenv(config.EnvVarSubdomainSupport)
		})

		It("should add job URL if annotations are nil", func() {
			AddJobUrlToAnnotations(resourceMeta, jobUrlToAdd)
			Expect(resourceMeta.Annotations).NotTo(BeNil())
			Expect(resourceMeta.Annotations[constants.AnnotationJobUrl]).To(Equal(jobUrlToAdd))
		})

		It("should add job URL if annotations are empty", func() {
			resourceMeta.Annotations = map[string]string{}
			AddJobUrlToAnnotations(resourceMeta, jobUrlToAdd)
			Expect(resourceMeta.Annotations[constants.AnnotationJobUrl]).To(Equal(jobUrlToAdd))
		})

		It("should append job URL if other URLs exist", func() {
			existingUrl := "http://other.example.com"
			resourceMeta.Annotations = map[string]string{
				constants.AnnotationJobUrl: existingUrl,
			}
			AddJobUrlToAnnotations(resourceMeta, jobUrlToAdd)
			expectedUrls := fmt.Sprintf("%s, %s", existingUrl, jobUrlToAdd)
			Expect(resourceMeta.Annotations[constants.AnnotationJobUrl]).To(Equal(expectedUrls))
		})

		It("should not add duplicate job URL", func() {
			resourceMeta.Annotations = map[string]string{
				constants.AnnotationJobUrl: jobUrlToAdd,
			}
			AddJobUrlToAnnotations(resourceMeta, jobUrlToAdd)
			Expect(resourceMeta.Annotations[constants.AnnotationJobUrl]).To(Equal(jobUrlToAdd))
		})

		It("should not add duplicate job URL when multiple exist", func() {
			existingUrl1 := "http://one.example.com"
			existingUrl2 := "http://two.example.com"
			allExisting := fmt.Sprintf("%s, %s, %s", existingUrl1, jobUrlToAdd, existingUrl2)
			resourceMeta.Annotations = map[string]string{
				constants.AnnotationJobUrl: allExisting,
			}
			AddJobUrlToAnnotations(resourceMeta, jobUrlToAdd)
			Expect(resourceMeta.Annotations[constants.AnnotationJobUrl]).To(Equal(allExisting))
		})

		It("should trim trailing slash from job URL", func() {
			jobUrlWithSlash := jobUrlToAdd + "/"
			AddJobUrlToAnnotations(resourceMeta, jobUrlWithSlash)
			Expect(resourceMeta.Annotations[constants.AnnotationJobUrl]).To(Equal(jobUrlToAdd))
		})

		Context("when on OpenShift with subdomain support", func() {
			BeforeEach(func() {
				// Use the helper to ensure original values are restored
				_ = os.Setenv(config.EnvVarOpenshift, "true")
				_ = os.Setenv(config.EnvVarSubdomainSupport, "true")
			})

			AfterEach(func() {
				_ = os.Unsetenv(config.EnvVarOpenshift)
				_ = os.Unsetenv(config.EnvVarSubdomainSupport)
			})

			It("should replace RunaiTrimStr from job URL", func() {
				urlWithRunai := fmt.Sprintf("http://%sexample.com/job", config.RunaiTrimStr)
				expectedUrl := "http://example.com/job"
				AddJobUrlToAnnotations(resourceMeta, urlWithRunai)
				Expect(resourceMeta.Annotations[constants.AnnotationJobUrl]).To(Equal(expectedUrl))
			})

			It("should replace RunaiTrimStr and trim slash", func() {
				urlWithRunaiAndSlash := fmt.Sprintf("http://%sexample.com/job/", config.RunaiTrimStr)
				expectedUrl := "http://example.com/job"
				AddJobUrlToAnnotations(resourceMeta, urlWithRunaiAndSlash)
				Expect(resourceMeta.Annotations[constants.AnnotationJobUrl]).To(Equal(expectedUrl))
			})
		})
	})

	Describe("ComposeExposedUrl tests", func() {
		const (
			jobName = "my-job"
			suffix  = "tool"
		)

		It("should return error for invalid cluster domain", func() {
			_ = os.Setenv(config.EnvVarClusterDomain, "://invalid-domain")
			_, err := ComposeExposedUrlInt(ctx, suffix, jobName)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid cluster domain"))
		})

		Context("with subdomain support enabled", func() {
			BeforeEach(func() {
				_ = os.Setenv(config.EnvVarClusterDomain, "http://cluster.example.com")
				_ = os.Setenv(config.EnvVarSubdomainSupport, "true")
			})

			AfterEach(func() {
				_ = os.Unsetenv(config.EnvVarClusterDomain)
				_ = os.Unsetenv(config.EnvVarSubdomainSupport)
			})

			It("should compose URL with suffix", func() {
				url, err := ComposeExposedUrlInt(ctx, suffix, jobName)
				Expect(err).NotTo(HaveOccurred())
				Expect(url.Host).To(Equal(fmt.Sprintf("test-project-%s-%s.cluster.example.com", jobName, suffix)))
				Expect(url.Path).To(Equal("/"))
			})

			It("should compose URL without suffix", func() {
				url, err := ComposeExposedUrlInt(ctx, "", jobName)
				Expect(err).NotTo(HaveOccurred())
				Expect(url.Host).To(Equal(fmt.Sprintf("test-project-%s.cluster.example.com", jobName)))
				Expect(url.Path).To(Equal("/"))
			})

			It("should compose URL with existing path in cluster domain", func() {
				_ = os.Setenv(config.EnvVarClusterDomain, "http://cluster.example.com/basepath")
				url, err := ComposeExposedUrlInt(ctx, suffix, jobName)
				Expect(err).NotTo(HaveOccurred())
				Expect(url.Host).To(Equal(fmt.Sprintf("test-project-%s-%s.cluster.example.com", jobName, suffix)))
				Expect(url.Path).To(Equal("/basepath"))
			})

			It("should shorten an overlong subdomain deterministically", func() {
				const (
					projectName  = "project-36e8"
					workloadName = "workload-f09e07a5be55-d3bcda"
					urlSuffix    = "url-authenticated-external-url"
				)

				first, err := ComposeExposedUrl(urlSuffix, workloadName, projectName, "https://cluster.example.com", ComposeUrlOptions{SubdomainSupported: true})
				Expect(err).NotTo(HaveOccurred())
				second, err := ComposeExposedUrl(urlSuffix, workloadName, projectName, "https://cluster.example.com", ComposeUrlOptions{SubdomainSupported: true})
				Expect(err).NotTo(HaveOccurred())

				label := strings.Split(first.Host, ".")[0]
				Expect(label).To(HaveLen(maxDNSLabelLength))
				Expect(first.Host).To(Equal(second.Host))
				Expect(first.Host).To(HaveSuffix(".cluster.example.com"))
			})

			It("should not collide when overlong subdomains share a prefix", func() {
				first := shortenDNSLabel("project-workload-url-authenticated-external-url-with-a-long-shared-prefix-a")
				second := shortenDNSLabel("project-workload-url-authenticated-external-url-with-a-long-shared-prefix-b")

				Expect(first).NotTo(Equal(second))
			})
		})

		Context("with subdomain support disabled", func() {
			BeforeEach(func() {
				_ = os.Setenv(config.EnvVarClusterDomain, "http://cluster.example.com")
				_ = os.Setenv(config.EnvVarSubdomainSupport, "false")
			})

			AfterEach(func() {
				_ = os.Unsetenv(config.EnvVarClusterDomain)
				_ = os.Unsetenv(config.EnvVarSubdomainSupport)
			})

			It("should compose URL with suffix", func() {
				url, err := ComposeExposedUrlInt(ctx, suffix, jobName)
				Expect(err).NotTo(HaveOccurred())
				Expect(url.Host).To(Equal("cluster.example.com"))
				Expect(url.Path).To(Equal(fmt.Sprintf("/test-project/%s/%s", jobName, suffix)))
			})

			It("should compose URL without suffix", func() {
				url, err := ComposeExposedUrlInt(ctx, "", jobName)
				Expect(err).NotTo(HaveOccurred())
				Expect(url.Host).To(Equal("cluster.example.com"))
				Expect(url.Path).To(Equal(fmt.Sprintf("/test-project/%s", jobName)))
			})

			It("should compose URL with existing path in cluster domain", func() {
				_ = os.Setenv(config.EnvVarClusterDomain, "http://cluster.example.com")
				url, err := ComposeExposedUrlInt(ctx, suffix, jobName)
				Expect(err).NotTo(HaveOccurred())
				Expect(url.Host).To(Equal("cluster.example.com"))
				Expect(url.Path).To(Equal(fmt.Sprintf("/test-project/%s/%s", jobName, suffix)))
			})
		})

		It("should return error if project name not in context", func() {
			_ = os.Setenv(config.EnvVarClusterDomain, "http://cluster.example.com")
			// Create a new context without the project
			emptyCtx := context.Background()
			_, err := ComposeExposedUrlInt(emptyCtx, suffix, jobName)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("project name not found in context"))
		})
	})
})
