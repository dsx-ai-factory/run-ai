package pseudo

import (
	"context"
	"io"
	"os"
	"strings"
	"time"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"go.uber.org/zap"

	"golang.org/x/term"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/apis"
	iostd "github.com/run-ai/runai/control-plane/runai-cli/internal/io"
)

type Streamer interface {
	Stream(ctx context.Context, workload *apis.Workload, options *Options) (io.ReadWriteCloser, error)
}

type Terminal struct {
	streamer Streamer
	output   io.Writer
}

func NewTerminal(streamer Streamer, output io.Writer) *Terminal {
	return &Terminal{
		streamer: streamer,
		output:   output,
	}
}

func (t *Terminal) Exec(ctx context.Context, workload *apis.Workload, options *Options, std *iostd.Std) error {
	stream, err := t.streamer.Stream(ctx, workload, options)
	if err != nil {
		return err
	}

	defer stream.Close()

	if options.Stdin || options.TTY {
		// we need this to enable the slave TTY handle the buffer
		oldState, stateErr := term.MakeRaw(int(os.Stdin.Fd()))
		if stateErr != nil {
			return stateErr
		}
		defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }() // Best effort.

		proxy := iostd.NewTwoWayProxy(time.Second)
		err = proxy.Start(ctx, stream, std)
	} else {
		proxy := iostd.NewOneWayProxy(time.Second)
		err = proxy.Start(ctx, stream, std)
	}

	if err != nil {
		if strings.Contains(err.Error(), "StatusNormalClosure") {
			logger := config.GetInstance().GetLogger()
			logger.Debug("normal connection closer", zap.Error(err))
			return nil
		}
		return err
	}
	return nil
}

func GetTerminalSize() (int, int, error) {
	return term.GetSize(int(os.Stdin.Fd()))
}
