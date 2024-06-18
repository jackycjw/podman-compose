package util

import (
	"fmt"
	"strings"
)

func TextColor(color int, str string) string {
	return fmt.Sprintf("\x1b[0;%dm%s\x1b[0m", color, str)
}

func FixSizeString(str string, length int, middle bool) string {
	if len(str) == length {
		return str
	} else if len(str) > length {
		return str[0:length]
	} else {
		if middle {
			left := length - len(str)
			before := left / 2
			after := left - before
			return strings.Repeat(" ", before) + str + strings.Repeat(" ", after)
		} else {
			return str + strings.Repeat(" ", length-len(str))
		}

	}
}

func GetMapSize(str string, length int, middle bool) string {
	if len(str) == length {
		return str
	} else if len(str) > length {
		return str[0:length]
	} else {
		if middle {
			left := length - len(str)
			before := left / 2
			after := left - before
			return strings.Repeat(" ", before) + str + strings.Repeat(" ", after)
		} else {
			return str + strings.Repeat(" ", length-len(str))
		}

	}
}
