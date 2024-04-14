package utils

import (
	"reflect"
	"strings"
)

func RepeatString(n any, s string) string {
	// n must be any of int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64
	if reflect.TypeOf(n).Kind() == reflect.String {
		return ""
	}

	// check if unsigned
	if reflect.TypeOf(n).Kind() == reflect.Uint || reflect.TypeOf(n).Kind() == reflect.Uint8 || reflect.TypeOf(n).Kind() == reflect.Uint16 || reflect.TypeOf(n).Kind() == reflect.Uint32 || reflect.TypeOf(n).Kind() == reflect.Uint64 {
		return strings.Repeat(s, int(reflect.ValueOf(n).Uint()))
	}

	return strings.Repeat(s, int(reflect.ValueOf(n).Int()))
}
