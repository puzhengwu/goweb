package goweb

import (
	"strconv"
)

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
