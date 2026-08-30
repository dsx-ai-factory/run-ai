package pseudo

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"go.uber.org/zap"

	"golang.org/x/term"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
)

const (
	queryParamPod       = "pod"
	queryParamContainer = "container"
	queryParamStdin     = "stdin"
	queryParamStdout    = "stdout"
	queryParamStderr    = "stderr"
	queryParamTTY       = "tty"
	queryParamTrue      = "true"
	queryParamWidth     = "width"
	queryParamHeight    = "height"
)

type Options struct {
	WorkloadName string
	Pod          string
	Container    string
	Project      string
	WorkloadType api.WorkloadType

	PodRunning time.Duration
	Stdin      bool
	TTY        bool

	WaitTimeout time.Duration
	Command     string

	// Terminal size
	Width  int
	Height int
}

func NewOptions(pod, container string, stdin, tty bool, cmd ...string) (*Options, error) {
	command, err := EncodeCommandString(cmd)
	if err != nil {
		return nil, err
	}

	options := &Options{
		Pod:       pod,
		Container: container,
		Stdin:     stdin,
		TTY:       tty,
		Command:   command,
	}

	// Set initial terminal size if TTY is enabled and we're in a terminal
	if tty && term.IsTerminal(int(os.Stdin.Fd())) {
		width, height, err := GetTerminalSize()
		if err != nil {
			logger := config.GetInstance().GetLogger()
			logger.Warn("fetching terminal size error", zap.Error(err))
		} else {
			options.Width = width
			options.Height = height
		}
	}

	return options, nil
}

func (o *Options) ToQueryParams() (*url.Values, error) {
	vars := url.Values{}
	vars.Set(queryParamStdout, queryParamTrue)
	vars.Set(queryParamStderr, queryParamTrue)

	if o.Pod != "" {
		vars.Set(queryParamPod, o.Pod)
	}
	if o.Container != "" {
		vars.Set(queryParamContainer, o.Container)
	}
	if o.Stdin {
		vars.Set(queryParamStdin, queryParamTrue)
	}
	if o.TTY {
		vars.Set(queryParamTTY, queryParamTrue)
		// Add terminal size parameters when TTY is enabled
		if o.Width > 0 {
			vars.Set(queryParamWidth, strconv.Itoa(int(o.Width)))
		}
		if o.Height > 0 {
			vars.Set(queryParamHeight, strconv.Itoa(int(o.Height)))
		}
	}
	return &vars, nil
}

func EncodeCommandString(cmd []string) (string, error) {
	strByte, err := json.Marshal(cmd)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(strByte), nil
}
