package list

import (
	"context"
	"fmt"
	"time"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cli/structs"
	wcgenerated "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

// TemplateLister defines the interface for fetching templates from the API
type TemplateLister interface {
	List(ctx context.Context, filterBy []string, includeAncestors bool, limit, offset *int32) (*wcgenerated.TemplatesListResponse, error)
}

// Service handles template list operations
type Service struct {
	lister    TemplateLister
	clusterId string
}

// NewService creates a new template list service
func NewService(lister TemplateLister, clusterId string) *Service {
	return &Service{
		lister:    lister,
		clusterId: clusterId,
	}
}

// GetListParams holds parameters for fetching template list
type GetListParams struct {
	ProjectId string
	All       bool
	Limit     *int32
	Offset    *int32
}

// GetTemplateList fetches templates and returns displayable output
func (s *Service) GetTemplateList(ctx context.Context, params GetListParams) (*Output, error) {
	var filterBy []string

	// Add project filter if not using -A flag
	var includeAncestors bool
	if !params.All && params.ProjectId != "" {
		filterBy = append(filterBy, "projectId=="+params.ProjectId)
		includeAncestors = true
	}

	// Call the thin client with explicit parameters
	response, err := s.lister.List(ctx, filterBy, includeAncestors, params.Limit, params.Offset)
	if err != nil {
		return nil, err
	}

	// Extract templates from raw API response
	templates := response.GetEntries()

	// Convert NextOffset to NextToken if available (Service layer transformation)
	var nextToken *structs.NextToken
	if pagination := response.Pagination; pagination != nil {
		if nextOffset := pagination.NextOffset; nextOffset != nil && *nextOffset > 0 {
			nextOffsetInt32 := int32(*nextOffset)
			nextToken, err = offsetToNextToken(nextOffsetInt32)
			if err != nil {
				return nil, fmt.Errorf("failed to create next token: %w", err)
			}
		}
	}

	return NewOutput(templates, nextToken)
}

// offsetToNextToken bridges the gap between two incompatible pagination systems:
//
// The CLI's pagination infrastructure (NextToken) expects timestamps for pagination:
//   - Used by workloads, projects, etc: "Get items created after 2024-11-11T10:30:00Z"
//   - NextToken internally stores a time.Time and encodes it as base64 RFC3339Nano
//
// But the Template API uses traditional offset-based pagination:
//   - Template API: "Get 50 items starting at position 100" (offset=100, limit=50)
//   - Returns nextOffset=150 to continue pagination
//
// This adapter allows offset-based Template API to work with the timestamp-based CLI framework
// by encoding the numeric offset as a Unix timestamp. This is a harmless encoding trick since:
//   - We only need to store/retrieve the number, not use it as an actual timestamp
//   - It allows templates to use the same --next-token flag as other CLI commands
//
// Example workflow:
//  1. Template API returns: nextOffset = 100
//  2. We convert 100 → Unix timestamp 100 seconds after epoch = "1970-01-01T00:01:40Z"
//  3. NextToken base64-encodes this timestamp → "MTk3MC0wMS0wMVQwMDowMTo0MFo="
//  4. User sees in CLI output: NEXTTOKEN: MTk3MC0wMS0wMVQwMDowMTo0MFo=
//  5. User runs: runai list template --next-token MTk3MC0wMS0wMVQwMDowMTo0MFo=
//  6. We decode token → "1970-01-01T00:01:40Z" → Unix seconds = 100 → offset = 100
//  7. We call Template API with offset=100 to get next page
func offsetToNextToken(offset int32) (*structs.NextToken, error) {
	// Validate that offset is non-negative
	// Valid tokens from our system: Always positive, never hit this error
	if offset < 0 {
		return nil, fmt.Errorf("offset cannot be negative: %d", offset)
	}

	// Encode offset as Unix epoch + offset seconds (e.g., 100 → Jan 1 1970, 00:01:40 UTC)
	timestamp := time.Unix(int64(offset), 0).UTC()
	return structs.NextTokenFromTimestamp(timestamp)
}

// NextTokenToOffset extracts the numeric offset from a NextToken
//
// This reverses the encoding done by offsetToNextToken, converting the fake timestamp
// back to the original offset number that the Template API expects.
//
// Example: When user provides --next-token MTk3MC0wMS0wMVQwMDowMTo0MFo=
//  1. NextToken decodes base64 → "1970-01-01T00:01:40Z"
//  2. We extract timestamp → time.Time representing Jan 1 1970, 00:01:40 UTC
//  3. We call timestamp.Unix() → 100 seconds since epoch
//  4. Return offset=100 to pass to Template API
func NextTokenToOffset(token *structs.NextToken) (*int32, error) {
	if token == nil || !token.IsSet() {
		return nil, nil
	}

	// Extract timestamp and convert back to offset (Unix seconds since epoch)
	timestamp := token.GetTimestamp()
	unixSeconds := timestamp.Unix()

	// Validate that the offset is non-negative
	// Valid tokens (from our system): Always positive, never hit this error
	// This check only catches invalid tokens
	if unixSeconds < 0 {
		return nil, fmt.Errorf("invalid token: offset cannot be negative")
	}

	offset := int32(unixSeconds)

	return &offset, nil
}
