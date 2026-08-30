package runaimcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type noInput struct{}

func (noInput) Validate() error { return nil }

type goodInput struct {
	ID string `json:"id" jsonschema:"the resource id to look up"`
}

func (goodInput) Validate() error { return nil }

type missingDescInput struct {
	ID string `json:"id"`
}

func (missingDescInput) Validate() error { return nil }

type emptyOut struct{}

func okHandler[In ToolInput](ctx context.Context, _ In) (emptyOut, error) {
	return emptyOut{}, nil
}

func TestLint_PassesOnWellFormedTools(t *testing.T) {
	errs := Lint(func(r *Registry) {
		Register(r, Tool[noInput, emptyOut]{Name: "list_things", Description: "ok", Handler: okHandler[noInput]})
		Register(r, Tool[goodInput, emptyOut]{Name: "get_thing", Description: "ok", Handler: okHandler[goodInput]})
	})
	assert.Empty(t, errs)
}

func TestLint_FailsOnNonSnakeCaseName(t *testing.T) {
	cases := []string{"ListThings", "list-things", "List_Things", "list__things", "1list", "list_"}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			errs := Lint(func(r *Registry) {
				Register(r, Tool[noInput, emptyOut]{
					Name: name, Description: "ok", Handler: okHandler[noInput],
				})
			})
			require.NotEmpty(t, errs, "expected lint failure for name %q", name)
			assert.Contains(t, errs[0].Error(), "snake_case")
		})
	}
}

func TestLint_FailsOnMissingFieldDescription(t *testing.T) {
	errs := Lint(func(r *Registry) {
		Register(r, Tool[missingDescInput, emptyOut]{
			Name: "get_thing", Description: "ok", Handler: okHandler[missingDescInput],
		})
	})
	require.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), `"id"`)
}

func TestLint_EmptyInputStructPasses(t *testing.T) {
	errs := Lint(func(r *Registry) {
		Register(r, Tool[noInput, emptyOut]{
			Name: "list_things", Description: "ok", Handler: okHandler[noInput],
		})
	})
	assert.Empty(t, errs)
}

func TestLint_CollectsMultipleViolations(t *testing.T) {
	errs := Lint(func(r *Registry) {
		Register(r, Tool[missingDescInput, emptyOut]{
			Name: "BadName", Description: "ok", Handler: okHandler[missingDescInput],
		})
	})
	// One for snake_case, one for the missing field description.
	assert.Len(t, errs, 2)
}

func TestLint_NoToolsRegisteredReturnsNil(t *testing.T) {
	assert.Nil(t, Lint(func(r *Registry) {}))
}

type missingDescOut struct {
	Result string `json:"result"`
}

func TestLint_FailsOnMissingOutputFieldDescription(t *testing.T) {
	errs := Lint(func(r *Registry) {
		Register(r, Tool[noInput, missingDescOut]{
			Name:        "list_things",
			Description: "ok",
			Handler: func(ctx context.Context, _ noInput) (missingDescOut, error) {
				return missingDescOut{}, nil
			},
		})
	})
	require.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), `"result"`)
}

type innerNested struct {
	Name string `json:"name"`
}

func (innerNested) Validate() error { return nil }

type nestedInput struct {
	Inner innerNested `json:"inner" jsonschema:"an inner object"`
}

func (nestedInput) Validate() error { return nil }

func TestLint_FailsOnMissingNestedFieldDescription(t *testing.T) {
	errs := Lint(func(r *Registry) {
		Register(r, Tool[nestedInput, emptyOut]{
			Name: "get_thing", Description: "ok", Handler: okHandler[nestedInput],
		})
	})
	require.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "inner.name")
}
