package log

import "fmt"

func FormatString(text string, length int, alignLeft bool) string {
	if len(text) > length {
		// Truncate the string if it's longer than the desired length
		return text[:length]
	}

	// CreateUser a format string based on alignment preference
	var formatStr string
	if alignLeft {
		formatStr = fmt.Sprintf("%%-%ds", length) // Left align
	} else {
		formatStr = fmt.Sprintf("%%%ds", length) // Right align
	}

	// Use the format string to format the text
	return fmt.Sprintf(formatStr, text)
}
