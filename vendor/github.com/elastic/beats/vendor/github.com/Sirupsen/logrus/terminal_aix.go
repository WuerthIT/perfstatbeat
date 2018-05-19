package logrus

import "io"

// Not yet implemented but have function available for compatibility.
// Always return false as of now.
func IsTerminal(f io.Writer) bool {
	return false
}
