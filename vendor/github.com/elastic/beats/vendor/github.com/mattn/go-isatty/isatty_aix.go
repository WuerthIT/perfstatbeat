package isatty

// Not yet implemented but have function available for compatibility.
// Always return false as of now.
func IsTerminal(fd uintptr) bool {
	return false
}
