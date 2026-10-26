/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package chart_inspection

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type inspectionResolver func(context.Context, string) ([]net.IPAddr, error)

func (r inspectionResolver) LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	return r(ctx, host)
}

func testRemoteFetcher(handler http.HandlerFunc) (*remoteFetcher, string, *[]string) {
	GinkgoHelper()

	server := httptest.NewTLSServer(handler)
	DeferCleanup(server.Close)
	address := server.Listener.Addr().String()
	_, port, _ := net.SplitHostPort(address)
	var mu sync.Mutex
	targets := []string{}
	resolver := inspectionResolver(func(context.Context, string) ([]net.IPAddr, error) {
		return []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}}, nil
	})
	dial := func(ctx context.Context, network, target string) (net.Conn, error) {
		mu.Lock()
		targets = append(targets, target)
		mu.Unlock()
		return (&net.Dialer{}).DialContext(ctx, network, address)
	}
	fetcher := newRemoteFetcher(resolver, dial)
	roots := x509.NewCertPool()
	roots.AddCert(server.Certificate())
	fetcher.tlsConfig = &tls.Config{RootCAs: roots, ServerName: server.Certificate().DNSNames[0], MinVersion: tls.VersionTLS12}
	return fetcher, "https://helm.ngc.nvidia.com:" + port, &targets
}

var _ = Describe("Remote chart acquisition", func() {
	It("rejects IPv6 site-local literals and DNS answers before dialing", func() {
		for _, address := range []string{"fec0::1", "feff:ffff::1"} {
			By("literal " + address)
			_, err := validateSourceURL("https://[" + address + "]/chart.tgz")
			Expect(err).To(HaveOccurred())
			By("DNS " + address)
			dialed := false
			fetcher := newRemoteFetcher(inspectionResolver(func(context.Context, string) ([]net.IPAddr, error) {
				return []net.IPAddr{{IP: net.ParseIP(address)}}, nil
			}), func(context.Context, string, string) (net.Conn, error) {
				dialed = true
				return nil, errors.New("unexpected dial")
			})
			_, dialErr := fetcher.safeDial(context.Background(), "tcp", "registry.example.com:443")
			Expect(dialErr).To(HaveOccurred())
			Expect(dialed).To(BeFalse())
		}
	})

	It("Fetch Remote Direct And Repository", func() {

		var paths []string
		fetcher, base, targets := testRemoteFetcher(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			paths = append(paths, r.URL.Path)
			username, password, ok := r.BasicAuth()
			Expect(!ok || username != "$oauthtoken" || password != "synthetic-test-key").To(BeFalse(),
				"missing scoped NGC authentication",
			)

			switch r.URL.Path {
			case "/repo/index.yaml":
				io.WriteString(w, "apiVersion: v1\nentries:\n  app:\n    - name: app\n      version: '1.0'\n      urls: [wrong.tgz]\n    - name: app\n      version: '2.0'\n      urls: ['../app.tgz?signature=synthetic']\n")
			case "/app.tgz":
				io.WriteString(w, "chart-archive")
			default:
				http.NotFound(w, r)
			}
		})
		got, err := fetcher.fetch(context.Background(), Selection{Repo: base + "/repo", ChartName: fixtureChartName, ChartVersion: "2.0"}, "synthetic-test-key")
		Expect(err != nil || string(got) != "chart-archive").To(BeFalse(), func() string {
			return fmt.Sprintf("got %q, %v", got, err)
		})
		Expect(strings.Join(paths, ",")).To(Equal("/repo/index.yaml,/app.tgz"), func() string {
			return fmt.Sprintf("paths %v", paths)
		})

		for _, address := range *targets {
			Expect(strings.HasPrefix(address, "8.8.8.8:")).To(BeTrue(), func() string {
				return fmt.Sprintf("validated IP not pinned: %s", address)
			})
		}
		got, err = fetcher.fetch(context.Background(), Selection{ChartURL: base + "/app.tgz"}, "synthetic-test-key")
		Expect(err != nil || string(got) != "chart-archive").To(BeFalse(), func() string {
			return fmt.Sprintf("direct got %q, %v", got, err)
		})
	})
	It("Fetch Remote Bounds Expanded Index Aliases", func() {
		index := "entries:\n  app: [{name: app, version: '1', urls: [app.tgz]}]\nlarge: &large " + strings.Repeat("x", 128<<10) + "\nrepeated: [*large" + strings.Repeat(", *large", 99) + "]\n"
		fetcher, base, _ := testRemoteFetcher(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/repo/index.yaml" {
				io.WriteString(w, index)
				return
			}
			io.WriteString(w, "chart-archive")
		})
		_, err := fetcher.fetch(context.Background(), Selection{Repo: base + "/repo", ChartName: fixtureChartName, ChartVersion: "1"}, "")
		var typed *api.Error
		Expect(!errors.As(err, &typed) || !strings.HasPrefix(typed.Message, "invalid_chart_index:")).To(BeFalse(), func() string {
			return fmt.Sprintf("expanded repository aliases were accepted: %v", err)
		})
	})
	It("Fetch Remote Rejects Unsafe Sources", func() {
		for _, raw := range []string{
			"file:///etc/passwd", "oci://registry.example.com/app", "http://127.0.0.1/a.tgz",
			"https://[::1]/a.tgz", "http://169.254.169.254/a",
			"https://user:password@example.com/a", "https://example.com/a#fragment",
		} {
			{
				By(raw)
				_, err := FetchRemote(context.Background(), Selection{ChartURL: raw}, "")
				Expect(err).NotTo(BeNil(), "unsafe source accepted")

				var typed *api.Error
				Expect(!errors.As(err, &typed) || typed.Code != http.StatusBadRequest).To(BeFalse(), func() string {
					return fmt.Sprintf("untyped error: %v", err)
				})
				Expect(strings.Contains(err.Error(), "password")).To(BeFalse(),
					"embedded credentials leaked",
				)
			}
		}
	})
	It("Fetch Remote Blocks Private DNS Before Dial", func() {
		dialed := false
		fetcher := newRemoteFetcher(inspectionResolver(func(context.Context, string) ([]net.IPAddr, error) {
			return []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}, {IP: net.ParseIP("10.1.2.3")}}, nil
		}), func(context.Context, string, string) (net.Conn, error) {
			dialed = true
			return nil, errors.New("unexpected dial")
		})
		_, err := fetcher.fetch(context.Background(), Selection{ChartURL: "https://registry.example.com/app.tgz"}, "")
		Expect(err == nil || dialed).To(BeFalse(), func() string {
			return fmt.Sprintf("private DNS accepted or dialed: %v, %v", err, dialed)
		})
	})
	It("Fetch Remote Redirect Rechecks DNS And Strips Credential", func() {

		fetcher, base, _ := testRemoteFetcher(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			if r.URL.Path == "/first" {
				http.Redirect(w, r, "https://cdn.example.com:"+strings.Split(r.Host, ":")[1]+"/archive?signature=synthetic", http.StatusFound)
				return
			}
			Expect(r.Header.Get("Authorization")).To(Equal(""), "NGC credential leaked to CDN")

			io.WriteString(w, "archive")
		})
		got, err := fetcher.fetch(context.Background(), Selection{ChartURL: base + "/first"}, "synthetic-test-key")
		Expect(err != nil || string(got) != "archive").To(BeFalse(), func() string {
			return fmt.Sprintf("signed CDN redirect: %q %v", got, err)
		})

		fetcher.resolver = inspectionResolver(func(_ context.Context, host string) ([]net.IPAddr, error) {
			if host == "cdn.example.com" {
				return []net.IPAddr{{IP: net.ParseIP("192.168.1.1")}}, nil
			}
			return []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}}, nil
		})
		_, err = fetcher.fetch(context.Background(), Selection{ChartURL: base + "/first"}, "synthetic-test-key")
		Expect(err).NotTo(BeNil(), "redirect to private DNS accepted")
		Expect(strings.Contains(err.Error(), "signature") || strings.Contains(err.Error(), "synthetic-test-key")).To(BeFalse(),
			"secret URL or credential leaked",
		)
	})
	It("Fetch Remote Drops Credential For Cross Origin Index Link", func() {

		fetcher, base, _ := testRemoteFetcher(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			if r.URL.Path == "/index.yaml" {
				io.WriteString(w, "entries:\n  app:\n    - version: '1'\n      urls: ['https://cdn.example.com:"+strings.Split(r.Host, ":")[1]+"/chart.tgz']\n")
				return
			}
			Expect(r.Header.Get("Authorization")).To(Equal(""), "credential leaked through package URL")

			io.WriteString(w, "archive")
		})
		got, err := fetcher.fetch(context.Background(), Selection{Repo: base, ChartName: fixtureChartName, ChartVersion: "1"}, "synthetic-test-key")
		Expect(err != nil || string(got) != "archive").To(BeFalse(), func() string {
			return fmt.Sprintf("got %q %v", got, err)
		})
	})
	It("Fetch Remote Validates Selection Index And Limits", func() {
		cases := []struct {
			name, body string
			selection  Selection
			wantStatus int
		}{
			{name: "duplicate index keys", body: "entries: {}\nentries: {}\n", wantStatus: http.StatusBadRequest},
			{name: "missing exact version", body: "entries:\n  app:\n    - version: '2'\n      urls: [chart.tgz]\n", wantStatus: http.StatusBadRequest},
			{name: "index too large", body: strings.Repeat("x", (1<<20)+1), wantStatus: http.StatusBadRequest},
		}
		for _, tc := range cases {
			{
				By(tc.name)
				fetcher, base, _ := testRemoteFetcher(func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, tc.body) })
				_, err := fetcher.fetch(context.Background(), Selection{Repo: base, ChartName: fixtureChartName, ChartVersion: "1"}, "")
				var typed *api.Error
				Expect(!errors.As(err, &typed) || typed.Code != int32(tc.wantStatus)).To(BeFalse(), func() string {
					return fmt.Sprintf("wrong failure %v", err)
				})
			}
		}
		for _, selection := range []Selection{
			{}, {Repo: fixtureRepositoryURL, ChartName: fixtureChartName},
			{ChartURL: "https://example.com/a", Repo: fixtureRepositoryURL, ChartName: fixtureChartName, ChartVersion: "1"},
			{ChartID: "stored"},
		} {
			{
				_, err := FetchRemote(context.Background(), selection, "")
				Expect(err).NotTo(BeNil(), func() string {
					return fmt.Sprintf("invalid selection accepted %+v", selection)
				})
			}
		}
		{
			_, err := FetchRemote(context.Background(), Selection{ChartURL: "https://notngc.nvidia.com/a"}, "synthetic-test-key")
			Expect(err).NotTo(BeNil(), "key accepted for non-NGC host")
		}
	})
	It("Fetch Remote Bounds Archive And Redacts Failures", func() {
		fetcher, base, _ := testRemoteFetcher(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/large" {
				w.Header().Set("Content-Length", "33554433")
				return
			}
			w.WriteHeader(http.StatusForbidden)
		})
		for _, path := range []string{"/large", "/forbidden?signature=sensitive"} {
			_, err := fetcher.fetch(context.Background(), Selection{ChartURL: base + path}, "")
			Expect(err == nil || strings.Contains(err.Error(), "sensitive") || strings.Contains(err.Error(), base)).To(BeFalse(), func() string {
				return fmt.Sprintf("missing bound or unredacted error: %v", err)
			})
		}
	})
	It("Fetch Remote Cancellation", func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := FetchRemote(ctx, Selection{ChartURL: "https://example.com/a"}, "")
		Expect(err).To(HaveOccurred())
		Expect(err.Code).To(BeNumerically("==", http.StatusRequestTimeout))
	})
	It("Fetch Remote Preserves Escaped Repository Path", func() {

		fetcher, base, _ := testRemoteFetcher(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			if r.URL.Path == "/repo/team/index.yaml" {
				Expect(r.URL.EscapedPath()).To(Equal("/repo%2Fteam/index.yaml"), func() string {
					return fmt.Sprintf("repository path changed: %s", r.URL.EscapedPath())
				})

				io.WriteString(w, "entries:\n  app:\n    - version: '1'\n      urls: [/archive.tgz]\n")
				return
			}
			io.WriteString(w, "archive")
		})
		{
			_, err := fetcher.fetch(context.Background(), Selection{Repo: base + "/repo%2Fteam", ChartName: fixtureChartName, ChartVersion: "1"}, "")
			Expect(err).NotTo(HaveOccurred())
		}
	})
	It("Fetch Remote Upstream HTTP Failures Return Service Unavailable", func() {
		for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusInternalServerError, http.StatusServiceUnavailable} {
			{
				By(http.StatusText(status))
				fetcher, base, _ := testRemoteFetcher(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(status)
					_, _ = io.WriteString(w, "upstream-sensitive-body")
				})
				for _, selection := range []Selection{
					{ChartURL: base + "/chart.tgz?signature=upstream-sensitive-query"},
					{Repo: base, ChartName: fixtureChartName, ChartVersion: "1"},
				} {
					_, err := fetcher.fetch(context.Background(), selection, "synthetic-test-key")
					var sourceErr *api.Error
					Expect(!errors.As(err, &sourceErr) || sourceErr.Code != http.StatusServiceUnavailable || !strings.HasPrefix(sourceErr.Message, "chart_download_failed:")).To(BeFalse(), func() string {
						return fmt.Sprintf("upstream HTTP %d returned %v", status, err)
					})
					Expect(strings.Contains(err.Error(), "upstream-sensitive") || strings.Contains(err.Error(), "synthetic-test-key")).To(BeFalse(), func() string {
						return fmt.Sprintf("upstream failure exposed source material: %v", err)
					})
				}
			}
		}
	})
})

var _ = Describe("Chart source address policy", func() {
	DescribeTable("keeps the address policy at CIDR boundaries", func(address string, blocked bool) {
		ip := net.ParseIP(address)
		Expect(ip).NotTo(BeNil())
		Expect(blockedIP(ip)).To(Equal(blocked))
		if ipv4 := ip.To4(); ipv4 != nil {
			Expect(blockedIP(ipv4)).To(Equal(blocked), "four-byte IPv4 representation")
		}
	},
		Entry("current network", "0.255.255.255", true),
		Entry("after current network", "1.0.0.0", false),
		Entry("before CGNAT", "100.63.255.255", false),
		Entry("CGNAT start", "100.64.0.0", true),
		Entry("CGNAT end", "100.127.255.255", true),
		Entry("after CGNAT", "100.128.0.0", false),
		Entry("before benchmark", "198.17.255.255", false),
		Entry("benchmark start", "198.18.0.0", true),
		Entry("benchmark end", "198.19.255.255", true),
		Entry("after benchmark", "198.20.0.0", false),
		Entry("reserved start", "240.0.0.0", true),
		Entry("reserved end", "255.255.255.255", true),
		Entry("before site-local", "fe7f:ffff:ffff:ffff:ffff:ffff:ffff:ffff", false),
		Entry("link-local", "fe80::1", true),
		Entry("site-local start", "fec0::", true),
		Entry("site-local end", "feff:ffff:ffff:ffff:ffff:ffff:ffff:ffff", true),
		Entry("unique-local", "fc00::1", true),
		Entry("IPv4-mapped CGNAT", "::ffff:100.64.0.1", true),
		Entry("public IPv6", "2606:4700:4700::1111", false),
		Entry("private IPv4", "10.0.0.1", true),
		Entry("IPv4 loopback", "127.0.0.1", true),
		Entry("IPv6 loopback", "::1", true),
	)
})
