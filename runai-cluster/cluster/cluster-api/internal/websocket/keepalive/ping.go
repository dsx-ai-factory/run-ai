/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package keepalive

import (
	"io"
	"time"

	"github.com/rs/zerolog/log"
)

const PingType = "ping"

type Ping struct {
	write  io.WriteCloser
	read   io.ReadCloser
	ticker *time.Ticker
}

func NewPing(write io.WriteCloser, read io.ReadCloser, tick time.Duration) *Ping {
	return &Ping{
		write:  write,
		read:   read,
		ticker: time.NewTicker(tick),
	}
}

func (p *Ping) GetType() string {
	return PingType
}

func (p *Ping) Execute() error {
	_, err := p.write.Write([]byte(""))
	if err != nil {
		log.Debug().Msg("ping closing connection")

		_ = p.read.Close()
		// validate connection closed (might be failed for other reasons)
		_ = p.write.Close()
	}

	return err
}

func (p *Ping) Channel() <-chan time.Time {
	return p.ticker.C
}
