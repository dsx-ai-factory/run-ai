package services

import (
	"context"
	"fmt"
	"github.com/run-ai/runai/control-plane/cli-exposer/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/cli-exposer/pkg/consts"
	"github.com/run-ai/runai/control-plane/cli-exposer/pkg/document"
	"github.com/run-ai/runai/control-plane/cli-exposer/pkg/installer"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/version"
	middlewares "github.com/run-ai/runai/runai-application-packages/old-auth-token-middleware"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"golang.org/x/mod/semver"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type RegexValidators struct {
	manualDocumentRegex *regexp.Regexp
	userAgentRegex      *regexp.Regexp
}

// CliApiService is a service that implements the logic for the CliApiService
// This service should implement the business logic for every endpoint for the CLI API.
// Include any external packages or services that will be required by this service.
type CliApiService struct {
	validators  *RegexValidators
	version     version.Version
	docTemplate document.HtmlContent
}

// NewCliApiService creates a default api service
func NewCliApiService(validators *RegexValidators, cliVersion version.Version,
	htmlContent document.HtmlContent) api.ResearcherCommandLineInterfaceAPIServicer {
	return &CliApiService{
		validators:  validators,
		version:     cliVersion,
		docTemplate: htmlContent,
	}
}

// GetVersion  - version check
func (cas *CliApiService) GetVersion(ctx context.Context) (api.ImplResponse, error) {
	cliServerVersion := cas.version.Version
	cliServerVersionV := cliServerVersion
	if !strings.HasPrefix(cliServerVersion, "v") {
		cliServerVersionV = fmt.Sprintf(consts.VersionPrefix, cliServerVersion)
	}

	if !semver.IsValid(cliServerVersionV) {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf(consts.ErrInternalCliVersion)
		return api.Response(http.StatusInternalServerError, consts.ErrInternalCliVersion), nil
	}

	userAgent, err := getUserAgentHeader(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf(consts.ErrorUserAgent)
		return api.Response(http.StatusBadRequest, consts.ErrorUserAgent), nil
	}

	versionRe := cas.validators.userAgentRegex.FindStringSubmatch(userAgent)
	if len(versionRe) < 2 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf(consts.ErrClientCliVersion)
		return api.Response(http.StatusBadRequest, consts.ErrClientCliVersion), nil
	}

	clientVersionStr := versionRe[1]
	clientVersionStrV := clientVersionStr
	if !strings.HasPrefix(clientVersionStr, "v") {
		clientVersionStrV = fmt.Sprintf(consts.VersionPrefix, clientVersionStr)
	}

	if !semver.IsValid(clientVersionStrV) {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf(consts.ErrClientCliVersion)
		return api.Response(http.StatusBadRequest, consts.ErrClientCliVersion), nil
	}

	needUpdate := false
	if !strings.HasSuffix(clientVersionStrV, "-dev") && !strings.HasSuffix(clientVersionStrV, "-local") {
		if semver.Compare(cliServerVersionV, clientVersionStrV) < 0 {
			msg := fmt.Sprintf(consts.ErrCliVersionMismatch, clientVersionStr, cliServerVersion)
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", msg)
			return api.Response(http.StatusBadRequest, msg), nil
		}
		if semver.Compare(cliServerVersionV, clientVersionStrV) > 0 {
			needUpdate = true
		}
	}

	response := api.Version{
		SdkVersion:  cliServerVersion,
		NeedUpgrade: needUpdate,
		CpVersion:   cas.version.ControlPlaneVersion,
	}

	return api.Response(http.StatusOK, response), nil
}

// GetBinary - get binary CLI file
func (cas *CliApiService) GetBinary(ctx context.Context, operationSystem string, arch string) (api.ImplResponse, error) {
	filePath, err := installer.ValidateBinaryPath(ctx, operationSystem, arch)
	if err != nil {
		return api.Response(http.StatusNotFound, err.Error()), nil
	}
	return api.Response(http.StatusOK, filePath), nil
}

// GetInstallerWindows - get installer script for Windows
func (cas *CliApiService) GetInstallerWindows(ctx context.Context) (api.ImplResponse, error) {
	body := []byte("Windows installer is not supported yet")
	return api.Response(http.StatusOK, body), nil
}

func getControlPlaneURL(ctx context.Context) (string, error) {
	var cpURL string
	var ok bool

	//Get the CP URL from the context
	httpServer := ctx.Value(consts.ContextHostUrl)
	if cpURL, ok = httpServer.(string); !ok {
		return "", fmt.Errorf(consts.ErrorGettingCpURL)
	}

	if strings.HasPrefix(cpURL, consts.LocalhostPrefix) {
		return fmt.Sprintf("%s%s", "http://", cpURL), nil
	}
	return fmt.Sprintf("%s%s", "https://", cpURL), nil
}

func getUserAgentHeader(ctx context.Context) (string, error) {
	var userAgent string
	var ok bool

	if userAgent, ok = ctx.Value(consts.UserAgentHeader).(string); !ok {
		return "", fmt.Errorf(consts.ErrorUserAgent)
	}
	return userAgent, nil
}

func getJWTToken(ctx context.Context) (string, error) {
	var token string
	var ok bool

	// Get the JWT token from the context
	if token, ok = ctx.Value(middlewares.ContextJwtToken).(string); !ok {
		return "", fmt.Errorf(consts.ErrMissingToken)
	}
	return token, nil
}

func (cas *CliApiService) GetInstallerWindowsCommands(ctx context.Context) (api.ImplResponse, error) {
	//cpURL, err := getControlPlaneURL(ctx)
	//if err != nil {
	//	return api.Response(http.StatusBadRequest, err.Error()), nil
	//}

	//response := api.Command{
	//	Command: fmt.Sprintf(consts.WindowsInstallationCommands, cpURL),
	//}

	return api.Response(http.StatusOK, consts.WindowsInstallationCommands), nil
}

func (cas *CliApiService) GetInstallerUnixCommands(ctx context.Context) (api.ImplResponse, error) {
	return generateUnixCommands(ctx, "unix")
}

// GetInstallerUnix - get installer script for Unix
func (cas *CliApiService) GetInstallerUnix(ctx context.Context) (api.ImplResponse, error) {
	setting := installer.Settings{}
	var err error

	if setting.Token, err = getJWTToken(ctx); err != nil {
		return api.Response(http.StatusBadRequest, err.Error()), nil
	}

	setting.CpURL, err = getControlPlaneURL(ctx)
	if err != nil {
		return api.Response(http.StatusBadRequest, err.Error()), nil
	}
	setting.AuthURL = setting.CpURL
	setting.SdkVersion = cas.version.Version
	setting.CpVersion = cas.version.ControlPlaneVersion
	tplOutput, err := installer.GetInstallerCode(setting)
	if err != nil {
		return api.Response(http.StatusInternalServerError, err.Error()), nil
	}
	return api.Response(http.StatusOK, tplOutput), nil
}

func (cas *CliApiService) GetInstallerMac(ctx context.Context) (api.ImplResponse, error) {
	return cas.GetInstallerUnix(ctx)
}

func (cas *CliApiService) GetInstallerMacCommands(ctx context.Context) (api.ImplResponse, error) {
	return generateUnixCommands(ctx, "mac")
}

func (cas *CliApiService) GetInstallerLinux(ctx context.Context) (api.ImplResponse, error) {
	return cas.GetInstallerUnix(ctx)
}

func (cas *CliApiService) GetInstallerLinuxCommands(ctx context.Context) (api.ImplResponse, error) {
	return generateUnixCommands(ctx, "linux")
}

// GetManualDocument - get manual document
func (cas *CliApiService) GetManualDocument(ctx context.Context, docName string) (api.ImplResponse, error) {
	// Validate the document name by regex
	if !cas.validators.manualDocumentRegex.MatchString(docName) {
		msg := fmt.Sprintf(consts.ErrForbiddenDocument, docName)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("%s", msg)
		return api.Response(http.StatusForbidden, msg), nil
	}

	// Check if the document does not exist
	fullFilePath := filepath.Join(consts.DocumentPath, docName)
	if _, err := os.Stat(fullFilePath); err != nil {
		msg := fmt.Sprintf(consts.ErrDocumentNotExist, docName)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("%s", msg)
		return api.Response(http.StatusNotFound, msg), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Reading document: %s", docName)
	doc, err := os.ReadFile(fullFilePath)
	if err != nil {
		return api.Response(http.StatusInternalServerError, "Error reading file"), nil
	}

	htmlTemplate := document.GenerateDocPage(doc, cas.docTemplate.Css, cas.docTemplate.Js)
	return api.Response(http.StatusOK, htmlTemplate), nil
}

func GetRegexValidators() (*RegexValidators, error) {
	// Define the regex pattern
	manualDocumentPattern := `^[0-9a-zA-Z_\-]+\.md$`
	// Compile the regex pattern
	manualDocumentRe, err := regexp.Compile(manualDocumentPattern)
	if err != nil {
		return nil, fmt.Errorf("error compiling regex pattern: %w", err)
	}

	userAgentPattern := `sdk/(\S+)`
	userAgentRe, err := regexp.Compile(userAgentPattern)
	if err != nil {
		return nil, fmt.Errorf("error compiling regex pattern: %w", err)
	}

	validators := &RegexValidators{
		manualDocumentRegex: manualDocumentRe,
		userAgentRegex:      userAgentRe,
	}

	return validators, nil
}

func generateUnixCommands(ctx context.Context, osType string) (api.ImplResponse, error) {
	token, err := getJWTToken(ctx)
	if err != nil {
		return api.Response(http.StatusBadRequest, err.Error()), nil
	}

	cpURL, err := getControlPlaneURL(ctx)
	if err != nil {
		return api.Response(http.StatusBadRequest, err.Error()), nil
	}

	body := fmt.Sprintf(consts.UnixInstallationCommands, cpURL, osType, token)
	response := api.Command{
		Command: body,
	}

	return api.Response(http.StatusOK, response), nil
}
