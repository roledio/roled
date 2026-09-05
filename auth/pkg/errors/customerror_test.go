package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomError_Error(t *testing.T) {
	err := CustomError{Msg: "test error"}
	assert.Equal(t, "test error", err.Error())
}

func TestCustomError_WithError(t *testing.T) {
	ce := CustomError{Msg: "test"}
	ce = ce.WithError(errors.New("wrapped"))
	assert.Equal(t, "wrapped", ce.Err.Error())
}

func TestCustomError_Is(t *testing.T) {
	baseErr := CustomError{Code: "test_code", Msg: "test msg"}
	wrappedErr := baseErr.WithError(errors.New("underlying error"))
	diffErr := CustomError{Code: "other_code", Msg: "other msg"}

	assert.True(t, errors.Is(wrappedErr, baseErr))
	assert.False(t, errors.Is(wrappedErr, diffErr))
	assert.False(t, errors.Is(wrappedErr, errors.New("other")))
}
