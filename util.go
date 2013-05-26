package goweb

import (
	"strconv"
)

func ToString(str string, defaultvalue string) string {
	if str == "" {
		return defaultvalue
	}
	return str
}

func ToInt(str string, defaultvalue int) int {
	val, err := strconv.Atoi(str)
	if err != nil {
		return defaultvalue
	}
	return val
}

func ToBool(str string, defaultvalue bool) bool {
	val, err := strconv.ParseBool(str)
	if err != nil {
		return defaultvalue
	}
	return val
}
