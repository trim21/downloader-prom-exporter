package utils

import (
	"fmt"
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

func byteCount(b int64, unit int64) string {
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}

	div, exp := unit, 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}

func ByteCountSI(b int64) string {
	return byteCount(b, 1000) //nolint:gomnd
}

func ByteCountIEC(b int64) string {
	return byteCount(b, 1024) //nolint:gomnd
}
