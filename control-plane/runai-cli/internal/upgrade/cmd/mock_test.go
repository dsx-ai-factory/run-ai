package cmd_test

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/cli"
)

type mockVersionChecker struct {
	hasUpdate  bool
	version    string
	checkErr   error
	upgradeErr error
}

func (m *mockVersionChecker) Check(_ context.Context) (bool, string, error) {
	return m.hasUpdate, m.version, m.checkErr
}

func mockUpgrader(checker *mockVersionChecker) func(ctx context.Context, cliClient *cli.Client, shell string, quit bool) (string, error) {
	return func(ctx context.Context, cliClient *cli.Client, shell string, quit bool) (string, error) {
		if checker.upgradeErr != nil {
			return "", checker.upgradeErr
		}
		return "success", nil
	}
}

type mockConfig struct {
	autoUpdate     bool
	autoUpdateInit bool
	updateShell    string
}

func (m *mockConfig) GetAutoUpdate() bool    { return m.autoUpdate }
func (m *mockConfig) IsAutoUpdateInit() bool { return m.autoUpdateInit }
func (m *mockConfig) GetUpdateShell() string { return m.updateShell }

type mockPrinter struct {
	messages []string
}

func (m *mockPrinter) Println(a ...interface{}) {
	m.messages = append(m.messages, fmt.Sprint(a...))
}

func (m *mockPrinter) Printf(format string, a ...interface{}) {
	m.messages = append(m.messages, fmt.Sprintf(format, a...))
}
