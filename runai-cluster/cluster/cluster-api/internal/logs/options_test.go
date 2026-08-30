package logs

import (
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

var _ = Describe("Options Logs", func() {
	Describe("Options", func() {
		It("should convert to k8s logs options", func() {
			options := Options{
				Container:    "container",
				TailLines:    ptr.To[int64](10),
				Timestamps:   true,
				Follow:       true,
				Previous:     false,
				LimitBytes:   100,
				SinceSeconds: 60 * time.Second,
				SinceTime:    "2023-05-12T15:04:05Z",
			}

			k8sOptions, err := options.ToK8sLogsOptions()

			Expect(err).NotTo(HaveOccurred())
			Expect(k8sOptions.Container).To(Equal("container"))
			Expect(*k8sOptions.TailLines).To(Equal(int64(10)))
			Expect(k8sOptions.Timestamps).To(BeTrue())
			Expect(k8sOptions.Follow).To(BeTrue())
			Expect(k8sOptions.Previous).To(BeFalse())
			Expect(*k8sOptions.LimitBytes).To(Equal(int64(100)))
			Expect(*k8sOptions.SinceSeconds).To(Equal(int64(60)))
			Expect(k8sOptions.SinceTime).NotTo(BeNil())
		})

		It("should validate options", func() {
			validOptions := Options{
				TailLines:    ptr.To[int64](10),
				LimitBytes:   100,
				SinceSeconds: 60 * time.Second,
			}
			Expect(validOptions.Validate()).NotTo(HaveOccurred())

			invalidOptions := []Options{
				{SinceTime: "2023-05-12T15:04:05Z", SinceSeconds: 60 * time.Second},
				{LimitBytes: -1},
				{SinceSeconds: -60 * time.Second},
				{TailLines: ptr.To[int64](-2)},
			}
			for _, options := range invalidOptions {
				Expect(options.Validate()).To(HaveOccurred())
			}
		})
		It("should error when sine time and since seconds are both set", func() {
			options := Options{
				SinceTime:    "2023-05-12T15:04:05Z",
				SinceSeconds: 60 * time.Second,
			}
			Expect(options.Validate()).To(HaveOccurred())
		})
	})

	Describe("OptionsBuilder", func() {
		It("should build options", func() {
			builder := NewOptionsBuilder().
				WithPod("pod").
				WithContainer("container").
				WithTailLines(10).
				WithTimestamps(true).
				WithFollow(true).
				WithPrevious(false).
				WithLimitBytes(100).
				WithSinceSeconds(60)

			options, err := builder.Build()

			Expect(err).NotTo(HaveOccurred())
			Expect(options.Pod).To(Equal("pod"))
			Expect(options.Container).To(Equal("container"))
			Expect(ptr.Deref(options.TailLines, 0)).To(Equal(int64(10)))
			Expect(options.Timestamps).To(BeTrue())
			Expect(options.Follow).To(BeTrue())
			Expect(options.Previous).To(BeFalse())
			Expect(options.LimitBytes).To(Equal(int64(100)))
			Expect(options.SinceSeconds).To(Equal(60 * time.Second))
		})

		It("should return an error for invalid options", func() {
			builder := NewOptionsBuilder().WithTailLines(-2)
			_, err := builder.Build()
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("FromGetRequest", func() {
		It("should parse options from GET request", func() {
			req, err := http.NewRequest("GET", "http://example.com/logs?pod=pod&container=container&tailLines=10&timestamps=true&follow=true&previous=false&limitBytes=100&sinceSeconds=60", nil)
			Expect(err).NotTo(HaveOccurred())

			options, err := FromGetRequest(req)

			Expect(err).NotTo(HaveOccurred())
			Expect(options.Pod).To(Equal("pod"))
			Expect(options.Container).To(Equal("container"))
			Expect(ptr.Deref(options.TailLines, 0)).To(Equal(int64(10)))
			Expect(options.Timestamps).To(BeTrue())
			Expect(options.Follow).To(BeTrue())
			Expect(options.Previous).To(BeFalse())
			Expect(options.LimitBytes).To(Equal(int64(100)))
			Expect(options.SinceSeconds).To(Equal(60 * time.Second))
			Expect(options.SinceTime).To(BeEmpty())
		})

		It("should parse options from GET request without follow", func() {
			req, err := http.NewRequest("GET", "http://example.com/logs?pod=pod&container=container&tailLines=10&timestamps=true&previous=false&limitBytes=100&sinceSeconds=60", nil)
			Expect(err).NotTo(HaveOccurred())

			options, err := FromGetRequest(req)

			Expect(err).NotTo(HaveOccurred())
			Expect(options.Pod).To(Equal("pod"))
			Expect(options.Container).To(Equal("container"))
			Expect(ptr.Deref(options.TailLines, 0)).To(Equal(int64(10)))
			Expect(options.Timestamps).To(BeTrue())
			Expect(options.Follow).To(BeFalse())
			Expect(options.Previous).To(BeFalse())
			Expect(options.LimitBytes).To(Equal(int64(100)))
			Expect(options.SinceSeconds).To(Equal(60 * time.Second))
			Expect(options.SinceTime).To(BeEmpty())
		})

		It("should return an error for invalid options", func() {
			req, err := http.NewRequest("GET", "http://example.com/logs?tailLines=invalid", nil)
			Expect(err).NotTo(HaveOccurred())

			_, err = FromGetRequest(req)
			Expect(err).To(HaveOccurred())
		})
	})
})
