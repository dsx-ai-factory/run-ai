package notifications

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdminMessage_Validate_WithStruct(t *testing.T) {
	adminMsg := &AdminMessage{}

	tests := []struct {
		name     string
		input    AdminMessage
		expected bool
	}{
		{
			name:     "valid empty message",
			input:    AdminMessage{Msg: ""},
			expected: true,
		},
		{
			name:     "valid simple message",
			input:    AdminMessage{Msg: "Hello world"},
			expected: true,
		},
		{
			name:     "valid message with allowed HTML tags",
			input:    AdminMessage{Msg: "<p>Hello <strong>world</strong></p>"},
			expected: true,
		},
		{
			name:     "valid message with link",
			input:    AdminMessage{Msg: "<a href='https://example.com'>Click here</a>"},
			expected: true,
		},
		{
			name:     "invalid message with script tag",
			input:    AdminMessage{Msg: "<script>alert('xss')</script>"},
			expected: false,
		},
		{
			name:     "invalid message with onclick",
			input:    AdminMessage{Msg: "<button onclick='alert()'>Click</button>"},
			expected: false,
		},
		{
			name:     "invalid message with javascript protocol",
			input:    AdminMessage{Msg: "<a href='javascript:alert()'>Click</a>"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adminMsg.Validate(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAdminMessage_Validate_WithPointer(t *testing.T) {
	adminMsg := &AdminMessage{}

	input := &AdminMessage{
		ID:            "test-id",
		Msg:           "Test message",
		CreatedAt:     1706886400000,
		EnableDismiss: true,
	}

	result := adminMsg.Validate(input)
	assert.True(t, result)
}

func TestAdminMessage_Validate_WithMap(t *testing.T) {
	adminMsg := &AdminMessage{}

	tests := []struct {
		name     string
		input    map[string]interface{}
		expected bool
	}{
		{
			name: "valid map with message",
			input: map[string]interface{}{
				"id":            "test-id",
				"msg":           "Hello world",
				"createdAt":     float64(1706886400000),
				"enableDismiss": true,
			},
			expected: true,
		},
		{
			name: "valid map with empty message",
			input: map[string]interface{}{
				"msg": "",
			},
			expected: true,
		},
		{
			name: "invalid map with script tag",
			input: map[string]interface{}{
				"msg": "<script>alert('xss')</script>",
			},
			expected: false,
		},
		{
			name:     "map without msg field",
			input:    map[string]interface{}{},
			expected: true, // empty msg is valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adminMsg.Validate(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAdminMessage_Validate_WithJSONString(t *testing.T) {
	adminMsg := &AdminMessage{}

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid JSON string",
			input:    `{"msg":"Hello world","id":"test-id","createdAt":1706886400000,"enableDismiss":false}`,
			expected: true,
		},
		{
			name:     "valid JSON string with empty message",
			input:    `{"msg":"","id":"test-id"}`,
			expected: true,
		},
		{
			name:     "valid JSON string with HTML",
			input:    `{"msg":"<p>Hello <strong>world</strong></p>"}`,
			expected: true,
		},
		{
			name:     "invalid JSON string with script tag",
			input:    `{"msg":"<script>alert('xss')</script>"}`,
			expected: false,
		},
		{
			name:     "invalid JSON string with onclick",
			input:    `{"msg":"<button onclick='alert()'>Click</button>"}`,
			expected: false,
		},
		{
			name:     "invalid JSON - not parseable",
			input:    `not valid json`,
			expected: false,
		},
		{
			name:     "empty JSON object",
			input:    `{}`,
			expected: true, // empty msg is valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adminMsg.Validate(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAdminMessage_Validate_WithInvalidType(t *testing.T) {
	adminMsg := &AdminMessage{}

	tests := []struct {
		name  string
		input interface{}
	}{
		{
			name:  "integer",
			input: 123,
		},
		{
			name:  "boolean",
			input: true,
		},
		{
			name:  "slice",
			input: []string{"a", "b"},
		},
		{
			name:  "nil",
			input: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adminMsg.Validate(tt.input)
			assert.False(t, result)
		})
	}
}

func TestAdminMessage_TransformForMigration_WithMap(t *testing.T) {
	adminMsg := &AdminMessage{}

	input := map[string]interface{}{
		"id":            "test-id",
		"msg":           "Test message",
		"createdAt":     float64(1706886400000),
		"enableDismiss": true,
	}

	result := adminMsg.TransformForMigration(input)

	transformed, ok := result.(AdminMessage)
	assert.True(t, ok, "result should be of type AdminMessage")
	assert.Equal(t, "test-id", transformed.ID)
	assert.Equal(t, "Test message", transformed.Msg)
	assert.Equal(t, int64(1706886400000), transformed.CreatedAt)
	assert.True(t, transformed.EnableDismiss)
}

func TestAdminMessage_TransformForMigration_WithJSONString(t *testing.T) {
	adminMsg := &AdminMessage{}

	input := `{"id":"test-id","msg":"Test message","createdAt":1706886400000,"enableDismiss":true}`

	result := adminMsg.TransformForMigration(input)

	transformed, ok := result.(AdminMessage)
	assert.True(t, ok, "result should be of type AdminMessage")
	assert.Equal(t, "test-id", transformed.ID)
	assert.Equal(t, "Test message", transformed.Msg)
	assert.Equal(t, int64(1706886400000), transformed.CreatedAt)
	assert.True(t, transformed.EnableDismiss)
}

func TestAdminMessage_TransformForMigration_WithPartialData(t *testing.T) {
	adminMsg := &AdminMessage{}

	tests := []struct {
		name     string
		input    map[string]interface{}
		expected AdminMessage
	}{
		{
			name: "only msg field",
			input: map[string]interface{}{
				"msg": "Hello",
			},
			expected: AdminMessage{
				Msg: "Hello",
			},
		},
		{
			name: "only id and enableDismiss",
			input: map[string]interface{}{
				"id":            "some-id",
				"enableDismiss": true,
			},
			expected: AdminMessage{
				ID:            "some-id",
				EnableDismiss: true,
			},
		},
		{
			name:     "empty map",
			input:    map[string]interface{}{},
			expected: AdminMessage{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adminMsg.TransformForMigration(tt.input)
			transformed, ok := result.(AdminMessage)
			assert.True(t, ok, "result should be of type AdminMessage")
			assert.Equal(t, tt.expected, transformed)
		})
	}
}

func TestAdminMessage_TransformForMigration_WithInvalidInput(t *testing.T) {
	adminMsg := &AdminMessage{}

	tests := []struct {
		name  string
		input interface{}
	}{
		{
			name:  "integer",
			input: 123,
		},
		{
			name:  "boolean",
			input: true,
		},
		{
			name:  "invalid JSON string",
			input: "not valid json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adminMsg.TransformForMigration(tt.input)
			assert.Equal(t, tt.input, result, "invalid input should be returned as-is")
		})
	}
}

func TestAdminMessage_TransformForMigration_RealisticMigrationScenario(t *testing.T) {
	adminMsg := &AdminMessage{}

	// This is the exact format that comes from the legacy backend during migration
	legacyJSON := `{"msg":"old message","id":"4b715a1e-a857-4c8f-86f6-c5a14592ef41","createdAt":1770029489685,"enableDismiss":false}`

	result := adminMsg.TransformForMigration(legacyJSON)

	transformed, ok := result.(AdminMessage)
	assert.True(t, ok, "result should be of type AdminMessage")
	assert.Equal(t, "4b715a1e-a857-4c8f-86f6-c5a14592ef41", transformed.ID)
	assert.Equal(t, "old message", transformed.Msg)
	assert.Equal(t, int64(1770029489685), transformed.CreatedAt)
	assert.False(t, transformed.EnableDismiss)

	// Also verify validation passes
	isValid := adminMsg.Validate(legacyJSON)
	assert.True(t, isValid)
}

func TestValidateHTMLContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "plain text",
			content:  "Hello world",
			expected: true,
		},
		{
			name:     "allowed tags",
			content:  "<p><strong>Bold</strong> and <em>italic</em></p>",
			expected: true,
		},
		{
			name:     "allowed link with https",
			content:  "<a href='https://example.com'>Link</a>",
			expected: true,
		},
		{
			name:     "allowed link with mailto",
			content:  "<a href='mailto:test@example.com'>Email</a>",
			expected: true,
		},
		{
			name:     "disallowed script tag",
			content:  "<script>alert('xss')</script>",
			expected: false,
		},
		{
			name:     "disallowed iframe tag",
			content:  "<iframe src='evil.com'></iframe>",
			expected: false,
		},
		{
			name:     "disallowed onclick handler",
			content:  "<div onclick='alert()'>Click</div>",
			expected: false,
		},
		{
			name:     "disallowed onerror handler",
			content:  "<img onerror='alert()' src='x'>",
			expected: false,
		},
		{
			name:     "disallowed javascript protocol",
			content:  "<a href='javascript:void(0)'>Click</a>",
			expected: false,
		},
		{
			name:     "disallowed vbscript protocol",
			content:  "<a href='vbscript:msgbox'>Click</a>",
			expected: false,
		},
		{
			name:     "disallowed data protocol",
			content:  "<a href='data:text/html,<script>alert(1)</script>'>Click</a>",
			expected: false,
		},
		{
			name:     "empty content",
			content:  "",
			expected: true,
		},
		{
			name:     "font tag allowed",
			content:  "<font color='red'>Red text</font>",
			expected: true,
		},
		{
			name:     "list tags allowed",
			content:  "<ul><li>Item 1</li><li>Item 2</li></ul>",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateHTMLContent(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAdminMessage_SetDefault(t *testing.T) {
	adminMsg := &AdminMessage{}
	adminMsg.SetDefault()

	assert.NotEmpty(t, adminMsg.ID, "ID should be set")
	assert.Equal(t, "", adminMsg.Msg)
	assert.Equal(t, int64(0), adminMsg.CreatedAt)
	assert.False(t, adminMsg.EnableDismiss)
}

func TestAdminMessage_Metadata(t *testing.T) {
	adminMsg := &AdminMessage{}

	assert.Equal(t, "Notifications", adminMsg.Category())
	assert.Equal(t, "adminMessage", adminMsg.Key())
	assert.Equal(t, "Admin message", adminMsg.DisplayName())
	assert.Contains(t, adminMsg.Description(), "message will be displayed")
}
