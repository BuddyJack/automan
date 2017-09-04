package ntos

import "strconv"

func F64toS(f float64) string {
	return strconv.FormatFloat(f, 'f', 0, 64)
}

func F64toS2(f float64, p int) string {
	return strconv.FormatFloat(f, 'f', p, 64)
}
