package jsonutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringify(t *testing.T) {
	data := map[string]string{"key": "value"}
	result := Stringify(data)
	expected := `{"key":"value"}`
	assert.Equal(t, expected, result)
}

func TestParse(t *testing.T) {
	str := `{"key":"value"}`
	var result map[string]string
	err := Parse(str, &result)
	assert.NoError(t, err)
	assert.Equal(t, "value", result["key"])
}

func TestStringify_InvalidData(t *testing.T) {
	// Data that cannot be marshaled
	data := make(chan int)
	result := Stringify(data)
	assert.Empty(t, result, "Stringify should return empty string for invalid data")
}

func TestParse_InvalidJSON(t *testing.T) {
	str := `invalid json`
	var result map[string]string
	err := Parse(str, &result)
	assert.Error(t, err, "Parse should error for invalid JSON")
}

func TestParse_TypeMismatch(t *testing.T) {
	str := `{"key":"value"}`
	var result int
	err := Parse(str, &result)
	assert.Error(t, err, "Parse should error for type mismatch")
}
