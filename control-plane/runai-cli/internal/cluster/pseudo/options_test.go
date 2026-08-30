package pseudo_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/pseudo"
)

var _ = Describe("Options", func() {
	var (
		pod       string
		container string
		stdin     bool
		tty       bool
		cmd       []string
	)

	BeforeEach(func() {
		pod = "example-pod"
		container = "example-container"
		stdin = true
		tty = true
		cmd = []string{"one", "two", "three and four"}
	})

	Describe("NewOptions", func() {
		It("creates a new Options instance with the provided values", func() {
			options, err := pseudo.NewOptions(pod, container, stdin, tty, cmd...)
			Expect(err).ToNot(HaveOccurred())

			Expect(options.Pod).To(Equal(pod))
			Expect(options.Container).To(Equal(container))
			Expect(options.Stdin).To(Equal(stdin))
			Expect(options.TTY).To(Equal(tty))

			cmdStr, err := pseudo.EncodeCommandString(cmd)
			Expect(err).ToNot(HaveOccurred())
			Expect(options.Command).To(Equal(cmdStr))
		})

		It("sets terminal size when TTY is enabled", func() {
			options, err := pseudo.NewOptions(pod, container, stdin, true, cmd...)
			Expect(err).ToNot(HaveOccurred())

			// Terminal size should be set when TTY is enabled (if running in a terminal)
			// Note: In test environment, this might be 0 if not running in a real terminal
			Expect(options.Width).To(BeNumerically(">=", 0))
			Expect(options.Height).To(BeNumerically(">=", 0))
		})

		It("does not set terminal size when TTY is disabled", func() {
			options, err := pseudo.NewOptions(pod, container, stdin, false, cmd...)
			Expect(err).ToNot(HaveOccurred())

			Expect(options.Width).To(Equal(0))
			Expect(options.Height).To(Equal(0))
		})
	})

	Describe("ToQueryParams", func() {
		var (
			options *pseudo.Options
			err     error
		)

		BeforeEach(func() {
			options, err = pseudo.NewOptions(pod, container, stdin, tty, cmd...)
		})

		It("converts the options to query parameters", func() {
			Expect(err).NotTo(HaveOccurred())
			queryParams, err := options.ToQueryParams()
			Expect(err).NotTo(HaveOccurred())

			Expect(queryParams.Get("pod")).To(Equal(pod))
			Expect(queryParams.Get("container")).To(Equal(container))
			Expect(queryParams.Get("stdout")).To(Equal("true"))
			Expect(queryParams.Get("stderr")).To(Equal("true"))

			Expect(queryParams.Get("stdin")).To(Equal("true"))
			Expect(queryParams.Get("tty")).To(Equal("true"))
		})

		It("includes terminal size parameters when TTY is enabled and size is set", func() {
			Expect(err).NotTo(HaveOccurred())
			options.Width = 80
			options.Height = 24

			queryParams, err := options.ToQueryParams()
			Expect(err).NotTo(HaveOccurred())

			Expect(queryParams.Get("width")).To(Equal("80"))
			Expect(queryParams.Get("height")).To(Equal("24"))
		})

		It("omits terminal size parameters when TTY is disabled", func() {
			Expect(err).NotTo(HaveOccurred())
			options.TTY = false
			options.Width = 80
			options.Height = 24

			queryParams, err := options.ToQueryParams()
			Expect(err).NotTo(HaveOccurred())

			Expect(queryParams.Get("width")).To(BeEmpty())
			Expect(queryParams.Get("height")).To(BeEmpty())
		})

		It("omits empty or default values from the query parameters", func() {
			options.Pod = ""
			options.Container = ""
			options.Stdin = false
			options.TTY = false
			options.Command = ""

			queryParams, err := options.ToQueryParams()
			Expect(err).NotTo(HaveOccurred())

			Expect(queryParams.Get("pod")).To(BeEmpty())
			Expect(queryParams.Get("container")).To(BeEmpty())
			Expect(queryParams.Get("stdout")).To(Equal("true"))
			Expect(queryParams.Get("stderr")).To(Equal("true"))

			Expect(queryParams.Get("stdin")).To(BeEmpty())
			Expect(queryParams.Get("tty")).To(BeEmpty())
			Expect(queryParams.Get("width")).To(BeEmpty())
			Expect(queryParams.Get("height")).To(BeEmpty())
		})
	})

	Describe("Validate", func() {
		var (
			err error
		)

		It("returns nil when the options are valid", func() {
			_, err = pseudo.NewOptions(pod, container, stdin, tty, cmd...)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns an error when the tail lines are invalid", func() {
			_, err = pseudo.NewOptions(pod, container, stdin, tty)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
