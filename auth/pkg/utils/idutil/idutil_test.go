package idutil

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/lithammer/shortuuid/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewID(t *testing.T) {
	id := NewID()
	fmt.Println(id)
	// 2JLwGJGHX5T9oaEc6a548K
	// 2JLwGKroaUJNoYQanAdHzh
	// 2JLwGLxUygUsYgDJXMwtDL
	// 2JLwGNEPmhBVFk3PrQ9huJ
	// 2JLwGPLfxyHAetpkiyw7pU
}

func TestNanoID_DefaultLength(t *testing.T) {
	result := NanoID()
	assert.Len(t, result, NanoIDDefaultSize, "NanoID() should have default length")
	validChars := regexp.MustCompile(fmt.Sprintf(`^[%s]+$`, shortuuid.DefaultAlphabet))
	assert.Regexp(t, validChars, result, "NanoID() should only contain custom characters")
}

func TestNanoID_CustomLength(t *testing.T) {
	customLength := 10
	result := NanoID(customLength)
	assert.Len(t, result, customLength, "NanoID(10) should have custom length")
	validChars := regexp.MustCompile(fmt.Sprintf(`^[%s]+$`, shortuuid.DefaultAlphabet))
	assert.Regexp(t, validChars, result, "NanoID(10) should only contain custom characters")
}

func TestNanoID_Randomness(t *testing.T) {
	result1 := NanoID()
	result2 := NanoID()
	assert.NotEqual(t, result1, result2, "NanoID() should produce different results")
}
