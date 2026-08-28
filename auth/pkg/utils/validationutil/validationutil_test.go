package validationutil

import (
	"context"
	"testing"

	"github.com/govalues/decimal"
	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Name  string          `validate:"required"`
	Email string          `validate:"required,email"`
	Age   int             `validate:"min=18"`
	Price decimal.Decimal `validate:"required"`
}

func TestValidateStruct(t *testing.T) {
	ctx := context.Background()
	d, _ := decimal.NewFromFloat64(100.0)
	valid := TestStruct{Name: "John", Email: "john@example.com", Age: 25, Price: d}
	err := ValidateStruct(ctx, &valid)
	assert.NoError(t, err)

	invalid := TestStruct{Name: "", Email: "invalid", Age: 15}
	err = ValidateStruct(ctx, &invalid)
	assert.Error(t, err)
}
