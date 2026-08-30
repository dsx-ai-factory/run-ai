package keepalive

import (
	"context"
	"net"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ping", func() {
	var (
		conn1 net.Conn
		conn2 net.Conn
		ping  *Ping
	)

	BeforeEach(func() {
		conn1, conn2 = net.Pipe()
		ping = NewPing(conn1, conn2, time.Millisecond)
	})

	AfterEach(func() {
		_ = conn1.Close()
		_ = conn2.Close()
	})

	Context("check channel timeout", func() {
		It("Ping connection timeout reached", func() {
			ctx, cancel := context.WithTimeout(context.TODO(), time.Millisecond*100)
			defer cancel()

			go func() {
				byteArr := make([]byte, 10)
				_, _ = conn2.Read(byteArr)
			}()

			select {
			case <-ping.Channel():
				err := ping.Execute()
				Expect(err).ToNot(HaveOccurred())
			case <-ctx.Done():
				Expect(ctx.Err()).ToNot(HaveOccurred())
			}
		})

		It("Ping connection execute single time", func() {
			var (
				err     error
				byteArr = make([]byte, 6)
			)
			ctx, cancel := context.WithTimeout(context.TODO(), time.Millisecond*100)
			defer cancel()

			Expect(conn2).ToNot(BeNil())
			go func() {
				// read once, terminate after, will conn1 write failed the next time it reads
				_, _ = conn2.Read(byteArr)
				_ = conn2.Close()
			}()

			select {
			case <-ping.Channel():
				err = ping.Execute()
				Expect(err).ToNot(HaveOccurred())
			case <-ctx.Done():
				Expect(ctx.Err()).ToNot(HaveOccurred())
			}

			select {
			case <-ping.Channel():
				err = ping.Execute()
				Expect(err).To(HaveOccurred())
			case <-ctx.Done():
				Expect(ctx.Err()).ToNot(HaveOccurred())
			}
		})
	})
})
