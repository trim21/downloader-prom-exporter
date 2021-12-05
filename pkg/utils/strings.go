package utils

import (
	"strconv"
	"strings"
)

func SplitByComma(raw string) []string {
	s := strings.Split(raw, ",")

	vsm := make([]string, len(s))
	for i, v := range s {
		vsm[i] = strings.TrimSpace(v)
	}

	return vsm
}

func ByteCountSI(b int64) string {
	const unit int64 = 1000
	if b < unit {
		return strconv.FormatInt(b, 10) + " B"
	}

	div, exp := unit, 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return strconv.FormatFloat(float64(b)/float64(div), 'f', 1, 64) +
		" " + string("kMGTPE"[exp]) + "iB"
}

func ByteCountIEC(b int64) string {
	const unit int64 = 1024
	if b < unit {
		return strconv.FormatInt(b, 10) + " B"
	}

	div, exp := unit, 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return strconv.FormatFloat(float64(b)/float64(div), 'f', 1, 64) +
		" " + string("KMGTPE"[exp]) + "iB"
}

func ByteCountIECFloat64(b float64) string {
	const unit = 1024.0
	if b < unit {
		return strconv.FormatFloat(b, 'f', 1, 64) + " B"
	}

	div, exp := unit, 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return strconv.FormatFloat(b/div, 'f', 1, 64) + " " + string("KMGTPE"[exp]) + "iB"
}
