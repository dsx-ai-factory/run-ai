package websocket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/config"

	"nhooyr.io/websocket"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TransportDialer", func() {
	var (
		token  string
		dialer *TransportDialer
		server *httptest.Server
	)

	BeforeEach(func() {
		token = "test-token"
		dialer = NewTransportDialer(token)
	})

	AfterEach(func() {
		if server != nil {
			server.Close()
		}
	})

	Describe("NewTransportDialer", func() {
		It("creates a new TransportDialer with the provided token", func() {
			Expect(dialer).NotTo(BeNil())
			Expect(dialer.token).To(Equal(token))
		})
	})

	Describe("Connect", func() {
		BeforeEach(func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.Method).To(Equal("GET"))
				Expect(r.URL.Path).To(Equal("/"))
				Expect(r.Header.Get(config.AuthorizationHeader)).To(Equal("Bearer " + token))

				options := &websocket.AcceptOptions{}
				conn, err := websocket.Accept(w, r, options)
				Expect(err).NotTo(HaveOccurred())
				defer conn.Close(websocket.StatusInternalError, "")
			}))
		})

		It("connects to the WebSocket server", func() {
			url, err := url.Parse(server.URL)
			Expect(err).NotTo(HaveOccurred())
			url.Scheme = "ws"

			ctx := context.Background()
			conn, resp, err := dialer.Connect(ctx, url, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(conn).NotTo(BeNil())
			Expect(resp).NotTo(BeNil())
			Expect(resp.StatusCode).To(Equal(http.StatusSwitchingProtocols))

			Expect(conn.Close()).To(Succeed())
		})
	})
})
