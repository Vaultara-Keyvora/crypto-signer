//go:build windows

package program

import (
	"os"
	"strings"
)

func GetProcessName() string {
	args := os.Args
	if len(args) > 0 {
		segments := strings.Split(args[0], `\`)
		if len(segments) > 0 {
			return segments[len(segments)-1]
		}
	}
	return ""
}
