package apperr

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestError_Error(t *testing.T) {
	t.Run("with message", func(t *testing.T) {
		err := New(CodeInvalid, "validation failed")
		require.Equal(t, "validation failed", err.Error())
	})

	t.Run("without message, with cause", func(t *testing.T) {
		cause := errors.New("database error")
		err := Wrap(cause, CodeInternal, "")
		require.Equal(t, "database error", err.Error())
	})

	t.Run("without message, without cause", func(t *testing.T) {
		err := New(CodeNotFound, "")
		require.Equal(t, "not_found", err.Error())
	})
}

func TestWrap(t *testing.T) {
	cause := errors.New("original error")
	err := Wrap(cause, CodeInvalid, "wrapped")

	require.Equal(t, CodeInvalid, err.Code)
	require.Equal(t, "wrapped", err.Msg)
	require.Equal(t, cause, err.Cause)
}

func TestCodeOf(t *testing.T) {
	t.Run("direct apperr", func(t *testing.T) {
		err := New(CodeNotFound, "not found")
		require.Equal(t, CodeNotFound, CodeOf(err))
	})

	t.Run("wrapped apperr", func(t *testing.T) {
		original := New(CodeConflict, "conflict")
		wrapped := Wrap(original, CodeInternal, "wrapped")
		require.Equal(t, CodeInternal, CodeOf(wrapped))
	})

	t.Run("non-apperr", func(t *testing.T) {
		err := errors.New("standard error")
		require.Equal(t, CodeInternal, CodeOf(err))
	})
}

func TestWithFields(t *testing.T) {
	err := New(CodeInvalid, "validation failed")
	fields := map[string]any{"field": "email"}

	withFields := WithFields(err, fields)
	require.Equal(t, CodeInvalid, withFields.Code)
	require.Equal(t, "validation failed", withFields.Msg)
	require.Equal(t, fields, withFields.Fields)
}
