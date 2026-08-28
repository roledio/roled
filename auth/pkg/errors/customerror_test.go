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
