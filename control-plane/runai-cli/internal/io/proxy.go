package io

import (
	"context"
	"io"
	"time"

	"golang.org/x/sync/errgroup"
)

//TODO: remove this after moving to mono repo

type OneWayProxy struct {
	shutdownGrace time.Duration
}

func NewOneWayProxy(shutdownGrace time.Duration) *OneWayProxy {
	return &OneWayProxy{shutdownGrace: shutdownGrace}
}

func (p *OneWayProxy) Start(ctx context.Context, remote io.ReadWriteCloser, local io.ReadWriteCloser) error {
	return proxy(ctx, remote, local, p.shutdownGrace)
}

type TwoWayProxy struct {
	shutdownGrace time.Duration
}

func NewTwoWayProxy(shutdownGrace time.Duration) *TwoWayProxy {
	return &TwoWayProxy{shutdownGrace: shutdownGrace}
}

func (p *TwoWayProxy) Start(ctx context.Context, remote io.ReadWriteCloser, local io.ReadWriteCloser) error {
	eg, errCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return proxy(errCtx, remote, local, p.shutdownGrace)
	})
	eg.Go(func() error {
		return proxy(errCtx, local, remote, p.shutdownGrace)
	})
	return eg.Wait()
}

// proxy is a function that handles data transfer between two io.ReadWriteCloser interfaces.
// It takes a context, two io.ReadWriteCloser interfaces (from and to), and a time.Duration for graceful shutdown.
// The function starts a goroutine that copies data from the 'from' ReadCloser to the 'to' WriteCloser.
// It listens for signals on the context's Done channel and an error channel.
// If a signal is received from the context's Done channel, it attempts to close both the 'from' and 'to' interfaces.
// If a signal is received from the error channel, it returns the error.
// If the context is done,it creates a new context with a cancellation
// and waits for the copy operation to finish.
func proxy(ctx context.Context, from io.ReadWriteCloser, to io.ReadWriteCloser, shutdownGrace time.Duration) error {
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
		case <-ctx.Done():
			// If a signal is received from the context's Done channel, attempt to close both the 'from' and 'to' interfaces
			from.Close()
			to.Close()

			// Create a new context with a timeout
			shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownGrace)
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
			// If a signal is received from the error channel outside the context's Done case, return the error
			if err != nil {
				return err
			}
			return nil
		}
	}
}
