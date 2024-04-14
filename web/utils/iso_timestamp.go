package utils

import (
	"reflect"
	"time"
)

func ISOTimestamp(t any) string {
	// Check if pointer
	var timeToConvert time.Time
	if reflect.TypeOf(t).Kind() == reflect.Ptr {
		// Check if nil
		if reflect.ValueOf(t).IsNil() {
			return "1970-01-01T00:00:00"
		}
		timeToConvert = reflect.ValueOf(t).Elem().Interface().(time.Time)
	} else {
		timeToConvert = t.(time.Time)
	}
	return timeToConvert.Format("2006-01-02T15:04:05")
}
