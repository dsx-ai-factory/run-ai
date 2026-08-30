package metrics

import (
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

func TestToolLabel_KnownTool(t *testing.T) {
	known := map[string]bool{"list_workloads": true, "get_project": true}
	assert.Equal(t, "list_workloads", toolLabel(&mcp.CallToolParamsRaw{Name: "list_workloads"}, known))
	assert.Equal(t, "get_project", toolLabel(&mcp.CallToolParamsRaw{Name: "get_project"}, known))
}

func TestToolLabel_UnknownTool(t *testing.T) {
	known := map[string]bool{"list_workloads": true}
	assert.Equal(t, "unknown", toolLabel(&mcp.CallToolParamsRaw{Name: "injected_tool"}, known))
	assert.Equal(t, "unknown", toolLabel(&mcp.CallToolParamsRaw{Name: ""}, known))
}

func TestToolLabel_NilParams(t *testing.T) {
	known := map[string]bool{"list_workloads": true}
	assert.Equal(t, "unknown", toolLabel(nil, known))
}

func TestToolLabel_CallToolParams(t *testing.T) {
	known := map[string]bool{"submit_job": true}
	assert.Equal(t, "submit_job", toolLabel(&mcp.CallToolParams{Name: "submit_job"}, known))
	assert.Equal(t, "unknown", toolLabel(&mcp.CallToolParams{Name: "other"}, known))
}

func TestNormalizeLabel_ShortString(t *testing.T) {
	s := "initialize"
	assert.Equal(t, s, normalizeLabel(s))
}

func TestNormalizeLabel_ExactlyMaxLen(t *testing.T) {
	s := strings.Repeat("a", maxLabelLen)
	assert.Equal(t, s, normalizeLabel(s))
}

func TestNormalizeLabel_LongString(t *testing.T) {
	s := strings.Repeat("x", maxLabelLen+1)
	got := normalizeLabel(s)
	assert.Len(t, got, 57) // 48 prefix + 1 tilde + 8 hex chars
	assert.Contains(t, got, "~")
}

func TestNormalizeLabel_TwoLongStringsAreDistinct(t *testing.T) {
	a := strings.Repeat("a", maxLabelLen+1)
	b := strings.Repeat("b", maxLabelLen+1)
	assert.NotEqual(t, normalizeLabel(a), normalizeLabel(b))
}
