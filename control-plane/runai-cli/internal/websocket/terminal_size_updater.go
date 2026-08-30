package websocket

import (
	"context"
	"encoding/json"
	"time"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/pseudo"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"go.uber.org/zap"
	"nhooyr.io/websocket"
)

const (
	versionStr                = "1.0"
	terminalSizeCheckInterval = 100 * time.Millisecond
)

type TerminalSizeUpdater struct {
	isRunning  bool
	enc        *ExtendedNetConnection
	cancelFunc context.CancelFunc
	lastHeight int
	lastWidth  int
}

func NewTerminalSizeUpdater(enc *ExtendedNetConnection, firstHeight, firstWidth int) *TerminalSizeUpdater {
	tsu := &TerminalSizeUpdater{
		enc:        enc,
		lastHeight: firstHeight,
		lastWidth:  firstWidth,
	}
	return tsu
}

func (tsu *TerminalSizeUpdater) IsRunning() bool {
	return tsu.isRunning
}

func (tsu *TerminalSizeUpdater) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	tsu.cancelFunc = cancel

	tsu.isRunning = true
	go tsu.handleSizeChange(ctx)
	config.GetInstance().GetLogger().Debug("terminal size updater started")
}

func (tsu *TerminalSizeUpdater) Stop() {
	tsu.cancelFunc()
	config.GetInstance().GetLogger().Debug("terminal size updater stopped")
}

func (tsu *TerminalSizeUpdater) handleSizeChange(ctx context.Context) {
	var (
		msg []byte
	)
	logger := config.GetInstance().GetLogger()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			width, height, err := pseudo.GetTerminalSize()
			if err != nil {
				logger.Warn("failed to get terminal size", zap.Error(err))
				continue
			}

			// Only update if size has changed
			if width != tsu.lastWidth || height != tsu.lastHeight {
				// Send the new size to the remote
				sizeMsg := Message{
					Version: versionStr,
					Type:    TypeTerminalSize,
					Params: map[string]any{
						IntMapFieldHeight: height,
						IntMapFieldWidth:  width,
					},
				}
				msg, err = json.Marshal(sizeMsg)
				if err != nil {
					logger.Warn("failed to marshal terminal size message", zap.Error(err))
					continue
				}
				err = tsu.enc.InnerWebSocketConn.Write(ctx, websocket.MessageText, msg)
				if err != nil {
					logger.Warn("failed to send terminal size update", zap.Error(err))
					tsu.cancelFunc()
				} else {
					tsu.lastHeight = height
					tsu.lastWidth = width
				}

			}
			time.Sleep(terminalSizeCheckInterval) // Check every 100ms
		}
	}
}
