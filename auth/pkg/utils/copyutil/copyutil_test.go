package copyutil

import (
	"testing"

	"github.com/govalues/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCopy(t *testing.T) {
	type Src struct {
		Name string
		Age  int
	}
	type Dst struct {
		Name string
		Age  int
	}
	src := Src{Name: "John", Age: 30}
	var dst Dst
	err := Copy(&src, &dst)
	assert.NoError(t, err)
	assert.Equal(t, "John", dst.Name)
	assert.Equal(t, 30, dst.Age)
}

func TestCopyWithDecimalToFloat64(t *testing.T) {
	type Src struct {
		Value decimal.Decimal
	}
	type Dst struct {
		Value float64
	}
	d, _ := decimal.NewFromFloat64(123.45)
	src := Src{Value: d}
	var dst Dst
	err := Copy(&src, &dst)
	assert.NoError(t, err)
	assert.Equal(t, 123.45, dst.Value)
}

func TestDecimalToFloat64Converter(t *testing.T) {
	d, _ := decimal.NewFromFloat64(123.45)
	f, err := DecimalToFloat64Converter().Fn(d)
	assert.NoError(t, err)
	assert.Equal(t, 123.45, f)
}

func TestDecimalToFloat64Converter_Error(t *testing.T) {
	conv := DecimalToFloat64Converter()
	_, err := conv.Fn("not a decimal")
	assert.Error(t, err)
	assert.Equal(t, "src type is not decimal.Decimal", err.Error())
}

func TestDecimalToFloat64Converter_Zero(t *testing.T) {
	d := decimal.Zero
	f, err := DecimalToFloat64Converter().Fn(d)
	assert.NoError(t, err)
	assert.Equal(t, 0.0, f)
}

func TestDecimalToFloat64Converter_Negative(t *testing.T) {
	d, _ := decimal.NewFromFloat64(-99.99)
	f, err := DecimalToFloat64Converter().Fn(d)
	assert.NoError(t, err)
	assert.Equal(t, -99.99, f)
}
