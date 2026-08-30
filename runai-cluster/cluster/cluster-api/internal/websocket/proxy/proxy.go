package proxy

import (
	"context"
	"io"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/websocket"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/websocket/keepalive"
	"golang.org/x/sync/errgroup"
)

const shutDownDuration = 1 * time.Second

type ProxyHandler interface {
	Start(ctx context.Context) error
}

type TwoWayProxy struct {
	remote           websocket.ReadWriteCloser
	local            websocket.ReadWriteCloser
	keepalive        keepalive.Keepalive
	shutDownDuration time.Duration
}

func NewTwoWayProxy(remote, local websocket.ReadWriteCloser, keepalive keepalive.Keepalive) *TwoWayProxy {
	return &TwoWayProxy{
		remote:           remote,
		local:            local,
		keepalive:        keepalive,
		shutDownDuration: shutDownDuration,
	}
}

func (p *TwoWayProxy) Start(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return proxy(ctx, p.remote, p.local, p.keepalive, p.shutDownDuration)
	})
	eg.Go(func() error {
		return proxy(ctx, p.local, p.remote, &keepalive.NoKeepalive{}, p.shutDownDuration)
	})

	return eg.Wait()
}

type OneWayProxy struct {
	remote    websocket.ReadCloser
	local     websocket.ReadWriteCloser
	keepalive keepalive.Keepalive
}

func NewOneWayProxy(
	remote websocket.ReadCloser, local websocket.ReadWriteCloser, keepalive keepalive.Keepalive,
) *OneWayProxy {
	return &OneWayProxy{
		remote:    remote,
		local:     local,
		keepalive: keepalive,
	}
}

func (p *OneWayProxy) Start(ctx context.Context) error {
	return proxy(ctx, p.remote, p.local, p.keepalive, shutDownDuration)
}

// proxy is a function that handles data transfer between two io.ReadWriteCloser interfaces.
// It takes a context, two io.ReadWriteCloser interfaces (from and to), and a time.Duration for graceful shutdown.
// The function starts a goroutine that copies data from the 'from' ReadCloser to the 'to' WriteCloser.
// It listens for signals on the context's Done channel and an error channel.
// If a signal is received from the context's Done channel, it attempts to close both the 'from' and 'to' interfaces.
// If a signal is received from the error channel, it returns the error.
// If the context is done,it creates a new context with a cancellation
// and waits for the copy operation to finish.
func proxy(
	ctx context.Context, from io.ReadCloser, to io.WriteCloser,
	keepalive keepalive.Keepalive, graceShutDown time.Duration,
) error {
	// Create an error channel to communicate errors from the copy goroutine
	errChan := make(chan error)

	// Start a goroutine to copy data from 'from' to 'to'
	go func() {
		_, err := io.Copy(to, from)

		errChan <- err

		close(errChan)
	}()

	// Continuously listen for signals on the context's Done channel and the error channel
	for {
		select {
		case <-keepalive.Channel():
			kEerr := keepalive.Execute()
			if kEerr != nil {
				log.Info().Msgf("failed to send keep-alive, error: %v", kEerr)
				return kEerr
			}
		case <-ctx.Done():
			// If a signal is received from the context's Done channel, attempt to close both the 'from' and 'to' interfaces
			err := from.Close()
			if err != nil {
				log.Warn().Err(err).Msg("Failed to close the Reader")
			}

			err = to.Close()
			if err != nil {
				log.Warn().Err(err).Msg("Failed to close the Writer")
			}

			// Create a new context with a timeout
			shutdownCtx, cancel := context.WithTimeout(context.Background(), graceShutDown)
			defer cancel()

			// Wait for either the copy goroutine to finish or the shutdown timeout to expire
			select {
			case <-errChan:
				// If a signal is received from the error channel, return the context's error
				return ctx.Err()
			case <-shutdownCtx.Done():
				// If the shutdown timeout expires, return an error
				return shutdownCtx.Err()
			}
		case err := <-errChan:
			// If a signal is received from the error channel outside of the context's Done case, return the error
			if err != nil {
				return err
			}

			return nil
		}
	}
}
