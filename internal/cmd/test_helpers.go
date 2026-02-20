package cmd

// FindWailsCLIForTest exposes Wails binary discovery for external tests.
func FindWailsCLIForTest() (string, error) {
	return findWailsCLI()
}

// DesktopBinaryArgsForTest exposes govard-desktop argument construction for tests.
func DesktopBinaryArgsForTest(background bool) []string {
	return buildDesktopBinaryArgs(background)
}
