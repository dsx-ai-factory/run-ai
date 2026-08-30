package version

import (
	"context"
	"errors"
	"strings"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	clierrors "github.com/run-ai/runai/control-plane/runai-cli/internal/errors"
	openapi "github.com/run-ai/runai/runai-clients/cli-exposer-client/generated"
)

const (
	NotAvailableVersion = "N/A"
)

// CLIVersionTooHighError represents an error when the CLI version is higher than the control plane version
type CLIVersionTooHighError struct {
	Message             string
	ControlPlaneVersion string
}

func (e *CLIVersionTooHighError) Error() string {
	return e.Message
}

type CliVersionGetter interface {
	GetVersion(ctx context.Context) (*openapi.Version, error)
}

type CLIChecker struct {
	getter CliVersionGetter
}

func NewCLIChecker(getter CliVersionGetter) *CLIChecker {
	return &CLIChecker{
		getter: getter,
	}
}

func (c *CLIChecker) Check(ctx context.Context) (bool, string, error) {
	version, err := c.getter.GetVersion(ctx)
	if err != nil {
		var httpError *clierrors.HttpError
		if errors.As(err, &httpError) && strings.Contains(httpError.Message, consts.VersionMismatchPrefix) {
			cpVersion := NotAvailableVersion
			if version != nil && version.GetCpVersion() != "" {
				cpVersion = version.GetCpVersion()
			}
			return false, cpVersion, &CLIVersionTooHighError{Message: httpError.Message, ControlPlaneVersion: cpVersion}
		}
		return false, "failed to get version", err
	}
	if version == nil {
		return false, "", errors.New("version does not exist")
	}
	if version.GetNeedUpgrade() {
		return true, version.GetCpVersion(), nil
	}
	return false, version.GetCpVersion(), nil
}
