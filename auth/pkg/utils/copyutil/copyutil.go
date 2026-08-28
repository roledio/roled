package copyutil

import (
	"errors"

	"github.com/govalues/decimal"
	"github.com/jinzhu/copier"
)

func Copy(from any, to any) error {
	return copier.CopyWithOption(to, from, copier.Option{
		Converters: []copier.TypeConverter{
			DecimalToFloat64Converter(),
		},
	})
}

func DecimalToFloat64Converter() copier.TypeConverter {
	return copier.TypeConverter{
		SrcType: decimal.Decimal{},
		DstType: copier.Float64,
		Fn: func(src any) (any, error) {
			d, ok := src.(decimal.Decimal)
			if !ok {
				return nil, errors.New("src type is not decimal.Decimal")
			}
			f, _ := d.Float64()
			return f, nil
		},
	}
}
