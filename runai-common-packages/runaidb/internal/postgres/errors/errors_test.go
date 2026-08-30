package pgerrors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func TestTranslate_Nil(t *testing.T) {
	require.Nil(t, Translate(nil))
}

func TestTranslate_MapsCodesToSentinels(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		sentinel error
	}{
		{"read only transaction", ErrCodeReadOnlyTransaction, ErrReadOnlyConnection},
		{"undefined table", ErrCodeUndefinedTable, ErrUndefinedTable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pgErr := &pgconn.PgError{Code: tt.code, Message: "boom"}
			wrapped := fmt.Errorf("exec failed: %w", pgErr)

			require.ErrorIs(t, Translate(pgErr), tt.sentinel)
			require.ErrorIs(t, Translate(wrapped), tt.sentinel)
		})
	}
}

func TestTranslate_SentinelsAreDistinct(t *testing.T) {
	translated := Translate(&pgconn.PgError{Code: ErrCodeUndefinedTable})

	require.ErrorIs(t, translated, ErrUndefinedTable)
	require.NotErrorIs(t, translated, ErrReadOnlyConnection)
}

func TestTranslate_PreservesOriginalError(t *testing.T) {
	pgErr := &pgconn.PgError{Code: ErrCodeReadOnlyTransaction, Message: "cannot execute UPDATE"}
	translated := Translate(fmt.Errorf("exec failed: %w", pgErr))

	var got *pgconn.PgError
	require.ErrorAs(t, translated, &got)
	require.Same(t, pgErr, got)
	require.Contains(t, translated.Error(), "cannot execute UPDATE")
}

func TestTranslate_IsIdempotent(t *testing.T) {
	once := Translate(&pgconn.PgError{Code: ErrCodeReadOnlyTransaction})

	require.Same(t, once, Translate(once))
	require.ErrorIs(t, Translate(once), ErrReadOnlyConnection)
}

func TestTranslate_UntranslatedCodesPassThrough(t *testing.T) {
	for _, code := range []string{ErrCodeDuplicateObject, ErrCodeInsufficientPrivilege, "23505"} {
		pgErr := &pgconn.PgError{Code: code}

		require.Same(t, pgErr, Translate(pgErr))
		require.NotErrorIs(t, Translate(pgErr), ErrReadOnlyConnection)
		require.NotErrorIs(t, Translate(pgErr), ErrUndefinedTable)
	}
}

func TestTranslate_GenericError(t *testing.T) {
	err := errors.New("connection refused")

	require.Same(t, err, Translate(err))
	require.NotErrorIs(t, Translate(err), ErrReadOnlyConnection)
}

func TestHasCode(t *testing.T) {
	pgErr := &pgconn.PgError{Code: ErrCodeDuplicateObject}

	require.True(t, HasCode(pgErr, ErrCodeDuplicateObject))
	require.True(t, HasCode(fmt.Errorf("create user: %w", pgErr), ErrCodeDuplicateObject))
	require.False(t, HasCode(pgErr, ErrCodeInsufficientPrivilege))
	require.False(t, HasCode(errors.New("connection refused"), ErrCodeDuplicateObject))
	require.False(t, HasCode(nil, ErrCodeDuplicateObject))
}

func TestHasCode_AfterTranslate(t *testing.T) {
	translated := Translate(&pgconn.PgError{Code: ErrCodeReadOnlyTransaction})

	require.True(t, HasCode(translated, ErrCodeReadOnlyTransaction))
	require.ErrorIs(t, translated, ErrReadOnlyConnection)
}
