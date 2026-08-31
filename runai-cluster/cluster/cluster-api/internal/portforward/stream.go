/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package portforward

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync/atomic"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/cluster/stream"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/websocket/keepalive"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/apimachinery/pkg/util/runtime"
)

type Stream struct {
	connection httpstream.Connection
	remote     int32
	requestId  atomic.Int64
	proxy      ProxyFactoryFn
}

func NewStream(connection httpstream.Connection, remote int32) *Stream {
	return &Stream{
		connection: connection,
		remote:     remote,
		requestId:  atomic.Int64{},
		proxy:      NewTwoWayProxyFactory(),
	}
}

// TunnelConnection Taken from the K8S code:
// https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/client-go
// /tools/portforward/portforward.go#L348
// small changes were made to match behaviour.
func (s *Stream) TunnelConnection(ctx context.Context, conn net.Conn, remotePort int32) error {
	defer func() {
		_ = conn.Close()
	}()

	localAddr := conn.LocalAddr().String()
	log.Info().Str("localAddr", localAddr).Int32("remotePort", remotePort).Msg("Handling port forward connection")

	s.requestId.Add(1)
	reqId := strconv.Itoa(int(s.requestId.Load()))

	// create error stream
	headers := generateErrorStreamHeaders(remotePort, reqId)
	log.Debug().Msgf("creating error stream for id (%s)", reqId)

	errorStream, err := s.connection.CreateStream(headers)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create error stream")
		return err
	}
	// we're not writing to this stream
	_ = errorStream.Close()

	defer func() {
		log.Debug().Msgf("removing error stream for id (%s)", reqId)
		s.connection.RemoveStreams(errorStream)
	}()

	errorChan := make(chan error)

	go func() {
		message, err := io.ReadAll(errorStream)

		switch {
		case err != nil:
			log.Error().Err(err).Str("localAddr", localAddr).Int32("remotePort", remotePort).
				Msg("Error reading from error stream")

			errorChan <- fmt.Errorf(
				"error reading from error stream for port %s -> %d: %v", localAddr, remotePort, err)
		case len(message) > 0:
			log.Error().Str("localAddr", localAddr).Int32("remotePort", remotePort).
				Str("errorMessage", string(message)).Msg("Error occurred forwarding")

			errorChan <- fmt.Errorf(
				"an error occurred forwarding %s -> %d: %v", localAddr, remotePort, string(message))
		}

		close(errorChan)
	}()

	// create data stream
	headers.Set(v1.StreamType, v1.StreamTypeData)

	log.Debug().Msgf("creating data stream for id (%s)", reqId)

	dataStream, err := s.connection.CreateStream(headers)
	if err != nil {
		log.Error().Err(err).Str("localAddr", localAddr).Int32("remotePort", remotePort).
			Msg("Error creating forwarding stream")
		runtime.HandleError(fmt.Errorf("error creating forwarding stream for port %s -> %d: %v",
			localAddr, remotePort, err))

		return err
	}

	defer func() {
		log.Debug().Msgf("removing data stream for id (%s)", reqId)
		s.connection.RemoveStreams(dataStream)
	}()

	ping, err := keepalive.NewIdle(
		conn, dataStream, stream.GetInstance().GetPingTime(), stream.GetInstance().GetIdleTime())
	if err != nil {
		log.Error().Err(err).Str("idle", localAddr).Int32("remotePort", remotePort).
			Msg("Error creating idle keepalive")
		runtime.HandleError(err)

		_ = s.connection.Close()

		return err
	}

	proxy := s.proxy(dataStream, conn, ping)

	// Start the proxy
	_ = proxy.Start(ctx)
	// always expect something on errorChan (it may be nil)
	err = <-errorChan
	if err != nil {
		log.Error().Err(err).Msg("Error occurred during port forwarding")
		runtime.HandleError(err)

		_ = s.connection.Close()
	}

	log.Info().Str("localAddr", localAddr).Int32("remotePort", remotePort).Msg("Port forwarding completed")

	return nil
}

func generateErrorStreamHeaders(remote int32, id string) http.Header {
	headers := http.Header{}
	headers.Set(v1.StreamType, v1.StreamTypeError)
	headers.Set(v1.PortHeader, fmt.Sprintf("%d", remote))
	headers.Set(v1.PortForwardRequestIDHeader, id)

	return headers
}
