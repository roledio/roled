package numberutil

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gookit/goutil/mathutil"
)

// Reference for byte unit to use (SI or IEC):
// https://www.reddit.com/r/computerscience/comments/u26awl/why_do_si_prefixes_and_iec_prefixes_both_exist/
// https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/

func ByteCountSI(val int) string {
	const unit = 1000
	if val < unit {
		return fmt.Sprintf("%d B", val)
	}
	div, exp := int64(unit), 0
	for n := val / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(val)/float64(div), "kMGTPE"[exp])
}

func ByteCountIEC(val int) string {
	const unit = 1024
	if val < unit {
		return fmt.Sprintf("%d B", val)
	}
	div, exp := int64(unit), 0
	for n := val / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(val)/float64(div), "KMGTPE"[exp])
}

// Format formats a number (int, uint, or float) with thousand separators
// according to the given locale ("id" or "en") and decimal length.
// If decimalLength == 0, the result has no decimal part.
func Format(num any, locale string, decimalLength int) string {
	var thousandSep, decimalSep string
	switch locale {
	case "id":
		thousandSep = "."
		decimalSep = ","
	default: // en
		thousandSep = ","
		decimalSep = "."
	}

	if decimalLength == 0 {
		str := strconv.FormatFloat(ToFloat(num), 'f', 0, 64)
		return formatIntegerString(str, thousandSep)
	}

	str := strconv.FormatFloat(ToFloat(num), 'f', decimalLength, 64)
	parts := strings.Split(str, ".")
	intPart := formatIntegerString(parts[0], thousandSep)
	return intPart + decimalSep + parts[1]
}

func ToFloat(num any) float64 {
	val, err := mathutil.ToFloatWith(num, mathutil.WithHandlePtr[float64])
	if err != nil {
		panic(err)
	}
	return val
}

// formatIntegerString inserts thousand separators into a numeric string.
func formatIntegerString(s string, sep string) string {
	n := len(s)
	if n <= 3 {
		return s
	}

	// handle negative numbers
	neg := ""
	if s[0] == '-' {
		neg = "-"
		s = s[1:]
		n--
	}

	var result []string
	for i, c := range s {
		if (n-i)%3 == 0 && i != 0 {
			result = append(result, sep)
		}
		result = append(result, string(c))
	}

	return neg + strings.Join(result, "")
}
