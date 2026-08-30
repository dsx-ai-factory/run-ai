package cmdtxn

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"

	clierrors "github.com/run-ai/runai/control-plane/runai-cli/internal/errors"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction/reporter"

	"github.com/spf13/cobra"
)

const (
	MissingParamsErrorMsg = "cmd missing params"

	complianceWithField    = "Field: %s, Details: %s"
	complianceWithoutField = "Details: %s"
)

func CmdErrorPrintAndSet(cmd *cobra.Command, msg string, err error) error {
	errStr := handleErrorsMsg(err)
	cmd.PrintErrf("Error: %s. %v\n", msg, errStr)
	return reporter.ReportCmd(cmd, err)
}

func constructErrorMsg(message clierrors.ApiGeneralErrorMessage) string {
	sb := &strings.Builder{}
	hasCompliance := len(message.Compliance) > 0
	// Skip the top-level message when compliance issues already contain the same text with more context (e.g. field name)
	if !hasCompliance || !complianceAlreadyCoversMessage(message) {
		sb.WriteString(message.Message)
	}
	if hasCompliance {
		messageWithCompliance(sb, message)
	}
	if message.Details != "" {
		sb.WriteString("\n")
		sb.WriteString(message.Details)
	}
	return sb.String()
}

func complianceAlreadyCoversMessage(message clierrors.ApiGeneralErrorMessage) bool {
	for _, issue := range message.Compliance {
		if issue.GetDetails() != message.Message {
			return false
		}
	}
	return true
}

func constructErrorMsgs(messages []clierrors.ApiGeneralErrorMessage) string {
	parts := make([]string, len(messages))
	for i, msg := range messages {
		parts[i] = constructErrorMsg(msg)
	}
	return strings.Join(parts, "\n")
}

func handleErrorsMsg(err error) string {
	var httpErr *clierrors.HttpError
	if errors.As(err, &httpErr) {
		if httpErr.Status == "403 Forbidden" {
			return consts.PermissionDenied
		}
		var message clierrors.ApiGeneralErrorMessage
		if err := json.Unmarshal([]byte(httpErr.Message), &message); err == nil {
			if errMsg := constructErrorMsg(message); errMsg != "" {
				return errMsg
			}
			return message.Message
		}
		var messages []clierrors.ApiGeneralErrorMessage
		if err := json.Unmarshal([]byte(httpErr.Message), &messages); err == nil {
			return constructErrorMsgs(messages)
		}
		return httpErr.Message
	}

	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		var netErr *net.OpError
		if errors.As(urlErr.Err, &netErr) {
			var dnsErr *net.DNSError
			if errors.As(netErr.Err, &dnsErr) {
				return fmt.Sprintf(consts.BadUrlMsg, dnsErr.Name, dnsErr.Err)
			}
		}

		if urlErr.Timeout() {
			return fmt.Sprintf(consts.ClusterTimedOutMsg, urlErr.URL, urlErr.Err)
		}
	}
	return err.Error()
}

func messageWithCompliance(builder *strings.Builder, message clierrors.ApiGeneralErrorMessage) {
	if builder == nil {
		return
	}
	for _, compliance := range message.Compliance {
		builder.WriteString("\n")
		if compliance.GetField() != "" {
			fmt.Fprintf(builder, complianceWithField, compliance.Field, compliance.Details)
		} else {
			fmt.Fprintf(builder, complianceWithoutField, compliance.Details)
		}
	}
}

func CmdErrorValidationMessage(cmd *cobra.Command, msg string, err error) error {
	errStr := handleErrorsMsg(err)
	cmd.PrintErrf(msg, errStr)
	cmd.Println()
	return reporter.ReportCmd(cmd, err)
}

func CmdErrorVersionMessage(cmd *cobra.Command, err error) error {
	cmd.PrintErrln(err)
	return reporter.ReportCmd(cmd, err)
}

func CmdErrorPrintAndSetSuffix(cmd *cobra.Command, msg string, err error, suffix string) error {
	cmd.PrintErrf("%s\nError: %v%s", msg, err, suffix)
	return reporter.ReportCmd(cmd, err)
}

func CmdErrorSetMissingParams(cmd *cobra.Command) error {
	return reporter.ReportCmd(cmd, errors.New(MissingParamsErrorMsg))
}

func CmdErrorPrintRaw(cmd *cobra.Command, msg string, err error) error {
	cmd.PrintErrf("Error: %v.\n\n%s", err, msg)
	return reporter.ReportCmd(cmd, err)
}
