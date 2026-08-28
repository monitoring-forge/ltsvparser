package ltsvparser

import "strconv"

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

// parseFloat uses a direct path for the common LTSV decimal format.
// Other formats use strconv so its full syntax and rounding guarantees remain intact.
func ParseFloat(value []byte) (float64, error) {
	parsed, ok := parseSimpleFloat(value)
	if ok {
		return parsed, nil
	}
	return strconv.ParseFloat(string(value), 64)
}

func parseSimpleFloat(value []byte) (float64, bool) {
	if len(value) == 0 {
		return 0, false
	}

	index := 0
	negative := false
	if value[index] == '-' {
		negative = true
		index++
		if index == len(value) {
			return 0, false
		}
	}

	var mantissa uint64
	decimalPlaces := 0
	hasDecimalPoint := false
	digits := 0
	for ; index < len(value); index++ {
		character := value[index]
		if character == '.' && !hasDecimalPoint {
			hasDecimalPoint = true
			continue
		}
		if character < '0' || character > '9' || digits == len(smallPowersOfTen) {
			return 0, false
		}

		mantissa = mantissa*10 + uint64(character-'0')
		digits++
		if hasDecimalPoint {
			decimalPlaces++
		}
	}
	if digits == 0 || decimalPlaces >= len(smallPowersOfTen) {
		return 0, false
	}

	result := float64(mantissa) / smallPowersOfTen[decimalPlaces]
	if negative {
		result = -result
	}
	return result, true
}
