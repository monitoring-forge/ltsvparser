package ltsvparser

import "strconv"

var lenSmallPowersOfTen = 15
var smallPowersOfTen = [...]float64{
	1,
	10,
	100,
	1_000,
	10_000,
	100_000,
	1_000_000,
	10_000_000,
	100_000_000,
	1_000_000_000,
	10_000_000_000,
	100_000_000_000,
	1_000_000_000_000,
	10_000_000_000_000,
	100_000_000_000_000,
}

// ParseFloat uses a direct path for the common LTSV decimal format.
// Other formats use strconv so its full syntax and rounding guarantees remain intact.
func ParseFloat(value []byte) (float64, error) {
	parsed, ok := parseSimpleFloat(value)
	if ok {
		return parsed, nil
	}
	return strconv.ParseFloat(string(value), 64)
}

//nolint:gocognit // Inline the fixed-precision path to avoid dispatch calls on other formats.
func parseSimpleFloat(value []byte) (float64, bool) {
	l := len(value)
	if l == 0 {
		return 0, false
	}

	index := 0
	negative := false
	if value[index] == '-' {
		negative = true
		index++
		if index == l {
			return 0, false
		}
	}

	// Common timing values have one integer digit and three or six decimal
	// places. Check the length first so other formats avoid extra byte loads.
	// Keep this in the existing parser to share sign handling and avoid a
	// separate dispatch call on the general path.
	remaining := l - index
	if (remaining == 5 || remaining == 8) && value[index+1] == '.' {
		v := value[index:]
		a, b, c, d := v[0]-'0', v[2]-'0', v[3]-'0', v[4]-'0'
		if a > 9 || b > 9 || c > 9 || d > 9 {
			return 0, false
		}
		mantissa := uint64(a)*1000 + uint64(b)*100 + uint64(c)*10 + uint64(d)
		divisor := float64(1000)
		if remaining == 8 {
			e, f, g := v[5]-'0', v[6]-'0', v[7]-'0'
			if e > 9 || f > 9 || g > 9 {
				return 0, false
			}
			mantissa = mantissa*1000 + uint64(e)*100 + uint64(f)*10 + uint64(g)
			divisor = 1000000
		}
		result := float64(mantissa) / divisor
		if negative {
			result = -result
		}
		return result, true
	}

	var mantissa uint64
	decimalPlaces := 0
	hasDecimalPoint := false
	digits := 0
	for ; index < l; index++ {
		character := value[index]
		if character == '.' && !hasDecimalPoint {
			hasDecimalPoint = true
			continue
		}
		if character < '0' || character > '9' || digits == lenSmallPowersOfTen {
			return 0, false
		}

		mantissa = mantissa*10 + uint64(character-'0')
		digits++
		if hasDecimalPoint {
			decimalPlaces++
		}
	}
	if digits == 0 || decimalPlaces >= lenSmallPowersOfTen {
		return 0, false
	}

	result := float64(mantissa) / smallPowersOfTen[decimalPlaces]
	if negative {
		result = -result
	}
	return result, true
}
