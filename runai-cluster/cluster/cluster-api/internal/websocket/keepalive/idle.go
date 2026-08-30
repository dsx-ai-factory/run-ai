package keepalive

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/websocket"
)

const IdleType = "idle"

type Idle struct {
	enc         *websocket.ExtendedNetConnection
	proxy       websocket.ReadWriteCloser
	ping        *time.Ticker
	idle        time.Duration
	lastPayload time.Time
}

func NewIdle(conn websocket.ReadWriteCloser, proxy websocket.ReadWriteCloser, ping, idle time.Duration) (*Idle, error) {
	enc, ok := conn.(*websocket.ExtendedNetConnection)
	if !ok {
		return nil, fmt.Errorf("idle handler require extended connection")
	}

	return &Idle{enc: enc, proxy: proxy, ping: time.NewTicker(ping), idle: idle, lastPayload: time.Now()}, nil
}

func (i *Idle) GetType() string {
	return IdleType
}

func (i *Idle) Execute() error {
	_, err := i.enc.Write([]byte(""))
	if err != nil {
		log.Debug().Msg("ping closing connection")
		i.close()

		return err
	}

	total := i.enc.ReadSize() + i.enc.WriteSize()
	if total != 0 {
		i.lastPayload = time.Now()
		return nil
	}

	return i.checkIdle()
}

func (i *Idle) Channel() <-chan time.Time {
	return i.ping.C
}

func (i *Idle) Reached() error {
	i.close()
	return fmt.Errorf("idle reach")
}

func (i *Idle) checkIdle() error {
	now := time.Now()
	if now.Before(i.lastPayload.Add(i.idle)) {
		return nil
	}

	return i.Reached()
}

func (i *Idle) close() {
	if i.proxy != nil {
		log.Debug().Msg("proxy is active. closing proxy connection due to idle.")

		_ = i.proxy.Close()
	}

	log.Debug().Msg("closing WebSocket connection due to idle.")

	_ = i.enc.InnerConnection.Close()
}
