package forward

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/status"

	ioproxy "github.com/run-ai/runai/control-plane/runai-cli/internal/io"

	clusterUrl "github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/url"

	"go.uber.org/zap"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/apis"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/portforward"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/portforward/options"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/websocket"
)

const (
	DefaultSleepTimeBetweenRetries = 5 * time.Second
	TCPProtocol                    = "tcp"

	endpoint       = "port-forward"
	queryParamPort = "port"
)

type Forward struct {
	token           string
	local           int32
	remote          int32
	listener        net.Listener
	options         *options.Options
	workload        *apis.Workload
	url             string
	baseUrl         *url.URL
	clusterMetadata *status.Metadata
}

func NewForward(token string, options *options.Options, local, remote int32, workload *apis.Workload, url string, clusterMetadata *status.Metadata) *Forward {
	return &Forward{
		token:           token,
		local:           local,
		remote:          remote,
		options:         options,
		workload:        workload,
		url:             url,
		clusterMetadata: clusterMetadata,
	}
}

func (f *Forward) Close() {
	if f.listener != nil {
		_ = f.listener.Close()
	}
}

func (f *Forward) GetLocalPort() int32 {
	return f.local
}

func (f *Forward) GetRemotePort() int32 {
	return f.remote
}

func (f *Forward) BindLocalPort(ctx context.Context) error {
	var err error
	var lc net.ListenConfig
	f.listener, err = lc.Listen(ctx, TCPProtocol, getAddressUrl(f.options.Address, f.local))
	if err != nil {
		return err
	}
	err = f.buildUrl()
	if err != nil {
		return err
	}
	return nil
}

func (f *Forward) Run(ctx context.Context) error {
	for {
		local, err := f.listener.Accept()
		if err != nil {
			return err
		}
		logger.Debug("connection created", zap.Int32("local", f.local), zap.String("loop", local.RemoteAddr().String()))

		dialer := websocket.NewTransportDialer(f.token)
		stream := portforward.NewStream(dialer, f.url, DefaultSleepTimeBetweenRetries, DefaultSleepTimeBetweenRetries, f.clusterMetadata)
		remote, err := stream.Stream(ctx, f.baseUrl)
		if err != nil {
			logger.Error("connection stream error", zap.Int32("local", f.local), zap.Error(err))
			return err
		}

		proxy := ioproxy.NewTwoWayProxy(time.Second)
		// Handle the actual forwarding to the remote
		go func() {
			defer logger.Debug("connection closed", zap.Int32("local", f.local), zap.String("loop", local.RemoteAddr().String()))
			_ = proxy.Start(ctx, remote, local)
		}()
	}
}

func (f *Forward) buildUrl() error {
	var err error
	var query *url.Values

	if f.options != nil {
		query, err = f.options.ToQueryParams()
		if err != nil {
			return err
		}
		query.Set(queryParamPort, strconv.FormatInt(int64(f.remote), 10))
	}
	f.baseUrl, err = clusterUrl.BuildClusterUrl(f.url, f.workload, endpoint, query)
	return err
}

func getAddressUrl(a string, p int32) string {
	return fmt.Sprintf("%s:%d", a, p)
}
