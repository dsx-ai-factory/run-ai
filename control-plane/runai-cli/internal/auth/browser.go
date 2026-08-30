package auth

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/pkg/browser"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
)

const (
	contentType         = "Content-Type"
	Http                = "http"
	Get                 = "GET"
	serverNotReady      = "server is not ready"
	checkServerInterval = 200 * time.Millisecond
)

type codeExchangeResult struct {
	Code string
	Err  error
}

type ExchangeCodeServerGetter interface {
	Start(ctx context.Context) error
	GetExchangeCode(ctx context.Context) (string, error)
	SetExpectedState(state string)
}

type BrowserAuthExchangeCodeHandler struct {
	localWebserver ExchangeCodeServerGetter
	printOnly      bool
}

func NewBrowserAuthExchangeCodeHandler(localWebserver ExchangeCodeServerGetter, printOnly bool) *BrowserAuthExchangeCodeHandler {
	return &BrowserAuthExchangeCodeHandler{
		localWebserver: localWebserver,
		printOnly:      printOnly,
	}
}

func (b *BrowserAuthExchangeCodeHandler) Handle(ctx context.Context, authUrl *url.URL) (string, error) {
	if expectedState := authUrl.Query().Get("state"); expectedState != "" {
		b.localWebserver.SetExpectedState(expectedState)
	}

	serverCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	err := b.localWebserver.Start(serverCtx)
	if err != nil {
		return "", err
	}
	if !b.printOnly {
		err = browser.OpenURL(authUrl.String())
		if err != nil {
			return "", fmt.Errorf(consts.OpenBrowserError, err)
		}
	} else {
		fmt.Printf(consts.PrintUrlBrowserModeMsg, authUrl.String())
	}

	code, err := b.localWebserver.GetExchangeCode(ctx)
	if err != nil {
		return "", err
	}
	return code, nil
}
