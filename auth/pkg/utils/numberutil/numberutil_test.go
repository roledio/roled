package numberutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestByteCountSI(t *testing.T) {
	assert.Equal(t, "500 B", ByteCountSI(500))
	assert.Equal(t, "1.5 kB", ByteCountSI(1500))
	assert.Equal(t, "2.0 MB", ByteCountSI(2000000))
	assert.Equal(t, "1.0 GB", ByteCountSI(1000000000))
	assert.NotEqual(t, "1500 B", ByteCountSI(1500))
}

func TestByteCountIEC(t *testing.T) {
	assert.Equal(t, "500 B", ByteCountIEC(500))
	assert.Equal(t, "1.5 KiB", ByteCountIEC(1500))
	assert.Equal(t, "1.9 MiB", ByteCountIEC(2000000))
	assert.Equal(t, "953.7 MiB", ByteCountIEC(1000000000))
}

func TestFormat(t *testing.T) {
	assert.Equal(t, "1,234,567", Format(1234567, "en", 0))
	assert.Equal(t, "1.234.567", Format(1234567, "id", 0))
	assert.Equal(t, "1,234.57", Format(1234.567, "en", 2))
	assert.Equal(t, "-1,234,567", Format(-1234567, "en", 0))
	assert.Equal(t, "123", Format(123, "en", 0))
	assert.Equal(t, "1,000", Format(1000, "en", 0))
}

func TestToFloat(t *testing.T) {
	assert.Equal(t, 123.0, ToFloat(123))
	assert.Equal(t, 123.45, ToFloat(123.45))
	assert.Panics(t, func() { ToFloat("invalid") })
}

func TestFormatIntegerString(t *testing.T) {
	assert.Equal(t, "123", formatIntegerString("123", ","))
	assert.Equal(t, "1,234", formatIntegerString("1234", ","))
	assert.Equal(t, "12,345", formatIntegerString("12345", ","))
	assert.Equal(t, "123,456", formatIntegerString("123456", ","))
	assert.Equal(t, "-1,234", formatIntegerString("-1234", ","))
	assert.Equal(t, "-123", formatIntegerString("-123", ","))
}
